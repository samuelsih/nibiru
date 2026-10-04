package sandbox

import (
	"context"
	"fmt"
	"uuid"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	DefaultListLimit = 20
	MaxListLimit     = 100
)

type Handler struct {
	db   *pgxpool.Pool
	repo Repo
}

func NewHandler(db *pgxpool.Pool) Handler {
	return Handler{
		db:   db,
		repo: NewRepo(db),
	}
}

type InstanceSummary struct {
	ID    uuid.UUID
	Name  string
	Spec  Spec
	State State
}

type ListResult struct {
	Items      []InstanceSummary
	NextCursor uuid.UUID
}

func (h Handler) List(ctx context.Context, ownerID, cursor uuid.UUID, limit int) (ListResult, error) {
	if ownerID == uuid.Nil() {
		return ListResult{}, ErrOwnerRequired
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
