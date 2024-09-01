package model

type UserLoginRequest struct {
	Username string `json:"username"`
}

type UserLogoutRequest struct {
	Username string `json:"username"`
	Token    string `json:"token"`
}
