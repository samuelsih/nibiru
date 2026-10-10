package runner

import (
	"context"
	"log/slog"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/samuelsih/golib/slogx"
)

func Consume(ctx context.Context, consumer jetstream.Consumer) {
	consumeCtx, err := consumer.Consume(func(msg jetstream.Msg) {
		slog.Info("Received message",
			slog.String("subject", msg.Subject()),
			slog.String("data", string(msg.Data())),
		)

		if err := msg.Ack(); err != nil {
			slog.Error("cannot ack message", slogx.ErrorAttr(err))
		}
	})
	if err != nil {
		slog.Error("cannot start consuming messages", slogx.ErrorAttr(err))
		return
	}

	defer consumeCtx.Stop()

	slog.Info("Starting to consume messages")
}
