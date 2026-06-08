package middleware

import (
	"net/http"

	"hxcoupon/internal/dto/response"
	"hxcoupon/internal/model"

	"github.com/gin-gonic/gin"
)

// RequireApproval checks that the current user has been approved (or is super_admin).
// This should be applied to all member-facing routes to block pending/rejected/suspended users.
func RequireApproval() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get("admin_role")
		if role == "super_admin" || role == "admin" {
			c.Next()
			return
		}

		as := GetApprovalStatus(c)
		if as != model.ApprovalApproved {
			var msg string
			switch as {
			case model.ApprovalPending:
				msg = "account pending approval"
			case model.ApprovalRejected:
				msg = "account has been rejected"
			case model.ApprovalSuspended:
				msg = "account has been suspended"
			default:
				msg = "account not approved"
			}
			c.AbortWithStatusJSON(http.StatusForbidden, response.Error(40300, msg))
			return
		}
		c.Next()
	}
}
