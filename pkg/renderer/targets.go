package renderer

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/ViBiOh/httputils/v3/pkg/crud"
	"github.com/ViBiOh/ketchup/pkg/target"
)

// SVG render a svg in given coolor
func (a app) targets() http.Handler {
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
	target := target.Target{
		Repository:     r.FormValue("repository"),
		CurrentVersion: r.FormValue("currentVersion"),
	}

	if errs := a.targetApp.Check(r.Context(), nil, target); len(errs) > 0 {
		a.handleCrudError(w, errs)
		return
	}

	if _, err := a.targetApp.Create(r.Context(), target); err != nil {
		a.errorHandler(w, http.StatusInternalServerError, err, nil)
		return
	}

	redirectWithMessage(w, r, fmt.Sprintf("%s created with success!", target.Repository))
}

func (a app) handleUpdate(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(strings.Trim(r.URL.Path, "/"), 10, 64)
	if err != nil {
		a.errorHandler(w, http.StatusBadRequest, err, nil)
		return
	}

	rawOldTarget, err := a.targetApp.Get(r.Context(), id)
	if err != nil {
		a.errorHandler(w, http.StatusBadRequest, err, nil)
		return
	}

	oldTarget := rawOldTarget.(target.Target)

	newTarget := target.Target{
		ID:             id,
		Repository:     r.FormValue("repository"),
		CurrentVersion: r.FormValue("currentVersion"),
		LatestVersion:  oldTarget.LatestVersion,
	}

	if errs := a.targetApp.Check(r.Context(), oldTarget, newTarget); len(errs) > 0 {
		a.handleCrudError(w, errs)
		return
	}

	if _, err := a.targetApp.Update(r.Context(), newTarget); err != nil {
		a.errorHandler(w, http.StatusInternalServerError, err, nil)
		return
	}

	redirectWithMessage(w, r, fmt.Sprintf("%s updated with success!", newTarget.Repository))
}

func (a app) handleDelete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(strings.Trim(r.URL.Path, "/"), 10, 64)
	if err != nil {
		a.errorHandler(w, http.StatusBadRequest, err, nil)
		return
	}

	rawTarget, err := a.targetApp.Get(r.Context(), id)
	if err != nil {
		a.errorHandler(w, http.StatusBadRequest, err, nil)
		return
	}

	target := rawTarget.(target.Target)

	if errs := a.targetApp.Check(r.Context(), target, nil); len(errs) > 0 {
		a.handleCrudError(w, errs)
		return
	}

	if err := a.targetApp.Delete(r.Context(), target); err != nil {
		a.errorHandler(w, http.StatusInternalServerError, err, nil)
		return
	}

	redirectWithMessage(w, r, fmt.Sprintf("%s deleted with success!", target.Repository))
}

func (a app) handleCrudError(w http.ResponseWriter, errs []crud.Error) {
	errorsValues := make([]error, 1+len(errs))
	errorsValues[0] = errors.New("invalid form")
	for i, err := range errs {
		errorsValues[i+1] = err
	}

	a.errorHandler(w, http.StatusBadRequest, errorsValues...)
}
