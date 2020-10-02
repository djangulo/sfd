package memory

import (
	"fmt"
	"time"

	"github.com/djangulo/sfd/crypto/session"
	"github.com/djangulo/sfd/db/models"
)

func (m *Memory) NewSession(ses session.Session) error {
	m.sessions = append(m.sessions, ses)
	return nil
}

func (m *Memory) ReadSession(id string) ([]byte, error) {
	for _, ses := range m.sessions {
		if ses.ID() == id {
			return ses.Bytes()
		}
	}
	return nil, models.ErrNotFound
}

func (m *Memory) DeleteSession(id string) error {
	for i, ses := range m.sessions {
		if ses.ID() == id {
			m.sessions[i] = m.sessions[len(m.sessions)-1]
			m.sessions[len(m.sessions)-1] = nil
			m.sessions = m.sessions[:len(m.sessions)-1]
			return nil
		}
	}
	return fmt.Errorf("%w: %v", models.ErrNotFound, id)
}

func (m *Memory) UpdateSession(ses session.Session) error {
	for _, s := range m.sessions {
		if s.ID() == ses.ID() {
			s = ses
			return nil
		}
	}

	return models.ErrNotFound
}

func (m *Memory) SessionGC() error {
	sessions := make([]session.Session, 0)
	for _, s := range m.sessions {
		if !s.Expiry().After(time.Now()) {
			sessions = append(sessions, s)
		}
	}
	m.sessions = sessions
	return nil
}
