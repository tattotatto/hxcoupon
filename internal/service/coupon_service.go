package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"hxcoupon/internal/dto/response"
	"hxcoupon/internal/model"
	"hxcoupon/internal/pkg/apperror"
	"hxcoupon/internal/pkg/couponcode"
	"hxcoupon/internal/pkg/errcode"
	redisutil "hxcoupon/internal/pkg/redis"
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

// Issue issues a coupon to a user. Idempotent via auto-generated idempotency_key.
func (s *CouponService) Issue(ctx context.Context, sourceStoreID uint64, templateID uint64, userPhone string) (*response.CouponIssueResponse, error) {
	// Generate idempotency key server-side
	idempotencyKey := s.generateIdempotencyKey()
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
			return apperror.NewWithMsg(errcode.CouponNotApply, "优惠券模板未启用")
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
		QrCodeURL:       s.resolveQrCodeURL(ctx, sourceStoreID),
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
		t, _ := s.getTemplateCached(ctx, ci.TemplateID)
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
			items[i].MpAppID, items[i].MpPagePath = s.resolveMpInfo(ctx, ci.TemplateID)
			items[i].Stackable = t.Stackable
		}
		items[i].QrCodeURL = s.resolveQrCodeURL(ctx, ci.SourceStoreID)
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
			return apperror.NewWithMsg(errcode.Forbidden, "优惠券不属于该用户")
		}

		// Validate status
		if ci.Status != "unused" {
			if ci.Status == "used" {
				return apperror.New(errcode.CouponUsed)
			}
			if ci.Status == "expired" {
				return apperror.New(errcode.CouponExpired)
			}
			return apperror.NewWithMsg(errcode.CouponNotApply, "优惠券状态异常: "+ci.Status)
		}

		// Validate validity period
		now := time.Now()
		if now.Before(ci.ValidStart) || now.After(ci.ValidEnd) {
			return apperror.New(errcode.CouponExpired)
		}

		// Validate store applicability
		t, err := s.getTemplateCached(ctx, ci.TemplateID)
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
			return apperror.NewWithMsg(errcode.Forbidden, "优惠券不属于该用户")
		}

		// Validate status
		if ci.Status != "used" {
			return apperror.NewWithMsg(errcode.CouponNotApply, "优惠券不在已使用状态")
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
		MpAppID         string    `json:"mp_appid,omitempty"`
		MpPagePath      string    `json:"mp_page_path,omitempty"`
		QrCodeURL       string    `json:"qr_code_url"`
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
		t, _ := s.getTemplateCached(ctx, ci.TemplateID)
		if t != nil {
			items[i].TemplateName = t.Name
			items[i].Type = t.Type
			items[i].DiscountValue = t.DiscountValue
			items[i].ThresholdAmount = t.ThresholdAmount
		}
		items[i].MpAppID, items[i].MpPagePath = s.resolveMpInfo(ctx, ci.TemplateID)
		items[i].QrCodeURL = s.resolveQrCodeURL(ctx, ci.SourceStoreID)
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

	t, _ := s.getTemplateCached(ctx, ci.TemplateID)
	templateName := ""
	if t != nil {
		templateName = t.Name
	}

	sourceStore, _ := s.getStoreCached(ctx, ci.SourceStoreID)
	sourceStoreName := ""
	if sourceStore != nil {
		sourceStoreName = sourceStore.Name
	}

	usedStoreName := ""
	if ci.UsedAtStoreID != nil {
		us, _ := s.getStoreCached(ctx, *ci.UsedAtStoreID)
		if us != nil {
			usedStoreName = us.Name
		}
	}

	records, _ := s.usageRecordRepo.ListByCouponID(ctx, ci.ID)
	recordBriefs := make([]response.CouponUsageRecordBrief, len(records))
	for i, r := range records {
		rs, _ := s.getStoreCached(ctx, r.StoreID)
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
		resp.MpAppID, resp.MpPagePath = s.resolveMpInfo(ctx, ci.TemplateID)
	}
	resp.QrCodeURL = s.resolveQrCodeURL(ctx, ci.SourceStoreID)
	return resp, nil
}

// VerifyStoreCredentials verifies store HMAC credentials (with Redis cache).
func (s *CouponService) VerifyStoreCredentials(ctx context.Context, appKey string) (uint64, error) {
	// Check Redis cache first
	cacheKey := redisutil.KeyCredential + appKey
	var cached struct {
		StoreID   uint64 `json:"store_id"`
		AppSecret string `json:"app_secret"`
	}
	if redisutil.CacheGet(ctx, cacheKey, &cached) {
		return cached.StoreID, nil
	}

	cred, err := s.credentialRepo.GetByAppKey(ctx, appKey)
	if err != nil {
		return 0, apperror.New(errcode.StoreNotAuth)
	}

	// Populate cache
	redisutil.CacheSet(ctx, cacheKey, map[string]interface{}{
		"store_id":   cred.StoreID,
		"app_secret": cred.AppSecret,
	}, redisutil.TTLCredential)

	return cred.StoreID, nil
}

