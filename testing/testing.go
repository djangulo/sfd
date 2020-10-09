package testing

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"text/tabwriter"

	"github.com/gofrs/uuid"

	"github.com/djangulo/sfd/db/models"
)

func DecodeJSON(t *testing.T, response *http.Response, dest interface{}) {
	t.Helper()
	err := json.NewDecoder(response.Body).Decode(dest)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
}

func AssertJSON(t *testing.T, response *http.Response) {
	t.Helper()
	got := response.Header.Get("Content-Type")
	if !strings.Contains(got, "application/json") {
		t.Errorf("expected %q got %q", "application/json", got)
	}
}

func AssertUserCollectionsEqual(t *testing.T, got, want []*models.User) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("lengths differ, got %d want %d", len(got), len(want))
	}

	for i := 0; i < len(got); i++ {
		AssertUsersEqual(t, got[i], want[i])
	}
}

func AssertUsersEqual(t *testing.T, got, want *models.User) {
	t.Helper()
	var b strings.Builder

	var gotID uuid.UUID
	if got.DBObj == nil {
		gotID = uuid.Nil
	} else {
		gotID = got.ID
	}

	if got.Username != want.Username {
		const format = "%v\t%v\t%v\n"
		tw := new(tabwriter.Writer).Init(&b, 0, 8, 2, ' ', 0)
		fmt.Fprintf(tw, format, "Field", "Got", "Want")
		fmt.Fprintf(tw, format, "-----", "---", "----")
		fmt.Fprintf(tw, format, "ID", gotID, want.ID)
		fmt.Fprintf(tw, format, "Username", got.Username, want.Username)
		fmt.Fprintf(tw, format, "IsAdmin", got.IsAdmin, want.IsAdmin)
		fmt.Fprintf(tw, format, "IsActive", got.Active, want.Active)
		fmt.Fprintf(tw, format, "ProfilePublic", got.ProfilePublic, want.ProfilePublic)
		tw.Flush()
		t.Errorf("wrong results\n%s", b.String())
	}
}

func AssertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Errorf("got an error but didn't expect one: %v", err)
	}
}

// AssertItemsEqual pretty print a field-by-field comparison for items.
// All direct fields are compared (Name, OwnerID, StartingPrice, MaxPrice,
// MinIncrement, BidInterval, BidDeadline, PublishedAt)
func AssertItemsEqual(t *testing.T, got, want *models.Item) {
	t.Helper()
	var b strings.Builder
	var gotID uuid.UUID
	if got.DBObj == nil {
		gotID = uuid.Nil
	} else {
		gotID = got.ID
	}

	if got.Name != want.Name ||
		*got.OwnerID != *want.OwnerID ||
		got.StartingPrice.Neq(want.StartingPrice) ||
		got.MaxPrice.Neq(want.MaxPrice) ||
		got.MinIncrement.Neq(want.MinIncrement) ||
		got.BidInterval != want.BidInterval ||
		got.BidDeadline.Unix() != want.BidDeadline.Unix() {
		const format = "%v\t%v\t%v\n"
		tw := new(tabwriter.Writer).Init(&b, 0, 8, 2, ' ', 0)
		fmt.Fprintf(tw, format, "Field", "Got", "Want")
		fmt.Fprintf(tw, format, "-----", "---", "----")
		fmt.Fprintf(tw, format, "ID", gotID, want.ID)
		fmt.Fprintf(tw, format, "OwnerID", got.OwnerID, want.OwnerID)
		fmt.Fprintf(tw, format, "Name", got.Name, want.Name)
		fmt.Fprintf(tw, format, "Slug", got.Slug, want.Slug)
		fmt.Fprintf(tw, format, "StartingPrice", got.StartingPrice, want.StartingPrice)
		fmt.Fprintf(tw, format, "MaxPrice", got.MaxPrice, want.MaxPrice)
		fmt.Fprintf(tw, format, "MinIncrement", got.MinIncrement, want.MinIncrement)
		fmt.Fprintf(tw, format, "BidInterval", got.BidInterval, want.BidInterval)
		fmt.Fprintf(tw, format, "BidDeadline", got.BidDeadline, want.BidDeadline)
		fmt.Fprintf(tw, format, "PublishedAt", got.PublishedAt, want.PublishedAt)
		fmt.Fprintf(tw, format, "Blind", got.Blind, want.Blind)
		fmt.Fprintf(tw, format, "Bids (len)", len(got.Bids), len(want.Bids))
		tw.Flush()
		t.Errorf("wrong results\n%s", b.String())
	}
}

func AssertItemCollectionsEqual(t *testing.T, got, want []*models.Item) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("lengths differ, got %d want %d", len(got), len(want))
	}

	for i := 0; i < len(got); i++ {
		AssertItemsEqual(t, got[i], want[i])
	}
}

// assertBidsEqual pretty print a field-by-field comparison for bids.
func AssertBidsEqual(t *testing.T, got, want *models.Bid) {
	t.Helper()
	var b strings.Builder

	if *got.UserID != *want.UserID || *got.ItemID != *want.ItemID {
		const format = "%v\t%v\t%v\n"
		tw := new(tabwriter.Writer).Init(&b, 0, 8, 2, ' ', 0)
		fmt.Fprintf(tw, format, "Field", "Got", "Want")
		fmt.Fprintf(tw, format, "-----", "---", "----")
		fmt.Fprintf(tw, format, "ID", got.ID, want.ID)
		fmt.Fprintf(tw, format, "UserID", got.UserID, want.UserID)
		fmt.Fprintf(tw, format, "ItemID", got.ItemID, want.ItemID)
		fmt.Fprintf(tw, format, "Amount", got.Amount, want.Amount)
		fmt.Fprintf(tw, format, "Valid", got.Valid, want.Valid)
		tw.Flush()
		t.Errorf("wrong results\n%s", b.String())
	}
}

func AssertBidCollectionsEqual(t *testing.T, got, want []*models.Bid) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("lengths differ, got %d want %d", len(got), len(want))
	}

	for i := 0; i < len(got); i++ {
		AssertBidsEqual(t, got[i], want[i])
	}
}

// AssertImagesEqual pretty print a field-by-field comparison for bids.
func AssertImagesEqual(t *testing.T, got, want *models.ItemImage) {
	t.Helper()
	var b strings.Builder

	if *got.ItemID != *want.ItemID || got.Order.Int64 != want.Order.Int64 {
		const format = "%v\t%v\t%v\n"
		tw := new(tabwriter.Writer).Init(&b, 0, 8, 2, ' ', 0)
		fmt.Fprintf(tw, format, "Field", "Got", "Want")
		fmt.Fprintf(tw, format, "-----", "---", "----")
		fmt.Fprintf(tw, format, "ID", got.ID, want.ID)
		fmt.Fprintf(tw, format, "ItemID", got.ItemID, want.ItemID)
		fmt.Fprintf(tw, format, "Path", got.Path, want.Path)
		fmt.Fprintf(tw, format, "AbsPath", got.AbsPath, want.AbsPath)
		fmt.Fprintf(tw, format, "AltText", got.AltText, want.AltText)
		fmt.Fprintf(tw, format, "FileExt", got.FileExt, want.FileExt)
		fmt.Fprintf(tw, format, "Order", got.Order, want.Order)
		tw.Flush()
		t.Errorf("wrong results\n%s", b.String())
	}
}

func AssertImageCollectionsEqual(t *testing.T, got, want []*models.ItemImage) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("lengths differ, got %d want %d", len(got), len(want))
	}

	for i := 0; i < len(got); i++ {
		AssertImagesEqual(t, got[i], want[i])
	}
}
