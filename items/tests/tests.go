package tests

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi"

	_ "github.com/djangulo/sfd/db/memory"
	"github.com/djangulo/sfd/db/mock"
	"github.com/djangulo/sfd/db/models"
	pkg "github.com/djangulo/sfd/items"
	"github.com/djangulo/sfd/pagination"
	_ "github.com/djangulo/sfd/storage/fs"
	testutils "github.com/djangulo/sfd/testing"
)

var (
	users []*models.User
	items []*models.Item
	bids  []*models.Bid
)

func init() {
	users = mock.Users()
	items = mock.Items(users)
	bids = mock.Bids(items, users)
}

func Test(t *testing.T, s *pkg.Server) {
	t.Run("TestListItems", func(t *testing.T) { TestListItems(t, s) })
	t.Run("TestGetItem", func(t *testing.T) { TestGetItem(t, s) })
	t.Run("TestCreateItem", func(t *testing.T) { TestCreateItem(t, s) })
	t.Run("TestItemBids", func(t *testing.T) { TestItemBids(t, s) })
	t.Run("TestPlaceBid", func(t *testing.T) { TestPlaceBid(t, s) })
	t.Run("TestWatchItem", func(t *testing.T) { TestWatchItem(t, s) })
	t.Run("TestUnWatchItem", func(t *testing.T) { TestUnWatchItem(t, s) })
	t.Run("TestInvalidateBid", func(t *testing.T) { TestInvalidateBid(t, s) })
	t.Run("TestItemImages", func(t *testing.T) { TestItemImages(t, s) })
	t.Run("TestAddItemImages", func(t *testing.T) { TestAddItemImages(t, s) })
}

func TestListItems(t *testing.T, s *pkg.Server) {

	t.Run("can get list of items", func(t *testing.T) {
		chain := s.ItemListCtx(http.HandlerFunc(pkg.ListItems))
		ts := httptest.NewServer(chain)
		defer ts.Close()

		res, err := http.Get(ts.URL)
		if err != nil {
			t.Fatal(err)
		}
		testutils.AssertJSON(t, res)
		var list = make([]*models.Item, 0)
		testutils.DecodeJSON(t, res, &list)
		testutils.AssertItemCollectionsEqual(t, list, items[:10])
	})
	t.Run("paginated", func(t *testing.T) {
		chain := pagination.Context(
			s.ItemListCtx(http.HandlerFunc(pkg.ListItems)),
		)
		ts := httptest.NewServer(chain)
		defer ts.Close()

		u, err := url.Parse(ts.URL)
		if err != nil {
			t.Fatal(err)
		}
		q := url.Values{}
		q.Add("last-id", items[5].ID.String())
		q.Add("last-created", items[5].CreatedAt.Format(time.RFC3339Nano))
		q.Add("limit", "5")
		u.RawQuery = q.Encode()

		res, err := http.Get(u.String())
		if err != nil {
			t.Fatal(err)
		}
		var list = make([]*models.Item, 0)
		testutils.AssertJSON(t, res)
		testutils.DecodeJSON(t, res, &list)
		testutils.AssertItemCollectionsEqual(t, list, items[6:11])
	})
}

func TestGetItem(t *testing.T, s *pkg.Server) {

	t.Run("can get item", func(t *testing.T) {
		chain := s.ItemCtx(s.ItemImagesCtx(http.HandlerFunc(pkg.GetItem)))
		// need the router for the chi.URLParam extraction
		r := chi.NewRouter()
		r.Get("/{itemID}", chain.ServeHTTP)

		// need TLS for the http.Push
		ts := httptest.NewTLSServer(r)
		defer ts.Close()

		client := ts.Client()
		res, err := client.Get(ts.URL + "/" + items[0].ID.String())
		if err != nil {
			t.Fatal(err)
		}
		var item = new(models.Item)
		testutils.AssertJSON(t, res)
		testutils.DecodeJSON(t, res, item)
		testutils.AssertItemsEqual(t, item, items[0])
	})
}

