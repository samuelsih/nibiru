package core

import "fmt"

type Spec struct {
	CPU      int
	MemoryGB int
	DiskGB   int
}

const (
	MinCPU      = 1
	MaxCPU      = 32
	MinMemoryGB = 1
	MaxMemoryGB = 128
	MinDiskGB   = 1
	MaxDiskGB   = 500
)

var (
	SmallSpec   = Spec{CPU: 2, MemoryGB: 4, DiskGB: 12}
	DefaultSpec = Spec{CPU: 4, MemoryGB: 8, DiskGB: 50}
	LargeSpec   = Spec{CPU: 8, MemoryGB: 16, DiskGB: 125}
)

func NewSpec(cpu, memoryGB, diskGB int) (Spec, error) {
	switch {
	case cpu < MinCPU || cpu > MaxCPU:
		return Spec{}, fmt.Errorf("%w: cpu must be between %d and %d", ErrInvalidSpec, MinCPU, MaxCPU)
	case memoryGB < MinMemoryGB || memoryGB > MaxMemoryGB:
		return Spec{}, fmt.Errorf("%w: memory must be between %dGB and %dGB", ErrInvalidSpec, MinMemoryGB, MaxMemoryGB)
	case diskGB < MinDiskGB || diskGB > MaxDiskGB:
		return Spec{}, fmt.Errorf("%w: disk must be between %dGB and %dGB", ErrInvalidSpec, MinDiskGB, MaxDiskGB)
	}

	return Spec{CPU: cpu, MemoryGB: memoryGB, DiskGB: diskGB}, nil
}

func (s Spec) IsZero() bool {
	return s == Spec{}
}

func (s Spec) Resize(next Spec) (Spec, error) {
	validated, err := NewSpec(next.CPU, next.MemoryGB, next.DiskGB)
	if err != nil {
		return Spec{}, err
	}

	if validated.DiskGB < s.DiskGB {
		return Spec{}, fmt.Errorf("%w: %dGB to %dGB", ErrDiskShrink, s.DiskGB, validated.DiskGB)
	}

	return validated, nil
}
