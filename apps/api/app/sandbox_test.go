package app

import (
	"strings"
	"testing"
	"uuid"

	"github.com/jackc/pgx/v5"
	"github.com/samuelsih/golib/assert"
	"github.com/samuelsih/nibiru/api/app/auth"
	"github.com/samuelsih/nibiru/api/app/sandbox"
)

type sandboxRecord struct {
	ID       uuid.UUID `db:"id"`
	OwnerID  uuid.UUID `db:"owner_id"`
	Name     string    `db:"name"`
	CPU      int       `db:"cpu"`
	MemoryGB int       `db:"memory_gb"`
	DiskGB   int       `db:"disk_gb"`
	State    string    `db:"state"`
}

func registerOwner(t *testing.T) auth.User {
	t.Helper()

	email := uuid.New().String() + "@example.com"

	err := auth.NewHandler(db, sessionTTL).Register(t.Context(), auth.RegisterRequest{
		Email:     email,
		Password:  testPassword,
		FirstName: "Sandi",
	})
	assert.NoError(t, err)

	return findUser(t, email)
}

func findSandbox(t *testing.T, id uuid.UUID) sandboxRecord {
	t.Helper()

	rows, err := db.Query(t.Context(),
		`SELECT id, owner_id, name, cpu, memory_gb, disk_gb, state
		 FROM sandboxes WHERE id = $1`, id)
	assert.NoError(t, err)

	record, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[sandboxRecord])
	assert.NoError(t, err)

	return record
}

func findSandboxCreatedPayload(t *testing.T, id uuid.UUID) []byte {
	t.Helper()

	var payload []byte

	err := db.QueryRow(t.Context(),
		`SELECT payload FROM outbox_messages
		 WHERE type = $1 AND payload ->> 'id' = $2`,
		sandbox.OutboxTypeCreated, id.String()).Scan(&payload)
	assert.NoError(t, err)

	return payload
}

func countSandboxes(t *testing.T, ownerID uuid.UUID) int {
	t.Helper()

	var count int
	err := db.QueryRow(t.Context(),
		`SELECT count(*) FROM sandboxes WHERE owner_id = $1`, ownerID).Scan(&count)
	assert.NoError(t, err)

	return count
}

func TestSandboxCreate(t *testing.T) {
	owner := registerOwner(t)
	handler := sandbox.NewHandler(db)

	instance, err := handler.Create(t.Context(), sandbox.CreateRequest{
		OwnerID:  uuid.UUID(owner.ID),
		Name:     "worker-1",
		CPU:      2,
		MemoryGB: 4,
		DiskGB:   12,
	})
	assert.NoError(t, err)

	assert.NotEqual(t, instance.ID, uuid.Nil())
	assert.Equal(t, instance.Name, "worker-1")
	assert.Equal(t, instance.Spec.CPU, 2)
	assert.Equal(t, instance.Spec.MemoryGB, 4)
	assert.Equal(t, instance.Spec.DiskGB, 12)
	assert.Equal(t, instance.State, sandbox.StateCreating)

	record := findSandbox(t, instance.ID)
	assert.Equal(t, record.OwnerID, uuid.UUID(owner.ID))
	assert.Equal(t, record.Name, "worker-1")
	assert.Equal(t, record.State, "creating")

	payload := findSandboxCreatedPayload(t, instance.ID)
	assert.Equal(t, assert.JSONPath(payload, "$.ownerId"), owner.ID.String())
	assert.Equal(t, assert.JSONPath(payload, "$.name"), "worker-1")
	assert.Equal(t, assert.JSONPath(payload, "$.cpu"), 2)
	assert.Equal(t, assert.JSONPath(payload, "$.memoryGb"), 4)
	assert.Equal(t, assert.JSONPath(payload, "$.diskGb"), 12)
	assert.Equal(t, assert.JSONPath(payload, "$.image"), sandbox.DefaultImage)
}

func TestSandboxCreateGeneratesName(t *testing.T) {
	owner := registerOwner(t)
	handler := sandbox.NewHandler(db)

	instance, err := handler.Create(t.Context(), sandbox.CreateRequest{
		OwnerID:  uuid.UUID(owner.ID),
		CPU:      2,
		MemoryGB: 4,
		DiskGB:   12,
	})
	assert.NoError(t, err)

	assert.True(t, strings.HasPrefix(instance.Name, "sandbox-"))
}

func TestSandboxCreateRejectsInvalidSpec(t *testing.T) {
	owner := registerOwner(t)
	handler := sandbox.NewHandler(db)

	_, err := handler.Create(t.Context(), sandbox.CreateRequest{
		OwnerID:  uuid.UUID(owner.ID),
		CPU:      0,
		MemoryGB: 4,
		DiskGB:   12,
	})
	assert.ErrorIs(t, err, sandbox.ErrInvalidSpec)
	assert.Equal(t, countSandboxes(t, uuid.UUID(owner.ID)), 0)
}