func TestCreateItem(t *testing.T, s *pkg.Server) {
	chain := http.HandlerFunc(s.CreateItem)
	ts := httptest.NewServer(chain)
	defer ts.Close()
	t.Run("can create", func(t *testing.T) {

		deadline := time.Now().Add(1000 * time.Minute)
		published := time.Now().Add(500 * time.Second)
		req := pkg.ItemRequest{
			Deadline:          deadline.Unix(),
			Published:         published.Unix(),
			BidIntervalAmount: 60,
			BidIntervalUnit:   "s",
			Item: &models.Item{
				Name:          "Test Item created",
				OwnerID:       &users[1].ID,
				StartingPrice: models.MustCurrency(10.00),
				MinIncrement:  models.MustCurrency(10.00),
				MaxPrice:      models.MustCurrency(1500.00),
			},
		}

		b, err := json.Marshal(req)
		if err != nil {
			t.Fatal(err)
		}

		res, err := http.Post(
			ts.URL,
			"application/json; charset=utf-8",
			bytes.NewReader(b),
		)
		if err != nil {
			t.Fatal(err)
		}
		var item = new(models.Item)
		testutils.AssertJSON(t, res)
		testutils.DecodeJSON(t, res, item)

		if req.Item.Name != item.Name {
			t.Errorf("expected Name %v got %v", req.Item.Name, item.Name)
		}
		if *req.Item.OwnerID != *item.OwnerID {
			t.Errorf("expected OwnerID %v got %v", *req.Item.OwnerID, *item.OwnerID)
		}
		if req.Item.StartingPrice.Neq(item.StartingPrice) {
			t.Errorf("expected StartingPrice %v got %v", req.Item.StartingPrice, item.StartingPrice)
		}
		if req.Item.MaxPrice.Neq(item.MaxPrice) {
			t.Errorf("expected MaxPrice %v got %v", req.Item.MaxPrice, item.MaxPrice)
		}
		if req.Item.MinIncrement.Neq(item.MinIncrement) {
			t.Errorf("expected MinIncrement %v got %v", req.Item.MinIncrement, item.MinIncrement)
		}
		newInterval, _ := time.ParseDuration(
			fmt.Sprintf("%d%s", req.BidIntervalAmount, req.BidIntervalUnit),
		)
		if int(newInterval.Seconds()) == item.BidInterval {
			t.Errorf("expected BidInterval %v got %v", newInterval, item.BidInterval)
		}
		newDeadline := time.Unix(req.Deadline, 0)
		if !newDeadline.Equal(item.BidDeadline) {
			t.Errorf("expected Deadline %v got %v", newDeadline, item.BidDeadline)
		}
	})
	t.Run("errors", func(t *testing.T) {
		// okDeadline := time.Now().Add(1000 * time.Minute)
		// okPublished := time.Now().Add(500 * time.Second)
		for _, tt := range []struct {
			name string
			in   pkg.ItemRequest
			want string
		}{
			{"all fields missing", pkg.ItemRequest{}, "missing required"},
			{"empty name", pkg.ItemRequest{Item: &models.Item{Name: ""}}, "cannot be empty"},
			{"nil ownerID", pkg.ItemRequest{Item: &models.Item{OwnerID: nil}}, "pointer is nil"},
			{"start < 0", pkg.ItemRequest{Item: &models.Item{StartingPrice: models.MustCurrency(-100)}}, "less than 0"},
			{"increment < 0", pkg.ItemRequest{Item: &models.Item{MinIncrement: models.MustCurrency(-100)}}, "less than 0"},
			{"max < 0", pkg.ItemRequest{Item: &models.Item{MaxPrice: models.MustCurrency(-100)}}, "less than 0"},
			{"publish before now", pkg.ItemRequest{Published: time.Now().Add(-10 * time.Second).Unix(), Item: &models.Item{}}, "cannot be before"},
			{"deadline before now", pkg.ItemRequest{Published: time.Now().Add(-10 * time.Second).Unix(), Item: &models.Item{}}, "cannot be before"},
			{"unrecognized unit", pkg.ItemRequest{BidIntervalUnit: "d", Item: &models.Item{}}, "unrecognized unit"},
			{"interval amount < 0", pkg.ItemRequest{BidIntervalAmount: -10, Item: &models.Item{}}, "less than 0"},
		} {
			b, err := json.Marshal(tt.in)
			if err != nil {
				t.Fatal(err)
			}

			res, err := http.Post(
				ts.URL,
				"application/json; charset=utf-8",
				bytes.NewReader(b),
			)
			if err != nil {
				t.Fatal(err)
			}
			testutils.AssertJSON(t, res)

			jsonB, err := ioutil.ReadAll(res.Body)
			if err != nil {
				t.Fatal(err)
			}
			defer res.Body.Close()
			if !strings.Contains(string(jsonB), tt.want) {
				t.Errorf("expected %q to contain %q", string(jsonB), tt.want)
			}
		}
	})
}

