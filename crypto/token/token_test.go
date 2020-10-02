package token

import (
	"crypto/sha256"
	"crypto/sha512"
	"database/sql"
	"hash"
	"testing"
	"time"

	"github.com/gofrs/uuid"

	"github.com/djangulo/sfd/config"
	"github.com/djangulo/sfd/db/models"
)

func TestManager(t *testing.T) {
	user := models.User{
		LastLogin:    sql.NullTime{Valid: true, Time: time.Now()},
		PasswordHash: "abcd1234",
		DBObj: &models.DBObj{
			ID: uuid.Must(uuid.NewV4()),
		},
	}
	conf := config.NewConfig()
	conf.Defaults()

	store := make(mockStore)

	for _, test := range []struct {
		name     string
		hashFunc func() hash.Hash
	}{
		{"SHA256", sha256.New},
		{"SHA224", sha256.New224},
		{"SHA512", sha512.New},
		{"SHA384", sha512.New384},
		{"SHA512-256", sha512.New512_256},
	} {
		t.Run(test.name, func(t *testing.T) {
			tg, _ := NewManager(store, test.hashFunc, conf)

			expiry := time.Now().Add(30 * time.Second)
			token, _ := tg.NewToken(
				user.LastLogin.Time,
				&user.ID,
				user.PasswordHash,
				Registration,
				&expiry,
			)

			tk, err := tg.CheckToken(token.Digest, Registration)
			if err != nil {
				t.Errorf("got an error but didn't expect one: %v", err)
			}
			if *tk.UserID != user.ID {
				t.Errorf("expected %v got %v", user.ID, *tk.UserID)
			}
		})

	}

	t.Run("failure", func(t *testing.T) {
		user2 := models.User{
			DBObj:        &models.DBObj{ID: uuid.UUID{}},
			LastLogin:    sql.NullTime{Valid: true, Time: time.Now()},
			PasswordHash: "4321dcba",
		}
		tg, _ := NewManager(store, sha256.New, conf)

		t.Run("expired", func(t *testing.T) {
			expiry := time.Now().Add(-1 * time.Second)
			token, _ := tg.NewToken(
				user2.LastLogin.Time,
				&user2.ID,
				user2.PasswordHash,
				Registration,
				&expiry,
			)

			if _, err := tg.CheckToken(token.Digest, Registration); err == nil {
				t.Errorf("expected an error but did not get one")
			}
		})
	})

}

type mockStore map[string]*Token

func (ms mockStore) GetToken(digest string, kind Kind) (*Token, error) {
	if t, ok := ms[digest]; ok {
		return t, nil
	}
	return nil, models.ErrNotFound
}
func (ms mockStore) SaveToken(token *Token) error {
	ms[token.Digest] = token
	return nil
}

// DeleteToken explicitly deletes a token from the store.
func (ms mockStore) DeleteToken(digest string) error {
	delete(ms, digest)
	return nil
}
func (ms mockStore) TokenGC() error {
	for k, v := range ms {
		if v.Expires.Before(time.Now()) {
			delete(ms, k)
		}
	}
	return nil
}
