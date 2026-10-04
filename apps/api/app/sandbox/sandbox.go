package sandbox

import "github.com/jackc/pgx/v5/pgxpool"

type Handler struct {
	db   *pgxpool.Pool
	repo Repo
}

func NewHandler(db *pgxpool.Pool) Handler {
	return Handler{
		db:   db,
		repo: NewRepo(db),
	}
}
