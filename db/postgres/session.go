package postgres

import (
	"github.com/djangulo/sfd/crypto/session"
)

func (pg *Postgres) NewSession(session session.Session) error {

	stmt := `INSERT INTO sfd.sessions (
		id, values, expires, created_at)
	VALUES
	($1, $2, $3, $4)
	RETURNING id;`

	b, err := session.Bytes()
	if err != nil {
		return err
	}

	id, created, _ := session.Metadata()

	if err := pg.db.QueryRowx(
		stmt,
		id,
		b,
		session.Expiry(),
		created,
	).StructScan(session); err != nil {
		return err
	}

	return nil
}

func (pg *Postgres) ReadSession(id string) ([]byte, error) {
	stmt := `SELECT values FROM sfd.sessions WHERE id = $1 LIMIT 1;`

	var b []byte
	if err := pg.db.Get(&b, stmt, id); err != nil {
		return nil, err
	}
	return b, nil
}

func (pg *Postgres) DeleteSession(id string) error {
	stmt := `DELETE FROM sfd.sessions WHERE id = $1;`

	if err := pg.db.QueryRowx(stmt, id).Err(); err != nil {
		return err
	}
	return nil
}

func (pg *Postgres) UpdateSession(ses session.Session) error {
	stmt := `
	UPDATE sfd.sessions
	SET (values, expires, updated_at) = ($1, $2, $3)
	WHERE id = $4;`
	id, _, updatedAt := ses.Metadata()
	b, err := ses.Bytes()
	if err != nil {
		return err
	}

	if err := pg.db.QueryRowx(
		stmt,
		b,
		ses.Expiry(),
		updatedAt,
		id,
	).Err(); err != nil {
		return err
	}

	return nil
}

func (pg *Postgres) SessionGC() error {
	stmt := `DELETE FROM sfd.sessions WHERE expires < NOW();`
	_, err := pg.db.Exec(stmt)
	if err != nil {
		return err
	}
	return nil
}
