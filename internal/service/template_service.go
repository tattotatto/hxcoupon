package service

import (
	"context"
	"encoding/json"
	"time"

	"hxcoupon/internal/dto/request"
	"hxcoupon/internal/dto/response"
	"hxcoupon/internal/model"
	"hxcoupon/internal/pkg/apperror"
	"hxcoupon/internal/pkg/errcode"
	"hxcoupon/internal/repository"

	"gorm.io/gorm"
)

type TemplateService struct {
	db                 *gorm.DB
	templateRepo       *repository.TemplateRepo
	templateStoreRepo  *repository.TemplateStoreRepo
	storeRepo          *repository.StoreRepo
}

func NewTemplateService(db *gorm.DB, tr *repository.TemplateRepo, tsr *repository.TemplateStoreRepo, sr *repository.StoreRepo) *TemplateService {
	return &TemplateService{db: db, templateRepo: tr, templateStoreRepo: tsr, storeRepo: sr}
}

func (s *TemplateService) Create(ctx context.Context, req *request.CreateTemplateRequest, createdBy string) (*response.TemplateResponse, error) {
	t := &model.CouponTemplate{
		Name:            req.Name,
		Type:            req.Type,
		DiscountValue:   req.DiscountValue,
		ThresholdAmount: req.ThresholdAmount,
		ApplicableScope: req.ApplicableScope,
		Stackable:       req.Stackable,
		MaxStackCount:   req.MaxStackCount,
		ValidityType:    req.ValidityType,
		ValidityDays:    req.ValidityDays,
		TotalQuantity:   req.TotalQuantity,
		PerUserLimit:    req.PerUserLimit,
		Status:          0,
		CreatedBy:       createdBy,
	}

	if req.ProductRestriction != "" {
		j := model.JSON(req.ProductRestriction)
		t.ProductRestriction = &j
	}

	if req.ValidityType == "fixed_date" && req.ValidStart != "" {
		start, _ := time.Parse("2006-01-02 15:04:05", req.ValidStart)
		t.ValidStart = &start
	}
	if req.ValidityType == "fixed_date" && req.ValidEnd != "" {
		end, _ := time.Parse("2006-01-02 15:04:05", req.ValidEnd)
		t.ValidEnd = &end
	}

	if req.ApplicableScope == "specific" && len(req.StoreIDs) == 0 {
		return nil, apperror.NewWithMsg(errcode.InvalidParams, "specific scope requires store_ids")
	}

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := s.templateRepo.Create(ctx, t); err != nil {
			return err
		}

		if req.ApplicableScope == "specific" {
			items := make([]model.CouponTemplateStore, len(req.StoreIDs))
			for i, sid := range req.StoreIDs {
				items[i] = model.CouponTemplateStore{TemplateID: t.ID, StoreID: sid}
			}
			if err := s.templateStoreRepo.BatchCreate(ctx, tx, items); err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return nil, apperror.NewWithErr(errcode.InternalError, err)
	}

	storeIDs, _ := s.templateStoreRepo.GetStoreIDsByTemplateID(ctx, t.ID)
	return response.ToTemplateResponse(t, storeIDs), nil
}

func (s *TemplateService) GetByID(ctx context.Context, id uint64) (*response.TemplateResponse, error) {
	t, err := s.templateRepo.GetByID(ctx, id)
	if err != nil {
		return nil, apperror.New(errcode.NotFound)
	}
	storeIDs, _ := s.templateStoreRepo.GetStoreIDsByTemplateID(ctx, id)
	return response.ToTemplateResponse(t, storeIDs), nil
}

func (s *TemplateService) List(ctx context.Context, f request.TemplateListRequest) (*response.PaginatedData, error) {
	filter := repository.TemplateListFilter{
		Keyword:  f.Keyword,
		Type:     f.Type,
		Status:   f.Status,
		Page:     f.Page,
		PageSize: f.PageSize,
	}
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}

	templates, total, err := s.templateRepo.List(ctx, filter)
	if err != nil {
		return nil, apperror.NewWithErr(errcode.InternalError, err)
	}

	items := make([]response.TemplateResponse, len(templates))
	for i, t := range templates {
		storeIDs, _ := s.templateStoreRepo.GetStoreIDsByTemplateID(ctx, t.ID)
		items[i] = *response.ToTemplateResponse(&t, storeIDs)
	}

	return &response.PaginatedData{
		Total:    total,
		Page:     filter.Page,
		PageSize: filter.PageSize,
		Items:    items,
	}, nil
}

