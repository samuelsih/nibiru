package trx

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// WithTx runs fn inside a database transaction.
func WithTx(ctx context.Context, db *pgxpool.Pool, fn func(pgx.Tx) error) (err error) {
	tx, err := db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		}

		if err != nil {
			_ = tx.Rollback(ctx)
			return
		}

		if err = tx.Commit(ctx); err != nil {
			err = fmt.Errorf("commit transaction: %w", err)
		}
	}()

	err = fn(tx)

	return err
}

// IsErrorDuplicate reports whether err is a Postgres unique-constraint violation.
func IsErrorDuplicate(err error) bool {
	pgErr, ok := errors.AsType[*pgconn.PgError](err)

	return ok && pgErr.Code == "23505"
}
