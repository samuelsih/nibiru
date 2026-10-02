package webapi

import (
	"encoding/json/v2"
	"io"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samuelsih/golib/oas"
	"github.com/samuelsih/nibiru/api"
	"github.com/samuelsih/nibiru/api/app/auth"
)

type Server struct {
	*oas.APIServer

	authHandler auth.Handler

	sessionCookieName   string
	sessionCookieSecure bool
}

func Init(r *oas.APIServer, cfg api.Config, db *pgxpool.Pool) {
	server := Server{
		APIServer:           r,
		authHandler:         auth.NewHandler(db, cfg.SessionTTL),
		sessionCookieName:   cfg.SessionCookieName,
		sessionCookieSecure: cfg.SessionCookieSecure,
	}

	server.auth()
}

func JSONUnmarshal[T any](r io.ReadCloser) (T, error) {
	defer r.Close()

	var out T

	err := json.UnmarshalRead(r, &out)
	if err != nil {
		return out, err
	}

	return out, nil
}

func JSONMarshal[T any](w http.ResponseWriter, body T) error {
	w.Header().Set("Content-Type", "application/json")
	return json.MarshalWrite(w, body)
}
