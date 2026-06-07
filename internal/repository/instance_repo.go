package repository

import (
	"context"
	"time"

	"hxcoupon/internal/model"

	"gorm.io/gorm"
)

type InstanceRepo struct {
	db *gorm.DB
}

func NewInstanceRepo(db *gorm.DB) *InstanceRepo {
	return &InstanceRepo{db: db}
}

func (r *InstanceRepo) Create(ctx context.Context, tx *gorm.DB, ci *model.CouponInstance) error {
	return tx.WithContext(ctx).Create(ci).Error
}

func (r *InstanceRepo) GetByID(ctx context.Context, id uint64) (*model.CouponInstance, error) {
	var ci model.CouponInstance
	err := r.db.WithContext(ctx).First(&ci, id).Error
	if err != nil {
		return nil, err
	}
	return &ci, nil
}

func (r *InstanceRepo) GetByCode(ctx context.Context, code string) (*model.CouponInstance, error) {
	var ci model.CouponInstance
	err := r.db.WithContext(ctx).Where("coupon_code = ?", code).First(&ci).Error
	if err != nil {
		return nil, err
	}
	return &ci, nil
}

// LockByCodeForUpdate locks the coupon row with FOR UPDATE for concurrency safety.
func (r *InstanceRepo) LockByCodeForUpdate(ctx context.Context, tx *gorm.DB, code string) (*model.CouponInstance, error) {
	var ci model.CouponInstance
	err := tx.WithContext(ctx).Raw(
		"SELECT * FROM coupon_instances WHERE coupon_code = ? FOR UPDATE", code,
	).Scan(&ci).Error
	if err != nil {
		return nil, err
	}
	return &ci, nil
}

func (r *InstanceRepo) GetByIdempotencyKey(ctx context.Context, sourceStoreID uint64, key string) (*model.CouponInstance, error) {
	var ci model.CouponInstance
	err := r.db.WithContext(ctx).
		Where("source_store_id = ? AND idempotency_key = ?", sourceStoreID, key).
		First(&ci).Error
	if err != nil {
		return nil, err
	}
	return &ci, nil
}

func (r *InstanceRepo) CountByTemplateAndPhone(ctx context.Context, tx *gorm.DB, templateID uint64, phone string) (int64, error) {
	var count int64
	err := tx.WithContext(ctx).Model(&model.CouponInstance{}).
		Where("template_id = ? AND user_phone = ? AND status != 'expired'", templateID, phone).
		Count(&count).Error
	return count, err
}

type AvailableCouponQuery struct {
	UserPhone   string
	StoreID     uint64
	OrderAmount float64
	Offset      int
	Limit       int
}

// FindAvailable returns coupons usable at the given store that meet the order amount threshold.
func (r *InstanceRepo) FindAvailable(ctx context.Context, q AvailableCouponQuery) ([]model.CouponInstance, int64, error) {
	baseQuery := `
		SELECT ci.*
		FROM coupon_instances ci
		JOIN coupon_templates ct ON ci.template_id = ct.id
		LEFT JOIN coupon_template_stores cts ON ct.id = cts.template_id
		WHERE ci.user_phone = ?
		  AND ci.status = 'unused'
		  AND ci.valid_end > ?
		  AND ci.valid_start <= ?
		  AND ct.status = 1
		  AND (ct.applicable_scope = 'all'
		       OR (ct.applicable_scope = 'specific' AND cts.store_id = ?))
		  AND (ct.threshold_amount = 0 OR ct.threshold_amount <= ?)
		GROUP BY ci.id
	`

	now := time.Now()
	var total int64
	countSQL := "SELECT COUNT(DISTINCT ci.id) FROM coupon_instances ci JOIN coupon_templates ct ON ci.template_id = ct.id LEFT JOIN coupon_template_stores cts ON ct.id = cts.template_id WHERE ci.user_phone = ? AND ci.status = 'unused' AND ci.valid_end > ? AND ci.valid_start <= ? AND ct.status = 1 AND (ct.applicable_scope = 'all' OR (ct.applicable_scope = 'specific' AND cts.store_id = ?)) AND (ct.threshold_amount = 0 OR ct.threshold_amount <= ?)"
	if err := r.db.WithContext(ctx).Raw(countSQL, q.UserPhone, now, now, q.StoreID, q.OrderAmount).Scan(&total).Error; err != nil {
		return nil, 0, err
	}

	var instances []model.CouponInstance
	err := r.db.WithContext(ctx).Raw(
		baseQuery+" ORDER BY ci.valid_end ASC LIMIT ? OFFSET ?",
		q.UserPhone, now, now, q.StoreID, q.OrderAmount, q.Limit, q.Offset,
	).Scan(&instances).Error
	if err != nil {
		return nil, 0, err
	}
	return instances, total, nil
}

