package gist

// User represents a GitHub authorized user on the system.
type User struct {
	ID          int    `json:"id"`
	Username    string `json:"username"`
	AccessToken string `json:"accessToken"`
}
