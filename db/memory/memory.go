package memory

import (
	nurl "net/url"
	"time"

	"github.com/djangulo/sfd/crypto/session"
	"github.com/djangulo/sfd/crypto/token"
	"github.com/djangulo/sfd/db"
	"github.com/djangulo/sfd/db/mock"
	"github.com/djangulo/sfd/db/models"
	"github.com/gofrs/uuid"
)

func init() {
	m := Memory{
		items:           make([]*models.Item, 0),
		bids:            make([]*models.Bid, 0),
		images:          make([]*models.ItemImage, 0),
		users:           make([]*models.User, 0),
		profilePictures: make([]*models.ProfilePicture, 0),
		userStats:       make([]*models.UserStats, 0),
		userPhones:      make([]*models.PhoneNumber, 0),
		nomail:          make([]*models.NoMail, 0),
		userItemsWatch:  make([]*itemWatch, 0),
	}
	db.Register("memory", &m)
}

type itemWatch struct {
	UserID    *uuid.UUID
	ItemID    *uuid.UUID
	CreatedAt time.Time
}

type Memory struct {
	items           []*models.Item
	images          []*models.ItemImage
	bids            []*models.Bid
	users           []*models.User
	profilePictures []*models.ProfilePicture
	userStats       []*models.UserStats
	userPhones      []*models.PhoneNumber
	tokens          []*token.Token
	sessions        []session.Session
	nomail          []*models.NoMail
	userItemsWatch  []*itemWatch
}

func (m *Memory) Open(urlString string) (db.Driver, error) {
	url, err := nurl.Parse(urlString)
	if err != nil {
		return nil, err
	}
	q := url.Query()
	if q.Get("prepopulate") == "true" {
		users := mock.Users()
		items := mock.Items(users)

		bids := mock.Bids(items, users)
		mock.Stats(users, items, bids)
		if err != nil {
			return nil, err
		}
		if m.items == nil {
			m.items = make([]*models.Item, 0)
		}
		if m.users == nil {
			m.users = make([]*models.User, 0)
		}
		if m.images == nil {
			m.images = make([]*models.ItemImage, 0)
		}
		if m.bids == nil {
			m.bids = make([]*models.Bid, 0)
		}
		if m.userStats == nil {
			m.userStats = make([]*models.UserStats, 0)
		}
		if m.userPhones == nil {
			m.userPhones = make([]*models.PhoneNumber, 0)
		}
		for _, user := range users {
			m.profilePictures = append(m.profilePictures, user.Picture)
			m.userPhones = append(m.userPhones, user.PhoneNumbers...)
			m.userStats = append(m.userStats, user.Stats)
		}
		for _, item := range items {
			m.images = append(m.images, item.Images...)
		}
		m.items = items
		m.users = users
		m.bids = bids
	}
	return m, nil
}

func (m *Memory) Close() error {
	m.items = make([]*models.Item, 0)
	m.bids = make([]*models.Bid, 0)
	return nil
}
