package app

import (
	"context"
	"errors"
	"time"
	"uuid"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/guregu/null/v6"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type RegisterRequest struct {
	Email     string      `json:"email"     example:"admin@gmail.com"`
	Password  string      `json:"password"  example:"Password.1"`
	FirstName string      `json:"firstName" example:"Admin"`
	LastName  null.String `json:"lastName"  example:"New"             required:"false"`
}

func (r RegisterRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Email,
			validation.Required,
			validation.Length(3, 320),
			is.Email,
		),
		validation.Field(&r.FirstName,
			validation.Required,
			validation.Length(5, 100),
		),
		validation.Field(&r.LastName, validation.By(func(any) error {
			if !r.LastName.Valid {
				return nil
			}

			return validation.Validate(r.LastName.String, validation.Length(0, 100))
		})),
		validation.Field(&r.Password,
			validation.Required,
			validation.Length(8, 72),
		),
	)
}

func Register(ctx context.Context, db *pgxpool.Pool, r RegisterRequest) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(r.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := User{
		ID:           uuid.NewV7(),
		Email:        r.Email,
		PasswordHash: string(hash),
		FirstName:    r.FirstName,
		LastName:     r.LastName,
		CreatedAt:    time.Now(),
	}

	return WithTx(ctx, db, func(tx pgx.Tx) error {
		_, err := FindUserByEmail(ctx, tx, r.Email)
		if err == nil {
			return ErrEmailDuplicate
		}

		if !errors.Is(err, ErrUserNotFound) {
			return err
		}

		return SaveUser(ctx, tx, user)
	})
}
