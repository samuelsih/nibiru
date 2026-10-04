package sandbox

import (
	"net/netip"
	"testing"
	"time"
	"uuid"

	"github.com/samuelsih/golib/assert"
)

const (
	testDesktopURL = "https://desktop.example"
	testHostedURL  = "https://app.example"
)

var (
	testNow       = time.Date(2026, time.October, 4, 12, 0, 0, 0, time.UTC)
	testMachineIP = netip.MustParseAddr("203.0.113.42")
)

func newSpec(t *testing.T, cpu, memoryGB, diskGB int) Spec {
	t.Helper()

	spec, err := NewSpec(cpu, memoryGB, diskGB)
	assert.NoError(t, err)

	return spec
}

func newLifetime(t *testing.T, autoStopAfter, autoDeleteAfter time.Duration) Lifetime {
	t.Helper()

	lifetime, err := NewLifetime(autoStopAfter, autoDeleteAfter)
	assert.NoError(t, err)

	return lifetime
}

func spawnInstance(t *testing.T) Instance {
	t.Helper()

	s, err := Spawn(SpawnRequest{OwnerID: uuid.NewV7(), Spec: DefaultSpec}, testNow)
	assert.NoError(t, err)

	return s
}

func runInstance(t *testing.T, s Instance, at time.Time) Instance {
	t.Helper()

	assert.NoError(t, s.MarkRunning(testMachineIP, testDesktopURL, testHostedURL, at))

	return s
}

func stopInstance(t *testing.T, s Instance, at time.Time) Instance {
	t.Helper()

	assert.NoError(t, s.Stop(at))
	assert.NoError(t, s.MarkStopped(at))

	return s
}

func runningInstance(t *testing.T) Instance {
	t.Helper()

	return runInstance(t, spawnInstance(t), testNow)
}

func stoppedInstance(t *testing.T) Instance {
	t.Helper()

	return stopInstance(t, runningInstance(t), testNow.Add(time.Minute))
}

func failedInstance(t *testing.T) Instance {
	t.Helper()

	s := spawnInstance(t)
	assert.NoError(t, s.Fail("boom", testNow))

	return s
}

func deletedInstance(t *testing.T) Instance {
	t.Helper()

	s := stopInstance(t, runningInstance(t), testNow.Add(time.Minute))
	assert.NoError(t, s.Delete(testNow.Add(2*time.Minute)))
	assert.NoError(t, s.MarkDeleted(testNow.Add(3*time.Minute)))

	return s
}

func TestSpawn(t *testing.T) {
	owner := uuid.NewV7()

	t.Run("defaults", func(t *testing.T) {
		s, err := Spawn(SpawnRequest{OwnerID: owner, Spec: DefaultSpec}, testNow)
		assert.NoError(t, err)

		assert.NotEqual(t, s.ID, uuid.Nil())
		assert.True(t, s.ID[6]>>4 == 7)
		assert.Equal(t, s.OwnerID, owner)
		assert.Equal(t, s.Name, "sandbox-"+s.ID.String()[:8])
		assert.Equal(t, s.Spec, DefaultSpec)
		assert.Equal(t, s.Image, DefaultImage)
		assert.Equal(t, s.Snapshot, uuid.Nil())
		assert.Equal(t, s.Lifetime, DefaultLifetime)
		assert.Equal(t, s.State, StateCreating)
		assert.Equal(t, s.CreatedAt, testNow)
		assert.Equal(t, s.UpdatedAt, testNow)
		assert.True(t, s.StartedAt.IsZero())
		assert.True(t, s.StoppedAt.IsZero())
		assert.True(t, s.DeletedAt.IsZero())
	})

	t.Run("explicit values", func(t *testing.T) {
		lifetime := newLifetime(t, 2*time.Hour, 30*time.Minute)

		s, err := Spawn(SpawnRequest{
			OwnerID:  owner,
			Name:     "worker-1",
			Spec:     LargeSpec,
			Image:    "debian:13",
			Lifetime: &lifetime,
		}, testNow)
		assert.NoError(t, err)

		assert.Equal(t, s.Name, "worker-1")
		assert.Equal(t, s.Spec, LargeSpec)
		assert.Equal(t, s.Image, "debian:13")
		assert.Equal(t, s.Lifetime, lifetime)
	})

	t.Run("snapshot source", func(t *testing.T) {
		snapshot := uuid.NewV7()

		s, err := Spawn(SpawnRequest{
			OwnerID:  owner,
			Spec:     DefaultSpec,
			Snapshot: snapshot,
		}, testNow)
		assert.NoError(t, err)

		assert.Equal(t, s.Snapshot, snapshot)
		assert.Equal(t, s.Image, "")
	})

	tests := []struct {
		name    string
		req     SpawnRequest
		wantErr error
	}{
		{
			name:    "missing owner",
			req:     SpawnRequest{Spec: DefaultSpec},
			wantErr: ErrOwnerRequired,
		},
		{
			name:    "missing spec",
			req:     SpawnRequest{OwnerID: owner},
			wantErr: ErrInvalidSpec,
		},
		{
			name: "ambiguous source",
			req: SpawnRequest{
				OwnerID:  owner,
				Spec:     DefaultSpec,
				Image:    DefaultImage,
				Snapshot: uuid.NewV7(),
			},
			wantErr: ErrInvalidSource,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := Spawn(tt.req, testNow)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, s, Instance{})
		})
	}
}

