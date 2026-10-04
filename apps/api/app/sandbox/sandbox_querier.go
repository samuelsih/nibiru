package sandbox

import (
	"context"
	"fmt"
	"uuid"

	core "github.com/samuelsih/nibiru/api/app/sandbox/internal"
)

const (
	DefaultListLimit = 20
	MaxListLimit     = 100
)

type InstanceSummary struct {
	ID    uuid.UUID  `db:"id"        json:"id"    example:"01924a7d-9b3e-7c1a-8f4b-7c1d2e3f4a5b"`
	Name  string     `db:"name"      json:"name"  example:"worker-1"`
	CPU   int        `db:"cpu"       json:"cpu"   example:"4"`
	RAM   int        `db:"memory_gb" json:"ram"   example:"8"`
	Disk  int        `db:"disk_gb"   json:"disk"  example:"50"`
	State core.State `db:"state"     json:"state" example:"running" enum:"creating,running,stopping,stopped,starting,deleting,failed"`
}

type ListResult struct {
	Items      []InstanceSummary
	NextCursor uuid.UUID
}

func (h Handler) List(ctx context.Context, ownerID, cursor uuid.UUID, limit int) (ListResult, error) {
	if ownerID == uuid.Nil() {
		return ListResult{}, core.ErrOwnerRequired
	}

	switch {
	case limit <= 0:
		limit = DefaultListLimit
	case limit > MaxListLimit:
		limit = MaxListLimit
	}

	items, err := h.repo.ListByOwner(ctx, ownerID, cursor, limit+1)
	if err != nil {
		return ListResult{}, fmt.Errorf("cannot list sandboxes: %w", err)
	}

	var next uuid.UUID
	if len(items) > limit {
		items = items[:limit]
		next = items[len(items)-1].ID
	}

	return ListResult{Items: items, NextCursor: next}, nil
}
