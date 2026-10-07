package webapi

import (
	"errors"
	"net/http"
	"strconv"
	"uuid"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/samuelsih/golib/httpx"
	"github.com/samuelsih/golib/oas"
	"github.com/samuelsih/nibiru/api/app"
)

func (s Server) sandboxErrorHandler() httpx.ErrorHandler {
	return func(w http.ResponseWriter, _ *http.Request, apperr error) httpx.HandleState {
		if !errors.Is(apperr, app.ErrInvalidSpec) {
			return httpx.HandleContinue
		}

		return httpx.HandleStop
	}
}

func (s Server) sandbox() {
	tag := []string{"Sandbox"}

	s.GroupPrefix("/sandbox", func(r *oas.APIServer) {
		r.Use(s.MiddlewareAuthenticated())

		r.Get("/", s.sandboxList).Spec(oas.Spec{
			OperationID: "sandboxList",
			Summary:     "List sandboxes",
			Description: "Lists the authenticated user's sandboxes, newest first, using cursor pagination.",
			Tags:        tag,
			Security:    securitySchemes,
			Params:      oas.SpecParams[listSandboxesParams](),
			Responses: []oas.ResponseSpec{
				{
					Status:      http.StatusOK,
					Description: "A page of sandboxes.",
					Body:        oas.SpecBody[app.ListResult](),
				},
			},
		})

		r.Post("/", s.sandboxCreate).Spec(oas.Spec{
			OperationID: "sandboxCreate",
			Summary:     "Create a sandbox",
			Tags:        tag,
			Security:    securitySchemes,
			Body:        oas.SpecBody[app.CreateSandboxRequest](),
			Responses: []oas.ResponseSpec{
				{
					Status:      http.StatusCreated,
					Description: "The created sandbox.",
					Body:        oas.SpecBody[app.SandboxSummary](),
				},
			},
		})
	})
}

type listSandboxesParams struct {
	Cursor uuid.UUID `query:"cursor" example:"01924a7d-9b3e-7c1a-8f4b-7c1d2e3f4a5b" required:"false"`
	Limit  int       `query:"limit" example:"20"  required:"false"`
}

func (p listSandboxesParams) Validate() error {
	return validation.ValidateStruct(&p,
		validation.Field(&p.Limit, validation.Min(1), validation.Max(app.MaxListLimit)),
	)
}

func (s Server) sandboxCreate(w http.ResponseWriter, r *http.Request) error {
	user, ok := r.Context().Value(userContextKey{}).(app.User)
	if !ok {
		return app.ErrInvalidSession
	}

	req, err := JSONUnmarshal[app.CreateSandboxRequest](r.Body)
	if err != nil {
		return err
	}

	if err := req.Validate(); err != nil {
		return err
	}

	req.OwnerID = user.ID

	instance, err := app.CreateSandbox(r.Context(), s.db, req)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	return JSONMarshal(w, app.SandboxSummary{
		ID:    instance.ID,
		Name:  instance.Name,
		CPU:   instance.CPU,
		RAM:   instance.MemoryGB,
		Disk:  instance.DiskGB,
		State: instance.State,
	})
}

func (s Server) sandboxList(w http.ResponseWriter, r *http.Request) error {
	user, ok := r.Context().Value(userContextKey{}).(app.User)
	if !ok {
		return app.ErrInvalidSession
	}

	query := r.URL.Query()
	params := listSandboxesParams{Limit: app.DefaultListLimit}

	if raw := query.Get("cursor"); raw != "" {
		cursor, err := uuid.Parse(raw)
		if err != nil {
			return validation.Errors{"cursor": errors.New("must be a valid UUID")}
		}

		params.Cursor = cursor
	}

	if raw := query.Get("limit"); raw != "" {
		limit, err := strconv.Atoi(raw)
		if err != nil {
			return validation.Errors{"limit": errors.New("must be an integer")}
		}

		params.Limit = limit
	}

	if err := params.Validate(); err != nil {
		return err
	}

	result, err := app.ListSandboxes(r.Context(), s.db, user.ID, params.Cursor, params.Limit)
	if err != nil {
		return err
	}

	return JSONMarshal(w, result)
}