func TestItemBids(t *testing.T, s *pkg.Server) {

	item0Bids := make([]*models.Bid, 0)
	for _, b := range bids {
		if *b.ItemID == items[0].ID {
			item0Bids = append(item0Bids, b)
		}
	}

	t.Run("can get list of bids", func(t *testing.T) {
		// need the router for the chi.URLParam extraction
		chain := s.ItemCtx(s.ItemBidsCtx(http.HandlerFunc(pkg.ItemBids)))
		r := chi.NewRouter()
		r.Get("/{itemID}/bids", chain.ServeHTTP)
		ts := httptest.NewServer(r)
		defer ts.Close()

		res, err := http.Get(ts.URL + "/" + items[0].ID.String() + "/bids")
		if err != nil {
			t.Fatal(err)
		}
		testutils.AssertJSON(t, res)
		var list = make([]*models.Bid, 0)
		testutils.DecodeJSON(t, res, &list)
		testutils.AssertBidCollectionsEqual(t, list, item0Bids[:10])
	})
	t.Run("paginated", func(t *testing.T) {
		chain := pagination.Context(
			s.ItemCtx(s.ItemBidsCtx(http.HandlerFunc(pkg.ItemBids))),
		)
		r := chi.NewRouter()
		r.Get("/{itemID}/bids", chain.ServeHTTP)
		ts := httptest.NewServer(r)
		defer ts.Close()

		u, err := url.Parse(ts.URL + "/" + items[0].ID.String() + "/bids")
		if err != nil {
			t.Fatal(err)
		}
		q := url.Values{}
		q.Add("last-amount", item0Bids[5].Amount.String())
		q.Add("limit", "5")
		u.RawQuery = q.Encode()

		res, err := http.Get(u.String())
		if err != nil {
			t.Fatal(err)
		}
		var list = make([]*models.Bid, 0)
		testutils.AssertJSON(t, res)
		testutils.DecodeJSON(t, res, &list)
		testutils.AssertBidCollectionsEqual(t, list, item0Bids[6:11])
	})
}

