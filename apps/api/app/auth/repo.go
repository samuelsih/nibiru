package auth

import (
	"context"
	"errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samuelsih/nibiru/api/pkg/trx"
)

type Repo struct {
	db      *pgxpool.Pool
	builder sq.StatementBuilderType
}

func NewRepo(db *pgxpool.Pool) Repo {
	return Repo{
		db:      db,
		builder: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (r Repo) SaveUser(ctx context.Context, tx pgx.Tx, user User) error {
	query, args, err := r.builder.Insert("users").
		Columns("id", "email", "password", "first_name", "last_name", "created_at", "updated_at").
		Values(user.ID, user.Email, user.PasswordHash, user.FirstName, user.LastName, user.CreatedAt, sq.Expr("NOW()")).
		ToSql()
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, query, args...)
	if err != nil {
		if trx.IsErrorDuplicate(err) {
			return ErrEmailDuplicate
		}

		return err
	}

	return nil
}

func (r Repo) FindUserByEmail(ctx context.Context, tx pgx.Tx, email string) (User, error) {
	query, args, err := r.builder.Select("id", "email", "password", "first_name", "last_name", "created_at").
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

func (r Repo) FindSessionByToken(ctx context.Context, token string) (Session, error) {
	query, args, err := r.builder.Select("id", "token", "user_id", "created_at", "expires_at").
		From("sessions").
		Where(sq.Eq{"token": token}).
		ToSql()
	if err != nil {
		return Session{}, err
	}

	rows, err := r.db.Query(ctx, query, args...)
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

func (r Repo) FindUserByID(ctx context.Context, id uuid.UUID) (User, error) {
	query, args, err := r.builder.Select("id", "email", "password", "first_name", "last_name", "created_at").
		From("users").
		Where(sq.Eq{"id": id.String()}).
		ToSql()
	if err != nil {
		return User{}, err
	}

	rows, err := r.db.Query(ctx, query, args...)
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

func (r Repo) SaveSession(ctx context.Context, tx pgx.Tx, session Session) error {
	query, args, err := r.builder.Insert("sessions").
		Columns("id", "token", "user_id", "created_at", "expires_at").
		Values(session.ID, session.Token, session.UserID, session.CreatedAt, session.ExpiresAt).
		ToSql()
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, query, args...)

	return err
}

func (r Repo) DeleteSessionByToken(ctx context.Context, tx pgx.Tx, token string) error {
	query, args, err := r.builder.Delete("sessions").
		Where(sq.Eq{"token": token}).
		ToSql()
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, query, args...)

	return err
}
