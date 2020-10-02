//Package helpers contains extended models.
package helpers

import (
	"github.com/djangulo/sfd/db/models"
	"github.com/djangulo/sfd/pagination"
)

type UserListData struct {
	Pages *pagination.Pages
	Users []*models.User
}

type UserDetailData struct {
	Pages *pagination.Pages
	User  *models.User
	Bids  []*models.Bid
}
