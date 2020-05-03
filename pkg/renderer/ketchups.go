package renderer

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/ViBiOh/ketchup/pkg/model"
)

func (a app) ketchups() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			a.errorHandler(w, http.StatusMethodNotAllowed, fmt.Errorf("invalid method %s", r.Method))
			return
		}

		if err := r.ParseForm(); err != nil {
			a.errorHandler(w, http.StatusBadRequest, err)
			return
		}

		method := strings.ToUpper(r.FormValue("method"))

		switch method {
		case http.MethodPost:
			a.handleCreate(w, r)
		case http.MethodPut:
			a.handleUpdate(w, r)
		case http.MethodDelete:
			a.handleDelete(w, r)
		default:
			a.errorHandler(w, http.StatusBadRequest, fmt.Errorf("invalid method %s", method))
		}
	})
}

func (a app) handleCreate(w http.ResponseWriter, r *http.Request) {
	item := model.Ketchup{
		Version: r.FormValue("version"),
		Repository: model.Repository{
			Name: r.FormValue("repository"),
		},
	}

	created, err := a.ketchupService.Create(r.Context(), item)
	if err != nil {
		a.errorHandler(w, http.StatusInternalServerError, err)
		return
	}

	redirectWithMessage(w, r, fmt.Sprintf("%s/", appPath), fmt.Sprintf("%s created with success!", created.Repository.Name))
}

func (a app) handleUpdate(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(strings.Trim(r.URL.Path, "/"), 10, 64)
	if err != nil {
		a.errorHandler(w, http.StatusBadRequest, err)
		return
	}

	item := model.Ketchup{
		Version: r.FormValue("version"),
		Repository: model.Repository{
			ID: id,
		},
	}

	updated, err := a.ketchupService.Update(r.Context(), item)
	if err != nil {
		a.errorHandler(w, http.StatusInternalServerError, err)
		return
	}

	redirectWithMessage(w, r, fmt.Sprintf("%s/", appPath), fmt.Sprintf("Updated %s with success!", updated.Version))
}

func (a app) handleDelete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(strings.Trim(r.URL.Path, "/"), 10, 64)
	if err != nil {
		a.errorHandler(w, http.StatusBadRequest, err)
		return
	}

	item := model.Ketchup{
		Repository: model.Repository{
			ID: id,
		},
	}

	if err := a.ketchupService.Delete(r.Context(), item); err != nil {
		a.errorHandler(w, http.StatusInternalServerError, err)
		return
	}

	redirectWithMessage(w, r, fmt.Sprintf("%s/", appPath), "Deleted with success!")
}
