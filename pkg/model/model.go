package model

// RowScanner describes scan ability of a row
type RowScanner interface {
	Scan(...interface{}) error
}

// Message for render
type Message struct {
	Level   string
	Content string
}
