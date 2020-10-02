package memory

import (
	"fmt"

	"github.com/djangulo/sfd/db/models"
	"github.com/gofrs/uuid"
)

func (m *Memory) ActivateUser(userID, approverID *uuid.UUID) error {
	for _, u := range m.users {
		if u.ID == *userID {
			u.Active = true
			return nil
		}
	}
	return fmt.Errorf("%w %v", models.ErrNotFound, *userID)
}

func (m *Memory) DeactivateUser(userID *uuid.UUID) error {
	for _, u := range m.users {
		if u.ID == *userID {
			u.Active = false
			return nil
		}
	}
	return fmt.Errorf("%w %v", models.ErrNotFound, *userID)
}

func (m *Memory) GrantAdmin(userID *uuid.UUID) error {
	for _, u := range m.users {
		if u.ID == *userID {
			u.IsAdmin = true
			return nil
		}
	}
	return fmt.Errorf("%w %v", models.ErrNotFound, *userID)
}

func (m *Memory) RevokeAdmin(userID *uuid.UUID) error {

	if userID == nil {
		return models.ErrNilPointer
	}
	for _, u := range m.users {
		if u.ID == *userID {
			u.IsAdmin = false
			return nil
		}
	}
	return fmt.Errorf("%w %v", models.ErrNotFound, userID)
}

func (m *Memory) ListUsers(opts *models.ListOptions) (int, []*models.User, error) {
	items := m.users

	if opts == nil {
		opts = models.NewOptions()
	}

	limit := opts.Limit
	if limit <= 0 {
		limit = 1000
	}

	offset := opts.Offset
	length := len(items)
	if offset >= length {
		return 0, nil, models.ErrNoResults
	}
	if length < limit {
		limit = length
	}
	if offset+limit > length {
		return length, items[(0 + offset):length], nil
	}
	return length, items[(0 + offset):(offset + limit)], nil
}
