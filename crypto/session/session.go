package session

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/render"
	"github.com/gofrs/uuid"

	"github.com/djangulo/sfd/config"
	"github.com/djangulo/sfd/crypto"
	"github.com/djangulo/sfd/db/models"
	sfdErrors "github.com/djangulo/sfd/errors"
)

func init() {
	gob.Register(uuid.UUID{})
	gob.Register(Values{})
	gob.Register(Key{})
}

var (
	// ErrNotFound not found.
	ErrNotFound = errors.New("not found")
	// ErrSessionNotFound session not found.
	ErrSessionNotFound = fmt.Errorf("session %w", ErrNotFound)
	// ErrUserNotFound user not found
	ErrUserNotFound = fmt.Errorf("user %w", ErrNotFound)
)

type Storer interface {
	// NewSession creates a new session. Values are encoded as []byte using
	// encoding/gob, so the store's don't have to bother with decoding
	// and encoding
	NewSession(Session) error
	ReadSession(id string) ([]byte, error)
	DeleteSession(id string) error
	UpdateSession(Session) error
	SessionGC() error
}

type Manager interface {
	New(values Values) (Session, error)
	Delete(id string) error
	Get(id string) (Session, error)
	Save(session Session) error
	// Rebuild rebuild an existing session from values
	Rebuild(id string, values []byte) (Session, error)
	GC(unit time.Duration, err chan<- error, cancel <-chan struct{})
	NewAuthCookie(session Session, domain string) (*http.Cookie, error)
	FromCookie(r *http.Request) (Session, error)
	Context(next http.Handler) http.Handler
	NoErrContext(next http.Handler) http.Handler
}

type Session interface {
	ID() string
	Set(key, value interface{}) error //set session value
	Get(key interface{}) interface{}  //get session value
	Delete(key interface{}) error     //delete session value
	// Bytes should return the byte-encode Values object. encoding/gob is the
	// current implementation, but any encodable/decodable solution should work.
	Bytes() ([]byte, error)
	// Save persists the changes in storage
	Save(m Manager) error
	Expiry() time.Time
	// GetUser is a shortcut method that returns the user.
	GetUser() (*models.User, error)
	// Metadata returns id, created, updated
	Metadata() (string, time.Time, *time.Time)
}

type sessionManager struct {
	sync.RWMutex
	store          Storer
	CookieName     string
	maxAge         int
	activeSessions int
	config         config.Configurer
}

func NewManager(store Storer, cookieName string, maxAge int, config config.Configurer) (Manager, error) {
	return &sessionManager{store: store, CookieName: cookieName, maxAge: maxAge, config: config}, nil
}

func (m *sessionManager) New(values Values) (Session, error) {
	m.Lock()
	defer m.Unlock()

	id := crypto.RandomString(64)

	var b bytes.Buffer
	enc := gob.NewEncoder(&b)
	if err := enc.Encode(values); err != nil {
		return nil, err
	}
	now := time.Now()
	expires := now.Add(time.Duration(m.maxAge) * time.Second)
	s := &sessionObject{
		id:        id,
		values:    values,
		createdAt: now,
		expiry:    expires,
	}

	if err := m.store.NewSession(s); err != nil {
		return nil, err
	}
	return s, nil
}

func (m *sessionManager) Delete(id string) error {
	m.RLock()
	defer m.RUnlock()
	return m.store.DeleteSession(id)
}

func (m *sessionManager) Rebuild(id string, values []byte) (Session, error) {
	b := bytes.NewBuffer(values)
	dec := gob.NewDecoder(b)
	ses := &sessionObject{id: id}
	if err := dec.Decode(&ses.values); err != nil {
		return nil, err
	}
	return ses, nil
}

func (m *sessionManager) Get(id string) (Session, error) {
	m.RLock()
	defer m.RUnlock()

	b, err := m.store.ReadSession(id)
	if err != nil {
		return nil, err
	}

	return m.Rebuild(id, b)
}

func (m *sessionManager) Save(ses Session) error {
	m.Lock()
	defer m.Unlock()

	if s, ok := ses.(*sessionObject); ok {
		now := time.Now()
		s.updatedAt = &now
	}

	if err := m.store.UpdateSession(ses); err != nil {
		return nil
	}
	return nil
}

// GC needs to be run in a goroutine to clean expired sessions. It'll start a
// time.Ticker every unit and periodically call the store GC method every unit.
func (m *sessionManager) GC(unit time.Duration, e chan<- error, cancel <-chan struct{}) {

	tick := time.Tick(1 * unit)
	for {
		select {
		case <-tick:
			m.Lock()
			if err := m.store.SessionGC(); err != nil {
				e <- err
			}
			m.Unlock()
		case <-cancel:
			return
		}
	}
}

type Values map[interface{}]interface{}

func (v Values) Set(key, value interface{}) error {
	if v == nil {
		v = make(Values)
	}
	v[key] = value
	return nil
}

