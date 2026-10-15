package responses

type Login struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

type Identity struct {
	UserID string `json:"user_id"`
}
