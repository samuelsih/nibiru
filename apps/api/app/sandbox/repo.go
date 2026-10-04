package sandbox

import (
	"context"
	"fmt"
	"uuid"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	core "github.com/samuelsih/nibiru/api/app/sandbox/internal"
)

type Repo struct {
	db      *pgxpool.Pool
	builder sq.StatementBuilderType
}

func NewRepo(db *pgxpool.Pool) Repo {
	return Repo{
		db:      db,
		builder: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (r Repo) ListByOwner(ctx context.Context, ownerID, cursor uuid.UUID, limit int) ([]InstanceSummary, error) {
	builder := r.builder.
		Select("id", "name", "cpu", "memory_gb", "disk_gb", "state").
		From("sandboxes").
		Where(sq.Eq{"owner_id": ownerID}).
		Where(sq.NotEq{"state": core.StateDeleted}).
		OrderBy("id DESC").
		Limit(uint64(limit))

	if cursor != uuid.Nil() {
		builder = builder.Where(sq.Lt{"id": cursor})
	}

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("cannot build sandbox list query: %w", err)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("cannot query sandboxes: %w", err)
	}

	records, err := pgx.CollectRows(rows, pgx.RowToStructByName[InstanceSummary])
	if err != nil {
		return nil, fmt.Errorf("cannot collect sandboxes: %w", err)
	}

	return records, nil
}
