//Package models holds common models, sorting interfaces, and errors for the app.
package models

import (
	"database/sql"
	"encoding/gob"
	"errors"
	"io"
	"regexp"
	"strings"
	"time"

	"github.com/djangulo/sfd/config"
	"github.com/gofrs/uuid"
)

var globalConfig config.Configurer

func init() {
	gob.Register(User{})
	globalConfig = config.Get()
}

type DBObj struct {
	ID        uuid.UUID `json:"id,omitempty" db:"id"`
	CreatedAt time.Time `json:"created_at,omitempty" db:"created_at"`
	UpdatedAt NullTime  `json:"updated_at,omitempty" db:"updated_at"`
	DeletedAt NullTime  `json:"deleted_at,omitempty" db:"deleted_at"`
}

type Translation struct {
	Lang        Language   `json:"lang" db:"lang"`
	Name        string     `json:"name" db:"name"`
	Slug        string     `json:"slug" db:"slug"`
	Description string     `json:"description" db:"description"`
	ItemID      *uuid.UUID `json:"item_id" db:"item_id"`
}

// Item struct holds an item's data. Note that amounts that refer to currency
// are unsigned integers * 100 (to retain decimal precision).
type Item struct {
	*DBObj
	OwnerID *uuid.UUID `json:"owner_id" db:"owner_id"`
	// Name item's name. Unique.
	Name string `json:"name" db:"name"`
	// Slug browser friendly representation. e.g. My-Item becomes my-item.
	Slug string `json:"slug" db:"slug"`
	// Description item description.
	Description string `json:"description" db:"description"`
	// StartingPrice starting price for item bids.
	StartingPrice Currency `json:"starting_price" db:"starting_price"`
	// MaxPrice maximum acceptable price. If set to -1 then there is no limit.
	MaxPrice Currency `json:"max_price" db:"max_price"`
	// MinIncrement next bid placed has to be at least WinningBid+MinIncrement
	// if the item bids are not blind. If the item is Blind, this value is
	// ignored.
	MinIncrement Currency `json:"min_increment" db:"min_increment"`
	// BidInterval time interval in seconds between subsequent bids
	// for a user. e.g. User1 can only place bids BidInterval seconds apart,
	// but User1 and User2 can place bids immediately, as long as they haven't
	// placed any bids since BidInterval seconds ago.
	BidInterval int `json:"bid_interval" db:"bid_interval"`
	// BidDeadline time when bids close and no more bids are accepted.
	BidDeadline time.Time `json:"bid_deadline" db:"bid_deadline"`
	Closed      bool      `json:"closed" db:"closed"`
	// Blind are the bids in this item hidden from everyone.
	Blind bool `json:"blind" db:"blind"`
	// PublishedAt item publication date. Nullable.
	PublishedAt   NullTime   `json:"published_at" db:"published_at"`
	UserNotified  bool       `json:"user_notified" db:"user_notified"`
	AdminApproved bool       `json:"admin_approved" db:"admin_approved"`
	CoverImage    *ItemImage `json:"cover_image" db:"cover_image"`
	// Images list of url to image queries /items/{id}/images/{imageID}
	Images       []*ItemImage   `json:"images"`
	Owner        *User          `json:"-" db:"owner"`
	Bids         []*Bid         `json:"bids"`
	WinningBid   *Bid           `json:"winning_bid" db:"winning_bid"`
	Rank         float64        `json:"rank" db:"rank"`
	Translations []*Translation `json:"translations"`
	// CoverImage image that will be shown in thumbnails and lists.
	// Images     []*ItemImage `json:"images"`
}

func NewItem(
	userID *uuid.UUID,
	name, description string,
	startingPrice, maxPrice, minIncrement Currency,
	blind bool,
	bidInterval int,
	bidDeadline, publishDate time.Time) (*Item, error) {
	item := &Item{
		DBObj: &DBObj{
			ID:        uuid.Must(uuid.NewV4()),
			CreatedAt: time.Now().In(globalConfig.TimeZone()),
			UpdatedAt: NullTime{NullTime: &sql.NullTime{Valid: false}},
			DeletedAt: NullTime{NullTime: &sql.NullTime{Valid: false}},
		},
		Name:          name,
		Description:   description,
		Slug:          Slugify(name),
		OwnerID:       userID,
		StartingPrice: startingPrice,
		MaxPrice:      maxPrice,
		MinIncrement:  minIncrement,
		WinningBid:    &Bid{},
		CoverImage:    &ItemImage{File: &File{}, ItemID: &uuid.Nil},
		BidDeadline:   bidDeadline,
		PublishedAt:   NewNullTime(publishDate),
		Blind:         blind,
	}

	return item, nil
}

