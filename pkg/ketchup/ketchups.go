package ketchup

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	httpModel "github.com/ViBiOh/httputils/v4/pkg/model"
	"github.com/ViBiOh/httputils/v4/pkg/renderer"
	"github.com/ViBiOh/ketchup/pkg/model"
	"github.com/ViBiOh/ketchup/pkg/semver"
)

func (s Service) ketchups() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			s.renderer.Error(w, r, nil, httpModel.WrapMethodNotAllowed(fmt.Errorf("invalid method %s", r.Method)))
			return
		}

		if err := r.ParseForm(); err != nil {
			s.renderer.Error(w, r, nil, httpModel.WrapInvalid(err))
			return
		}

		method := strings.ToUpper(r.FormValue("method"))

		switch method {
		case http.MethodPost:
			s.handleCreate(w, r)
		case http.MethodPut:
			s.handleUpdate(w, r)
		case http.MethodDelete:
			s.handleDelete(w, r)
		default:
			s.renderer.Error(w, r, nil, httpModel.WrapInvalid(fmt.Errorf("invalid method %s", method)))
		}
	})
}

func (s Service) handleCreate(w http.ResponseWriter, r *http.Request) {
	repositoryKind, err := model.ParseRepositoryKind(r.FormValue("kind"))
	if err != nil {
		s.renderer.Error(w, r, nil, httpModel.WrapInvalid(err))
		return
	}

	ketchupFrequency, err := model.ParseKetchupFrequency(r.FormValue("frequency"))
	if err != nil {
		s.renderer.Error(w, r, nil, httpModel.WrapInvalid(err))
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

	created, err := s.ketchup.Create(r.Context(), item)
	if err != nil {
		s.renderer.Error(w, r, nil, toHttpError(err))
		return
	}

	if err := s.cache.EvictOnSuccess(ctx, model.ReadUser(ctx), nil); err != nil {
		slog.LogAttrs(ctx, slog.LevelError, "evict suggests cache", slog.Any("error", err))
	}

	s.renderer.Redirect(w, r, fmt.Sprintf("%s/", appPath), renderer.NewSuccessMessage(fmt.Sprintf("%s created with success!", created.Repository.Name)))
}

func (s Service) handleUpdate(w http.ResponseWriter, r *http.Request) {
	rawID := strings.Trim(r.URL.Path, "/")
	if rawID == "all" {
		if err := s.ketchup.UpdateAll(r.Context()); err != nil {
			s.renderer.Error(w, r, nil, toHttpError(err))
		} else {
			s.renderer.Redirect(w, r, fmt.Sprintf("%s/", appPath), renderer.NewSuccessMessage("All ketchups are up-to-date!"))
		}

		return
	}

	id, err := strconv.ParseUint(strings.Trim(r.URL.Path, "/"), 10, 64)
	if err != nil {
		s.renderer.Error(w, r, nil, httpModel.WrapInvalid(err))
		return
	}

	ketchupFrequency, err := model.ParseKetchupFrequency(r.FormValue("frequency"))
	if err != nil {
		s.renderer.Error(w, r, nil, httpModel.WrapInvalid(err))
		return
	}

	updateWhenNotify := r.FormValue("update-when-notify") == "true"

	item := model.NewKetchup(r.FormValue("pattern"), r.FormValue("version"), ketchupFrequency, updateWhenNotify, model.NewGithubRepository(model.Identifier(id), "")).WithID()

	updated, err := s.ketchup.Update(r.Context(), r.FormValue("old-pattern"), item)
	if err != nil {
		s.renderer.Error(w, r, nil, toHttpError(err))
		return
	}

	s.renderer.Redirect(w, r, fmt.Sprintf("%s/", appPath), renderer.NewSuccessMessage(fmt.Sprintf("Updated %s with success!", updated.Version)))
}

func (s Service) handleDelete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(strings.Trim(r.URL.Path, "/"), 10, 64)
	if err != nil {
		s.renderer.Error(w, r, nil, httpModel.WrapInvalid(err))
		return
	}

	ctx := r.Context()
	item := model.NewKetchup(r.FormValue("pattern"), "", model.Daily, false, model.NewGithubRepository(model.Identifier(id), "")).WithID()

	if err := s.ketchup.Delete(ctx, item); err != nil {
		s.renderer.Error(w, r, nil, toHttpError(err))
		return
	}

	if err := s.cache.EvictOnSuccess(ctx, model.ReadUser(ctx), nil); err != nil {
		slog.LogAttrs(ctx, slog.LevelError, "evict suggests cache", slog.Any("error", err))
	}

	s.renderer.Redirect(w, r, fmt.Sprintf("%s/", appPath), renderer.NewSuccessMessage("Deleted with success!"))
}

func toHttpError(err error) error {
	switch {
	case errors.Is(err, semver.ErrPatternInvalid) || errors.Is(err, semver.ErrPrefixInvalid):
		return httpModel.WrapInvalid(err)
	default:
		return err
	}
}
