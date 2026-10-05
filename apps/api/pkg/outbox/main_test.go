package outbox

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/ory/dockertest/v4"
	"github.com/samuelsih/golib/sqlmigration"
	migrationpgx "github.com/samuelsih/golib/sqlmigration/pgx"
)

var (
	db       *pgxpool.Pool
	natsConn *nats.Conn
	js       jetstream.JetStream
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	pool, _ := dockertest.NewPool(ctx, "")
	if err := setup(ctx, pool); err != nil {
		_ = pool.Close(ctx)
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	code := m.Run()

	if natsConn != nil {
		natsConn.Close()
	}

	_ = pool.Close(ctx)
	os.Exit(code)
}

func setup(ctx context.Context, pool dockertest.ClosablePool) error {
	postgres, err := pool.Run(ctx, "docker.io/library/postgres",
		dockertest.WithTag("18.6-trixie@sha256:5a5a84b19854a9ffaa54082c166ff4ec27473a361e496e5ea167f298f2da9722"),
		dockertest.WithEnv([]string{
			"POSTGRES_PASSWORD=secret",
			"POSTGRES_DB=testdb",
		}),
	)
	if err != nil {
		return fmt.Errorf("Cannot run postgres: %w", err)
	}

	dsn := fmt.Sprintf("postgres://postgres:secret@%s/testdb?sslmode=disable", postgres.GetHostPort("5432/tcp"))
	db, err = pgxpool.New(ctx, dsn)
	if err != nil {
		return fmt.Errorf("Cannot connect postgres: %w", err)
	}

	err = pool.Retry(ctx, 30*time.Second, func() error {
		return db.Ping(ctx)
	})
	if err != nil {
		return fmt.Errorf("Cannot ping postgres: %w", err)
	}

	if err = sqlmigration.Up(ctx, migrationpgx.New(db), os.DirFS("../.."), "migrations"); err != nil {
		return fmt.Errorf("cannot run migrations: %w", err)
	}

	natsContainer, err := pool.Run(ctx, "docker.io/library/nats",
		dockertest.WithTag("2.15.0@sha256:cd3fcd4ecdda44e3a66728a5334af0a959bc3979b32810e033d1c547241cd0f4"),
		dockertest.WithCmd([]string{"-js", "-sd", "/data", "-m", "8222"}),
	)
	if err != nil {
		return fmt.Errorf("Cannot run nats: %w", err)
	}

	natsURL := "nats://" + natsContainer.GetHostPort("4222/tcp")

	err = pool.Retry(ctx, 30*time.Second, func() error {
		if natsConn == nil {
			natsConn, err = nats.Connect(natsURL, nats.Name("nibiru-outbox-test"))
			if err != nil {
				return err
			}

			js, err = jetstream.New(natsConn)
			if err != nil {
				return err
			}
		}

		_, err = js.AccountInfo(ctx)

		return err
	})
	if err != nil {
		return fmt.Errorf("Cannot connect nats: %w", err)
	}

	if err = EnsureStream(ctx, js); err != nil {
		return fmt.Errorf("cannot ensure outbox stream: %w", err)
	}

	return nil
}
