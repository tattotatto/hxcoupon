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

type TemplateHandler struct {
	templateService *service.TemplateService
}

func NewTemplateHandler(templateService *service.TemplateService) *TemplateHandler {
	return &TemplateHandler{templateService: templateService}
}

func (h *TemplateHandler) List(c *gin.Context) {
	var req request.TemplateListRequest
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	req.Page = page
	req.PageSize = pageSize
	req.Keyword = c.Query("keyword")
	req.Type = c.Query("type")
	if statusStr := c.Query("status"); statusStr != "" {
		statusVal, _ := strconv.Atoi(statusStr)
		s := int8(statusVal)
		req.Status = &s
	}

	data, err := h.templateService.List(c.Request.Context(), req)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Success(data))
}

func (h *TemplateHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error(errcode.InvalidParams, "invalid id"))
		return
	}

	t, err := h.templateService.GetByID(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Success(t))
}

func (h *TemplateHandler) Create(c *gin.Context) {
	var req request.CreateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(errcode.InvalidParams, err.Error()))
		return
	}

	username, _ := c.Get("admin_username")
	createdBy := ""
	if username != nil {
		createdBy = username.(string)
	}

	t, err := h.templateService.Create(c.Request.Context(), &req, createdBy)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, response.Success(t))
}

func (h *TemplateHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error(errcode.InvalidParams, "invalid id"))
		return
	}

	var req request.UpdateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(errcode.InvalidParams, err.Error()))
		return
	}

	t, err := h.templateService.Update(c.Request.Context(), id, &req)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Success(t))
}

func (h *TemplateHandler) UpdateStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error(errcode.InvalidParams, "invalid id"))
		return
	}

	var req request.UpdateTemplateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(errcode.InvalidParams, err.Error()))
		return
	}

	if err := h.templateService.UpdateStatus(c.Request.Context(), id, req.Status); err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Success(nil))
}

func (h *TemplateHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error(errcode.InvalidParams, "invalid id"))
		return
	}

	if err := h.templateService.Delete(c.Request.Context(), id); err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Success(nil))
}

// Browse lists all published templates for consumers.
func (h *TemplateHandler) Browse(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	data, err := h.templateService.ListPublished(c.Request.Context(), page, pageSize)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Success(data))
}

// BrowseDetail returns a single published template detail for consumers.
func (h *TemplateHandler) BrowseDetail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error(errcode.InvalidParams, "invalid id"))
		return
	}

	t, err := h.templateService.GetByID(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Success(t))
}
