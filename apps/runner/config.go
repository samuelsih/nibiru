package runner

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/samuelsih/golib/slogx"
)

const (
	OutboxStream          = "OUTBOX"
	OutboxStreamSubjects  = "outbox.>"
	OutboxConsumerDurable = "runner"
	PubSubName            = "nibiru-runner"
)

type Config struct {
	PubSubURL string `env:"PUBSUB_URL,required,notEmpty"`
}

func (c Config) ConnectPubSub(ctx context.Context) (jetstream.JetStream, jetstream.Consumer, error) {
	conn, err := nats.Connect(c.PubSubURL,
		nats.Name(PubSubName),
		nats.DisconnectErrHandler(func(_ *nats.Conn, err error) {
			if err != nil {
				slog.Warn("NATS disconnected", slogx.ErrorAttr(err))
			}
		}),
		nats.ClosedHandler(func(_ *nats.Conn) {
			slog.Info("NATS connection closed", slog.String("url", c.PubSubURL))
		}),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot connect to pubsub instance: %w", err)
	}

	js, err := jetstream.New(conn)
	if err != nil {
		conn.Close()
		return nil, nil, fmt.Errorf("cannot create jetstream context: %w", err)
	}

	if _, err = js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:     OutboxStream,
		Subjects: []string{OutboxStreamSubjects},
	}); err != nil {
		conn.Close()
		return nil, nil, fmt.Errorf("cannot create or update outbox stream: %w", err)
	}

	consumer, err := js.CreateOrUpdateConsumer(ctx, OutboxStream, jetstream.ConsumerConfig{
		Durable:       OutboxConsumerDurable,
		FilterSubject: OutboxStreamSubjects,
		AckPolicy:     jetstream.AckExplicitPolicy,
	})
	if err != nil {
		conn.Close()
		return nil, nil, fmt.Errorf("cannot create or update runner consumer: %w", err)
	}

	return js, consumer, nil
}
