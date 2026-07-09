package openapi

import (
	"net/http"

	"hxcoupon/internal/dto/request"
	"hxcoupon/internal/dto/response"
	"hxcoupon/internal/pkg/errcode"
	"hxcoupon/internal/service"

	"github.com/gin-gonic/gin"
)

type TemplateHandler struct {
	templateService *service.TemplateService
	storeService    *service.StoreService
}

func NewTemplateHandler(templateService *service.TemplateService, storeService *service.StoreService) *TemplateHandler {
	return &TemplateHandler{templateService: templateService, storeService: storeService}
}

// Create creates a new coupon template via Open API (HMAC auth).
// The authenticated store becomes the owner and, when applicable_scope is
// "specific" without explicit store_ids, is automatically assigned as the
// target store.
func (h *TemplateHandler) Create(c *gin.Context) {
	storeID, _ := c.Get("store_id")

	var req request.CreateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(errcode.InvalidParams, err.Error()))
		return
	}

	// Resolve store name for the created_by audit field.
	createdBy := ""
	store, err := h.storeService.GetByID(c.Request.Context(), storeID.(uint64))
	if err == nil {
		createdBy = store.Name
	}

	// Auto-assign the calling store when scope is specific but no store_ids given.
	if req.ApplicableScope == "specific" && len(req.StoreIDs) == 0 {
		req.StoreIDs = []uint64{storeID.(uint64)}
	}

	t, err := h.templateService.Create(c.Request.Context(), &req, createdBy)
	if err != nil {
		handleError(c, err)
		return
	}

	// Open API templates are published immediately so stores can start issuing.
	if err := h.templateService.UpdateStatus(c.Request.Context(), t.ID, 1); err != nil {
		handleError(c, err)
		return
	}
	t.Status = 1

	c.JSON(http.StatusCreated, response.Success(t))
}
