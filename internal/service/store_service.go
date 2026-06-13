package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"

	"hxcoupon/internal/dto/request"
	"hxcoupon/internal/dto/response"
	"hxcoupon/internal/model"
	"hxcoupon/internal/pkg/apperror"
	"hxcoupon/internal/pkg/errcode"
	redisutil "hxcoupon/internal/pkg/redis"
	"hxcoupon/internal/repository"

	"gorm.io/gorm"
)

type StoreService struct {
	db             *gorm.DB
	storeRepo      *repository.StoreRepo
	credentialRepo *repository.CredentialRepo
}

func NewStoreService(db *gorm.DB, storeRepo *repository.StoreRepo, credRepo *repository.CredentialRepo) *StoreService {
	return &StoreService{db: db, storeRepo: storeRepo, credentialRepo: credRepo}
}

func (s *StoreService) Create(ctx context.Context, req *request.CreateStoreRequest) (*response.StoreWithCredentialsResponse, error) {
	// Auto-generate unique 5-char alphanumeric store code
	code, err := s.generateStoreCode(ctx)
	if err != nil {
		return nil, apperror.NewWithMsg(errcode.InternalError, "failed to generate store code")
	}

	store := &model.Store{
		Name:         req.Name,
		Code:         code,
		AppID:        req.AppID,
		Type:         req.Type,
		Status:       1,
		ContactName:  req.ContactName,
		ContactPhone: req.ContactPhone,
		Remark:       req.Remark,
	}
	if req.MpAppID != "" {
		store.MpAppID = &req.MpAppID
	}
	if req.MpPagePath != "" {
		store.MpPagePath = &req.MpPagePath
	}

	if err := s.storeRepo.Create(ctx, store); err != nil {
		return nil, apperror.NewWithErr(errcode.InternalError, err)
	}

	appKey, appSecret, err := s.generateCredentials(ctx, store.ID)
	if err != nil {
		return nil, err
	}

	resp := &response.StoreWithCredentialsResponse{
		StoreResponse: *response.ToStoreResponse(store),
		Credentials: &response.CredentialResponse{
			AppKey:    appKey,
			AppSecret: appSecret,
		},
	}
	return resp, nil
}

func (s *StoreService) GetByID(ctx context.Context, id uint64) (*response.StoreResponse, error) {
	store, err := s.getStoreCached(ctx, id)
	if err != nil {
		return nil, apperror.New(errcode.NotFound)
	}
	return response.ToStoreResponse(store), nil
}

