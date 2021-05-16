package ketchup

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	httpModel "github.com/ViBiOh/httputils/v4/pkg/model"
	"github.com/ViBiOh/httputils/v4/pkg/renderer"
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

	ketchupFrequency, err := model.ParseKetchupFrequency(r.FormValue("frequency"))
	if err != nil {
		a.rendererApp.Error(w, httpModel.WrapInvalid(err))
		return
	}

	var repository model.Repository
	name := r.FormValue("name")

	switch repositoryKind {
	case model.Github:
		repository = model.NewGithubRepository(0, name)
	case model.Helm:
		repository = model.NewHelmRepository(0, strings.TrimSuffix(name, "/"), r.FormValue("part"))
	default:
		a.rendererApp.Error(w, httpModel.WrapInternal(fmt.Errorf("unhandled repository kind `%s`", repositoryKind)))
	}

	item := model.NewKetchup(r.FormValue("pattern"), r.FormValue("version"), ketchupFrequency, repository)

	created, err := a.ketchupService.Create(r.Context(), item)
	if err != nil {
		a.rendererApp.Error(w, err)
		return
	}

	a.rendererApp.Redirect(w, r, fmt.Sprintf("%s/", appPath), renderer.NewSuccessMessage(fmt.Sprintf("%s created with success!", created.Repository.Name)))
}

func (a app) handleUpdate(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(strings.Trim(r.URL.Path, "/"), 10, 64)
	if err != nil {
		a.rendererApp.Error(w, httpModel.WrapInvalid(err))
		return
	}

	ketchupFrequency, err := model.ParseKetchupFrequency(r.FormValue("frequency"))
	if err != nil {
		a.rendererApp.Error(w, httpModel.WrapInvalid(err))
		return
	}

	item := model.NewKetchup(r.FormValue("pattern"), r.FormValue("version"), ketchupFrequency, model.NewGithubRepository(id, ""))

	updated, err := a.ketchupService.Update(r.Context(), item)
	if err != nil {
		a.rendererApp.Error(w, err)
		return
	}

	a.rendererApp.Redirect(w, r, fmt.Sprintf("%s/", appPath), renderer.NewSuccessMessage(fmt.Sprintf("Updated %s with success!", updated.Version)))
}

func (a app) handleDelete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(strings.Trim(r.URL.Path, "/"), 10, 64)
	if err != nil {
		a.rendererApp.Error(w, httpModel.WrapInvalid(err))
		return
	}

	item := model.NewKetchup("", "", model.Daily, model.NewGithubRepository(id, ""))

	if err := a.ketchupService.Delete(r.Context(), item); err != nil {
		a.rendererApp.Error(w, err)
		return
	}

	a.rendererApp.Redirect(w, r, fmt.Sprintf("%s/", appPath), renderer.NewSuccessMessage("Deleted with success!"))
}
