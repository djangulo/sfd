//Package mocksql links all files inside the db/migrations dir into
// db/mock/migrations and appends migrations with the mocked data
// (starting at 990).
// Test databases can then use this migrations dir to run tests on mocked
// data.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/djangulo/sfd/db/mock"
)

var (
	source   string
	target   string
	dbtables tables
)

func init() {
	gopath := os.Getenv("GOPATH")
	rootDir := filepath.Join(gopath, "src/github.com/djangulo/sfd")
	flag.StringVar(
		&target,
		"target",
		filepath.Join(rootDir, "db/mock/migrations"),
		"target directory to clone/extend migrations to",
	)
	flag.StringVar(
		&source,
		"source",
		filepath.Join(rootDir, "db/migrations"),
		"source directory to clone migrations from",
	)
	flag.Var(&dbtables, "tables", "map of database table names")

}

func main() {
	flag.Parse()

	if err := mock.Migrations(source, target, dbtables); err != nil {
		fmt.Fprintf(os.Stderr, "error writing file %v", err)
		os.Exit(1)
	}

	os.Exit(0)

}

type tables map[string]string

func (t *tables) String() string {
	var s = make([]string, 0)
	for k, v := range *t {
		s = append(s, k+":"+v)
	}
	return strings.Join(s, ",")
}

func (t *tables) Set(value string) error {
	set := map[string]struct{}{
		"users":            {},
		"user_phones":      {},
		"profile_pictures": {},
		"user_preferences": {},
		"user_stats":       {},
		"items":            {},
		"item_bids":        {},
		"item_images":      {},
	}

	for _, pair := range strings.Split(value, ",") {
		s := strings.SplitN(pair, ":", 2)
		if _, ok := set[s[0]]; !ok {
			fmt.Fprintf(os.Stderr, "invalid key for dbtables %s\n", s[0])
			os.Exit(1)
		}
		(*t)[s[0]] = s[1]
	}
	return nil
}

func (t *tables) Values(key string) (s string) {
	s = (*t)[key]
	return
}
