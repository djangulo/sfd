package memory

import (
	"testing"

	"github.com/djangulo/sfd/db"
	dbtest "github.com/djangulo/sfd/db/tests"
)

func Test(t *testing.T) {
	d, err := db.Open("memory://irrelevanthost/?prepopulate=true")
	if err != nil {
		t.Fatal(err)
	}
	t.Run("db/memory:Test", func(t *testing.T) { dbtest.Test(t, d) })
}
