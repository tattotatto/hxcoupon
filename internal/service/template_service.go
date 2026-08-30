package service

import (
	"context"
	"fmt"
	"time"

	"hxcoupon/internal/dto/request"
	"hxcoupon/internal/dto/response"
	"hxcoupon/internal/model"
	"hxcoupon/internal/pkg/apperror"
	"hxcoupon/internal/pkg/errcode"
	redisutil "hxcoupon/internal/pkg/redis"
	"hxcoupon/internal/repository"

	"gorm.io/gorm"
)

type TemplateService struct {
	db                *gorm.DB
	templateRepo      *repository.TemplateRepo
	templateStoreRepo *repository.TemplateStoreRepo
	storeRepo         *repository.StoreRepo
}

func NewTemplateService(db *gorm.DB, tr *repository.TemplateRepo, tsr *repository.TemplateStoreRepo, sr *repository.StoreRepo) *TemplateService {
	return &TemplateService{db: db, templateRepo: tr, templateStoreRepo: tsr, storeRepo: sr}
}

func (s *TemplateService) Create(ctx context.Context, req *request.CreateTemplateRequest, createdBy string) (*response.TemplateResponse, error) {
	// The owning/creator store: explicit create_store_id wins; otherwise the
	// first applicable store for specific-scope templates. Legacy all-scope
	// templates keep NULL and coupon resolution falls back to the issuing store.
	var createStoreID *uint64
	if req.CreateStoreID != nil && *req.CreateStoreID > 0 {
		createStoreID = req.CreateStoreID
	} else if req.ApplicableScope == "specific" && len(req.StoreIDs) > 0 {
		sid := req.StoreIDs[0]
		createStoreID = &sid
	}

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
		StoreID:         createStoreID,
		CreatedBy:       createdBy,
	}

	if req.ProductRestriction != "" {
		j := model.JSON(req.ProductRestriction)
		t.ProductRestriction = &j
	}

	if req.ValidityType == "fixed_date" && req.ValidStart != "" {
		start, _ := time.Parse("2006-01-02", req.ValidStart)
		t.ValidStart = &start
	}
	if req.ValidityType == "fixed_date" && req.ValidEnd != "" {
		end, _ := time.Parse("2006-01-02", req.ValidEnd)
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

	// No need to invalidate on create since it's a new ID, but populate cache
	redisutil.CacheSet(ctx, fmt.Sprintf("%s%d", redisutil.KeyTemplate, t.ID), *t, redisutil.TTLTemplate)

	storeIDs, _ := s.templateStoreRepo.GetStoreIDsByTemplateID(ctx, t.ID)
	resp := response.ToTemplateResponse(t, storeIDs)
	s.fillMpInfo(ctx, resp, t)
	return resp, nil
}

func (s *TemplateService) GetByID(ctx context.Context, id uint64) (*response.TemplateResponse, error) {
	t, err := s.templateRepo.GetByID(ctx, id)
	if err != nil {
		return nil, apperror.New(errcode.NotFound)
	}
	storeIDs, _ := s.templateStoreRepo.GetStoreIDsByTemplateID(ctx, id)
	resp := response.ToTemplateResponse(t, storeIDs)
	s.fillMpInfo(ctx, resp, t)
	return resp, nil
}

func (s *TemplateService) List(ctx context.Context, f request.TemplateListRequest) (*response.PaginatedData, error) {
	// If store_id is provided, filter templates by the store assignment
	if f.StoreID != nil {
		return s.listByStoreID(ctx, *f.StoreID, f.Status)
	}

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
		s.fillMpInfo(ctx, &items[i], &t)
	}

	return &response.PaginatedData{
		Total:    total,
		Page:     filter.Page,
		PageSize: filter.PageSize,
		Items:    items,
	}, nil
}

