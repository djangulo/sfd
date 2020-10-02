package mock

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/djangulo/sfd/db/models"
)

const (
	usersT           = "users"
	userPhonesT      = "user_phones"
	userAddressesT   = "user_addresses"
	profilePicturesT = "profile_pictures"
	userPreferencesT = "user_preferences"
	userStats        = "user_stats"
	itemsT           = "items"
	itemBidsT        = "item_bids"
	itemImagesT      = "item_images"
)

var (
	defaultTables map[string]string
)

func init() {
	defaultTables = map[string]string{
		usersT:           "sfd.users",
		userPhonesT:      "sfd.user_phones",
		userAddressesT:   "sfd.user_addresses",
		profilePicturesT: "sfd.profile_pictures",
		userPreferencesT: "sfd.user_preferences",
		userStats:        "sfd.user_stats",
		itemsT:           "sfd.items",
		itemBidsT:        "sfd.item_bids",
		itemImagesT:      "sfd.item_images",
	}
}

// Migrations links all files inside the db/migrations dir into
// db/mock/migrations and appends migrations with the mocked data
// (starting at 990).
// Test databases can then use this migrations dir to run tests on mocked
// data.
func Migrations(source, target string, tables map[string]string) error {
	if tables != nil {
		for k, v := range tables {
			if _, ok := defaultTables[k]; ok {
				defaultTables[k] = v
			}
		}
	}

	if _, err := os.Stat(source); os.IsNotExist(err) {
		return fmt.Errorf("source migrations dir (%s) not found", source)
	}
	// remove unconditionally
	if err := os.RemoveAll(target); err != nil {
		return fmt.Errorf("error cleaning %s: %v", target, err)
	}

	if err := os.Mkdir(target, 0755); err != nil {
		return fmt.Errorf("error creating target (%s): %v", target, err)
	}

	if err := linkFiles(source, target); err != nil {
		return fmt.Errorf("error linking files %v", err)
	}

	users := Users()
	items := Items(users)
	bids := Bids(items, users)
	Stats(users, items, bids)

	for filename, data := range map[string]string{
		"990_users_data.up.sql":   usersUp(users),
		"990_users_data.down.sql": usersDown(),
		"991_items_data.up.sql":   itemsUp(items),
		"991_items_data.down.sql": itemsDown(),
		"992_bids_data.up.sql":    bidsUp(bids),
		"992_bids_data.down.sql":  bidsDown(),
	} {
		path := filepath.Join(target, filename)
		if err := writeFile(path, data); err != nil {
			return fmt.Errorf("error writing file %s: %v", path, err)
		}
	}

	return nil

}

func linkFiles(src, tgt string) error {
	err := filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Fprintf(
				os.Stderr,
				"failure accessing a path %q: %v\n",
				path,
				err,
			)
			return err
		}
		if !info.IsDir() {
			target := filepath.Join(tgt, filepath.Base(path))
			if err := os.Symlink(path, target); err != nil {
				fmt.Printf("symlink %q -> %q; error: %v\n", path, target, err)
				return err
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("error walking the path %q: %v", src, err)
	}
	return nil
}

func writeFile(targetName string, data string) error {
	fh, err := os.Create(targetName)
	if err != nil {
		return err
	}
	defer fh.Close()
	_, err = fh.WriteString(data)
	if err != nil {
		return err
	}
	return nil
}

func usersUp(users []*models.User) string {
	stmt := fmt.Sprintf(`BEGIN;

	-- insert test users
INSERT INTO %s
(id, username, email, full_name, password_hash, active, created_at, updated_at, last_login, profile_public)
VALUES
`, defaultTables[usersT])
	for i, user := range users {
		stmt += fmt.Sprintf(
			`('%s', '%s', '%s', '%s', '%s', %t, '%v', '%v', '%v', %t)`,
			user.ID.String(),
			user.Username,
			user.Email,
			user.FullName,
			user.PasswordHash,
			user.Active,
			user.CreatedAt.Format(time.RFC3339Nano),
			user.UpdatedAt.Time.Format(time.RFC3339Nano),
			user.LastLogin.Time.Format(time.RFC3339Nano),
			user.ProfilePublic,
		)
		if i == len(users)-1 {
			stmt += ";\n"
		} else {
			stmt += ",\n"
		}
	}

	stmt += fmt.Sprintf(`

--insert test user stats
INSERT INTO %s (user_id, login_count, items_created, bids_created, bids_won)
VALUES
`, defaultTables[userStats])
	for i, user := range users {
		s := user.Stats
		stmt += fmt.Sprintf(
			`('%s', %d, %d, %d, %d)`,
			s.UserID.String(),
			s.LoginCount,
			s.ItemsCreated,
			s.BidsCreated,
			s.BidsWon,
		)
		if i == len(users)-1 {
			stmt += ";\n"
		} else {
			stmt += ",\n"
		}
	}

	stmt += fmt.Sprintf(`

-- insert test user phone numbers
INSERT INTO %s (id, user_id, number)
VALUES
`, defaultTables[userPhonesT])
	for i, user := range users {
		for j, p := range user.PhoneNumbers {
			stmt += fmt.Sprintf(
				`('%s', '%s', '%s')`,
				p.ID.String(),
				p.UserID.String(),
				p.Number,
			)
			if i == len(users)-1 && j == len(user.PhoneNumbers)-1 {
				stmt += ";\n"
			} else {
				stmt += ",\n"
			}
		}
	}

	stmt += fmt.Sprintf(`

-- insert test user addresses
INSERT INTO %s (id, user_id, address, kind)
VALUES
`, defaultTables[userAddressesT])
	for i, user := range users {
		for j, p := range user.Addresses {
			stmt += fmt.Sprintf(
				`('%s', '%s', '%s', %d)`,
				p.ID.String(),
				p.UserID.String(),
				p.Address,
				p.Kind,
			)
			if i == len(users)-1 && j == len(user.Addresses)-1 {
				stmt += ";\n"
			} else {
				stmt += ",\n"
			}
		}
	}

	stmt += fmt.Sprintf(`

-- insert test user pictures
INSERT INTO %s (id, path, abs_path, original_filename, alt_text, file_ext, "order", user_id)
VALUES
`, defaultTables[profilePicturesT])
	for i, user := range users {
		p := user.Picture
		stmt += fmt.Sprintf(
			`('%s', '%s', '%s', '%s', '%s', '%s', %d, '%s')`,
			p.ID.UUID.String(),
			p.Path.String,
			p.AbsPath.String,
			p.OriginalFilename.String,
			p.AltText.String,
			p.FileExt.String,
			p.Order.Int64,
			p.UserID.String(),
		)
		if i == len(users)-1 {
			stmt += ";\n"
		} else {
			stmt += ",\n"
		}
	}

	stmt += fmt.Sprintf(`

-- insert test user preferences
INSERT INTO %s (user_id, language)
VALUES
`, defaultTables[userPreferencesT])
	for i, user := range users {
		p := user.Preferences
		stmt += fmt.Sprintf(
			`('%s', '%s')`,
			p.UserID.String(),
			p.Language,
		)
		if i == len(users)-1 {
			stmt += ";\n"
		} else {
			stmt += ",\n"
		}
	}
	stmt += "\nCOMMIT;"
	return stmt
}

