package admin

import (
	"net/http"

	"hxcoupon/internal/dto/response"
	"hxcoupon/internal/service"

	"github.com/gin-gonic/gin"
)

type StatisticsHandler struct {
	statisticsService *service.StatisticsService
}

func NewStatisticsHandler(statisticsService *service.StatisticsService) *StatisticsHandler {
	return &StatisticsHandler{statisticsService: statisticsService}
}

func (h *StatisticsHandler) Overview(c *gin.Context) {
	data, err := h.statisticsService.Overview(c.Request.Context())
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Success(data))
}

func (h *StatisticsHandler) Trend(c *gin.Context) {
	startDate := c.DefaultQuery("start_date", "")
	endDate := c.DefaultQuery("end_date", "")
	if startDate == "" || endDate == "" {
		c.JSON(http.StatusBadRequest, response.Error(40001, "start_date and end_date required"))
		return
	}

	data, err := h.statisticsService.Trend(c.Request.Context(), startDate, endDate)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Success(data))
}
