//Package mock has some utilities to instantiate data for tests.
// Relationshipts are populated from the parent-down, to avoid having 10
// different functions: i.e. Users() returns a fully populated *models.User
// with its own *models.UserStats, []*models.PhoneNumber, *models.ProfilePicture,
// etc.
package mock

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofrs/uuid"

	"github.com/djangulo/sfd/crypto/password"
	"github.com/djangulo/sfd/db/models"
	"github.com/djangulo/sfd/storage"
	_ "github.com/djangulo/sfd/storage/fs"
)

var (
	testImages = []struct{ ext, oFilename, alt, data string }{
		{".png", "test-image-1.png", "Test image 1", "iVBORw0KGgoAAAANSUhEUgAAAJYAAAB8BAMAAAB58FQ1AAAAG1BMVEXMzMyWlpaqqqq3t7fFxcW+vr6xsbGjo6OcnJyLKnDGAAAACXBIWXMAAA7EAAAOxAGVKw4bAAABCUlEQVRoge3SsU7DMBSF4RPHdjLaTcUMUxkTwcAYS1SsRjAwJqKCjq2QmK2WwmvjECI2JGc+3xZL+ZXrXICIiIiIiIiIiFJk0NZ22B7/jgxw+Qkok9q6gHh/8flZ304nwkAcbzqUqa3GoqiBohMd9FJWQG4NMp8HrFJb6wbFY5y0jS/jIHaAfDBwcUB1Sp7RIdsv4FoVW03j44kcGjqU13Nam6fO/RSKBaZW71/ljFa8ojB+l66m1t1JfsxrKTPeV/n125KHNrfW1smtBnoX/+M5sL2vx9bwcLtetsmt8q2v4355yEqEsZU9X9WYM6Pax50f9l6EYb+GhrNxydJbRERERERERET/+QbEaSM2tsdQ3wAAAABJRU5ErkJggg=="},
		{".png", "test-image-2.png", "Test image 2", "iVBORw0KGgoAAAANSUhEUgAAAJYAAACWBAMAAADOL2zRAAAAG1BMVEXMzMyWlpaqqqq3t7fFxcW+vr6xsbGjo6OcnJyLKnDGAAAACXBIWXMAAA7EAAAOxAGVKw4bAAABAElEQVRoge3SMW+DMBiE4YsxJqMJtHOTITPeOsLQnaodGImEUMZEkZhRUqn92f0MaTubtfeMh/QGHANEREREREREREREtIJJ0xbH299kp8l8FaGtLdTQ19HjofxZlJ0m1+eBKZcikd9PWtXC5DoDotRO04B9YOvFIXmXLy2jEbiqE6Df7DTleA5socLqvEFVxtJyrpZFWz/pHM2CVte0lS8g2eDe6prOyqPglhzROL+Xye4tmT4WvRcQ2/m81p+/rdguOi8Hc5L/8Qk4vhZzy08DduGt9eVQyP2qoTM1zi0/uf4hvBWf5c77e69Gf798y08L7j0RERERERERERH9P99ZpSVRivB/rgAAAABJRU5ErkJggg=="},
		{".png", "test-image-3.png", "Test image 3", "iVBORw0KGgoAAAANSUhEUgAAASwAAAEsBAMAAACLU5NGAAAAG1BMVEXMzMyWlpacnJyqqqrFxcWxsbGjo6O3t7e+vr6He3KoAAAACXBIWXMAAA7EAAAOxAGVKw4bAAAD90lEQVR4nO3cwW+bSBQH4AcGw5HnJDhHaN3dHO1su9ojNGnPtrUb7dFuIiVHnEo5263Uv3vfGwab1myVA5DV6vcpgeD35HmeGYbJxUQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAD/zOb3N5BRexlu9/Jo+NCQFl/HbWrRK7s6Amcdy3jCfaftyOT/OmsnLxSFqkzu04Ns1Z+RxPOMtUc63fH6U5HP8O5/uo1Vyh9IJhTylwSjz0pV0y4Tex0dJ7iij3ck+WiV3J9RPvVhRLgO5O5V+KOSl7MesnXSRH++jNrlDAWurEW0i6ZOz8jI9mlwaDXkftckd8nXEdgnNVjI2sf6Q+VvLSMiMHJnupHC0j9rkrmlL87Lhs7JK86oM1fowVFq0jdrkjn2QKbMuTEvD8aGsfCQ9th9PbzHeR21yt1KWkUq3et+Tq4tDHpnXfZ67+7Zdltu1itrkbrEuRWVLWdmwHbl0shlXSQ7LLVtFbXLXZUmLphHOHK3IsWVtTg6Lk6PFV1Gb3G1Z9I1Xjb015NpSHq7jfntL7reoaW7JhD+pJQ2537llVuyGO1Em17iWJMt7f3ei/zeZcdGlKLDr1saW5XPV9F9bM2pV1CZ3yDxDZFx0HZcF0z+s8rpwVcuWPo5k1KqoTe7QwD58mp6Js/PUTn4tVEatx2ei3lAzu4M4t3uErQl5PN3YOb84NR+gitrkDnl8J51QNO23hjLH7SqQxxnp0trbfotmo9t0RE27U9k9hFw2PuBfLnVD0d/u9KMs8hNq2svrxFqXJXprZtmg9riXp5v0jTRI4afyn5lv1X8+gRaQ22XA/zT6sxatkgEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAD65tjf/5gXLYtHb/8l8kNZkVw5zEwjIjei8ru7rtJ7/YqcO3ISorTFsiLvt+eXZY7xlp5sWd6b7KscrpeZ80DBus2y6D1dviY3C+QP/9WUnGWkp8GrhZa1fE3DQiK1ssYrurdlDeblwZ86TzTctFuWf/dxPihy+kw31+/IuTOnm2v98I6EwoTe1cuKLsLEluVm5cFLHHf7pc2JKIPoZl4STpfFHzSRfnEyc5pQrmVJiO7l13yRHpdlPQ0LW5ZTHSInWN23WZZMedMJycUq0aa1FT1F1dyK6MugoHpvuY903Fv0a9Jqb+n7apesHlY0KSvRU6233CV9V5Z/RsdzixbzlsvSuUXL4nFOT9mVtq2nw9yiYPx9WebCHGt3IrW7yOnby51IuyzPKEgv9M31dLgTKUgayioH+oqrdavlsp5hWPTb3jM9vnQBjZyLl64AAP43/gHVSaMe2vmdiAAAAABJRU5ErkJggg=="},
		{".png", "test-image-4.png", "Test image 4", "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"},
	}
)