// getStoreCached returns store model from cache or DB.
func (s *StoreService) getStoreCached(ctx context.Context, id uint64) (*model.Store, error) {
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

func (s *StoreService) List(ctx context.Context, page, pageSize int) (*response.PaginatedData, error) {
	stores, total, err := s.storeRepo.List(ctx, page, pageSize)
	if err != nil {
		return nil, apperror.NewWithErr(errcode.InternalError, err)
	}

	items := make([]response.StoreResponse, len(stores))
	for i, s := range stores {
		items[i] = *response.ToStoreResponse(&s)
	}

	return &response.PaginatedData{
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Items:    items,
	}, nil
}

func (s *StoreService) Update(ctx context.Context, id uint64, req *request.UpdateStoreRequest) (*response.StoreResponse, error) {
	store, err := s.getStoreCached(ctx, id)
	if err != nil {
		return nil, apperror.New(errcode.NotFound)
	}

	store.Name = req.Name
	store.ContactName = req.ContactName
	store.ContactPhone = req.ContactPhone
	store.Remark = req.Remark
	if req.MpAppID != "" {
		store.MpAppID = &req.MpAppID
	} else {
		store.MpAppID = nil
	}
	if req.MpPagePath != "" {
		store.MpPagePath = &req.MpPagePath
	} else {
		store.MpPagePath = nil
	}

	if err := s.storeRepo.Update(ctx, store); err != nil {
		return nil, apperror.NewWithErr(errcode.InternalError, err)
	}
	redisutil.CacheDelete(ctx, fmt.Sprintf("%s%d", redisutil.KeyStore, id))
	return response.ToStoreResponse(store), nil
}

func (s *StoreService) UpdateStatus(ctx context.Context, id uint64, status int8) error {
	if _, err := s.getStoreCached(ctx, id); err != nil {
		return apperror.New(errcode.NotFound)
	}
	if err := s.storeRepo.UpdateStatus(ctx, id, status); err != nil {
		return err
	}
	redisutil.CacheDelete(ctx, fmt.Sprintf("%s%d", redisutil.KeyStore, id))
	return nil
}

// ListActive returns all active stores for dropdown selects.
func (s *StoreService) ListActive(ctx context.Context) ([]model.Store, error) {
	return s.storeRepo.ListActive(ctx)
}

// Delete soft-deletes a store by ID (sets status to -1).
func (s *StoreService) Delete(ctx context.Context, id uint64) error {
	if _, err := s.getStoreCached(ctx, id); err != nil {
		return apperror.New(errcode.NotFound)
	}
	if err := s.storeRepo.UpdateStatus(ctx, id, -1); err != nil {
		return err
	}
	redisutil.CacheDelete(ctx, fmt.Sprintf("%s%d", redisutil.KeyStore, id))
	return nil
}

func (s *StoreService) GenerateCredentials(ctx context.Context, storeID uint64) (*response.CredentialResponse, error) {
	store, err := s.getStoreCached(ctx, storeID)
	if err != nil {
		return nil, apperror.New(errcode.NotFound)
	}

	// Check for existing credentials to preserve the AppKey
	oldCred, _ := s.credentialRepo.GetByStoreID(ctx, store.ID)
	var appKey string
	if oldCred != nil {
		appKey = oldCred.AppKey
		// Invalidate cached credential for old key
		redisutil.CacheDelete(ctx, fmt.Sprintf("%s%s", redisutil.KeyCredential, oldCred.AppKey))
		// Disable old credential
		_ = s.credentialRepo.DisableByStoreID(ctx, store.ID)
	}

	appKey, appSecret, err := s.generateSecret(ctx, store.ID, appKey)
	if err != nil {
		return nil, err
	}

	return &response.CredentialResponse{
		AppKey:    appKey,
		AppSecret: appSecret,
	}, nil
}

// generateSecret creates a new credential with a fresh app_secret.
// If existingAppKey is empty, a new AppKey is generated; otherwise the existing AppKey is reused.
// Returns the appKey and the raw secret.
func (s *StoreService) generateSecret(ctx context.Context, storeID uint64, existingAppKey string) (string, string, error) {
	appKey := existingAppKey
	if appKey == "" {
		appKeyBytes := make([]byte, 16)
		if _, err := rand.Read(appKeyBytes); err != nil {
			return "", "", apperror.NewWithErr(errcode.InternalError, err)
		}
		appKey = "ak_" + hex.EncodeToString(appKeyBytes)
	}

	appSecretBytes := make([]byte, 32)
	if _, err := rand.Read(appSecretBytes); err != nil {
		return "", "", apperror.NewWithErr(errcode.InternalError, err)
	}
	rawSecret := "sk_" + hex.EncodeToString(appSecretBytes)

	cred := &model.StoreAPICredential{
		StoreID:   storeID,
		AppKey:    appKey,
		AppSecret: rawSecret,
		Status:    1,
	}

	if err := s.credentialRepo.Create(ctx, cred); err != nil {
		return "", "", apperror.NewWithErr(errcode.InternalError, err)
	}

	return appKey, rawSecret, nil
}

func (s *StoreService) generateCredentials(ctx context.Context, storeID uint64) (string, string, error) {
	return s.generateSecret(ctx, storeID, "")
}

// generateStoreCode generates a unique 5-character alphanumeric code.
// It retries up to 10 times on collision.
func (s *StoreService) generateStoreCode(ctx context.Context) (string, error) {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	const codeLen = 5
	const maxRetries = 10

	for i := 0; i < maxRetries; i++ {
		code := make([]byte, codeLen)
		for j := range code {
			n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
			if err != nil {
				return "", err
			}
			code[j] = charset[n.Int64()]
		}
		codeStr := string(code)

		existing, _ := s.storeRepo.GetByCode(ctx, codeStr)
		if existing == nil {
			return codeStr, nil
		}
	}

	return "", fmt.Errorf("failed to generate unique store code after %d retries", maxRetries)
}
