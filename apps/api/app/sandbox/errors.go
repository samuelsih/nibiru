package sandbox

import "errors"

var (
	ErrInvalidSpec       = errors.New("invalid sandbox spec")
	ErrInvalidSource     = errors.New("invalid sandbox source")
	ErrInvalidLifetime   = errors.New("invalid sandbox lifetime")
	ErrInvalidMachineIP  = errors.New("invalid sandbox machine ip")
	ErrInvalidTransition = errors.New("invalid sandbox state transition")
	ErrOwnerRequired     = errors.New("sandbox owner is required")
	ErrDiskShrink        = errors.New("sandbox disk can only grow")
	ErrResizeNotAllowed  = errors.New("sandbox cannot be resized in its current state")
	ErrAutoStopNotDue    = errors.New("sandbox auto-stop is not due")
	ErrAutoDeleteNotDue  = errors.New("sandbox auto-delete is not due")
)
