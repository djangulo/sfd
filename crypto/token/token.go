package token

import (
	"context"
	"crypto/hmac"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/render"
	"github.com/gofrs/uuid"

	"github.com/djangulo/sfd/config"
	"github.com/djangulo/sfd/crypto"
	"github.com/djangulo/sfd/db/models"
	sfdErrors "github.com/djangulo/sfd/errors"
)

var (
	ErrNotFound          = errors.New("not found")
	ErrCSRFTokenNotFound = fmt.Errorf("csrf token %w", ErrNotFound)
	ErrUnrecognizedKind  = errors.New("unrecognized token kind")
)

// Storer token storage interface.
type Storer interface {
	GetToken(digest string, kind Kind) (*Token, error)
	SaveToken(token *Token) error
	// DeleteToken explicitly deletes a token from the store.
	DeleteToken(digest string) error
	TokenGC() error
}

type Manager interface {
	GC(unit time.Duration, err chan<- error, cancel <-chan struct{})
	CheckToken(digest string, kind Kind) (*Token, error)
	DropToken(digest string) error
	NewToken(login time.Time, id *uuid.UUID, pwHash string, kind Kind, expiry *time.Time) (*Token, error)
	CSRFContext(next http.Handler) http.Handler
	NewCSRFCookie(userID *uuid.UUID, domain string) (*http.Cookie, error)
	CSRFFromCookie(r *http.Request) (*Token, error)
}

type Kind int

const (
	// Registration token.
	Registration Kind = iota + 1
	// PasswordReset token.
	PasswordReset
	// CSRF token.
	CSRF
	// Redirect token
	Redirect
	// State if valid, restore stateful data to the frontend
	State
)

var TokenKinds = []string{
	Registration:  "Registration",
	PasswordReset: "PasswordReset",
	CSRF:          "CSRF",
	Redirect:      "Redirect",
	State:         "State",
}

func (kind Kind) String() string {
	return TokenKinds[kind]
}

type Token struct {
	Digest    string       `json:"digest" db:"digest"`
	Expires   time.Time    `json:"expires" db:"expires"`
	Kind      Kind         `json:"kind" db:"kind"`
	CreatedAt time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt sql.NullTime `json:"updated_at" db:"updated_at"`
	UserID    *uuid.UUID   `json:"user_id" db:"user_id"`
	PWHash    string       `json:"pw_hash" db:"pw_hash"`
	LastLogin time.Time    `json:"last_login" db:"last_login"`
}

// CSRFToken simple wrapper around Token to add utility methods.
type CSRFToken struct {
	*Token
}

func (csrf *CSRFToken) UnmarshalForm(form url.Values) error {
	csrf.Token.Digest = form.Get("csrf_token")
	if csrf.Token.Digest == "" {
		return ErrCSRFTokenNotFound
	}
	return nil
}

// AsInput renders the CSRFToken in an input[type="hidden"].
func (csrf *CSRFToken) AsInput() template.HTML {
	str := fmt.Sprintf(`<input type="hidden" name="csrf_token" id="csrf_token" value="%s">`, csrf.Digest)
	return template.HTML(str)
}

type tokenManager struct {
	lock     sync.Mutex
	hashFunc func() hash.Hash
	// MaxAge sets the maximum age in seconds for the tokens to live.
	MaxAge int
	store  Storer
	config config.Configurer
}

// GC needs to be run in a goroutine to clean expired tokens. It'll start a
// time.Ticker every unit and periodically call the store GC method.
func (m *tokenManager) GC(unit time.Duration, errChan chan<- error, cancel <-chan struct{}) {

	tick := time.Tick(1 * unit)
	for {
		select {
		case <-tick:
			m.lock.Lock()
			if err := m.store.TokenGC(); err != nil {
				errChan <- err
			}
			m.lock.Unlock()
		case <-cancel:
			return
		}
	}
}

func NewManager(
	store Storer,
	hashFunc func() hash.Hash,
	cfg config.Configurer) (Manager, error) {
	return &tokenManager{store: store, hashFunc: hashFunc, config: cfg}, nil
}

func (m *tokenManager) NewToken(
	lastLogin time.Time,
	id *uuid.UUID,
	pwHash string,
	kind Kind,
	expiry *time.Time,
) (*Token, error) {
	now := time.Now()
	timestamp := now.Sub(time.Date(2020, 1, 1, 0, 0, 0, 0, m.config.TimeZone()))
	var login time.Time
	if !lastLogin.IsZero() {
		login = lastLogin
	} else {
		login = time.Time{} // zero as a flag value
	}
	token := Token{
		Kind:      kind,
		CreatedAt: now,
		UpdatedAt: sql.NullTime{Valid: true, Time: now},
		UserID:    id,
		LastLogin: login,
		PWHash:    pwHash,
	}
	token.Digest = m.tokenWithTimestamp(
		id,
		pwHash,
		lastLogin.In(m.config.TimeZone()),
		timestamp.Truncate(time.Second),
		kind,
	)

	if expiry != nil && !expiry.IsZero() {
		token.Expires = *expiry
	} else {
		var dur time.Duration
		switch kind {
		case CSRF:
			dur = m.config.CSRFTokenExpiry()
		case PasswordReset:
			dur = m.config.PassResetTokenExpiry()
		case Registration:
			dur = m.config.AccountConfirmationEmailExpiry()
		default:
			return nil, fmt.Errorf("%w: %s (%[1]d)", ErrUnrecognizedKind, kind)
		}
		token.Expires = time.Now().Add(dur).Truncate(time.Microsecond).In(m.config.TimeZone())
	}

	if err := m.store.SaveToken(&token); err != nil {
		return nil, err
	}
	return &token, nil
}

