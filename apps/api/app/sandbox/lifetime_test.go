package sandbox

import (
	"testing"
	"time"

	"github.com/samuelsih/golib/assert"
)

func TestNewLifetime(t *testing.T) {
	tests := []struct {
		name       string
		autoStop   time.Duration
		autoDelete time.Duration
		wantErr    error
	}{
		{name: "disabled"},
		{name: "default", autoStop: time.Hour},
		{name: "minimum stop", autoStop: MinAutoStopAfter},
		{name: "maximum stop", autoStop: MaxAutoStopAfter},
		{name: "maximum delete", autoStop: time.Hour, autoDelete: MaxAutoDeleteAfter},
		{name: "negative stop", autoStop: -time.Minute, wantErr: ErrInvalidLifetime},
		{name: "too short stop", autoStop: time.Second, wantErr: ErrInvalidLifetime},
		{name: "too long stop", autoStop: MaxAutoStopAfter + time.Minute, wantErr: ErrInvalidLifetime},
		{name: "negative delete", autoDelete: -time.Minute, wantErr: ErrInvalidLifetime},
		{name: "too long delete", autoDelete: MaxAutoDeleteAfter + time.Minute, wantErr: ErrInvalidLifetime},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lifetime, err := NewLifetime(tt.autoStop, tt.autoDelete)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Equal(t, lifetime, Lifetime{})
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, lifetime.AutoStopAfter, tt.autoStop)
			assert.Equal(t, lifetime.AutoDeleteAfter, tt.autoDelete)
		})
	}
}

func TestLifetimeDefaults(t *testing.T) {
	assert.Equal(t, DefaultLifetime.AutoStopAfter, time.Hour)
	assert.Equal(t, DefaultLifetime.AutoDeleteAfter, time.Duration(0))
}

func TestLifetimeDeadlines(t *testing.T) {
	disabled, err := NewLifetime(0, 0)
	assert.NoError(t, err)

	_, ok := disabled.StopDeadline(testNow)
	assert.False(t, ok)

	_, ok = disabled.DeleteDeadline(testNow)
	assert.False(t, ok)

	lifetime, err := NewLifetime(2*time.Hour, 30*time.Minute)
	assert.NoError(t, err)

	stopDeadline, ok := lifetime.StopDeadline(testNow)
	assert.True(t, ok)
	assert.Equal(t, stopDeadline, testNow.Add(2*time.Hour))

	deleteDeadline, ok := lifetime.DeleteDeadline(testNow)
	assert.True(t, ok)
	assert.Equal(t, deleteDeadline, testNow.Add(30*time.Minute))
}
