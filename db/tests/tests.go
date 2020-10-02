//Package testing is a common set of unit tests for db.Driver implementations.
// All that should be tested is that:
//    - Data can be retrieved/modified/deleted successfully.
//    - Data is in the expected shape. See the db.Driver interface for these
//      details.
package testing

import (
	"errors"
	"testing"

	"github.com/gofrs/uuid"

	"github.com/djangulo/sfd/db"
	"github.com/djangulo/sfd/db/mock"
	"github.com/djangulo/sfd/db/models"
	testutils "github.com/djangulo/sfd/testing"
)

func Test(t *testing.T, d db.Driver) {
	t.Run("TestGetBySlug", func(t *testing.T) { TestGetBySlug(t, d) })
	// t.Run("TestListItems", func(t *testing.T) { TestListItems(t, d) })
	// t.Run("TestItemBids", func(t *testing.T) { TestItemBids(t, d) })
	t.Run("TestCreateItem", func(t *testing.T) { TestCreateItem(t, d) })

	t.Run("TestActivate", func(t *testing.T) { TestActivate(t, d) })
	// t.Run("TestDeactivateUser", func(t *testing.T) { TestDeactivateUser(t, d) })
	// t.Run("TestGrantAdmin", func(t *testing.T) { TestGrantAdmin(t, d) })
	// t.Run("TestRevokeAdmin", func(t *testing.T) { TestRevokeAdmin(t, d) })
	// t.Run("TestListUsers", func(t *testing.T) { TestListUsers(t, d) })
	t.Run("TestUserByUsername", func(t *testing.T) { TestUserByUsername(t, d) })
	// t.Run("TestCreateUser", func(t *testing.T) { TestCreateUser(t, d) })
}

var (
	items []*models.Item
	users []*models.User
	bids  []*models.Bid
)

func init() {
	users = mock.Users()
	items = mock.Items(users)
	bids = mock.Bids(items, users)
}

func TestGetBySlug(t *testing.T, d db.Driver) {
	t.Run("errors", func(t *testing.T) {
		for _, test := range []struct {
			name string
			in   string
			want error
		}{
			{"not found", "i do not exist", models.ErrNotFound},
		} {
			t.Run(test.name, func(t *testing.T) {
				_, got := d.GetBySlug(test.in)
				if !errors.Is(got, test.want) {
					t.Errorf("got %q want %q", got, test.want)
				}
			})
		}
	})
	t.Run("success", func(t *testing.T) {

		for _, item := range items {
			got, err := d.GetBySlug(item.Slug)
			if err != nil {
				t.Errorf("got an error but didn't want one: %v", err)
			}
			testutils.AssertItemsEqual(t, got, item)
		}
	})
}

// func TestListItems(t *testing.T, d db.Driver) {
// 	t.Run("success", func(t *testing.T) {
// 		for _, test := range []struct {
// 			name  string
// 			in    *models.ListOptions
// 			count int
// 		}{

// 			{"limit=2", &models.ListOptions{Limit: 2}, 2},
// 			{"limit=4", &models.ListOptions{Limit: 4}, 4},
// 			{"offset=2", &models.ListOptions{Limit: 1000, Offset: 2}, len(items)},
// 			{"offset=4", &models.ListOptions{Limit: 1000, Offset: 4}, len(items)},
// 			{"limit=2+offset=2", &models.ListOptions{Limit: 2, Offset: 2}, 2},
// 			{"limit=4+offset=4", &models.ListOptions{Limit: 4, Offset: 4}, 4},
// 		} {
// 			t.Run(test.name, func(t *testing.T) {

// 				got, err := d.ListItems(test.in)
// 				if err != nil {
// 					t.Errorf("got an error but didn't want one %v", err)
// 				}
// 				if count <= 0 {
// 					t.Errorf("expected count > 0 but count is %d", count)
// 				}
// 				if len(got) != test.wantLen {
// 					t.Errorf("got %d want %d", len(got), test.wantLen)
// 				}
// 			})
// 		}

// 	})
// }
// func TestItemBids(t *testing.T, d db.Driver) {
// 	t.Run("success", func(t *testing.T) {
// 		// contains 10 sample items,
// 		itemID := items[0].ID

// 		iBids := make([]*models.Bid, 0)
// 		for _, b := range bids {
// 			if *b.ItemID == itemID {
// 				iBids = append(iBids, b)
// 			}
// 		}

