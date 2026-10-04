package core

import (
	"fmt"
	"net/netip"
	"time"
	"uuid"
)

const DefaultImage = "ubuntu:24.04"

type Instance struct {
	ID          uuid.UUID
	OwnerID     uuid.UUID
	Name        string
	Spec        Spec
	Image       string
	Snapshot    uuid.UUID
	Lifetime    Lifetime
	State       State
	MachineIP   netip.Addr
	DesktopURL  string
	HostedURL   string
	Usage       Meter
	ErrorReason string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	StartedAt   time.Time
	StoppedAt   time.Time
	DeletedAt   time.Time
}

type SpawnRequest struct {
	OwnerID  uuid.UUID
	Name     string
	Spec     Spec
	Image    string
	Snapshot uuid.UUID
	Lifetime *Lifetime
}

func Spawn(req SpawnRequest, now time.Time) (Instance, error) {
	if req.OwnerID == uuid.Nil() {
		return Instance{}, ErrOwnerRequired
	}

	if req.Spec.IsZero() {
		return Instance{}, fmt.Errorf("%w: spec is required", ErrInvalidSpec)
	}

	spec, err := NewSpec(req.Spec.CPU, req.Spec.MemoryGB, req.Spec.DiskGB)
	if err != nil {
		return Instance{}, err
	}

	image, snapshot := req.Image, req.Snapshot
	switch {
	case image == "" && snapshot == uuid.Nil():
		image = DefaultImage
	case image != "" && snapshot != uuid.Nil():
		return Instance{}, fmt.Errorf("%w: image and snapshot are mutually exclusive", ErrInvalidSource)
	}

	lifetime := DefaultLifetime
	if req.Lifetime != nil {
		validated, err := NewLifetime(req.Lifetime.AutoStopAfter, req.Lifetime.AutoDeleteAfter)
		if err != nil {
			return Instance{}, err
		}

		lifetime = validated
	}

	id := uuid.NewV7()

	name := req.Name
	if name == "" {
		name = "sandbox-" + id.String()[:8]
	}

	return Instance{
		ID:        id,
		OwnerID:   req.OwnerID,
		Name:      name,
		Spec:      spec,
		Image:     image,
		Snapshot:  snapshot,
		Lifetime:  lifetime,
		State:     StateCreating,
		Usage:     NewMeter(),
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func (s *Instance) MarkRunning(machineIP netip.Addr, desktopURL, hostedURL string, now time.Time) error {
	if !machineIP.IsValid() {
		return ErrInvalidMachineIP
	}

	if err := s.transitionTo(StateRunning, now); err != nil {
		return err
	}

	s.MachineIP = machineIP
	s.DesktopURL = desktopURL
	s.HostedURL = hostedURL
	s.ErrorReason = ""
	s.StartedAt = now
	s.StoppedAt = time.Time{}
	s.Usage = s.Usage.Start(now)

	return nil
}

func (s *Instance) Stop(now time.Time) error {
	return s.transitionTo(StateStopping, now)
}

func (s *Instance) MarkStopped(now time.Time) error {
	if err := s.transitionTo(StateStopped, now); err != nil {
		return err
	}

	s.StoppedAt = now
	s.Usage = s.Usage.Pause(now)

	return nil
}

func (s *Instance) Start(now time.Time) error {
	return s.transitionTo(StateStarting, now)
}

func (s *Instance) Delete(now time.Time) error {
	if err := s.transitionTo(StateDeleting, now); err != nil {
		return err
	}

	s.Usage = s.Usage.Pause(now)

	return nil
}

func (s *Instance) MarkDeleted(now time.Time) error {
	if err := s.transitionTo(StateDeleted, now); err != nil {
		return err
	}

	s.DeletedAt = now
	s.MachineIP = netip.Addr{}
	s.DesktopURL = ""
	s.HostedURL = ""

	return nil
}

func (s *Instance) Fail(reason string, now time.Time) error {
	if err := s.transitionTo(StateFailed, now); err != nil {
		return err
	}

	s.ErrorReason = reason
	s.Usage = s.Usage.Pause(now)

	return nil
}

func (s *Instance) Resize(spec Spec, now time.Time) error {
	if s.State != StateStopped {
		return fmt.Errorf("%w: sandbox is %s", ErrResizeNotAllowed, s.State)
	}

	next, err := s.Spec.Resize(spec)
	if err != nil {
		return err
	}

	s.Spec = next
	s.UpdatedAt = now

	return nil
}

func (s Instance) StopDeadline() (time.Time, bool) {
	if s.State != StateRunning || s.StartedAt.IsZero() {
		return time.Time{}, false
	}

	return s.Lifetime.StopDeadline(s.StartedAt)
}

func (s Instance) ShouldAutoStop(now time.Time) bool {
	deadline, ok := s.StopDeadline()

	return ok && !now.Before(deadline)
}

func (s *Instance) AutoStop(now time.Time) error {
	if !s.ShouldAutoStop(now) {
		return ErrAutoStopNotDue
	}

	return s.Stop(now)
}

func (s Instance) DeleteDeadline() (time.Time, bool) {
	if s.State != StateStopped || s.StoppedAt.IsZero() {
		return time.Time{}, false
	}

	return s.Lifetime.DeleteDeadline(s.StoppedAt)
}

func (s Instance) ShouldAutoDelete(now time.Time) bool {
	deadline, ok := s.DeleteDeadline()

	return ok && !now.Before(deadline)
}

func (s *Instance) AutoDelete(now time.Time) error {
	if !s.ShouldAutoDelete(now) {
		return ErrAutoDeleteNotDue
	}

	return s.Delete(now)
}

func (s Instance) RunningDuration(now time.Time) time.Duration {
	return s.Usage.Total(now)
}

func (s Instance) SSH(user string) string {
	return user + "@" + s.MachineIP.String()
}

func (s Instance) IsRunning() bool {
	return s.State == StateRunning
}

func (s Instance) IsStopped() bool {
	return s.State == StateStopped
}

func (s Instance) IsTerminal() bool {
	return s.State.Terminal()
}

func (s Instance) IsActive() bool {
	return !s.State.Terminal() && s.State != StateFailed
}

func (s *Instance) transitionTo(next State, now time.Time) error {
	if !s.State.CanTransitionTo(next) {
		return &TransitionError{From: s.State, To: next}
	}

	s.State = next
	s.UpdatedAt = now

	return nil
}
