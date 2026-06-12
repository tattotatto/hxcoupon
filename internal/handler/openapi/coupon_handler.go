package openapi

import (
	"net/http"
	"strconv"

	"hxcoupon/internal/dto/request"
	"hxcoupon/internal/dto/response"
	"hxcoupon/internal/service"

	"github.com/gin-gonic/gin"
)

type CouponHandler struct {
	couponService *service.CouponService
}

func NewCouponHandler(couponService *service.CouponService) *CouponHandler {
	return &CouponHandler{couponService: couponService}
}

func (h *CouponHandler) Issue(c *gin.Context) {
	storeID, _ := c.Get("store_id")

	var req request.IssueCouponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(40001, err.Error()))
		return
	}

	result, err := h.couponService.Issue(c.Request.Context(), storeID.(uint64), req.TemplateID, req.UserPhone)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Success(result))
}

func (h *CouponHandler) Available(c *gin.Context) {
	var req request.AvailableCouponRequest
	req.UserPhone = c.Query("user_phone")
	storeIDStr := c.Query("store_id")
	orderAmountStr := c.Query("order_amount")

	if req.UserPhone == "" || storeIDStr == "" {
		c.JSON(http.StatusBadRequest, response.Error(40001, "user_phone and store_id required"))
		return
	}

	req.StoreID, _ = strconv.ParseUint(storeIDStr, 10, 64)
	req.OrderAmount, _ = strconv.ParseFloat(orderAmountStr, 64)
	req.Page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	req.PageSize, _ = strconv.Atoi(c.DefaultQuery("page_size", "20"))

	data, err := h.couponService.GetAvailable(c.Request.Context(), req.UserPhone, req.StoreID, req.OrderAmount, req.Page, req.PageSize)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Success(data))
}

func (h *CouponHandler) Consume(c *gin.Context) {
	var req request.ConsumeCouponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(40001, err.Error()))
		return
	}

	result, err := h.couponService.Consume(c.Request.Context(), req.CouponCode, req.UserPhone, req.StoreID, req.OrderID, req.OrderAmount)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Success(result))
}

func (h *CouponHandler) Refund(c *gin.Context) {
	var req request.RefundCouponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(40001, err.Error()))
		return
	}

	result, err := h.couponService.Refund(c.Request.Context(), req.CouponCode, req.UserPhone, req.StoreID, req.OrderID)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Success(result))
}

func (h *CouponHandler) ListByUser(c *gin.Context) {
	userPhone := c.Query("user_phone")
	status := c.DefaultQuery("status", "")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if userPhone == "" {
		c.JSON(http.StatusBadRequest, response.Error(40001, "user_phone required"))
		return
	}

	data, err := h.couponService.ListByUser(c.Request.Context(), userPhone, status, page, pageSize)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Success(data))
}

func (h *CouponHandler) Detail(c *gin.Context) {
	couponCode := c.Param("coupon_code")
	if couponCode == "" {
		c.JSON(http.StatusBadRequest, response.Error(40001, "coupon_code required"))
		return
	}

	result, err := h.couponService.GetDetail(c.Request.Context(), couponCode)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Success(result))
}
