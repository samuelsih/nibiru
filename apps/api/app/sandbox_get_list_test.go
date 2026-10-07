package app

import (
	"testing"
	"uuid"

	"github.com/samuelsih/golib/assert"
)

func createTestSandbox(t *testing.T, ownerID uuid.UUID, name string) Sandbox {
	t.Helper()

	sandbox, err := CreateSandbox(t.Context(), db, CreateSandboxRequest{
		OwnerID:  ownerID,
		Name:     name,
		CPU:      2,
		MemoryGB: 4,
		DiskGB:   12,
	})
	assert.NoError(t, err)

	return sandbox
}

func TestListSandboxes(t *testing.T) {
	owner := registerUser(t, newEmail())
	other := registerUser(t, newEmail())

	first := createTestSandbox(t, owner.ID, "first")
	second := createTestSandbox(t, owner.ID, "second")
	third := createTestSandbox(t, owner.ID, "third")
	createTestSandbox(t, other.ID, "foreign")

	result, err := ListSandboxes(t.Context(), db, owner.ID, uuid.Nil(), 10)
	assert.NoError(t, err)
	assert.Equal(t, len(result.Items), 3)
	assert.Equal(t, result.Items[0].ID, third.ID)
	assert.Equal(t, result.Items[1].ID, second.ID)
	assert.Equal(t, result.Items[2].ID, first.ID)
	assert.False(t, result.NextCursor.Valid)

	_, err = db.Exec(t.Context(), `UPDATE sandboxes SET state = $1 WHERE id = $2`, StateDeleted, first.ID)
	assert.NoError(t, err)

	result, err = ListSandboxes(t.Context(), db, owner.ID, uuid.Nil(), 10)
	assert.NoError(t, err)
	assert.Equal(t, len(result.Items), 2)
}

func TestListSandboxesPagination(t *testing.T) {
	owner := registerUser(t, newEmail())

	first := createTestSandbox(t, owner.ID, "first")
	second := createTestSandbox(t, owner.ID, "second")
	third := createTestSandbox(t, owner.ID, "third")

	page, err := ListSandboxes(t.Context(), db, owner.ID, uuid.Nil(), 2)
	assert.NoError(t, err)
	assert.Equal(t, len(page.Items), 2)
	assert.Equal(t, page.Items[0].ID, third.ID)
	assert.Equal(t, page.Items[1].ID, second.ID)
	assert.True(t, page.NextCursor.Valid)
	assert.Equal(t, page.NextCursor.V, second.ID)

	next, err := ListSandboxes(t.Context(), db, owner.ID, page.NextCursor.V, 2)
	assert.NoError(t, err)
	assert.Equal(t, len(next.Items), 1)
	assert.Equal(t, next.Items[0].ID, first.ID)
	assert.False(t, next.NextCursor.Valid)
}