// GetStoreSecret returns the raw app_secret for HMAC verification (with Redis cache).
func (s *CouponService) GetStoreSecret(ctx context.Context, appKey string) (string, error) {
	// Check Redis cache first
	cacheKey := redisutil.KeyCredential + appKey
	var cached struct {
		StoreID   uint64 `json:"store_id"`
		AppSecret string `json:"app_secret"`
	}
	if redisutil.CacheGet(ctx, cacheKey, &cached) {
		return cached.AppSecret, nil
	}

	cred, err := s.credentialRepo.GetByAppKey(ctx, appKey)
	if err != nil {
		return "", apperror.New(errcode.StoreNotAuth)
	}

	// Populate cache
	redisutil.CacheSet(ctx, cacheKey, map[string]interface{}{
		"store_id":   cred.StoreID,
		"app_secret": cred.AppSecret,
	}, redisutil.TTLCredential)

	return cred.AppSecret, nil
}

// InvalidateCredentialCache removes cached credential for an appKey.
func (s *CouponService) InvalidateCredentialCache(ctx context.Context, appKey string) {
	cacheKey := fmt.Sprintf("%s%s", redisutil.KeyCredential, appKey)
	redisutil.CacheDelete(ctx, cacheKey)
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
		store, err := s.getStoreCached(ctx, storeIDs[0])
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

// getTemplateCached returns a cached template by ID to avoid N+1 DB queries.
func (s *CouponService) getTemplateCached(ctx context.Context, id uint64) (*model.CouponTemplate, error) {
	cacheKey := fmt.Sprintf("%s%d", redisutil.KeyTemplate, id)
	var cached model.CouponTemplate
	if redisutil.CacheGet(ctx, cacheKey, &cached) {
		return &cached, nil
	}
	t, err := s.templateRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	redisutil.CacheSet(ctx, cacheKey, *t, redisutil.TTLTemplate)
	return t, nil
}

// getStoreCached returns a cached store by ID to avoid N+1 DB queries.
func (s *CouponService) getStoreCached(ctx context.Context, id uint64) (*model.Store, error) {
	cacheKey := fmt.Sprintf("%s%d", redisutil.KeyStore, id)
	var cached model.Store
	if redisutil.CacheGet(ctx, cacheKey, &cached) {
		return &cached, nil
	}
	store, err := s.storeRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	redisutil.CacheSet(ctx, cacheKey, *store, redisutil.TTLStore)
	return store, nil
}

func (s *CouponService) buildIssueResponse(ctx context.Context, ci *model.CouponInstance) (*response.CouponIssueResponse, error) {
	t, err := s.getTemplateCached(ctx, ci.TemplateID)
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
		QrCodeURL:       s.resolveQrCodeURL(ctx, ci.SourceStoreID),
	}, nil
}

// GetByID returns a single coupon instance by ID.
func (s *CouponService) GetByID(ctx context.Context, id uint64) (*response.CouponDetailResponse, error) {
	ci, err := s.instanceRepo.GetByID(ctx, id)
	if err != nil {
		return nil, apperror.New(errcode.NotFound)
	}
	return s.buildDetail(ctx, ci)
}

// ListAdminRecords returns paginated coupon instances for admin view.
func (s *CouponService) ListAdminRecords(ctx context.Context, page, pageSize int) (*response.PaginatedData, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	instances, total, err := s.instanceRepo.ListAll(ctx, offset, pageSize)
	if err != nil {
		return nil, apperror.NewWithErr(errcode.InternalError, err)
	}

	type recordItem struct {
		ID              uint64     `json:"id"`
		CouponCode      string     `json:"coupon_code"`
		TemplateName    string     `json:"template_name"`
		Type            string     `json:"type"`
		DiscountValue   float64    `json:"discount_value"`
		UserPhone       string     `json:"user_phone"`
		Status          string     `json:"status"`
		SourceStoreName string     `json:"source_store_name"`
		ReceiveTime     time.Time  `json:"receive_time"`
		UseTime         *time.Time `json:"use_time"`
	}

	items := make([]recordItem, len(instances))
	for i, ci := range instances {
		items[i] = recordItem{
			ID:         ci.ID,
			CouponCode: ci.CouponCode,
			UserPhone:  ci.UserPhone,
			Status:     ci.Status,
			ReceiveTime: ci.ReceiveTime,
			UseTime:    ci.UseTime,
		}
		if t, _ := s.getTemplateCached(ctx, ci.TemplateID); t != nil {
			items[i].TemplateName = t.Name
			items[i].Type = t.Type
			items[i].DiscountValue = t.DiscountValue
		}
		if s, _ := s.getStoreCached(ctx, ci.SourceStoreID); s != nil {
			items[i].SourceStoreName = s.Name
		}
	}

	return &response.PaginatedData{
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Items:    items,
	}, nil
}

