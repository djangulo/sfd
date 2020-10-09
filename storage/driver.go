package storage

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	nurl "net/url"
	"sync"
)

var (
	driversMu sync.RWMutex
	drivers   map[string]Driver
)

func init() {
	drivers = make(map[string]Driver)
}

type Driver interface {
	Open(url string) (Driver, error)
	Close() error
	// Accepts determines if this storage driver will allow ext.
	Accepts(ext string) bool
	// Path returns the path of the storage driver.
	Path() string
	// Root returns the root dir of the storage driver.
	Root() string
	// NormalizePath returns the passed entries with the necessary prefix
	// e.g. driver.NormalizePath("path/to/file") == "/assets/path/to/file"
	NormalizePath(entries ...string) string
	// AddFile saves the contents of r to path, and returns the absolute path on disk of the
	// created file  and any potential errors.
	AddFile(r io.Reader, path string) (string, error)

	RemoveFile(path string) error
	Dir() http.FileSystem
}

func Open(url string) (Driver, error) {
	u, err := nurl.Parse(url)
	if err != nil {
		return nil, err
	}

	if u.Scheme == "" {
		return nil, fmt.Errorf("storage driver: invalid URL scheme")
	}

	driversMu.RLock()
	d, ok := drivers[u.Scheme]
	driversMu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("storage driver: unknown driver %v (forgotten import)", u.Scheme)
	}

	return d.Open(url)
}

func Register(name string, driver Driver) {
	driversMu.Lock()
	defer driversMu.Unlock()

	if driver == nil {
		panic("storage: Register driver is nil")
	}
	if _, dup := drivers[name]; dup {
		panic("storage: register called twice for driver " + name)
	}
	drivers[name] = driver
}

var (
	// ErrAlreadyExists file already exists.
	ErrAlreadyExists = errors.New("file already exists")
	// ErrInvalidExtension invalid extension.
	ErrInvalidExtension = errors.New("invalid extension")
)
