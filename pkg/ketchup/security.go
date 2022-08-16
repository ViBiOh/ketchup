package ketchup

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/logger"
	"github.com/ViBiOh/httputils/v4/pkg/uuid"
)

type securityQuestion struct {
	Question string
	Answer   string
}

type securityPayload struct {
	UUID     string `json:"uuid"`
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

func (a App) generateToken(ctx context.Context) (securityPayload, error) {
	questionID, err := rand.Int(rand.Reader, big.NewInt(int64(len(colors))))
	if err != nil {
		return securityPayload{}, fmt.Errorf("generate random int: %w", err)
	}

	token, err := uuid.New()
	if err != nil {
		return securityPayload{}, fmt.Errorf("generate uuid: %w", err)
	}

	id := questionID.Int64()

	if err := a.redisApp.Store(ctx, tokenKey(token), fmt.Sprintf("%d", id), time.Minute*5); err != nil {
		return securityPayload{}, fmt.Errorf("store token: %w", err)
	}

	return securityPayload{
		UUID:     token,
		Question: colors[id].Question,
	}, nil
}

func (a App) validateToken(ctx context.Context, token, answer string) bool {
	questionIDString, err := a.redisApp.Load(ctx, tokenKey(token))
	if err != nil {
		logger.Warn("retrieve captcha token: %s", err)
		return false
	}

	questionID, err := strconv.ParseInt(questionIDString, 10, 64)
	if err != nil {
		logger.Error("question id is not numerical: %s", err)
		return false
	}

	if colors[questionID].Answer != strings.TrimSpace(answer) {
		logger.WithField("answer", answer).Warn("invalid question answer")
		return false
	}

	return true
}

func (a App) cleanToken(ctx context.Context, token string) {
	if err := a.redisApp.Delete(ctx, tokenKey(token)); err != nil {
		logger.WithField("token", token).Error("delete token: %s", err)
	}
}

func tokenKey(token string) string {
	return fmt.Sprintf("%s:token:%s", cachePrefix(), token)
}