func TestInstanceLifecycle(t *testing.T) {
	s := spawnInstance(t)
	assert.Equal(t, s.State, StateCreating)

	runningAt := testNow.Add(time.Minute)
	assert.NoError(t, s.MarkRunning(testMachineIP, testDesktopURL, testHostedURL, runningAt))
	assert.Equal(t, s.State, StateRunning)
	assert.Equal(t, s.MachineIP, testMachineIP)
	assert.Equal(t, s.DesktopURL, testDesktopURL)
	assert.Equal(t, s.HostedURL, testHostedURL)
	assert.Equal(t, s.SSH("ubuntu"), "ubuntu@203.0.113.42")
	assert.Equal(t, s.StartedAt, runningAt)
	assert.Equal(t, s.UpdatedAt, runningAt)
	assert.True(t, s.Usage.Running())

	stoppingAt := runningAt.Add(10 * time.Minute)
	assert.NoError(t, s.Stop(stoppingAt))
	assert.Equal(t, s.State, StateStopping)

	stoppedAt := stoppingAt.Add(time.Second)
	assert.NoError(t, s.MarkStopped(stoppedAt))
	assert.Equal(t, s.State, StateStopped)
	assert.Equal(t, s.StoppedAt, stoppedAt)
	assert.False(t, s.Usage.Running())

	startingAt := stoppedAt.Add(time.Minute)
	assert.NoError(t, s.Start(startingAt))
	assert.Equal(t, s.State, StateStarting)

	resumedAt := startingAt.Add(time.Second)
	assert.NoError(t, s.MarkRunning(testMachineIP, testDesktopURL, testHostedURL, resumedAt))
	assert.Equal(t, s.State, StateRunning)
	assert.Equal(t, s.StartedAt, resumedAt)
	assert.True(t, s.StoppedAt.IsZero())

	deletingAt := resumedAt.Add(time.Minute)
	assert.NoError(t, s.Delete(deletingAt))
	assert.Equal(t, s.State, StateDeleting)

	deletedAt := deletingAt.Add(time.Second)
	assert.NoError(t, s.MarkDeleted(deletedAt))
	assert.Equal(t, s.State, StateDeleted)
	assert.Equal(t, s.DeletedAt, deletedAt)
	assert.Equal(t, s.MachineIP, netip.Addr{})
	assert.Equal(t, s.DesktopURL, "")
	assert.Equal(t, s.HostedURL, "")
	assert.True(t, s.State.Terminal())
	assert.True(t, s.IsTerminal())
	assert.False(t, s.IsActive())
}

