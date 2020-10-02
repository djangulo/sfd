package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/djangulo/sfd/db/models"
	"github.com/djangulo/sfd/pagination"
)

func (pg *Postgres) GetItem(ctx context.Context, id *uuid.UUID) (*models.Item, error) {
	stmt := `SELECT * FROM sfd.items_owner WHERE id = $1 LIMIT 1;`

	var item = new(models.Item)
	if err := pg.db.GetContext(ctx, item, stmt, id); err != nil {
		return nil, err
	}
	stmt = `SELECT * FROM sfd.item_images WHERE item_id = $1`
	rows, err := pg.db.QueryxContext(ctx, stmt, id)
	if err != nil {
		return nil, err
	}
	item.Images = make([]*models.ItemImage, 0)
	for rows.Next() {
		var img = new(models.ItemImage)
		if err := rows.StructScan(img); err != nil {
			return nil, err
		}
		item.Images = append(item.Images, img)
	}
	item.WinningBid, err = pg.ItemWinningBid(&item.ID)
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (pg *Postgres) GetBid(ctx context.Context, id *uuid.UUID) (*models.Bid, error) {
	stmt := `SELECT * FROM sfd.bids_full WHERE id = $1 LIMIT 1;`

	var bid = new(models.Bid)
	if err := pg.db.GetContext(ctx, bid, stmt, id); err != nil {
		return nil, err
	}
	return bid, nil
}

func (pg *Postgres) ListItems(ctx context.Context) ([]*models.Item, error) {
	var (
		limit       int        = 10
		lastID      *uuid.UUID = &uuid.Nil
		lastCreated *time.Time = &time.Time{}
		rows        *sqlx.Rows
		err         error
	)
	opts, ok := ctx.Value(pagination.CtxKey).(pagination.Options)
	if ok {
		limit, lastID, lastCreated = opts.Limit(), opts.LastID(), opts.LastCreated()
	}

	stmt := `SELECT * FROM sfd.items_owner`
	if *lastID != uuid.Nil && !lastCreated.IsZero() {
		stmt += ` WHERE (created_at, id) < ($1, $2)
		AND admin_approved = TRUE
		ORDER BY created_at DESC, id DESC LIMIT $3`
		rows, err = pg.db.Queryx(stmt, lastCreated, lastID, limit)
	} else {
		stmt += " WHERE admin_approved = TRUE ORDER BY created_at DESC, id DESC LIMIT $1;"
		rows, err = pg.db.Queryx(stmt, limit)
	}
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}

	var objs = make([]*models.Item, 0)
	for rows.Next() {
		var obj = new(models.Item)
		err = rows.StructScan(obj)
		if err != nil {
			return nil, err
		}
		objs = append(objs, obj)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}

	return objs, nil
}

