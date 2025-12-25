package cap

import (
	"context"
	"flag"
	"fmt"

	"github.com/ViBiOh/flags"
	"github.com/ViBiOh/httputils/v4/pkg/httpjson"
	"github.com/ViBiOh/httputils/v4/pkg/request"
)

type Service struct {
	siteURL   string
	secretKey string
}

type Config struct {
	URL       string
	SiteKey   string
	SecretKey string
}

func Flags(fs *flag.FlagSet, prefix string, overrides ...flags.Override) *Config {
	var config Config

	flags.New("URL", "Instance URL").Prefix(prefix).DocPrefix("cap").StringVar(fs, &config.URL, "http://cap", overrides)
	flags.New("SiteKey", "Site Key").Prefix(prefix).DocPrefix("cap").StringVar(fs, &config.SiteKey, "", overrides)
	flags.New("SecretKey", "Secret Key").Prefix(prefix).DocPrefix("cap").StringVar(fs, &config.SecretKey, "", overrides)

	return &config
}

func New(config *Config) Service {
	return Service{
		siteURL:   fmt.Sprintf("%s/%s/", config.URL, config.SiteKey),
		secretKey: config.SecretKey,
	}
}

func (s Service) SiteURL() string {
	return s.siteURL
}

func (s Service) Verify(ctx context.Context, token string) (bool, error) {
	resp, err := request.New().Post(s.siteURL+"siteverify").JSON(ctx, map[string]string{
		"secret":   s.secretKey,
		"response": token,
	})
	if err != nil {
		return false, fmt.Errorf("send: %w", err)
	}

	answer, err := httpjson.Read[VerifyAnswer](resp)
	if err != nil {
		return false, fmt.Errorf("json: %w", err)
	}

	return answer.Success, nil
}
