package auth

import (
	"time"

	"github.com/google/uuid"
	"github.com/samuelsih/golib/stringx"
)

type Session struct {
	ID        uuid.UUID
	Token     string
	UserID    uuid.UUID
	CreatedAt time.Time
	ExpiresAt time.Time
}

func NewSession(userID uuid.UUID, ttl time.Duration) Session {
	now := time.Now()

	return Session{
		ID:        uuid.New(),
		Token:     stringx.Base64(40),
		UserID:    userID,
		CreatedAt: now,
		ExpiresAt: now.Add(ttl),
	}
}

func (s Session) Expired() bool {
	return time.Now().After(s.ExpiresAt)
}
