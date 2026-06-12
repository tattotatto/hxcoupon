package repository

import (
	"context"

	"hxcoupon/internal/model"

	"gorm.io/gorm"
)

type TemplateStoreRepo struct {
	db *gorm.DB
}

func NewTemplateStoreRepo(db *gorm.DB) *TemplateStoreRepo {
	return &TemplateStoreRepo{db: db}
}

func (r *TemplateStoreRepo) BatchCreate(ctx context.Context, tx *gorm.DB, items []model.CouponTemplateStore) error {
	if len(items) == 0 {
		return nil
	}
	return tx.WithContext(ctx).Create(&items).Error
}

func (r *TemplateStoreRepo) GetStoreIDsByTemplateID(ctx context.Context, templateID uint64) ([]uint64, error) {
	var ids []uint64
	err := r.db.WithContext(ctx).Model(&model.CouponTemplateStore{}).
		Where("template_id = ?", templateID).
		Pluck("store_id", &ids).Error
	return ids, err
}

func (r *TemplateStoreRepo) DeleteByTemplateID(ctx context.Context, tx *gorm.DB, templateID uint64) error {
	return tx.WithContext(ctx).Where("template_id = ?", templateID).Delete(&model.CouponTemplateStore{}).Error
}

func (r *TemplateStoreRepo) IsStoreApplicable(ctx context.Context, templateID, storeID uint64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.CouponTemplateStore{}).
		Where("template_id = ? AND store_id = ?", templateID, storeID).
		Count(&count).Error
	return count > 0, err
}

// GetTemplateIDsByStoreID returns all template IDs assigned to a store.
func (r *TemplateStoreRepo) GetTemplateIDsByStoreID(ctx context.Context, storeID uint64) ([]uint64, error) {
	var ids []uint64
	err := r.db.WithContext(ctx).Model(&model.CouponTemplateStore{}).
		Where("store_id = ?", storeID).
		Pluck("template_id", &ids).Error
	return ids, err
}
