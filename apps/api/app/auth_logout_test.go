package app

import (
	"testing"

	"github.com/samuelsih/golib/assert"
)

func TestLogout(t *testing.T) {
	email := newEmail()
	registerUser(t, email)

	first := loginUser(t, email)
	second := loginUser(t, email)

	assert.NoError(t, Logout(t.Context(), db, first.Token))
	assert.Equal(t, countSessions(t, first.Token), 0)
	assert.Equal(t, countSessions(t, second.Token), 1)

	assert.NoError(t, Logout(t.Context(), db, "unknown-token"))
}
