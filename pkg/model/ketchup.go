package model

var (
	// NoneKetchup is an undefined ketchup
	NoneKetchup = Ketchup{}
)

// Ketchup of app
type Ketchup struct {
	Version    string     `json:"version"`
	Repository Repository `json:"repository"`
}
