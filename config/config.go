//Package config contains configuration for the project.
// It revolves around the Config struct.
// It initially sets default values, then tries to populate from values
// defined in the config.conf file, then from the environment (if set).
//
// Environment variables are the same keys as defined in this file,
// uppercased and prepended with "SFD", e.g. storage_url env key is
// SFD_STORAGE_URL.
//
// All values are stored as a string, but the Configurer interface, which
// is what dependent packages should use, returns a proper, usable data
// structure; e.g. password_reset_token_expiry sits in file and in the
// singleton as "1h", but the Configurer interface returns a time.Duration
// value.
package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Configuration keys.
const (
	databaseURL                    = "database_url"
	migrationsDir                  = "migrations_dir"
	storageURL                     = "storage_url"
	emailURL                       = "email_url"
	siteAdmins                     = "site_admins"
	defaultFromEmail               = "default_from_email"
	secretKey                      = "secret_key"
	tokenSalt                      = "token_salt"
	siteHost                       = "site_host"
	publicURL                      = "public_url"
	authCookieName                 = "auth_cookie_name"
	smtpHost                       = "smtp_host"
	smtpPort                       = "smtp_port"
	smtpUser                       = "smtp_user"
	smtpPass                       = "smtp_pass"
	authMountPoint                 = "auth_mount_point"
	passwordResetTokenExpiry       = "password_reset_token_expiry"
	accountConfirmationEmailExpiry = "account_confirmation_email_expiry"
	csrfTokenExpiry                = "csrf_token_expiry"
	csrfCookieName                 = "csrf_cookie_name"
	defaultPageSize                = "default_page_size"
	timeZone                       = "time_zone"
	passwordResetEndpoint          = "password_reset_endpoint"
	emailVerificationEndpoint      = "email_verification_endpoint"
	loginEndpoint                  = "login_endpoint"
	logoutEndpoint                 = "logout_endpoint"
	webFrontendMountpoint          = "web_frontend_mountpoint"
)

// Config struct holds config.
type Config struct {
	sync.Mutex
	keys   keyHandler
	Values map[string]string
}

func NewConfig() *Config {
	c := Config{}
	c.Values = make(map[string]string)
	c.keys = make(keyHandler)
	allKeys := []string{
		databaseURL,
		migrationsDir,
		storageURL,
		emailURL,
		siteAdmins,
		defaultFromEmail,
		secretKey,
		tokenSalt,
		siteHost,
		publicURL,
		authCookieName,
		smtpHost,
		smtpPort,
		smtpUser,
		smtpPass,
		authMountPoint,
		passwordResetTokenExpiry,
		accountConfirmationEmailExpiry,
		csrfTokenExpiry,
		defaultPageSize,
		timeZone,
		passwordResetEndpoint,
		emailVerificationEndpoint,
		csrfCookieName,
		loginEndpoint,
		logoutEndpoint,
		webFrontendMountpoint,
	}
	for _, k := range allKeys {
		c.keys[k] = struct{}{}
	}
	return &c
}

func (c *Config) Defaults() {
	if c.Values == nil {
		c.Values = make(map[string]string)
	}
	c.Values[databaseURL] = "memory://irrelevanthost/?prepopulate=true"
	c.Values[migrationsDir] = "db/migrations"
	c.Values[storageURL] = "fs://./assets?accept=.jpg&accept=.png&accept=.jpeg&accept=.svg&root=/tmp/assets"
	c.Values[emailURL] = "console://./"
	c.Values[siteAdmins] = "webmaster@sfd-app.com"
	c.Values[defaultFromEmail] = "noreply@sfd.com"
	c.Values[secretKey] = "changeme"
	c.Values[tokenSalt] = "changeme"
	c.Values[siteHost] = "localhost"
	c.Values[publicURL] = "https://localhost:9000"
	c.Values[authCookieName] = "sfd-session-id"
	c.Values[authMountPoint] = "/accounts"
	c.Values[passwordResetTokenExpiry] = "1h"
	c.Values[accountConfirmationEmailExpiry] = "72h"
	c.Values[csrfTokenExpiry] = "10m"
	c.Values[defaultPageSize] = "10"
	c.Values[timeZone] = "UTC"
	c.Values[passwordResetEndpoint] = "/api/password/token"
	c.Values[emailVerificationEndpoint] = "/api/register/token"
	c.Values[csrfCookieName] = "X-CSRF-Token"
	c.Values[loginEndpoint] = "/api/login"
	c.Values[logoutEndpoint] = "/api/logout"
	c.Values[webFrontendMountpoint] = "/"
}