func TestInstanceInvalidTransitions(t *testing.T) {
	tests := []struct {
		name    string
		prepare func(t *testing.T) Instance
		act     func(s *Instance) error
		from    State
		to      State
	}{
		{
			name:    "run twice",
			prepare: runningInstance,
			act:     func(s *Instance) error { return s.MarkRunning(testMachineIP, testDesktopURL, testHostedURL, testNow) },
			from:    StateRunning,
			to:      StateRunning,
		},
		{
			name:    "stop while creating",
			prepare: spawnInstance,
			act:     func(s *Instance) error { return s.Stop(testNow) },
			from:    StateCreating,
			to:      StateStopping,
		},
		{
			name:    "stop while stopped",
			prepare: stoppedInstance,
			act:     func(s *Instance) error { return s.Stop(testNow) },
			from:    StateStopped,
			to:      StateStopping,
		},
		{
			name:    "mark stopped while running",
			prepare: runningInstance,
			act:     func(s *Instance) error { return s.MarkStopped(testNow) },
			from:    StateRunning,
			to:      StateStopped,
		},
		{
			name:    "start while running",
			prepare: runningInstance,
			act:     func(s *Instance) error { return s.Start(testNow) },
			from:    StateRunning,
			to:      StateStarting,
		},
		{
			name:    "mark deleted while running",
			prepare: runningInstance,
			act:     func(s *Instance) error { return s.MarkDeleted(testNow) },
			from:    StateRunning,
			to:      StateDeleted,
		},
		{
			name:    "fail twice",
			prepare: failedInstance,
			act:     func(s *Instance) error { return s.Fail("boom again", testNow) },
			from:    StateFailed,
			to:      StateFailed,
		},
		{
			name:    "run after failed",
			prepare: failedInstance,
			act:     func(s *Instance) error { return s.MarkRunning(testMachineIP, testDesktopURL, testHostedURL, testNow) },
			from:    StateFailed,
			to:      StateRunning,
		},
		{
			name:    "delete twice",
			prepare: deletedInstance,
			act:     func(s *Instance) error { return s.Delete(testNow) },
			from:    StateDeleted,
			to:      StateDeleting,
		},
		{
			name:    "start when deleted",
			prepare: deletedInstance,
			act:     func(s *Instance) error { return s.Start(testNow) },
			from:    StateDeleted,
			to:      StateStarting,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.prepare(t)
			err := tt.act(&s)

			assert.ErrorIs(t, err, ErrInvalidTransition)

			transition := assert.ErrorAsType[*TransitionError](t, err)
			if transition == nil {
				return
			}

			assert.Equal(t, transition.From, tt.from)
			assert.Equal(t, transition.To, tt.to)
		})
	}
}

func TestInstanceMarkRunningRejectsInvalidMachineIP(t *testing.T) {
	s := spawnInstance(t)

	err := s.MarkRunning(netip.Addr{}, testDesktopURL, testHostedURL, testNow)
	assert.ErrorIs(t, err, ErrInvalidMachineIP)
	assert.Equal(t, s.State, StateCreating)
	assert.Equal(t, s.Usage.Total(testNow), time.Duration(0))
}

func TestInstanceFail(t *testing.T) {
	s := runInstance(t, spawnInstance(t), testNow)

	assert.NoError(t, s.Fail("boot failed", testNow.Add(30*time.Second)))
	assert.Equal(t, s.State, StateFailed)
	assert.Equal(t, s.ErrorReason, "boot failed")
	assert.False(t, s.Usage.Running())
	assert.Equal(t, s.RunningDuration(testNow.Add(time.Hour)), 30*time.Second)
	assert.False(t, s.IsActive())
}

func TestInstanceAutoStop(t *testing.T) {
	lifetime := newLifetime(t, 2*time.Hour, 0)

	s, err := Spawn(SpawnRequest{OwnerID: uuid.NewV7(), Spec: DefaultSpec, Lifetime: &lifetime}, testNow)
	assert.NoError(t, err)

	deadline, ok := s.StopDeadline()
	assert.False(t, ok)
	assert.True(t, deadline.IsZero())
	assert.False(t, s.ShouldAutoStop(testNow))

	assert.NoError(t, s.MarkRunning(testMachineIP, testDesktopURL, testHostedURL, testNow))

	deadline, ok = s.StopDeadline()
	assert.True(t, ok)
	assert.Equal(t, deadline, testNow.Add(2*time.Hour))
	assert.False(t, s.ShouldAutoStop(testNow.Add(2*time.Hour-time.Second)))
	assert.True(t, s.ShouldAutoStop(testNow.Add(2*time.Hour)))

	err = s.AutoStop(testNow.Add(2*time.Hour - time.Second))
	assert.ErrorIs(t, err, ErrAutoStopNotDue)
	assert.Equal(t, s.State, StateRunning)

	assert.NoError(t, s.AutoStop(testNow.Add(2*time.Hour)))
	assert.Equal(t, s.State, StateStopping)
}

