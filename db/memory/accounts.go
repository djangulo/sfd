package memory

import (
	"fmt"

	"github.com/djangulo/sfd/db/models"
	"github.com/gofrs/uuid"
)

func (m *Memory) LoginUser(user *models.User) (*models.User, error) {
	for _, u := range m.users {
		if u.ID == user.ID {
			for _, p := range m.profilePictures {
				if *p.UserID == u.ID {
					user.Picture = p
					break
				}
			}
			u = user
			return user, nil
		}
	}
	return nil, models.ErrNotFound
}

func (m *Memory) CreateUser(user *models.User) error {
	if user == nil {
		return models.ErrNilPointer
	}
	if user.ID != uuid.Nil {
		for _, u := range m.users {
			if u.ID == user.ID || u.Username == user.Username {
				return models.ErrAlreadyExists
			}
		}
	}
	m.users = append(m.users, user)
	return nil
}

func (m *Memory) UserByUsernameOrEmail(username string) (*models.User, error) {
	if username == "" {
		return nil, fmt.Errorf("%w: username is empty", models.ErrInvalidInput)
	}

	for _, u := range m.users {
		if u.Username == username || u.Email == username {
			for _, p := range m.profilePictures {
				if *p.UserID == u.ID {
					u.Picture = p
					break
				}
			}
			return u, nil
		}
	}
	return nil, models.ErrNotFound
}

// GetPasswordHash returns the password hash for the given userID. This method
// exists because none of the other user-returning methods include the hash.
func (m *Memory) GetPasswordHash(userID *uuid.UUID) (string, error) {
	for _, u := range m.users {
		if u.ID == *userID {
			return u.PasswordHash, nil
		}
	}
	return "", models.ErrNotFound
}

func (m *Memory) UserByID(id *uuid.UUID) (*models.User, error) {
	if id == nil {
		return nil, fmt.Errorf("%w: id is nil ", models.ErrNilPointer)
	}
	if *id == uuid.Nil {
		return nil, fmt.Errorf("%w: id is uuid.Nil ", models.ErrInvalidInput)
	}

	for _, u := range m.users {
		if u.ID == *id {
			for _, p := range m.profilePictures {
				if *p.UserID == u.ID {
					u.Picture = p
					break
				}
			}
			return u, nil
		}
	}
	return nil, models.ErrNotFound
}

func (m *Memory) ResetPassword(userID *uuid.UUID, newPasswordHash string) error {
	for _, user := range m.users {
		if user.ID == *userID {
			user.PasswordHash = newPasswordHash
			return nil
		}
	}
	return models.ErrNotFound
}

func (m *Memory) ChangePassword(user *models.User, newPasswordHash string) (*models.User, error) {
	for _, u := range m.users {
		if u.ID == user.ID {
			u.PasswordHash = newPasswordHash
			return user, nil
		}
	}
	return nil, models.ErrNotFound
}

func (m *Memory) AddProfilePic(userID *uuid.UUID, image *models.ProfilePicture) error {
	m.profilePictures = append(m.profilePictures, image)
	return nil

}

func (m *Memory) RemoveProfilePic(id *uuid.UUID) error {
	for i, img := range m.profilePictures {
		if img.ID.Valid {
			if img.ID.UUID != uuid.Nil {
				if img.ID.UUID == *id {
					m.profilePictures[i] = m.profilePictures[len(m.profilePictures)-1]
					m.profilePictures[len(m.profilePictures)-1] = nil
					m.profilePictures = m.profilePictures[:len(m.profilePictures)-1]
					return nil
				}
			}
		}
	}
	return fmt.Errorf("%w: %v", models.ErrNotFound, id)
}

// AddPhoneNumbers builds a bulk insert for user phone numbers.
func (m *Memory) AddPhoneNumbers(userID *uuid.UUID, phoneNumbers ...*models.PhoneNumber) error {
	for _, phone := range phoneNumbers {
		m.userPhones = append(m.userPhones, phone)
	}
	return nil
}

func (m *Memory) RemovePhoneNumber(numberID *uuid.UUID) error {
	for i, p := range m.userPhones {
		if p.ID == *numberID {
			m.userPhones[i] = m.userPhones[len(m.userPhones)-1]
			m.userPhones[len(m.userPhones)-1] = nil
			m.userPhones = m.userPhones[:len(m.userPhones)-1]
			return nil
		}
	}
	return models.ErrNotFound
}

func (m *Memory) UnsubEmail(email string, kind models.NoMailKind) error {
	m.nomail = append(m.nomail, &models.NoMail{Email: email, Kind: kind})
	return nil
}
func (m *Memory) IsUnsub(email string, kind models.NoMailKind) bool {
	for _, nm := range m.nomail {
		if nm.Email == email && nm.Kind == kind {
			return true
		}
	}
	return false
}

func (m *Memory) ResubEmail(email string, kind models.NoMailKind) error {
	for i, nm := range m.nomail {
		if nm.Email == email && nm.Kind == kind {
			m.nomail[i] = m.nomail[len(m.nomail)-1]
			m.nomail[len(m.nomail)-1] = nil
			m.nomail = m.nomail[:len(m.nomail)-1]
			return nil
		}
	}
	return models.ErrNotFound
}
