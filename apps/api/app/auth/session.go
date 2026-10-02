package auth

import (
	"time"

	"github.com/google/uuid"
	"github.com/samuelsih/golib/stringx"
)

type Session struct {
	ID        uuid.UUID `db:"id"`
	Token     string    `db:"token"`
	UserID    uuid.UUID `db:"user_id"`
	CreatedAt time.Time `db:"created_at"`
	ExpiresAt time.Time `db:"expires_at"`
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