// CheckToken checks:
//   1. That the token exists in the store, if not, it's probably expired and has
//      been garbage collected.
//   2. That the token in question is not expired.
//   3. Compares the digest with a newly generated token.
//
// If all checks pass, CheckToken deletes the token from the store and return
// nil.
func (m *tokenManager) CheckToken(digest string, kind Kind) (*Token, error) {
	token, err := m.store.GetToken(digest, kind)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return nil, ErrInvalidToken
		}
		return nil, err
	}

	if time.Now().After(token.Expires) {
		return nil, ErrExpiredToken
	}

	if digest == "" {
		return nil, ErrInvalidToken
	}
	b64timestamp := strings.Split(digest, "-")[0]
	strTimestamp, err := base64.StdEncoding.DecodeString(b64timestamp)
	if err != nil {
		return nil, err
	}
	timestamp, err := time.ParseDuration(string(strTimestamp) + "s")

	if err != nil {
		return nil, err
	}

	otherToken := m.tokenWithTimestamp(
		token.UserID,
		token.PWHash,
		token.LastLogin.In(m.config.TimeZone()),
		timestamp.Truncate(time.Second),
		kind,
	)
	if !hmac.Equal([]byte(otherToken), []byte(digest)) {
		// no match, invalid token
		return nil, ErrInvalidToken
	}

	// all checks passed,return token, no error
	return token, nil
}

func (m *tokenManager) DropToken(digest string) error {
	if err := m.store.DeleteToken(digest); err != nil {
		log.Println(err)
		return err
	}
	return nil
}

type CtxKey int

const (
	// CSRFCtxKey used to store CSRF token in context.
	CSRFCtxKey CtxKey = iota + 2000
)

func CSRFTokenFromContext(ctx context.Context) (*CSRFToken, error) {
	t, ok := ctx.Value(CSRFCtxKey).(*CSRFToken)
	if !ok {
		return nil, ErrCSRFTokenNotFound
	}
	return t, nil
}

// NewCSRFCookie create a new secure cookie for CSRF.
func (m *tokenManager) NewCSRFCookie(userID *uuid.UUID, domain string) (*http.Cookie, error) {

	if domain == "" {
		domain = m.config.SiteHost()
	}

	expiry := time.Now().Add(m.config.CSRFTokenExpiry())
	t, err := m.NewToken(
		time.Now(),
		userID,
		crypto.RandomString(64),
		CSRF,
		&expiry,
	)
	if err != nil {
		return nil, err
	}

	return &http.Cookie{
		Name:     m.config.CSRFCookieName(),
		Value:    t.Digest,
		Expires:  expiry,
		MaxAge:   int(expiry.Sub(time.Now()).Seconds()),
		Secure:   true,
		HttpOnly: true,
		Domain:   domain,
		SameSite: http.SameSiteDefaultMode,
		Path:     "/",
	}, nil
}

// CSRFContext checks that a proper CSRF token exists as a cookie.
func (m *tokenManager) CSRFContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		csrfCookie, err := r.Cookie(m.config.CSRFCookieName())
		if err != nil {
			render.Render(w, r, sfdErrors.ErrInvalidRequest(err))
			return
		}

		t, err := m.CheckToken(csrfCookie.Value, CSRF)
		if err != nil {
			render.Render(w, r, sfdErrors.ErrInvalidRequest(err))
			return
		}
		ctx := context.WithValue(r.Context(), CSRFCtxKey, &CSRFToken{t})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *tokenManager) CSRFFromCookie(r *http.Request) (*Token, error) {
	cookie, err := r.Cookie(m.config.CSRFCookieName())
	if err != nil {
		return nil, err
	}
	t, err := m.CheckToken(cookie.Value, CSRF)
	if err != nil {
		return nil, err
	}
	return t, nil
}

var (
	// ErrInvalidToken token is invalid
	ErrInvalidToken = errors.New("token is invalid")
	// ErrExpiredToken token is expired
	ErrExpiredToken           = errors.New("token is expired")
	baseTime        time.Time = time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
)

// tokenWithTimestamp the timestamp is the number of seconds since baseTime
func (m *tokenManager) tokenWithTimestamp(
	id *uuid.UUID,
	pwHash string,
	lastLogin time.Time,
	timestamp time.Duration,
	kind Kind,
) string {
	strTimestamp := strconv.FormatFloat(timestamp.Seconds(), 'f', 6, 64)
	b64timestamp := base64.StdEncoding.EncodeToString([]byte(strTimestamp))

	hashString := m.saltedHMAC(
		m.config.DefaultTokenSalt(),
		hashValue(id, pwHash, lastLogin, timestamp, kind),
		m.config.SecretKey(),
	).Sum(nil)
	return fmt.Sprintf("%s-%s", b64timestamp, hex.EncodeToString(hashString))
}

func hashValue(
	id *uuid.UUID,
	pwHash string,
	lastLogin time.Time,
	timestamp time.Duration,
	kind Kind,
) string {
	var loginTimestamp string
	if lastLogin.IsZero() {
		loginTimestamp = ""
	} else {
		loginTimestamp = lastLogin.Truncate(time.Second).String()
	}
	encodedHash := fmt.Sprintf(
		"%s-%s-%s-%s-%s",
		id.String(),
		pwHash,
		loginTimestamp,
		timestamp.String(),
		kind.String(),
	)
	return encodedHash
}

func (m *tokenManager) saltedHMAC(keySalt, value, secret string) hash.Hash {
	if secret == "" {
		secret = m.config.SecretKey()
	}
	if keySalt == "" {
		keySalt = m.config.DefaultTokenSalt()
	}
	hasher := m.hashFunc()
	hasher.Write([]byte(keySalt + secret))
	key := hasher.Sum(nil)

	mac := hmac.New(m.hashFunc, key)
	mac.Write([]byte(value))
	return mac
}
