package webapi

import "github.com/samuelsih/golib/oas"

func (s Server) sandbox() {
	tag := []string{"Authentication"}

	s.GroupPrefix("/sandbox", func(r *oas.APIServer) {
		r.Get("/", nil).Spec(oas.Spec{
			Tags:     tag,
			Security: securitySchemes,
		})
		r.Post("/", nil).Spec(oas.Spec{
			Tags:     tag,
			Security: securitySchemes,
		})
	})
}
