package target

// Target is a watched repository for checking update
type Target struct {
	ID         uint64 `json:"id"`
	Owner      string `json:"owner"`
	Repository string `json:"repository"`
	Version    string `json:"version"`
}
