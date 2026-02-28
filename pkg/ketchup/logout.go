package ketchup

import (
	"net/http"

	"github.com/ViBiOh/httputils/v4/pkg/renderer"
)

func (s Service) Logout() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.logout.Logout(w, r)
		s.renderer.Redirect(w, r, "/", renderer.NewSuccessMessage("Logged out"))
	})
}
