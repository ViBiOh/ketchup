package ketchup

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	httpModel "github.com/ViBiOh/httputils/v3/pkg/model"
	"github.com/ViBiOh/httputils/v3/pkg/renderer"
	"github.com/ViBiOh/ketchup/pkg/model"
)

func (a app) ketchups() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			a.rendererApp.Error(w, httpModel.WrapMethodNotAllowed(fmt.Errorf("invalid method %s", r.Method)))
			return
		}

		if err := r.ParseForm(); err != nil {
			a.rendererApp.Error(w, httpModel.WrapInvalid(err))
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
			a.rendererApp.Error(w, httpModel.WrapInvalid(fmt.Errorf("invalid method %s", method)))
		}
	})
}

func (a app) handleCreate(w http.ResponseWriter, r *http.Request) {
	repositoryKind, err := model.ParseRepositoryKind(r.FormValue("kind"))
	if err != nil {
		a.rendererApp.Error(w, httpModel.WrapInvalid(err))
		return
	}

	item := model.Ketchup{
		Pattern: r.FormValue("pattern"),
		Version: r.FormValue("version"),
		Repository: model.Repository{
			Name: r.FormValue("repository"),
			Kind: repositoryKind,
		},
	}

	created, err := a.ketchupService.Create(r.Context(), item)
	if err != nil {
		fmt.Println(err)
		a.rendererApp.Error(w, err)
		return
	}

	renderer.Redirect(w, r, fmt.Sprintf("%s/", appPath), renderer.NewSuccessMessage(fmt.Sprintf("%s created with success!", created.Repository.Name)))
}

func (a app) handleUpdate(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(strings.Trim(r.URL.Path, "/"), 10, 64)
	if err != nil {
		a.rendererApp.Error(w, httpModel.WrapInvalid(err))
		return
	}

	item := model.Ketchup{
		Pattern: r.FormValue("pattern"),
		Version: r.FormValue("version"),
		Repository: model.Repository{
			ID: id,
		},
	}

	updated, err := a.ketchupService.Update(r.Context(), item)
	if err != nil {
		a.rendererApp.Error(w, err)
		return
	}

	renderer.Redirect(w, r, fmt.Sprintf("%s/", appPath), renderer.NewSuccessMessage(fmt.Sprintf("Updated %s with success!", updated.Version)))
}

func (a app) handleDelete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(strings.Trim(r.URL.Path, "/"), 10, 64)
	if err != nil {
		a.rendererApp.Error(w, httpModel.WrapInvalid(err))
		return
	}

	item := model.Ketchup{
		Repository: model.Repository{
			ID: id,
		},
	}

	if err := a.ketchupService.Delete(r.Context(), item); err != nil {
		a.rendererApp.Error(w, err)
		return
	}

	renderer.Redirect(w, r, fmt.Sprintf("%s/", appPath), renderer.NewSuccessMessage("Deleted with success!"))
}
