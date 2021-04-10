package redis

import (
	"context"
	"flag"
	"strings"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/flags"
	"github.com/ViBiOh/httputils/v4/pkg/logger"
	"github.com/go-redis/redis/v8"
)

// App of package
type App interface {
	Store(context.Context, string, string, time.Duration) error
	Load(context.Context, string) (string, error)
	Delete(context.Context, string) error
	DoExclusive(context.Context, string, time.Duration, func(context.Context) error) (bool, error)
	Ping() error
}

// Config of package
type Config struct {
	redisAddress  *string
	redisPassword *string
	redisDatabase *int
}

type app struct {
	redisClient *redis.Client
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string, overrides ...flags.Override) Config {
	return Config{
		redisAddress:  flags.New(prefix, "redis").Name("Address").Default(flags.Default("Address", "localhost:6379", overrides)).Label("Redis Address").ToString(fs),
		redisPassword: flags.New(prefix, "redis").Name("Password").Default(flags.Default("Password", "", overrides)).Label("Redis Password, if any").ToString(fs),
		redisDatabase: flags.New(prefix, "redis").Name("Database").Default(flags.Default("Database", 0, overrides)).Label("Redis Database").ToInt(fs),
	}
}

// New creates new App from Config
func New(config Config) App {
	return app{
		redisClient: redis.NewClient(&redis.Options{
			Addr:     strings.TrimSpace(*config.redisAddress),
			Password: strings.TrimSpace(*config.redisPassword),
			DB:       *config.redisDatabase,
		}),
	}
}

func (a app) Ping() error {
	return a.redisClient.Ping(context.Background()).Err()
}

func (a app) Store(ctx context.Context, key, value string, duration time.Duration) error {
	return a.redisClient.SetEX(ctx, key, value, duration).Err()
}

func (a app) Load(ctx context.Context, key string) (string, error) {
	return a.redisClient.Get(ctx, key).Result()
}

func (a app) Delete(ctx context.Context, key string) error {
	return a.redisClient.Del(ctx, key).Err()
}

func (a app) DoExclusive(ctx context.Context, lockName string, timeout time.Duration, action func(context.Context) error) (bool, error) {
	if !a.redisClient.SetNX(ctx, lockName, "acquired", timeout).Val() {
		return false, nil
	}

	defer func() {
		if err := a.redisClient.Del(ctx, lockName).Err(); err != nil {
			logger.Warn("unable to release lock for `%s`: %s", lockName, err)
		}
	}()

	actionCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return true, action(actionCtx)
}
