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

func (r *AdminUserRepo) UpdateLastLogin(ctx context.Context, id uint64) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&model.AdminUser{}).Where("id = ?", id).Update("last_login_at", now).Error
}
