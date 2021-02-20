package ketchup

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

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
		questionID, ok := a.tokenStore.Load(token)
		if !ok {
			a.rendererApp.Error(w, errors.New("token has expired"))
			return
		}

		if colors[questionID.(int64)].Answer != strings.TrimSpace(r.FormValue("answer")) {
			a.rendererApp.Error(w, errors.New("invalid question answer"))
			return
		}

		user := model.NewUser(0, r.FormValue("email"), authModel.User{
			Login:    r.FormValue("login"),
			Password: r.FormValue("password"),
		})

		if _, err := a.userService.Create(r.Context(), user); err != nil {
			a.rendererApp.Error(w, err)
			return
		}

		a.tokenStore.Delete(token)

		renderer.Redirect(w, r, fmt.Sprintf("%s/", appPath), renderer.NewSuccessMessage("Welcome to ketchup!"))
	})
}
