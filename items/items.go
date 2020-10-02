package items

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/gofrs/uuid"

	"github.com/djangulo/sfd/db"
	"github.com/djangulo/sfd/db/models"
	"github.com/djangulo/sfd/db/validators"
	sfdErrors "github.com/djangulo/sfd/errors"
	"github.com/djangulo/sfd/storage"
)

type Server struct {
	store db.ItemStorer
	fs    storage.Driver
}

func NewServer(
	storer db.ItemStorer,
	fs storage.Driver) (*Server, error) {

	s := Server{
		store: storer,
		fs:    fs,
	}
	// server.registerRoutes()

	return &s, nil
}

type filters struct {
	addInactive    bool
	addUnpublished bool
}

func (f *filters) Inactive() bool {
	return f.addInactive
}

func (f *filters) Unpublished() bool {
	return f.addUnpublished
}

/*
Api endpoints
	GET:
		/items	list items
		/items/{itemId} 	item details
		/items/{itemId}/bids bids for this item
		/items/{itemId}/images images for this item
		/items/{itemId}/images/cover cover image for this item
		/bids/{itemId} details on a specific bid
	POST:
		/items create an item
		/items/{itemId}/images images for this item
		/items/{itemId}/bids place a bid on this item
		/items/{itemId}/subscribe sub user-id to item
		/items/{itemId}/unsubscribe sub user-id to item

		/bids/{itemId}/invalidate invalidate bid
	PATCH:
		/items/{itemId} update
	DELETE:
		/items/{itemId} delete item
*/

type ItemResponse struct {
	*models.Item
}

func NewItemResponse(item *models.Item) *ItemResponse {
	res := &ItemResponse{Item: item}
	return res
}