func TestPlaceBid(t *testing.T, s *pkg.Server) {
	getItem := func(blind bool) *models.Item {
		for _, item := range items {
			if item.Blind == blind {
				return item
			}
		}
		return nil
	}
	findHighest := func(item *models.Item) *models.Bid {
		b := make([]*models.Bid, 0)
		for _, bid := range bids {
			if *bid.ItemID == item.ID {
				b = append(b, bid)
			}
		}
		models.BidsOrderedBy(models.SortBidsByAmountDesc).Sort(b)
		return b[0]
	}
	chain := s.ItemCtx(http.HandlerFunc(s.PlaceBid))
	r := chi.NewRouter()
	r.Post("/{itemID}/bids", chain.ServeHTTP)
	ts := httptest.NewServer(r)
	defer ts.Close()

	t.Run("blind item", func(t *testing.T) {
		t.Run("success", func(t *testing.T) {
			item := getItem(true)
			amount := item.StartingPrice.Add(1000)
			req := pkg.BidRequest{
				Bid: &models.Bid{
					ItemID: &item.ID,
					Amount: amount,
					UserID: &users[0].ID,
					Valid:  true,
				},
			}

			b, err := json.Marshal(req)
			if err != nil {
				t.Fatal(err)
			}

			res, err := http.Post(
				ts.URL+"/"+item.ID.String()+"/bids",
				"application/json; charset=utf-8",
				bytes.NewReader(b),
			)
			if err != nil {
				t.Fatal(err)
			}

			var bid = new(models.Bid)
			testutils.AssertJSON(t, res)
			testutils.DecodeJSON(t, res, bid)

			if req.Bid.Amount.Neq(bid.Amount) {
				t.Errorf("expected Amount %v got %v", req.Bid.Amount, bid.Amount)
			}
			if *req.Bid.UserID != *bid.UserID {
				t.Errorf("expected UserID %v got %v", *req.Bid.UserID, *bid.UserID)
			}
			if *req.Bid.ItemID != *bid.ItemID {
				t.Errorf("expected ItemID %v got %v", req.Bid.ItemID, bid.ItemID)
			}
			if !bid.Valid {
				t.Error("expected Valid to be true but is false")
			}
		})
		t.Run("errors", func(t *testing.T) {
			item := getItem(true)
			for _, tt := range []struct {
				name   string
				item   *models.Item
				amount models.Currency
				want   string
			}{
				{
					"lower than min",
					item,
					item.StartingPrice.Sub(1000),
					"cannot be less than",
				},
				{
					"greater than max",
					item,
					item.MaxPrice.Add(1000),
					"cannot be greater than",
				},
			} {
				t.Run(tt.name, func(t *testing.T) {
					req := pkg.BidRequest{
						Bid: &models.Bid{
							ItemID: &tt.item.ID,
							Amount: tt.amount,
							UserID: &users[0].ID,
							Valid:  true,
						},
					}

					b, err := json.Marshal(req)
					if err != nil {
						t.Fatal(err)
					}

					res, err := http.Post(
						ts.URL+"/"+tt.item.ID.String()+"/bids",
						"application/json; charset=utf-8",
						bytes.NewReader(b),
					)
					if err != nil {
						t.Fatal(err)
					}

					testutils.AssertJSON(t, res)
					jsonB, err := ioutil.ReadAll(res.Body)
					if err != nil {
						t.Fatal(err)
					}
					defer res.Body.Close()
					if !strings.Contains(string(jsonB), tt.want) {
						t.Errorf("expected %q to contain %q", string(jsonB), tt.want)
					}
				})
			}
		})
	})
	t.Run("non-blind item", func(t *testing.T) {
		t.Run("success", func(t *testing.T) {
			item := getItem(false)
			highestBid := findHighest(item)
			amount := highestBid.Amount.Add(item.MinIncrement)
			req := pkg.BidRequest{
				Bid: &models.Bid{
					ItemID: &item.ID,
					Amount: amount,
					UserID: &users[0].ID,
					Valid:  true,
				},
			}

			b, err := json.Marshal(req)
			if err != nil {
				t.Fatal(err)
			}

			res, err := http.Post(
				ts.URL+"/"+item.ID.String()+"/bids",
				"application/json; charset=utf-8",
				bytes.NewReader(b),
			)
			if err != nil {
				t.Fatal(err)
			}

			var bid = new(models.Bid)
			testutils.AssertJSON(t, res)
			testutils.DecodeJSON(t, res, bid)

			if req.Bid.Amount.Neq(bid.Amount) {
				t.Errorf("expected Amount %v got %v", req.Bid.Amount, bid.Amount)
			}
			if *req.Bid.UserID != *bid.UserID {
				t.Errorf("expected UserID %v got %v", *req.Bid.UserID, *bid.UserID)
			}
			if *req.Bid.ItemID != *bid.ItemID {
				t.Errorf("expected ItemID %v got %v", req.Bid.ItemID, bid.ItemID)
			}
			if !bid.Valid {
				t.Error("expected Valid to be true but is false")
			}
		})
		t.Run("errors", func(t *testing.T) {
			item := getItem(false)
			highest := findHighest(item)
			for _, tt := range []struct {
				name   string
				item   *models.Item
				amount models.Currency
				want   string
			}{
				{
					"lower than highest+increment",
					item,
					highest.Amount.Add(item.MinIncrement).Sub(1000), // 10
					"cannot be less than",
				},
				{
					"greater than max",
					item,
					item.MaxPrice.Add(1000),
					"cannot be greater than",
				},
			} {
				t.Run(tt.name, func(t *testing.T) {
					req := pkg.BidRequest{
						Bid: &models.Bid{
							ItemID: &tt.item.ID,
							Amount: tt.amount,
							UserID: &users[0].ID,
							Valid:  true,
						},
					}

					b, err := json.Marshal(req)
					if err != nil {
						t.Fatal(err)
					}

					res, err := http.Post(
						ts.URL+"/"+tt.item.ID.String()+"/bids",
						"application/json; charset=utf-8",
						bytes.NewReader(b),
					)
					if err != nil {
						t.Fatal(err)
					}

					testutils.AssertJSON(t, res)
					jsonB, err := ioutil.ReadAll(res.Body)
					if err != nil {
						t.Fatal(err)
					}
					if !strings.Contains(string(jsonB), tt.want) {
						t.Errorf("expected %q to contain %q", string(jsonB), tt.want)
					}
				})
			}
		})
	})
}

