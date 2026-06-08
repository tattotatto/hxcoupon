package request

type AdminLoginRequest struct {
	Username string `json:"username" validate:"required,max=64"`
	Password string `json:"password" validate:"required,max=128"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// RegisterRequest for self-registration
type RegisterRequest struct {
	Username        string `json:"username" validate:"required,min=3,max=64"`
	Password        string `json:"password" validate:"required,min=6,max=128"`
	MemberType      string `json:"member_type" validate:"required,oneof=issuer consumer both"`
	CompanyName     string `json:"company_name" validate:"required,max=128"`
	ContactName     string `json:"contact_name" validate:"required,max=64"`
	ContactPhone    string `json:"contact_phone" validate:"required,max=20"`
	Email           string `json:"email" validate:"max=128"`
	BusinessLicense string `json:"business_license" validate:"max=256"`
}
