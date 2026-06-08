package repository

import (
	"context"

	"hxcoupon/internal/model"

	"gorm.io/gorm"
)

type ApprovalRecordRepo struct {
	db *gorm.DB
}

func NewApprovalRecordRepo(db *gorm.DB) *ApprovalRecordRepo {
	return &ApprovalRecordRepo{db: db}
}

func (r *ApprovalRecordRepo) Create(ctx context.Context, record *model.ApprovalRecord) error {
	return r.db.WithContext(ctx).Create(record).Error
}

func (r *ApprovalRecordRepo) ListByUserID(ctx context.Context, userID uint64) ([]model.ApprovalRecord, error) {
	var records []model.ApprovalRecord
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC").Find(&records).Error
	return records, err
}
