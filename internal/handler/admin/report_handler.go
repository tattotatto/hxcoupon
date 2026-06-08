package admin

import (
	"net/http"

	"hxcoupon/internal/dto/response"
	"hxcoupon/internal/middleware"
	"hxcoupon/internal/pkg/errcode"
	"hxcoupon/internal/service"

	"github.com/gin-gonic/gin"
)

type ReportHandler struct {
	reportService *service.ReportService
}

func NewReportHandler(reportService *service.ReportService) *ReportHandler {
	return &ReportHandler{reportService: reportService}
}

func (h *ReportHandler) Overview(c *gin.Context) {
	role := middleware.GetUserRole(c)
	userID := middleware.GetUserID(c)

	result, err := h.reportService.Overview(c.Request.Context(), role, userID, nil)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Success(result))
}

func (h *ReportHandler) Trend(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	if startDate == "" || endDate == "" {
		c.JSON(http.StatusBadRequest, response.Error(errcode.InvalidParams, "start_date and end_date required"))
		return
	}

	result, err := h.reportService.Trend(c.Request.Context(), startDate, endDate)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Success(result))
}

func (h *ReportHandler) ExportCoupons(c *gin.Context) {
	data, err := h.reportService.ExportCouponsCSV(c.Request.Context())
	if err != nil {
		handleError(c, err)
		return
	}
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=coupons.csv")
	c.Data(http.StatusOK, "text/csv; charset=utf-8", data)
}

func (h *ReportHandler) ExportUsage(c *gin.Context) {
	data, err := h.reportService.ExportUsageCSV(c.Request.Context())
	if err != nil {
		handleError(c, err)
		return
	}
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=usage_records.csv")
	c.Data(http.StatusOK, "text/csv; charset=utf-8", data)
}
