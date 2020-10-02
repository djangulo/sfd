package postgres

import (
	"github.com/djangulo/sfd/crypto/token"
)

func (pg *Postgres) GetToken(digest string, kind token.Kind) (*token.Token, error) {

	stmt := `SELECT * FROM sfd.tokens WHERE digest = $1 AND kind = $2 LIMIT 1;`

	var t token.Token
	if err := pg.db.QueryRowx(stmt, digest, kind).StructScan(&t); err != nil {
		return nil, err
	}

	return &t, nil
}

func (pg *Postgres) DeleteToken(digest string) error {
	stmt := `DELETE FROM sfd.tokens WHERE digest = $1;`

	if err := pg.db.QueryRowx(stmt, digest).Err(); err != nil {
		return err
	}
	return nil
}

func (pg *Postgres) SaveToken(t *token.Token) error {
	stmt := `INSERT INTO sfd.tokens (
		digest, expires, created_at, updated_at, user_id, pw_hash, last_login, kind
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	ON CONFLICT (digest, kind)
	DO UPDATE
	SET
		expires = EXCLUDED.expires,
		updated_at = EXCLUDED.updated_at;`

	if err := pg.db.QueryRowx(
		stmt,
		t.Digest,
		t.Expires,
		t.CreatedAt,
		t.UpdatedAt,
		t.UserID,
		t.PWHash,
		t.LastLogin,
		t.Kind,
	).Err(); err != nil {
		return err
	}
	return nil
}

func (pg *Postgres) TokenGC() error {
	stmt := `DELETE FROM sfd.tokens WHERE expires < NOW();`
	_, err := pg.db.Exec(stmt)
	if err != nil {
		return err
	}
	return nil
}
