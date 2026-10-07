package app

import (
	"testing"

	"github.com/samuelsih/golib/assert"
)

func TestSandboxStateCanTransitionTo(t *testing.T) {
	tests := []struct {
		name string
		from SandboxState
		to   SandboxState
		want bool
	}{
		{name: "creating to running", from: StateCreating, to: StateRunning, want: true},
		{name: "creating to deleting", from: StateCreating, to: StateDeleting, want: true},
		{name: "creating to stopped", from: StateCreating, to: StateStopped, want: false},
		{name: "running to stopping", from: StateRunning, to: StateStopping, want: true},
		{name: "running to deleted", from: StateRunning, to: StateDeleted, want: false},
		{name: "stopping to stopped", from: StateStopping, to: StateStopped, want: true},
		{name: "stopping to running", from: StateStopping, to: StateRunning, want: false},
		{name: "stopped to starting", from: StateStopped, to: StateStarting, want: true},
		{name: "stopped to running", from: StateStopped, to: StateRunning, want: false},
		{name: "starting to running", from: StateStarting, to: StateRunning, want: true},
		{name: "starting to stopped", from: StateStarting, to: StateStopped, want: false},
		{name: "deleting to deleted", from: StateDeleting, to: StateDeleted, want: true},
		{name: "deleting to running", from: StateDeleting, to: StateRunning, want: false},
		{name: "failed to deleting", from: StateFailed, to: StateDeleting, want: true},
		{name: "failed to running", from: StateFailed, to: StateRunning, want: false},
		{name: "deleted to running", from: StateDeleted, to: StateRunning, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.from.CanTransitionTo(tt.to), tt.want)
		})
	}
}

func TestTransitionError(t *testing.T) {
	err := &TransitionError{From: StateRunning, To: StateStopped}

	assert.ErrorIs(t, err, ErrInvalidTransition)
	assert.Equal(t, err.Error(), "invalid sandbox state transition: running -> stopped")
}
