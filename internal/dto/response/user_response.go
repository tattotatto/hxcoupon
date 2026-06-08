package response

import "time"

// UserResponse for admin user management
type UserResponse struct {
	ID              uint64     `json:"id"`
	Username        string     `json:"username"`
	Role            string     `json:"role"`
	MemberType      *string    `json:"member_type"`
	Status          int8       `json:"status"`
	ApprovalStatus  int8       `json:"approval_status"`
	CompanyName     *string    `json:"company_name"`
	ContactName     *string    `json:"contact_name"`
	ContactPhone    *string    `json:"contact_phone"`
	Email           *string    `json:"email"`
	RejectReason    *string    `json:"reject_reason"`
	BusinessLicense *string    `json:"business_license"`
	RegisteredAt    *time.Time `json:"registered_at"`
	ApprovedAt      *time.Time `json:"approved_at"`
	LastLoginAt     *time.Time `json:"last_login_at"`
	CreatedAt       time.Time  `json:"created_at"`
}

// ProfileResponse for the current user's own profile
type ProfileResponse struct {
	ID             uint64  `json:"id"`
	Username       string  `json:"username"`
	Role           string  `json:"role"`
	MemberType     *string `json:"member_type"`
	ApprovalStatus int8    `json:"approval_status"`
	CompanyName    *string `json:"company_name"`
	ContactName    *string `json:"contact_name"`
	ContactPhone   *string `json:"contact_phone"`
	Email          *string `json:"email"`
	RejectReason   *string `json:"reject_reason"`
}

// ApprovalRecordResponse for audit log display
type ApprovalRecordResponse struct {
	ID         uint64    `json:"id"`
	UserID     uint64    `json:"user_id"`
	Action     string    `json:"action"`
	Reason     *string   `json:"reason"`
	OperatedBy uint64    `json:"operated_by"`
	CreatedAt  time.Time `json:"created_at"`
}
