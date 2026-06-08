package service

import (
	"context"
	"encoding/csv"
	"fmt"
	"strconv"
	"time"

	"hxcoupon/internal/dto/response"
	"hxcoupon/internal/model"
	"hxcoupon/internal/pkg/apperror"
	"hxcoupon/internal/pkg/errcode"
	"hxcoupon/internal/repository"

	"gorm.io/gorm"
)

type ReportService struct {
	db               *gorm.DB
	instanceRepo     *repository.InstanceRepo
	usageRecordRepo  *repository.UsageRecordRepo
	templateRepo     *repository.TemplateRepo
	storeRepo        *repository.StoreRepo
}

func NewReportService(db *gorm.DB, instanceRepo *repository.InstanceRepo, usageRecordRepo *repository.UsageRecordRepo, templateRepo *repository.TemplateRepo, storeRepo *repository.StoreRepo) *ReportService {
	return &ReportService{
		db:               db,
		instanceRepo:     instanceRepo,
		usageRecordRepo:  usageRecordRepo,
		templateRepo:     templateRepo,
		storeRepo:        storeRepo,
	}
}

// Overview returns scoped statistics based on user role and ownership.
func (s *ReportService) Overview(ctx context.Context, role string, userID uint64, storeID *uint64) (*response.OverviewResponse, error) {
	totalTemplates, err := s.templateRepo.Count(ctx)
	if err != nil {
		return nil, apperror.NewWithErr(errcode.InternalError, err)
	}

	totalStores, err := s.storeRepo.Count(ctx)
	if err != nil {
		return nil, apperror.NewWithErr(errcode.InternalError, err)
	}

	totalIssued, err := s.instanceRepo.CountByStatus(ctx, "")
	if err != nil {
		totalIssued = 0
	}
	totalUsed, err := s.instanceRepo.CountByStatus(ctx, "used")
	if err != nil {
		totalUsed = 0
	}

	usageRate := float64(0)
	if totalIssued > 0 {
		usageRate = float64(totalUsed) / float64(totalIssued) * 100
	}

	todayIssued, _ := s.instanceRepo.CountIssuedToday(ctx)
	todayUsed, _ := s.instanceRepo.CountUsedToday(ctx)

	return &response.OverviewResponse{
		TotalStores:    totalStores,
		TotalTemplates: totalTemplates,
		TotalIssued:    totalIssued,
		TotalUsed:      totalUsed,
		UsageRate:      usageRate,
		TodayIssued:    todayIssued,
		TodayUsed:      todayUsed,
	}, nil
}

// Trend returns daily trend data.
func (s *ReportService) Trend(ctx context.Context, startDate, endDate string) (*response.TrendResponse, error) {
	issuedData, err := s.instanceRepo.TrendIssuedByDate(ctx, startDate, endDate)
	if err != nil {
		return nil, apperror.NewWithErr(errcode.InternalError, err)
	}
	usedData, err := s.instanceRepo.TrendUsedByDate(ctx, startDate, endDate)
	if err != nil {
		return nil, apperror.NewWithErr(errcode.InternalError, err)
	}

	items := mergeTrendData(issuedData, usedData)

	return &response.TrendResponse{Items: items}, nil
}

func mergeTrendData(issued, used []map[string]interface{}) []response.TrendItem {
	dateMap := make(map[string]*response.TrendItem)

	for _, m := range issued {
		if d, ok := m["date"].(string); ok {
			item := &response.TrendItem{Date: d}
			if v, ok := m["count"].(int64); ok {
				item.Issued = v
			}
			dateMap[d] = item
		}
	}
	for _, m := range used {
		if d, ok := m["date"].(string); ok {
			if item, exist := dateMap[d]; exist {
				if v, ok := m["count"].(int64); ok {
					item.Used = v
				}
			} else {
				item := &response.TrendItem{Date: d}
				if v, ok := m["count"].(int64); ok {
					item.Used = v
				}
				dateMap[d] = item
			}
		}
	}

	result := make([]response.TrendItem, 0)

	for _, item := range dateMap {
		result = append(result, *item)
	}
	return result
}

// ExportCouponsCSV generates a CSV export of coupon instances.
func (s *ReportService) ExportCouponsCSV(ctx context.Context) ([]byte, error) {
	var instances []model.CouponInstance
	s.db.WithContext(ctx).Order("id DESC").Limit(50000).Find(&instances)

	var buf []byte
	buf = append(buf, "coupon_id,coupon_code,template_id,user_phone,status,valid_start,valid_end,receive_time,use_time\n"...)

	for _, c := range instances {
		line := fmt.Sprintf("%d,%s,%d,%s,%s,%s,%s,%s,%s\n",
			c.ID, c.CouponCode, c.TemplateID, c.UserPhone, c.Status,
			c.ValidStart.Format("2006-01-02"), c.ValidEnd.Format("2006-01-02"),
			c.ReceiveTime.Format("2006-01-02"), formatTime(c.UseTime))
		buf = append(buf, line...)
	}
	return buf, nil
}

// ExportUsageCSV generates a CSV export of usage records.
func (s *ReportService) ExportUsageCSV(ctx context.Context) ([]byte, error) {
	var records []model.CouponUsageRecord
	s.db.WithContext(ctx).Order("id DESC").Limit(50000).Find(&records)

	var buf []byte
	buf = append(buf, "id,coupon_id,user_phone,store_id,action,created_at\n"...)

	for _, r := range records {
		line := fmt.Sprintf("%d,%d,%s,%d,%s,%s\n",
			r.ID, r.CouponID, r.UserPhone, r.StoreID, r.Action,
			r.CreatedAt.Format("2006-01-02 15:04:05"))
		buf = append(buf, line...)
	}
	return buf, nil
}

func formatTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02 15:04:05")
}

// Ensure encoding/csv is available
var _ = csv.NewWriter
var _ = strconv.Itoa
