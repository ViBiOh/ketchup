package ui

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/ViBiOh/ketchup/pkg/target"
)

// SVG render a svg in given coolor
func (a app) targets() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			a.handleError(w, http.StatusMethodNotAllowed, fmt.Errorf("invalid method %s", r.Method), nil)
			return
		}

		if err := r.ParseForm(); err != nil {
			a.handleError(w, http.StatusBadRequest, err, nil)
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
			a.handleError(w, http.StatusBadRequest, fmt.Errorf("invalid method %s", method), nil)
		}
	})
}

func (a app) handleCreate(w http.ResponseWriter, r *http.Request) {
	target := target.Target{
		Repository:     r.FormValue("repository"),
		CurrentVersion: r.FormValue("currentVersion"),
	}

	if errors := a.targetApp.Check(r.Context(), nil, target); len(errors) > 0 {
		a.handleError(w, http.StatusBadRequest, fmt.Errorf("invalid form"), errors)
		return
	}

	if _, err := a.targetApp.Create(r.Context(), target); err != nil {
		a.handleError(w, http.StatusInternalServerError, err, nil)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func (a app) handleUpdate(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(strings.Trim(r.URL.Path, "/"), 10, 64)
	if err != nil {
		a.handleError(w, http.StatusBadRequest, err, nil)
		return
	}

	rawOldTarget, err := a.targetApp.Get(r.Context(), id)
	if err != nil {
		a.handleError(w, http.StatusBadRequest, err, nil)
		return
	}

	oldTarget := rawOldTarget.(target.Target)

	newTarget := target.Target{
		ID:             id,
		Repository:     r.FormValue("repository"),
		CurrentVersion: r.FormValue("currentVersion"),
		LatestVersion:  oldTarget.LatestVersion,
	}

	if errors := a.targetApp.Check(r.Context(), oldTarget, newTarget); len(errors) > 0 {
		a.handleError(w, http.StatusBadRequest, fmt.Errorf("invalid form"), errors)
		return
	}

	if _, err := a.targetApp.Update(r.Context(), newTarget); err != nil {
		a.handleError(w, http.StatusInternalServerError, err, nil)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func (a app) handleDelete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(strings.Trim(r.URL.Path, "/"), 10, 64)
	if err != nil {
		a.handleError(w, http.StatusBadRequest, err, nil)
		return
	}

	target, err := a.targetApp.Get(r.Context(), id)
	if err != nil {
		a.handleError(w, http.StatusBadRequest, err, nil)
		return
	}

	if errors := a.targetApp.Check(r.Context(), target, nil); len(errors) > 0 {
		a.handleError(w, http.StatusBadRequest, fmt.Errorf("invalid form"), errors)
		return
	}

	if err := a.targetApp.Delete(r.Context(), target); err != nil {
		a.handleError(w, http.StatusInternalServerError, err, nil)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}
