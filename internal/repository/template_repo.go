package repository

import (
	"context"
	"fmt"

	"hxcoupon/internal/model"

	"gorm.io/gorm"
)

type TemplateRepo struct {
	db *gorm.DB
}

func NewTemplateRepo(db *gorm.DB) *TemplateRepo {
	return &TemplateRepo{db: db}
}

func (r *TemplateRepo) Create(ctx context.Context, t *model.CouponTemplate) error {
	return r.db.WithContext(ctx).Create(t).Error
}

func (r *TemplateRepo) GetByID(ctx context.Context, id uint64) (*model.CouponTemplate, error) {
	var t model.CouponTemplate
	err := r.db.WithContext(ctx).First(&t, id).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *TemplateRepo) GetByIDForUpdate(ctx context.Context, tx *gorm.DB, id uint64) (*model.CouponTemplate, error) {
	var t model.CouponTemplate
	err := tx.WithContext(ctx).Clauses().Where("id = ?", id).First(&t).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *TemplateRepo) LockByID(ctx context.Context, tx *gorm.DB, id uint64) (*model.CouponTemplate, error) {
	var t model.CouponTemplate
	err := tx.WithContext(ctx).Raw(
		"SELECT * FROM coupon_templates WHERE id = ? FOR UPDATE", id,
	).Scan(&t).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

type TemplateListFilter struct {
	Keyword string
	Type    string
	Status  *int8
	Page    int
	PageSize int
}

func (r *TemplateRepo) List(ctx context.Context, f TemplateListFilter) ([]model.CouponTemplate, int64, error) {
	var templates []model.CouponTemplate
	var total int64

	query := r.db.WithContext(ctx).Model(&model.CouponTemplate{})

	if f.Keyword != "" {
		query = query.Where("name LIKE ?", "%"+f.Keyword+"%")
	}
	if f.Type != "" {
		query = query.Where("type = ?", f.Type)
	}
	if f.Status != nil {
		query = query.Where("status = ?", *f.Status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (f.Page - 1) * f.PageSize
	err := query.Order("id DESC").Offset(offset).Limit(f.PageSize).Find(&templates).Error
	return templates, total, err
}

func (r *TemplateRepo) Update(ctx context.Context, t *model.CouponTemplate) error {
	return r.db.WithContext(ctx).Save(t).Error
}

func (r *TemplateRepo) UpdateStatus(ctx context.Context, id uint64, status int8) error {
	return r.db.WithContext(ctx).Model(&model.CouponTemplate{}).Where("id = ?", id).Update("status", status).Error
}

func (r *TemplateRepo) IncrementIssued(ctx context.Context, tx *gorm.DB, id uint64) error {
	result := tx.WithContext(ctx).Model(&model.CouponTemplate{}).
		Where("id = ? AND (total_quantity = 0 OR issued_count < total_quantity)", id).
		Update("issued_count", gorm.Expr("issued_count + 1"))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("inventory exhausted or template not found")
	}
	return nil
}

// ListByIDs returns templates matching the given IDs, optionally filtered by status.
func (r *TemplateRepo) ListByIDs(ctx context.Context, ids []uint64, status *int8) ([]model.CouponTemplate, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var templates []model.CouponTemplate
	query := r.db.WithContext(ctx).Model(&model.CouponTemplate{}).Where("id IN ?", ids)
	if status != nil {
		query = query.Where("status = ?", *status)
	}
	err := query.Order("id DESC").Find(&templates).Error
	return templates, err
}

func (r *TemplateRepo) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.CouponTemplate{}).Count(&count).Error
	return count, err
}
