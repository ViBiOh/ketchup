package ketchup

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	authModel "github.com/ViBiOh/auth/v2/pkg/model"
	rendererModel "github.com/ViBiOh/httputils/v3/pkg/renderer/model"
	"github.com/ViBiOh/ketchup/pkg/model"
)

func (a app) Signup() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			a.rendererApp.Error(w, rendererModel.WrapMethodNotAllowed(fmt.Errorf("invalid method %s", r.Method)))
			return
		}

		if err := r.ParseForm(); err != nil {
			a.rendererApp.Error(w, rendererModel.WrapInvalid(err))
			return
		}

		token := r.FormValue("token")
		questionID, ok := a.tokenStore.Load(token)
		if !ok {
			a.rendererApp.Error(w, errors.New("token has expired"))
			return
		}

		if colors[questionID.(int64)].Answer != strings.TrimSpace(r.FormValue("answer")) {
			a.rendererApp.Error(w, errors.New("invalid question answer"))
			return
		}

		user := model.User{
			Email: r.FormValue("email"),
			Login: authModel.User{
				Login:    r.FormValue("login"),
				Password: r.FormValue("password"),
			},
		}

		if _, err := a.userService.Create(r.Context(), user); err != nil {
			a.rendererApp.Error(w, err)
			return
		}

		a.tokenStore.Delete(token)

		a.rendererApp.Redirect(w, r, fmt.Sprintf("/%s", appPath), rendererModel.NewSuccessMessage("Welcome to ketchup!"))
	})
}