type Image interface {
	Data() io.Reader
	Obj() *File
}

// File retains file properties to save in the filesystem. Most fields
// are nullable because while images are nullable, views return them with
// the object they belong to.
type File struct {
	ID        uuid.NullUUID `json:"id,omitempty" db:"id"`
	CreatedAt time.Time     `json:"created_at,omitempty" db:"created_at"`
	UpdatedAt NullTime      `json:"updated_at,omitempty" db:"updated_at"`
	DeletedAt NullTime      `json:"deleted_at,omitempty" db:"deleted_at"`
	// Path relative path to asset on fileserver.
	Path NullString `json:"path,omitempty" db:"path"`
	// AbsPath absolute path to image.
	AbsPath NullString `json:"abs_path,omitempty" db:"abs_path"`
	// OriginalFilename the image's original filename.
	OriginalFilename NullString `json:"original_filename,omitempty" db:"original_filename"`
	FileExt          NullString `json:"file_ext,omitempty" db:"file_ext"`
	AltText          NullString `json:"alt_text,omitempty" db:"alt_text"`
	Order            NullInt64  `json:"order,omitempty" db:"order"`
}

type ItemImage struct {
	*File
	ItemID *uuid.UUID `json:"item_id" db:"item_id"`
}

// NewItemImage creates a new ItemImage for Item with itemID.
// Its Path and AbsPath have to be set after creation, as they are not known
// untile the storage driver creates them.
func NewItemImage(
	itemID *uuid.UUID,
	path, abspath, originalFilename, ext, altText string,
	order int) (*ItemImage, error) {
	id, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}
	now := time.Now().In(globalConfig.TimeZone())
	img := &ItemImage{
		File: &File{
			ID:               uuid.NullUUID{Valid: true, UUID: id},
			CreatedAt:        now,
			Path:             NewNullString(path),
			AbsPath:          NewNullString(abspath),
			OriginalFilename: NewNullString(originalFilename),
			FileExt:          NewNullString(ext),
			AltText:          NewNullString(altText),
			Order:            NewNullInt64(order),
		},
		ItemID: itemID,
	}

	return img, nil
}

type Bid struct {
	*DBObj
	Amount Currency   `json:"amount,omitempty" db:"amount" form:"amount"`
	UserID *uuid.UUID `json:"user_id,omitempty" db:"user_id" form:"user_id"`
	ItemID *uuid.UUID `json:"item_id,omitempty" db:"item_id" form:"item_id"`
	Valid  bool       `json:"valid,omitempty" db:"valid" form:"valid"`
	User   *User      `json:"user,omitempty" db:"user" form:"user"`
	Item   *Item      `json:"item,omitempty" db:"item" form:"item"`
}

func NewBid(itemID, userID *uuid.UUID, amount Currency) (*Bid, error) {
	id, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}
	bid := &Bid{
		DBObj: &DBObj{
			ID:        id,
			CreatedAt: time.Now(),
			UpdatedAt: NullTime{NullTime: &sql.NullTime{Valid: false}},
		},
		ItemID: itemID,
		UserID: userID,
		Amount: amount,
		Valid:  true,
	}
	return bid, nil
}

func FilterValid(bids []*Bid) []*Bid {
	out := make([]*Bid, 0)
	for _, bid := range bids {
		if bid.Valid {
			out = append(out, bid)
		}
	}
	BidsOrderedBy(SortBidsByAmountDesc).Sort(out)
	return out
}

func FilterUserBids(bids []*Bid, userID *uuid.UUID) []*Bid {
	out := make([]*Bid, 0)
	for _, bid := range bids {
		if *bid.UserID == *userID {
			out = append(out, bid)
		}
	}
	BidsOrderedBy(SortBidsByAmountDesc).Sort(out)
	return out
}

