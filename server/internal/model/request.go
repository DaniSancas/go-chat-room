package model

type UserLoginRequest struct {
	Username string `json:"username"`
}

type UserWithTokenRequest struct {
	Username string `json:"username"`
	Token    string `json:"token"`
}
