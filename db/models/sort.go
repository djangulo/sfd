package models

import "sort"

type ItemSorter struct {
	items []*Item
	less  []itemLess
}

type itemLess func(i1, i2 *Item) bool

func (is *ItemSorter) Sort(items []*Item) {
	is.items = items
	sort.Sort(is)
}
func (is *ItemSorter) Len() int {
	return len(is.items)
}
func (is *ItemSorter) Swap(i, j int) {
	is.items[i], is.items[j] = is.items[j], is.items[i]
}

func (is *ItemSorter) Less(i, j int) bool {
	p, q := is.items[i], is.items[j]

	var k int
	for k = 0; k < len(is.less)-1; k++ {
		less := is.less[k]
		switch {
		case less(p, q):
			return true
		case less(q, p):
			return false
		}
	}
	return is.less[k](p, q)
}

func ItemsOrderedBy(less ...itemLess) *ItemSorter {
	return &ItemSorter{
		less: less,
	}
}

func SortItemsByCreatedAtAsc(i1, i2 *Item) bool {
	return i1.CreatedAt.Before(i2.CreatedAt)
}
func SortItemsByCreatedAtDesc(i1, i2 *Item) bool {
	return i1.CreatedAt.After(i2.CreatedAt)
}

func SortItemsByIDAsc(i1, i2 *Item) bool {
	return i1.ID.String() < i2.ID.String()
}
func SortItemsByIDDesc(i1, i2 *Item) bool {
	return i1.ID.String() > i2.ID.String()
}

type BidSorter struct {
	bids []*Bid
	less []bidLess
}

type bidLess func(b1, b2 *Bid) bool

func (bs *BidSorter) Sort(bids []*Bid) {
	bs.bids = bids
	sort.Sort(bs)
}

func BidsOrderedBy(less ...bidLess) *BidSorter {
	return &BidSorter{
		less: less,
	}
}

func (bs *BidSorter) Len() int {
	return len(bs.bids)
}

func (bs *BidSorter) Swap(i, j int) {
	bs.bids[i], bs.bids[j] = bs.bids[j], bs.bids[i]
}

func (bs *BidSorter) Less(i, j int) bool {
	p, q := bs.bids[i], bs.bids[j]

	var k int
	for k = 0; k < len(bs.less)-1; k++ {
		less := bs.less[k]
		switch {
		case less(p, q):
			return true
		case less(q, p):
			return false
		}
	}
	return bs.less[k](p, q)
}

func SortBidsByCreatedAtAsc(b1, b2 *Bid) bool {
	return b1.CreatedAt.Before(b2.CreatedAt)
}
func SortBidsByCreatedAtDesc(b1, b2 *Bid) bool {
	return b1.CreatedAt.After(b2.CreatedAt)
}

func SortBidsByIDAsc(b1, b2 *Bid) bool {
	return b1.ID.String() < b2.ID.String()
}
func SortBidsByIDDesc(b1, b2 *Bid) bool {
	return b1.ID.String() > b2.ID.String()
}

func SortBidsByAmountDesc(b1, b2 *Bid) bool {
	return b1.Amount.Gt(b2.Amount)
}
func SortBidsByAmountAsc(b1, b2 *Bid) bool {
	return b1.Amount.Lt(b2.Amount)
}

type UserSorter struct {
	users []*User
	less  []userLess
}

type userLess func(u1, u2 *User) bool

func (us *UserSorter) Sort(users []*User) {
	us.users = users
	sort.Sort(us)
}

func UsersOrderedBy(less ...userLess) *UserSorter {
	return &UserSorter{
		less: less,
	}
}

func (us *UserSorter) Len() int {
	return len(us.users)
}

func (us *UserSorter) Swap(i, j int) {
	us.users[i], us.users[j] = us.users[j], us.users[i]
}

func (us *UserSorter) Less(i, j int) bool {
	p, q := us.users[i], us.users[j]

	var k int
	for k = 0; k < len(us.less)-1; k++ {
		less := us.less[k]
		switch {
		case less(p, q):
			return true
		case less(q, p):
			return false
		}
	}
	return us.less[k](p, q)
}

func SortUsersByCreatedAtAsc(u1, u2 *User) bool {
	return u1.CreatedAt.Before(u2.CreatedAt)
}
func SortUsersByCreatedAtDesc(u1, u2 *User) bool {
	return u1.CreatedAt.After(u2.CreatedAt)
}

func SortUsersByIDAsc(u1, u2 *User) bool {
	return u1.ID.String() < u2.ID.String()
}
func SortUsersByIDDesc(u1, u2 *User) bool {
	return u1.ID.String() > u2.ID.String()
}

type ImageSorter struct {
	images []*ItemImage
	less   []imageLess
}

type imageLess func(b1, b2 *ItemImage) bool

func (is *ImageSorter) Sort(images []*ItemImage) {
	is.images = images
	sort.Sort(is)
}

func ImagesOrderedBy(less ...imageLess) *ImageSorter {
	return &ImageSorter{
		less: less,
	}
}

func (is *ImageSorter) Len() int {
	return len(is.images)
}

func (is *ImageSorter) Swap(i, j int) {
	is.images[i], is.images[j] = is.images[j], is.images[i]
}

func (is *ImageSorter) Less(i, j int) bool {
	p, q := is.images[i], is.images[j]

	var k int
	for k = 0; k < len(is.less)-1; k++ {
		less := is.less[k]
		switch {
		case less(p, q):
			return true
		case less(q, p):
			return false
		}
	}
	return is.less[k](p, q)
}

func SortImagesByOrderDesc(i1, i2 *ItemImage) bool {
	return i1.Order > i2.Order
}
func SortImagesByOrderAsc(i1, i2 *ItemImage) bool {
	return i1.Order < i2.Order
}
