package outbox

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"time"
	"uuid"

	sq "github.com/Masterminds/squirrel"
	"github.com/guregu/null/v6"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/samuelsih/golib/concx"
	"github.com/samuelsih/golib/slicex"
	"github.com/samuelsih/golib/slogx"
	"github.com/samuelsih/nibiru/api/pkg/trx"
)

const (
	Batch         = 10
	TypeUnknown   = "unknown"
	Stream        = "OUTBOX"
	SubjectPrefix = "outbox"
	PollInterval  = 3 * time.Second
)

type Message struct {
	ID          uuid.UUID   `db:"id"`
	Type        string      `db:"type"`
	Metadata    []byte      `db:"metadata"`
	Payload     []byte      `db:"payload"`
	Error       null.String `db:"error"`
	DeliveredAt null.Time   `db:"delivered_at"`
	CreatedAt   time.Time   `db:"created_at"`
}

func (m Message) Subject() string {
	return SubjectPrefix + "." + m.Type
}

func (m *Message) MarkAsDelivered() {
	m.DeliveredAt = null.TimeFrom(time.Now())
	m.Error = null.StringFromPtr(nil)
}

func (m *Message) MarkAsError(err error) {
	m.DeliveredAt = null.TimeFromPtr(nil)
	m.Error = null.StringFrom(err.Error())
}

type InsertParam[T, U any] struct {
	Type     string
	Metadata null.Value[T]
	Payload  U
}

func Insert[T, U any](ctx context.Context, tx pgx.Tx, p InsertParam[T, U]) error {
	if p.Type == "" {
		p.Type = TypeUnknown
	}

	var metadata T

	if p.Metadata.Valid {
		metadata = p.Metadata.V
	}

	query, args, err := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
		Insert("outbox_messages").
		Columns("id", "type", "metadata", "payload", "created_at").
		Values(uuid.New(), p.Type, metadata, p.Payload, time.Now()).
		ToSql()
	if err != nil {
		return fmt.Errorf("cannot build outbox insert query: %w", err)
	}

	if _, err = tx.Exec(ctx, query, args...); err != nil {
		return fmt.Errorf("cannot insert outbox message: %w", err)
	}

	return nil
}

func GetMessages(ctx context.Context, tx pgx.Tx) ([]Message, error) {
	query, args, err := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
		Select("id", "type", "metadata", "payload", "error", "delivered_at", "created_at").
		From("outbox_messages").
		Where(sq.Eq{"delivered_at": nil}).
		OrderBy("created_at").
		Limit(Batch).
		Suffix("FOR UPDATE SKIP LOCKED").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("cannot build outbox select query: %w", err)
	}

	rows, err := tx.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("cannot query outbox messages: %w", err)
	}

	messages, err := pgx.CollectRows(rows, pgx.RowToStructByName[Message])
	if err != nil {
		return nil, fmt.Errorf("cannot collect outbox messages: %w", err)
	}

	return messages, nil
}

func UpdateMessages(ctx context.Context, tx pgx.Tx, messages []Message) error {
	builder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	deliveredMsgs := slicex.Filter(messages, func(message Message) bool { return message.Error.IsZero() })
	failedMsgs := slicex.Filter(messages, func(message Message) bool { return message.Error.Valid })

	deliveredIDs := slicex.Transform(deliveredMsgs, func(message Message) string { return message.ID.String() })
	failedIDs := slicex.Transform(failedMsgs, func(message Message) string { return message.ID.String() })

	if len(deliveredMsgs) > 0 {
		query, args, err := builder.
			Update("outbox_messages").
			Set("delivered_at", time.Now()).
			Set("error", nil).
			Where(sq.Eq{"id": deliveredIDs}).
			ToSql()
		if err != nil {
			return fmt.Errorf("cannot build outbox delivered query: %w", err)
		}

		if _, err = tx.Exec(ctx, query, args...); err != nil {
			return fmt.Errorf("cannot update delivered outbox messages: %w", err)
		}
	}

	if len(failedMsgs) > 0 {
		cases := sq.Case("id")
		for message := range slices.Values(failedMsgs) {
			cases = cases.When(sq.Expr("?", message.ID.String()), sq.Expr("?", message.Error.String))
		}

		query, args, err := builder.
			Update("outbox_messages").
			Set("error", cases).
			Where(sq.Eq{"id": failedIDs}).
			ToSql()
		if err != nil {
			return fmt.Errorf("cannot build outbox failed query: %w", err)
		}

		if _, err = tx.Exec(ctx, query, args...); err != nil {
			return fmt.Errorf("cannot update failed outbox messages: %w", err)
		}
	}

	return nil
}

func EnsureStream(ctx context.Context, js jetstream.JetStream) error {
	_, err := js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:     Stream,
		Subjects: []string{SubjectPrefix + ".>"},
	})
	if err != nil {
		return fmt.Errorf("cannot create or update outbox stream: %w", err)
	}

	return nil
}

func Poll(ctx context.Context, db *pgxpool.Pool, js jetstream.JetStream) {
	ticker := time.NewTicker(PollInterval)
	defer ticker.Stop()

	slog.Info("Starting to polling outbox")

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			err := PublishBatch(ctx, db, js)
			if err != nil && !errors.Is(err, context.Canceled) {
				slog.Error("cannot poll outbox messages", slogx.ErrorAttr(err))
			}
		}
	}
}

func PublishBatch(ctx context.Context, db *pgxpool.Pool, js jetstream.JetStream) error {
	return trx.WithTx(ctx, db, func(tx pgx.Tx) error {
		messages, err := GetMessages(ctx, tx)
		if err != nil {
			return fmt.Errorf("cannot get outbox messages: %w", err)
		}

		results := concx.ForEachRes(messages, func(message Message) Message {
			_, publishErr := js.Publish(ctx, message.Subject(), message.Payload, jetstream.WithMsgID(message.ID.String()))
			if publishErr != nil {
				message.MarkAsError(fmt.Errorf("cannot publish outbox message %s: %w", message.ID, publishErr))
			} else {
				message.MarkAsDelivered()
			}

			return message
		})

		if err = UpdateMessages(ctx, tx, results); err != nil {
			return fmt.Errorf("cannot update outbox messages: %w", err)
		}

		return nil
	})
}