const (
	testUsers = "sfd_test_users.json"
	testItems = "sfd_test_items.json"
	testBids  = "sfd_test_bids.json"
)

var storageDrv storage.Driver

func init() {
	rand.Seed(time.Now().UnixNano())
	var err error
	u := "fs://./assets?accept=.jpg&accept=.png&accept=.jpeg&accept=.svg&root=/tmp/assets"
	storageDrv, err = storage.Open(u)
	if err != nil {
		log.Fatalf("mock: failed to initialize storage driver: %v", err)
	}
}

func Users() []*models.User {
	users := make([]*models.User, 0)

	fpath := filepath.Join(os.TempDir(), testUsers)
	if _, err := os.Stat(fpath); os.IsNotExist(err) {
		for i := 0; i < 4; i++ {
			now := time.Now()
			updated := now.Add(time.Hour * time.Duration(time.Hour*time.Duration(rand.Int31n(48))))
			username := fmt.Sprintf("testuser-%.2d", i)

			var profilePublic bool
			if i%2 == 0 {
				profilePublic = true
			}
			user, err := models.NewUser(
				username,
				fmt.Sprintf("%s@email.com", username),

				password.MustHash(username),
				"en-US",
				profilePublic,
			)
			if err != nil {
				panic(err)
			}
			user.LastLogin = models.NewNullTime(updated)
			user.Stats.LoginCount = 1

			decoder := base64.NewDecoder(
				base64.StdEncoding,
				strings.NewReader(testImages[0].data),
			)
			ppicID := uuid.Must(uuid.NewV4())
			path := storageDrv.NormalizePath(
				"users",
				user.ID.String(),
				fmt.Sprintf("%s%s", ppicID.String(), testImages[0].ext),
			)
			absPath, err := storageDrv.AddFile(decoder, path)
			if err != nil {
				log.Fatalf("mock.Users: %v", err)
			}
			ppic, err := models.NewProfilePicture(
				&user.ID,
				path,
				absPath,
				testImages[0].oFilename,
				".png",
				fmt.Sprintf("Profile picture of user %s", username))
			if err != nil {
				panic(err)
			}
			user.Picture = ppic

			user.PhoneNumbers = make([]*models.PhoneNumber, 0)
			for k := 0; k <= i; k++ {
				p := models.PhoneNumber{
					ID:     uuid.Must(uuid.NewV4()),
					UserID: &user.ID,
					Number: fmt.Sprintf("555-555-%.2d%.2d", k, i),
				}
				user.PhoneNumbers = append(user.PhoneNumbers, &p)
			}

			user.Addresses = make([]*models.Address, 0)
			for k := 0; k <= i; k++ {
				var kind models.AddressKind
				if k%2 == 0 {
					kind = models.Billing
				} else {
					kind = models.Shipping
				}
				a := models.Address{
					DBObj: &models.DBObj{
						ID: uuid.Must(uuid.NewV4()),
					},
					UserID:  &user.ID,
					Address: fmt.Sprintf("Addresss %d of user %s", k+1, user.FullName),
					Kind:    kind,
				}
				user.Addresses = append(user.Addresses, &a)
			}

			users = append(users, user)
		}
		models.UsersOrderedBy(
			models.SortUsersByCreatedAtDesc,
			models.SortUsersByIDDesc,
		).Sort(users)
		b, err := json.MarshalIndent(users, "", "  ")
		err = ioutil.WriteFile(fpath, b, 0644)
		if err != nil {
			log.Fatalf("mock.Users: error writing %s: %v", fpath, err)
		}
	} else {
		b, err := getBytes(fpath)
		if err != nil {
			log.Fatalf("mock.Users: error reading %s: %v", fpath, err)
		}
		err = json.Unmarshal(b, &users)
		if err != nil {
			log.Fatalf("mock.Users: error unmarshaling %s into []*models.User: %v", fpath, err)
		}
	}

	return users
}

