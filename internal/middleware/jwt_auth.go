package middleware

import (
	"net/http"
	"strings"

	"hxcoupon/internal/pkg/apperror"
	"hxcoupon/internal/pkg/errcode"
	"hxcoupon/internal/service"

	"github.com/gin-gonic/gin"
)

func JWTAuth(authService *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    errcode.AuthFailed,
				"message": "missing authorization header",
			})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    errcode.AuthFailed,
				"message": "invalid authorization format",
			})
			return
		}

		claims, err := authService.ParseToken(parts[1])
		if err != nil {
			appErr, ok := err.(*apperror.AppError)
			status := http.StatusUnauthorized
			if ok {
				status = http.StatusUnauthorized
			}
			c.AbortWithStatusJSON(status, gin.H{
				"code":    appErr.Code,
				"message": appErr.Message,
			})
			return
		}

		c.Set("admin_user_id", claims.UserID)
		c.Set("admin_username", claims.Username)
		c.Set("admin_role", claims.Role)
		c.Next()
	}
}
