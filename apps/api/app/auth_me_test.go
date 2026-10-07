package app

import (
	"testing"
	"uuid"

	"github.com/samuelsih/golib/assert"
)

func TestWhoAmI(t *testing.T) {
	email := newEmail()
	user := registerUser(t, email)
	session := loginUser(t, email)

	expiredToken := uuid.NewV7().String()
	_, err := db.Exec(t.Context(),
		`INSERT INTO sessions (id, token, user_id, created_at, expires_at)
		 VALUES ($1, $2, $3, NOW() - INTERVAL '2 hours', NOW() - INTERVAL '1 hour')`,
		uuid.NewV7(), expiredToken, user.ID)
	assert.NoError(t, err)

	found, err := WhoAmI(t.Context(), db, session.Token)
	assert.NoError(t, err)
	assert.Equal(t, found, user)

	_, err = WhoAmI(t.Context(), db, "unknown-token")
	assert.ErrorIs(t, err, ErrInvalidSession)

	_, err = WhoAmI(t.Context(), db, expiredToken)
	assert.ErrorIs(t, err, ErrInvalidSession)
}