func (c *Config) FromFile(path string) error {
	fh, err := os.Open(path)
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(fh)
	for scanner.Scan() {
		split := strings.SplitN(scanner.Text(), "=", 2)
		if len(split) == 2 {
			key, value := strings.TrimSpace(split[0]), strings.TrimSpace(split[1])
			if handler.valid(key) {
				c.Lock()
				if c.Values == nil {
					c.Values = make(map[string]string)
				}
				c.Values[key] = value
				c.Unlock()
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

// FromEnv overrides all values if a key is found.
func (c *Config) FromEnv() {
	for key := range c.keys {
		envKey := "SFD_" + strings.ToUpper(key)
		c.Lock()

		if v := os.Getenv(envKey); c.keys.valid(key) && v != "" {
			c.Values[key] = v
		}
		c.Unlock()
	}
}

var (
	// Cnf configuration singleton.
	cnf *Config
)

func Get() Configurer {
	cnf = NewConfig()
	cnf.Defaults()
	gopath := os.Getenv("GOPATH")
	path := filepath.Join(gopath, "src", "github.com", "djangulo", "sfd", "config", "config.conf")
	if err := cnf.FromFile(path); err != nil {
		panic(err)
	}
	cnf.FromEnv()
	return cnf
}

type Configurer interface {
	// DatabaseURL returns the URL for the db connection.
	DatabaseURL() string
	// MigrationsDir return the directory that holds migration files.
	MigrationsDir() string
	// StorageURL default storage engine.
	StorageURL() string
	// EmailURL config url.
	EmailURL() string
	// SiteAdminis admins for the site, these will be notified on items
	// that require admin attention. This value can be modified at runtime.
	SiteAdmins() []string
	// AddAdmin adds adminEmail to the admis string and returns the modified array.
	AddAdmin(adminEmails ...string) []string
	DefaultFromEmail() string
	DefaultTokenSalt() string
	// SecretKey returns the config secret, used for various implementations
	// throughout application.
	// If needed, generate one with
	//    head /dev/urandom | tr -dc 'a-zA-Z0-9' | head -c 64
	SecretKey() string
	// SiteHost host for application. This is decoupled from actual host, and it must
	// configured to match. It's used by emails and notifications to create external
	// links to the site.
	SiteHost() string
	PublicURL() string
	// SMTPSettings returns a [4]string with the host, port, user, and password.
	// It will panic if it's not configured.
	SMTPSettings() [4]string
	AuthCookieName() string
	// AuthMountPoint is where the authentication module is mounted. It's used in many
	// internal URLs.
	AuthMountPoint() string
	// PassResetTokenExpiry returns a time.Duration for a password reset token
	// max age.
	PassResetTokenExpiry() time.Duration
	// CSRFTokenExpiry returns a time.Duration for a csrf token max age.
	CSRFTokenExpiry() time.Duration
	// CSRFCookieName returns the config csrf cookie name.
	CSRFCookieName() string
	// AccountConfirmationEmailExpiry returns a time.Duration for an account
	// confirmation token max age.
	AccountConfirmationEmailExpiry() time.Duration
	// PageSize returns the default page size. File key: default_page_size,
	// env key: SFD_DEFAULT_PAGE_SIZE
	PageSize() int
	TimeZone() *time.Location
	PasswordResetEndpoint() string
	EmailVerificationEndpoint() string
	LoginEndpoint() string
	LogoutEndpoint() string
	WebFrontendMountpoint() string
}

type keyHandler map[string]struct{}

var handler = make(keyHandler)

func (h keyHandler) valid(key string) bool {
	_, ok := h[key]
	return ok
}

func (c *Config) DatabaseURL() string {
	c.Lock()
	defer c.Unlock()

	return c.Values[databaseURL]
}
func (c *Config) LoginEndpoint() string {
	c.Lock()
	defer c.Unlock()

	return c.Values[loginEndpoint]
}
func (c *Config) LogoutEndpoint() string {
	c.Lock()
	defer c.Unlock()

	return c.Values[logoutEndpoint]
}

func (c *Config) StorageURL() string {
	c.Lock()
	defer c.Unlock()

	return c.Values[storageURL]
}

func (c *Config) EmailURL() string {
	c.Lock()
	defer c.Unlock()

	return c.Values[emailURL]
}

func (c *Config) SiteAdmins() []string {
	c.Lock()
	defer c.Unlock()
	set := make(map[string]struct{})
	admins := c.Values[siteAdmins]
	for _, admin := range strings.Split(admins, ",") {
		if _, ok := set[admin]; !ok {
			set[admin] = struct{}{}
		}
	}
	aadmins := make([]string, 0)
	for k := range set {
		aadmins = append(aadmins, k)
	}
	return aadmins
}

func (c *Config) AddAdmin(adminEmails ...string) []string {
	c.Lock()
	defer c.Unlock()

	admins := c.SiteAdmins()
	set := make(map[string]struct{})
	for _, admin := range adminEmails {
		if _, ok := set[admin]; !ok {
			set[admin] = struct{}{}
		}
	}
	for k := range set {
		admins = append(admins, k)
	}
	c.Values[siteAdmins] = strings.Join(admins, ",")
	return admins

}
func (c *Config) DefaultFromEmail() string {
	c.Lock()
	defer c.Unlock()

	return c.Values[defaultFromEmail]
}
func (c *Config) DefaultTokenSalt() string {
	c.Lock()
	defer c.Unlock()

	return c.Values[tokenSalt]
}

// SecretKey returns the config secret, used for various implementations
// throughout the application.
//
// If needed, generate one with
//    head /dev/urandom | tr -dc 'a-zA-Z0-9' | head -c 64
func (c *Config) SecretKey() string {
	c.Lock()
	defer c.Unlock()

	return c.Values[secretKey]
}

// SiteHost host for application. This is decoupled from actual host, and it must
// configured to match. It's used with auth cookies.
func (c *Config) SiteHost() string {
	c.Lock()
	defer c.Unlock()

	return c.Values[siteHost]
}

// PublicURL full application URL host for application. This is decoupled from actual host, and it must
// configured to match.
func (c *Config) PublicURL() string {
	c.Lock()
	defer c.Unlock()

	return c.Values[publicURL]
}

func (c *Config) SMTPSettings() [4]string {
	c.Lock()
	defer c.Unlock()

	missing := make([]string, 0)
	host, ok := c.Values[smtpHost]
	if !ok {
		missing = append(missing, "host")
	}
	port, ok := c.Values[smtpPort]
	if !ok {
		missing = append(missing, "port")
	}
	user, ok := c.Values[smtpUser]
	if !ok {
		missing = append(missing, "user")
	}
	pass, ok := c.Values[smtpPass]
	if !ok {
		missing = append(missing, "pass")
	}

	if len(missing) > 0 {
		panicF("SMTP improperly configured, missing %s", strings.Join(missing, ", "))
	}
	return [...]string{host, port, user, pass}
}

func (c *Config) WebFrontendMountpoint() string {
	c.Lock()
	defer c.Unlock()

	return c.Values[webFrontendMountpoint]
}

func (c *Config) AuthCookieName() string {
	c.Lock()
	defer c.Unlock()

	return c.Values[authCookieName]
}

func (c *Config) AuthMountPoint() string {
	c.Lock()
	defer c.Unlock()

	return c.Values[authMountPoint]
}

// PassResetTokenExpiry returns a time.Duration for a password reset token max
// age.
func (c *Config) PassResetTokenExpiry() time.Duration {
	c.Lock()
	defer c.Unlock()

	str := c.Values[passwordResetTokenExpiry]
	d, err := time.ParseDuration(str)
	if err != nil {
		panicF(`pkg config: Improperly configured: %s: should be able to be parsed by time.ParseDuration
see https://pkg.go.dev/time?tab=doc#ParseDuration`, passwordResetTokenExpiry)
	}
	return d
}

// CSRFTokenExpiry returns a time.Duration for a csrf token max age.
func (c *Config) CSRFTokenExpiry() time.Duration {
	c.Lock()
	defer c.Unlock()

	str := c.Values[csrfTokenExpiry]
	d, err := time.ParseDuration(str)
	if err != nil {
		panicF(`pkg config: Improperly configured: %s: should be able to be parsed by time.ParseDuration
see https://pkg.go.dev/time?tab=doc#ParseDuration`, csrfTokenExpiry)
	}
	return d
}

// CSRFCookieName the csrf cookie name.
func (c *Config) CSRFCookieName() string {
	c.Lock()
	defer c.Unlock()

	return c.Values[csrfCookieName]
}

// AccountConfirmationEmailExpiry returns a time.Duration for an account
// confirmation token max age.
func (c *Config) AccountConfirmationEmailExpiry() time.Duration {
	c.Lock()
	defer c.Unlock()

	str := c.Values[accountConfirmationEmailExpiry]
	d, err := time.ParseDuration(str)
	if err != nil {
		panicF(`pkg config: Improperly configured: %s: should be able to be parsed by time.ParseDuration
see https://pkg.go.dev/time?tab=doc#ParseDuration`, accountConfirmationEmailExpiry)
	}
	return d
}

// MigrationsDir.
func (c *Config) MigrationsDir() string {
	c.Lock()
	defer c.Unlock()

	return c.Values[migrationsDir]
}

func (c *Config) PageSize() int {
	c.Lock()
	defer c.Unlock()

	size, err := strconv.Atoi(c.Values[defaultPageSize])
	if err != nil {
		panicF(`pkg config: Improperly configured: %s: should be able to be parsed by strconv.Atoi
see https://pkg.go.dev/strconv?tab=doc#Atoi`, defaultPageSize)
	}
	return size
}

func (c *Config) TimeZone() *time.Location {
	tz, err := time.LoadLocation(c.Values[timeZone])
	if err != nil {
		panicF("pkg config: Improperly configured: %v", err)
	}
	return tz
}

func (c *Config) PasswordResetEndpoint() string {
	c.Lock()
	defer c.Unlock()

	return c.Values[passwordResetEndpoint]

}
func (c *Config) EmailVerificationEndpoint() string {
	c.Lock()
	defer c.Unlock()

	return c.Values[emailVerificationEndpoint]
}

// panicF like fmt.Sprintf but panicky.
func panicF(format string, a ...interface{}) {
	panic(fmt.Sprintf(format, a...))
}
