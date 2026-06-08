package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"hxcoupon/internal/dto/request"
	"hxcoupon/internal/dto/response"
	"hxcoupon/internal/model"
	"hxcoupon/internal/pkg/apperror"
	"hxcoupon/internal/pkg/errcode"
	"hxcoupon/internal/repository"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AdminUserService struct {
	db          *gorm.DB
	adminRepo   *repository.AdminUserRepo
	storeRepo   *repository.StoreRepo
	credRepo    *repository.CredentialRepo
	approvalRepo *repository.ApprovalRecordRepo
}

func NewAdminUserService(db *gorm.DB, adminRepo *repository.AdminUserRepo, storeRepo *repository.StoreRepo, credRepo *repository.CredentialRepo, approvalRepo *repository.ApprovalRecordRepo) *AdminUserService {
	return &AdminUserService{
		db:          db,
		adminRepo:   adminRepo,
		storeRepo:   storeRepo,
		credRepo:    credRepo,
		approvalRepo: approvalRepo,
	}
}

// Register a new platform member. Consumer type is auto-approved.
func (s *AdminUserService) Register(ctx context.Context, req *request.RegisterRequest) (*model.AdminUser, error) {
	exists, err := s.adminRepo.ExistsByUsername(ctx, req.Username)
	if err != nil {
		return nil, apperror.NewWithErr(errcode.InternalError, err)
	}
	if exists {
		return nil, apperror.NewWithMsg(errcode.InvalidParams, "username already exists")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, apperror.NewWithErr(errcode.InternalError, err)
	}

	now := time.Now()
	user := &model.AdminUser{
		Username:     req.Username,
		PasswordHash: string(hash),
		Role:         "member",
		MemberType:   &req.MemberType,
		Status:       1,
		CompanyName:  &req.CompanyName,
		ContactName:  &req.ContactName,
		ContactPhone: &req.ContactPhone,
		RegisteredAt: &now,
	}

	if req.Email != "" {
		user.Email = &req.Email
	}
	if req.BusinessLicense != "" {
		user.BusinessLicense = &req.BusinessLicense
	}

	// Consumer is auto-approved; issuer/both require manual approval
	if req.MemberType == model.MemberTypeConsumer {
		user.ApprovalStatus = model.ApprovalApproved
		user.ApprovedAt = &now
	} else {
		user.ApprovalStatus = model.ApprovalPending
	}

	if err := s.adminRepo.Create(ctx, user); err != nil {
		return nil, apperror.NewWithErr(errcode.InternalError, err)
	}

	// Auto-create default store and credentials for auto-approved consumers
	if user.ApprovalStatus == model.ApprovalApproved {
		if err := s.createDefaultStore(ctx, user); err != nil {
			// Log but don't fail registration — admin can fix later
			_ = err
		}
	}

	return user, nil
}

// ApproveUser approves a pending user and creates their default store.
func (s *AdminUserService) ApproveUser(ctx context.Context, userID uint64, operatedBy uint64, reason *string) error {
	user, err := s.adminRepo.GetByID(ctx, userID)
	if err != nil {
		return apperror.New(errcode.NotFound)
	}
	if user.ApprovalStatus != model.ApprovalPending {
		return apperror.NewWithMsg(errcode.InvalidParams, "user is not in pending status")
	}

	now := time.Now()
	fields := map[string]interface{}{
		"approval_status": model.ApprovalApproved,
		"approved_at":     now,
		"approved_by":     operatedBy,
		"reject_reason":   nil,
	}
	if err := s.adminRepo.UpdateFields(ctx, userID, fields); err != nil {
		return apperror.NewWithErr(errcode.InternalError, err)
	}

	// Create default store + credentials
	if err := s.createDefaultStore(ctx, user); err != nil {
		return err
	}

	// Record approval
	rec := &model.ApprovalRecord{
		UserID:     userID,
		Action:     "approve",
		Reason:     reason,
		OperatedBy: operatedBy,
	}
	_ = s.approvalRepo.Create(ctx, rec)

	return nil
}

