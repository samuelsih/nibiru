package app

import (
	"context"
	"uuid"

	"github.com/guregu/null/v6"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	DefaultListLimit = 20
	MaxListLimit     = 100
)

type SandboxSummary struct {
	ID    uuid.UUID    `db:"id"        json:"id"    example:"01924a7d-9b3e-7c1a-8f4b-7c1d2e3f4a5b"`
	Name  string       `db:"name"      json:"name"  example:"worker-1"`
	CPU   int          `db:"cpu"       json:"cpu"   example:"4"`
	RAM   int          `db:"memory_gb" json:"ram"   example:"8"`
	Disk  int          `db:"disk_gb"   json:"disk"  example:"50"`
	State SandboxState `db:"state"     json:"state" example:"running" enum:"creating,running,stopping,stopped,starting,deleting,failed"`
}

type ListResult struct {
	Items      []SandboxSummary      `json:"items"`
	NextCursor null.Value[uuid.UUID] `json:"nextCursor"`
}

func ListSandboxes(ctx context.Context, db *pgxpool.Pool, ownerID, cursor uuid.UUID, limit int) (ListResult, error) {
	switch {
	case limit <= 0:
		limit = DefaultListLimit
	case limit > MaxListLimit:
		limit = MaxListLimit
	}

	items, err := ListSandboxesByOwner(ctx, db, ownerID, cursor, limit+1)
	if err != nil {
		return ListResult{}, err
	}

	var next uuid.UUID
	if len(items) > limit {
		items = items[:limit]
		next = items[len(items)-1].ID
	}

	return ListResult{
		Items:      items,
		NextCursor: null.NewValue(next, next != uuid.Nil()),
	}, nil
}