// 		for _, test := range []struct {
// 			name     string
// 			in       *options
// 			wantLen  int
// 			wantBids []*models.Bid
// 		}{
// 			{"limit 2", &options{limit: 2}, 2, iBids[:2]},
// 			{"limit 4", &options{limit: 4}, 4, iBids[:4]},
// 			{"limit overflow", &options{limit: 99999}, len(iBids), iBids},
// 			//			{"keyed 2", &store.PaginateOptions{LastID: &iBids[1].ID, LastCreated: iBids[1].CreatedAt, Limit: 2}, 2, iBids[2:4]},
// 			//			{"keyed 4", &store.PaginateOptions{LastID: &iBids[3].ID, LastCreated: iBids[3].CreatedAt, Limit: 4}, 4, iBids[4:8]},
// 		} {
// 			t.Run(test.name, func(t *testing.T) {
// 				got, err := d.ItemBids(conte&itemID, test.in)
// 				if err != nil {
// 					t.Fatalf("got an error but didn't want one %v", err)
// 				}
// 				if len(got) != test.wantLen {
// 					t.Fatalf("got %d want %d", len(got), test.wantLen)
// 				}
// 				assertBidCollectionsEqual(t, got, test.wantBids)
// 			})
// 		}

// 	})
// }

func TestCreateItem(t *testing.T, d db.Driver) {
	t.Run("success", func(t *testing.T) {
		for _, test := range []struct {
			name string
			item *models.Item
		}{
			{name: "item", item: &models.Item{DBObj: &models.DBObj{ID: uuid.Must(uuid.NewV4())}, Name: "Test create item", Slug: "test-create-item", OwnerID: &users[0].ID}},
		} {
			err := d.CreateItem(test.item)
			if err != nil {
				t.Errorf("got an error but didn't expect one: %v", err)
			}
			_, err = d.GetBySlug(test.item.Slug)
			if err != nil {
				t.Errorf("got an error but didn't expect one: %v", err)
			}
		}
	})
}

func TestAddImages(t *testing.T, d db.Driver) {

}

func TestActivate(t *testing.T, d db.Driver) {
	users[1].IsAdmin = true
	t.Run("success", func(t *testing.T) {
		err := d.ActivateUser(&users[0].ID, &users[1].ID)
		testutils.AssertNoError(t, err)
		u, err := d.UserByID(&users[0].ID)
		testutils.AssertNoError(t, err)
		if !u.Active {
			t.Error("expected Active to be true but it isn't")
		}
	})
	t.Run("errors", func(t *testing.T) {
		for _, test := range []struct {
			name string
			user *models.User
		}{
			{"non-existing user", &models.User{Username: "non-existing", DBObj: &models.DBObj{ID: uuid.Must(uuid.NewV4())}}},
		} {
			t.Run(test.name, func(t *testing.T) {
				err := d.ActivateUser(&test.user.ID, &users[1].ID)
				if err == nil {
					t.Error("expected an error but didn't get one")
				}
			})
		}
	})
}

// func TestDeactivateUser(t *testing.T, d db.Driver) {
// 	t.Run("success", func(t *testing.T) {
// 		err := d.DeactivateUser(&users[0].ID)
// 		assertNoError(t, err)
// 		u, _ := d.UserByID(&users[0].ID)
// 		if u.Active {
// 			t.Error("expected IsActive to be false but it isn't")
// 		}
// 	})
// 	t.Run("errors", func(t *testing.T) {
// 		for _, test := range []struct {
// 			name string
// 			user *models.User
// 			want error
// 		}{
// 			{"non-existing user", &models.User{Username: "non-existing", ID: uuid.Must(uuid.NewV4())}, models.ErrNotFound},
// 			{"nil id user", &models.User{Username: "i_have_a_nil_id", ID: uuid.Nil}, models.ErrInvalidInput},
// 		} {
// 			t.Run(test.name, func(t *testing.T) {
// 				err := d.ActivateUser(&test.user.ID, &users[1].ID)
// 				if !errors.Is(err, test.want) {
// 					t.Errorf("unexpected error: %v", err)
// 				}
// 			})
// 		}
// 	})
// }

// func TestGrantAdmin(t *testing.T, d db.Driver) {
// 	t.Run("success", func(t *testing.T) {
// 		err := d.GrantAdmin(&users[0].ID)
// 		assertNoError(t, err)
// 		u, _ := d.UserByID(&users[0].ID)
// 		if !u.IsAdmin {
// 			t.Error("expected IsAdmin to be true but it isn't")
// 		}
// 	})
// 	t.Run("errors", func(t *testing.T) {
// 		for _, test := range []struct {
// 			name string
// 			user *models.User
// 			want error
// 		}{
// 			{"non-existing user", &models.User{Username: "non-existing", ID: uuid.Must(uuid.NewV4())}, models.ErrNotFound},
// 			{"nil id user", &models.User{Username: "i_have_a_nil_id", ID: uuid.Nil}, models.ErrInvalidInput},
// 		} {
// 			t.Run(test.name, func(t *testing.T) {
// 				err := d.GrantAdmin(&test.user.ID)
// 				if !errors.Is(err, test.want) {
// 					t.Errorf("unexpected error: %v", err)
// 				}
// 			})
// 		}
// 	})
// }

