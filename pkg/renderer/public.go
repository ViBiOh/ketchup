package renderer

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	authModel "github.com/ViBiOh/auth/v2/pkg/model"
	"github.com/ViBiOh/httputils/v3/pkg/httperror"
	"github.com/ViBiOh/httputils/v3/pkg/logger"
	"github.com/ViBiOh/httputils/v3/pkg/templates"
	"github.com/ViBiOh/ketchup/pkg/model"
)

func (a app) generateToken() (string, int64) {
	questionID := a.rand.Int63n(int64(len(colors)))
	return a.tokenStore.Store(questionID, time.Minute*5), questionID
}

func (a app) publicHandler(w http.ResponseWriter, r *http.Request, status int, message model.Message) {
	token, questionID := a.generateToken()

	suggests, err := a.repositoryService.Suggest(r.Context(), []uint64{0}, 3)
	if err != nil {
		logger.Warn("unable to get publics suggestions: %s", err)
	}

	content := map[string]interface{}{
		"Version":  a.version,
		"Token":    token,
		"Question": colors[questionID].Question,
		"Suggests": suggests,
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

	token := r.FormValue("token")
	questionID, ok := a.tokenStore.Load(token)
	if !ok {
		a.errorHandler(w, http.StatusInternalServerError, errors.New("token has expired"))
		return
	}

	if colors[questionID.(int64)].Answer != strings.TrimSpace(r.FormValue("answer")) {
		a.errorHandler(w, http.StatusInternalServerError, errors.New("invalid question answer"))
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

	a.tokenStore.Delete(token)

	redirectWithMessage(w, r, fmt.Sprintf("https://%s:%s@%s%s/", user.Login.Login, user.Login.Password, a.uiPath, appPath), "Welcome to ketchup!")
}
