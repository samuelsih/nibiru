package sandbox

import (
	"context"
	"fmt"
	"slices"
	"uuid"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
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

type instanceSummaryRecord struct {
	ID       uuid.UUID `db:"id"`
	Name     string    `db:"name"`
	CPU      int       `db:"cpu"`
	MemoryGB int       `db:"memory_gb"`
	DiskGB   int       `db:"disk_gb"`
	State    State     `db:"state"`
}

func (r Repo) ListByOwner(ctx context.Context, ownerID, cursor uuid.UUID, limit int) ([]InstanceSummary, error) {
	builder := r.builder.
		Select("id", "name", "cpu", "memory_gb", "disk_gb", "state").
		From("sandboxes").
		Where(sq.Eq{"owner_id": ownerID}).
		Where(sq.NotEq{"state": StateDeleted}).
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

	records, err := pgx.CollectRows(rows, pgx.RowToStructByName[instanceSummaryRecord])
	if err != nil {
		return nil, fmt.Errorf("cannot collect sandboxes: %w", err)
	}

	summaries := make([]InstanceSummary, 0, len(records))
	for record := range slices.Values(records) {
		summary, err := record.summary()
		if err != nil {
			return nil, err
		}

		summaries = append(summaries, summary)
	}

	return summaries, nil
}

func (r instanceSummaryRecord) summary() (InstanceSummary, error) {
	spec, err := NewSpec(r.CPU, r.MemoryGB, r.DiskGB)
	if err != nil {
		return InstanceSummary{}, fmt.Errorf("cannot read sandbox %s: %w", r.ID, err)
	}

	return InstanceSummary{
		ID:    r.ID,
		Name:  r.Name,
		Spec:  spec,
		State: r.State,
	}, nil
}
