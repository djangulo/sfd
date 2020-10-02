package memory

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/gofrs/uuid"

	"github.com/djangulo/sfd/db/models"
	"github.com/djangulo/sfd/pagination"
)

func (m *Memory) GetItem(ctx context.Context, id *uuid.UUID) (*models.Item, error) {
	var i *models.Item
	var err error
	for _, item := range m.items {
		if item.ID == *id {
			i = item
		}
	}
	if i == nil {
		return nil, fmt.Errorf("id: %q %w", id, models.ErrNotFound)
	}
	i.WinningBid, err = m.ItemWinningBid(&i.ID)
	if err != nil {
		return nil, err
	}
	return i, nil
}

func (m *Memory) GetBid(ctx context.Context, id *uuid.UUID) (*models.Bid, error) {
	for _, bid := range m.bids {
		if bid.ID == *id {
			return bid, nil
		}
	}
	return nil, fmt.Errorf("id: %q %w", id, models.ErrNotFound)
}

func (m *Memory) ListItems(ctx context.Context) ([]*models.Item, error) {
	var (
		limit       int        = 10
		lastID      *uuid.UUID = &uuid.Nil
		lastCreated *time.Time = &time.Time{}
	)
	opts, ok := ctx.Value(pagination.CtxKey).(pagination.Options)
	if ok {
		limit, lastID, lastCreated = opts.Limit(), opts.LastID(), opts.LastCreated()
	}

	models.ItemsOrderedBy(
		models.SortItemsByCreatedAtDesc,
		models.SortItemsByIDDesc,
	).Sort(m.items)

	var items = make([]*models.Item, 0)
	for _, item := range m.items {
		if item.AdminApproved {
			items = append(items, item)
		}
	}

	if lastID != nil && *lastID != uuid.Nil && !lastCreated.IsZero() {

		for i, item := range items {
			created := item.CreatedAt
			if item.ID == *lastID && created.Equal(*lastCreated) {
				if itemsLeft := len(items[(i + 1):]); limit > itemsLeft {
					limit = itemsLeft
				}
				return items[(i + 1):(i + 1 + limit)], nil
			}
		}
	} else {
		if limit > len(items) {
			limit = len(items)
		}
		return items[:limit], nil
	}
	return nil, models.ErrNoResults
}

func (m *Memory) GetBySlug(slug string) (*models.Item, error) {
	for _, item := range m.items {
		if item.Slug == slug {
			return item, nil
		}
	}
	return nil, fmt.Errorf("slug: %q %w", slug, models.ErrNotFound)
}

func (m *Memory) PublishItem(itemID *uuid.UUID, datetime time.Time) error {
	for _, item := range m.items {
		if item.ID == *&item.ID {
			item.PublishedAt = sql.NullTime{Valid: true, Time: datetime}
		}
	}
	return nil
}

func (m *Memory) WatchItem(itemID, userID *uuid.UUID) error {
	m.userItemsWatch = append(m.userItemsWatch, &itemWatch{
		UserID:    userID,
		ItemID:    itemID,
		CreatedAt: time.Now(),
	})
	return nil
}

func (m *Memory) UnWatchItem(itemID, userID *uuid.UUID) error {
	for i, watch := range m.userItemsWatch {
		if *watch.UserID == *userID && *watch.ItemID == *itemID {
			m.userItemsWatch[i] = m.userItemsWatch[len(m.userItemsWatch)-1]
			m.userItemsWatch[len(m.userItemsWatch)-1] = nil
			m.userItemsWatch = m.userItemsWatch[:len(m.userItemsWatch)-1]
			return nil
		}
	}
	return fmt.Errorf("%w: %v %v", models.ErrNotFound, itemID, userID)
}

func (m *Memory) CreateItem(item *models.Item) error {
	m.items = append(m.items, item)
	return nil
}
func (m *Memory) ItemBids(ctx context.Context, itemID *uuid.UUID) ([]*models.Bid, error) {
	var (
		limit      int     = 10
		lastAmount float64 = 0.0
	)
	opts, ok := ctx.Value(pagination.CtxKey).(pagination.Options)
	if ok {
		limit, lastAmount = opts.Limit(), opts.LastAmount()
	}

	models.BidsOrderedBy(
		models.SortBidsByAmountDesc,
	).Sort(m.bids)

	var bids = make([]*models.Bid, 0)
	for _, bid := range m.bids {
		if *bid.ItemID == *itemID {
			bids = append(bids, bid)
		}
	}

	if lastAmount > 0 {

		amt := models.MustCurrency(lastAmount)
		for i, bid := range bids {
			if bid.Amount.Eq(amt) {
				if bidsLeft := len(bids[(i + 1):]); limit > bidsLeft {
					limit = bidsLeft
				}
				return bids[(i + 1):(i + 1 + limit)], nil
			}
		}
	} else {
		if limit > len(bids) {
			limit = len(bids)
		}
		return bids[:limit], nil
	}
	return nil, models.ErrNoResults
}

func (m *Memory) ItemWinningBid(itemID *uuid.UUID) (*models.Bid, error) {
	bids := make([]*models.Bid, 0)
	for _, bid := range m.bids {
		if *bid.ItemID == *itemID {
			bids = append(bids, bid)
		}
	}
	models.BidsOrderedBy(
		models.SortBidsByAmountDesc,
	).Sort(bids)
	return bids[0], nil
}

