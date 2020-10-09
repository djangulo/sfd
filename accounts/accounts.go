//Package accounts handles the creation and management of User
// accounts, and provides authentication handlers for User sessions.
package accounts

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/djangulo/sfd/config"
	"github.com/djangulo/sfd/crypto/password"
	"github.com/djangulo/sfd/crypto/session"
	"github.com/djangulo/sfd/crypto/token"
	"github.com/djangulo/sfd/db"
	"github.com/djangulo/sfd/db/models"
	"github.com/djangulo/sfd/db/validators"
	sfdErrors "github.com/djangulo/sfd/errors"
	"github.com/djangulo/sfd/mail"
	"github.com/djangulo/sfd/storage"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/gofrs/uuid"
)

/*
API endpoints
	GET:
		/users/{username} profile, self, or if public
		/check-username	username available
		/register/token email verification token check
		/password/token verify email token

	POST:
		/login
		/logout
		/register	handle registration form
		/password/reset reset password
		/password/confirm password reset form after verification
		/password/change change password

*/

// template names, should match their handlers when appropriate
const (
	notifyAdminNewUserTXT  = "notifyAdminNewUserTXT"
	regVerifyUserHTMLEmail = "regVerifyUserHTMLEmail"
	regVerifyUserTXTEmail  = "regVerifyUserTXTEmail"

	// Password reset templates
	enPassResetTXTEmail  = "enPassResetTXTEmail"
	enPassResetHTMLEmail = "enPassResetHTMLEmail"
	esPassResetTXTEmail  = "esPassResetTXTEmail"
	esPassResetHTMLEmail = "esPassResetHTMLEmail"
)

func init() {

	for _, setup := range []struct {
		isText       bool
		templateName string
		assets       []string
		funcMap      map[string]interface{}
	}{
		{true, notifyAdminNewUserTXT, []string{"accounts/templates/email/notify-admin-new-user.txt"}, nil},
		{true, regVerifyUserTXTEmail, []string{"accounts/templates/email/es-DO/registration-verify-user.txt"}, nil},
		{false, regVerifyUserHTMLEmail, []string{"accounts/templates/email/es-DO/registration-verify-user.html"}, nil},
		{false, esPassResetHTMLEmail, []string{"accounts/templates/email/es-DO/password-reset.html"}, nil},
		{true, esPassResetTXTEmail, []string{"accounts/templates/email/es-DO/password-reset.txt"}, nil},
		{false, enPassResetHTMLEmail, []string{"accounts/templates/email/en-US/password-reset.html"}, nil},
		{true, enPassResetTXTEmail, []string{"accounts/templates/email/en-US/password-reset.txt"}, nil},
	} {
		if setup.isText {
			mail.RegisterTextTemplate(
				setup.templateName,
				Asset,
				setup.funcMap,
				setup.assets...,
			)
			continue
		}
		mail.RegisterHTMLTemplate(
			setup.templateName,
			Asset,
			setup.funcMap,
			setup.assets...,
		)
	}
}

type assetFn func(string) ([]byte, error)

func getAssets(asset assetFn, assetPaths ...string) [][]byte {
	assets := make([][]byte, 0)
	for _, path := range assetPaths {
		b, err := asset(path)
		if err != nil {
			panic(err)
		}
		assets = append(assets, b)
	}
	return assets
}

// Server exported struct glues package together.
type Server struct {
	store          db.AccountStorer
	mail           mail.Mailer
	tokenManager   token.Manager
	sessionManager session.Manager
	config         config.Configurer
	storage        storage.Driver
}

// NewServer returns a new accounts.Server instance.
func NewServer(
	storer db.AccountStorer,
	mailer mail.Mailer,
	config config.Configurer,
	storage storage.Driver,
	tm token.Manager,
	sm session.Manager) (*Server, error) {

	server := Server{
		store:          storer,
		mail:           mailer,
		config:         config,
		tokenManager:   tm,
		sessionManager: sm,
		storage:        storage,
	}

	return &server, nil
}

func (s *Server) ComparePassword(userID *uuid.UUID, pass string) (bool, error) {
	hash, err := s.store.GetPasswordHash(userID)
	if err != nil {
		return false, err
	}
	match, err := password.Compare(pass, hash)
	if err != nil {
		return false, err
	}
	return match, nil
}

