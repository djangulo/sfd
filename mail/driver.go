package mail

import (
	"fmt"
	htmlTemplate "html/template"
	"io"
	nurl "net/url"
	"sync"
	txtTemplate "text/template"
)

type Recipient struct {
	Name    string
	Address string
}

// Normalize sets the Name to the email address
// if the Name field is zero value ("").
func (r *Recipient) Normalize() {
	if r.Name == "" {
		r.Name = r.Address
	}
}

type AssetFn func(string) ([]byte, error)

type Mailer interface {
	// SendMail sends an email with the parameters passed to the recipients passed.
	SendMail(contentType, subject, body, from string, recipients ...Recipient) error
	// SendMultipart email sends an email with a text body and an html body.
	SendMultipart(subject, txtBody, htmlBody, from string, recipients ...Recipient) error
	// SendTemplate refer to a template by name as registered with sfd-app/templates
	SendTemplate(contentType, tplName string, data interface{}, subject, from string, recipients ...Recipient) error
	// SendMultipartTemplate
	SendMultipartTemplate(txtTpl string,
		txtData interface{},
		htmlTpl string,
		htmlData interface{},
		subject, from string, recipients ...Recipient) error
}

func getAssets(asset AssetFn, assetPaths ...string) [][]byte {
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

func RegisterHTMLTemplate(tplName string, asset AssetFn, funcs htmlTemplate.FuncMap, tplPaths ...string) {
	templatesMu.Lock()
	defer templatesMu.Unlock()

	// register URL, absolute URL and UserLoggedIn globally as it will be used by every
	// template

	if _, dup := templates[tplName]; dup {
		panic(fmt.Sprintf("mail: Register called twice for template %s", tplName))
	}

	t, err := register(tplName, asset, funcs, tplPaths...)
	if err != nil {
		panic(err)
	}

	templates[tplName] = t
}

// RegisterTxt globally registers a text/template for later usage.
func RegisterTextTemplate(tplName string, asset AssetFn, funcs txtTemplate.FuncMap, tplPaths ...string) {
	templatesMu.Lock()
	defer templatesMu.Unlock()

	if _, dup := txtTemplates[tplName]; dup {
		panic(fmt.Sprintf("mail: RegisterTxt called twice for template %s", tplName))
	}

	t, err := registerTxt(tplName, asset, funcs, tplPaths...)
	if err != nil {
		panic(err)
	}

	txtTemplates[tplName] = t
}

type Driver interface {
	// Open creates a connection to the driver.
	Open(url string) (Driver, error)
	// Close closes the driver connection.
	Close() error
	Mailer
}

var (
	driversMu sync.RWMutex
	drivers   map[string]Driver
)

var (
	templatesMu  sync.Mutex
	templates    map[string]*htmlTemplate.Template
	txtTemplates map[string]*txtTemplate.Template
)

func init() {
	templates = make(map[string]*htmlTemplate.Template)
	txtTemplates = make(map[string]*txtTemplate.Template)
	drivers = make(map[string]Driver)
}

// Open returns a new driver instance.
func Open(url string) (Driver, error) {
	u, err := nurl.Parse(url)
	if err != nil {
		return nil, err
	}

	if u.Scheme == "" {
		return nil, fmt.Errorf("mail : invalid URL scheme")
	}

	driversMu.RLock()
	d, ok := drivers[u.Scheme]
	driversMu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("mail driver: unknown driver %v (forgotten import?)", u.Scheme)
	}

	return d.Open(url)
}

// Register globally registers a driver.
func Register(name string, driver Driver) {
	driversMu.Lock()
	defer driversMu.Unlock()

	if driver == nil {
		panic("mail: Register driver is nil")
	}
	if _, dup := drivers[name]; dup {
		panic("mail: Register called twice for driver " + name)
	}
	drivers[name] = driver
}

// register reads assets into a template sequentially.
func register(templateName string, asset AssetFn, funcs htmlTemplate.FuncMap, assetPaths ...string) (*htmlTemplate.Template, error) {
	assets := getAssets(asset, assetPaths...)
	t, err := htmlTemplate.New(templateName).Funcs(funcs).Parse(string(assets[0]))
	if err != nil {
		return nil, err
	}

	for _, asset := range assets[1:] {
		t, err = t.Parse(string(asset))
		if err != nil {
			return nil, err
		}
	}

	return t, nil
}

// register reads assets into a template sequentially.
func registerTxt(templateName string, asset AssetFn, funcs txtTemplate.FuncMap, assetPaths ...string) (*txtTemplate.Template, error) {
	assets := getAssets(asset, assetPaths...)
	t, err := txtTemplate.New(templateName).Funcs(funcs).Parse(string(assets[0]))
	if err != nil {
		return nil, err
	}

	for _, asset := range assets[1:] {
		t, err = t.Parse(string(asset))
		if err != nil {
			return nil, err
		}
	}

	return t, nil
}

func Execute(tplName string, w io.Writer, data interface{}) error {

	var isHTML, isTxt bool
	_, isHTML = templates[tplName]
	_, isTxt = txtTemplates[tplName]
	if isHTML {
		if err := templates[tplName].Execute(w, data); err != nil {
			return err
		}
		return nil
	}
	if isTxt {
		if err := txtTemplates[tplName].Execute(w, data); err != nil {
			return err
		}
		return nil
	}
	return fmt.Errorf("mail: not found: %s", tplName)
}
