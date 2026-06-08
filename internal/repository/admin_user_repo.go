package repository

import (
	"context"
	"time"

	"hxcoupon/internal/model"

	"gorm.io/gorm"
)

type AdminUserRepo struct {
	db *gorm.DB
}

func NewAdminUserRepo(db *gorm.DB) *AdminUserRepo {
	return &AdminUserRepo{db: db}
}

func (r *AdminUserRepo) GetByUsername(ctx context.Context, username string) (*model.AdminUser, error) {
	var user model.AdminUser
	err := r.db.WithContext(ctx).Where("username = ? AND status = 1", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *AdminUserRepo) GetByID(ctx context.Context, id uint64) (*model.AdminUser, error) {
	var user model.AdminUser
	err := r.db.WithContext(ctx).First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *AdminUserRepo) Create(ctx context.Context, user *model.AdminUser) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *AdminUserRepo) Update(ctx context.Context, user *model.AdminUser) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *AdminUserRepo) UpdateFields(ctx context.Context, id uint64, fields map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&model.AdminUser{}).Where("id = ?", id).Updates(fields).Error
}

func (r *AdminUserRepo) List(ctx context.Context, page, pageSize int, keyword, role, memberType string, approvalStatus *int8) ([]model.AdminUser, int64, error) {
	var users []model.AdminUser
	var total int64

	q := r.db.WithContext(ctx).Model(&model.AdminUser{})
	if keyword != "" {
		q = q.Where("username LIKE ? OR company_name LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	if role != "" {
		q = q.Where("role = ?", role)
	}
	if memberType != "" {
		q = q.Where("member_type = ?", memberType)
	}
	if approvalStatus != nil {
		q = q.Where("approval_status = ?", *approvalStatus)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := q.Order("id DESC").Offset(offset).Limit(pageSize).Find(&users).Error
	return users, total, err
}

func (r *AdminUserRepo) UpdateLastLogin(ctx context.Context, id uint64) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&model.AdminUser{}).Where("id = ?", id).Update("last_login_at", now).Error
}

func (r *AdminUserRepo) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.AdminUser{}).Where("username = ?", username).Count(&count).Error
	return count > 0, err
}
