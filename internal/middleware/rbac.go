package middleware

import (
	"net/http"

	"hxcoupon/internal/dto/response"
	"hxcoupon/internal/model"

	"github.com/gin-gonic/gin"
)

// RequirePermission checks the user has the required permission.
// super_admin always passes. Members must have the permission or be "both" type.
func RequirePermission(perm string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get("admin_role")
		if role == "super_admin" {
			c.Next()
			return
		}

		memberType, _ := c.Get("admin_member_type")
		mt, _ := memberType.(*string)

		if !hasMemberPermission(mt, perm) {
			c.AbortWithStatusJSON(http.StatusForbidden, response.Error(40300, "权限不足"))
			return
		}
		c.Next()
	}
}

func hasMemberPermission(memberType *string, perm string) bool {
	if memberType == nil {
		return false
	}
	switch *memberType {
	case model.MemberTypeBoth:
		return true
	case model.MemberTypeIssuer:
		return perm == model.PermIssueCoupons || perm == model.PermManageStores
	case model.MemberTypeConsumer:
		return perm == model.PermConsumeCoupons || perm == model.PermManageTemplates || perm == model.PermManageStores || perm == model.PermViewReports
	default:
		return false
	}
}

// GetUserID extracts the authenticated user ID from context
func GetUserID(c *gin.Context) uint64 {
	id, _ := c.Get("admin_user_id")
	return id.(uint64)
}

// GetUserRole extracts the authenticated user role from context
func GetUserRole(c *gin.Context) string {
	role, _ := c.Get("admin_role")
	return role.(string)
}

// GetMemberType extracts the member type from context
func GetMemberType(c *gin.Context) *string {
	mt, _ := c.Get("admin_member_type")
	if mt == nil {
		return nil
	}
	return mt.(*string)
}

// GetApprovalStatus extracts approval status from context
func GetApprovalStatus(c *gin.Context) int8 {
	as, _ := c.Get("admin_approval_status")
	return as.(int8)
}
