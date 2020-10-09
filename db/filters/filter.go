package filters

import (
	"context"
	"net/http"
	"strings"
	"time"
)

type key int

const CtxKey key = iota + 1

// Options interface returns the LastID and the LastCreated.
// The methods are guaranteed to be non-nil, returning uuid.Nil
// and a zero-value time.Time.
type Options interface {
	Closed() State
	Blind() State
	AdminApproved() State
}

type State int

const (
	True State = 1 << iota
	False
	Off State = True | False
)

func (s State) String() string {
	return [...]string{
		True:  "TRUE",
		False: "FALSE",
		Off:   "NOT NULL",
	}[s]
}

func (s State) On() bool {
	return s != Off
}

func (s State) True() bool {
	return s == True
}

func (s State) False() bool {
	return s == False
}

func NewOptions(opts ...OptionFunc) Options {
	return newOptions(opts...)
}

func newOptions(opts ...OptionFunc) *options {

	o := &options{
		closed:         Off,
		blind:          Off,
		adminApproved:  True,
		deadlineStart:  time.Time{},
		deadlineEnd:    time.Time{},
		publishedStart: time.Time{},
		publishedEnd:   time.Time{},
	}
	for _, opt := range opts {
		opt(o)
	}
	return o
}

type OptionFunc func(Options)

func WithBlind(state State) OptionFunc {
	return func(o Options) {
		u := o.(*options)
		u.blind = state
	}
}
func WithClosed(state State) OptionFunc {
	return func(o Options) {
		u := o.(*options)
		u.closed = state
	}
}
func WithAdminApproved(state State) OptionFunc {
	return func(o Options) {
		u := o.(*options)
		u.adminApproved = state
	}
}

type options struct {
	blind          State
	adminApproved  State
	closed         State
	deadlineStart  time.Time
	deadlineEnd    time.Time
	publishedStart time.Time
	publishedEnd   time.Time
}

func (o *options) Blind() State {
	return o.blind
}

func (o *options) AdminApproved() State {
	return o.adminApproved
}

func (o *options) Closed() State {
	return o.closed
}

func (o *options) Deadline() (time.Time, time.Time) {
	return o.deadlineStart, o.deadlineEnd
}

func (o *options) PublishedAt() (time.Time, time.Time) {
	return o.publishedStart, o.publishedEnd
}

func Context(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()

		var opts = newOptions()

		if s := q.Get("blind"); s != "" {
			s = strings.ToLower(s)
			if s == "1" || s == "on" || s == "true" || s == "t" {
				opts.blind = True
			} else if s == "0" || s == "off" || s == "false" || s == "f" {
				opts.blind = False
			}
		}

		if s := q.Get("closed"); s != "" {
			s = strings.ToLower(s)
			if s == "1" || s == "on" || s == "true" || s == "t" {
				opts.closed = True
			} else if s == "0" || s == "off" || s == "false" || s == "f" {
				opts.closed = False
			}
		}
		if s := q.Get("admin-approved"); s != "" {
			s = strings.ToLower(s)
			if s == "1" || s == "on" || s == "true" || s == "t" {
				opts.adminApproved = True
			} else if s == "0" || s == "off" || s == "false" || s == "f" {
				opts.adminApproved = False
			}
		}
		var o Options
		o = opts

		ctx := context.WithValue(r.Context(), CtxKey, o)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
