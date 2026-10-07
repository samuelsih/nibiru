package app

import (
	"errors"
	"fmt"
	"slices"
)

var ErrInvalidTransition = errors.New("invalid sandbox state transition")

type SandboxState string

const (
	StateCreating SandboxState = "creating"
	StateRunning  SandboxState = "running"
	StateStopping SandboxState = "stopping"
	StateStopped  SandboxState = "stopped"
	StateStarting SandboxState = "starting"
	StateDeleting SandboxState = "deleting"
	StateDeleted  SandboxState = "deleted"
	StateFailed   SandboxState = "failed"
)

var transitions = map[SandboxState][]SandboxState{
	StateCreating: {StateRunning, StateDeleting, StateFailed},
	StateRunning:  {StateStopping, StateDeleting, StateFailed},
	StateStopping: {StateStopped, StateDeleting, StateFailed},
	StateStopped:  {StateStarting, StateDeleting, StateFailed},
	StateStarting: {StateRunning, StateDeleting, StateFailed},
	StateDeleting: {StateDeleted, StateFailed},
	StateFailed:   {StateDeleting},
	StateDeleted:  nil,
}

func (s SandboxState) CanTransitionTo(next SandboxState) bool {
	return slices.Contains(transitions[s], next)
}

type TransitionError struct {
	From SandboxState
	To   SandboxState
}

func (e *TransitionError) Error() string {
	return fmt.Sprintf("%s: %s -> %s", ErrInvalidTransition, e.From, e.To)
}

func (e *TransitionError) Unwrap() error {
	return ErrInvalidTransition
}
