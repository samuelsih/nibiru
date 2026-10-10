package app

import (
	"context"
	"errors"
	"uuid"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrEmailDuplicate = errors.New("email already exists")
	ErrUserNotFound   = errors.New("user not found")
)

func SaveUser(ctx context.Context, tx pgx.Tx, user User) error {
	query, args, err := builder.Insert("users").
		Columns("id", "email", "password", "first_name", "last_name", "created_at", "updated_at").
		Values(user.ID, user.Email, user.PasswordHash, user.FirstName, user.LastName, user.CreatedAt, sq.Expr("NOW()")).
		ToSql()
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, query, args...)
	if err != nil {
		if IsErrorDuplicate(err) {
			return ErrEmailDuplicate
		}

		return err
	}

	return nil
}

func FindUserByEmail(ctx context.Context, tx pgx.Tx, email string) (User, error) {
	query, args, err := builder.Select("id", "email", "password", "first_name", "last_name", "created_at").
		From("users").
		Where(sq.Eq{"email": email}).
		ToSql()
	if err != nil {
		return User{}, err
	}

	rows, err := tx.Query(ctx, query, args...)
	if err != nil {
		return User{}, err
	}

	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[User])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, ErrUserNotFound
		}

		return User{}, err
	}

	return user, nil
}

func FindUserByID(ctx context.Context, db *pgxpool.Pool, id uuid.UUID) (User, error) {
	query, args, err := builder.Select("id", "email", "password", "first_name", "last_name", "created_at").
		From("users").
		Where(sq.Eq{"id": id.String()}).
		ToSql()
	if err != nil {
		return User{}, err
	}

	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return User{}, err
	}

	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[User])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, ErrUserNotFound
		}

		return User{}, err
	}

	return user, nil
}
