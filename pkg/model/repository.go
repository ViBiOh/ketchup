package model

var (
	// NoneRepository is an undefined repository
	NoneRepository = Repository{}
)

// Repository of app
type Repository struct {
	ID      uint64 `json:"id"`
	Name    string `json:"name"`
	Version string `json:"version"`
}
