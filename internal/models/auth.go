package models

type AuthResponse struct {
	AuthTokensResponse
	User UserResponse `json:"user"`
}

type AuthTokensResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}
