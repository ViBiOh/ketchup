package renderer

// SecurityQuestion is a question for fighting against bot
type SecurityQuestion struct {
	Question string
	Answer   string
}

var (
	colors = map[int64]SecurityQuestion{
		0: {"❄️ and 🧻 are both ?", "white"},
		1: {"☀️ and 🍋 are both ?", "yellow"},
		2: {"🥦 and 🥝 are both ?", "green"},
		3: {"👖 and 💧 are both ?", "blue"},
		4: {"🍅 and 🩸 are both ?", "red"},
		5: {"🍆 and ☂️ are both ?", "purple"},
		6: {"🦓 and 🕳 are both ?", "black"},
		7: {"🍊 and 🥕 are both ?", "orange"},
	}
)
