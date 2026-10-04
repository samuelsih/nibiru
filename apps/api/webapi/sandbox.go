package webapi

import (
	"errors"
	"net/http"
	"strconv"
	"uuid"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/guregu/null/v6"
	"github.com/samuelsih/golib/oas"
	"github.com/samuelsih/nibiru/api/app/auth"
	"github.com/samuelsih/nibiru/api/app/sandbox"
)

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
					Body:        oas.SpecBody[listSandboxesResponse](),
				},
			},
		})

		r.Post("/", nil).Spec(oas.Spec{
			OperationID: "sandboxCreate",
			Summary:     "Create a sandbox",
			Tags:        tag,
			Security:    securitySchemes,
			Body:        oas.SpecBody[createSandboxRequest](),
			Responses: []oas.ResponseSpec{
				{
					Status:      http.StatusCreated,
					Description: "The created sandbox.",
					Body:        oas.SpecBody[sandbox.InstanceSummary](),
				},
			},
		})
	})
}

type listSandboxesParams struct {
	Cursor uuid.UUID `query:"cursor" description:"Cursor returned as nextCursor by the previous page." example:"01924a7d-9b3e-7c1a-8f4b-7c1d2e3f4a5b" required:"false"`
	Limit  int       `query:"limit"  description:"Maximum number of sandboxes per page."               example:"20"                                   required:"false"`
}

func (p listSandboxesParams) Validate() error {
	return validation.ValidateStruct(&p,
		validation.Field(&p.Limit, validation.Min(1), validation.Max(sandbox.MaxListLimit)),
	)
}

type listSandboxesResponse struct {
	Items      []sandbox.InstanceSummary `json:"items"`
	NextCursor null.Value[uuid.UUID]     `json:"nextCursor"`
}

type createSandboxRequest struct {
	Name string `json:"name" example:"worker-1" required:"false"`
	CPU  int    `json:"cpu"  example:"4"`
	RAM  int    `json:"ram"  example:"8"`
	Disk int    `json:"disk" example:"50"`
}

func (r createSandboxRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Name, validation.Length(0, 255)),
		validation.Field(&r.CPU, validation.Required),
		validation.Field(&r.RAM, validation.Required),
		validation.Field(&r.Disk, validation.Required),
	)
}

func (s Server) sandboxList(w http.ResponseWriter, r *http.Request) error {
	user, ok := r.Context().Value(userContextKey{}).(auth.User)
	if !ok {
		return auth.ErrInvalidSession
	}

	query := r.URL.Query()
	params := listSandboxesParams{Limit: sandbox.DefaultListLimit}

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

	result, err := s.sandboxHandler.List(r.Context(), uuid.UUID(user.ID), params.Cursor, params.Limit)
	if err != nil {
		return err
	}

	var nextCursor null.Value[uuid.UUID]
	if result.NextCursor != uuid.Nil() {
		nextCursor = null.ValueFrom(result.NextCursor)
	}

	return JSONMarshal(w, listSandboxesResponse{
		Items:      result.Items,
		NextCursor: nextCursor,
	})
}
