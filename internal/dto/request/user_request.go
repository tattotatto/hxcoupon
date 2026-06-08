package request

// UserListRequest for listing platform users
type UserListRequest struct {
	Page           int    `form:"page" validate:"min=1"`
	PageSize       int    `form:"page_size" validate:"min=1,max=100"`
	Keyword        string `form:"keyword"`
	Role           string `form:"role"`
	MemberType     string `form:"member_type"`
	ApprovalStatus *int   `form:"approval_status"`
}

// ApproveUserRequest for approving a pending user
type ApproveUserRequest struct {
	Reason *string `json:"reason"`
}

// RejectUserRequest for rejecting a pending user
type RejectUserRequest struct {
	Reason string `json:"reason" validate:"required,max=512"`
}

// SuspendUserRequest for suspending an approved user
type SuspendUserRequest struct {
	Reason string `json:"reason" validate:"required,max=512"`
}

// UpdateProfileRequest for users updating their own profile
type UpdateProfileRequest struct {
	ContactName  *string `json:"contact_name" validate:"omitempty,max=64"`
	ContactPhone *string `json:"contact_phone" validate:"omitempty,max=20"`
	Email        *string `json:"email" validate:"omitempty,max=128"`
	CompanyName  *string `json:"company_name" validate:"omitempty,max=128"`
	OldPassword  *string `json:"old_password"`
	NewPassword  *string `json:"new_password" validate:"omitempty,min=6,max=128"`
}
