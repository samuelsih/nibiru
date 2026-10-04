package app

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"
	"time"
	"uuid"

	"github.com/guregu/null/v6"
	"github.com/jackc/pgx/v5"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/samuelsih/golib/assert"
	"github.com/samuelsih/nibiru/api/pkg/trx"
)

func insertOutboxMessage[T, U any](t *testing.T, p InsertOutboxParam[T, U]) {
	t.Helper()

	err := trx.WithTx(t.Context(), db, func(tx pgx.Tx) error {
		return InsertOutbox(t.Context(), tx, p)
	})
	assert.NoError(t, err)

	t.Cleanup(func() {
		_, err := db.Exec(context.Background(),
			`DELETE FROM outbox_messages WHERE payload = $1`, p.Payload)
		assert.NoError(t, err)
	})
}

func getOutboxMessages(t *testing.T) []OutboxMessage {
	t.Helper()

	var messages []OutboxMessage

	err := trx.WithTx(t.Context(), db, func(tx pgx.Tx) error {
		var err error

		messages, err = GetOutboxMessages(t.Context(), tx)

		return err
	})
	assert.NoError(t, err)

	return messages
}

func findOutboxMessage(t *testing.T, payload any) OutboxMessage {
	t.Helper()

	rows, err := db.Query(t.Context(), `
		SELECT id, type, metadata, payload, error, delivered_at, created_at
		FROM outbox_messages
		WHERE payload = $1`, payload)
	assert.NoError(t, err)

	message, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[OutboxMessage])
	assert.NoError(t, err)

	return message
}

func countOutboxMessages(t *testing.T, payload any) int {
	t.Helper()

	var count int
	err := db.QueryRow(t.Context(),
		`SELECT count(*) FROM outbox_messages WHERE payload = $1`, payload).Scan(&count)
	assert.NoError(t, err)

	return count
}

func outboxMessageMarkers(messages []OutboxMessage) []string {
	markers := make([]string, 0, len(messages))

	for message := range slices.Values(messages) {
		if marker, ok := assert.JSONPath(message.Payload, "$.marker").(string); ok {
			markers = append(markers, marker)
		}
	}

	return markers
}

func TestInsertOutbox(t *testing.T) {
	t.Run("stores message fields", func(t *testing.T) {
		marker := uuid.New().String()
		payload := map[string]any{"user_id": marker}

		insertOutboxMessage(t, InsertOutboxParam[map[string]any, map[string]any]{
			Type:     "user.registered",
			Metadata: null.ValueFrom(map[string]any{"trace_id": marker}),
			Payload:  payload,
		})

		message := findOutboxMessage(t, payload)

		assert.NotEqual(t, message.ID, uuid.Nil)
		assert.Equal(t, message.Type, "user.registered")
		assert.Equal(t, assert.JSONPath(message.Payload, "$.user_id"), marker)
		assert.Equal(t, assert.JSONPath(message.Metadata, "$.trace_id"), marker)
		assert.False(t, message.Error.Valid)
		assert.False(t, message.DeliveredAt.Valid)
		assert.True(t, time.Since(message.CreatedAt) < time.Minute)
	})

	t.Run("stores struct metadata", func(t *testing.T) {
		type metadata struct {
			TraceID string `json:"trace_id"`
			Attempt int    `json:"attempt"`
		}

		marker := uuid.New().String()
		payload := map[string]any{"user_id": marker}

		insertOutboxMessage(t, InsertOutboxParam[metadata, map[string]any]{
			Type:     "user.registered",
			Metadata: null.ValueFrom(metadata{TraceID: marker, Attempt: 3}),
			Payload:  payload,
		})

		message := findOutboxMessage(t, payload)

		assert.Equal(t, assert.JSONPath(message.Metadata, "$.trace_id"), marker)
		assert.Equal(t, assert.JSONPath(message.Metadata, "$.attempt"), 3)
	})

	t.Run("stores unknown type when empty", func(t *testing.T) {
		payload := map[string]any{"marker": uuid.New().String()}

		insertOutboxMessage(t, InsertOutboxParam[map[string]any, map[string]any]{
			Payload: payload,
		})

		message := findOutboxMessage(t, payload)

		assert.Equal(t, message.Type, OutboxTypeUnknown)
		assert.Nil(t, message.Metadata)
	})

	t.Run("stores no message when transaction rolls back", func(t *testing.T) {
		payload := map[string]any{"marker": uuid.New().String()}

		tx, err := db.Begin(t.Context())
		assert.NoError(t, err)

		defer func() { _ = tx.Rollback(context.Background()) }()

		err = InsertOutbox(t.Context(), tx, InsertOutboxParam[map[string]any, map[string]any]{
			Type:    "user.registered",
			Payload: payload,
		})
		assert.NoError(t, err)
		assert.NoError(t, tx.Rollback(t.Context()))

		assert.Equal(t, countOutboxMessages(t, payload), 0)
	})
}

