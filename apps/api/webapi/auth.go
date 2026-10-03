package webapi

import (
	"errors"
	"net/http"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/guregu/null/v6"
	"github.com/samuelsih/golib/httpx"
	"github.com/samuelsih/golib/httpx/pbd"
	"github.com/samuelsih/golib/oas"
	"github.com/samuelsih/nibiru/api/app/auth"
)

func (s Server) auth() {
	tag := []string{"Authentication"}

	s.Router().RegisterErrorHandler(func(w http.ResponseWriter, _ *http.Request, apperr error) httpx.HandleState {
		var status int

		switch {
		case errors.Is(apperr, auth.ErrEmailDuplicate):
			status = http.StatusConflict
		case errors.Is(apperr, auth.ErrInvalidCredentials), errors.Is(apperr, auth.ErrInvalidSession):
			status = http.StatusUnauthorized
		case errors.Is(apperr, auth.ErrUserNotFound):
			status = http.StatusNotFound
		default:
			return httpx.HandleContinue
		}

		_ = pbd.New(status, pbd.WithDetail(apperr.Error())).Write(w)

		return httpx.HandleStop
	})

	s.Router().RegisterErrorHandler(s.fallbackErrorHandler())

	s.GroupPrefix("/auth", func(r *oas.APIServer) {
		r.Post("/register", s.authRegister).Spec(oas.Spec{
			OperationID: "authRegister",
			Tags:        tag,
			Body:        oas.SpecBody[registerRequest](),
			Responses: []oas.ResponseSpec{
				{Status: http.StatusNoContent, Description: "User registered."},
			},
		})

		r.Post("/login", s.authLogin).Spec(oas.Spec{
			OperationID: "authLogin",
			Tags:        tag,
			Body:        oas.SpecBody[loginRequest](),
			Responses: []oas.ResponseSpec{
				{
					Status:      http.StatusOK,
					Description: "Login succeeded and session cookie set.",
					Body:        oas.SpecBody[userResponse](),
				},
			},
		})

		r.Get("/me", s.authMe, s.MiddlewareAuthenticated()).Spec(oas.Spec{
			OperationID: "authMe",
			Tags:        tag,
			Security:    []oas.SecurityRequirement{{SessionSecurityScheme: {}}},
			Responses: []oas.ResponseSpec{
				{Status: http.StatusOK, Description: "Authenticated user.", Body: oas.SpecBody[userResponse]()},
			},
		})

		r.Post("/logout", s.authLogout).Spec(oas.Spec{
			OperationID: "authLogout",
			Tags:        tag,
			Responses: []oas.ResponseSpec{
				{Status: http.StatusNoContent, Description: "Session deleted and session cookie cleared."},
			},
		})
	})
}

type registerRequest struct {
	Email     string      `json:"email"     example:"admin@gmail.com"`
	FirstName string      `json:"firstName" example:"Admin"`
	LastName  null.String `json:"lastName"  example:"New"             required:"false"`
	Password  string      `json:"password"  example:"Password.1"`
}

func (r registerRequest) Validate() error {
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

func (s Server) authRegister(w http.ResponseWriter, r *http.Request) error {
	req, err := JSONUnmarshal[registerRequest](r.Body)
	if err != nil {
		return err
	}

	if err := req.Validate(); err != nil {
		return err
	}

	err = s.authHandler.Register(r.Context(), auth.RegisterRequest{
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Password:  req.Password,
	})
	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusNoContent)

	return nil
}

type loginRequest struct {
	Email    string `json:"email"    example:"admin@gmail.com"`
	Password string `json:"password" example:"Password.1"`
}

func (r loginRequest) Validate() error {
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

type userResponse struct {
	Email     string      `json:"email"     example:"admin@gmail.com"`
	FirstName string      `json:"firstName" example:"Admin"`
	LastName  null.String `json:"lastName"  example:"New"`
	CreatedAt time.Time   `json:"createdAt" example:"2026-01-02T15:04:05Z"`
}

func (s Server) authLogin(w http.ResponseWriter, r *http.Request) error {
	req, err := JSONUnmarshal[loginRequest](r.Body)
	if err != nil {
		return err
	}

	if err := req.Validate(); err != nil {
		return err
	}

	user, session, err := s.authHandler.Login(r.Context(), auth.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	})
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

	return JSONMarshal(w, userResponse{
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		CreatedAt: user.CreatedAt,
	})
}

func (s Server) authMe(w http.ResponseWriter, r *http.Request) error {
	user, ok := r.Context().Value(userContextKey{}).(auth.User)
	if !ok {
		return auth.ErrInvalidSession
	}

	return JSONMarshal(w, userResponse{
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		CreatedAt: user.CreatedAt,
	})
}

func (s Server) authLogout(w http.ResponseWriter, r *http.Request) error {
	cookie, err := r.Cookie(s.sessionCookieName)
	if err != nil && !errors.Is(err, http.ErrNoCookie) {
		return err
	}

	if err == nil {
		if err := s.authHandler.Logout(r.Context(), cookie.Value); err != nil {
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
