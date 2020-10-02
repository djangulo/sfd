package db

import (
	"fmt"
	"path/filepath"

	"github.com/gofrs/uuid"

	"github.com/djangulo/sfd/config"
)

var cfg = config.Get()

// ItemImagePath provide a standard path to save an ItemImage.
func ItemImagePath(itemID, imageID *uuid.UUID, ext, storageRoot string) string {
	return filepath.Join(
		cfg.PublicURL(),
		storageRoot,
		"items",
		itemID.String(),
		"images",
		fmt.Sprintf("%s%s", imageID.String(), ext),
	)
}

// ProfilePicturePath provide a standard path to save user's ProfilePictures.
func ProfilePicturePath(userID, imageID *uuid.UUID, filename, storageRoot string) string {
	return filepath.Join(
		cfg.PublicURL(),
		storageRoot,
		"users",
		userID.String(),
		"profile_pictures",
		fmt.Sprintf("%s%s", imageID.String(), filepath.Ext(filename)),
	)
}
