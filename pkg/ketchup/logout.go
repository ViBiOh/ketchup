package ketchup

import (
	"net/http"

	"github.com/ViBiOh/httputils/v4/pkg/renderer"
	"github.com/ViBiOh/httputils/v4/pkg/telemetry"
)

func (s Service) Logout() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		telemetry.SetRouteTag(ctx, "/logout")

		s.logout.Logout(w, r)

		s.renderer.Serve(w, r, renderer.NewPage("auth", http.StatusOK, map[string]any{
			"Redirect": "/",
			"Message":  renderer.NewSuccessMessage("Logout success!"),
		}))
	})
}