func FilterItemBids(bids []*Bid, itemID *uuid.UUID) []*Bid {
	out := make([]*Bid, 0)
	for _, bid := range bids {
		if *bid.ItemID == *itemID {
			out = append(out, bid)
		}
	}
	BidsOrderedBy(SortBidsByAmountDesc).Sort(out)
	return out
}

func Winner(bids []*Bid) *Bid {
	var bid = &Bid{}
	if len(bids) == 0 {
		return bid
	}
	for _, b := range bids {
		if b.Valid && b.Amount.Gte(bid.Amount) {
			*bid = *b
		}
	}
	return bid
}

type User struct {
	*DBObj
	Username      string           `json:"username,omitempty" db:"username" form:"username"`
	Email         string           `json:"email,omitempty" db:"email" form:"email"`
	FullName      string           `json:"full_name,omitempty" db:"full_name" form:"full_name"`
	PasswordHash  string           `json:"password_hash,omitempty" db:"password_hash" form:"password_hash"`
	IsAdmin       bool             `json:"is_admin,omitempty" db:"is_admin" form:"is_admin"`
	Active        bool             `json:"active,omitempty" db:"active" form:"active"`
	LastLogin     NullTime         `json:"last_login,omitempty" db:"last_login"`
	Picture       *ProfilePicture  `json:"picture,omitempty" db:"picture"`
	Stats         *UserStats       `json:"stats,omitempty" db:"stats"`
	ProfilePublic bool             `json:"profile_public,omitempty" db:"profile_public"`
	Preferences   *UserPreferences `json:"preferences,omitempty" db:"preferences"`
	Addresses     []*Address       `json:"addresses,omitempty" db:"addresses"`
	PhoneNumbers  []*PhoneNumber   `json:"phone_numbers,omitempty" db:"phone_numbers"`
}

func NewUser(
	username, email, passwordHash, preferredLanguage string,
	profilePublic bool) (*User, error) {
	id, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}
	bid := &User{
		DBObj: &DBObj{
			ID:        id,
			CreatedAt: time.Now(),
			UpdatedAt: NewNullTime(time.Time{}),
			DeletedAt: NewNullTime(time.Time{}),
		},
		Username:      username,
		Email:         email,
		PasswordHash:  passwordHash,
		ProfilePublic: profilePublic,
		Picture:       &ProfilePicture{UserID: &id},
		Stats:         &UserStats{UserID: &id},
		Preferences:   &UserPreferences{UserID: &id, Language: preferredLanguage},
		Addresses:     make([]*Address, 0),
		PhoneNumbers:  make([]*PhoneNumber, 0),
		IsAdmin:       false,
		Active:        false,
	}
	return bid, nil
}

type PhoneNumber struct {
	ID     uuid.UUID  `json:"id" db:"id"`
	UserID *uuid.UUID `json:"user_id" db:"user_id"`
	Number string     `json:"number" db:"number"`
}

type AddressKind int

func (a AddressKind) String() string {
	return addresses[a]
}

const (
	Billing AddressKind = iota + 1
	Shipping
)

var addresses = []string{
	Billing:  "Billing",
	Shipping: "Shipping",
}

type Address struct {
	*DBObj
	UserID  *uuid.UUID  `json:"user_id" db:"user_id"`
	Address string      `json:"address" db:"address"`
	Kind    AddressKind `json:"kind" db:"kind"`
}

type UserStats struct {
	UserID       *uuid.UUID `json:"user_id" db:"user_id"`
	LoginCount   int        `json:"login_count" db:"login_count"`
	ItemsCreated int        `json:"items_created" db:"items_created"`
	BidsCreated  int        `json:"bids_created" db:"bids_created"`
	BidsWon      int        `json:"bids_won" db:"bids_won"`
}

type ColorPallette int

const (
	Colorblind ColorPallette = iota + 1
	Monochromatic
)

var themes = []string{
	Colorblind:    "Colorblind",
	Monochromatic: "Monocromatic",
}

type UserPreferences struct {
	UserID     *uuid.UUID    `json:"user_id" db:"user_id"`
	Language   string        `json:"language" db:"language"`
	ColorTheme ColorPallette `json:"color_theme" db:"color_theme"`
}

