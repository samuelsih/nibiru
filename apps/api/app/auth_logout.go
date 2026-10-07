package app

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Logout(ctx context.Context, db *pgxpool.Pool, token string) error {
	return WithTx(ctx, db, func(tx pgx.Tx) error {
		return DeleteSessionByToken(ctx, tx, token)
	})
}
