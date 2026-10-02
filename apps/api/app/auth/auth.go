package auth

import (
	"context"
	"errors"
	"time"

	"github.com/guregu/null/v6"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samuelsih/nibiru/api/pkg/trx"
)

var (
	ErrEmailDuplicate     = errors.New("email already exists")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid email or password")
)

type Handler struct {
	repo       Repo
	sessionTTL time.Duration
}

func NewHandler(db *pgxpool.Pool, sessionTTL time.Duration) Handler {
	return Handler{
		repo:       NewRepo(db),
		sessionTTL: sessionTTL,
	}
}

type RegisterRequest struct {
	Email     string
	Password  string
	FirstName string
	LastName  null.String
}

func (h Handler) Register(ctx context.Context, r RegisterRequest) error {
	user, err := NewUser(r.Email, r.Password, r.FirstName, r.LastName)
	if err != nil {
		return err
	}

	return trx.WithTx(ctx, h.repo.db, func(tx pgx.Tx) error {
		_, err := h.repo.FindUserByEmail(ctx, tx, r.Email)
		if err == nil {
			return ErrEmailDuplicate
		}

		if !errors.Is(err, ErrUserNotFound) {
			return err
		}

		return h.repo.SaveUser(ctx, tx, user)
	})
}

type LoginRequest struct {
	Email    string
	Password string
}

func (h Handler) Login(ctx context.Context, r LoginRequest) (User, Session, error) {
	var (
		user    User
		session Session
	)

	err := trx.WithTx(ctx, h.repo.db, func(tx pgx.Tx) error {
		var err error

		user, err = h.repo.FindUserByEmail(ctx, tx, r.Email)
		if err != nil {
			if errors.Is(err, ErrUserNotFound) {
				return ErrInvalidCredentials
			}

			return err
		}

		if !user.CorrectPassword(r.Password) {
			return ErrInvalidCredentials
		}

		session = NewSession(user.ID, h.sessionTTL)

		return h.repo.SaveSession(ctx, tx, session)
	})
	if err != nil {
		return User{}, Session{}, err
	}

	return user, session, nil
}

func (h Handler) Logout(ctx context.Context, token string) error {
	return trx.WithTx(ctx, h.repo.db, func(tx pgx.Tx) error {
		return h.repo.DeleteSessionByToken(ctx, tx, token)
	})
}
