package admin

import (
	"net/http"
	"strconv"

	"hxcoupon/internal/dto/request"
	"hxcoupon/internal/dto/response"
	"hxcoupon/internal/middleware"
	"hxcoupon/internal/pkg/errcode"
	"hxcoupon/internal/service"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	adminUserService *service.AdminUserService
}

func NewUserHandler(adminUserService *service.AdminUserService) *UserHandler {
	return &UserHandler{adminUserService: adminUserService}
}

// List users (super_admin only)
func (h *UserHandler) List(c *gin.Context) {
	var req request.UserListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(errcode.InvalidParams, err.Error()))
		return
	}

	result, err := h.adminUserService.ListUsers(c.Request.Context(), &req)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Success(result))
}

// Get a single user (super_admin only)
func (h *UserHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error(errcode.InvalidParams, "invalid id"))
		return
	}

	user, err := h.adminUserService.GetUser(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Success(user))
}

// ApproveUser approves a pending user
func (h *UserHandler) ApproveUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error(errcode.InvalidParams, "invalid id"))
		return
	}

	var req request.ApproveUserRequest
	c.ShouldBindJSON(&req)

	if err := h.adminUserService.ApproveUser(c.Request.Context(), id, middleware.GetUserID(c), req.Reason); err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Success(gin.H{"message": "user approved"}))
}

// RejectUser rejects a pending user
func (h *UserHandler) RejectUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error(errcode.InvalidParams, "invalid id"))
		return
	}

	var req request.RejectUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(errcode.InvalidParams, err.Error()))
		return
	}

	if err := h.adminUserService.RejectUser(c.Request.Context(), id, middleware.GetUserID(c), req.Reason); err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Success(gin.H{"message": "user rejected"}))
}

// SuspendUser suspends an approved user
func (h *UserHandler) SuspendUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error(errcode.InvalidParams, "invalid id"))
		return
	}

	var req request.SuspendUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(errcode.InvalidParams, err.Error()))
		return
	}

	if err := h.adminUserService.SuspendUser(c.Request.Context(), id, middleware.GetUserID(c), req.Reason); err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Success(gin.H{"message": "user suspended"}))
}

// UnsuspendUser restores a suspended user
func (h *UserHandler) UnsuspendUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error(errcode.InvalidParams, "invalid id"))
		return
	}

	if err := h.adminUserService.UnsuspendUser(c.Request.Context(), id, middleware.GetUserID(c)); err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Success(gin.H{"message": "user unsuspended"}))
}

// GetProfile returns the current user's profile
func (h *UserHandler) GetProfile(c *gin.Context) {
	profile, err := h.adminUserService.GetProfile(c.Request.Context(), middleware.GetUserID(c))
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Success(profile))
}

// UpdateProfile updates the current user's profile
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	var req request.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(errcode.InvalidParams, err.Error()))
		return
	}

	if err := h.adminUserService.UpdateProfile(c.Request.Context(), middleware.GetUserID(c), &req); err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Success(gin.H{"message": "profile updated"}))
}