// Session note ByteValues field is db-encoded as values.
type sessionObject struct {
	sync.RWMutex
	id        string `db:"id"`
	values    Values
	expiry    time.Time  `db:"expires"`
	createdAt time.Time  `db:"created_at"`
	updatedAt *time.Time `db:"updated_at"`
}

func (s *sessionObject) ID() string {
	s.RLock()
	defer s.RUnlock()
	return s.id
}

func (s *sessionObject) Set(key, value interface{}) error {
	s.Lock()
	defer s.Unlock()
	if s.values == nil {
		s.values = make(Values)
	}
	s.values[key] = value
	return nil
}

func (s *sessionObject) Bytes() ([]byte, error) {
	s.RLock()
	defer s.RUnlock()
	var b bytes.Buffer
	enc := gob.NewEncoder(&b)
	if err := enc.Encode(s.values); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}
func (s *sessionObject) Expiry() time.Time {
	s.RLock()
	defer s.RUnlock()
	return s.expiry
}

func (s *sessionObject) Get(key interface{}) interface{} {
	s.RLock()
	defer s.RUnlock()
	if key != nil {
		if v, ok := s.values[key]; ok {
			return v
		}
	}
	return nil
}

func (s *sessionObject) Delete(key interface{}) error {
	s.Lock()
	defer s.Unlock()
	delete(s.values, key)
	return nil
}

func (s *sessionObject) Save(m Manager) error {
	return m.Save(s)
}

func (s *sessionObject) GetUser() (*models.User, error) {
	// gob encoded values return structs, not pointers
	if user, ok := s.Get(UserKey).(models.User); ok {
		return &user, nil
	}
	return nil, ErrUserNotFound
}

func (s *sessionObject) Metadata() (string, time.Time, *time.Time) {
	return s.id, s.createdAt, s.updatedAt
}

type Sessioner interface {
	Set(key, value interface{}) error //set session value
	Get(key interface{}) interface{}  //get session value
	Delete(key interface{}) error     //delete session value
	// Save persists the changes in storage
	Save(m *sessionManager) error
	SessionID() string //back current sessionID
}

// NewAuthCookie create a new secure cookie for the session
func (m *sessionManager) NewAuthCookie(ses Session, domain string) (*http.Cookie, error) {
	expiry := time.Now().Add(time.Duration(m.maxAge) * time.Second) // 1 week

	if domain == "" {
		domain = m.config.SiteHost()
	}
	id, _, _ := ses.Metadata()

	return &http.Cookie{
		Name:     m.CookieName,
		Value:    id,
		Expires:  expiry,
		MaxAge:   int(expiry.Sub(time.Now()).Seconds()),
		Secure:   true,
		HttpOnly: true,
		Domain:   domain,
		SameSite: http.SameSiteStrictMode,
		Path:     "/", // needs research and testing
	}, nil
}

type ctxKey int

const (
	// CtxKey used to store sessions in context.
	CtxKey ctxKey = iota + 1
	// AuthCookieCtxKey used to store AuthCookie in context.
	AuthCookieCtxKey
	// UserCtxKey used to store a *models.User in context.
	UserCtxKey
	// AuthCtxKey used to store a bool in context.
	AuthCtxKey
	// DataCtxKey used to store some initialized data in context.
	DataCtxKey
)

// Key default session keys for certain values.
type Key struct {
	Name string
}

var (
	// UserKey session key ("user").
	UserKey = Key{"user"}
	// AuthKey session key ("is_authenticated").
	AuthKey = Key{"is_authenticated"}
)

func (k Key) String() string {
	return k.Name
}

// FromCookie extracts a *Session from the *http.Request using the session
// cookie.
func (m *sessionManager) FromCookie(r *http.Request) (Session, error) {
	cookie, err := r.Cookie(m.CookieName)
	if err != nil {
		return nil, fmt.Errorf("cookie %s: %w", m.CookieName, ErrNotFound)
	}
	ses, err := m.Get(cookie.Value)
	if err != nil {
		return nil, fmt.Errorf("session: %w", ErrNotFound)
	}
	return ses, nil
}

// FromContext extracts a *Session from ctx.
func FromContext(ctx context.Context) (Session, error) {
	ses, ok := ctx.Value(CtxKey).(Session)
	if !ok {
		return nil, fmt.Errorf("%w in context", ErrNotFound)
	}
	return ses, nil
}

var ErrRedirectToLogin = errors.New("session not available, redirect to login")

type RedirectToLoginResponse struct {
	Next string `json:"next"`
}

// Context reads session cookie and add session to request context. Renders
// an error if the cookie is not present. It's up to the consumer to redirect
// to login.
func (m *sessionManager) Context(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		ses, err := m.FromCookie(r)
		if err != nil {
			render.Render(w, r, sfdErrors.ErrRedirectToLogin)
			return
		}
		ctx := context.WithValue(r.Context(), CtxKey, ses)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// NoErrContext reads session cookie and add session to request context.
// It does not render any errors if the session is not found.
func (m *sessionManager) NoErrContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		ses, err := m.FromCookie(r)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}
		ctx := context.WithValue(r.Context(), CtxKey, ses)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