func (r *InstanceRepo) ListByUser(ctx context.Context, phone, status string, offset, limit int) ([]model.CouponInstance, int64, error) {
	var instances []model.CouponInstance
	var total int64

	query := r.db.WithContext(ctx).Model(&model.CouponInstance{}).Where("user_phone = ?", phone)
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("id DESC").Offset(offset).Limit(limit).Find(&instances).Error
	return instances, total, err
}

// ConsumeUpdate updates the coupon to 'used' state with optimistic lock check.
func (r *InstanceRepo) ConsumeUpdate(ctx context.Context, tx *gorm.DB, id uint64, version uint, storeID uint64, orderID string, orderAmount float64) (bool, error) {
	result := tx.WithContext(ctx).Model(&model.CouponInstance{}).
		Where("id = ? AND version = ? AND status = 'unused'", id, version).
		Updates(map[string]interface{}{
			"status":           "used",
			"use_time":         time.Now(),
			"used_at_store_id": storeID,
			"use_order_id":     orderID,
			"use_order_amount": orderAmount,
			"version":          version + 1,
		})
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

// RefundRestore restores coupon to 'unused' or 'expired' based on validity.
func (r *InstanceRepo) RefundRestore(ctx context.Context, tx *gorm.DB, id uint64, version uint, newStatus string) (bool, error) {
	updates := map[string]interface{}{
		"status":           newStatus,
		"use_time":         nil,
		"used_at_store_id": nil,
		"use_order_id":     nil,
		"use_order_amount": nil,
		"version":          version + 1,
	}
	result := tx.WithContext(ctx).Model(&model.CouponInstance{}).
		Where("id = ? AND version = ? AND status = 'used'", id, version).
		Updates(updates)
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

// ExpireBatch marks expired unused coupons in batches.
func (r *InstanceRepo) ExpireBatch(ctx context.Context, limit int) (int64, error) {
	result := r.db.WithContext(ctx).Model(&model.CouponInstance{}).
		Where("status = 'unused' AND valid_end <= ?", time.Now()).
		Limit(limit).
		Update("status", "expired")
	return result.RowsAffected, result.Error
}

func (r *InstanceRepo) CountByStatus(ctx context.Context, status string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.CouponInstance{}).Where("status = ?", status).Count(&count).Error
	return count, err
}

func (r *InstanceRepo) CountIssuedToday(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.CouponInstance{}).
		Where("DATE(receive_time) = CURDATE()").Count(&count).Error
	return count, err
}

func (r *InstanceRepo) CountUsedToday(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.CouponInstance{}).
		Where("status = 'used' AND DATE(use_time) = CURDATE()").Count(&count).Error
	return count, err
}

func (r *InstanceRepo) TrendIssuedByDate(ctx context.Context, start, end string) ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	err := r.db.WithContext(ctx).Raw(
		"SELECT DATE(receive_time) as date, COUNT(*) as count FROM coupon_instances WHERE receive_time BETWEEN ? AND ? GROUP BY DATE(receive_time) ORDER BY date",
		start, end,
	).Scan(&results).Error
	return results, err
}

func (r *InstanceRepo) TrendUsedByDate(ctx context.Context, start, end string) ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	err := r.db.WithContext(ctx).Raw(
		"SELECT DATE(use_time) as date, COUNT(*) as count FROM coupon_instances WHERE status = 'used' AND use_time BETWEEN ? AND ? GROUP BY DATE(use_time) ORDER BY date",
		start, end,
	).Scan(&results).Error
	return results, err
}
