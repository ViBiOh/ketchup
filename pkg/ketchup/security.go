package ketchup

import (
	"context"
	"crypto/rand"
	"fmt"
	"log/slog"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/id"
)

type securityQuestion struct {
	Question string
	Answer   string
}

type securityPayload struct {
	Token    string `json:"token"`
	Question string `json:"question"`
}

var colors = map[int64]securityQuestion{
	0:  {"❄️ and 🧻 are both?", "white"},
	1:  {"☀️ and 🍋 are both?", "yellow"},
	2:  {"🥦 and 🥝 are both?", "green"},
	3:  {"👖 and 💧 are both?", "blue"},
	4:  {"🍅 and 🩸 are both?", "red"},
	5:  {"🍆 and ☂️ are both?", "purple"},
	6:  {"🦓 and 🕳 are both?", "black"},
	7:  {"🍊 and 🥕 are both?", "orange"},
	8:  {"🟥 mixed with 🟦 give?", "purple"},
	9:  {"🟨 mixed with 🟥 give?", "orange"},
	10: {"🟦 mixed with 🟨 give?", "green"},
}

func (s Service) generateToken(ctx context.Context) (securityPayload, error) {
	questionID, err := rand.Int(rand.Reader, big.NewInt(int64(len(colors))))
	if err != nil {
		return securityPayload{}, fmt.Errorf("generate random int: %w", err)
	}

	token := id.New()

	id := questionID.Int64()

	if err := s.redis.Store(ctx, tokenKey(token), fmt.Sprintf("%d", id), time.Minute*5); err != nil {
		return securityPayload{}, fmt.Errorf("store token: %w", err)
	}

	return securityPayload{
		Token:    token,
		Question: colors[id].Question,
	}, nil
}

func (s Service) validateToken(ctx context.Context, token, answer string) bool {
	questionIDString, err := s.redis.Load(ctx, tokenKey(token))
	if err != nil {
		slog.LogAttrs(ctx, slog.LevelWarn, "retrieve captcha token", slog.Any("error", err))
		return false
	}

	questionID, err := strconv.ParseInt(string(questionIDString), 10, 64)
	if err != nil {
		slog.LogAttrs(ctx, slog.LevelError, "question id is not numerical", slog.Any("error", err))
		return false
	}

	if colors[questionID].Answer != strings.TrimSpace(strings.ToLower(answer)) {
		slog.LogAttrs(ctx, slog.LevelWarn, "invalid question answer", slog.String("answer", answer))
		return false
	}

	return true
}

func (s Service) cleanToken(ctx context.Context, token string) {
	if err := s.redis.Delete(ctx, tokenKey(token)); err != nil {
		slog.LogAttrs(ctx, slog.LevelError, "delete token", slog.String("token", token), slog.Any("error", err))
	}
}

func tokenKey(token string) string {
	return fmt.Sprintf("%s:token:%s", cachePrefix, token)
}
