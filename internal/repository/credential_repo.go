package repository

import (
	"context"

	"hxcoupon/internal/model"

	"gorm.io/gorm"
)

type CredentialRepo struct {
	db *gorm.DB
}

func NewCredentialRepo(db *gorm.DB) *CredentialRepo {
	return &CredentialRepo{db: db}
}

func (r *CredentialRepo) Create(ctx context.Context, cred *model.StoreAPICredential) error {
	return r.db.WithContext(ctx).Create(cred).Error
}

func (r *CredentialRepo) GetByAppKey(ctx context.Context, appKey string) (*model.StoreAPICredential, error) {
	var cred model.StoreAPICredential
	err := r.db.WithContext(ctx).Where("app_key = ? AND status = 1", appKey).First(&cred).Error
	if err != nil {
		return nil, err
	}
	return &cred, nil
}

func (r *CredentialRepo) GetByStoreID(ctx context.Context, storeID uint64) (*model.StoreAPICredential, error) {
	var cred model.StoreAPICredential
	err := r.db.WithContext(ctx).Where("store_id = ? AND status = 1", storeID).First(&cred).Error
	if err != nil {
		return nil, err
	}
	return &cred, nil
}

func (r *CredentialRepo) DisableByStoreID(ctx context.Context, storeID uint64) error {
	return r.db.WithContext(ctx).Model(&model.StoreAPICredential{}).
		Where("store_id = ?", storeID).
		Update("status", 0).Error
}
