package items

import (
	"context"
	"net/http"

	"github.com/djangulo/sfd/db/models"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/gofrs/uuid"
)

type key int

const (
	ItemCtxKey key = iota + 1
	ItemListCtxKey
	ItemBidsCtxKey
	ItemImagesCtxKey
	BidCtxKey
)

func (s *Server) ItemCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var item *models.Item
		var err error

		ctx := r.Context()

		if itemID := chi.URLParam(r, "itemID"); itemID != "" {
			var uid uuid.UUID
			uid, err = uuid.FromString(itemID)
			if err != nil {
				item, err = s.store.GetBySlug(itemID)
				if err != nil {
					render.Render(w, r, ErrInvalidRequest(err))
					return
				}
			} else {
				item, err = s.store.GetItem(ctx, &uid)
			}
		} else {
			render.Render(w, r, ErrNotFound)
			return
		}
		if err != nil {
			render.Render(w, r, ErrNotFound)
			return
		}

		ctx = context.WithValue(ctx, ItemCtxKey, item)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *Server) ItemListCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var items []*models.Item
		var err error
		q := r.URL.Query()
		if query := q.Get("q"); query != "" {
			items, err = s.store.SearchItems(r.Context(), query)
		} else {
			items, err = s.store.ListItems(r.Context())
		}
		if err != nil {
			render.Render(w, r, ErrNotFound)
			return
		}

		ctx := context.WithValue(r.Context(), ItemListCtxKey, items)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *Server) ItemImagesCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var images []*models.ItemImage
		var err error

		ctx := r.Context()

		item := r.Context().Value(ItemCtxKey).(*models.Item)
		images, err = s.store.ItemImages(ctx, &item.ID)
		if err != nil {
			render.Render(w, r, ErrNotFound)
			return
		}

		ctx = context.WithValue(ctx, ItemImagesCtxKey, images)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *Server) ItemBidsCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var bids []*models.Bid
		var err error

		ctx := r.Context()

		item := r.Context().Value(ItemCtxKey).(*models.Item)
		if item.Blind {
			render.Render(w, r, ErrItemIsBlind)
			return
		}

		bids, err = s.store.ItemBids(ctx, &item.ID)
		if err != nil {
			render.Render(w, r, ErrNotFound)
			return
		}

		ctx = context.WithValue(ctx, ItemBidsCtxKey, bids)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// BidCtx fetches the bid under {bidID} and sends it down the request context.
// Renders an error if the item is Blind.
func (s *Server) BidCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		ctx := r.Context()

		item := ctx.Value(ItemCtxKey).(*models.Item)
		if item.Blind {
			render.Render(w, r, ErrItemIsBlind)
			return
		}
		var bid *models.Bid
		var err error

		if bidID := chi.URLParam(r, "bidID"); bidID != "" {
			var uid uuid.UUID
			uid, err = uuid.FromString(bidID)
			if err != nil {
				render.Render(w, r, ErrInvalidRequest(err))
				return
			}
			bid, err = s.store.GetBid(ctx, &uid)
		} else {
			render.Render(w, r, ErrNotFound)
			return
		}
		if err != nil {
			render.Render(w, r, ErrNotFound)
			return
		}

		ctx = context.WithValue(ctx, BidCtxKey, bid)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
