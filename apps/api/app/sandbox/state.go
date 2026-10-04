package sandbox

import (
	"fmt"
	"slices"
)

type State string

const (
	StateCreating State = "creating"
	StateRunning  State = "running"
	StateStopping State = "stopping"
	StateStopped  State = "stopped"
	StateStarting State = "starting"
	StateDeleting State = "deleting"
	StateDeleted  State = "deleted"
	StateFailed   State = "failed"
)

var transitions = map[State][]State{
	StateCreating: {StateRunning, StateDeleting, StateFailed},
	StateRunning:  {StateStopping, StateDeleting, StateFailed},
	StateStopping: {StateStopped, StateDeleting, StateFailed},
	StateStopped:  {StateStarting, StateDeleting, StateFailed},
	StateStarting: {StateRunning, StateDeleting, StateFailed},
	StateDeleting: {StateDeleted, StateFailed},
	StateFailed:   {StateDeleting},
	StateDeleted:  nil,
}

func (s State) Terminal() bool {
	return s == StateDeleted
}

func (s State) CanTransitionTo(next State) bool {
	return slices.Contains(transitions[s], next)
}

type TransitionError struct {
	From State
	To   State
}

func (e *TransitionError) Error() string {
	return fmt.Sprintf("%s: %s -> %s", ErrInvalidTransition, e.From, e.To)
}

func (e *TransitionError) Unwrap() error {
	return ErrInvalidTransition
}
