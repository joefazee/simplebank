package token

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Different types of error that returned from the VerifyToken
var (
	ErrExpiredToken = errors.New("token has expired")
	ErrInvalidToken = errors.New("invalid token")
)

// Payload contains the payload data of the token
type Payload struct {
	ID        uuid.UUID
	Username  string
	IssuedAt  time.Time
	ExpiredAt time.Time
}

// NewPayload creates a new token with a specific username and duration
func NewPayload(username string, duration time.Duration) (*Payload, error) {
	tokenID, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	payload := &Payload{
		ID:        tokenID,
		Username:  username,
		IssuedAt:  time.Now(),
		ExpiredAt: time.Now().Add(duration),
	}

	return payload, nil
}

func (p *Payload) Valid() error {

	if time.Now().After(p.ExpiredAt) {
		return ErrExpiredToken
	}

	return nil
}