func (pg *Postgres) CreateItem(item *models.Item) error {
	tx, err := pg.db.Beginx()
	if err != nil {
		return fmt.Errorf("transaction error: %w", err)
	}
	stmt := `
	INSERT INTO sfd.items
	(
		id, owner_id, name, slug, description, starting_price, max_price,
		min_increment, bid_interval, bid_deadline, blind, created_at
	)
	VALUES 
	($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12);
	`

	if err := tx.QueryRowx(
		stmt,
		item.ID,
		item.OwnerID,
		item.Name,
		item.Slug,
		item.Description,
		item.StartingPrice,
		item.MaxPrice,
		item.MinIncrement,
		item.BidInterval,
		item.BidDeadline,
		item.Blind,
		item.CreatedAt,
	).Err(); err != nil {
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

func (pg *Postgres) ItemBids(ctx context.Context, itemID *uuid.UUID) ([]*models.Bid, error) {
	var (
		limit      int     = 10
		lastAmount float64 = 0.0
		rows       *sqlx.Rows
		err        error
	)
	opts, ok := ctx.Value(pagination.CtxKey).(pagination.Options)
	if ok {
		limit, lastAmount = opts.Limit(), opts.LastAmount()
	}

	stmt := `SELECT * FROM sfd.bids_full`
	if lastAmount > 0.0 {
		stmt += ` WHERE item_id = $1
		AND (amount) < ($2)
		ORDER BY amount DESC LIMIT $3`
		rows, err = pg.db.Queryx(stmt, itemID, lastAmount, limit)
	} else {
		stmt += " WHERE item_id = $1 ORDER BY amount DESC LIMIT $2;"
		rows, err = pg.db.Queryx(stmt, itemID, limit)
	}
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}

	var objs = make([]*models.Bid, 0)
	for rows.Next() {
		var obj = new(models.Bid)
		err = rows.StructScan(obj)
		if err != nil {
			return nil, err
		}
		objs = append(objs, obj)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}

	return objs, nil
}

func (pg *Postgres) ItemWinningBid(itemID *uuid.UUID) (*models.Bid, error) {
	stmt := `
	SELECT * FROM sfd.bids_full
	WHERE item_id = $1
	ORDER BY amount DESC
	LIMIT 1;`
	var bid = new(models.Bid)
	if err := pg.db.Get(bid, stmt, itemID); err != nil {
		return nil, err
	}
	return bid, nil
}

// func (pg *Postgres) ItemBids2(itemID *uuid.UUID, opts *models.ListOptions) (int, []*models.Bid, error) {
// 	if opts == nil {
// 		opts = models.NewOptions()
// 	}
// 	stmt := `
// 	SELECT * FROM sfd.bids_full
// 	WHERE item_id = $1
// 	ORDER BY amount DESC`
// 	// Filter, no filters for this query yet
// 	// if opts.Inactive() {
// 	// 	stmt += " AND published_at != NULL"
// 	// }

// 	stmt += fmt.Sprintf(" LIMIT %d", opts.Limit)
// 	if opts.Offset > 0 {
// 		stmt += fmt.Sprintf(" OFFSET %d", opts.Offset)
// 	}
// 	stmt += ";"

// 	rows, err := pg.db.Queryx(stmt, itemID)
// 	if err != nil {
// 		return 0, nil, err
// 	}

// 	var objs = make([]*models.Bid, 0)
// 	var i, count int
// 	for rows.Next() {
// 		var obj = struct {
// 			Count int `db:"count"`
// 			*models.Bid
// 		}{Count: 0, Bid: &models.Bid{}}
// 		err = rows.StructScan(&obj)
// 		if err != nil {
// 			log.Println(err)
// 		}
// 		if i == 0 {
// 			count = obj.Count
// 			i++
// 		}
// 		objs = append(objs, obj.Bid)
// 	}
// 	if err := rows.Close(); err != nil {
// 		return 0, nil, err
// 	}

// 	return count, objs, nil
// }

func (pg *Postgres) PublishItem(itemID *uuid.UUID, datetime time.Time) error {
	stmt := `
	UPDATE sfd.items
	SET published_at = $2
	WHERE id = $1;`
	if err := pg.db.QueryRowx(stmt, itemID, datetime).Err(); err != nil {
		return err
	}
	return nil
}

func (pg *Postgres) GetBySlug(slug string) (*models.Item, error) {
	stmt := `SELECT * FROM sfd.items_owner WHERE slug = $1 LIMIT 1;`
	var obj models.Item
	if err := pg.db.Get(&obj, stmt, slug); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrNotFound
		}
		return nil, err
	}
	return &obj, nil
}

