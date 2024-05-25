package ketchup

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	authModel "github.com/ViBiOh/auth/v2/pkg/model"
	httpModel "github.com/ViBiOh/httputils/v4/pkg/model"
	"github.com/ViBiOh/httputils/v4/pkg/renderer"
	"github.com/ViBiOh/ketchup/pkg/model"
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

		token := r.FormValue("token")
		if !s.validateToken(r.Context(), token, r.FormValue("answer")) {
			s.renderer.Error(w, r, nil, httpModel.WrapInvalid(errors.New("validate security question")))
			return
		}

		user := model.NewUser(0, r.FormValue("email"), authModel.User{
			Login:    r.FormValue("login"),
			Password: r.FormValue("password"),
		})

		if _, err := s.user.Create(r.Context(), user); err != nil {
			s.renderer.Error(w, r, nil, err)
			return
		}

		go s.cleanToken(context.WithoutCancel(r.Context()), token)

		s.renderer.Redirect(w, r, fmt.Sprintf("%s/", appPath), renderer.NewSuccessMessage("Welcome to ketchup!"))
	})
}
