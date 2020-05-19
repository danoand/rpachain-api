package models

// Account models an application account
type Account struct {
	ID       string `json:"id"`       // request id
	Username string `json:"username"` // timestamp
	Password string `json:"password"` // general meta data
}