func TestWatchItem(t *testing.T, s *pkg.Server) {
	chain := s.ItemCtx(http.HandlerFunc(s.WatchItem))
	r := chi.NewRouter()
	r.Post("/{itemID}/watch", chain.ServeHTTP)
	ts := httptest.NewServer(r)
	defer ts.Close()
	t.Run("can watch", func(t *testing.T) {

		req := pkg.SubRequest{
			UserID: &users[0].ID,
		}
		b, err := json.Marshal(req)
		if err != nil {
			t.Fatal(err)
		}

		res, err := http.Post(
			ts.URL+"/"+items[0].ID.String()+"/watch",
			"application/json; charset=utf-8",
			bytes.NewReader(b),
		)
		if err != nil {
			t.Fatal(err)
		}
		testutils.AssertJSON(t, res)
		var cb = new(pkg.CallBackResponse)
		testutils.DecodeJSON(t, res, cb)
		if cb.Status != "OK" {
			t.Errorf("expected Status OK got %v", cb.Status)
		}
	})
}

func TestUnWatchItem(t *testing.T, s *pkg.Server) {
	chain := s.ItemCtx(http.HandlerFunc(s.UnWatchItem))
	r := chi.NewRouter()
	r.Post("/{itemID}/unwatch", chain.ServeHTTP)
	ts := httptest.NewServer(r)
	defer ts.Close()
	t.Run("can unwatch", func(t *testing.T) {

		req := pkg.SubRequest{
			UserID: &users[0].ID,
		}
		b, err := json.Marshal(req)
		if err != nil {
			t.Fatal(err)
		}

		res, err := http.Post(
			ts.URL+"/"+items[0].ID.String()+"/unwatch",
			"application/json; charset=utf-8",
			bytes.NewReader(b),
		)
		if err != nil {
			t.Fatal(err)
		}
		testutils.AssertJSON(t, res)
		var cb = new(pkg.CallBackResponse)
		testutils.DecodeJSON(t, res, cb)
		if cb.Status != "OK" {
			t.Errorf("expected Status OK got %v", cb.Status)
		}
	})
}

func TestInvalidateBid(t *testing.T, s *pkg.Server) {
	chain := http.HandlerFunc(s.InvalidateBid)
	r := chi.NewRouter()
	r.Post("/{bidID}", chain.ServeHTTP)
	ts := httptest.NewServer(r)
	defer ts.Close()
	t.Run("can watch", func(t *testing.T) {

		req := pkg.SubRequest{
			UserID: &users[0].ID,
		}
		b, err := json.Marshal(req)
		if err != nil {
			t.Fatal(err)
		}

		res, err := http.Post(
			ts.URL+"/"+bids[0].ID.String(),
			"application/json; charset=utf-8",
			bytes.NewReader(b),
		)
		if err != nil {
			t.Fatal(err)
		}
		testutils.AssertJSON(t, res)
		var cb = new(pkg.CallBackResponse)
		testutils.DecodeJSON(t, res, cb)
		if cb.Status != "OK" {
			t.Errorf("expected Status OK got %v", cb.Status)
		}
	})
}