func Items(users []*models.User) []*models.Item {
	items := make([]*models.Item, 0)

	fpath := filepath.Join(os.TempDir(), testItems)
	if _, err := os.Stat(fpath); os.IsNotExist(err) {
		for j := 0; j < len(users); j++ {
			var user = users[j]
			for i := 0; i < 30; i++ {
				min, _ := models.NewCurrency(rand.Int63n(1000))
				max, _ := models.NewCurrency(150000)

				max = min.Add(max)
				now := time.Now()
				var publishedAt time.Time
				if i%2 == 0 {
					publishedAt = now.Add(time.Hour * time.Duration(rand.Int31n(100)))
				}
				increment, _ := models.NewCurrency((i % 5) * 100)

				var blind bool
				if i%3 == 0 {
					blind = true
				}

				created := now.Add(time.Hour * time.Duration((j*30)+i*2))
				updated := created.Add(time.Hour * time.Duration(i*2))

				someTimeAgo := now.Add(time.Duration(-48+(j*30)+(i)) * time.Hour)

				name := fmt.Sprintf("Test Item %.3d", (j*30)+i+1)

				item, err := models.NewItem(
					&user.ID,
					name,
					fmt.Sprintf("Description of %s", name),
					min,
					max,
					increment,
					blind,
					rand.Intn(7200),
					someTimeAgo,
					publishedAt,
				)
				if err != nil {
					panic(err)
				}
				item.AdminApproved = true
				item.Translations = []*models.Translation{
					{
						Lang:        models.English,
						Name:        name,
						Slug:        models.Slugify(name),
						Description: fmt.Sprintf("Description of %s", name),
						ItemID:      &item.ID,
					},
					{
						Lang:        models.Spanish,
						Name:        fmt.Sprintf("Artículo de prueba %.3d", (j*30)+i+1),
						Slug:        models.Slugify(fmt.Sprintf("Artículo de prueba %.3d", (j*30)+i+1)),
						Description: fmt.Sprintf("Descripción del Artículo de prueba of %.3d", (j*30)+i+1),
						ItemID:      &item.ID,
					},
				}

				for k, img := range testImages {
					decoder := base64.NewDecoder(
						base64.StdEncoding,
						strings.NewReader(testImages[0].data),
					)
					imgID := uuid.Must(uuid.NewV4())
					path := storageDrv.NormalizePath("items", item.ID.String(), fmt.Sprintf("%s%s", imgID.String(), img.ext))
					absPath, err := storageDrv.AddFile(decoder, path)
					if err != nil {
						log.Fatalf("mock.Items: %v", err)
					}
					image := models.ItemImage{
						File: &models.File{
							ID:               uuid.NullUUID{Valid: true, UUID: imgID},
							CreatedAt:        updated,
							Path:             models.NewNullString(path),
							AbsPath:          models.NewNullString(absPath),
							AltText:          models.NewNullString(fmt.Sprintf("Image %d of 4", k+1)),
							OriginalFilename: models.NewNullString(img.oFilename),
							FileExt:          models.NewNullString(".png"),
							Order:            models.NewNullInt64(k + 1),
						},
						ItemID: &item.ID,
					}
					if image.Order.Int64 == 1 {
						item.CoverImage = &image
					}
					item.Images = append(item.Images, &image)
				}

				items = append(items, item)
			}
		}

		models.ItemsOrderedBy(
			models.SortItemsByCreatedAtDesc,
			models.SortItemsByIDDesc,
		).Sort(items)
		b, err := json.MarshalIndent(items, "", "  ")
		err = ioutil.WriteFile(fpath, b, 0644)
		if err != nil {
			log.Fatalf("mock.Items: error writing %s: %v", fpath, err)
		}

	} else {
		b, err := getBytes(fpath)
		if err != nil {
			log.Fatalf("mock.Items: error reading %s: %v", fpath, err)
		}
		err = json.Unmarshal(b, &items)
		if err != nil {
			log.Fatalf("mock.Items: error unmarshaling %s into []*models.Item: %v", fpath, err)
		}
	}

	return items
}

