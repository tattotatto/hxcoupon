package repository

import (
	"context"

	"hxcoupon/internal/model"

	"gorm.io/gorm"
)

type UsageRecordRepo struct {
	db *gorm.DB
}

func NewUsageRecordRepo(db *gorm.DB) *UsageRecordRepo {
	return &UsageRecordRepo{db: db}
}

func (r *UsageRecordRepo) Create(ctx context.Context, tx *gorm.DB, record *model.CouponUsageRecord) error {
	return tx.WithContext(ctx).Create(record).Error
}

func (r *UsageRecordRepo) ListByCouponID(ctx context.Context, couponID uint64) ([]model.CouponUsageRecord, error) {
	var records []model.CouponUsageRecord
	err := r.db.WithContext(ctx).
		Where("coupon_id = ?", couponID).
		Order("created_at DESC").
		Find(&records).Error
	return records, err
}

type UsageRecordListFilter struct {
	StoreID   *uint64
	Action    string
	StartDate string
	EndDate   string
	Page      int
	PageSize  int
}

func (r *UsageRecordRepo) List(ctx context.Context, f UsageRecordListFilter) ([]model.CouponUsageRecord, int64, error) {
	var records []model.CouponUsageRecord
	var total int64

	query := r.db.WithContext(ctx).Model(&model.CouponUsageRecord{})

	if f.StoreID != nil {
		query = query.Where("store_id = ?", *f.StoreID)
	}
	if f.Action != "" {
		query = query.Where("action = ?", f.Action)
	}
	if f.StartDate != "" {
		query = query.Where("created_at >= ?", f.StartDate)
	}
	if f.EndDate != "" {
		query = query.Where("created_at <= ?", f.EndDate+" 23:59:59")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (f.Page - 1) * f.PageSize
	err := query.Order("id DESC").Offset(offset).Limit(f.PageSize).Find(&records).Error
	return records, total, err
}