// RejectUser rejects a pending user.
func (s *AdminUserService) RejectUser(ctx context.Context, userID uint64, operatedBy uint64, reason string) error {
	user, err := s.adminRepo.GetByID(ctx, userID)
	if err != nil {
		return apperror.New(errcode.NotFound)
	}
	if user.ApprovalStatus != model.ApprovalPending {
		return apperror.NewWithMsg(errcode.InvalidParams, "user is not in pending status")
	}

	if err := s.adminRepo.UpdateFields(ctx, userID, map[string]interface{}{
		"approval_status": model.ApprovalRejected,
		"reject_reason":   reason,
	}); err != nil {
		return apperror.NewWithErr(errcode.InternalError, err)
	}

	rec := &model.ApprovalRecord{
		UserID:     userID,
		Action:     "reject",
		Reason:     &reason,
		OperatedBy: operatedBy,
	}
	_ = s.approvalRepo.Create(ctx, rec)

	return nil
}

// SuspendUser suspends an approved user.
func (s *AdminUserService) SuspendUser(ctx context.Context, userID uint64, operatedBy uint64, reason string) error {
	user, err := s.adminRepo.GetByID(ctx, userID)
	if err != nil {
		return apperror.New(errcode.NotFound)
	}
	if user.ApprovalStatus != model.ApprovalApproved {
		return apperror.NewWithMsg(errcode.InvalidParams, "user is not in approved status")
	}

	if err := s.adminRepo.UpdateFields(ctx, userID, map[string]interface{}{
		"approval_status": model.ApprovalSuspended,
	}); err != nil {
		return apperror.NewWithErr(errcode.InternalError, err)
	}

	rec := &model.ApprovalRecord{
		UserID:     userID,
		Action:     "suspend",
		Reason:     &reason,
		OperatedBy: operatedBy,
	}
	_ = s.approvalRepo.Create(ctx, rec)

	return nil
}

// UnsuspendUser restores a suspended user.
func (s *AdminUserService) UnsuspendUser(ctx context.Context, userID uint64, operatedBy uint64) error {
	user, err := s.adminRepo.GetByID(ctx, userID)
	if err != nil {
		return apperror.New(errcode.NotFound)
	}
	if user.ApprovalStatus != model.ApprovalSuspended {
		return apperror.NewWithMsg(errcode.InvalidParams, "user is not suspended")
	}

	if err := s.adminRepo.UpdateFields(ctx, userID, map[string]interface{}{
		"approval_status": model.ApprovalApproved,
	}); err != nil {
		return apperror.NewWithErr(errcode.InternalError, err)
	}

	rec := &model.ApprovalRecord{
		UserID:     userID,
		Action:     "unsuspend",
		OperatedBy: operatedBy,
	}
	_ = s.approvalRepo.Create(ctx, rec)

	return nil
}

// ListUsers returns a paginated list of platform users.
func (s *AdminUserService) ListUsers(ctx context.Context, req *request.UserListRequest) (*response.PaginatedData, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 20
	}

	var approvalStatus *int8
	if req.ApprovalStatus != nil {
		v := int8(*req.ApprovalStatus)
		approvalStatus = &v
	}
	users, total, err := s.adminRepo.List(ctx, req.Page, req.PageSize, req.Keyword, req.Role, req.MemberType, approvalStatus)
	if err != nil {
		return nil, apperror.NewWithErr(errcode.InternalError, err)
	}

	items := make([]response.UserResponse, len(users))
	for i, u := range users {
		items[i] = *toUserResponse(&u)
	}

	return &response.PaginatedData{
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
		Items:    items,
	}, nil
}

// GetUser returns a single user by ID.
func (s *AdminUserService) GetUser(ctx context.Context, id uint64) (*response.UserResponse, error) {
	user, err := s.adminRepo.GetByID(ctx, id)
	if err != nil {
		return nil, apperror.New(errcode.NotFound)
	}
	return toUserResponse(user), nil
}

// GetProfile returns the current user's profile.
func (s *AdminUserService) GetProfile(ctx context.Context, userID uint64) (*response.ProfileResponse, error) {
	user, err := s.adminRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, apperror.New(errcode.NotFound)
	}
	return &response.ProfileResponse{
		ID:             user.ID,
		Username:       user.Username,
		Role:           user.Role,
		MemberType:     user.MemberType,
		ApprovalStatus: user.ApprovalStatus,
		CompanyName:    user.CompanyName,
		ContactName:    user.ContactName,
		ContactPhone:   user.ContactPhone,
		Email:          user.Email,
		RejectReason:   user.RejectReason,
	}, nil
}