func (pg *Postgres) UserBids(userID *uuid.UUID, opts *models.ListOptions) (int, []*models.Bid, error) {
	if opts == nil {
		opts = models.NewOptions()
	}
	stmt := `
	SELECT
	*,
	(SELECT COUNT(*) FROM sfd.bids_full WHERE user_id = $1) AS count
	FROM sfd.bids_full
	WHERE user_id = $1
	ORDER BY created_at DESC`
	// Filter, no filters for this query yet
	// if filters.Unpublished() {
	// 	stmt += " AND published_at != NULL"
	// }

	stmt += fmt.Sprintf(" LIMIT %d", opts.Limit)
	if opts.Offset > 0 {
		stmt += fmt.Sprintf(" OFFSET %d", opts.Offset)
	}
	stmt += ";"

	rows, err := pg.db.Queryx(stmt, userID)
	if err != nil {
		return 0, nil, err
	}

	var objs = make([]*models.Bid, 0)
	var i, count int
	for rows.Next() {
		var obj = struct {
			Count int `db:"count"`
			*models.Bid
		}{}
		err = rows.StructScan(&obj)
		if i == 0 {
			count = obj.Count
			i++
		}
		objs = append(objs, obj.Bid)
	}
	if err := rows.Close(); err != nil {
		return 0, nil, err
	}

	return count, objs, nil
}

func (pg *Postgres) UserItemBids(ctx context.Context, userID, itemID *uuid.UUID) ([]*models.Bid, error) {
	var (
		limit       int        = 10
		lastID      *uuid.UUID = &uuid.Nil
		lastCreated *time.Time = &time.Time{}
		rows        *sqlx.Rows
		err         error
	)
	opts, ok := ctx.Value(pagination.CtxKey).(pagination.Options)
	if ok {
		limit, lastID, lastCreated = opts.Limit(), opts.LastID(), opts.LastCreated()
	}

	stmt := `SELECT * FROM sfd.bids_full`
	if lastID != nil && *lastID != uuid.Nil && !lastCreated.IsZero() {
		stmt += ` WHERE item_id = $1 AND user_id = $2
		AND (created_at, id) < ($3, $4)
		ORDER BY created_at DESC, id DESC LIMIT $5`
		rows, err = pg.db.Queryx(stmt, itemID, userID, lastCreated, lastID, limit)
	} else {
		stmt += ` WHERE item_id = $1 AND user_id = $2
		ORDER BY created_at DESC, id DESC
		LIMIT $3;`
		rows, err = pg.db.Queryx(stmt, itemID, userID, limit)
	}
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}

	var objs = make([]*models.Bid, 0)
	for rows.Next() {
		var obj = new(models.Bid)
		err = rows.StructScan(obj)
		if err != nil {
			return nil, err
		}
		objs = append(objs, obj)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}

	return objs, nil
}

func (pg *Postgres) ItemImages(ctx context.Context, itemID *uuid.UUID) ([]*models.ItemImage, error) {
	var (
		limit     int = 10
		lastOrder int = 0
		rows      *sqlx.Rows
		err       error
	)
	opts, ok := ctx.Value(pagination.CtxKey).(pagination.Options)
	if ok {
		limit, lastOrder = opts.Limit(), opts.LastOrder()
	}

	stmt := `SELECT * FROM sfd.item_images`
	if lastOrder > 0 {
		stmt += ` WHERE item_id = $1
		AND ("order") > ($2)
		ORDER BY "order" ASC LIMIT $3`
		rows, err = pg.db.Queryx(stmt, itemID, lastOrder, limit)
	} else {
		stmt += ` WHERE item_id = $1 ORDER BY "order" ASC LIMIT $2;`
		rows, err = pg.db.Queryx(stmt, itemID, limit)
	}
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}

	var objs = make([]*models.ItemImage, 0)
	for rows.Next() {
		var obj = new(models.ItemImage)
		err = rows.StructScan(obj)
		if err != nil {
			return nil, err
		}
		objs = append(objs, obj)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}

	return objs, nil
}

func (pg *Postgres) DeleteItem(id *uuid.UUID) error {
	stmt := `UPDATE sfd.items SET deleted_at = $2 WHERE id = $1;`

	if err := pg.db.QueryRowx(stmt, id, time.Now()).Err(); err != nil {
		return err
	}
	return nil
}

