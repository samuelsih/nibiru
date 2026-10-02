package app

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ory/dockertest/v4"
	"github.com/samuelsih/golib/sqlmigration"
	migrationpgx "github.com/samuelsih/golib/sqlmigration/pgx"
)

var db *pgxpool.Pool

func TestMain(m *testing.M) {
	ctx := context.Background()

	pool, _ := dockertest.NewPool(ctx, "")
	if err := setup(ctx, pool); err != nil {
		_ = pool.Close(ctx)
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	code := m.Run()
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

	if err = sqlmigration.Up(ctx, migrationpgx.New(db), os.DirFS(".."), "migrations"); err != nil {
		return fmt.Errorf("cannot run migrations: %w", err)
	}

	return nil
}