func TestItemImages(t *testing.T, s *pkg.Server) {
	imageList := func(item *models.Item) []*models.ItemImage {
		images := make([]*models.ItemImage, 0)
		for _, it := range items {
			if it.ID == item.ID {
				for _, im := range it.Images {
					images = append(images, im)
				}
				return images
			}
		}
		return images
	}

	t.Run("can get list of images", func(t *testing.T) {
		chain := s.ItemCtx(s.ItemImagesCtx(http.HandlerFunc(pkg.ItemImages)))
		r := chi.NewRouter()
		r.Get("/{itemID}/images", chain.ServeHTTP)
		ts := httptest.NewTLSServer(r)
		defer ts.Close()

		client := ts.Client()
		res, err := client.Get(ts.URL + "/" + items[0].ID.String() + "/images")
		if err != nil {
			t.Fatal(err)
		}
		testutils.AssertJSON(t, res)
		var list = make([]*models.ItemImage, 0)
		testutils.DecodeJSON(t, res, &list)
		want := imageList(items[0])
		testutils.AssertImageCollectionsEqual(t, list, want)
	})
	t.Run("paginated", func(t *testing.T) {
		chain := pagination.Context(
			s.ItemCtx(s.ItemImagesCtx(http.HandlerFunc(pkg.ItemImages))),
		)
		r := chi.NewRouter()
		r.Get("/{itemID}/images", chain.ServeHTTP)
		ts := httptest.NewTLSServer(r)
		defer ts.Close()

		u, err := url.Parse(ts.URL + "/" + items[0].ID.String() + "/images")
		if err != nil {
			t.Fatal(err)
		}
		q := url.Values{}
		images := imageList(items[0])
		models.ImagesOrderedBy(models.SortImagesByOrderAsc).Sort(images)
		q.Add("last-order", fmt.Sprintf("%d", images[1].Order.Int64))
		q.Add("limit", "10")
		u.RawQuery = q.Encode()

		client := ts.Client()
		res, err := client.Get(u.String())
		if err != nil {
			t.Fatal(err)
		}
		testutils.AssertJSON(t, res)
		var list = make([]*models.ItemImage, 0)
		testutils.DecodeJSON(t, res, &list)
		testutils.AssertImageCollectionsEqual(t, list, images[2:])
	})
}

// func createMultipartFormData(t *testing.T, fieldName, fileName string) (bytes.Buffer, *multipart.Writer) {
// 	var b bytes.Buffer
// 	var err error
// 	w := multipart.NewWriter(&b)
// 	var fw io.Writer
// 	file := mustOpen(fileName)
// 	if fw, err = w.CreateFormFile(fieldName, file.Name()); err != nil {
// 		t.Errorf("Error creating writer: %v", err)
// 	}
// 	if _, err = io.Copy(fw, file); err != nil {
// 		t.Errorf("Error with io.Copy: %v", err)
// 	}
// 	w.Close()
// 	return b, w
// }

// NewMultipartFormData creates a multipart/form-data payload. Caller must close the
// *multipart.Writer after acquiring it.
func NewMultipartFormData(values map[string]io.Reader) (io.ReadWriter, *multipart.Writer, error) {
	var err error
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	for key, r := range values {
		var fw io.Writer
		if x, ok := r.(io.Closer); ok {
			defer x.Close()
		}
		if x, ok := r.(*os.File); ok {
			if fw, err = w.CreateFormFile(key, x.Name()); err != nil {
				return nil, nil, err
			}
		} else {
			if fw, err = w.CreateFormField(key); err != nil {
				return nil, nil, err
			}
		}
		if _, err = io.Copy(fw, r); err != nil {
			return nil, nil, err
		}
	}
	return &b, w, nil
}

