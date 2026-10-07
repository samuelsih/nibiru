package app

import (
	"context"
	"time"
	"uuid"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	DefaultImage      = "ubuntu:24.04"
	OutboxTypeCreated = "sandbox.created"
)

type CreateSandboxRequest struct {
	OwnerID  uuid.UUID `json:"-"`
	Name     string    `json:"name" example:"worker-1" required:"false"`
	CPU      int       `json:"cpu"  example:"4"`
	MemoryGB int       `json:"ram"  example:"8"`
	DiskGB   int       `json:"disk" example:"50"`
}

func (r CreateSandboxRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Name, validation.Length(0, 255)),
		validation.Field(&r.CPU, validation.Required, validation.Min(1), validation.Max(32)),
		validation.Field(&r.MemoryGB, validation.Required, validation.Min(1), validation.Max(128)),
		validation.Field(&r.DiskGB, validation.Required, validation.Min(1), validation.Max(500)),
	)
}

type SandboxCreatedPayload struct {
	ID       uuid.UUID `json:"id"`
	OwnerID  uuid.UUID `json:"ownerId"`
	Name     string    `json:"name"`
	CPU      int       `json:"cpu"`
	MemoryGB int       `json:"memoryGb"`
	DiskGB   int       `json:"diskGb"`
	Image    string    `json:"image"`
}

func CreateSandbox(ctx context.Context, db *pgxpool.Pool, r CreateSandboxRequest) (Sandbox, error) {
	now := time.Now()

	sandbox := Sandbox{
		ID:        uuid.NewV7(),
		OwnerID:   r.OwnerID,
		Name:      r.Name,
		CPU:       r.CPU,
		MemoryGB:  r.MemoryGB,
		DiskGB:    r.DiskGB,
		Image:     DefaultImage,
		State:     StateCreating,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := WithTx(ctx, db, func(tx pgx.Tx) error {
		if err := SaveSandbox(ctx, tx, sandbox); err != nil {
			return err
		}

		return OutboxInsert(ctx, tx, OutboxInsertParam[map[string]any, SandboxCreatedPayload]{
			Type: OutboxTypeCreated,
			Payload: SandboxCreatedPayload{
				ID:       sandbox.ID,
				OwnerID:  sandbox.OwnerID,
				Name:     sandbox.Name,
				CPU:      sandbox.CPU,
				MemoryGB: sandbox.MemoryGB,
				DiskGB:   sandbox.DiskGB,
				Image:    sandbox.Image,
			},
		})
	})
	if err != nil {
		return Sandbox{}, err
	}

	return sandbox, nil
}
