package token

import (
	"context"
	"crypto/rand"
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/flags"
	"github.com/ViBiOh/httputils/v4/pkg/logger"
	"github.com/go-redis/redis/v8"
)

// App of package
type App interface {
	Store(context.Context, string, time.Duration) (string, error)
	Load(context.Context, string) (string, error)
	Delete(context.Context, string) error
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
		redisAddress:  flags.New(prefix, "token").Name("RedisAddress").Default(flags.Default("RedisAddress", "localhost:6379", overrides)).Label("Redis Address").ToString(fs),
		redisPassword: flags.New(prefix, "token").Name("RedisPassword").Default(flags.Default("RedisPassword", "", overrides)).Label("Redis Password, if any").ToString(fs),
		redisDatabase: flags.New(prefix, "token").Name("RedisDatabase").Default(flags.Default("RedisDatabase", 0, overrides)).Label("Redis Database").ToInt(fs),
	}
}

// New creates new App from Config
func New(config Config) App {
	address := strings.TrimSpace(*config.redisAddress)

	return app{
		redisClient: redis.NewClient(&redis.Options{
			Addr:     address,
			Password: strings.TrimSpace(*config.redisPassword),
			DB:       *config.redisDatabase,
		}),
	}
}

func (a app) Store(ctx context.Context, value string, duration time.Duration) (string, error) {
	token := uuid()

	return token, a.redisClient.SetNX(ctx, token, value, duration).Err()
}

func (a app) Load(ctx context.Context, key string) (string, error) {
	value, err := a.redisClient.Get(ctx, key).Result()
	if err != nil {
		return "", err
	}

	return value, nil
}

func (a app) Delete(ctx context.Context, key string) error {
	return a.redisClient.Del(ctx, key).Err()
}

func uuid() string {
	raw := make([]byte, 16)
	if _, err := rand.Read(raw); err != nil {
		logger.Fatal(err)
		return ""
	}

	raw[8] = raw[8]&^0xc0 | 0x80
	raw[6] = raw[6]&^0xf0 | 0x40

	return fmt.Sprintf("%x-%x-%x-%x-%x", raw[0:4], raw[4:6], raw[6:8], raw[8:10], raw[10:])
}
