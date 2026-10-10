package app

import (
	"context"
	"uuid"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func ListSandboxesByOwner(ctx context.Context, db *pgxpool.Pool, ownerID, cursor uuid.UUID, limit int) ([]SandboxSummary, error) {
	q := builder.
		Select("id", "name", "cpu", "memory_gb", "disk_gb", "state").
		From("sandboxes").
		Where(sq.Eq{"owner_id": ownerID.String()}).
		Where(sq.NotEq{"state": StateDeleted}).
		OrderBy("id DESC").
		Limit(uint64(limit))

	if cursor != uuid.Nil() {
		q = q.Where(sq.Lt{"id": cursor.String()})
	}

	query, args, err := q.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	return pgx.CollectRows(rows, pgx.RowToStructByName[SandboxSummary])
}

func SaveSandbox(ctx context.Context, tx pgx.Tx, sandbox Sandbox) error {
	query, args, err := builder.
		Insert("sandboxes").
		Columns("id", "owner_id", "name", "cpu", "memory_gb", "disk_gb", "state", "created_at", "updated_at").
		Values(sandbox.ID, sandbox.OwnerID, sandbox.Name, sandbox.CPU, sandbox.MemoryGB, sandbox.DiskGB, sandbox.State, sandbox.CreatedAt, sandbox.UpdatedAt).
		ToSql()
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, query, args...)

	return err
}
