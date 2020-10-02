package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/gofrs/uuid"

	"github.com/djangulo/sfd/db/models"
)

func (pg *Postgres) LoginUser(user *models.User) (*models.User, error) {

	tx, err := pg.db.Beginx()
	if err != nil {
		return nil, fmt.Errorf("transaction error: %w", err)
	}
	stmt := `SELECT * FROM sfd.loginUser($1, $2);`
	if err := tx.QueryRowx(stmt, user.ID, time.Now()).StructScan(user); err != nil {
		return nil, fmt.Errorf("exec error: %w", err)
	}
	if err := tx.Commit(); err != nil {
		if err := tx.Rollback(); err != nil {
			return nil, fmt.Errorf("rollback error: %w", err)
		}
		return nil, fmt.Errorf("commit error: %w", err)
	}
	return user, nil
}

func (pg *Postgres) LogoutUser(userID *uuid.UUID) error {
	stmt := `UPDATE sfd.users SET is_logged_in = FALSE WHERE id = $1;`
	if err := pg.db.QueryRowx(stmt, userID).Err(); err != nil {
		return err
	}
	return nil
}

func (pg *Postgres) CreateUser(user *models.User) error {
	tx, err := pg.db.Beginx()
	if err != nil {
		return fmt.Errorf("transaction error: %w", err)
	}
	stmt := `SELECT * FROM sfd.createUser($1, $2, $3, $4, $5, $6, $7, $8);`

	if err := tx.Get(
		user,
		stmt,
		user.ID,
		user.Username,
		user.Email,
		user.FullName,
		false,      //is admin,
		false,      //active
		time.Now(), // created_at
		user.PasswordHash,
	); err != nil {
		return fmt.Errorf("exec error: %w", err)
	}
	if err := tx.Commit(); err != nil {
		if err := tx.Rollback(); err != nil {
			return fmt.Errorf("rollback error: %w", err)
		}
		return fmt.Errorf("commit error: %w", err)
	}
	return nil
}

// UserByUsernameOrEmail returns a *models.User by username or email
func (pg *Postgres) UserByUsernameOrEmail(usernameOrEmail string) (*models.User, error) {
	stmt := `SELECT * FROM sfd.users_stats WHERE username = $1 OR email = $1 LIMIT 1;`
	var obj models.User
	err := pg.db.Get(&obj, stmt, usernameOrEmail)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrNotFound
		}
		return nil, err
	}
	return &obj, nil
}

// UserByID returns a *models.User by username or email
func (pg *Postgres) UserByID(id *uuid.UUID) (*models.User, error) {
	stmt := `SELECT * FROM sfd.users_stats WHERE id = $1 LIMIT 1;`
	var obj models.User
	if err := pg.db.Get(&obj, stmt, id); err != nil {
		return nil, err
	}
	return &obj, nil
}

// ChangePassword returns a *models.User by username or email
func (pg *Postgres) ResetPassword(userID *uuid.UUID, newPasswordHash string) error {
	stmt := `
	UPDATE sfd.users
	SET password_hash = crypt($2, gen_salt('md5'))
	WHERE id = $1;`
	if err := pg.db.QueryRowx(stmt, userID, newPasswordHash).Err(); err != nil {
		return err
	}
	return nil
}

// ChangePassword returns a *models.User by username or email
func (pg *Postgres) ChangePassword(user *models.User, newPasswordHash string) (*models.User, error) {
	stmt := `
	UPDATE sfd.users
	SET password_hash = crypt($2, gen_salt('md5'))
	WHERE id = $1
	RETURNING password_hash;`
	if err := pg.db.QueryRowx(stmt, user.ID, newPasswordHash).StructScan(&user); err != nil {
		return nil, err
	}
	return user, nil
}

func (pg *Postgres) AddProfilePic(userID *uuid.UUID, img *models.ProfilePicture) error {

	stmt := `INSERT INTO sfd.profile_pictures
	(id, user_id, path, abs_path, original_filename, alt_text, order, created_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8);`
	if err := pg.db.QueryRowx(
		stmt,
		img.ID,
		userID,
		img.Path,
		img.AbsPath,
		img.OriginalFilename,
		img.AltText,
		img.Order,
		img.CreatedAt,
	).Err(); err != nil {
		return fmt.Errorf("exec error: %w", err)
	}
	return nil
}

func (pg *Postgres) RemoveProfilePic(userID *uuid.UUID) error {
	stmt := `DELETE FROM sfd.profile_pictures WHERE user_id = $1;`
	if err := pg.db.QueryRowx(stmt, userID).Err(); err != nil {
		return fmt.Errorf("exec error: %w", err)
	}
	return nil
}

// AddPhoneNumbers builds a bulk insert for user phone numbers.
func (pg *Postgres) AddPhoneNumbers(userID *uuid.UUID, phoneNumbers ...*models.PhoneNumber) error {

	valueStrings := BuildValueStrings(len(phoneNumbers), 3)
	args := make([]interface{}, 0)
	for _, phone := range phoneNumbers {
		args = append(args, phone.ID)
		args = append(args, userID)
		args = append(args, phone.Number)
	}

	stmt := fmt.Sprintf(`
	INSERT INTO sfd.user_phones (
		id, user_id, "number"
	) VALUES
	%s
	ON CONFLICT DO NOTHING;`, valueStrings)

	tx, err := pg.db.Beginx()
	if err != nil {
		return fmt.Errorf("transaction error: %w", err)
	}

	if _, err := tx.Exec(stmt, args...); err != nil {
		return fmt.Errorf("exec error: %w", err)
	}

	if err := tx.Commit(); err != nil {
		if err := tx.Rollback(); err != nil {
			return fmt.Errorf("rollback error: %w", err)
		}
		return fmt.Errorf("commit error: %w", err)
	}

	return nil
}

func (pg *Postgres) RemovePhoneNumber(phoneID *uuid.UUID) error {
	stmt := `DELETE FROM sfd.user_phones WHERE id = $1`
	if err := pg.db.QueryRowx(stmt, phoneID).Err(); err != nil {
		return fmt.Errorf("exec error: %w", err)
	}
	return nil
}

func (pg *Postgres) UnsubEmail(email string, kind models.NoMailKind) error {
	stmt := `INSERT INTO sfd.nomail_list (email, kind, created_at)
	VALUES ($1, $2, $3);`
	if err := pg.db.QueryRowx(stmt, email, kind, time.Now()).Err(); err != nil {
		return fmt.Errorf("exec error: %w", err)
	}
	return nil
}

func (pg *Postgres) ResubEmail(email string, kind models.NoMailKind) error {
	stmt := `DELETE FROM sfd.nomail_list WHERE email = $1 AND kind = $2;`
	if err := pg.db.QueryRowx(stmt, email, kind).Err(); err != nil {
		return fmt.Errorf("exec error: %w", err)
	}
	return nil
}

func (pg *Postgres) IsUnsub(email string, kind models.NoMailKind) bool {
	stmt := `SELECT email FROM sfd.nomail_list WHERE email = $1 AND kind = $2;`
	if err := pg.db.QueryRowx(stmt, email, kind).Err(); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false
		}
		// any other errors, we should assume that the user is on the list
		return true
	}
	return true
}
