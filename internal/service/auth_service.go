package service

import (
	"context"
	"time"

	"hxcoupon/config"
	"hxcoupon/internal/pkg/apperror"
	"hxcoupon/internal/pkg/errcode"
	"hxcoupon/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	adminRepo  *repository.AdminUserRepo
	jwtCfg     config.JWTConfig
}

func NewAuthService(adminRepo *repository.AdminUserRepo, jwtCfg config.JWTConfig) *AuthService {
	return &AuthService{adminRepo: adminRepo, jwtCfg: jwtCfg}
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

type JWTClaims struct {
	UserID   uint64 `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

func (s *AuthService) Login(ctx context.Context, username, password string) (*TokenPair, error) {
	user, err := s.adminRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, apperror.New(errcode.AuthFailed)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, apperror.New(errcode.AuthFailed)
	}

	_ = s.adminRepo.UpdateLastLogin(ctx, user.ID)

	accessTTL := s.jwtCfg.AccessTokenTTL
	refreshTTL := s.jwtCfg.RefreshTokenTTL
	now := time.Now()

	accessClaims := &JWTClaims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(accessTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			Subject:   user.Username,
		},
	}

	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString([]byte(s.jwtCfg.Secret))
	if err != nil {
		return nil, apperror.NewWithErr(errcode.InternalError, err)
	}

	refreshClaims := &JWTClaims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(refreshTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			Subject:   user.Username,
		},
	}

	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(s.jwtCfg.Secret))
	if err != nil {
		return nil, apperror.NewWithErr(errcode.InternalError, err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(accessTTL.Seconds()),
	}, nil
}

func (s *AuthService) RefreshToken(refreshTokenStr string) (*TokenPair, error) {
	claims := &JWTClaims{}
	token, err := jwt.ParseWithClaims(refreshTokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(s.jwtCfg.Secret), nil
	})
	if err != nil || !token.Valid {
		return nil, apperror.New(errcode.AuthFailed)
	}

	accessTTL := s.jwtCfg.AccessTokenTTL
	now := time.Now()

	accessClaims := &JWTClaims{
		UserID:   claims.UserID,
		Username: claims.Username,
		Role:     claims.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(accessTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			Subject:   claims.Username,
		},
	}

	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString([]byte(s.jwtCfg.Secret))
	if err != nil {
		return nil, apperror.NewWithErr(errcode.InternalError, err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshTokenStr,
		ExpiresIn:    int64(accessTTL.Seconds()),
	}, nil
}

func (s *AuthService) ParseToken(tokenStr string) (*JWTClaims, error) {
	claims := &JWTClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(s.jwtCfg.Secret), nil
	})
	if err != nil || !token.Valid {
		return nil, apperror.New(errcode.AuthFailed)
	}
	return claims, nil
}