// ListConsumeRecords returns paginated usage records for admin view.
func (s *CouponService) ListConsumeRecords(ctx context.Context, page, pageSize int) (*response.PaginatedData, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	f := repository.UsageRecordListFilter{
		Action:   "consume",
		Page:     page,
		PageSize: pageSize,
	}

	records, total, err := s.usageRecordRepo.List(ctx, f)
	if err != nil {
		return nil, apperror.NewWithErr(errcode.InternalError, err)
	}

	type consumeRecordItem struct {
		ID              uint64    `json:"id"`
		CouponID        uint64    `json:"coupon_id"`
		UserPhone       string    `json:"user_phone"`
		StoreName       string    `json:"store_name"`
		Action          string    `json:"action"`
		OrderInfo       *model.JSON `json:"order_info,omitempty"`
		CreatedAt       time.Time `json:"created_at"`
	}

	items := make([]consumeRecordItem, len(records))
	for i, r := range records {
		items[i] = consumeRecordItem{
			ID:        r.ID,
			CouponID:  r.CouponID,
			UserPhone: r.UserPhone,
			Action:    r.Action,
			OrderInfo: r.OrderInfo,
			CreatedAt: r.CreatedAt,
		}
		if s, _ := s.getStoreCached(ctx, r.StoreID); s != nil {
			items[i].StoreName = s.Name
		}
	}

	return &response.PaginatedData{
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Items:    items,
	}, nil
}

// generateIdempotencyKey creates a UUID v4-style idempotency key.
func (s *CouponService) generateIdempotencyKey() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	// Set version 4 and variant bits
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%s-%s-%s-%s-%s",
		hex.EncodeToString(b[0:4]),
		hex.EncodeToString(b[4:6]),
		hex.EncodeToString(b[6:8]),
		hex.EncodeToString(b[8:10]),
		hex.EncodeToString(b[10:16]),
	)
}

// Default QR code URL used as fallback when a store has no QR code configured.
const defaultQrCodeURL = "https://ynhx.oss-cn-chengdu.aliyuncs.com/qrcodes/store_1_1782303867445088958.jpg"

// resolveQrCodeURL returns the QR code URL for a store, or the default if not configured.
func (s *CouponService) resolveQrCodeURL(ctx context.Context, storeID uint64) string {
	store, err := s.getStoreCached(ctx, storeID)
	if err != nil || store.QrCodeURL == nil {
		return defaultQrCodeURL
	}
	return *store.QrCodeURL
}

// resolveMpInfo returns mini-program redirect info for a template's first applicable store.
func (s *CouponService) resolveMpInfo(ctx context.Context, templateID uint64) (mpAppID, mpPagePath string) {
	t, err := s.getTemplateCached(ctx, templateID)
	if err != nil {
		return "", ""
	}
	var storeIDs []uint64
	if t.ApplicableScope == "specific" {
		storeIDs, _ = s.templateStoreRepo.GetStoreIDsByTemplateID(ctx, templateID)
	}
	if len(storeIDs) == 0 {
		stores, err := s.storeRepo.ListActive(ctx)
		if err == nil && len(stores) > 0 {
			storeIDs = []uint64{stores[0].ID}
		}
	}
	for _, sid := range storeIDs {
		store, err := s.getStoreCached(ctx, sid)
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

func (s *CouponService) buildDetail(ctx context.Context, ci *model.CouponInstance) (*response.CouponDetailResponse, error) {
	t, _ := s.getTemplateCached(ctx, ci.TemplateID)
	templateName := ""
	if t != nil {
		templateName = t.Name
	}

	sourceStore, _ := s.getStoreCached(ctx, ci.SourceStoreID)
	sourceStoreName := ""
	if sourceStore != nil {
		sourceStoreName = sourceStore.Name
	}

	usedStoreName := ""
	if ci.UsedAtStoreID != nil {
		us, _ := s.getStoreCached(ctx, *ci.UsedAtStoreID)
		if us != nil {
			usedStoreName = us.Name
		}
	}

	records, _ := s.usageRecordRepo.ListByCouponID(ctx, ci.ID)
	recordBriefs := make([]response.CouponUsageRecordBrief, len(records))
	for i, r := range records {
		rs, _ := s.getStoreCached(ctx, r.StoreID)
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
		resp.MpAppID, resp.MpPagePath = s.resolveMpInfo(ctx, ci.TemplateID)
	}
	resp.QrCodeURL = s.resolveQrCodeURL(ctx, ci.SourceStoreID)
	return resp, nil
}
