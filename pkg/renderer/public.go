package renderer

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	authModel "github.com/ViBiOh/auth/v2/pkg/model"
	"github.com/ViBiOh/httputils/v3/pkg/httperror"
	"github.com/ViBiOh/httputils/v3/pkg/templates"
	"github.com/ViBiOh/ketchup/pkg/model"
)

func (a app) publicHandler(w http.ResponseWriter, r *http.Request, status int, message model.Message) {
	questionID := a.rand.Int63n(int64(len(colors)))

	content := map[string]interface{}{
		"Version":    a.version,
		"QuestionID": questionID,
		"Question":   colors[questionID].Question,
	}

	if len(message.Content) > 0 {
		content["Message"] = message
	}

	if err := templates.ResponseHTMLTemplate(a.tpl.Lookup("public"), w, content, status); err != nil {
		httperror.InternalServerError(w, err)
	}
}

func (a app) signup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		a.errorHandler(w, http.StatusMethodNotAllowed, fmt.Errorf("invalid method %s", r.Method))
		return
	}

	if err := r.ParseForm(); err != nil {
		a.errorHandler(w, http.StatusBadRequest, err)
		return
	}

	questionID, err := strconv.ParseInt(r.FormValue("question"), 10, 64)
	if err != nil {
		a.errorHandler(w, http.StatusBadRequest, err)
		return
	}

	if colors[questionID].Answer != strings.TrimSpace(r.FormValue("answer")) {
		a.errorHandler(w, http.StatusBadRequest, errors.New("invalid answer for question"))
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
		a.errorHandler(w, http.StatusInternalServerError, err)
		return
	}

	redirectWithMessage(w, r, fmt.Sprintf("https://%s:%s@%s%s/", user.Login.Login, user.Login.Password, a.uiPath, appPath), "Welcome to ketchup!")
}
