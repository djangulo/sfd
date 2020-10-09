//Package postgres contains postgres-specific implementation of the
// db.Driver interface.
// Each file contains the functions that are concerned with each package
// e.g. admin.go has the AdminStorer functions, that in turn deals with the
// admin package.
package postgres

import (
	"fmt"
	"strings"

	"github.com/djangulo/sfd/db"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// Postgres sturct to hold db.
type Postgres struct {
	db *sqlx.DB
}

func init() {
	p := &Postgres{}
	db.Register("postgres", p)
}

// Open creates a connection to the Postgres DB as an sqlx.DB object.
func (pg *Postgres) Open(urlString string) (db.Driver, error) {
	var err error
	pg.db, err = sqlx.Open("postgres", urlString)
	if err != nil {
		return nil, err
	}

	return pg, nil
}

// Close closes the database
func (pg *Postgres) Close() error {
	return pg.db.Close()
}

// BuildValueStrings returns postgres-escaped VALUES strings for fields of length
// insertions, to be used in bulk-inserts.
// e.g.
// calling buildValueStrings(length=2, fields=3) to insert 3 fields of
// an array of length 2, returns
//     ($1, $2, $3), ($4, $5, $6)
// which then results into the query:
//     INSERT INTO table (a, b, c)
//     VALUES
//         ($1, $2, $3),
//         ($4, $5, $6);
func BuildValueStrings(length, fields int) string {
	var b strings.Builder

	for i := 0; i < length; i++ {
		b.WriteRune('(')
		for j := 1; j <= fields; j++ {
			fmt.Fprintf(&b, "$%d", i*fields+j)
			if j < fields {
				b.WriteRune(',')
			}
		}
		if i == length-1 {
			b.WriteString(")")
		} else {
			b.WriteString("),\n")
		}
	}
	return b.String()
}
