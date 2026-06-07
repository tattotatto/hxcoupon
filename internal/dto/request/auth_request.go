package request

type AdminLoginRequest struct {
	Username string `json:"username" validate:"required,max=64"`
	Password string `json:"password" validate:"required,max=128"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}
