package repository

import (
	"context"

	"hxcoupon/internal/model"

	"gorm.io/gorm"
)

type StoreRepo struct {
	db *gorm.DB
}

func NewStoreRepo(db *gorm.DB) *StoreRepo {
	return &StoreRepo{db: db}
}

func (r *StoreRepo) Create(ctx context.Context, store *model.Store) error {
	return r.db.WithContext(ctx).Create(store).Error
}

func (r *StoreRepo) GetByID(ctx context.Context, id uint64) (*model.Store, error) {
	var store model.Store
	err := r.db.WithContext(ctx).First(&store, id).Error
	if err != nil {
		return nil, err
	}
	return &store, nil
}

func (r *StoreRepo) GetByCode(ctx context.Context, code string) (*model.Store, error) {
	var store model.Store
	err := r.db.WithContext(ctx).Where("code = ?", code).First(&store).Error
	if err != nil {
		return nil, err
	}
	return &store, nil
}

func (r *StoreRepo) GetByAppID(ctx context.Context, appID string) (*model.Store, error) {
	var store model.Store
	err := r.db.WithContext(ctx).Where("app_id = ?", appID).First(&store).Error
	if err != nil {
		return nil, err
	}
	return &store, nil
}

func (r *StoreRepo) List(ctx context.Context, page, pageSize int) ([]model.Store, int64, error) {
	var stores []model.Store
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Store{})

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&stores).Error
	return stores, total, err
}

func (r *StoreRepo) Update(ctx context.Context, store *model.Store) error {
	return r.db.WithContext(ctx).Save(store).Error
}

func (r *StoreRepo) UpdateStatus(ctx context.Context, id uint64, status int8) error {
	return r.db.WithContext(ctx).Model(&model.Store{}).Where("id = ?", id).Update("status", status).Error
}

func (r *StoreRepo) ListActive(ctx context.Context) ([]model.Store, error) {
	var stores []model.Store
	err := r.db.WithContext(ctx).Where("status = 1").Find(&stores).Error
	return stores, err
}

func (r *StoreRepo) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Store{}).Count(&count).Error
	return count, err
}