// NotifyAdminsNewUser sends an email to every site admin when a new user account
// is created.
func (s *Server) NotifyAdminsNewUser(user *models.User, userLink string) error {

	data := struct {
		User      *models.User
		AdminLink string
	}{user, userLink}

	recipients := make([]mail.Recipient, 0)
	for _, a := range s.config.SiteAdmins() {
		recipients = append(recipients, mail.Recipient{Name: a, Address: a})
	}

	err := s.mail.SendTemplate(
		"text/plain",
		notifyAdminNewUserTXT,
		data,
		fmt.Sprintf("Nuevo usuario registrado: %s", user.Email),
		s.config.DefaultFromEmail(),
		recipients...,
	)
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) sendVerificationEmail(user *models.User) error {

	expiry := time.Now().Add(s.config.AccountConfirmationEmailExpiry())
	token, err := s.tokenManager.NewToken(
		user.LastLogin.Time,
		&user.ID,
		user.PasswordHash,
		token.Registration,
		&expiry)
	if err != nil {
		return err
	}

	verifyURL, err := url.Parse(s.config.PublicURL() + s.config.EmailVerificationEndpoint())
	if err != nil {
		return err
	}
	params := url.Values{}
	params.Add("token", token.Digest)
	verifyURL.RawQuery = params.Encode()

	data := struct {
		User             *models.User
		SiteHost         string
		VerificationLink string
		UnsubscribeLink  string
	}{
		User:             user,
		SiteHost:         s.config.SiteHost(),
		VerificationLink: verifyURL.String(),
	}

	// Set both Name and Address to the user's email so as not to leak the user's name
	recipients := []mail.Recipient{
		{Name: user.Email, Address: user.Email},
	}

	err = s.mail.SendMultipartTemplate(
		regVerifyUserTXTEmail,
		data,
		regVerifyUserHTMLEmail,
		data,
		"Confirmación de correo electrónico",
		s.config.DefaultFromEmail(),
		recipients...,
	)
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) sendPassResetEmail(user *models.User) error {

	var htmlTpl, txtTpl, subject string
	switch user.Preferences.Language {
	case "es", "es-DO":
		txtTpl = esPassResetTXTEmail
		htmlTpl = esPassResetHTMLEmail
		subject = "Reinicio de contraseña"
	default:
		txtTpl = enPassResetTXTEmail
		htmlTpl = enPassResetHTMLEmail
		subject = "Password reset"
	}

	expiry := time.Now().Add(s.config.PassResetTokenExpiry())
	token, err := s.tokenManager.NewToken(
		user.LastLogin.Time,
		&user.ID,
		user.PasswordHash,
		token.PasswordReset,
		&expiry)
	if err != nil {
		return err
	}

	resetURL, err := url.Parse(s.config.PublicURL() + s.config.PasswordResetEndpoint())
	if err != nil {
		return err
	}
	params := url.Values{}
	params.Add("token", token.Digest)
	resetURL.RawQuery = params.Encode()

	host, err := url.Parse(s.config.PublicURL())
	if err != nil {
		return err
	}
	// unsubLink := fmt.Sprintf("%s/accounts/registration-unsub?email=%s", cfg.SiteHost(), user.Email)

	data := struct {
		User            *models.User
		SiteHost        template.URL
		PassResetLink   template.URL
		UnsubscribeLink template.URL
	}{
		User:          user,
		SiteHost:      template.URL(host.String()),
		PassResetLink: template.URL(resetURL.String()),
	}

	// Set both Name and Address to the user's email so as not to leak the user's name
	recipients := []mail.Recipient{
		{Name: user.Email, Address: user.Email},
	}
	err = s.mail.SendMultipartTemplate(
		txtTpl,
		data,
		htmlTpl,
		data,
		subject,
		s.config.DefaultFromEmail(),
		recipients...,
	)
	if err != nil {
		return err
	}
	return nil
}

type PassResetInitRequest struct {
	UsernameOrEmail string `json:"username"`
}

func (prr *PassResetInitRequest) Bind(r *http.Request) error {
	e := validators.NewErrors()
	if err := validators.NotEmpty(prr.UsernameOrEmail); err != nil {
		e.Add("username", err)
	}
	if e.Len() > 0 {
		return e
	}
	return nil
}

type CallBackResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func (cbr *CallBackResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func NewCallBackResponse(status, msg string) *CallBackResponse {
	return &CallBackResponse{Status: status, Message: msg}
}

func (s *Server) PasswordResetInit(w http.ResponseWriter, r *http.Request) {
	data := &PassResetInitRequest{}
	if err := render.Bind(r, data); err != nil {
		if e, ok := err.(*validators.Errors); ok {
			render.Render(w, r, e)
			return
		}
		render.Render(w, r, sfdErrors.ErrInvalidRequest(err))
		return
	}

	user, _ := s.store.UserByUsernameOrEmail(data.UsernameOrEmail)
	if user != nil {
		if err := s.sendPassResetEmail(user); err != nil {
			log.Println(err)
		}
	}

	if err := render.Render(w, r, NewCallBackResponse("OK", "reset email sent")); err != nil {
		sfdErrors.ErrRender(err)
		return
	}
}

type ValidationTokenResponse struct {
	Token  string    `json:"token"`
	Expiry time.Time `json:"expiry"`
}

func (vtr *ValidationTokenResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (s *Server) CheckPassResetToken(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	tokenStr := q.Get("token")
	e := validators.NewErrors()
	if tokenStr == "" {
		e.Add("token", fmt.Errorf("token is invalid"))
		if err := render.Render(w, r, e); err != nil {
			sfdErrors.ErrRender(err)
			return
		}
		return
	}
	t, err := s.tokenManager.CheckToken(tokenStr, token.PasswordReset)
	if err != nil {
		e.Add("token", fmt.Errorf("token is invalid"))
		if err := render.Render(w, r, e); err != nil {
			sfdErrors.ErrRender(err)
			return
		}
		return
	}

	// original registration token was used: drop it
	if err := s.tokenManager.DropToken(t.Digest); err != nil {
		log.Println(err)
		// if err := render.Render(w, r, sfdErrors.ErrInternalServerError(err)); err != nil {
		// 	sfdErrors.ErrRender(err)
		// 	return
		// }
		// return
	}

	csrfCookie, err := s.tokenManager.NewCSRFCookie(t.UserID, "")
	if err != nil {
		render.Render(w, r, sfdErrors.ErrInternalServerError(err))
		return
	}
	http.SetCookie(w, csrfCookie)

	if err := render.Render(w, r, NewCallBackResponse("OK", "token is valid")); err != nil {
		render.Render(w, r, sfdErrors.ErrRender(err))
		return
	}
}

type PassResetRequest struct {
	Password       string `json:"password"`
	RepeatPassword string `json:"repeat_password"`
}

var ErrPasswordsDoNotMatch = errors.New("passwords do not match")

func (psr *PassResetRequest) Bind(r *http.Request) error {
	var e = validators.NewErrors()
	if psr.Password != psr.RepeatPassword {
		e.Add("password", ErrPasswordsDoNotMatch)
	}
	if err := validators.Password(psr.Password); err != nil {
		e.Add("password", err...)
	}
	if e.Len() > 0 {
		return e
	}
	return nil
}

func (s *Server) PasswordResetConfirm(w http.ResponseWriter, r *http.Request) {
	data := &PassResetRequest{}
	if err := render.Bind(r, data); err != nil {
		if e, ok := err.(*validators.Errors); ok {
			render.Render(w, r, e)
			return
		}
		render.Render(w, r, sfdErrors.ErrInvalidRequest(err))
		return
	}

	t, err := s.tokenManager.CSRFFromCookie(r)
	if err != nil {
		render.Render(w, r, sfdErrors.ErrInvalidToken(err))
		return
	}

	user, err := s.store.UserByID(t.UserID)
	if err != nil {
		render.Render(w, r, sfdErrors.ErrInvalidToken(err))
		return
	}

	hash, err := password.Hash(data.Password, password.DefaultHashParams)
	if err != nil {
		render.Render(w, r, sfdErrors.ErrInternalServerError(err))
		return
	}

	if err := s.store.ResetPassword(&user.ID, hash); err != nil {
		render.Render(w, r, sfdErrors.ErrInternalServerError(err))
		return
	}

	// unset csrf cookie
	cookie, _ := r.Cookie(s.config.CSRFCookieName())
	cookie.MaxAge = -1
	http.SetCookie(w, cookie)
	if err := s.tokenManager.DropToken(t.Digest); err != nil {
		log.Println(err)
	}
	res := NewCallBackResponse("OK", "password reset successfully")
	if err := render.Render(w, r, res); err != nil {
		sfdErrors.ErrRender(err)
		return
	}
}

// RegistrationFormData what the application expects from the registration
// form.
type RegistrationRequest struct {
	Username       string `json:"username"`
	FullName       string `json:"full_name"`
	Email          string `json:"email"`
	Password       string `json:"password"`
	RepeatPassword string `json:"repeat_password"`
	AcceptedTOS    bool   `json:"accept_tos"`
}

func (rr *RegistrationRequest) Bind(r *http.Request) error {
	e := validators.NewErrors()

	if rr.Password != rr.RepeatPassword {
		e.Add("password", fmt.Errorf("passwords do not match"))
	}
	if err := validators.Password(rr.Password); err != nil {
		e.Add("password", err...)
	}
	if err := validators.NotEmpty(rr.Username); err != nil {
		e.Add("username", err)
	}
	if err := validators.MaxLength(rr.Username, validators.ByteLength); err != nil {
		e.Add("username", err)
	}
	if err := validators.NotEmpty(rr.Email); err != nil {
		e.Add("email", err)
	}
	if err := validators.MaxLength(rr.Email, validators.ByteLength); err != nil {
		e.Add("email", err)
	}
	if err := validators.NotEmpty(rr.FullName); err != nil {
		e.Add("full_name", err)
	}
	if !rr.AcceptedTOS {
		e.Add("accept_tos", fmt.Errorf("you must accept the terms of service"))
	}
	if len(e.Values) > 0 {
		return e
	}
	return nil
}

var errNotAvailable = errors.New("not available")

// RegisterUser
//   - Create a data object.
//   - Populate it with form data via forms.Bind.
//   - Validate the form as defined by RegistrationForm.Validate method.
//   - Perform additional db layer validation (uniqueness of username, email).
//   - Create user if all validation passed.
//   - Send notification emails (admin, user).
//   - Serve notification view.
func (s *Server) RegisterUser(w http.ResponseWriter, r *http.Request) {

	data := &RegistrationRequest{}
	if err := render.Bind(r, data); err != nil {
		if e, ok := err.(*validators.Errors); ok {
			render.Render(w, r, e)
			return
		}
		render.Render(w, r, sfdErrors.ErrInvalidRequest(err))
		return
	}

	e := validators.NewErrors()

	if user, _ := s.store.UserByUsernameOrEmail(data.Username); user != nil {
		e.Add("username", errNotAvailable)
	}
	if user, _ := s.store.UserByUsernameOrEmail(data.Email); user != nil {
		e.Add("email", errNotAvailable)
	}

	if e.Len() > 0 {
		render.Render(w, r, e)
		return
	}

	hash, err := password.Hash(data.Password, password.DefaultHashParams)
	if err != nil {
		render.Render(w, r, sfdErrors.ErrInternalServerError(err))
		return
	}

	user, err := models.NewUser(
		data.Username,
		data.Email,
		hash,
		"en-US",
		false,
	)
	if err != nil {
		render.Render(w, r, sfdErrors.ErrInternalServerError(err))
		return
	}
	if err := s.store.CreateUser(user); err != nil {
		render.Render(w, r, sfdErrors.ErrInternalServerError(err))
		return
	}
	if err := s.NotifyAdminsNewUser(user, ""); err != nil {
		log.Println(err)
	}
	if err := s.sendVerificationEmail(user); err != nil {
		log.Println(err)
	}

	res := NewCallBackResponse("OK", "registration successful")
	if err := render.Render(w, r, res); err != nil {
		render.Render(w, r, sfdErrors.ErrRender(err))
		return
	}
}

func (s *Server) VerifyEmailToken(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	t := q.Get("token")
	if t == "" {
		if err := render.Render(w, r, sfdErrors.ErrInvalidToken(fmt.Errorf("no token"))); err != nil {
			render.Render(w, r, sfdErrors.ErrRender(err))
			return
		}
		return
	}
	_, err := s.tokenManager.CheckToken(t, token.Registration)
	if err != nil {
		if err := render.Render(w, r, sfdErrors.ErrInvalidToken(err)); err != nil {
			render.Render(w, r, sfdErrors.ErrRender(err))
			return
		}
	}

	// token validation passed, update as true, drop and return
	if err := s.tokenManager.DropToken(t); err != nil {
		log.Println(err)
	}

	res := NewCallBackResponse("OK", "successfully validated")
	if err := render.Render(w, r, res); err != nil {
		render.Render(w, r, sfdErrors.ErrRender(err))
		return
	}
}

type key int

const (
	UserCtxKey key = iota + 1
)

func (s *Server) UserContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var user *models.User
		var err error

		ctx := r.Context()

		if userID := chi.URLParam(r, "userID"); userID != "" {
			var uid uuid.UUID
			uid, err = uuid.FromString(userID)
			if err != nil {
				render.Render(w, r, sfdErrors.ErrInvalidRequest(err))
				return
			}
			user, err = s.store.UserByID(&uid)
		} else if username := chi.URLParam(r, "username"); username != "" {
			user, err = s.store.UserByUsernameOrEmail(username)
		} else {
			render.Render(w, r, sfdErrors.ErrNotFound)
			return
		}
		if err != nil {
			render.Render(w, r, sfdErrors.ErrNotFound)
			return
		}
		if user.ProfilePublic {
			ctx = context.WithValue(ctx, UserCtxKey, user)
		}
		if ses, err := session.FromContext(ctx); err == nil {
			currentUser, err := ses.GetUser()
			if err != nil {
				render.Render(w, r, sfdErrors.ErrInternalServerError(err))
				return
			}
			if currentUser.IsAdmin || currentUser.ID == user.ID {
				ctx = context.WithValue(ctx, UserCtxKey, user)
			}
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type UserProfileResponse struct {
	*models.User
}

func (upr *UserProfileResponse) Render(w http.ResponseWriter, r *http.Request) error {
	// prevent leaking the hash
	upr.User.PasswordHash = ""
	return nil
}

func GetUser(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(UserCtxKey).(*models.User)
	if !ok {
		render.Render(w, r, sfdErrors.ErrInvalidRequest(fmt.Errorf("profile is private")))
		return
	}

	if pusher, ok := w.(http.Pusher); ok {
		img := user.Picture
		if img.Path.Valid {
			if img.Path.String != "" {
				contentType := "image/" + strings.ReplaceAll(img.FileExt.String, ".", "")
				options := &http.PushOptions{
					Header: http.Header{
						"Content-Type":    []string{contentType},
						"Accept-Encoding": r.Header["Accept-Encoding"],
					},
				}
				if err := pusher.Push(img.Path.String, options); err != nil {
					log.Println(err)
				}
			}
		}
	}

	res := &UserProfileResponse{User: user}
	if err := render.Render(w, r, res); err != nil {
		render.Render(w, r, sfdErrors.ErrRender(err))
	}
}

type LoginRequest struct {
	Username string `form:"username"`
	Password string `form:"password"`
}

func (lr *LoginRequest) Bind(r *http.Request) error {
	e := validators.NewErrors()

	if err := validators.NotEmpty(lr.Password); err != nil {
		e.Add("password", err)
	}
	if err := validators.NotEmpty(lr.Username); err != nil {
		e.Add("username", err)
	}
	if e.Len() > 0 {
		return e
	}
	return nil
}

type LoginResponse struct {
	*models.User
}

// NewLoginResponse wraps *models.User.
func NewLoginResponse(user *models.User) *LoginResponse {
	return &LoginResponse{User: user}
}

// Render renders LoginResponse. Cuts down the user to minimal fields, as the rest
// of the fields are dependent on whether the user profile is public or not.
func (lr *LoginResponse) Render(w http.ResponseWriter, r *http.Request) error {
	var p = &models.UserPreferences{}
	if lr.User.Preferences != nil {
		p = lr.Preferences
	}
	var pic = &models.ProfilePicture{}
	if lr.User.Picture != nil {
		pic = lr.Picture
	}
	lr.User = &models.User{
		DBObj:       &models.DBObj{ID: lr.User.ID},
		Username:    lr.User.Username,
		Preferences: p,
		IsAdmin:     lr.User.IsAdmin,
		Active:      lr.User.Active,
		Picture:     pic,
	}
	return nil
}

const stateCookieName = "sfd-user-state-restore"

func (s *Server) LoginUser(w http.ResponseWriter, r *http.Request) {
	data := &LoginRequest{}
	if err := render.Bind(r, data); err != nil {
		if e, ok := err.(*validators.Errors); ok {
			render.Render(w, r, e)
			return
		}
		render.Render(w, r, sfdErrors.ErrInvalidRequest(err))
		return
	}

	ses, _ := session.FromContext(r.Context())
	if ses != nil {
		render.Render(w, r, sfdErrors.ErrAlreadyAuthenticated)
		return
	}
	user, err := s.store.UserByUsernameOrEmail(data.Username)
	if err != nil {
		render.Render(w, r, sfdErrors.ErrInvalidUsernameOrPassword)
		return
	}

	match, err := s.ComparePassword(&user.ID, data.Password)
	if err != nil {
		render.Render(w, r, sfdErrors.ErrInvalidUsernameOrPassword)
		return
	}

	if !match {
		render.Render(w, r, sfdErrors.ErrInvalidUsernameOrPassword)
		return
	}

	user.UpdatedAt = models.NewNullTime(time.Now())
	user.LastLogin = user.UpdatedAt

	user, err = s.store.LoginUser(user)
	if err != nil {
		render.Render(w, r, sfdErrors.ErrInternalServerError(err))
		return
	}

	values := make(session.Values)
	values.Set(session.UserKey, user)

	ses, err = s.sessionManager.New(values)
	if err != nil {
		render.Render(w, r, sfdErrors.ErrInternalServerError(err))
		return
	}

	cookie, err := s.sessionManager.NewAuthCookie(ses, "")
	if err != nil {
		render.Render(w, r, sfdErrors.ErrInternalServerError(err))
		return
	}
	http.SetCookie(w, cookie)

	t, err := s.tokenManager.NewToken(
		user.LastLogin.Time,
		&user.ID,
		user.PasswordHash,
		token.State,
		&cookie.Expires,
	)
	if err != nil {
		log.Println(err)
	}

	// set a cookie to restore user state
	cookie = &http.Cookie{
		Name:     stateCookieName,
		Value:    t.Digest,
		Expires:  t.Expires,
		MaxAge:   int(t.Expires.Sub(time.Now()).Seconds()),
		HttpOnly: false,
		Secure:   false,
	}
	http.SetCookie(w, cookie)

	res := NewLoginResponse(user)
	if err := render.Render(w, r, res); err != nil {
		render.Render(w, r, sfdErrors.ErrRender(err))
		return
	}
}

func (s *Server) Logout(w http.ResponseWriter, r *http.Request) {

	cookie, err := r.Cookie(s.sessionManager.CookieName())
	if err != nil {
		log.Println(err)
		return
	}

	cookie.MaxAge = -1
	if err := s.sessionManager.Delete(cookie.Value); err != nil {
		render.Render(w, r, sfdErrors.ErrInternalServerError(err))
		return
	}
	http.SetCookie(w, cookie)

	cookie, err = r.Cookie(stateCookieName)
	if err != nil {
		log.Println(err)
		return
	}
	cookie.MaxAge = -1
	http.SetCookie(w, cookie)

	res := NewCallBackResponse("OK", "succesfully logged out")
	if err := render.Render(w, r, res); err != nil {
		render.Render(w, r, sfdErrors.ErrRender(err))
		return
	}
}

type IsLoggedInRequest struct {
	UserID string `json:"user_id"`
}

// Bind implements the render.Binder interface.
func (ilir *IsLoggedInRequest) Bind(r *http.Request) error {
	e := validators.NewErrors()
	if err := validators.NotEmpty(ilir.UserID); err != nil {
		e.Add("user_id", err)
	}
	if e.Len() > 0 {
		return e
	}
	return nil
}

// RestoreUserState a client with the right cookie can restore a user object.
func (s *Server) RestoreUserState(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(stateCookieName)
	if err != nil {
		render.Render(w, r, sfdErrors.ErrInvalidRequest(fmt.Errorf("state cookie not found")))
		return
	}
	t, err := s.tokenManager.CheckToken(cookie.Value, token.State)
	if err != nil {
		render.Render(w, r, sfdErrors.ErrInvalidRequest(fmt.Errorf("state token is invalid")))
		return
	}
	user, err := s.store.UserByID(t.UserID)
	if err != nil {
		render.Render(w, r, sfdErrors.ErrInternalServerError(err))
		return
	}

	res := NewLoginResponse(user)
	if err := render.Render(w, r, res); err != nil {
		render.Render(w, r, sfdErrors.ErrRender(err))
		return
	}
}
