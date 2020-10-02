package db

import (
	"context"
	"time"

	"github.com/gofrs/uuid"

	"github.com/djangulo/sfd/crypto/session"
	"github.com/djangulo/sfd/crypto/token"
	"github.com/djangulo/sfd/db/models"
)

// BidStorer holds the Bids related storage methods.
type ItemStorer interface {
	GetItem(ctx context.Context, itemID *uuid.UUID) (*models.Item, error)
	GetBid(ctx context.Context, bidID *uuid.UUID) (*models.Bid, error)
	// CreateItem creates a new item.
	CreateItem(item *models.Item) error
	// PublishItem sets the PublishedAt property on an item.
	PublishItem(itemID *uuid.UUID, datetime time.Time) error
	// GetBySlug returns an item by its unique slug.
	GetBySlug(slug string) (*models.Item, error)
	// ListItems returns a list of items.
	ListItems(ctx context.Context) ([]*models.Item, error)
	// ItemBids gets a list of bids for an item.
	ItemBids(ctx context.Context, itemID *uuid.UUID) ([]*models.Bid, error)
	WatchItem(itemID, userID *uuid.UUID) error
	UnWatchItem(itemID, userID *uuid.UUID) error
	ItemWinningBid(itemID *uuid.UUID) (*models.Bid, error)

	// // CoverImage returns the lowest-order image of an item.
	// CoverImage(itemID *uuid.UUID) ([]*models.ItemImage, error)
	// ItemImages returns the images belonging to an item.
	ItemImages(ctx context.Context, itemID *uuid.UUID) ([]*models.ItemImage, error)
	// DeleteItem soft deletes an Item by ID.
	DeleteItem(id *uuid.UUID) error
	// PlaceBid creates a models.Bid and places the bid on the Bid.ItemID.
	PlaceBid(bid *models.Bid) error
	// InvalidateBid invalidates a bid by ID.
	InvalidateBid(bidID *uuid.UUID) error
	// BulkInvalidateBids invalidate bids in bulk.
	BulkInvalidateBids(ids ...*uuid.UUID) error
	// BidsByID
	BidsByID(ids ...*uuid.UUID) ([]*models.Bid, error)
	// AddImages adds an image to the item.
	AddImages(itemID *uuid.UUID, images ...*models.ItemImage) error
	// RemoveImage removes an image to the item.
	RemoveImage(id *uuid.UUID) error
	UserBids(userID *uuid.UUID, opts *models.ListOptions) (int, []*models.Bid, error)
	UserItemBids(ctx context.Context, userID *uuid.UUID, itemID *uuid.UUID) ([]*models.Bid, error)
}

// AccountStorer accounts related storage methods: Actions that users can
// perform on their own accounts or using the registration system.
type AccountStorer interface {
	// CreateUser creates a new user.
	CreateUser(user *models.User) error
	// UserByUsernameOrEmail returns a user by its username or email.
	UserByUsernameOrEmail(username string) (*models.User, error)
	// UserByID returns a user by its username.
	UserByID(id *uuid.UUID) (*models.User, error)

	// ChangePassword replaces the user's PasswordHash.
	ChangePassword(user *models.User, newPasswordHash string) (*models.User, error)
	// ResetPassword replaces the user's PasswordHash. Users resetting their passwords
	// are most likely logged off, so no need to return the User object.
	ResetPassword(userID *uuid.UUID, newPasswordHash string) error
	LoginUser(user *models.User) (*models.User, error)
	// RemoveProfilePic removes the user's profile picture.
	RemoveProfilePic(userID *uuid.UUID) error
	AddProfilePic(userID *uuid.UUID, image *models.ProfilePicture) error
	AddPhoneNumbers(userID *uuid.UUID, phoneNumbers ...*models.PhoneNumber) error
	RemovePhoneNumber(numberID *uuid.UUID) error
	UnsubEmail(email string, kind models.NoMailKind) error
	ResubEmail(email string, kind models.NoMailKind) error
	IsUnsub(email string, kind models.NoMailKind) bool
}

// AdminStorer admin-related store functions.
type AdminStorer interface {
	// ActivateUser user activates a user in system (allowing authenticated access).
	ActivateUser(userID, approverID *uuid.UUID) error
	DeactivateUser(userID *uuid.UUID) error
	GrantAdmin(userID *uuid.UUID) error
	RevokeAdmin(userID *uuid.UUID) error
	ListUsers(opts *models.ListOptions) (int, []*models.User, error)
}

// Driver all storage engines need to implement this interface.
type Driver interface {
	// Open creates a connection to the driver.
	Open(url string) (Driver, error)
	// Close closes the driver connection.
	Close() error
	ItemStorer
	AccountStorer
	AdminStorer
	session.Storer
	token.Storer
}
