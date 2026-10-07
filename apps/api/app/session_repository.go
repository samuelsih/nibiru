package app

import (
	"context"
	"errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrInvalidSession = errors.New("invalid session")

func SaveSession(ctx context.Context, tx pgx.Tx, session Session) error {
	query, args, err := builder.Insert("sessions").
		Columns("id", "token", "user_id", "created_at", "expires_at").
		Values(session.ID, session.Token, session.UserID, session.CreatedAt, session.ExpiresAt).
		ToSql()
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, query, args...)

	return err
}

func DeleteSessionByToken(ctx context.Context, tx pgx.Tx, token string) error {
	query, args, err := builder.Delete("sessions").
		Where(sq.Eq{"token": token}).
		ToSql()
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, query, args...)

	return err
}

func FindSessionByToken(ctx context.Context, db *pgxpool.Pool, token string) (Session, error) {
	query, args, err := builder.Select("id", "token", "user_id", "created_at", "expires_at").
		From("sessions").
		Where(sq.Eq{"token": token}).
		ToSql()
	if err != nil {
		return Session{}, err
	}

	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return Session{}, err
	}

	session, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[Session])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Session{}, ErrInvalidSession
		}

		return Session{}, err
	}

	return session, nil
}