func (m *Memory) UserBids(userID *uuid.UUID, opts *models.ListOptions) (int, []*models.Bid, error) {
	bids := make([]*models.Bid, 0)
	for _, b := range m.bids {
		if *b.UserID == *userID {
			bids = append(bids, b)
		}
	}
	for _, b := range bids {
		for _, i := range m.items {
			for _, bid := range m.bids {
				i.Bids = make([]*models.Bid, 0)
				if *bid.ItemID == i.ID {
					i.Bids = append(i.Bids, bid)
				}
			}
			if *b.ItemID == i.ID {
				b.Item = i
			}

		}
	}
	var count = len(bids)

	models.BidsOrderedBy(
		models.SortBidsByCreatedAtDesc,
	).Sort(bids)

	if opts == nil {
		length := len(bids)
		if length < 1000 {
			return count, bids[:length], nil
		}
		return count, bids[:1000], nil
	}

	limit := opts.Limit
	length := len(bids)
	if length < limit {
		return count, bids[:length], nil
	}
	return count, bids[:limit], nil

}

func (m *Memory) CoverImage(itemID *uuid.UUID) ([]*models.ItemImage, error) {
	var images = make([]*models.ItemImage, 1, 1)
	for _, img := range m.images {
		if *(img.ItemID) == *itemID && img.Order.Int64 == 1 {
			images[0] = img
			return images, nil
		}
	}
	return nil, models.ErrNoResults
}

// PlaceBid places a bid on item.
func (m *Memory) PlaceBid(bid *models.Bid) error {
	m.bids = append(m.bids, bid)
	return nil
}

// InvalidateBid invalidates a bid by ID.
func (m *Memory) InvalidateBid(bidID *uuid.UUID) error {
	for _, bid := range m.bids {
		if bid.ID == *bidID {
			bid.Valid = false
			return nil
		}
	}
	return models.ErrNotFound
}

func (m *Memory) ItemImages(ctx context.Context, itemID *uuid.UUID) ([]*models.ItemImage, error) {
	var (
		limit     int = 10
		lastOrder int = 0
	)
	opts, ok := ctx.Value(pagination.CtxKey).(pagination.Options)
	if ok {
		limit, lastOrder = opts.Limit(), opts.LastOrder()
	}

	models.ImagesOrderedBy(
		models.SortImagesByOrderAsc,
	).Sort(m.images)

	var images = make([]*models.ItemImage, 0)
	for _, image := range m.images {
		if *image.ItemID == *itemID {
			images = append(images, image)
		}
	}

	if lastOrder > 0 {
		for i, bid := range images {
			if bid.Order.Int64 == int64(lastOrder) {
				if imagesLeft := len(images[(i + 1):]); limit > imagesLeft {
					limit = imagesLeft
				}
				return images[(i + 1):(i + 1 + limit)], nil
			}
		}
	} else {
		if limit > len(images) {
			limit = len(images)
		}
		return images[:limit], nil
	}
	return nil, models.ErrNoResults
}

func (m *Memory) AddImages(itemID *uuid.UUID, images ...*models.ItemImage) error {
	for _, i := range images {
		m.images = append(m.images, i)
	}
	return nil

}

func (m *Memory) RemoveImage(id *uuid.UUID) error {
	for i, img := range m.images {
		if img.ID.Valid {
			if img.ID.UUID == *id {
				m.images[i] = m.images[len(m.images)-1]
				m.images[len(m.images)-1] = nil
				m.images = m.images[:len(m.images)-1]
				return nil
			}
		}
	}
	return fmt.Errorf("%w: %v", models.ErrNotFound, id)
}

func (m *Memory) DeleteItem(id *uuid.UUID) error {
	for i, item := range m.items {
		if item.ID == *id {
			m.items[i] = m.items[len(m.items)-1]
			m.items[len(m.items)-1] = nil
			m.items = m.items[:len(m.items)-1]
			return nil
		}
	}
	return fmt.Errorf("%w: %v", models.ErrNotFound, id)
}

func (m *Memory) UserItemBids(ctx context.Context, userID, itemID *uuid.UUID) ([]*models.Bid, error) {
	var (
		limit       int        = 10
		lastID      *uuid.UUID = &uuid.Nil
		lastCreated *time.Time = &time.Time{}
	)
	opts, ok := ctx.Value(pagination.CtxKey).(pagination.Options)
	if ok {
		limit, lastID, lastCreated = opts.Limit(), opts.LastID(), opts.LastCreated()
	}

	models.BidsOrderedBy(
		models.SortBidsByCreatedAtDesc,
		models.SortBidsByIDDesc,
	).Sort(m.bids)

	var bids = make([]*models.Bid, 0)
	for _, bid := range m.bids {
		if *bid.ItemID == *itemID {
			bids = append(bids, bid)
		}
	}

	if lastID != nil && *lastID != uuid.Nil && !lastCreated.IsZero() {

		for i, bid := range bids {
			created := bid.CreatedAt.Truncate(time.Second)
			if bid.ID == *lastID && created.Equal(*lastCreated) {
				if bidsLeft := len(bids[(i + 1):]); limit > bidsLeft {
					limit = bidsLeft
				}
				return bids[(i + 1):(i + 1 + limit)], nil
			}
		}
	} else {
		if limit > len(bids) {
			limit = len(bids)
		}
		return bids[:limit], nil
	}
	return nil, models.ErrNoResults
}

func (m *Memory) BulkInvalidateBids(ids ...*uuid.UUID) error {
	set := make(map[*uuid.UUID]struct{})
	for _, id := range ids {
		set[id] = struct{}{}
	}
	for _, bid := range m.bids {
		if _, ok := set[&bid.ID]; ok {
			bid.Valid = false
		}
	}
	return nil
}

func (m *Memory) BidsByID(ids ...*uuid.UUID) ([]*models.Bid, error) {
	set := make(map[*uuid.UUID]struct{})
	for _, id := range ids {
		set[id] = struct{}{}
	}

	var bids = make([]*models.Bid, 0)
	for _, bid := range m.bids {
		if _, ok := set[&bid.ID]; ok {
			bids = append(bids, bid)
		}
	}
	return bids, nil
}
