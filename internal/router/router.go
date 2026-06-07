package router

import (
	"context"
	"time"

	"hxcoupon/config"
	"hxcoupon/internal/handler/admin"
	"hxcoupon/internal/handler/openapi"
	"hxcoupon/internal/middleware"
	"hxcoupon/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Services struct {
	Auth       *service.AuthService
	Store      *service.StoreService
	Template   *service.TemplateService
	Coupon     *service.CouponService
	Statistics *service.StatisticsService
}

type Handlers struct {
	AdminAuth       *admin.AuthHandler
	AdminStore      *admin.StoreHandler
	AdminTemplate   *admin.TemplateHandler
	AdminStatistics *admin.StatisticsHandler
	OpenCoupon      *openapi.CouponHandler
}

func NewHandlers(svc *Services) *Handlers {
	return &Handlers{
		AdminAuth:       admin.NewAuthHandler(svc.Auth),
		AdminStore:      admin.NewStoreHandler(svc.Store),
		AdminTemplate:   admin.NewTemplateHandler(svc.Template),
		AdminStatistics: admin.NewStatisticsHandler(svc.Statistics),
		OpenCoupon:      openapi.NewCouponHandler(svc.Coupon),
	}
}

func Setup(r *gin.Engine, h *Handlers, svc *Services, cfg *config.Config, logger *zap.Logger) {
	r.Use(middleware.Logger(logger))
	r.Use(middleware.Recovery(logger))
	r.Use(middleware.CORS())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	api := r.Group("/api/v1")

	// --- Admin auth (public) ---
	api.POST("/admin/login", h.AdminAuth.Login)

	// --- Admin routes (JWT protected) ---
	adminGroup := api.Group("/admin")
	adminGroup.Use(middleware.JWTAuth(svc.Auth))
	{
		adminGroup.POST("/refresh", h.AdminAuth.Refresh)

		adminGroup.GET("/stores", h.AdminStore.List)
		adminGroup.GET("/stores/:id", h.AdminStore.Get)
		adminGroup.POST("/stores", h.AdminStore.Create)
		adminGroup.PUT("/stores/:id", h.AdminStore.Update)
		adminGroup.PATCH("/stores/:id/status", h.AdminStore.UpdateStatus)
		adminGroup.POST("/stores/:id/credentials", h.AdminStore.GenerateCredentials)

		adminGroup.GET("/templates", h.AdminTemplate.List)
		adminGroup.GET("/templates/:id", h.AdminTemplate.Get)
		adminGroup.POST("/templates", h.AdminTemplate.Create)
		adminGroup.PUT("/templates/:id", h.AdminTemplate.Update)
		adminGroup.PATCH("/templates/:id/status", h.AdminTemplate.UpdateStatus)
		adminGroup.DELETE("/templates/:id", h.AdminTemplate.Delete)

		adminGroup.GET("/statistics/overview", h.AdminStatistics.Overview)
		adminGroup.GET("/statistics/trend", h.AdminStatistics.Trend)
	}

	// --- Open API (HMAC store auth + rate limiting) ---
	credentialGetter := func(appKey string) (uint64, string, error) {
		ctx := context.Background()
		id, err := svc.Coupon.VerifyStoreCredentials(ctx, appKey)
		if err != nil {
			return 0, "", err
		}
		secret, err := svc.Coupon.GetStoreSecret(ctx, appKey)
		return id, secret, err
	}

	storeAuthMW := middleware.StoreAuth(cfg.StoreAuth, credentialGetter)

	openGroup := api.Group("/coupons")
	openGroup.Use(storeAuthMW)
	{
		openGroup.POST("/issue",
			middleware.RateLimit("rl:issue", 100, 1*time.Minute),
			h.OpenCoupon.Issue,
		)

		openGroup.GET("/available",
			middleware.RateLimit("rl:query", 300, 1*time.Minute),
			h.OpenCoupon.Available,
		)

		openGroup.GET("/user",
			middleware.RateLimit("rl:query", 300, 1*time.Minute),
			h.OpenCoupon.ListByUser,
		)

		openGroup.GET("/:coupon_code",
			middleware.RateLimit("rl:query", 300, 1*time.Minute),
			h.OpenCoupon.Detail,
		)

		openGroup.POST("/consume",
			middleware.RateLimit("rl:action", 200, 1*time.Minute),
			h.OpenCoupon.Consume,
		)

		openGroup.POST("/refund",
			middleware.RateLimit("rl:action", 200, 1*time.Minute),
			h.OpenCoupon.Refund,
		)
	}
}
