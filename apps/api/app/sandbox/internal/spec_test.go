package core

import (
	"testing"

	"github.com/samuelsih/golib/assert"
)

func TestNewSpec(t *testing.T) {
	tests := []struct {
		name     string
		cpu      int
		memoryGB int
		diskGB   int
		wantErr  error
	}{
		{name: "small", cpu: 2, memoryGB: 4, diskGB: 12},
		{name: "default", cpu: 4, memoryGB: 8, diskGB: 50},
		{name: "large", cpu: 8, memoryGB: 16, diskGB: 125},
		{name: "minimum bounds", cpu: MinCPU, memoryGB: MinMemoryGB, diskGB: MinDiskGB},
		{name: "maximum bounds", cpu: MaxCPU, memoryGB: MaxMemoryGB, diskGB: MaxDiskGB},
		{name: "zero cpu", cpu: 0, memoryGB: 8, diskGB: 50, wantErr: ErrInvalidSpec},
		{name: "cpu over limit", cpu: MaxCPU + 1, memoryGB: 8, diskGB: 50, wantErr: ErrInvalidSpec},
		{name: "zero memory", cpu: 4, memoryGB: 0, diskGB: 50, wantErr: ErrInvalidSpec},
		{name: "memory over limit", cpu: 4, memoryGB: MaxMemoryGB + 1, diskGB: 50, wantErr: ErrInvalidSpec},
		{name: "zero disk", cpu: 4, memoryGB: 8, diskGB: 0, wantErr: ErrInvalidSpec},
		{name: "negative disk", cpu: 4, memoryGB: 8, diskGB: -1, wantErr: ErrInvalidSpec},
		{name: "disk over limit", cpu: 4, memoryGB: 8, diskGB: MaxDiskGB + 1, wantErr: ErrInvalidSpec},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec, err := NewSpec(tt.cpu, tt.memoryGB, tt.diskGB)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Equal(t, spec, Spec{})
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, spec.CPU, tt.cpu)
			assert.Equal(t, spec.MemoryGB, tt.memoryGB)
			assert.Equal(t, spec.DiskGB, tt.diskGB)
		})
	}
}

func TestSpecPresets(t *testing.T) {
	assert.Equal(t, SmallSpec.CPU, 2)
	assert.Equal(t, SmallSpec.MemoryGB, 4)
	assert.Equal(t, SmallSpec.DiskGB, 12)
	assert.Equal(t, DefaultSpec.CPU, 4)
	assert.Equal(t, DefaultSpec.MemoryGB, 8)
	assert.Equal(t, DefaultSpec.DiskGB, 50)
	assert.Equal(t, LargeSpec.CPU, 8)
	assert.Equal(t, LargeSpec.MemoryGB, 16)
	assert.Equal(t, LargeSpec.DiskGB, 125)
}

func TestSpecIsZero(t *testing.T) {
	assert.True(t, Spec{}.IsZero())
	assert.False(t, SmallSpec.IsZero())
}

func TestSpecResize(t *testing.T) {
	next, err := NewSpec(8, 16, 100)
	assert.NoError(t, err)

	grown, err := DefaultSpec.Resize(next)
	assert.NoError(t, err)
	assert.Equal(t, grown, next)

	same, err := DefaultSpec.Resize(DefaultSpec)
	assert.NoError(t, err)
	assert.Equal(t, same, DefaultSpec)

	smaller, err := NewSpec(8, 16, 10)
	assert.NoError(t, err)

	_, err = DefaultSpec.Resize(smaller)
	assert.ErrorIs(t, err, ErrDiskShrink)

	_, err = DefaultSpec.Resize(Spec{})
	assert.ErrorIs(t, err, ErrInvalidSpec)
}