func TestInstanceAutoDelete(t *testing.T) {
	lifetime := newLifetime(t, 0, time.Hour)

	s, err := Spawn(SpawnRequest{OwnerID: uuid.NewV7(), Spec: DefaultSpec, Lifetime: &lifetime}, testNow)
	assert.NoError(t, err)
	assert.NoError(t, s.MarkRunning(testMachineIP, testDesktopURL, testHostedURL, testNow))

	stoppedAt := testNow.Add(10 * time.Minute)
	assert.NoError(t, s.Stop(stoppedAt))
	assert.NoError(t, s.MarkStopped(stoppedAt))

	deadline, ok := s.DeleteDeadline()
	assert.True(t, ok)
	assert.Equal(t, deadline, stoppedAt.Add(time.Hour))

	assert.False(t, s.ShouldAutoDelete(stoppedAt.Add(time.Hour-time.Second)))
	assert.True(t, s.ShouldAutoDelete(stoppedAt.Add(time.Hour)))

	err = s.AutoDelete(stoppedAt.Add(time.Hour - time.Second))
	assert.ErrorIs(t, err, ErrAutoDeleteNotDue)
	assert.Equal(t, s.State, StateStopped)

	assert.NoError(t, s.AutoDelete(stoppedAt.Add(time.Hour)))
	assert.Equal(t, s.State, StateDeleting)
}

func TestInstanceMeter(t *testing.T) {
	s := runInstance(t, spawnInstance(t), testNow)

	assert.Equal(t, s.RunningDuration(testNow.Add(90*time.Second)), 90*time.Second)

	stopAt := testNow.Add(2 * time.Minute)
	assert.NoError(t, s.Stop(stopAt))
	assert.NoError(t, s.MarkStopped(stopAt))
	assert.Equal(t, s.RunningDuration(stopAt.Add(time.Hour)), 2*time.Minute)

	resumeAt := stopAt.Add(time.Hour)
	assert.NoError(t, s.Start(resumeAt))
	assert.NoError(t, s.MarkRunning(testMachineIP, testDesktopURL, testHostedURL, resumeAt.Add(time.Second)))
	assert.Equal(t, s.RunningDuration(resumeAt.Add(31*time.Second)), 2*time.Minute+30*time.Second)

	assert.NoError(t, s.Fail("boom", resumeAt.Add(time.Minute)))
	assert.Equal(t, s.RunningDuration(resumeAt.Add(time.Hour)), 2*time.Minute+59*time.Second)
}

func TestInstanceResize(t *testing.T) {
	s := runInstance(t, spawnInstance(t), testNow)

	err := s.Resize(newSpec(t, 4, 8, 100), testNow.Add(time.Minute))
	assert.ErrorIs(t, err, ErrResizeNotAllowed)
	assert.Equal(t, s.Spec, DefaultSpec)

	assert.NoError(t, s.Stop(testNow.Add(time.Minute)))
	assert.NoError(t, s.MarkStopped(testNow.Add(time.Minute)))

	next := newSpec(t, 4, 8, 100)
	assert.NoError(t, s.Resize(next, testNow.Add(2*time.Minute)))
	assert.Equal(t, s.Spec, next)
	assert.Equal(t, s.UpdatedAt, testNow.Add(2*time.Minute))

	err = s.Resize(newSpec(t, 4, 8, 10), testNow.Add(3*time.Minute))
	assert.ErrorIs(t, err, ErrDiskShrink)
	assert.Equal(t, s.Spec, next)
}

func TestInstanceQueries(t *testing.T) {
	creating := spawnInstance(t)
	assert.False(t, creating.IsRunning())
	assert.False(t, creating.IsStopped())
	assert.True(t, creating.IsActive())

	running := runInstance(t, creating, testNow)
	assert.True(t, running.IsRunning())
	assert.True(t, running.IsActive())

	stopped := stopInstance(t, running, testNow.Add(time.Minute))
	assert.True(t, stopped.IsStopped())
	assert.True(t, stopped.IsActive())

	deleted := deletedInstance(t)
	assert.True(t, deleted.IsTerminal())
	assert.False(t, deleted.IsActive())
}
