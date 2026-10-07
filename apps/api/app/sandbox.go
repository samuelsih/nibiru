package app

import (
	"errors"
	"net"
	"time"
	"uuid"
)

var (
	ErrInvalidSpec = errors.New("invalid sandbox spec")
	ErrDiskShrink  = errors.New("sandbox disk can only grow")
)

type Sandbox struct {
	ID        uuid.UUID
	OwnerID   uuid.UUID
	Name      string
	CPU       int
	MemoryGB  int
	DiskGB    int
	Image     string
	Snapshot  uuid.UUID
	State     SandboxState
	MachineIP net.IP
	CreatedAt time.Time
	UpdatedAt time.Time
	StartedAt time.Time
	StoppedAt time.Time
}
