package webapi

import (
	"errors"
	"net/http"

	"github.com/samuelsih/golib/httpx"
	"github.com/samuelsih/golib/httpx/pbd"
	"github.com/samuelsih/golib/oas"
	"github.com/samuelsih/nibiru/api/app"
)

func (s Server) authErrorHandler() httpx.ErrorHandler {
	return func(w http.ResponseWriter, _ *http.Request, apperr error) httpx.HandleState {
		var status int

		switch {
		case errors.Is(apperr, app.ErrEmailDuplicate):
			status = http.StatusConflict
		case errors.Is(apperr, app.ErrInvalidCredentials), errors.Is(apperr, app.ErrInvalidSession):
			status = http.StatusUnauthorized
		case errors.Is(apperr, app.ErrUserNotFound):
			status = http.StatusNotFound
		default:
			return httpx.HandleContinue
		}

		_ = pbd.New(status, pbd.WithDetail(apperr.Error())).Write(w)

		return httpx.HandleStop
	}
}

func (s Server) auth() {
	tag := []string{"Authentication"}

	s.GroupPrefix("/auth", func(r *oas.APIServer) {
		r.Post("/register", s.authRegister).Spec(oas.Spec{
			OperationID: "authRegister",
			Tags:        tag,
			Body:        oas.SpecBody[app.RegisterRequest](),
			Responses: []oas.ResponseSpec{
				{Status: http.StatusNoContent, Description: "User registered."},
			},
		})

		r.Post("/login", s.authLogin).Spec(oas.Spec{
			OperationID: "authLogin",
			Tags:        tag,
			Body:        oas.SpecBody[app.LoginRequest](),
			Responses: []oas.ResponseSpec{
				{
					Status:      http.StatusOK,
					Description: "Login succeeded and session cookie set.",
					Body:        oas.SpecBody[app.User](),
				},
			},
		})

		r.Get("/me", s.authMe, s.MiddlewareAuthenticated()).Spec(oas.Spec{
			OperationID: "authMe",
			Tags:        tag,
			Security:    securitySchemes,
			Responses: []oas.ResponseSpec{
				{Status: http.StatusOK, Description: "Authenticated user.", Body: oas.SpecBody[app.User]()},
			},
		})

		r.Post("/logout", s.authLogout).Spec(oas.Spec{
			OperationID: "authLogout",
			Tags:        tag,
			Security:    securitySchemes,
			Responses: []oas.ResponseSpec{
				{Status: http.StatusNoContent, Description: "Session deleted and session cookie cleared."},
			},
		})
	})
}

func (s Server) authRegister(w http.ResponseWriter, r *http.Request) error {
	req, err := JSONUnmarshal[app.RegisterRequest](r.Body)
	if err != nil {
		return err
	}

	if err := req.Validate(); err != nil {
		return err
	}

	if err := app.Register(r.Context(), s.db, req); err != nil {
		return err
	}

	w.WriteHeader(http.StatusNoContent)

	return nil
}

func (s Server) authLogin(w http.ResponseWriter, r *http.Request) error {
	req, err := JSONUnmarshal[app.LoginRequest](r.Body)
	if err != nil {
		return err
	}

	if err := req.Validate(); err != nil {
		return err
	}

	user, session, err := app.Login(r.Context(), s.db, s.sessionTTL, req)
	if err != nil {
		return err
	}

	http.SetCookie(w, &http.Cookie{ //nolint:gosec // Secure is configured through SESSION_COOKIE_SECURE.
		Name:     s.sessionCookieName,
		Value:    session.Token,
		Path:     "/",
		Expires:  session.ExpiresAt,
		HttpOnly: true,
		Secure:   s.sessionCookieSecure,
		SameSite: http.SameSiteLaxMode,
	})

	return JSONMarshal(w, user)
}

func (s Server) authMe(w http.ResponseWriter, r *http.Request) error {
	user, ok := r.Context().Value(userContextKey{}).(app.User)
	if !ok {
		return app.ErrInvalidSession
	}

	return JSONMarshal(w, user)
}

func (s Server) authLogout(w http.ResponseWriter, r *http.Request) error {
	cookie, err := r.Cookie(s.sessionCookieName)
	if err != nil && !errors.Is(err, http.ErrNoCookie) {
		return err
	}

	if err == nil {
		if err := app.Logout(r.Context(), s.db, cookie.Value); err != nil {
			return err
		}
	}

	http.SetCookie(w, &http.Cookie{ //nolint:gosec // Secure is configured through SESSION_COOKIE_SECURE.
		Name:     s.sessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   s.sessionCookieSecure,
		SameSite: http.SameSiteLaxMode,
	})

	w.WriteHeader(http.StatusNoContent)
	return nil
}
