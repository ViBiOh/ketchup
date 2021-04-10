package ketchup

import (
	"crypto/rand"
	"fmt"

	"github.com/ViBiOh/httputils/v4/pkg/logger"
)

// SecurityQuestion is a question for fighting against bot
type SecurityQuestion struct {
	Question string
	Answer   string
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
