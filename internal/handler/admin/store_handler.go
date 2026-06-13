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

type StoreHandler struct {
	storeService *service.StoreService
}

func NewStoreHandler(storeService *service.StoreService) *StoreHandler {
	return &StoreHandler{storeService: storeService}
}

func (h *StoreHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	// Members only see their own stores
	role, _ := c.Get("admin_role")
	if role == "member" {
		userID, _ := c.Get("admin_user_id")
		data, err := h.storeService.ListByUser(c.Request.Context(), userID.(uint64), page, pageSize)
		if err != nil {
			handleError(c, err)
			return
		}
		c.JSON(http.StatusOK, response.Success(data))
		return
	}

	data, err := h.storeService.List(c.Request.Context(), page, pageSize)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Success(data))
}

// guardOwnership aborts if the current user is a member and does not own the store.
func (h *StoreHandler) guardOwnership(c *gin.Context, storeID uint64) bool {
	role, _ := c.Get("admin_role")
	if role != "member" {
		return true
	}
	userID, _ := c.Get("admin_user_id")
	if err := h.storeService.VerifyStoreOwnership(c.Request.Context(), storeID, userID.(uint64)); err != nil {
		c.JSON(http.StatusForbidden, response.Error(40300, "access denied: not your store"))
		return false
	}
	return true
}

func (h *StoreHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error(errcode.InvalidParams, "invalid id"))
		return
	}

	if !h.guardOwnership(c, id) {
		return
	}

	store, err := h.storeService.GetByID(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Success(store))
}

func (h *StoreHandler) Create(c *gin.Context) {
	var req request.CreateStoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(errcode.InvalidParams, err.Error()))
		return
	}

	// Associate the store with the current user if member
	role, _ := c.Get("admin_role")
	var userID *uint64
	if role == "member" {
		uid, _ := c.Get("admin_user_id")
		id := uid.(uint64)
		userID = &id
	}

	storeResp, err := h.storeService.Create(c.Request.Context(), &req, userID)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, response.Success(storeResp))
}

func (h *StoreHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error(errcode.InvalidParams, "invalid id"))
		return
	}

	if !h.guardOwnership(c, id) {
		return
	}

	var req request.UpdateStoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(errcode.InvalidParams, err.Error()))
		return
	}

	storeResp, err := h.storeService.Update(c.Request.Context(), id, &req)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Success(storeResp))
}

func (h *StoreHandler) UpdateStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error(errcode.InvalidParams, "invalid id"))
		return
	}

	if !h.guardOwnership(c, id) {
		return
	}

	var req request.UpdateStoreStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(errcode.InvalidParams, err.Error()))
		return
	}

	if err := h.storeService.UpdateStatus(c.Request.Context(), id, req.Status); err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Success(nil))
}

func (h *StoreHandler) GenerateCredentials(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error(errcode.InvalidParams, "invalid id"))
		return
	}

	if !h.guardOwnership(c, id) {
		return
	}

	cred, err := h.storeService.GenerateCredentials(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Success(cred))
}

// Options returns all active stores as a simplified list for dropdown selects.
func (h *StoreHandler) Options(c *gin.Context) {
	stores, err := h.storeService.ListActive(c.Request.Context())
	if err != nil {
		handleError(c, err)
		return
	}

	type storeOption struct {
		ID   uint64 `json:"id"`
		Name string `json:"name"`
		Code string `json:"code"`
	}
	items := make([]storeOption, len(stores))
	for i, s := range stores {
		items[i] = storeOption{ID: s.ID, Name: s.Name, Code: s.Code}
	}
	c.JSON(http.StatusOK, response.Success(items))
}

// DeleteStore soft-deletes a store.
func (h *StoreHandler) DeleteStore(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error(errcode.InvalidParams, "invalid id"))
		return
	}

	if !h.guardOwnership(c, id) {
		return
	}

	if err := h.storeService.Delete(c.Request.Context(), id); err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Success(nil))
}
