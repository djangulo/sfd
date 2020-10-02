package pagination

import (
	"context"
	"html/template"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/gofrs/uuid"

	"github.com/djangulo/sfd/db/models"
)

type Reference struct {
	Link string
	Page int
}

func (r *Reference) String() string {
	return r.Link
}

type Pages struct {
	FirstLink  *Reference
	LastLink   *Reference
	TotalPages int
	Options    *models.ListOptions
	Refs       []*Reference
}

func Paginate(path string, count int, opts *models.ListOptions) *Pages {

	q := url.Values{}
	if opts != nil {
		if opts.Inactive {
			q.Set("inactive", "true")
		} else {
			q.Del("inactive")
		}
		if opts.Unpublished {
			q.Set("unpublished", "true")
		} else {
			q.Del("unpublished")
		}
	}
	if opts == nil {
		opts = models.NewOptions()
	}

	total := (int(count) / opts.Limit) + 1
	p := Pages{
		TotalPages: total,
		Options:    opts,
		Refs:       make([]*Reference, total, total),
	}

	u, _ := url.Parse(path)
	u.RawQuery = q.Encode()
	p.FirstLink = &Reference{u.String(), 1}
	q.Set("page", strconv.Itoa(total))
	u.RawQuery = q.Encode()
	p.LastLink = &Reference{u.String(), total}

	p.Refs[0] = p.FirstLink
	for i := 1; i < total; i++ {
		q.Set("page", strconv.Itoa(i+1))
		u.RawQuery = q.Encode()
		p.Refs[i] = &Reference{Link: u.String(), Page: i + 1}
	}
	return &p
}

type key int

const CtxKey key = iota + 1

// Options interface returns the LastID and the LastCreated.
// The methods are guaranteed to be non-nil, returning uuid.Nil
// and a zero-value time.Time.
type Options interface {
	Limit() int
	LastID() *uuid.UUID
	LastCreated() *time.Time
	LastAmount() float64
	LastOrder() int
}

func NewOptions(opts ...OptionFunc) Options {

	o := &options{
		limit:       10,
		lastID:      &uuid.Nil,
		lastCreated: &time.Time{},
		lastAmount:  0.0,
		lastOrder:   1,
	}
	for _, opt := range opts {
		opt(o)
	}
	return o
}

type OptionFunc func(Options)

func WithLimit(n int) OptionFunc {
	if n == 0 {
		n = 10
	}
	return func(o Options) {
		u := o.(*options)
		u.limit = n
	}
}

func WithLastID(id *uuid.UUID) OptionFunc {
	return func(o Options) {
		u := o.(*options)
		u.lastID = id
	}
}

func WithLastCreated(lastCreated *time.Time) OptionFunc {
	return func(o Options) {
		u := o.(*options)
		u.lastCreated = lastCreated
	}
}

func WithLastAmount(lastAmount float64) OptionFunc {
	return func(o Options) {
		u := o.(*options)
		u.lastAmount = lastAmount
	}
}

func WithLastOrder(lastOrder int) OptionFunc {
	return func(o Options) {
		u := o.(*options)
		u.lastOrder = lastOrder
	}
}

type options struct {
	limit       int
	lastOrder   int
	lastID      *uuid.UUID
	lastCreated *time.Time
	lastAmount  float64
}

func (o *options) Limit() int {
	return o.limit
}

func (o *options) LastID() *uuid.UUID {
	return o.lastID
}

func (o *options) LastCreated() *time.Time {
	return o.lastCreated
}

func (o *options) LastAmount() float64 {
	return o.lastAmount
}

func (o *options) LastOrder() int {
	return o.lastOrder
}

func Context(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()

		var limit int = 10
		if n, err := strconv.Atoi(q.Get("limit")); err == nil {
			limit = n
		}

		var lastID *uuid.UUID
		if uid, err := uuid.FromString(q.Get("last-id")); err == nil {
			lastID = &uid
		} else {
			lastID = &uuid.Nil
		}

		var lastCreated *time.Time
		if t, err := time.Parse(time.RFC3339Nano, q.Get("last-created")); err == nil {
			lastCreated = &t
		} else {
			lastCreated = &(time.Time{})
		}

		var lastAmount float64
		if f, err := strconv.ParseFloat(q.Get("last-amount"), 64); err == nil {
			lastAmount = f
		} else {
			lastAmount = 0.0
		}

		var lastOrder int
		if n, err := strconv.Atoi(q.Get("last-order")); err == nil {
			lastOrder = n
		} else {
			lastOrder = 1
		}

		var opts Options
		opts = &options{
			limit:       limit,
			lastID:      lastID,
			lastCreated: lastCreated,
			lastAmount:  lastAmount,
			lastOrder:   lastOrder,
		}

		ctx := context.WithValue(r.Context(), CtxKey, opts)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func FuncMap() template.FuncMap {
	return template.FuncMap{
		"displayPage": func(index, currentPage int) bool {
			index++ // index starts at 0
			if math.Abs(float64(currentPage-index)) == 0 {
				return true
			} else if math.Abs(float64(currentPage-index)) <= 3 {
				return true
			}
			return false
		},
		"paginationDots": func(value, currentPage int) bool {
			if math.Abs(float64(currentPage-value)) > 3 {
				return true
			}
			return false
		},
	}
}
