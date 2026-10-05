package sandbox

import (
	"context"
	"fmt"
	"time"
	"uuid"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samuelsih/nibiru/api/pkg/outbox"
	"github.com/samuelsih/nibiru/api/pkg/trx"
)

const OutboxTypeCreated = "sandbox.created"

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

type CreateRequest struct {
	OwnerID  uuid.UUID
	Name     string
	CPU      int
	MemoryGB int
	DiskGB   int
}

type CreatedPayload struct {
	ID              uuid.UUID     `json:"id"`
	OwnerID         uuid.UUID     `json:"ownerId"`
	Name            string        `json:"name"`
	CPU             int           `json:"cpu"`
	MemoryGB        int           `json:"memoryGb"`
	DiskGB          int           `json:"diskGb"`
	Image           string        `json:"image"`
	AutoStopAfter   time.Duration `json:"autoStopAfter"`
	AutoDeleteAfter time.Duration `json:"autoDeleteAfter"`
}

func (h Handler) Create(ctx context.Context, req CreateRequest) (Instance, error) {
	now := time.Now()

	instance, err := Spawn(SpawnRequest{
		OwnerID: req.OwnerID,
		Name:    req.Name,
		Spec:    Spec{CPU: req.CPU, MemoryGB: req.MemoryGB, DiskGB: req.DiskGB},
	}, now)
	if err != nil {
		return Instance{}, err
	}

	err = trx.WithTx(ctx, h.db, func(tx pgx.Tx) error {
		if err := h.repo.SaveInstance(ctx, tx, instance); err != nil {
			return err
		}

		return outbox.Insert(ctx, tx, outbox.InsertParam[map[string]any, CreatedPayload]{
			Type: OutboxTypeCreated,
			Payload: CreatedPayload{
				ID:              instance.ID,
				OwnerID:         instance.OwnerID,
				Name:            instance.Name,
				CPU:             instance.Spec.CPU,
				MemoryGB:        instance.Spec.MemoryGB,
				DiskGB:          instance.Spec.DiskGB,
				Image:           instance.Image,
				AutoStopAfter:   instance.Lifetime.AutoStopAfter,
				AutoDeleteAfter: instance.Lifetime.AutoDeleteAfter,
			},
		})
	})
	if err != nil {
		return Instance{}, fmt.Errorf("cannot create sandbox: %w", err)
	}

	return instance, nil
}
