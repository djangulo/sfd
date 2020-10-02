package postgres

import (
	"fmt"

	"github.com/djangulo/sfd/db/models"
	"github.com/gofrs/uuid"
)

func (pg *Postgres) ActivateUser(userID, approverID *uuid.UUID) error {
	tx, err := pg.db.Beginx()
	if err != nil {
		return fmt.Errorf("transaction error: %w", err)
	}

	stmt := `SELECT * FROM sfd.activateUser($1, $2);`
	if err := tx.QueryRowx(stmt, userID, approverID).Err(); err != nil {
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

func (pg *Postgres) DeactivateUser(userID *uuid.UUID) error {
	stmt := `UPDATE sfd.users SET active = FALSE WHERE id = $1;`
	if err := pg.db.QueryRowx(stmt, userID).Err(); err != nil {
		return fmt.Errorf("exec error: %w", err)
	}
	return nil
}

func (pg *Postgres) GrantAdmin(userID *uuid.UUID) error {
	stmt := `UPDATE sfd.users SET is_admin = TRUE WHERE id = $1;`
	if err := pg.db.QueryRowx(stmt, userID).Err(); err != nil {
		return fmt.Errorf("exec error: %w", err)
	}
	return nil
}

func (pg *Postgres) RevokeAdmin(userID *uuid.UUID) error {
	stmt := `UPDATE sfd.users SET is_admin = FALSE WHERE id = $1;`
	if err := pg.db.QueryRowx(stmt, userID).Err(); err != nil {
		return fmt.Errorf("exec error: %w", err)
	}
	return nil
}

func (pg *Postgres) ListUsers(opts *models.ListOptions) (int, []*models.User, error) {
	// stmt := `UPDATE sfd.users SET is_admin = FALSE WHERE id = $1;`
	// if err := pg.db.QueryRowx(stmt, userID).Err(); err != nil {
	// 	return fmt.Errorf("exec error: %w", err)
	// }
	// return nil
	stmt := `
	SELECT
	*,
	(SELECT COUNT(*) FROM sfd.users_stats) AS count
	FROM sfd.users_stats`

	stmt += fmt.Sprintf(" LIMIT %d", opts.Limit)
	if opts.Offset > 0 {
		stmt += fmt.Sprintf(" OFFSET %d", opts.Offset)
	}
	stmt += ";"

	rows, err := pg.db.Queryx(stmt)
	if err != nil {
		return 0, nil, err
	}

	var users = make([]*models.User, 0)
	var i, count int
	for rows.Next() {
		var obj = struct {
			Count int `db:"count"`
			*models.User
		}{}
		err = rows.StructScan(&obj)
		if i == 0 {
			count = obj.Count
			i++
		}
		users = append(users, obj.User)
	}
	if err := rows.Close(); err != nil {
		return 0, nil, err
	}

	return count, users, nil
}
