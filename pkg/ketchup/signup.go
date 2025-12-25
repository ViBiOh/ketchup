package ketchup

import (
	"errors"
	"fmt"
	"net/http"

	httpModel "github.com/ViBiOh/httputils/v4/pkg/model"
	"github.com/ViBiOh/httputils/v4/pkg/renderer"
)

func (s Service) Signup() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			s.renderer.Error(w, r, nil, httpModel.WrapMethodNotAllowed(fmt.Errorf("invalid method %s", r.Method)))
			return
		}

		if err := r.ParseForm(); err != nil {
			s.renderer.Error(w, r, nil, httpModel.WrapInvalid(err))
			return
		}

		success, err := s.cap.Verify(r.Context(), r.FormValue("cap-token"))
		if err != nil {
			s.renderer.Error(w, r, nil, httpModel.WrapInternal(err))
			return
		}

		if !success {
			s.renderer.Error(w, r, nil, httpModel.WrapInternal(errors.New("invalid token")))
			return
		}

		if _, err := s.user.Create(r.Context(), r.FormValue("email"), r.FormValue("login"), r.FormValue("password")); err != nil {
			s.renderer.Error(w, r, nil, err)
			return
		}

		s.renderer.Redirect(w, r, fmt.Sprintf("%s/", appPath), renderer.NewSuccessMessage("Welcome to ketchup!"))
	})
}
