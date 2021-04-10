package ketchup

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	authModel "github.com/ViBiOh/auth/v2/pkg/model"
	"github.com/ViBiOh/httputils/v4/pkg/logger"
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
		questionIDString, err := a.tokenApp.Load(r.Context(), token)
		if err != nil {
			a.rendererApp.Error(w, httpModel.WrapInvalid(fmt.Errorf("unable to retrieve captcha token: %s", err)))
			return
		}

		questionID, err := strconv.ParseInt(questionIDString, 10, 64)
		if err != nil {
			a.rendererApp.Error(w, fmt.Errorf("question id is not numerical: %s", err))
			return
		}

		if colors[questionID].Answer != strings.TrimSpace(r.FormValue("answer")) {
			a.rendererApp.Error(w, httpModel.WrapInvalid(errors.New("invalid question answer")))
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

		if err := a.tokenApp.Delete(r.Context(), token); err != nil {
			logger.Warn("unable to delete token: %s", err)
		}

		renderer.Redirect(w, r, fmt.Sprintf("%s/", appPath), renderer.NewSuccessMessage("Welcome to ketchup!"))
	})
}
