package testing

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/djangulo/sfd/storage"
)

func Test(t *testing.T, d storage.Driver) {
	t.Run("TestAddFile", func(t *testing.T) { TestAddFile(t, d) })
	t.Run("TestRemoveFile", func(t *testing.T) { TestRemoveFile(t, d) })
}

func TestAddFile(t *testing.T, d storage.Driver) {
	t.Run("success", func(t *testing.T) {
		for _, test := range []struct {
			name string
			r    io.Reader
			path string
		}{
			{"hello world", strings.NewReader("hello world"), "test1/test.tst"},
		} {
			t.Run(test.name, func(t *testing.T) {
				path, err := d.AddFile(test.r, test.path)

				if err != nil {
					t.Errorf("got an error but didn't expect one: %v", err)
				}
				if _, err := os.Stat(path); os.IsNotExist(err) {
					t.Errorf("file was not created at path %s", path)
				}

			})
		}
	})
	t.Run("create dir if none exists", func(t *testing.T) {
		root := "dir1"
		path := filepath.Join(os.TempDir(), root, "dir2", "file.tst")
		path, err := d.AddFile(strings.NewReader("create dir if none exists"), path)
		if err != nil {
			t.Errorf("got an error but didn't expect one: %v", err)
		}
		if _, err := os.Stat(filepath.Dir(path)); os.IsNotExist(err) {
			t.Errorf("directory not created for %s", path)
		}
		os.RemoveAll(filepath.Join(os.TempDir(), root))
	})

	t.Run("errors", func(t *testing.T) {
		t.Run("file already exists", func(t *testing.T) {

			path := "fail.tst"
			d.AddFile(strings.NewReader("duplicate"), path)

			if _, err := d.AddFile(strings.NewReader("duplicate"), path); err != nil && !errors.Is(err, storage.ErrAlreadyExists) {
				t.Errorf("unexpected error %v", err)
			}
		})
	})

}

func TestRemoveFile(t *testing.T, d storage.Driver) {
	t.Run("success", func(t *testing.T) {
		path := "fail.tst"
		d.AddFile(strings.NewReader("duplicate"), path)
		err := d.RemoveFile(path)
		if err != nil {
			t.Errorf("unexpected error %v", err)
		}
	})
}