func (s *TemplateService) Update(ctx context.Context, id uint64, req *request.UpdateTemplateRequest) (*response.TemplateResponse, error) {
	t, err := s.templateRepo.GetByID(ctx, id)
	if err != nil {
		return nil, apperror.New(errcode.NotFound)
	}

	if t.Status != 0 {
		return nil, apperror.NewWithMsg(errcode.Forbidden, "only draft templates can be updated")
	}

	t.Name = req.Name
	t.DiscountValue = req.DiscountValue
	t.ThresholdAmount = req.ThresholdAmount
	t.Stackable = req.Stackable
	t.MaxStackCount = req.MaxStackCount
	t.ValidityDays = req.ValidityDays
	t.TotalQuantity = req.TotalQuantity
	t.PerUserLimit = req.PerUserLimit

	if req.ProductRestriction != "" {
		j := model.JSON(req.ProductRestriction)
		t.ProductRestriction = &j
	} else {
		t.ProductRestriction = nil
	}

	if req.ValidStart != "" {
		start, _ := time.Parse("2006-01-02 15:04:05", req.ValidStart)
		t.ValidStart = &start
	}
	if req.ValidEnd != "" {
		end, _ := time.Parse("2006-01-02 15:04:05", req.ValidEnd)
		t.ValidEnd = &end
	}

	if err := s.templateRepo.Update(ctx, t); err != nil {
		return nil, apperror.NewWithErr(errcode.InternalError, err)
	}

	storeIDs, _ := s.templateStoreRepo.GetStoreIDsByTemplateID(ctx, id)
	return response.ToTemplateResponse(t, storeIDs), nil
}

func (s *TemplateService) UpdateStatus(ctx context.Context, id uint64, status int8) error {
	if _, err := s.templateRepo.GetByID(ctx, id); err != nil {
		return apperror.New(errcode.NotFound)
	}
	return s.templateRepo.UpdateStatus(ctx, id, status)
}

func (s *TemplateService) Delete(ctx context.Context, id uint64) error {
	t, err := s.templateRepo.GetByID(ctx, id)
	if err != nil {
		return apperror.New(errcode.NotFound)
	}
	if t.Status != 0 {
		return apperror.NewWithMsg(errcode.Forbidden, "only draft templates can be deleted")
	}
	return s.templateRepo.UpdateStatus(ctx, id, 3) // soft-delete as status=3
}

// GetTemplateByID is a lightweight lookup for internal use.
func (s *TemplateService) GetTemplateByID(ctx context.Context, id uint64) (*model.CouponTemplate, error) {
	return s.templateRepo.GetByID(ctx, id)
}

// GetStoreCode returns the 5-char code for a store.
func (s *TemplateService) GetStoreCode(ctx context.Context, storeID uint64) (string, error) {
	store, err := s.storeRepo.GetByID(ctx, storeID)
	if err != nil {
		return "", err
	}
	return store.Code, nil
}

// IsStoreApplicable checks if a template applies to a store.
func (s *TemplateService) IsStoreApplicable(ctx context.Context, templateID, storeID uint64) (bool, error) {
	t, err := s.templateRepo.GetByID(ctx, templateID)
	if err != nil {
		return false, err
	}
	if t.ApplicableScope == "all" {
		return true, nil
	}
	return s.templateStoreRepo.IsStoreApplicable(ctx, templateID, storeID)
}

// GetApplicableStoreID returns a store ID for coupon code generation.
func (s *TemplateService) GetApplicableStoreID(ctx context.Context, templateID uint64) (uint64, error) {
	storeIDs, err := s.templateStoreRepo.GetStoreIDsByTemplateID(ctx, templateID)
	if err != nil || len(storeIDs) == 0 {
		return 0, err
	}
	return storeIDs[0], nil
}

// GetSourceStoreID returns the store_id from the template's first applicable store for code generation
func (s *TemplateService) GetSourceStoreID(ctx context.Context, templateID uint64) (uint64, error) {
	return s.GetApplicableStoreID(ctx, templateID)
}