func Bids(items []*models.Item, users []*models.User) []*models.Bid {
	bids := make([]*models.Bid, 0)

	fpath := filepath.Join(os.TempDir(), testBids)
	if _, err := os.Stat(fpath); os.IsNotExist(err) {
		for _, item := range items {
			for _, user := range users {
				for i := 0; i < 3; i++ {
					created := time.Now().Add(time.Hour * time.Duration(i*2))
					updated := created.Add(time.Hour * time.Duration(i*2))
					var amount models.Currency
					r := rand.Intn(int(item.MaxPrice.Sub(models.Currency(50000)).AsInt()))
					amount, _ = models.NewCurrency(r)

					bid, err := models.NewBid(&item.ID, &user.ID, amount)
					if err != nil {
						panic(err)
					}
					bid.UpdatedAt = models.NewNullTime(updated)

					for bid.Amount.Lt(item.StartingPrice) {
						bid.Amount = bid.Amount.Add(models.Currency(100))
					}

					bids = append(bids, bid)
				}
			}
		}

		models.BidsOrderedBy(
			models.SortBidsByAmountDesc,
		).Sort(bids)
		b, err := json.MarshalIndent(bids, "", "  ")
		err = ioutil.WriteFile(fpath, b, 0644)
		if err != nil {
			log.Fatalf("mock.Bids: error writing %s: %v", fpath, err)
		}

	} else {
		b, err := getBytes(fpath)
		if err != nil {
			log.Fatalf("mock.Bids: error reading %s: %v", fpath, err)
		}
		err = json.Unmarshal(b, &bids)
		if err != nil {
			log.Fatalf("mock.Bids: error unmarshaling %s into []*models.Bid: %v", fpath, err)
		}
	}
	return bids
}

func Stats(users []*models.User, items []*models.Item, bids []*models.Bid) {
	for _, user := range users {
		var itemCount, bidCount, bidsWon int
		for _, item := range items {
			if *item.OwnerID == user.ID {
				itemCount++
			}
			winner := models.Winner(item.Bids)
			if winner.UserID != nil && *winner.UserID == user.ID {
				bidsWon++
			}
		}
		for _, bid := range bids {
			if *bid.UserID == user.ID {
				bidCount++
			}
		}
		user.Stats.ItemsCreated = itemCount
		user.Stats.BidsCreated = bidCount
		user.Stats.BidsWon = bidsWon
	}
}

func getBytes(path string) ([]byte, error) {
	fh, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer fh.Close()
	b, err := ioutil.ReadAll(fh)
	if err != nil {
		return nil, err
	}

	return b, nil
}