func usersDown() string {
	return fmt.Sprintf(`BEGIN;
TRUNCATE %s;
TRUNCATE %s;
TRUNCATE %s;
TRUNCATE %s;
TRUNCATE %s;
TRUNCATE %s;
COMMIT;`,
		defaultTables[userPreferencesT],
		defaultTables[profilePicturesT],
		defaultTables[userAddressesT],
		defaultTables[userPhonesT],
		defaultTables[userStats],
		defaultTables[usersT],
	)
}

func itemsUp(items []*models.Item) string {
	stmt := fmt.Sprintf(`-- insert test items
INSERT INTO %s
(
	id, owner_id, name, slug, description, starting_price,
	max_price, min_increment, bid_interval, bid_deadline, blind,
	published_at, created_at, admin_approved
)
VALUES
`, defaultTables[itemsT])
	for i, item := range items {
		stmt += fmt.Sprintf(
			`(
	'%s', -- id
	'%s', -- owner_id
	'%s','%s','%s', -- name, slug, description
	%s,	%s, %s,	%d, -- starting_price, max_price, min_increment, bid_interval
	'%v', %t, -- bids_deadline, blind
	'%v', -- published
	'%v', -- created
	%t -- admin_approved
)`,
			item.ID.String(),
			item.OwnerID.String(),
			item.Name,
			item.Slug,
			item.Description,
			item.StartingPrice,
			item.MaxPrice,
			item.MinIncrement,
			item.BidInterval,
			item.BidDeadline.Format(time.RFC3339Nano),
			item.Blind,
			item.PublishedAt.Time.Format(time.RFC3339Nano),
			item.CreatedAt.Format(time.RFC3339Nano),
			item.AdminApproved,
		)
		if i == len(items)-1 {
			stmt += ";\n"
		} else {
			stmt += ",\n"
		}
	}

	stmt += fmt.Sprintf(`

-- insert test item images
INSERT INTO %s (id, path, abs_path, original_filename, alt_text, file_ext, "order", item_id)
VALUES
`, defaultTables[itemImagesT])
	for i, item := range items {
		for j, p := range item.Images {
			stmt += fmt.Sprintf(
				`('%s', '%s', '%s', '%s', '%s', '%s', %d, '%s')`,
				p.ID.UUID.String(),
				p.Path.String,
				p.AbsPath.String,
				p.OriginalFilename.String,
				p.AltText.String,
				p.FileExt.String,
				p.Order.Int64,
				p.ItemID,
			)
			if i == len(items)-1 && j == len(item.Images)-1 {
				stmt += ";\n"
			} else {
				stmt += ",\n"
			}
		}
	}

	return stmt
}

func itemsDown() string {
	return fmt.Sprintf(`BEGIN;
TRUNCATE %s;
TRUNCATE %s;
COMMIT;`,
		defaultTables[itemsT],
		defaultTables[itemImagesT],
	)
}

func bidsUp(bids []*models.Bid) string {
	stmt := fmt.Sprintf(`-- insert test bids
INSERT INTO %s
(
	id, user_id, item_id, amount, valid,
	created_at, updated_at
)
VALUES
`, defaultTables[itemBidsT])
	for i, bid := range bids {
		stmt += fmt.Sprintf(
			`('%s', '%s', '%s', %s, %t, '%v', '%v')`,
			bid.ID.String(),
			bid.UserID.String(),
			bid.ItemID.String(),
			bid.Amount,
			bid.Valid,
			bid.CreatedAt.Format(time.RFC3339),
			bid.UpdatedAt.Time.Format(time.RFC3339),
		)
		if i == len(bids)-1 {
			stmt += ";\n"
		} else {
			stmt += ",\n"
		}
	}

	return stmt
}

func bidsDown() string {
	return fmt.Sprintf(`BEGIN;
TRUNCATE %s;
COMMIT;`,
		defaultTables[itemBidsT],
	)
}
