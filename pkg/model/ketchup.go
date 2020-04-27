package model

var (
	// NoneKetchup is an undefined ketchup
	NoneKetchup = Ketchup{}
)

// Ketchup of app
type Ketchup struct {
	ID         uint64     `json:"id"`
	Version    string     `json:"version"`
	Repository Repository `json:"repository"`
}
