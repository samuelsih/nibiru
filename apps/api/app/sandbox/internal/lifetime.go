package core

import (
	"fmt"
	"time"
)

type Lifetime struct {
	AutoStopAfter   time.Duration
	AutoDeleteAfter time.Duration
}

const (
	MinAutoStopAfter   = time.Minute
	MaxAutoStopAfter   = 7 * 24 * time.Hour
	MaxAutoDeleteAfter = 30 * 24 * time.Hour
)

var DefaultLifetime = Lifetime{AutoStopAfter: time.Hour}

func NewLifetime(autoStopAfter, autoDeleteAfter time.Duration) (Lifetime, error) {
	switch {
	case autoStopAfter < 0:
		return Lifetime{}, fmt.Errorf("%w: auto-stop must not be negative", ErrInvalidLifetime)
	case autoStopAfter > 0 && autoStopAfter < MinAutoStopAfter:
		return Lifetime{}, fmt.Errorf("%w: auto-stop must be at least %s", ErrInvalidLifetime, MinAutoStopAfter)
	case autoStopAfter > MaxAutoStopAfter:
		return Lifetime{}, fmt.Errorf("%w: auto-stop must be at most %s", ErrInvalidLifetime, MaxAutoStopAfter)
	case autoDeleteAfter < 0:
		return Lifetime{}, fmt.Errorf("%w: auto-delete must not be negative", ErrInvalidLifetime)
	case autoDeleteAfter > MaxAutoDeleteAfter:
		return Lifetime{}, fmt.Errorf("%w: auto-delete must be at most %s", ErrInvalidLifetime, MaxAutoDeleteAfter)
	}

	return Lifetime{AutoStopAfter: autoStopAfter, AutoDeleteAfter: autoDeleteAfter}, nil
}

func (l Lifetime) StopDeadline(startedAt time.Time) (time.Time, bool) {
	if l.AutoStopAfter <= 0 {
		return time.Time{}, false
	}

	return startedAt.Add(l.AutoStopAfter), true
}

func (l Lifetime) DeleteDeadline(stoppedAt time.Time) (time.Time, bool) {
	if l.AutoDeleteAfter <= 0 {
		return time.Time{}, false
	}

	return stoppedAt.Add(l.AutoDeleteAfter), true
}
