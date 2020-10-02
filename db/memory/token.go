package memory

import (
	"time"

	"github.com/djangulo/sfd/crypto/token"
	"github.com/djangulo/sfd/db/models"
)

func (m *Memory) GetToken(digest string, kind token.Kind) (*token.Token, error) {

	for _, t := range m.tokens {
		if t.Digest == digest && t.Kind == kind {
			return t, nil
		}
	}
	return nil, models.ErrNotFound
}

func (m *Memory) DeleteToken(digest string) error {
	var nt = make([]*token.Token, 0)
	for _, token := range m.tokens {
		if token.Digest == digest {
			continue
		}
		nt = append(nt, token)
	}
	m.tokens = nt
	return nil
}

func (m *Memory) SaveToken(t *token.Token) error {
	m.tokens = append(m.tokens, t)
	return nil
}

func (m *Memory) TokenGC() error {
	tokens := make([]*token.Token, 0)
	for _, t := range m.tokens {
		if !t.Expires.After(time.Now()) {
			tokens = append(tokens, t)
		}
	}
	m.tokens = tokens
	return nil
}
