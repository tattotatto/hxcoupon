package admin

import (
	"net/http"
	"strconv"

	"hxcoupon/internal/dto/request"
	"hxcoupon/internal/dto/response"
	"hxcoupon/internal/pkg/apperror"
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

	data, err := h.storeService.List(c.Request.Context(), page, pageSize)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Success(data))
}

func (h *StoreHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error(errcode.InvalidParams, "invalid id"))
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

	storeResp, err := h.storeService.Create(c.Request.Context(), &req)
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

	cred, err := h.storeService.GenerateCredentials(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Success(cred))
}
