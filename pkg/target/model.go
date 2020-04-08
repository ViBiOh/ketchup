package target

// Target is a watched repository for checking update
type Target struct {
	ID             uint64 `json:"id"`
	Repository     string `json:"repository"`
	CurrentVersion string `json:"current_version"`
	LatestVersion  string `json:"latest_version"`
}
