// +build memory all

package items_test

import (
	"testing"

	"github.com/djangulo/sfd/db"
	_ "github.com/djangulo/sfd/db/memory"
	pkg "github.com/djangulo/sfd/items"
	pkgTest "github.com/djangulo/sfd/items/tests"
	"github.com/djangulo/sfd/storage"
	_ "github.com/djangulo/sfd/storage/fs"
)

func TestMemory(t *testing.T) {
	fsConnStr := "fs://irrelevant/?accept=.jpg&accept=.png&accept=.svg&root=/tmp/assets"
	storageDrv, err := storage.Open(fsConnStr)
	if err != nil {
		panic(err)
	}

	memConnStr := "memory://irrelevanthost/?prepopulate=true"
	store, err := db.Open(memConnStr)
	if err != nil {
		panic(err)
	}

	server, err := pkg.NewServer(store, storageDrv)
	if err != nil {
		panic(err)
	}
	pkgTest.Test(t, server)
}
