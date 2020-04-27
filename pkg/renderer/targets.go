package renderer

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/ViBiOh/httputils/v3/pkg/crud"
	"github.com/ViBiOh/ketchup/pkg/model"
)

func (a app) ketchups() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			a.errorHandler(w, http.StatusMethodNotAllowed, fmt.Errorf("invalid method %s", r.Method), nil)
			return
		}

		if err := r.ParseForm(); err != nil {
			a.errorHandler(w, http.StatusBadRequest, err, nil)
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
			a.errorHandler(w, http.StatusBadRequest, fmt.Errorf("invalid method %s", method), nil)
		}
	})
}

func (a app) handleCreate(w http.ResponseWriter, r *http.Request) {
	ketchup := model.Ketchup{
		Version: r.FormValue("currentVersion"),
		Repository: model.Repository{
			Name: r.FormValue("repository"),
		},
	}

	if errs := a.ketchupApp.Check(r.Context(), nil, ketchup); len(errs) > 0 {
		a.handleCrudError(w, errs)
		return
	}

	if _, err := a.ketchupApp.Create(r.Context(), ketchup); err != nil {
		a.errorHandler(w, http.StatusInternalServerError, err, nil)
		return
	}

	redirectWithMessage(w, r, fmt.Sprintf("%s created with success!", ketchup.Repository.Name))
}

func (a app) handleUpdate(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(strings.Trim(r.URL.Path, "/"), 10, 64)
	if err != nil {
		a.errorHandler(w, http.StatusBadRequest, err, nil)
		return
	}

	rawOldKetchup, err := a.ketchupApp.Get(r.Context(), id)
	if err != nil {
		a.errorHandler(w, http.StatusBadRequest, err, nil)
		return
	}

	oldKetchup := rawOldKetchup.(model.Ketchup)

	newKetchup := model.Ketchup{
		Version: r.FormValue("currentVersion"),
		Repository: model.Repository{
			Name: r.FormValue("repository"),
		},
	}

	if errs := a.ketchupApp.Check(r.Context(), oldKetchup, newKetchup); len(errs) > 0 {
		a.handleCrudError(w, errs)
		return
	}

	if _, err := a.ketchupApp.Update(r.Context(), newKetchup); err != nil {
		a.errorHandler(w, http.StatusInternalServerError, err, nil)
		return
	}

	redirectWithMessage(w, r, fmt.Sprintf("%s updated with success!", newKetchup.Repository.Name))
}

func (a app) handleDelete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(strings.Trim(r.URL.Path, "/"), 10, 64)
	if err != nil {
		a.errorHandler(w, http.StatusBadRequest, err, nil)
		return
	}

	rawKetchup, err := a.ketchupApp.Get(r.Context(), id)
	if err != nil {
		a.errorHandler(w, http.StatusBadRequest, err, nil)
		return
	}

	ketchup := rawKetchup.(model.Ketchup)

	if errs := a.ketchupApp.Check(r.Context(), ketchup, nil); len(errs) > 0 {
		a.handleCrudError(w, errs)
		return
	}

	if err := a.ketchupApp.Delete(r.Context(), ketchup); err != nil {
		a.errorHandler(w, http.StatusInternalServerError, err, nil)
		return
	}

	redirectWithMessage(w, r, fmt.Sprintf("%s deleted with success!", ketchup.Repository.Name))
}

func (a app) handleCrudError(w http.ResponseWriter, errs []crud.Error) {
	errorsValues := make([]error, 1+len(errs))
	errorsValues[0] = errors.New("invalid form")
	for i, err := range errs {
		errorsValues[i+1] = err
	}

	a.errorHandler(w, http.StatusBadRequest, errorsValues...)
}
