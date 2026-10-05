package sandbox

import (
	"testing"
	"time"

	"github.com/samuelsih/golib/assert"
)

func TestNewMeter(t *testing.T) {
	meter := NewMeter()

	assert.False(t, meter.Running())
	assert.Equal(t, meter.Total(testNow), time.Duration(0))
}

func TestMeterStartPauseTotal(t *testing.T) {
	meter := NewMeter().Start(testNow)
	assert.True(t, meter.Running())

	assert.Equal(t, meter.Total(testNow.Add(90*time.Second)), 90*time.Second)
	assert.Equal(t, meter.Total(testNow.Add(-time.Minute)), time.Duration(0))

	startedAgain := meter.Start(testNow.Add(time.Minute))
	assert.Equal(t, startedAgain.Total(testNow.Add(90*time.Second)), 90*time.Second)

	paused := meter.Pause(testNow.Add(2 * time.Minute))
	assert.False(t, paused.Running())
	assert.Equal(t, paused.Total(testNow.Add(time.Hour)), 2*time.Minute)

	pausedAgain := paused.Pause(testNow.Add(time.Hour))
	assert.Equal(t, pausedAgain.Total(testNow.Add(time.Hour)), 2*time.Minute)

	resumed := paused.Start(testNow.Add(time.Hour))
	assert.True(t, resumed.Running())
	assert.Equal(t, resumed.Total(testNow.Add(time.Hour+30*time.Second)), 2*time.Minute+30*time.Second)
}