// listByStoreID returns templates available to a specific store:
// templates assigned to the store (specific scope) + global templates (all scope).
func (s *TemplateService) listByStoreID(ctx context.Context, storeID uint64, status *int8) (*response.PaginatedData, error) {
	// Get store-specific template IDs from junction table
	storeTemplateIDs, err := s.templateStoreRepo.GetTemplateIDsByStoreID(ctx, storeID)
	if err != nil {
		return nil, apperror.NewWithErr(errcode.InternalError, err)
	}

	// Get global template IDs (applicable_scope = "all")
	globalTemplateIDs, err := s.templateRepo.GetGlobalTemplateIDs(ctx, status)
	if err != nil {
		return nil, apperror.NewWithErr(errcode.InternalError, err)
	}

	// Merge and deduplicate
	idSet := make(map[uint64]struct{})
	for _, id := range storeTemplateIDs {
		idSet[uint64(id)] = struct{}{}
	}
	for _, id := range globalTemplateIDs {
		idSet[uint64(id)] = struct{}{}
	}

	if len(idSet) == 0 {
		return &response.PaginatedData{
			Total: 0,
			Items: []response.TemplateResponse{},
		}, nil
	}

	ids := make([]uint64, 0, len(idSet))
	for id := range idSet {
		ids = append(ids, id)
	}

	templates, err := s.templateRepo.ListByIDs(ctx, ids, status)
	if err != nil {
		return nil, apperror.NewWithErr(errcode.InternalError, err)
	}

	items := make([]response.TemplateResponse, len(templates))
	for i, t := range templates {
		storeIDs, _ := s.templateStoreRepo.GetStoreIDsByTemplateID(ctx, t.ID)
		items[i] = *response.ToTemplateResponse(&t, storeIDs)
		s.fillMpInfo(ctx, &items[i], &t)
	}

	return &response.PaginatedData{
		Total:    int64(len(templates)),
		Page:     1,
		PageSize: len(templates),
		Items:    items,
	}, nil
}

// ListPublished returns all enabled templates for consumers to browse.
func (s *TemplateService) ListPublished(ctx context.Context, page, pageSize int) (*response.PaginatedData, error) {
	status := int8(1)
	filter := repository.TemplateListFilter{
		Status:   &status,
		Page:     page,
		PageSize: pageSize,
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
		s.fillMpInfo(ctx, &items[i], &t)
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
		start, _ := time.Parse("2006-01-02", req.ValidStart)
		t.ValidStart = &start
	}
	if req.ValidEnd != "" {
		end, _ := time.Parse("2006-01-02", req.ValidEnd)
		t.ValidEnd = &end
	}

	if err := s.templateRepo.Update(ctx, t); err != nil {
		return nil, apperror.NewWithErr(errcode.InternalError, err)
	}

	s.invalidateTemplateCache(ctx, id)

	storeIDs, _ := s.templateStoreRepo.GetStoreIDsByTemplateID(ctx, id)
	resp := response.ToTemplateResponse(t, storeIDs)
	s.fillMpInfo(ctx, resp, t)
	return resp, nil
}

func (s *TemplateService) UpdateStatus(ctx context.Context, id uint64, status int8) error {
	if _, err := s.templateRepo.GetByID(ctx, id); err != nil {
		return apperror.New(errcode.NotFound)
	}
	if err := s.templateRepo.UpdateStatus(ctx, id, status); err != nil {
		return err
	}
	s.invalidateTemplateCache(ctx, id)
	return nil
}

func (s *TemplateService) ResetToDraft(ctx context.Context, id uint64) error {
	t, err := s.templateRepo.GetByID(ctx, id)
	if err != nil {
		return apperror.New(errcode.NotFound)
	}
	if t.Status == 0 {
		return apperror.NewWithMsg(errcode.Forbidden, "template is already a draft")
	}
	if t.IssuedCount > 0 {
		return apperror.NewWithMsg(errcode.Forbidden, "only templates with 0 issued count can be reset to draft")
	}
	if err := s.templateRepo.UpdateStatus(ctx, id, 0); err != nil {
		return err
	}
	s.invalidateTemplateCache(ctx, id)
	return nil
}