func (u *User) String() string {
	return u.Username
}

type ProfilePicture struct {
	*File
	UserID *uuid.UUID `json:"user_id" db:"user_id"`
}

func NewProfilePicture(
	userID *uuid.UUID,
	path, absPath, originalFilename, fileExt, altText string,
) (*ProfilePicture, error) {
	id, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}
	ppic := ProfilePicture{
		File: &File{
			ID:               uuid.NullUUID{Valid: true, UUID: id},
			CreatedAt:        time.Now(),
			Path:             NewNullString(path),
			AbsPath:          NewNullString(absPath),
			OriginalFilename: NewNullString(originalFilename),
			FileExt:          NewNullString(fileExt),
			AltText:          NewNullString(altText),
			Order:            NewNullInt64(1),
		},
		UserID: userID,
	}
	return &ppic, nil

}

var (
	reClean      = regexp.MustCompile(`[^\w\s\-]`)
	diacriticsRe = map[string]*regexp.Regexp{
		"a": regexp.MustCompile(`[áàâäåā]`),
		"A": regexp.MustCompile(`[ÁÀÂÄÅĀ]`),
		"e": regexp.MustCompile(`[ééêëē]`),
		"E": regexp.MustCompile(`[ÉÉÊËĒ]`),
		"i": regexp.MustCompile(`[íìîïī]`),
		"I": regexp.MustCompile(`[ÍÌÎÏĪ]`),
		"o": regexp.MustCompile(`[óòôöō]`),
		"O": regexp.MustCompile(`[ÓÒÔÖŌ]`),
		"u": regexp.MustCompile(`[úùûüūů]`),
		"U": regexp.MustCompile(`[ÚÙÛÜŪŮ]`),
		"c": regexp.MustCompile(`[ç]`),
	}
	reNoSpaces = regexp.MustCompile(`[-\s]+`)
)

func Slugify(str string) string {
	// replace all diacritic vowels by their ascii counterpart
	for k, re := range diacriticsRe {
		str = re.ReplaceAllString(str, k)
	}
	str = reClean.ReplaceAllString(str, " ")
	str = strings.ToLower(strings.TrimSpace(str))
	return reNoSpaces.ReplaceAllString(str, "-")
}

type NoMailKind int

const (
	RegistrationUnsub NoMailKind = iota + 1
	NotificationUnsub
	Newsletter
	None
)

var noMailKinds = []string{
	RegistrationUnsub: "Registration",
	NotificationUnsub: "Notification",
	Newsletter:        "Newsletter",
	None:              "None",
}

func (a NoMailKind) String() string {
	return addresses[a]
}

type NoMail struct {
	Email string
	Kind  NoMailKind
}

// errors

var (
	// ErrNotFound not found.
	ErrNotFound = errors.New("not found")
	// ErrInvalidInput invalid input.
	ErrInvalidInput = errors.New("invalid input")
	// ErrNoResults no results.
	ErrNoResults = errors.New("no results")
	// ErrNilPointer nil pointer passed.
	ErrNilPointer = errors.New("nil pointer passed")
	// ErrAlreadyExists already exists.
	ErrAlreadyExists = errors.New("already exists")
)

type ListOptions struct {
	Limit       int
	Offset      int
	Page        int
	Unpublished bool
	Inactive    bool
}

var defaultPageSize int

func init() {
	cnf := config.Get()
	defaultPageSize = cnf.PageSize()
}

func NewOptions(opts ...Option) *ListOptions {

	o := &ListOptions{
		Limit: defaultPageSize,
		Page:  1,
	}
	for _, opt := range opts {
		opt(o)
	}
	return o
}

type Option func(*ListOptions)

func WithLimit(n int) Option {
	if n == 0 {
		n = defaultPageSize
	}
	return func(o *ListOptions) {
		o.Limit = n
	}
}

func WithOffset(n int) Option {
	return func(o *ListOptions) {
		o.Offset = n
	}
}

func WithUnpublished(n int) Option {
	return func(o *ListOptions) {
		o.Unpublished = true
	}
}

func WithInactive() Option {
	return func(o *ListOptions) {
		o.Inactive = true
	}
}
func WithPage(page int) Option {
	return func(o *ListOptions) {
		o.Page = page
	}
}
