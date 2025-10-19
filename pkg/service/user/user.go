package user

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	authModel "github.com/ViBiOh/auth/v3/pkg/model"
	httpModel "github.com/ViBiOh/httputils/v4/pkg/model"
	"github.com/ViBiOh/ketchup/pkg/model"
)

type Service struct {
	store model.UserStore
	auth  model.AuthService
}

func New(userStore model.UserStore, authApp model.AuthService) Service {
	return Service{
		store: userStore,
		auth:  authApp,
	}
}

func (s Service) StoreInContext(ctx context.Context) context.Context {
	id := authModel.ReadUser(ctx).ID
	if id == 0 {
		slog.WarnContext(ctx, "no login user in context")
		return ctx
	}

	item, err := s.store.GetByLoginID(ctx, id)
	if err != nil || item.IsZero() {
		slog.LogAttrs(ctx, slog.LevelError, "get user with login", slog.Uint64("id", id), slog.Any("error", err))
		return ctx
	}

	return model.StoreUser(ctx, item)
}

func (s Service) Create(ctx context.Context, email, login, password string) (model.User, error) {
	if err := s.check(ctx, email, login, password); err != nil {
		return model.User{}, httpModel.WrapInvalid(err)
	}

	var output model.User

	err := s.store.DoAtomic(ctx, func(ctx context.Context) error {
		loginUser, err := s.auth.CreateBasic(ctx, login, password)
		if err != nil {
			return httpModel.WrapInternal(fmt.Errorf("create user: %w", err))
		}

		user := model.NewUser(0, email, loginUser)

		id, err := s.store.Create(ctx, user)
		if err != nil {
			return httpModel.WrapInternal(fmt.Errorf("create: %w", err))
		}

		user.ID = id
		output = user

		return nil
	})

	return output, err
}

func (s Service) check(ctx context.Context, email, login, password string) error {
	var output []error

	if len(strings.TrimSpace(email)) == 0 {
		output = append(output, errors.New("email is required"))
	}

	if len(strings.TrimSpace(login)) == 0 {
		output = append(output, errors.New("login is required"))
	}

	if len(strings.TrimSpace(password)) == 0 {
		output = append(output, errors.New("password is required"))
	}

	if userWithEmail, err := s.store.GetByEmail(ctx, email); err != nil {
		output = append(output, errors.New("check if email already exists"))
	} else if !userWithEmail.ID.IsZero() {
		output = append(output, errors.New("email already used"))
	}

	return httpModel.ConcatError(output)
}

func (s Service) Count(ctx context.Context) (uint64, error) {
	return s.store.Count(ctx)
}