// IncreaseQuantity adds more inventory to a published or disabled template.
func (s *TemplateService) IncreaseQuantity(ctx context.Context, id uint64, amount uint) error {
	t, err := s.templateRepo.GetByID(ctx, id)
	if err != nil {
		return apperror.New(errcode.NotFound)
	}
	if t.Status == 0 {
		return apperror.NewWithMsg(errcode.Forbidden, "草稿状态的模板请直接编辑发行总量")
	}
	if err := s.templateRepo.IncreaseTotalQuantity(ctx, id, amount); err != nil {
		return apperror.NewWithErr(errcode.InternalError, err)
	}
	s.invalidateTemplateCache(ctx, id)
	return nil
}

func (s *TemplateService) Delete(ctx context.Context, id uint64) error {
	t, err := s.templateRepo.GetByID(ctx, id)
	if err != nil {
		return apperror.New(errcode.NotFound)
	}
	if t.Status != 0 {
		return apperror.NewWithMsg(errcode.Forbidden, "only draft templates can be deleted")
	}
	if err := s.templateRepo.UpdateStatus(ctx, id, 3); err != nil {
		return err
	}
	s.invalidateTemplateCache(ctx, id)
	return nil
}

// GetTemplateByID is a lightweight lookup for internal use (with Redis cache).
func (s *TemplateService) GetTemplateByID(ctx context.Context, id uint64) (*model.CouponTemplate, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("%s%d", redisutil.KeyTemplate, id)
	var cached model.CouponTemplate
	if redisutil.CacheGet(ctx, cacheKey, &cached) {
		return &cached, nil
	}

	t, err := s.templateRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Populate cache
	redisutil.CacheSet(ctx, cacheKey, *t, redisutil.TTLTemplate)
	return t, nil
}

// invalidateTemplateCache removes cached template by id.
func (s *TemplateService) invalidateTemplateCache(ctx context.Context, id uint64) {
	redisutil.CacheDelete(ctx, fmt.Sprintf("%s%d", redisutil.KeyTemplate, id))
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

// fillMpInfo populates the mini-program and store info on a TemplateResponse.
// The creator store (store_id) takes priority over the applicable stores so
// template cards and coupon details agree on which store owns the template.
func (s *TemplateService) fillMpInfo(ctx context.Context, resp *response.TemplateResponse, t *model.CouponTemplate) {
	resp.MpAppID, resp.MpPagePath = s.ResolveMpInfo(ctx, t)
	for _, sid := range s.leaderStoreIDs(ctx, t) {
		store, err := s.storeRepo.GetByID(ctx, sid)
		if err != nil {
			continue
		}
		resp.StoreName = store.Name
		break
	}
}

// leaderStoreIDs returns the candidate stores that represent a template: the
// creator store (store_id) first, then the applicable stores, and finally any
// active store for legacy all-scope templates that record no creator store.
func (s *TemplateService) leaderStoreIDs(ctx context.Context, t *model.CouponTemplate) []uint64 {
	var ids []uint64
	seen := make(map[uint64]struct{})
	add := func(id uint64) {
		if id == 0 {
			return
		}
		if _, ok := seen[id]; ok {
			return
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}

	if t.StoreID != nil {
		add(*t.StoreID)
	}
	if t.ApplicableScope == "specific" {
		appIDs, _ := s.templateStoreRepo.GetStoreIDsByTemplateID(ctx, t.ID)
		for _, id := range appIDs {
			add(id)
		}
	}
	if len(ids) == 0 {
		stores, err := s.storeRepo.ListActive(ctx)
		if err == nil && len(stores) > 0 {
			ids = append(ids, stores[0].ID)
		}
	}
	return ids
}

// ResolveMpInfo returns the mini-program AppID and page path for a template's
// owning store: the creator store first, then the first applicable store.
func (s *TemplateService) ResolveMpInfo(ctx context.Context, t *model.CouponTemplate) (mpAppID, mpPagePath string) {
	for _, sid := range s.leaderStoreIDs(ctx, t) {
		store, err := s.storeRepo.GetByID(ctx, sid)
		if err != nil {
			continue
		}
		if store.MpAppID != nil && *store.MpAppID != "" {
			mpAppID = *store.MpAppID
			if store.MpPagePath != nil {
				mpPagePath = *store.MpPagePath
			}
			return
		}
	}
	return "", ""
}
