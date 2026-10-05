package webapi

import (
	"context"
	"encoding/json/v2"
	"errors"
	"io"
	"log/slog"
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samuelsih/golib/httpx"
	"github.com/samuelsih/golib/httpx/pbd"
	"github.com/samuelsih/golib/oas"
	"github.com/samuelsih/golib/slogx"
	"github.com/samuelsih/nibiru/api"
	"github.com/samuelsih/nibiru/api/app/auth"
	"github.com/samuelsih/nibiru/api/app/sandbox"
)

const (
	SessionSecurityScheme = "cookieAuth"
)

var (
	securitySchemes = []oas.SecurityRequirement{{SessionSecurityScheme: {}}}
)

type Server struct {
	*oas.APIServer

	authHandler    auth.Handler
	sandboxHandler sandbox.Handler

	sessionCookieName   string
	sessionCookieSecure bool
}

func Init(r *oas.APIServer, cfg api.Config, db *pgxpool.Pool) {
	server := Server{
		APIServer:           r,
		authHandler:         auth.NewHandler(db, cfg.SessionTTL),
		sandboxHandler:      sandbox.NewHandler(db),
		sessionCookieName:   cfg.SessionCookieName,
		sessionCookieSecure: cfg.SessionCookieSecure,
	}

	server.Router().RegisterErrorHandler(server.authErrorHandler())
	server.Router().RegisterErrorHandler(server.sandboxErrorHandler())
	server.Router().RegisterErrorHandler(server.fallbackErrorHandler())

	server.auth()
	server.sandbox()
}

func (s Server) fallbackErrorHandler() httpx.ErrorHandler {
	return func(w http.ResponseWriter, _ *http.Request, apperr error) httpx.HandleState {
		if problem, ok := errors.AsType[*pbd.Problem](apperr); ok {
			_ = problem.Write(w)
			return httpx.HandleStop
		}

		if errs, ok := errors.AsType[validation.Errors](apperr); ok {
			fields := make(map[string]any, len(errs))
			for field, fieldErr := range errs {
				fields[field] = fieldErr.Error()
			}

			_ = pbd.New(http.StatusUnprocessableEntity, pbd.WithErrors(fields)).Write(w)

			return httpx.HandleStop
		}

		_, bodyTooLarge := errors.AsType[*http.MaxBytesError](apperr)

		var status int

		switch {
		case bodyTooLarge:
			status = http.StatusRequestEntityTooLarge
		case errors.Is(apperr, httpx.ErrRouteNotFound):
			status = http.StatusNotFound
		case errors.Is(apperr, httpx.ErrMethodNotAllowed):
			status = http.StatusMethodNotAllowed
		case errors.Is(apperr, httpx.ErrInvalidRequestBody):
			status = http.StatusBadRequest
		case errors.Is(apperr, httpx.ErrTimeoutExceeded), errors.Is(apperr, context.DeadlineExceeded):
			status = http.StatusGatewayTimeout
		case errors.Is(apperr, context.Canceled):
			status = 499
		default:
			status = http.StatusInternalServerError
			slog.Error("unhandled error", slogx.ErrorAttr(apperr))
		}

		_ = pbd.New(status, pbd.WithDetail(apperr.Error())).Write(w)

		return httpx.HandleStop
	}
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
