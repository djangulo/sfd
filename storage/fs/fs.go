package fs

import (
	"fmt"
	"io"
	"net/http"
	nurl "net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/djangulo/sfd/storage"
)

type Filesystem struct {
	// root folder on disk
	root string
	// path to access assets
	path   string
	accept []string
}

func init() {
	fs := &Filesystem{}
	storage.Register("fs", fs)
}

func (fs *Filesystem) Path() string {
	return fs.path
}

func (fs *Filesystem) Root() string {
	return fs.root
}

func (fs *Filesystem) Accepts(ext string) bool {
	for _, x := range fs.accept {
		if x == ext {
			return true
		}
	}
	return false
}

func (fs *Filesystem) NormalizePath(entries ...string) string {
	entries = append([]string{fs.path}, entries...)
	return filepath.Join(entries...)
}

// Open creates a filesystem object rooted at the path of the urlString.
// 'accept' querystring is a crude validation for the acceptable filetypes.
func (fs *Filesystem) Open(urlString string) (storage.Driver, error) {
	url, err := nurl.Parse(urlString)
	if err != nil {
		return nil, err
	}

	q := url.Query()
	root := q.Get("root")
	if root == "" {
		root = filepath.Join(os.TempDir(), "assets")
	}
	if _, err := os.Stat(root); os.IsNotExist(err) {
		err := os.MkdirAll(root, 0777)
		if err != nil {
			panic(err)
		}
	}
	accept, ok := q["accept"]
	if !ok {
		accept = []string{".jgp", ".jpeg", ".png", ".svg"}
	}
	fs = &Filesystem{root: root, path: url.Path, accept: accept}
	return fs, nil
}

func (fs *Filesystem) appendRoot(path string) string {
	return filepath.Join(fs.root, path)
}

func (fs *Filesystem) acceptsExt(ext string) bool {
	set := make(map[string]struct{})
	for _, a := range fs.accept {
		set[a] = struct{}{}
	}
	if _, ok := set[ext]; !ok {
		return false
	}
	return true
}

// Close noop
func (fs *Filesystem) Close() error {
	return nil
}

func (fs *Filesystem) AddFile(r io.Reader, path string) (string, error) {
	path = strings.TrimPrefix(path, fs.path)
	absPath := filepath.Join(fs.root, path)
	dir := filepath.Dir(absPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.MkdirAll(dir, 0777)
		if err != nil {
			return "", err
		}
	}

	if _, err := os.Stat(absPath); os.IsExist(err) {
		return "", fmt.Errorf("%w at %s", storage.ErrAlreadyExists, path)
	}

	if ext := filepath.Ext(absPath); !fs.acceptsExt(ext) {
		return "", fmt.Errorf("%w %s", storage.ErrInvalidExtension, ext)
	}

	file, err := os.Create(absPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	_, err = io.Copy(file, r)
	if err != nil {
		return "", err
	}

	return absPath, nil
}

func (fs *Filesystem) RemoveFile(path string) error {
	if err := os.Remove(filepath.Join(fs.root, path)); err != nil {
		return err
	}
	return nil
}

func (fs *Filesystem) Dir() http.FileSystem {
	return http.Dir(fs.root)
}
