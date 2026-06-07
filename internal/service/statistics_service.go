package service

import (
	"context"

	"hxcoupon/internal/dto/response"
	"hxcoupon/internal/pkg/apperror"
	"hxcoupon/internal/pkg/errcode"
	"hxcoupon/internal/repository"
)

type StatisticsService struct {
	storeRepo    *repository.StoreRepo
	templateRepo *repository.TemplateRepo
	instanceRepo *repository.InstanceRepo
}

func NewStatisticsService(sr *repository.StoreRepo, tr *repository.TemplateRepo, ir *repository.InstanceRepo) *StatisticsService {
	return &StatisticsService{storeRepo: sr, templateRepo: tr, instanceRepo: ir}
}

func (s *StatisticsService) Overview(ctx context.Context) (*response.OverviewResponse, error) {
	totalStores, _ := s.storeRepo.Count(ctx)
	totalTemplates, _ := s.templateRepo.Count(ctx)
	totalIssued, _ := s.instanceRepo.CountByStatus(ctx, "")
	totalUsed, _ := s.instanceRepo.CountByStatus(ctx, "used")
	todayIssued, _ := s.instanceRepo.CountIssuedToday(ctx)
	todayUsed, _ := s.instanceRepo.CountUsedToday(ctx)

	usageRate := 0.0
	if totalIssued > 0 {
		usageRate = float64(totalUsed) / float64(totalIssued)
	}

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

func (s *StatisticsService) Trend(ctx context.Context, startDate, endDate string) ([]response.TrendItem, error) {
	issuedResults, err := s.instanceRepo.TrendIssuedByDate(ctx, startDate, endDate)
	if err != nil {
		return nil, apperror.NewWithErr(errcode.InternalError, err)
	}
	usedResults, err := s.instanceRepo.TrendUsedByDate(ctx, startDate, endDate)
	if err != nil {
		return nil, apperror.NewWithErr(errcode.InternalError, err)
	}

	// Merge issued and used by date
	trendMap := make(map[string]*response.TrendItem)
	for _, r := range issuedResults {
		date, _ := r["date"].(string)
		count, _ := r["count"].(int64)
		trendMap[date] = &response.TrendItem{Date: date, Issued: count, Used: 0}
	}
	for _, r := range usedResults {
		date, _ := r["date"].(string)
		count, _ := r["count"].(int64)
		if item, ok := trendMap[date]; ok {
			item.Used = count
		} else {
			trendMap[date] = &response.TrendItem{Date: date, Issued: 0, Used: count}
		}
	}

	items := make([]response.TrendItem, 0, len(trendMap))
	for _, v := range trendMap {
		items = append(items, *v)
	}
	return items, nil
}
