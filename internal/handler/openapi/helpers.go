package openapi

import (
	"net/http"

	"hxcoupon/internal/pkg/apperror"
	"hxcoupon/internal/pkg/errcode"

	"github.com/gin-gonic/gin"
)

func handleError(c *gin.Context, err error) {
	appErr, ok := err.(*apperror.AppError)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    errcode.InternalError,
			"message": err.Error(),
		})
		return
	}

	status := http.StatusBadRequest
	switch appErr.Code {
	case errcode.NotFound:
		status = http.StatusNotFound
	case errcode.AuthFailed, errcode.StoreNotAuth:
		status = http.StatusUnauthorized
	case errcode.Forbidden:
		status = http.StatusForbidden
	case errcode.Conflict:
		status = http.StatusConflict
	case errcode.CouponExpired, errcode.CouponUsed, errcode.CouponNotApply, errcode.NoInventory, errcode.PerUserLimit, errcode.BelowThreshold, errcode.RefundMismatch:
		status = http.StatusOK
	}

	c.JSON(status, gin.H{
		"code":    appErr.Code,
		"message": appErr.Message,
	})
}
