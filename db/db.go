//Package db implements the drivers needed for Database access.
// It's its own package to prevent cyclic imports: reduces duplication by
// embedding its consumer's interfaces.
// db also abstracts database interactions by implementing at the package
// level the same functions as the db.Driver interface using a global
// *manager object. This way the store implementations can stay as dumb as
// possible, while the db package handles all validation.
// See db/validators for validation details.
package db

import (
	"fmt"
	nurl "net/url"
	"sync"
)

var (
	driversMu sync.RWMutex
	drivers   map[string]Driver
	m         *manager
)

type manager struct {
	driver Driver
}

func init() {
	m = &manager{}
	drivers = make(map[string]Driver)
}

// Open returns a new driver instance.
func Open(url string) (Driver, error) {
	u, err := nurl.Parse(url)
	if err != nil {
		return nil, err
	}

	if u.Scheme == "" {
		return nil, fmt.Errorf("source driver: invalid URL scheme")
	}

	driversMu.RLock()
	d, ok := drivers[u.Scheme]
	driversMu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("source driver: unknown driver %v (forgotten import?)", u.Scheme)
	}

	return d.Open(url)
}

// Register globally registers a
func Register(name string, drv Driver) {
	driversMu.Lock()
	defer driversMu.Unlock()

	if drv == nil {
		panic("bid: Register driver is nil")
	}
	if _, dup := drivers[name]; dup {
		panic("bid: Register called twice for driver " + name)
	}
	drivers[name] = drv
}
