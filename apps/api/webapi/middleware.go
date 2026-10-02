package webapi

import (
	"context"
	"errors"
	"net/http"

	"github.com/samuelsih/golib/httpx"
	"github.com/samuelsih/nibiru/api/app/auth"
)

type userContextKey struct{}

func (s Server) MiddlewareAuthenticated() httpx.Middleware {
	return func(next httpx.Handler) httpx.Handler {
		return func(w http.ResponseWriter, r *http.Request) error {
			cookie, err := r.Cookie(s.sessionCookieName)
			if err != nil {
				return auth.ErrInvalidSession
			}

			user, err := s.authHandler.WhoAmI(r.Context(), cookie.Value)
			if err != nil {
				return err
			}

			ctx := context.WithValue(r.Context(), userContextKey{}, user)

			return next(w, r.WithContext(ctx))
		}
	}
}

func handleAuthError(w http.ResponseWriter, _ *http.Request, err error) httpx.HandleState {
	if errors.Is(err, auth.ErrInvalidSession) {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)

		return httpx.HandleStop
	}

	return httpx.HandleContinue
}
