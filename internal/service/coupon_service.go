package service

import (
	"context"
	"encoding/json"
	"time"

	"hxcoupon/internal/dto/response"
	"hxcoupon/internal/model"
	"hxcoupon/internal/pkg/apperror"
	"hxcoupon/internal/pkg/couponcode"
	"hxcoupon/internal/pkg/errcode"
	"hxcoupon/internal/repository"

	"gorm.io/gorm"
)

type CouponService struct {
	db                 *gorm.DB
	instanceRepo       *repository.InstanceRepo
	templateRepo       *repository.TemplateRepo
	templateStoreRepo  *repository.TemplateStoreRepo
	usageRecordRepo    *repository.UsageRecordRepo
	storeRepo          *repository.StoreRepo
	credentialRepo     *repository.CredentialRepo
}

func NewCouponService(
	db *gorm.DB,
	ir *repository.InstanceRepo,
	tr *repository.TemplateRepo,
	tsr *repository.TemplateStoreRepo,
	urr *repository.UsageRecordRepo,
	sr *repository.StoreRepo,
	cr *repository.CredentialRepo,
) *CouponService {
	return &CouponService{
		db:                db,
		instanceRepo:      ir,
		templateRepo:      tr,
		templateStoreRepo: tsr,
		usageRecordRepo:   urr,
		storeRepo:         sr,
		credentialRepo:    cr,
	}
}

