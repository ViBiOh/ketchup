package user

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	authModel "github.com/ViBiOh/auth/v2/pkg/model"
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
		slog.ErrorContext(ctx, "get user with login", "error", err, "id", id)
		return ctx
	}

	return model.StoreUser(ctx, item)
}

func (s Service) ListReminderUsers(ctx context.Context) ([]model.User, error) {
	return s.store.ListReminderUsers(ctx)
}

func (s Service) Create(ctx context.Context, item model.User) (model.User, error) {
	if err := s.check(ctx, model.User{}, item); err != nil {
		return model.User{}, httpModel.WrapInvalid(err)
	}

	if err := s.auth.Check(ctx, authModel.User{}, item.Login); err != nil {
		return model.User{}, httpModel.WrapInvalid(err)
	}

	var output model.User

	err := s.store.DoAtomic(ctx, func(ctx context.Context) error {
		loginUser, err := s.auth.Create(ctx, item.Login)
		if err != nil {
			return httpModel.WrapInternal(fmt.Errorf("create login: %w", err))
		}

		item.Login = loginUser

		id, err := s.store.Create(ctx, item)
		if err != nil {
			return httpModel.WrapInternal(fmt.Errorf("create: %w", err))
		}

		item.ID = id
		output = item

		return nil
	})

	return output, err
}

func (s Service) check(ctx context.Context, _, new model.User) error {
	if new.IsZero() {
		return nil
	}

	var output []error

	if len(strings.TrimSpace(new.Email)) == 0 {
		output = append(output, errors.New("email is required"))
	}

	if userWithEmail, err := s.store.GetByEmail(ctx, new.Email); err != nil {
		output = append(output, errors.New("check if email already exists"))
	} else if !userWithEmail.ID.IsZero() {
		output = append(output, errors.New("email already used"))
	}

	return httpModel.ConcatError(output)
}

func (s Service) Count(ctx context.Context) (uint64, error) {
	return s.store.Count(ctx)
}
