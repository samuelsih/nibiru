package app

import (
	"testing"
	"time"

	"github.com/samuelsih/golib/assert"
)

func TestSessionExpired(t *testing.T) {
	assert.True(t, Session{ExpiresAt: time.Now().Add(-time.Second)}.Expired())
	assert.False(t, Session{ExpiresAt: time.Now().Add(time.Second)}.Expired())
}
