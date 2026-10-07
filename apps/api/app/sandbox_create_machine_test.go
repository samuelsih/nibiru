package app

import (
	"strings"
	"testing"
	"uuid"

	"github.com/jackc/pgx/v5"
	"github.com/samuelsih/golib/assert"
)

func findSandbox(t *testing.T, id uuid.UUID) Sandbox {
	t.Helper()

	rows, err := db.Query(t.Context(),
		`SELECT id, owner_id, name, cpu, memory_gb, disk_gb, state, created_at, updated_at
		 FROM sandboxes WHERE id = $1`, id)
	assert.NoError(t, err)

	sandbox, err := pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[Sandbox])
	assert.NoError(t, err)

	return sandbox
}

func TestCreateMachine(t *testing.T) {
	owner := registerUser(t, newEmail())

	sandbox, err := CreateSandbox(t.Context(), db, CreateSandboxRequest{
		OwnerID:  owner.ID,
		Name:     "worker-1",
		CPU:      2,
		MemoryGB: 4,
		DiskGB:   12,
	})
	assert.NoError(t, err)

	assert.NotEqual(t, sandbox.ID, uuid.Nil())
	assert.Equal(t, sandbox.Name, "worker-1")
	assert.Equal(t, sandbox.CPU, 2)
	assert.Equal(t, sandbox.MemoryGB, 4)
	assert.Equal(t, sandbox.DiskGB, 12)
	assert.Equal(t, sandbox.Image, DefaultImage)
	assert.Equal(t, sandbox.State, StateCreating)

	record := findSandbox(t, sandbox.ID)
	assert.Equal(t, record.OwnerID, owner.ID)
	assert.Equal(t, record.Name, "worker-1")
	assert.Equal(t, record.CPU, 2)
	assert.Equal(t, record.MemoryGB, 4)
	assert.Equal(t, record.DiskGB, 12)
	assert.Equal(t, record.State, StateCreating)
}

func TestCreateSandboxRequestValidate(t *testing.T) {
	tests := []struct {
		name    string
		req     CreateSandboxRequest
		wantErr bool
	}{
		{
			name: "valid",
			req:  CreateSandboxRequest{CPU: 2, MemoryGB: 4, DiskGB: 12},
		},
		{
			name:    "missing cpu",
			req:     CreateSandboxRequest{MemoryGB: 4, DiskGB: 12},
			wantErr: true,
		},
		{
			name:    "cpu too high",
			req:     CreateSandboxRequest{CPU: 33, MemoryGB: 4, DiskGB: 12},
			wantErr: true,
		},
		{
			name:    "missing memory",
			req:     CreateSandboxRequest{CPU: 2, DiskGB: 12},
			wantErr: true,
		},
		{
			name:    "memory too high",
			req:     CreateSandboxRequest{CPU: 2, MemoryGB: 129, DiskGB: 12},
			wantErr: true,
		},
		{
			name:    "missing disk",
			req:     CreateSandboxRequest{CPU: 2, MemoryGB: 4},
			wantErr: true,
		},
		{
			name:    "disk too high",
			req:     CreateSandboxRequest{CPU: 2, MemoryGB: 4, DiskGB: 501},
			wantErr: true,
		},
		{
			name:    "name too long",
			req:     CreateSandboxRequest{Name: strings.Repeat("a", 256), CPU: 2, MemoryGB: 4, DiskGB: 12},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()

			if tt.wantErr {
				assert.NotNil(t, err)

				return
			}

			assert.NoError(t, err)
		})
	}
}
