// +build memory all

package accounts_test

import (
	"crypto/sha256"
	"testing"

	pkg "github.com/djangulo/sfd/accounts"
	pkgTest "github.com/djangulo/sfd/accounts/tests"
	"github.com/djangulo/sfd/config"
	"github.com/djangulo/sfd/crypto/token"
	"github.com/djangulo/sfd/db"
	_ "github.com/djangulo/sfd/db/memory"
	"github.com/djangulo/sfd/mail"
	_ "github.com/djangulo/sfd/mail/console"
	"github.com/djangulo/sfd/storage"
	_ "github.com/djangulo/sfd/storage/fs"
)

func Test(t *testing.T) {
	storageDriver, err := storage.Open("fs://./assets?accept=.jpg&accept=.png&accept=.jpeg&accept=.svg&root=/tmp/assets")
	if err != nil {
		panic(err)
	}

	dbDriver, err := db.Open("memory://irrelevanthost/?prepopulate=true")
	if err != nil {
		panic(err)
	}

	mailDriver, err := mail.Open("console://./?quiet=true")
	if err != nil {
		panic(err)
	}

	cnf := config.NewConfig()
	cnf.Defaults()

	tm, err := token.NewManager(dbDriver, sha256.New, cnf)
	if err != nil {
		panic(err)
	}

	server, err := pkg.NewServer(dbDriver, mailDriver, cnf, storageDriver, tm)
	if err != nil {
		panic(err)
	}
	pkgTest.Test(t, server)
}
