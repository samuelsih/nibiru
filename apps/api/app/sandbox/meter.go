package sandbox

import "time"

type Meter struct {
	accumulated time.Duration
	runningFrom time.Time
}

func NewMeter() Meter {
	return Meter{}
}

func (m Meter) Running() bool {
	return !m.runningFrom.IsZero()
}

func (m Meter) Start(now time.Time) Meter {
	if m.Running() {
		return m
	}

	m.runningFrom = now

	return m
}

func (m Meter) Pause(now time.Time) Meter {
	if !m.Running() {
		return m
	}

	if now.After(m.runningFrom) {
		m.accumulated += now.Sub(m.runningFrom)
	}

	m.runningFrom = time.Time{}

	return m
}

func (m Meter) Total(now time.Time) time.Duration {
	if !m.Running() || !now.After(m.runningFrom) {
		return m.accumulated
	}

	return m.accumulated + now.Sub(m.runningFrom)
}
