package model

var (
	// NoneRepository is an undefined repository
	NoneRepository = Repository{}
)

// Repository of app
type Repository struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	ID      uint64 `json:"id"`
}