// UpdateProfile updates the current user's profile.
func (s *AdminUserService) UpdateProfile(ctx context.Context, userID uint64, req *request.UpdateProfileRequest) error {
	fields := map[string]interface{}{}
	if req.ContactName != nil {
		fields["contact_name"] = *req.ContactName
	}
	if req.ContactPhone != nil {
		fields["contact_phone"] = *req.ContactPhone
	}
	if req.Email != nil {
		fields["email"] = *req.Email
	}
	if req.CompanyName != nil {
		fields["company_name"] = *req.CompanyName
	}

	// Password change requires old password verification
	if req.NewPassword != nil {
		if req.OldPassword == nil {
			return apperror.NewWithMsg(errcode.InvalidParams, "old password required")
		}
		user, err := s.adminRepo.GetByID(ctx, userID)
		if err != nil {
			return apperror.NewWithErr(errcode.NotFound, err)
		}
		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(*req.OldPassword)); err != nil {
			return apperror.NewWithMsg(errcode.InvalidParams, "old password is incorrect")
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(*req.NewPassword), bcrypt.DefaultCost)
		if err != nil {
			return apperror.NewWithErr(errcode.InternalError, err)
		}
		fields["password_hash"] = string(hash)
	}

	if len(fields) == 0 {
		return nil
	}
	return s.adminRepo.UpdateFields(ctx, userID, fields)
}

// createDefaultStore creates a default application and API credentials for a newly approved user.
func (s *AdminUserService) createDefaultStore(ctx context.Context, user *model.AdminUser) error {
	uid := user.ID
	company := ""
	if user.CompanyName != nil {
		company = *user.CompanyName
	}
	if company == "" {
		company = user.Username
	}

	// Generate unique 5-char code
	code := generateStoreCode()
	appID := fmt.Sprintf("app_%s_%d", hex.EncodeToString(make([]byte, 4)), uid%10000)

	store := &model.Store{
		Name:        company + " - Default",
		Code:        code,
		AppID:       appID,
		Type:        "api",
		Status:      1,
		UserID:      &uid,
		ContactName: "",
		ContactPhone: "",
		Remark:      "auto-created on approval",
	}

	if err := s.storeRepo.Create(ctx, store); err != nil {
		return apperror.NewWithErr(errcode.InternalError, err)
	}

	// Generate API credentials
	return s.createCredentials(ctx, store.ID)
}

func (s *AdminUserService) createCredentials(ctx context.Context, storeID uint64) error {
	appKeyBytes := make([]byte, 16)
	appSecretBytes := make([]byte, 32)
	rand.Read(appKeyBytes)
	rand.Read(appSecretBytes)

	appKey := "ak_" + hex.EncodeToString(appKeyBytes)
	rawSecret := "sk_" + hex.EncodeToString(appSecretBytes)

	hashedSecret, err := bcrypt.GenerateFromPassword([]byte(rawSecret), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	cred := &model.StoreAPICredential{
		StoreID:   storeID,
		AppKey:    appKey,
		AppSecret: string(hashedSecret),
		Status:    1,
	}

	return s.credRepo.Create(ctx, cred)
}

func generateStoreCode() string {
	b := make([]byte, 3)
	rand.Read(b)
	return fmt.Sprintf("S%04X", hex.EncodeToString(b)[:4])
}

func toUserResponse(u *model.AdminUser) *response.UserResponse {
	return &response.UserResponse{
		ID:              u.ID,
		Username:        u.Username,
		Role:            u.Role,
		MemberType:      u.MemberType,
		Status:          u.Status,
		ApprovalStatus:  u.ApprovalStatus,
		CompanyName:     u.CompanyName,
		ContactName:     u.ContactName,
		ContactPhone:    u.ContactPhone,
		Email:           u.Email,
		RejectReason:    u.RejectReason,
		BusinessLicense: u.BusinessLicense,
		RegisteredAt:    u.RegisteredAt,
		ApprovedAt:      u.ApprovedAt,
		LastLoginAt:     u.LastLoginAt,
		CreatedAt:       u.CreatedAt,
	}
}
