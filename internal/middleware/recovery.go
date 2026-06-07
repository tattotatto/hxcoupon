package middleware

import (
	"net/http"

	"hxcoupon/internal/dto/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Recovery(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("panic recovered", zap.Any("error", err))
				c.AbortWithStatusJSON(http.StatusInternalServerError, response.Error(50000, "internal server error"))
			}
		}()
		c.Next()
	}
}
