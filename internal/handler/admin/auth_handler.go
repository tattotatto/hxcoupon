package admin

import (
	"net/http"

	"hxcoupon/internal/dto/request"
	"hxcoupon/internal/dto/response"
	"hxcoupon/internal/pkg/apperror"
	"hxcoupon/internal/pkg/errcode"
	"hxcoupon/internal/service"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req request.AdminLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(errcode.InvalidParams, err.Error()))
		return
	}

	tokens, err := h.authService.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		appErr, ok := err.(*apperror.AppError)
		if ok {
			c.JSON(http.StatusUnauthorized, response.Error(appErr.Code, appErr.Message))
			return
		}
		c.JSON(http.StatusInternalServerError, response.Error(errcode.InternalError, err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(response.LoginResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresIn:    tokens.ExpiresIn,
		TokenType:    "Bearer",
	}))
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var req request.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(errcode.InvalidParams, err.Error()))
		return
	}

	tokens, err := h.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, response.Error(errcode.AuthFailed, "invalid refresh token"))
		return
	}

	c.JSON(http.StatusOK, response.Success(response.RefreshResponse{
		AccessToken: tokens.AccessToken,
		ExpiresIn:   tokens.ExpiresIn,
		TokenType:   "Bearer",
	}))
}