func TestGetOutboxMessages(t *testing.T) {
	t.Run("returns undelivered messages ordered by created_at", func(t *testing.T) {
		first := uuid.New().String()
		second := uuid.New().String()
		traceID := uuid.New().String()

		insertOutboxMessage(t, InsertOutboxParam[map[string]any, map[string]any]{
			Type:     "first",
			Metadata: null.ValueFrom(map[string]any{"trace_id": traceID}),
			Payload:  map[string]any{"marker": first},
		})
		time.Sleep(time.Millisecond)
		insertOutboxMessage(t, InsertOutboxParam[map[string]any, map[string]any]{
			Type:    "second",
			Payload: map[string]any{"marker": second},
		})

		messages := getOutboxMessages(t)

		assert.Equal(t, outboxMessageMarkers(messages), []string{first, second})
		assert.Equal(t, messages[0].Type, "first")
		assert.Equal(t, assert.JSONPath(messages[0].Metadata, "$.trace_id"), traceID)
		assert.False(t, messages[0].DeliveredAt.Valid)
	})

	t.Run("excludes delivered messages", func(t *testing.T) {
		delivered := uuid.New().String()
		undelivered := uuid.New().String()

		insertOutboxMessage(t, InsertOutboxParam[map[string]any, map[string]any]{
			Type:    "delivered",
			Payload: map[string]any{"marker": delivered},
		})
		insertOutboxMessage(t, InsertOutboxParam[map[string]any, map[string]any]{
			Type:    "undelivered",
			Payload: map[string]any{"marker": undelivered},
		})

		_, err := db.Exec(t.Context(), `
			UPDATE outbox_messages SET delivered_at = NOW()
			WHERE payload = $1`, map[string]any{"marker": delivered})
		assert.NoError(t, err)

		messages := getOutboxMessages(t)

		assert.Equal(t, outboxMessageMarkers(messages), []string{undelivered})
	})

	t.Run("limits messages to OutboxBatch", func(t *testing.T) {
		markers := make([]string, 0, OutboxBatch+1)

		for range OutboxBatch + 1 {
			marker := uuid.New().String()
			markers = append(markers, marker)

			insertOutboxMessage(t, InsertOutboxParam[map[string]any, map[string]any]{
				Type:    "batched",
				Payload: map[string]any{"marker": marker},
			})
			time.Sleep(time.Millisecond)
		}

		messages := getOutboxMessages(t)

		assert.Equal(t, len(messages), OutboxBatch)
		assert.Equal(t, outboxMessageMarkers(messages), markers[:OutboxBatch])
	})

	t.Run("skips messages locked by another transaction", func(t *testing.T) {
		marker := uuid.New().String()

		insertOutboxMessage(t, InsertOutboxParam[map[string]any, map[string]any]{
			Type:    "locked",
			Payload: map[string]any{"marker": marker},
		})

		first, err := db.Begin(t.Context())
		assert.NoError(t, err)

		defer func() { _ = first.Rollback(context.Background()) }()

		locked, err := GetOutboxMessages(t.Context(), first)
		assert.NoError(t, err)
		assert.True(t, slices.Contains(outboxMessageMarkers(locked), marker))

		second, err := db.Begin(t.Context())
		assert.NoError(t, err)

		defer func() { _ = second.Rollback(context.Background()) }()

		skipped, err := GetOutboxMessages(t.Context(), second)
		assert.NoError(t, err)
		assert.False(t, slices.Contains(outboxMessageMarkers(skipped), marker))
	})
}

