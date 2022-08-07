package ketchup

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/ViBiOh/httputils/v4/pkg/cache"
	"github.com/ViBiOh/httputils/v4/pkg/logger"
	httpModel "github.com/ViBiOh/httputils/v4/pkg/model"
	"github.com/ViBiOh/httputils/v4/pkg/renderer"
	"github.com/ViBiOh/ketchup/pkg/model"
)

func (a App) ketchups() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			a.rendererApp.Error(w, r, nil, httpModel.WrapMethodNotAllowed(fmt.Errorf("invalid method %s", r.Method)))
			return
		}

		if err := r.ParseForm(); err != nil {
			a.rendererApp.Error(w, r, nil, httpModel.WrapInvalid(err))
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
			a.rendererApp.Error(w, r, nil, httpModel.WrapInvalid(fmt.Errorf("invalid method %s", method)))
		}
	})
}

func (a App) handleCreate(w http.ResponseWriter, r *http.Request) {
	repositoryKind, err := model.ParseRepositoryKind(r.FormValue("kind"))
	if err != nil {
		a.rendererApp.Error(w, r, nil, httpModel.WrapInvalid(err))
		return
	}

	ketchupFrequency, err := model.ParseKetchupFrequency(r.FormValue("frequency"))
	if err != nil {
		a.rendererApp.Error(w, r, nil, httpModel.WrapInvalid(err))
		return
	}

	updateWhenNotify := r.FormValue("update-when-notify") == "true"

	var repository model.Repository
	name := r.FormValue("name")

	switch repositoryKind {
	case model.Helm:
		repository = model.NewHelmRepository(model.Identifier(0), strings.TrimSuffix(name, "/"), r.FormValue("part"))
	case model.Docker:
		repository = model.NewRepository(0, repositoryKind, strings.TrimPrefix(name, "docker.io/"), "")
	default:
		repository = model.NewRepository(0, repositoryKind, name, "")
	}

	ctx := r.Context()
	item := model.NewKetchup(r.FormValue("pattern"), r.FormValue("version"), ketchupFrequency, updateWhenNotify, repository).WithID()

	created, err := a.ketchupService.Create(r.Context(), item)
	if err != nil {
		a.rendererApp.Error(w, r, nil, err)
		return
	}

	if err := cache.EvictOnSuccess(ctx, a.redisApp, suggestCacheKey(model.ReadUser(ctx)), nil); err != nil {
		logger.Error("evict suggests cache: %s", err)
	}

	a.rendererApp.Redirect(w, r, fmt.Sprintf("%s/", appPath), renderer.NewSuccessMessage(fmt.Sprintf("%s created with success!", created.Repository.Name)))
}

func (a App) handleUpdate(w http.ResponseWriter, r *http.Request) {
	rawID := strings.Trim(r.URL.Path, "/")
	if rawID == "all" {
		if err := a.ketchupService.UpdateAll(r.Context()); err != nil {
			a.rendererApp.Error(w, r, nil, err)
		} else {
			a.rendererApp.Redirect(w, r, fmt.Sprintf("%s/", appPath), renderer.NewSuccessMessage("All ketchups are up-to-date!"))
		}

		return
	}

	id, err := strconv.ParseUint(strings.Trim(r.URL.Path, "/"), 10, 64)
	if err != nil {
		a.rendererApp.Error(w, r, nil, httpModel.WrapInvalid(err))
		return
	}

	ketchupFrequency, err := model.ParseKetchupFrequency(r.FormValue("frequency"))
	if err != nil {
		a.rendererApp.Error(w, r, nil, httpModel.WrapInvalid(err))
		return
	}

	updateWhenNotify := r.FormValue("update-when-notify") == "true"

	item := model.NewKetchup(r.FormValue("pattern"), r.FormValue("version"), ketchupFrequency, updateWhenNotify, model.NewGithubRepository(model.Identifier(id), "")).WithID()

	updated, err := a.ketchupService.Update(r.Context(), r.FormValue("old-pattern"), item)
	if err != nil {
		a.rendererApp.Error(w, r, nil, err)
		return
	}

	a.rendererApp.Redirect(w, r, fmt.Sprintf("%s/", appPath), renderer.NewSuccessMessage(fmt.Sprintf("Updated %s with success!", updated.Version)))
}

func (a App) handleDelete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(strings.Trim(r.URL.Path, "/"), 10, 64)
	if err != nil {
		a.rendererApp.Error(w, r, nil, httpModel.WrapInvalid(err))
		return
	}

	ctx := r.Context()
	item := model.NewKetchup(r.FormValue("pattern"), "", model.Daily, false, model.NewGithubRepository(model.Identifier(id), "")).WithID()

	if err := a.ketchupService.Delete(ctx, item); err != nil {
		a.rendererApp.Error(w, r, nil, err)
		return
	}

	if err := cache.EvictOnSuccess(ctx, a.redisApp, suggestCacheKey(model.ReadUser(ctx)), nil); err != nil {
		logger.Error("evict suggests cache: %s", err)
	}

	a.rendererApp.Redirect(w, r, fmt.Sprintf("%s/", appPath), renderer.NewSuccessMessage("Deleted with success!"))
}
