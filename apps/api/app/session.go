package app

import (
	"time"
	"uuid"
)

type Session struct {
	ID        uuid.UUID `db:"id"`
	Token     string    `db:"token"`
	UserID    uuid.UUID `db:"user_id"`
	CreatedAt time.Time `db:"created_at"`
	ExpiresAt time.Time `db:"expires_at"`
}

func (s Session) Expired() bool {
	return time.Now().After(s.ExpiresAt)
}