func TestPollOutboxMessages(t *testing.T) {
	t.Run("publishes messages and marks them delivered", func(t *testing.T) {
		eventType := "test." + uuid.New().String()
		firstMarker := uuid.New().String()
		secondMarker := uuid.New().String()

		insertOutboxMessage(t, InsertOutboxParam[map[string]any, map[string]any]{
			Type:    eventType,
			Payload: map[string]any{"marker": firstMarker},
		})
		insertOutboxMessage(t, InsertOutboxParam[map[string]any, map[string]any]{
			Type:    eventType,
			Payload: map[string]any{"marker": secondMarker},
		})

		err := pollOutboxMessages(t.Context(), db, js)
		assert.NoError(t, err)

		subject := OutboxSubjectPrefix + "." + eventType

		for marker := range slices.Values([]string{firstMarker, secondMarker}) {
			message := findOutboxMessage(t, map[string]any{"marker": marker})

			assert.True(t, message.DeliveredAt.Valid)
			assert.False(t, message.Error.Valid)
		}

		stream, err := js.Stream(t.Context(), OutboxStream)
		assert.NoError(t, err)

		consumer, err := stream.CreateOrUpdateConsumer(t.Context(), jetstream.ConsumerConfig{
			Name:          "test-" + uuid.New().String(),
			FilterSubject: subject,
			AckPolicy:     jetstream.AckExplicitPolicy,
		})
		assert.NoError(t, err)

		published := map[string]string{}

		for {
			msg, err := consumer.Next(jetstream.FetchMaxWait(time.Second))
			if err != nil {
				break
			}

			if marker, ok := assert.JSONPath(msg.Data(), "$.marker").(string); ok {
				published[marker] = msg.Subject()
			}

			_ = msg.Ack()
		}

		assert.Equal(t, len(published), 2)
		assert.Equal(t, published[firstMarker], subject)
		assert.Equal(t, published[secondMarker], subject)
	})

	t.Run("marks message failed and keeps it undelivered when publish fails", func(t *testing.T) {
		eventType := "test." + uuid.New().String()
		marker := uuid.New().String()

		insertOutboxMessage(t, InsertOutboxParam[map[string]any, map[string]any]{
			Type:    eventType,
			Payload: map[string]any{"marker": marker},
		})

		assert.NoError(t, js.DeleteStream(t.Context(), OutboxStream))

		t.Cleanup(func() {
			assert.NoError(t, EnsureOutboxStream(context.Background(), js))
		})

		err := pollOutboxMessages(t.Context(), db, js)
		assert.NoError(t, err)

		message := findOutboxMessage(t, map[string]any{"marker": marker})

		assert.False(t, message.DeliveredAt.Valid)
		assert.True(t, message.Error.Valid)
		assert.True(t, strings.Contains(message.Error.String, "cannot publish outbox message"))
	})
}

func TestUpdateOutboxMessages(t *testing.T) {
	t.Run("updates delivered and failed messages in batches", func(t *testing.T) {
		deliveredMarker := uuid.New().String()
		firstFailedMarker := uuid.New().String()
		secondFailedMarker := uuid.New().String()

		for marker := range slices.Values([]string{deliveredMarker, firstFailedMarker, secondFailedMarker}) {
			insertOutboxMessage(t, InsertOutboxParam[map[string]any, map[string]any]{
				Type:    "test.batch",
				Payload: map[string]any{"marker": marker},
			})
		}

		delivered := findOutboxMessage(t, map[string]any{"marker": deliveredMarker})
		delivered.MarkAsDelivered()

		firstFailed := findOutboxMessage(t, map[string]any{"marker": firstFailedMarker})
		firstFailed.MarkAsError(errors.New("first failure"))

		secondFailed := findOutboxMessage(t, map[string]any{"marker": secondFailedMarker})
		secondFailed.MarkAsError(errors.New("second failure"))

		err := trx.WithTx(t.Context(), db, func(tx pgx.Tx) error {
			return UpdateOutboxMessages(t.Context(), tx, []OutboxMessage{delivered, firstFailed, secondFailed})
		})
		assert.NoError(t, err)

		updated := findOutboxMessage(t, map[string]any{"marker": deliveredMarker})
		assert.True(t, updated.DeliveredAt.Valid)
		assert.False(t, updated.Error.Valid)

		firstUpdated := findOutboxMessage(t, map[string]any{"marker": firstFailedMarker})
		assert.False(t, firstUpdated.DeliveredAt.Valid)
		assert.Equal(t, firstUpdated.Error.String, "first failure")

		secondUpdated := findOutboxMessage(t, map[string]any{"marker": secondFailedMarker})
		assert.False(t, secondUpdated.DeliveredAt.Valid)
		assert.Equal(t, secondUpdated.Error.String, "second failure")
	})
}
