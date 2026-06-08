package model

// Permission constants
const (
	PermIssueCoupons    = "issue_coupons"
	PermConsumeCoupons  = "consume_coupons"
	PermManageStores    = "manage_stores"
	PermManageTemplates = "manage_templates"
	PermViewReports     = "view_reports"
	PermExportData      = "export_data"
	PermManageUsers     = "manage_users"
)

// Approval status constants
const (
	ApprovalPending   int8 = 0
	ApprovalApproved  int8 = 1
	ApprovalRejected  int8 = 2
	ApprovalSuspended int8 = 3
)

// Member type constants
const (
	MemberTypeIssuer   = "issuer"
	MemberTypeConsumer = "consumer"
	MemberTypeBoth     = "both"
)

// IsSuperAdmin checks if a role is super_admin
func IsSuperAdmin(role string) bool { return role == "super_admin" }

// IsMember checks if a role is a registered member
func IsMember(role string) bool { return role == "member" }

// IsApproved checks if a member is approved to use the platform
func IsApproved(role string, approvalStatus int8) bool {
	if role == "super_admin" || role == "admin" {
		return true
	}
	return approvalStatus == ApprovalApproved
}
