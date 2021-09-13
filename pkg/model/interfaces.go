package model

import (
	"context"

	authModel "github.com/ViBiOh/auth/v2/pkg/model"
	"github.com/ViBiOh/ketchup/pkg/semver"
	mailerModel "github.com/ViBiOh/mailer/pkg/model"
)

// Mailer interface client
type Mailer interface {
	Enabled() bool
	Send(context.Context, mailerModel.MailRequest) error
}

// AuthService defines interaction with underlying User provider
type AuthService interface {
	Create(context.Context, authModel.User) (authModel.User, error)
	Check(context.Context, authModel.User, authModel.User) error
}

// UserService for storing user in context
type UserService interface {
	StoreInContext(context.Context) context.Context
}

// UserStore defines interaction with User storage
type UserStore interface {
	DoAtomic(context.Context, func(context.Context) error) error
	GetByLoginID(context.Context, uint64) (User, error)
	GetByEmail(context.Context, string) (User, error)
	Create(context.Context, User) (uint64, error)
	Count(context.Context) (uint64, error)
}

// GenericProvider defines interaction with common providers
type GenericProvider interface {
	LatestVersions(string, []string) (map[string]semver.Version, error)
}

// HelmProvider defines interaction with helm
type HelmProvider interface {
	FetchIndex(string, map[string][]string) (map[string]map[string]semver.Version, error)
	LatestVersions(string, string, []string) (map[string]semver.Version, error)
}

// RepositoryService defines interaction with repository
type RepositoryService interface {
	List(context.Context, uint, string) ([]Repository, uint64, error)
	ListByKinds(context.Context, uint, string, ...RepositoryKind) ([]Repository, uint64, error)
	Suggest(context.Context, []uint64, uint64) ([]Repository, error)
	GetOrCreate(context.Context, RepositoryKind, string, string, string) (Repository, error)
	Update(context.Context, Repository) error
	Clean(context.Context) error
	LatestVersions(Repository) (map[string]semver.Version, error)
}

// KetchupService defines interaction with ketchup
type KetchupService interface {
	List(ctx context.Context, pageSize uint, last string) ([]Ketchup, uint64, error)
	ListForRepositories(ctx context.Context, repositories []Repository, frequency KetchupFrequency) ([]Ketchup, error)
	ListOutdatedByFrequency(ctx context.Context, frequency KetchupFrequency) ([]Ketchup, error)
	Create(ctx context.Context, item Ketchup) (Ketchup, error)
	Update(ctx context.Context, oldPattern string, item Ketchup) (Ketchup, error)
	UpdateAll(ctx context.Context) error
	Delete(ctx context.Context, item Ketchup) error
}
