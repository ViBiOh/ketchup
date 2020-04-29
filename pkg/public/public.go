package public

import (
	"errors"
	"net/http"

	"github.com/ViBiOh/httputils/v3/pkg/httperror"
	"github.com/ViBiOh/httputils/v3/pkg/httpjson"
	"github.com/ViBiOh/httputils/v3/pkg/request"
	"github.com/ViBiOh/ketchup/pkg/service"
	"github.com/ViBiOh/ketchup/pkg/service/user"
)

// App of package
type App interface {
	Handler() http.Handler
}

type app struct {
	userService user.App
}

// New creates new App from Config
func New(userService user.App) App {
	return app{
		userService: userService,
	}
}

// Handler for request. Should be use with net/http
func (a app) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/signup" {
			if r.Method == http.MethodPost {
				a.signup(w, r)
				return
			}

			w.WriteHeader(http.StatusMethodNotAllowed)
		}

		httperror.NotFound(w)
	})
}

func (a app) signup(w http.ResponseWriter, r *http.Request) {
	payload, err := request.ReadBodyRequest(r)
	if err != nil {
		httperror.BadRequest(w, err)
		return
	}

	item, err := a.userService.Unmarshal(payload, r.Header.Get("Content-Type"))
	if err != nil {
		httperror.BadRequest(w, err)
		return
	}

	user, err := a.userService.Create(r.Context(), item)
	if err != nil {
		if errors.Is(err, service.ErrInvalid) {
			httperror.BadRequest(w, err)
			return
		}

		httperror.InternalServerError(w, err)
		return
	}

	httpjson.ResponseJSON(w, http.StatusCreated, user, httpjson.IsPretty(r))
}
