package model

type UserLoginResponse struct {
	Token string `json:"token"`
}

type UserLogoutResponse struct {
	Message string `json:"message"`
}

type WebsocketWelcomeResponse struct {
	Welcome string `json:"welcome"`
}