func (ir *ItemResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func NewItemListResponse(items []*models.Item) []render.Renderer {
	list := []render.Renderer{}
	for _, item := range items {
		list = append(list, NewItemResponse(item))
	}
	return list
}

type BidResponse struct {
	*models.Bid
}

func (br *BidResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func NewBidResponse(bid *models.Bid) render.Renderer {
	return &BidResponse{Bid: bid}
}

func NewBidListResponse(bids []*models.Bid) []render.Renderer {
	list := []render.Renderer{}
	for _, bid := range bids {
		list = append(list, NewBidResponse(bid))
	}
	return list
}

func GetItem(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	item := ctx.Value(ItemCtxKey).(*models.Item)
	images := ctx.Value(ItemImagesCtxKey).([]*models.ItemImage)

	for _, img := range images {
		if pusher, ok := w.(http.Pusher); ok {
			if img.Path.Valid {
				if img.Path.String != "" {
					contentType := "image/" + strings.ReplaceAll(img.FileExt.String, ".", "")
					options := &http.PushOptions{
						Header: http.Header{
							"Content-Type":    []string{contentType},
							"Accept-Encoding": r.Header["Accept-Encoding"],
						},
					}
					if err := pusher.Push(img.Path.String, options); err != nil {
						render.Render(w, r, ErrInternalServerError(err))
						return
					}
				}
			}
		}
	}

	if err := render.Render(w, r, NewItemResponse(item)); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
}

func ListItems(w http.ResponseWriter, r *http.Request) {

	items := r.Context().Value(ItemListCtxKey).([]*models.Item)

	for _, item := range items {
		if pusher, ok := w.(http.Pusher); ok {
			if item.CoverImage.Path.Valid {
				if item.CoverImage.Path.String != "" {
					contentType := "image/" + strings.ReplaceAll(item.CoverImage.FileExt.String, ".", "")
					options := &http.PushOptions{
						Header: http.Header{
							"Content-Type":    []string{contentType},
							"Accept-Encoding": r.Header["Accept-Encoding"],
						},
					}
					if err := pusher.Push(item.CoverImage.Path.String, options); err != nil {
						render.Render(w, r, ErrInternalServerError(err))
						return
					}
				}
			}
		}
	}
	if err := render.RenderList(w, r, NewItemListResponse(items)); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
}

type ItemRequest struct {
	// Deadline unix timestamp of bid closure datetime.
	Deadline int64 `json:"deadline"`
	// Published unix timestamp of publish date.
	Published         int64  `json:"published"`
	BidIntervalAmount int    `json:"bid_interval_amount"`
	BidIntervalUnit   string `json:"bid_interval_unit"`
	*models.Item
	User *models.User
}

func (ir *ItemRequest) Bind(r *http.Request) error {
	e := validators.NewErrors()
	if ir.Item == nil {
		e.Add("form", fmt.Errorf("missing required Item fields"))
		return e
	}

	if err := validators.NotEmpty(ir.Name); err != nil {
		e.Add("name", err)
	}
	if err := validators.NotNil(ir.Item.OwnerID); err != nil {
		e.Add("owner_id", err)
	}
	if err := validators.Min(int(ir.StartingPrice), 0); err != nil {
		e.Add("starting_price", err)
	}
	if err := validators.Min(int(ir.MinIncrement), 0); err != nil {
		e.Add("min_increment", err)
	}
	if err := validators.Min(int(ir.MaxPrice), 0); err != nil {
		e.Add("max_price", err)
	}
	deadline := time.Unix(ir.Deadline, 0)
	if err := validators.After(deadline, time.Now()); err != nil {
		e.Add("bid_deadline", err)
	}
	publish := time.Unix(ir.Published, 0)
	if err := validators.After(publish, time.Now()); err != nil {
		e.Add("bid_deadline", err)
	}

	if !(ir.BidIntervalUnit == "s" ||
		ir.BidIntervalUnit == "m" ||
		ir.BidIntervalUnit == "h") {
		e.Add("bid_interval", fmt.Errorf("unrecognized unit: %s", ir.BidIntervalUnit))
	}
	if err := validators.Min(ir.BidIntervalAmount, 0); err != nil {
		e.Add("bid_interval", err)
	}
	interval, err := time.ParseDuration(
		fmt.Sprintf("%d%s", ir.BidIntervalAmount, ir.BidIntervalUnit),
	)
	if err != nil {
		e.Add("bid_interval", err)
	}
	ir.Item, err = models.NewItem(
		ir.OwnerID,
		ir.Name,
		ir.Description,
		ir.StartingPrice,
		ir.MaxPrice,
		ir.MinIncrement,
		ir.Blind,
		int(interval.Seconds()),
		deadline,
		publish,
	)
	if err != nil {
		e.Add("form", err)
	}
	if e.Len() > 0 {
		return e
	}
	return nil

}

func (s *Server) CreateItem(w http.ResponseWriter, r *http.Request) {
	data := &ItemRequest{}
	if err := render.Bind(r, data); err != nil {
		e := err.(*validators.Errors)
		render.Render(w, r, e)
		return
	}

	item := data.Item
	if err := s.store.CreateItem(item); err != nil {
		render.Render(w, r, ErrInternalServerError(err))
		return
	}

	render.Status(r, http.StatusCreated)
	if err := render.Render(w, r, NewItemResponse(item)); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
}

type ItemImageResponse struct {
	*models.ItemImage
}

func (iir *ItemImageResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func NewItemImageResponse(image *models.ItemImage) render.Renderer {
	return &ItemImageResponse{ItemImage: image}
}

func NewItemImageListResponse(images []*models.ItemImage) []render.Renderer {
	list := []render.Renderer{}
	for _, image := range images {
		list = append(list, NewItemImageResponse(image))
	}
	return list
}

type ItemImagesRequest struct {
	Count int `json:"count"`
	// Format is an fmt-compatible format for the image count i.e. "images-%d"
	// same format will be appended to find the image order "images-%d-order" and
	// alt-text "images-%d-alt"
	Format string `json:"format"`
}

func (iir *ItemImagesRequest) Bind(r *http.Request) error {
	return nil
}

func (s *Server) AddItemImages(w http.ResponseWriter, r *http.Request) {
	item := r.Context().Value(ItemCtxKey).(*models.Item)
	r.ParseMultipartForm(10 << 20)

	data := &ItemImagesRequest{}
	var e = validators.NewErrors()
	if c := r.FormValue("count"); c != "" {
		n, err := strconv.Atoi(c)
		if err != nil {
			render.Render(w, r, ErrInternalServerError(err))
			return
		}
		if err := validators.Min(n, 1); err != nil {
			e.Add("count", err)
		} else {
			data.Count = n
		}
	}
	if f := r.FormValue("format"); f != "" {
		if err := validators.Regex(f, "%(d|v)"); err != nil {
			e.Add("format", fmt.Errorf("missing \"%%d\" or \"%%v\" in format: %w", err))
		} else {
			data.Format = f
		}
	}
	if e.Len() > 0 {
		render.Render(w, r, e)
		return
	}

	var images = make([]*models.ItemImage, 0)

	for i := 1; i <= data.Count; i++ {
		name := fmt.Sprintf(data.Format, i)
		file, handler, err := r.FormFile(name)
		if err != nil {
			render.Render(w, r, ErrInvalidRequest(err))
			return
		}
		ext := filepath.Ext(handler.Filename)
		if !s.fs.Accepts(ext) {
			render.Render(w, r, ErrInvalidRequest(fmt.Errorf("extension %q not allowed", ext)))
			return
		}
		var order int = i + 1
		if o := r.FormValue(fmt.Sprintf(data.Format+"-order", i)); o != "" {
			n, err := strconv.Atoi(o)
			if err == nil {
				order = n
			}
		}
		var alt string
		if a := r.FormValue(fmt.Sprintf(data.Format+"-alt", i)); a != "" {
			alt = a
		}

		image, err := models.NewItemImage(&item.ID, "", "", handler.Filename, ext, alt, order)
		if err != nil {
			render.Render(w, r, ErrInternalServerError(err))
			return
		}
		image.Path.String = db.ItemImagePath(&item.ID, &image.ID.UUID, ext, s.fs.Root())
		image.Path.Valid = true
		image.AbsPath.String, err = s.fs.AddFile(file, image.Path.String)
		if err != nil {
			render.Render(w, r, ErrInternalServerError(err))
			return
		}
		image.AbsPath.Valid = true

		if err := file.Close(); err != nil {
			render.Render(w, r, ErrInternalServerError(err))
			return
		}

		images = append(images, image)
	}
	if err := s.store.AddImages(&item.ID, images...); err != nil {
		for _, img := range images {
			if err := s.fs.RemoveFile(img.Path.String); err != nil {
				log.Printf("error removing image %s, must proceed manually\n", img.AbsPath.String)
			}
		}
		render.Render(w, r, ErrInternalServerError(err))
		return
	}
	if err := render.RenderList(w, r, NewItemImageListResponse(images)); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}

}

func ItemImages(w http.ResponseWriter, r *http.Request) {
	images := r.Context().Value(ItemImagesCtxKey).([]*models.ItemImage)

	for _, img := range images {
		if pusher, ok := w.(http.Pusher); ok {
			if img.Path.Valid {
				if img.Path.String != "" {
					contentType := "image/" + strings.ReplaceAll(img.FileExt.String, ".", "")
					options := &http.PushOptions{
						Header: http.Header{
							"Content-Type":    []string{contentType},
							"Accept-Encoding": r.Header["Accept-Encoding"],
						},
					}
					if err := pusher.Push(img.Path.String, options); err != nil {
						render.Render(w, r, ErrInternalServerError(err))
						return
					}
				}
			}
		}
	}

	if err := render.RenderList(w, r, NewItemImageListResponse(images)); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
}

func ItemBids(w http.ResponseWriter, r *http.Request) {
	bids := r.Context().Value(ItemBidsCtxKey).([]*models.Bid)

	if err := render.RenderList(w, r, NewBidListResponse(bids)); err != nil {
		fmt.Println(err)
		render.Render(w, r, ErrRender(err))
		return
	}
}

type BidRequest struct {
	*models.Bid
	User *models.User
}

func (br *BidRequest) Bind(r *http.Request) error {
	return nil
}

func (s *Server) PlaceBid(w http.ResponseWriter, r *http.Request) {
	data := &BidRequest{}

	ctx := r.Context()
	item := ctx.Value(ItemCtxKey).(*models.Item)
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	var e = validators.NewErrors()
	// Blind items work stand-alone, i.e. no information is known
	// about the winning bid, just a min bid and a max bid.
	if item.Blind {
		if data.Amount.Gt(item.MaxPrice) {
			e.Add("amount", fmt.Errorf("cannot be greater than %v", item.MaxPrice))
		}
		if data.Amount.Lt(item.StartingPrice) {
			e.Add("amount", fmt.Errorf("cannot be less than %v", item.StartingPrice))
		}
	} else {
		// Non-blind items, on the other hand, have a composite min and max,
		// comprised of winning_bid+min_increment. If winning_bid+min_increment
		// is greater than the maximum, then it's set at the maximum and the item
		// is closed.
		winningBid, err := s.store.ItemWinningBid(&item.ID)
		if err != nil {
			render.Render(w, r, ErrInternalServerError(err))
			return
		}
		if data.Amount.Gt(item.MaxPrice) {
			e.Add("amount", fmt.Errorf("cannot be greater than %v", item.MaxPrice))
		}
		if data.Amount.Lt(winningBid.Amount.Add(item.MinIncrement)) {
			e.Add(
				"amount",
				fmt.Errorf(
					"cannot be less than %v",
					winningBid.Amount.Add(item.MinIncrement),
				),
			)
		}
	}
	if e.Len() > 0 {
		render.Render(w, r, e)
		return
	}

	bid := data.Bid
	bid, err := models.NewBid(bid.ItemID, bid.UserID, bid.Amount)
	if err != nil {
		render.Render(w, r, ErrInternalServerError(err))
		return
	}
	if err := s.store.PlaceBid(bid); err != nil {
		render.Render(w, r, ErrInternalServerError(err))
		return
	}

	render.Status(r, http.StatusCreated)
	if err := render.Render(w, r, NewBidResponse(bid)); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
}

type SubRequest struct {
	UserID *uuid.UUID `json:"userID"`
}

func (sr *SubRequest) Bind(r *http.Request) error {
	var e = validators.NewErrors()
	if err := validators.UUID(sr.UserID); err != nil {
		e.Add("user_id", err...)
	}
	if e.Len() > 0 {
		return e
	}
	return nil
}

type CallBackResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func (cbr *CallBackResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func NewCallBackResponse(status, msg string) *CallBackResponse {
	return &CallBackResponse{Status: status, Message: msg}
}

func (s *Server) WatchItem(w http.ResponseWriter, r *http.Request) {
	item := r.Context().Value(ItemCtxKey).(*models.Item)
	data := &SubRequest{}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	if err := s.store.WatchItem(&item.ID, data.UserID); err != nil {
		render.Render(w, r, ErrInternalServerError(err))
		return
	}

	render.Status(r, http.StatusAccepted)
	if err := render.Render(w, r, NewCallBackResponse(
		"OK",
		fmt.Sprintf("Item %s is now watched", item.Name),
	)); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
}

func (s *Server) UnWatchItem(w http.ResponseWriter, r *http.Request) {
	item := r.Context().Value(ItemCtxKey).(*models.Item)
	data := &SubRequest{}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	if err := s.store.UnWatchItem(&item.ID, data.UserID); err != nil {
		render.Render(w, r, ErrInternalServerError(err))
		return
	}

	render.Status(r, http.StatusAccepted)
	if err := render.Render(w, r, NewCallBackResponse(
		"OK",
		fmt.Sprintf("Item %s is no longer watched", item.Name),
	)); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
}

func GetBid(w http.ResponseWriter, r *http.Request) {
	bid := r.Context().Value(BidCtxKey).(*models.Bid)
	if err := render.Render(w, r, NewBidResponse(bid)); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
}

func (s *Server) InvalidateBid(w http.ResponseWriter, r *http.Request) {
	var err error
	var uid uuid.UUID

	if bidID := chi.URLParam(r, "bidID"); bidID != "" {
		uid, err = uuid.FromString(bidID)
		if err != nil {
			render.Render(w, r, ErrInvalidRequest(err))
			return
		}
		err = s.store.InvalidateBid(&uid)
	} else {
		render.Render(w, r, ErrNotFound)
		return
	}
	if err != nil {
		render.Render(w, r, ErrNotFound)
		return
	}

	render.Status(r, http.StatusAccepted)
	if err := render.Render(w, r, NewCallBackResponse(
		"OK",
		fmt.Sprintf("Successfully invalidated bid %v", uid),
	)); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
}

type WinningBidResponse struct {
	*models.Bid
}

func NewWinningBidResponse(bid *models.Bid) *WinningBidResponse {
	return &WinningBidResponse{Bid: bid}
}

func (wbr *WinningBidResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (s *Server) ItemWinningBid(w http.ResponseWriter, r *http.Request) {
	item := r.Context().Value(ItemCtxKey).(*models.Item)
	var err error
	var bid = new(models.Bid)
	if bid, err = s.store.ItemWinningBid(&item.ID); err != nil {
		render.Render(w, r, sfdErrors.ErrNotFound)
		return
	}

	if err := render.Render(w, r, NewWinningBidResponse(bid)); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}

}