// Issue issues a coupon to a user. Idempotent via idempotency_key.
func (s *CouponService) Issue(ctx context.Context, sourceStoreID uint64, templateID uint64, userPhone, idempotencyKey string) (*response.CouponIssueResponse, error) {
	// Check idempotency first (outside transaction for performance)
	existing, err := s.instanceRepo.GetByIdempotencyKey(ctx, sourceStoreID, idempotencyKey)
	if err == nil && existing != nil {
		return s.buildIssueResponse(ctx, existing)
	}

	var result *response.CouponIssueResponse

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Lock template row
		t, err := s.templateRepo.LockByID(ctx, tx, templateID)
		if err != nil {
			return apperror.New(errcode.NotFound)
		}

		// Validate template status
		if t.Status != 1 {
			return apperror.NewWithMsg(errcode.CouponNotApply, "template is not enabled")
		}

		// Check inventory
		if t.TotalQuantity > 0 && t.IssuedCount >= t.TotalQuantity {
			return apperror.New(errcode.NoInventory)
		}

		// Check fixed-date validity
		if t.ValidityType == "fixed_date" {
			if t.ValidEnd != nil && t.ValidEnd.Before(time.Now()) {
				return apperror.New(errcode.CouponExpired)
			}
		}

		// Double-check idempotency within transaction
		existing, err := s.instanceRepo.GetByIdempotencyKey(ctx, sourceStoreID, idempotencyKey)
		if err == nil && existing != nil {
			resp, buildErr := s.buildIssueResponse(ctx, existing)
			if buildErr != nil {
				return buildErr
			}
			result = resp
			return nil
		}

		// Check per-user limit
		count, err := s.instanceRepo.CountByTemplateAndPhone(ctx, tx, templateID, userPhone)
		if err != nil {
			return apperror.NewWithErr(errcode.InternalError, err)
		}
		if count >= int64(t.PerUserLimit) {
			return apperror.New(errcode.PerUserLimit)
		}

		// Calculate instance validity
		validStart, validEnd := s.calculateValidity(t)

		// Generate coupon code
		code, err := s.generateCouponCode(ctx, t)
		if err != nil {
			return apperror.NewWithErr(errcode.InternalError, err)
		}

		now := time.Now()
		ci := &model.CouponInstance{
			TemplateID:     templateID,
			UserPhone:      userPhone,
			SourceStoreID:  sourceStoreID,
			CouponCode:     code,
			Status:         "unused",
			ReceiveTime:    now,
			ValidStart:     validStart,
			ValidEnd:       validEnd,
			IdempotencyKey: idempotencyKey,
			Version:        0,
		}

		if err := s.instanceRepo.Create(ctx, tx, ci); err != nil {
			return apperror.NewWithErr(errcode.InternalError, err)
		}

		// Increment issued count
		if err := s.templateRepo.IncrementIssued(ctx, tx, templateID); err != nil {
			return apperror.NewWithErr(errcode.NoInventory, err)
		}

		result = &response.CouponIssueResponse{
			CouponID:        ci.ID,
			CouponCode:      ci.CouponCode,
			TemplateName:    t.Name,
			Type:            t.Type,
			DiscountValue:   t.DiscountValue,
			ThresholdAmount: t.ThresholdAmount,
			ValidStart:      ci.ValidStart,
			ValidEnd:        ci.ValidEnd,
			Status:          ci.Status,
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetAvailable returns coupons usable at a store for a given order amount.
func (s *CouponService) GetAvailable(ctx context.Context, userPhone string, storeID uint64, orderAmount float64, page, pageSize int) (*response.PaginatedData, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	q := repository.AvailableCouponQuery{
		UserPhone:   userPhone,
		StoreID:     storeID,
		OrderAmount: orderAmount,
		Offset:      (page - 1) * pageSize,
		Limit:       pageSize,
	}

	instances, total, err := s.instanceRepo.FindAvailable(ctx, q)
	if err != nil {
		return nil, apperror.NewWithErr(errcode.InternalError, err)
	}

	items := make([]response.CouponAvailableResponse, len(instances))
	for i, ci := range instances {
		t, _ := s.templateRepo.GetByID(ctx, ci.TemplateID)
		templateName := ""
		if t != nil {
			templateName = t.Name
		}

		items[i] = response.CouponAvailableResponse{
			CouponID:        ci.ID,
			CouponCode:      ci.CouponCode,
			TemplateID:      ci.TemplateID,
			TemplateName:    templateName,
			Type:            "",
			DiscountValue:   0,
			ThresholdAmount: 0,
			ValidEnd:        ci.ValidEnd,
			Stackable:       false,
		}
		if t != nil {
			items[i].Type = t.Type
			items[i].DiscountValue = t.DiscountValue
			items[i].ThresholdAmount = t.ThresholdAmount
			items[i].Stackable = t.Stackable
		}
	}

	return &response.PaginatedData{
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Items:    items,
	}, nil
}

// Consume uses a coupon. Returns the discount info.
func (s *CouponService) Consume(ctx context.Context, couponCode, userPhone string, storeID uint64, orderID string, orderAmount float64) (*response.CouponConsumeResponse, error) {
	var result *response.CouponConsumeResponse

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Lock coupon row
		ci, err := s.instanceRepo.LockByCodeForUpdate(ctx, tx, couponCode)
		if err != nil {
			return apperror.New(errcode.NotFound)
		}

		// Validate ownership
		if ci.UserPhone != userPhone {
			return apperror.NewWithMsg(errcode.Forbidden, "coupon does not belong to this user")
		}

		// Validate status
		if ci.Status != "unused" {
			if ci.Status == "used" {
				return apperror.New(errcode.CouponUsed)
			}
			if ci.Status == "expired" {
				return apperror.New(errcode.CouponExpired)
			}
			return apperror.NewWithMsg(errcode.CouponNotApply, "coupon status: "+ci.Status)
		}

		// Validate validity period
		now := time.Now()
		if now.Before(ci.ValidStart) || now.After(ci.ValidEnd) {
			return apperror.New(errcode.CouponExpired)
		}

		// Validate store applicability
		t, err := s.templateRepo.GetByID(ctx, ci.TemplateID)
		if err != nil {
			return apperror.NewWithErr(errcode.NotFound, err)
		}
		if t.ApplicableScope == "specific" {
			ok, _ := s.templateStoreRepo.IsStoreApplicable(ctx, ci.TemplateID, storeID)
			if !ok {
				return apperror.New(errcode.CouponNotApply)
			}
		}

		// Validate order amount threshold
		if t.ThresholdAmount > 0 && orderAmount < t.ThresholdAmount {
			return apperror.New(errcode.BelowThreshold)
		}

		// Calculate discount
		discount := s.calculateDiscount(t, orderAmount)

		// Update coupon with optimistic lock
		ok, err := s.instanceRepo.ConsumeUpdate(ctx, tx, ci.ID, ci.Version, storeID, orderID, orderAmount)
		if err != nil {
			return apperror.NewWithErr(errcode.InternalError, err)
		}
		if !ok {
			return apperror.New(errcode.CouponUsed) // concurrent modification
		}

		// Record audit
		orderInfo, _ := json.Marshal(map[string]interface{}{
			"order_id":     orderID,
			"order_amount": orderAmount,
		})
		j := model.JSON(orderInfo)
		record := &model.CouponUsageRecord{
			CouponID:  ci.ID,
			UserPhone: userPhone,
			StoreID:   storeID,
			Action:    "consume",
			OrderInfo: &j,
		}
		if err := s.usageRecordRepo.Create(ctx, tx, record); err != nil {
			return apperror.NewWithErr(errcode.InternalError, err)
		}

		result = &response.CouponConsumeResponse{
			CouponID:      ci.ID,
			CouponCode:    ci.CouponCode,
			DiscountValue: discount,
			ActualAmount:  orderAmount - discount,
			UsedAt:        now,
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}

// Refund restores a used coupon if not expired, or marks it expired.
func (s *CouponService) Refund(ctx context.Context, couponCode, userPhone string, storeID uint64, orderID string) (*response.CouponRefundResponse, error) {
	var result *response.CouponRefundResponse

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		ci, err := s.instanceRepo.LockByCodeForUpdate(ctx, tx, couponCode)
		if err != nil {
			return apperror.New(errcode.NotFound)
		}

		// Validate ownership
		if ci.UserPhone != userPhone {
			return apperror.NewWithMsg(errcode.Forbidden, "coupon does not belong to this user")
		}

		// Validate status
		if ci.Status != "used" {
			return apperror.NewWithMsg(errcode.CouponNotApply, "coupon is not in used state")
		}

		// Validate order match
		if ci.UseOrderID != orderID {
			return apperror.New(errcode.RefundMismatch)
		}
		if ci.UsedAtStoreID == nil || *ci.UsedAtStoreID != storeID {
			return apperror.New(errcode.RefundMismatch)
		}

		// Determine new status based on validity
		newStatus := "unused"
		restored := true
		if time.Now().After(ci.ValidEnd) {
			newStatus = "expired"
			restored = false
		}

		ok, err := s.instanceRepo.RefundRestore(ctx, tx, ci.ID, ci.Version, newStatus)
		if err != nil {
			return apperror.NewWithErr(errcode.InternalError, err)
		}
		if !ok {
			return apperror.New(errcode.CouponUsed) // concurrent modification
		}

		// Record audit
		orderInfo, _ := json.Marshal(map[string]interface{}{
			"order_id": orderID,
		})
		j := model.JSON(orderInfo)
		record := &model.CouponUsageRecord{
			CouponID:  ci.ID,
			UserPhone: userPhone,
			StoreID:   storeID,
			Action:    "refund",
			OrderInfo: &j,
		}
		if err := s.usageRecordRepo.Create(ctx, tx, record); err != nil {
			return apperror.NewWithErr(errcode.InternalError, err)
		}

		result = &response.CouponRefundResponse{
			CouponID:   ci.ID,
			CouponCode: ci.CouponCode,
			NewStatus:  newStatus,
			Restored:   restored,
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}

// ListByUser returns all coupons for a user.
func (s *CouponService) ListByUser(ctx context.Context, userPhone, status string, page, pageSize int) (*response.PaginatedData, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	instances, total, err := s.instanceRepo.ListByUser(ctx, userPhone, status, offset, pageSize)
	if err != nil {
		return nil, apperror.NewWithErr(errcode.InternalError, err)
	}

	type userCouponItem struct {
		CouponID        uint64    `json:"coupon_id"`
		CouponCode      string    `json:"coupon_code"`
		TemplateName    string    `json:"template_name"`
		Type            string    `json:"type"`
		DiscountValue   float64   `json:"discount_value"`
		ThresholdAmount float64   `json:"threshold_amount"`
		Status          string    `json:"status"`
		ValidStart      time.Time `json:"valid_start"`
		ValidEnd        time.Time `json:"valid_end"`
	}

	items := make([]userCouponItem, len(instances))
	for i, ci := range instances {
		items[i] = userCouponItem{
			CouponID:   ci.ID,
			CouponCode: ci.CouponCode,
			Status:     ci.Status,
			ValidStart: ci.ValidStart,
			ValidEnd:   ci.ValidEnd,
		}
		t, _ := s.templateRepo.GetByID(ctx, ci.TemplateID)
		if t != nil {
			items[i].TemplateName = t.Name
			items[i].Type = t.Type
			items[i].DiscountValue = t.DiscountValue
			items[i].ThresholdAmount = t.ThresholdAmount
		}
	}

	return &response.PaginatedData{
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Items:    items,
	}, nil
}

// GetDetail returns a single coupon's detail with usage records.
func (s *CouponService) GetDetail(ctx context.Context, couponCode string) (*response.CouponDetailResponse, error) {
	ci, err := s.instanceRepo.GetByCode(ctx, couponCode)
	if err != nil {
		return nil, apperror.New(errcode.NotFound)
	}

	t, _ := s.templateRepo.GetByID(ctx, ci.TemplateID)
	templateName := ""
	if t != nil {
		templateName = t.Name
	}

	sourceStore, _ := s.storeRepo.GetByID(ctx, ci.SourceStoreID)
	sourceStoreName := ""
	if sourceStore != nil {
		sourceStoreName = sourceStore.Name
	}

	usedStoreName := ""
	if ci.UsedAtStoreID != nil {
		us, _ := s.storeRepo.GetByID(ctx, *ci.UsedAtStoreID)
		if us != nil {
			usedStoreName = us.Name
		}
	}

	records, _ := s.usageRecordRepo.ListByCouponID(ctx, ci.ID)
	recordBriefs := make([]response.CouponUsageRecordBrief, len(records))
	for i, r := range records {
		rs, _ := s.storeRepo.GetByID(ctx, r.StoreID)
		storeName := ""
		if rs != nil {
			storeName = rs.Name
		}
		recordBriefs[i] = response.CouponUsageRecordBrief{
			Action:    r.Action,
			StoreName: storeName,
			CreatedAt: r.CreatedAt,
		}
	}

	resp := response.ToCouponDetailResponse(ci, templateName, sourceStoreName, usedStoreName, recordBriefs)
	if t != nil {
		resp.Type = t.Type
		resp.DiscountValue = t.DiscountValue
		resp.ThresholdAmount = t.ThresholdAmount
	}
	return resp, nil
}

// VerifyStoreCredentials verifies store HMAC credentials.
func (s *CouponService) VerifyStoreCredentials(ctx context.Context, appKey string) (uint64, error) {
	cred, err := s.credentialRepo.GetByAppKey(ctx, appKey)
	if err != nil {
		return 0, apperror.New(errcode.StoreNotAuth)
	}
	return cred.StoreID, nil
}

// GetStoreSecret returns the raw app_secret for HMAC verification.
func (s *CouponService) GetStoreSecret(ctx context.Context, appKey string) (string, error) {
	cred, err := s.credentialRepo.GetByAppKey(ctx, appKey)
	if err != nil {
		return "", apperror.New(errcode.StoreNotAuth)
	}
	return cred.AppSecret, nil
}

// --- Private helpers ---

func (s *CouponService) calculateValidity(t *model.CouponTemplate) (time.Time, time.Time) {
	now := time.Now()
	if t.ValidityType == "fixed_date" {
		start := now
		if t.ValidStart != nil {
			start = *t.ValidStart
		}
		end := now.Add(365 * 24 * time.Hour)
		if t.ValidEnd != nil {
			end = *t.ValidEnd
		}
		return start, end
	}

	// days_after_receive
	start := now
	end := now.Add(time.Duration(t.ValidityDays) * 24 * time.Hour)
	return start, end
}

func (s *CouponService) generateCouponCode(ctx context.Context, t *model.CouponTemplate) (string, error) {
	storeCode := ""
	if t.ApplicableScope == "specific" {
		storeIDs, err := s.templateStoreRepo.GetStoreIDsByTemplateID(ctx, t.ID)
		if err != nil || len(storeIDs) == 0 {
			return "", apperror.NewWithErr(errcode.InternalError, err)
		}
		store, err := s.storeRepo.GetByID(ctx, storeIDs[0])
		if err != nil {
			return "", apperror.NewWithErr(errcode.InternalError, err)
		}
		storeCode = store.Code
	}
	return couponcode.Generate(storeCode), nil
}

func (s *CouponService) calculateDiscount(t *model.CouponTemplate, orderAmount float64) float64 {
	switch t.Type {
	case "fixed_amount":
		return t.DiscountValue
	case "full_reduction":
		return t.DiscountValue
	case "discount":
		// discount_value as percentage (e.g., 80 = 20% off)
		discount := orderAmount * (1 - t.DiscountValue/100)
		return discount
	default:
		return 0
	}
}

func (s *CouponService) buildIssueResponse(ctx context.Context, ci *model.CouponInstance) (*response.CouponIssueResponse, error) {
	t, err := s.templateRepo.GetByID(ctx, ci.TemplateID)
	if err != nil {
		return nil, apperror.NewWithErr(errcode.InternalError, err)
	}
	return &response.CouponIssueResponse{
		CouponID:        ci.ID,
		CouponCode:      ci.CouponCode,
		TemplateName:    t.Name,
		Type:            t.Type,
		DiscountValue:   t.DiscountValue,
		ThresholdAmount: t.ThresholdAmount,
		ValidStart:      ci.ValidStart,
		ValidEnd:        ci.ValidEnd,
		Status:          ci.Status,
	}, nil
}
