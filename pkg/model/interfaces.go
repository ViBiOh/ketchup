package model

import (
	"context"

	authModel "github.com/ViBiOh/auth/v2/pkg/model"
	"github.com/ViBiOh/ketchup/pkg/semver"
	mailerModel "github.com/ViBiOh/mailer/pkg/model"
)

type Identifier uint64

func (i Identifier) IsZero() bool {
	return i == 0
}

//go:generate mockgen -source interfaces.go -destination ../mocks/interfaces.go -package mocks -mock_names Mailer=Mailer,AuthService=AuthService,UserService=UserService,UserStore=UserStore,GenericProvider=GenericProvider,HelmProvider=HelmProvider,RepositoryService=RepositoryService,RepositoryStore=RepositoryStore,KetchupService=KetchupService,KetchupStore=KetchupStore

type Mailer interface {
	Enabled() bool
	Send(context.Context, mailerModel.MailRequest) error
}

type AuthService interface {
	Create(context.Context, authModel.User) (authModel.User, error)
	Check(context.Context, authModel.User, authModel.User) error
}

type UserService interface {
	StoreInContext(context.Context) context.Context
}

type UserStore interface {
	DoAtomic(context.Context, func(context.Context) error) error
	GetByLoginID(context.Context, uint64) (User, error)
	GetByEmail(context.Context, string) (User, error)
	Create(context.Context, User) (Identifier, error)
	Count(context.Context) (uint64, error)
}

type GenericProvider interface {
	LatestVersions(context.Context, string, []string) (map[string]semver.Version, error)
}

type HelmProvider interface {
	FetchIndex(context.Context, string, map[string][]string) (map[string]map[string]semver.Version, error)
	LatestVersions(context.Context, string, string, []string) (map[string]semver.Version, error)
}

type RepositoryService interface {
	List(context.Context, uint, string) ([]Repository, uint64, error)
	ListByKinds(context.Context, uint, string, ...RepositoryKind) ([]Repository, uint64, error)
	Suggest(context.Context, []Identifier, uint64) ([]Repository, error)
	GetOrCreate(context.Context, RepositoryKind, string, string, string) (Repository, error)
	Update(context.Context, Repository) error
	Clean(context.Context) error
	LatestVersions(context.Context, Repository) (map[string]semver.Version, error)
}

type RepositoryStore interface {
	DoAtomic(ctx context.Context, action func(context.Context) error) error
	List(ctx context.Context, pageSize uint, last string) ([]Repository, uint64, error)
	ListByKinds(ctx context.Context, pageSize uint, last string, kinds ...RepositoryKind) ([]Repository, uint64, error)
	Suggest(ctx context.Context, ignoreIds []Identifier, count uint64) ([]Repository, error)
	Get(ctx context.Context, id Identifier, forUpdate bool) (Repository, error)
	GetByName(ctx context.Context, repositoryKind RepositoryKind, name, part string) (Repository, error)
	Create(ctx context.Context, o Repository) (Identifier, error)
	UpdateVersions(ctx context.Context, o Repository) error
	DeleteUnused(ctx context.Context) error
	DeleteUnusedVersions(ctx context.Context) error
}

type KetchupService interface {
	List(ctx context.Context, pageSize uint, last string) ([]Ketchup, uint64, error)
	ListForRepositories(ctx context.Context, repositories []Repository, frequencies ...KetchupFrequency) ([]Ketchup, error)
	ListOutdated(ctx context.Context, users ...User) ([]Ketchup, error)
	Create(ctx context.Context, item Ketchup) (Ketchup, error)
	Update(ctx context.Context, oldPattern string, item Ketchup) (Ketchup, error)
	UpdateAll(ctx context.Context) error
	UpdateVersion(ctx context.Context, userID, repositoryID Identifier, pattern, version string) error
	Delete(ctx context.Context, item Ketchup) error
}

type KetchupStore interface {
	DoAtomic(ctx context.Context, action func(context.Context) error) error
	List(ctx context.Context, page uint, last string) ([]Ketchup, uint64, error)
	ListByRepositoriesIDAndFrequencies(ctx context.Context, ids []Identifier, frequencies ...KetchupFrequency) ([]Ketchup, error)
	ListOutdated(ctx context.Context, usersIds ...Identifier) ([]Ketchup, error)
	GetByRepository(ctx context.Context, id Identifier, pattern string, forUpdate bool) (Ketchup, error)
	Create(ctx context.Context, o Ketchup) (Identifier, error)
	Update(ctx context.Context, o Ketchup, oldPattern string) error
	UpdateAll(ctx context.Context) error
	UpdateVersion(ctx context.Context, userID, repositoryID Identifier, pattern, version string) error
	Delete(ctx context.Context, o Ketchup) error
}
