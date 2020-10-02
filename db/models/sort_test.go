package models

import (
	"testing"
	"time"

	"github.com/gofrs/uuid"
)

var (
	id1 = uuid.Must(uuid.FromString("11111111-1111-1111-1111-111111111111"))
	id2 = uuid.Must(uuid.FromString("21111111-1111-1111-1111-111111111111"))
)

func TestSortItems(t *testing.T) {
	var items = []*Item{
		{
			Name: "1",
			DBObj: &DBObj{
				ID:        id1,
				CreatedAt: time.Now(),
			},
		},
		{
			Name: "2",
			DBObj: &DBObj{
				ID:        id2,
				CreatedAt: time.Now().Add(1 * time.Second),
			},
		},
	}

	t.Run("id ascending", func(t *testing.T) {
		ItemsOrderedBy(SortItemsByIDAsc).Sort(items)
		a, b := items[0], items[1]
		if a.Name != "1" || b.Name != "2" {
			t.Error("wrong order for items")
		}
	})
	t.Run("id descending", func(t *testing.T) {
		ItemsOrderedBy(SortItemsByIDDesc).Sort(items)
		a, b := items[0], items[1]
		if a.Name != "2" || b.Name != "1" {
			t.Error("wrong order for items")
		}
	})
	t.Run("created at ascending", func(t *testing.T) {
		ItemsOrderedBy(SortItemsByCreatedAtAsc).Sort(items)
		a, b := items[0], items[1]
		if a.Name != "1" || b.Name != "2" {
			t.Error("wrong order for items")
		}
	})
	t.Run("created at descending", func(t *testing.T) {
		ItemsOrderedBy(SortItemsByCreatedAtDesc).Sort(items)
		a, b := items[0], items[1]
		if a.Name != "2" || b.Name != "1" {
			t.Error("wrong order for items")
		}
	})
}

func TestSortUsers(t *testing.T) {
	var users = []*User{
		{
			Username: "1",
			DBObj: &DBObj{
				ID:        id1,
				CreatedAt: time.Now(),
			},
		},
		{
			Username: "2",
			DBObj: &DBObj{
				ID:        id2,
				CreatedAt: time.Now().Add(1 * time.Second),
			},
		},
	}

	t.Run("id ascending", func(t *testing.T) {
		UsersOrderedBy(SortUsersByIDAsc).Sort(users)
		a, b := users[0], users[1]
		if a.Username != "1" || b.Username != "2" {
			t.Error("wrong order for users")
		}
	})
	t.Run("id descending", func(t *testing.T) {
		UsersOrderedBy(SortUsersByIDDesc).Sort(users)
		a, b := users[0], users[1]
		if a.Username != "2" || b.Username != "1" {
			t.Error("wrong order for users")
		}
	})
	t.Run("created at ascending", func(t *testing.T) {
		UsersOrderedBy(SortUsersByCreatedAtAsc).Sort(users)
		a, b := users[0], users[1]
		if a.Username != "1" || b.Username != "2" {
			t.Error("wrong order for users")
		}
	})
	t.Run("created at descending", func(t *testing.T) {
		UsersOrderedBy(SortUsersByCreatedAtDesc).Sort(users)
		a, b := users[0], users[1]
		if a.Username != "2" || b.Username != "1" {
			t.Error("wrong order for users")
		}
	})
}
