package app

import (
	"context"
	"errors"
	"time"
	"uuid"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samuelsih/golib/stringx"
	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidCredentials = errors.New("invalid email or password")

type LoginRequest struct {
	Email    string `json:"email"    example:"admin@gmail.com"`
	Password string `json:"password" example:"Password.1"`
}

func (r LoginRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Email,
			validation.Required,
			validation.Length(3, 320),
			is.Email,
		),
		validation.Field(&r.Password,
			validation.Required,
		),
	)
}

func Login(ctx context.Context, db *pgxpool.Pool, ttl time.Duration, r LoginRequest) (User, Session, error) {
	var (
		user    User
		session Session
	)

	err := WithTx(ctx, db, func(tx pgx.Tx) error {
		var err error

		user, err = FindUserByEmail(ctx, tx, r.Email)
		if err != nil {
			if errors.Is(err, ErrUserNotFound) {
				return ErrInvalidCredentials
			}

			return err
		}

		if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(r.Password)) != nil {
			return ErrInvalidCredentials
		}

		now := time.Now()
		session = Session{
			ID:        uuid.NewV7(),
			Token:     stringx.Base64(40),
			UserID:    user.ID,
			CreatedAt: now,
			ExpiresAt: now.Add(ttl),
		}

		return SaveSession(ctx, tx, session)
	})
	if err != nil {
		return User{}, Session{}, err
	}

	return user, session, nil
}