// func TestRevokeAdmin(t *testing.T, d db.Driver) {
// 	t.Run("success", func(t *testing.T) {
// 		err := d.RevokeAdmin(&users[0].ID)
// 		assertNoError(t, err)
// 		u, _ := d.UserByID(&users[0].ID)
// 		if u.IsAdmin {
// 			t.Error("expected IsAdmin to be false but it isn't")
// 		}
// 	})
// 	t.Run("errors", func(t *testing.T) {
// 		for _, test := range []struct {
// 			name string
// 			user *models.User
// 			want error
// 		}{
// 			{"non-existing user", &models.User{Username: "non-existing", ID: uuid.Must(uuid.NewV4())}, models.ErrNotFound},
// 			{"nil id user", &models.User{Username: "i_have_a_nil_id", ID: uuid.Nil}, models.ErrInvalidInput},
// 		} {
// 			t.Run(test.name, func(t *testing.T) {
// 				err := d.GrantAdmin(&test.user.ID)
// 				if !errors.Is(err, test.want) {
// 					t.Errorf("unexpected error: %v", err)
// 				}
// 			})
// 		}
// 	})
// }

// func TestCreateUser(t *testing.T, d db.Driver) {
// 	t.Run("success", func(t *testing.T) {
// 		user := models.User{
// 			ID:       uuid.Must(uuid.NewV4()),
// 			Username: "Test_user_newly_created",
// 			IsAdmin:  false,
// 			Active:   false,
// 		}
// 		err := d.CreateUser(&user)
// 		assertNoError(t, err)
// 		if user.ID == uuid.Nil {
// 			t.Error("expected a populated ID but returned nil")
// 		}
// 	})
// 	t.Run("errors", func(t *testing.T) {
// 		for _, test := range []struct {
// 			name string
// 			user *models.User
// 			want error
// 		}{
// 			{"existing user", users[0], models.ErrAlreadyExists},
// 		} {
// 			t.Run(test.name, func(t *testing.T) {
// 				err := d.CreateUser(test.user)
// 				if !errors.Is(err, test.want) {
// 					t.Errorf("unexpected error: %v", err)
// 				}
// 			})
// 		}
// 	})
// }

func TestUserByUsername(t *testing.T, d db.Driver) {
	t.Run("success", func(t *testing.T) {
		user, err := d.UserByUsernameOrEmail(users[0].Username)
		testutils.AssertNoError(t, err)
		testutils.AssertUsersEqual(t, user, users[0])
	})
	t.Run("errors", func(t *testing.T) {
		for _, test := range []struct {
			name, username string
			want           error
		}{
			{"not found", "doesn't really exist", models.ErrNotFound},
		} {
			t.Run(test.name, func(t *testing.T) {
				_, err := d.UserByUsernameOrEmail(test.username)
				if !errors.Is(err, test.want) {
					t.Errorf("unexpected error: %v", err)
				}
			})
		}
	})
}

// func TestListUsers(t *testing.T, d db.Driver) {

// 	for _, test := range []struct {
// 		name    string
// 		in      *options
// 		wantLen int
// 		want    []*models.User
// 	}{
// 		{"limit=2", &options{limit: 2}, 2, users[:2]},
// 		{"limit=4", &options{limit: 4}, 4, users[:4]},
// 		{"limit overflow", &options{limit: 99999}, len(users), users},
// 		{"offset=1", &options{limit: 1000, offset: 1}, len(users) - 1, users[1:]},
// 		{"offset=3", &options{limit: 1000, offset: 3}, len(users) - 3, users[3:]},
// 		{"limit=2+offset=2", &options{limit: 2, offset: 2}, 2, users[2:4]},
// 	} {
// 		t.Run(test.name, func(t *testing.T) {
// 			_, got, err := d.ListUsers(test.in)
// 			if err != nil {
// 				t.Fatalf("got an error but didn't want one %v", err)
// 			}
// 			if len(got) != test.wantLen {
// 				t.Errorf("got %d want %d", len(got), test.wantLen)
// 			}
// 			assertUserCollectionsEqual(t, got, test.want)
// 		})
// 	}

// }
