// +build postgres all

package postgres

import (
	"database/sql"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/djangulo/sfd/db"
	"github.com/djangulo/sfd/db/mock"
	pkgTest "github.com/djangulo/sfd/db/tests"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq" // blank import
	"github.com/ory/dockertest"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func TestPostgres(t *testing.T) {

	gopath := os.Getenv("GOPATH")
	rootDir := filepath.Join(gopath, "src/github.com/djangulo/sfd")
	source := filepath.Join(rootDir, "db/migrations")
	target := filepath.Join(rootDir, "db/mock/migrations")

	if err := mock.Migrations(source, target, nil); err != nil {
		panic(fmt.Sprintf("error writing file %v", err))
	}

	pool, err := dockertest.NewPool("")
	if err != nil {
		t.Fatalf("could not connect to docker: %s", err)
	}

	resource, err := pool.Run(
		"postgres",
		"12.4",
		[]string{
			"POSTGRES_PASSWORD=abcd1234",
			"POSTGRES_DB=sfd_db",
		})
	if err != nil {
		t.Fatalf("could not start resource: %s", err)
	}
	if err := resource.Expire(60); err != nil {
		t.Fatalf("error setting container expiry: %v", err)
	}

	connStr := fmt.Sprintf(
		"postgres://postgres:abcd1234@localhost:%s/%s?sslmode=disable",
		resource.GetPort("5432/tcp"),
		"sfd_db",
	)

	if err = pool.Retry(func() error {
		var err error
		pgDB, err := sql.Open("postgres", connStr)
		if err != nil {
			return err
		}
		return pgDB.Ping()
	}); err != nil {
		t.Fatalf("Could not connect to docker: %s", err)
	}

	// migrate
	m, err := migrate.New("file://"+target, connStr)
	if err != nil {
		t.Fatalf("error migrating test db: %v\n", err)
	}
	if err := m.Up(); err != nil {
		t.Fatalf("migrate up error: %v\n", err)
	}
	if _, err := m.Close(); err != nil {
		t.Fatalf("migrate close error: %v\n", err)
	}

	d, err := db.Open(connStr)
	if err != nil {
		t.Fatal(err)
	}

	// Test using db
	t.Run("db/postgres:Test", func(t *testing.T) { pkgTest.Test(t, d) })

	// When you're done, kill and remove the container
	err = pool.Purge(resource)
	if err != nil {
		fmt.Println(err)
	}
}
