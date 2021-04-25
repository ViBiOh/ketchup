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
)

// SecurityQuestion is a question for fighting against bot
type SecurityQuestion struct {
	Question string
	Answer   string
}

// SecurityPayload is a question sent to the browser
type SecurityPayload struct {
	UUID     string `json:"uuid"`
	Question string `json:"question"`
}

var (
	colors = map[int64]SecurityQuestion{
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
)

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

func (a app) generateToken(ctx context.Context) (SecurityPayload, error) {
	questionID, err := rand.Int(rand.Reader, big.NewInt(int64(len(colors))))
	if err != nil {
		return SecurityPayload{}, fmt.Errorf("unable to generate random int: %w", err)
	}

	token := uuid()
	id := questionID.Int64()

	if err := a.redisApp.Store(ctx, token, fmt.Sprintf("%d", id), time.Minute*5); err != nil {
		return SecurityPayload{}, fmt.Errorf("unable to store token: %s", err)
	}

	return SecurityPayload{
		UUID:     token,
		Question: colors[id].Question,
	}, nil
}

func (a app) validateToken(ctx context.Context, token, answer string) bool {
	questionIDString, err := a.redisApp.Load(ctx, token)
	if err != nil {
		logger.Warn("unable to retrieve captcha token: %s", err)
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

func (a app) cleanToken(ctx context.Context, token string) {
	if err := a.redisApp.Delete(ctx, token); err != nil {
		logger.WithField("token", token).Error("unable to delete token: %s", err)
	}
}
