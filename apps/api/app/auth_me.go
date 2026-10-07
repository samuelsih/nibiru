package app

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

func WhoAmI(ctx context.Context, db *pgxpool.Pool, token string) (User, error) {
	session, err := FindSessionByToken(ctx, db, token)
	if err != nil {
		return User{}, err
	}

	if session.Expired() {
		return User{}, ErrInvalidSession
	}

	return FindUserByID(ctx, db, session.UserID)
}
