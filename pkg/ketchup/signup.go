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

func (a app) Signup() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			a.rendererApp.Error(w, httpModel.WrapMethodNotAllowed(fmt.Errorf("invalid method %s", r.Method)))
			return
		}

		if err := r.ParseForm(); err != nil {
			a.rendererApp.Error(w, httpModel.WrapInvalid(err))
			return
		}

		token := r.FormValue("token")
		if !a.validateToken(r.Context(), token, r.FormValue("answer")) {
			a.rendererApp.Error(w, httpModel.WrapInvalid(errors.New("unable to validate security question")))
		}

		user := model.NewUser(0, r.FormValue("email"), authModel.User{
			Login:    r.FormValue("login"),
			Password: r.FormValue("password"),
		})

		if _, err := a.userService.Create(r.Context(), user); err != nil {
			a.rendererApp.Error(w, err)
			return
		}

		go a.cleanToken(context.Background(), token)

		renderer.Redirect(w, r, fmt.Sprintf("%s/", appPath), renderer.NewSuccessMessage("Welcome to ketchup!"))
	})
}