func (pg *Postgres) PlaceBid(bid *models.Bid) error {
	tx, err := pg.db.Beginx()
	if err != nil {
		return fmt.Errorf("transaction error: %w", err)
	}

	stmt := `
	INSERT INTO sfd.item_bids
	(
		id, created_at, amount, user_id, item_id, valid
	)
	VALUES 
	($1, $2, $3, $4, $5, $6);
	`

	if err := tx.QueryRowx(
		stmt,
		bid.ID,
		bid.CreatedAt,
		bid.Amount,
		bid.UserID,
		bid.ItemID,
		bid.Valid,
	).Err(); err != nil {
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

func (pg *Postgres) WatchItem(itemID, userID *uuid.UUID) error {
	stmt := `INSERT INTO sfd.users_item_watch
	(user_id, item_id, created_at)
VALUES ($1, $2, $3);`
	if err := pg.db.QueryRowx(stmt, userID, itemID, time.Now()).Err(); err != nil {
		return err
	}
	return nil
}

func (pg *Postgres) UnWatchItem(itemID, userID *uuid.UUID) error {
	stmt := `DELETE FROM sfd.users_item_watch WHERE item_id = $1 AND user_id = $2;`
	if err := pg.db.QueryRowx(stmt, userID, itemID).Err(); err != nil {
		return err
	}
	return nil
}

func (pg *Postgres) InvalidateBid(id *uuid.UUID) error {
	stmt := `UPDATE sfd.item_bids SET valid = FALSE WHERE id = $1;`
	if err := pg.db.QueryRowx(stmt, id).Err(); err != nil {
		return err
	}
	return nil
}

func (pg *Postgres) BulkInvalidateBids(ids ...*uuid.UUID) error {
	valueStrings := "("
	args := make([]interface{}, 0)
	for i := 1; i <= len(ids); i++ {
		valueStrings += fmt.Sprintf("$%d", i)
		if i == len(ids) {
			valueStrings += ")"
		} else {
			valueStrings += ","
		}
		args = append(args, ids[i-1])
	}
	stmt := fmt.Sprintf(
		`UPDATE sfd.item_bids SET valid=FALSE WHERE id IN %s;`,
		valueStrings,
	)
	if err := pg.db.QueryRowx(stmt, args...).Err(); err != nil {
		return err
	}
	return nil
}

func (pg *Postgres) BidsByID(ids ...*uuid.UUID) ([]*models.Bid, error) {
	valueStrings := "("
	args := make([]interface{}, 0)
	for i := 1; i <= len(ids); i++ {
		valueStrings += fmt.Sprintf("$%d", i)
		if i == len(ids) {
			valueStrings += ")"
		} else {
			valueStrings += ","
		}
		args = append(args, ids[i-1])
	}
	stmt := fmt.Sprintf(
		`SELECT * FROM sfd.bids_full WHERE id IN %s;`,
		valueStrings,
	)
	rows, err := pg.db.Queryx(stmt, args...)
	if err != nil {
		return nil, err
	}

	var objs = make([]*models.Bid, 0)
	for rows.Next() {
		var obj models.Bid
		if err := rows.StructScan(&obj); err != nil {
			return nil, err
		}
		objs = append(objs, &obj)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	return objs, nil
}

func (pg *Postgres) AddImages(itemID *uuid.UUID, images ...*models.ItemImage) error {
	valueStrings := BuildValueStrings(len(images), 8)
	args := make([]interface{}, 0)
	for _, img := range images {
		args = append(args, img.ID)
		args = append(args, img.Path)
		args = append(args, img.AbsPath)
		args = append(args, img.OriginalFilename)
		args = append(args, img.FileExt)
		args = append(args, img.AltText)
		args = append(args, img.CreatedAt)
		args = append(args, img.Order)
	}

	stmt := fmt.Sprintf(`
	INSERT INTO sfd.item_images (
		id, path, abs_path, original_filename, file_ext, alt_text, created_at, "order"
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

func (pg *Postgres) RemoveImage(id *uuid.UUID) error {
	stmt := `UPDATE sfd.item_images SET deleted_at = $2 WHERE id = $1;`

	if err := pg.db.QueryRowx(stmt, id, time.Now()).Err(); err != nil {
		return err
	}
	return nil
}
