package model

// RowScanner describes scan ability of a row
type RowScanner interface {
	Scan(...interface{}) error
}
