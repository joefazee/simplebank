package token

import (
	"fmt"
	"time"

	"golang.org/x/crypto/chacha20poly1305"

	"github.com/o1egl/paseto"
)

// PasetoMaker is a Paseto token maker
type PasetoMaker struct {
	peseto       *paseto.V2
	symmetricKey []byte
}

func NewPasetoMaker(symmetricKey string) (Maker, error) {
	if len(symmetricKey) != chacha20poly1305.KeySize {
		return nil, fmt.Errorf("invalid key size: key must be %d characters", chacha20poly1305.KeySize)
	}

	maker := PasetoMaker{
		peseto:       paseto.NewV2(),
		symmetricKey: []byte(symmetricKey),
	}

	return &maker, nil
}

// CreateToken creates a new token for a specific username and duration
func (m *PasetoMaker) CreateToken(username string, duration time.Duration) (string, error) {

	payload, err := NewPayload(username, duration)
	if err != nil {
		return "", err
	}
	return m.peseto.Encrypt(m.symmetricKey, payload, nil)
}

// VerifyToken checks if the token is valid or not
func (m *PasetoMaker) VerifyToken(token string) (*Payload, error) {

	payload := &Payload{}
	err := m.peseto.Decrypt(token, m.symmetricKey, payload, nil)

	if err != nil {
		return nil, ErrInvalidToken
	}

	err = payload.Valid()
	if err != nil {
		return nil, err
	}

	return payload, nil
}