func TestAddItemImages(t *testing.T, s *pkg.Server) {
	chain := s.ItemCtx(http.HandlerFunc(s.AddItemImages))
	r := chi.NewRouter()
	r.Post("/{itemID}/images", chain.ServeHTTP)
	ts := httptest.NewTLSServer(r)
	defer ts.Close()
	client := ts.Client()

	gopath := os.Getenv("GOPATH")
	testImgPath := filepath.Join(
		gopath,
		"src",
		"github.com",
		"djangulo",
		"sfd",
		"db",
		"mock",
		"150x150.png",
	)

	t.Run("add 1", func(t *testing.T) {

		img1, err := os.OpenFile(testImgPath, os.O_RDONLY, 0755)
		if err != nil {
			t.Fatal(err)
		}
		defer img1.Close()
		values := map[string]io.Reader{
			"image-1":       img1,
			"count":         strings.NewReader("1"),
			"format":        strings.NewReader("image-%d"),
			"image-1-order": strings.NewReader("1"),
			"image-1-alt":   strings.NewReader("new test image added"),
		}
		item := items[0]
		want := []*models.ItemImage{
			{
				File: &models.File{
					AltText: sql.NullString{Valid: true, String: "new test image added"},
					Order:   sql.NullInt64{Valid: true, Int64: 1},
				},
				ItemID: &item.ID,
			},
		}
		b, w, err := NewMultipartFormData(values)
		if err != nil {
			t.Fatal(err)
		}
		w.Close()
		req, err := http.NewRequest(
			http.MethodPost,
			ts.URL+"/"+item.ID.String()+"/images",
			b,
		)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", w.FormDataContentType())

		res, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}

		testutils.AssertJSON(t, res)

		var list = make([]*models.ItemImage, 0)
		testutils.DecodeJSON(t, res, &list)
		testutils.AssertImageCollectionsEqual(t, list, want)
		// t.Fail()
	})
	t.Run("add many", func(t *testing.T) {

		img1, err := os.OpenFile(testImgPath, os.O_RDONLY, 0755)
		if err != nil {
			t.Fatal(err)
		}
		defer img1.Close()
		values := map[string]io.Reader{
			"count":         strings.NewReader("3"),
			"format":        strings.NewReader("image-%d"),
			"image-1":       img1,
			"image-1-order": strings.NewReader("1"),
			"image-1-alt":   strings.NewReader("new test image 1"),
			"image-2":       img1,
			"image-2-order": strings.NewReader("2"),
			"image-2-alt":   strings.NewReader("new test image 2"),
			"image-3":       img1,
			"image-3-order": strings.NewReader("3"),
			"image-3-alt":   strings.NewReader("new test image 3"),
		}
		item := items[0]
		want := []*models.ItemImage{
			{
				File: &models.File{
					AltText: sql.NullString{Valid: true, String: "new test image 1"},
					Order:   sql.NullInt64{Valid: true, Int64: 1},
				},
				ItemID: &item.ID,
			},
			{
				File: &models.File{
					AltText: sql.NullString{Valid: true, String: "new test image 2"},
					Order:   sql.NullInt64{Valid: true, Int64: 2},
				},
				ItemID: &item.ID,
			},
			{
				File: &models.File{
					AltText: sql.NullString{Valid: true, String: "new test image 3"},
					Order:   sql.NullInt64{Valid: true, Int64: 3},
				},
				ItemID: &item.ID,
			},
		}
		b, w, err := NewMultipartFormData(values)
		if err != nil {
			t.Fatal(err)
		}
		w.Close()
		req, err := http.NewRequest(
			http.MethodPost,
			ts.URL+"/"+item.ID.String()+"/images",
			b,
		)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", w.FormDataContentType())

		res, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}

		testutils.AssertJSON(t, res)

		var list = make([]*models.ItemImage, 0)
		testutils.DecodeJSON(t, res, &list)
		testutils.AssertImageCollectionsEqual(t, list, want)
	})
}
