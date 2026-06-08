package admin

import (
	"net/http"
	"strconv"

	"hxcoupon/internal/dto/request"
	"hxcoupon/internal/dto/response"
	"hxcoupon/internal/pkg/errcode"
	"hxcoupon/internal/service"

	"github.com/gin-gonic/gin"
)

type CouponHandler struct {
	couponService *service.CouponService
}

func NewCouponHandler(couponService *service.CouponService) *CouponHandler {
	return &CouponHandler{couponService: couponService}
}

// Issue issues a coupon from the admin's specified store.
func (h *CouponHandler) Issue(c *gin.Context) {
	var req request.AdminIssueCouponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(errcode.InvalidParams, err.Error()))
		return
	}

	result, err := h.couponService.Issue(c.Request.Context(), req.StoreID, req.TemplateID, req.UserPhone, req.IdempotencyKey)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Success(result))
}

// ListRecords lists coupon instances issued by the admin's stores.
func (h *CouponHandler) ListRecords(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	data, err := h.couponService.ListAdminRecords(c.Request.Context(), page, pageSize)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Success(data))
}

// GetRecord returns a single coupon instance detail.
func (h *CouponHandler) GetRecord(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error(errcode.InvalidParams, "invalid id"))
		return
	}

	detail, err := h.couponService.GetByID(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Success(detail))
}

// Consume marks a coupon as used (admin-side).
func (h *CouponHandler) Consume(c *gin.Context) {
	var req request.ConsumeCouponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(errcode.InvalidParams, err.Error()))
		return
	}

	result, err := h.couponService.Consume(c.Request.Context(), req.CouponCode, req.UserPhone, req.StoreID, req.OrderID, req.OrderAmount)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Success(result))
}

// ListConsumeRecords lists consume records.
func (h *CouponHandler) ListConsumeRecords(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	data, err := h.couponService.ListConsumeRecords(c.Request.Context(), page, pageSize)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Success(data))
}
