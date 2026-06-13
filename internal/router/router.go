package router

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
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
	Auth        *service.AuthService
	Store       *service.StoreService
	Template    *service.TemplateService
	Coupon      *service.CouponService
	Statistics  *service.StatisticsService
	AdminUser   *service.AdminUserService
	Report      *service.ReportService
}

type Handlers struct {
	AdminAuth       *admin.AuthHandler
	AdminStore      *admin.StoreHandler
	AdminTemplate   *admin.TemplateHandler
	AdminStatistics *admin.StatisticsHandler
	AdminUser       *admin.UserHandler
	AdminReport     *admin.ReportHandler
	AdminCoupon     *admin.CouponHandler
	OpenCoupon      *openapi.CouponHandler
}

func NewHandlers(svc *Services) *Handlers {
	return &Handlers{
		AdminAuth:       admin.NewAuthHandler(svc.Auth, svc.AdminUser),
		AdminStore:      admin.NewStoreHandler(svc.Store),
		AdminTemplate:   admin.NewTemplateHandler(svc.Template),
		AdminStatistics: admin.NewStatisticsHandler(svc.Statistics),
		AdminUser:       admin.NewUserHandler(svc.AdminUser),
		AdminReport:     admin.NewReportHandler(svc.Report),
		AdminCoupon:     admin.NewCouponHandler(svc.Coupon),
		OpenCoupon:      openapi.NewCouponHandler(svc.Coupon, svc.Template),
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

	// --- Public routes ---
	api.POST("/auth/register", h.AdminAuth.Register)
	api.POST("/admin/login", h.AdminAuth.Login)
	api.POST("/admin/refresh", h.AdminAuth.Refresh)

	// --- Admin routes (JWT protected) ---
	adminGroup := api.Group("/admin")
	adminGroup.Use(middleware.JWTAuth(svc.Auth))
	{
		// Profile (all authenticated users)
		adminGroup.GET("/profile", h.AdminUser.GetProfile)
		adminGroup.PUT("/profile", h.AdminUser.UpdateProfile)

		// User management (super_admin + manage_users permission)
		userGroup := adminGroup.Group("/users")
		userGroup.Use(middleware.RequirePermission("manage_users"))
		{
			userGroup.GET("", h.AdminUser.List)
			userGroup.GET("/:id", h.AdminUser.Get)
			userGroup.POST("/:id/approve", h.AdminUser.ApproveUser)
			userGroup.POST("/:id/reject", h.AdminUser.RejectUser)
			userGroup.POST("/:id/suspend", h.AdminUser.SuspendUser)
			userGroup.POST("/:id/unsuspend", h.AdminUser.UnsuspendUser)
		}

		// Store management (existing routes, JWT only for now)
		adminGroup.GET("/stores/options", h.AdminStore.Options)
		adminGroup.GET("/stores", h.AdminStore.List)
		adminGroup.GET("/stores/:id", h.AdminStore.Get)
		adminGroup.POST("/stores", h.AdminStore.Create)
		adminGroup.PUT("/stores/:id", h.AdminStore.Update)
		adminGroup.PATCH("/stores/:id/status", h.AdminStore.UpdateStatus)
		adminGroup.POST("/stores/:id/credentials", h.AdminStore.GenerateCredentials)

		// Template management
		adminGroup.GET("/templates", h.AdminTemplate.List)
		adminGroup.GET("/templates/:id", h.AdminTemplate.Get)
		// Template write operations require manage_templates permission (用券方/综合 only)
		templateWrite := adminGroup.Group("")
		templateWrite.Use(middleware.RequirePermission("manage_templates"))
		{
			templateWrite.POST("/templates", h.AdminTemplate.Create)
			templateWrite.PUT("/templates/:id", h.AdminTemplate.Update)
			templateWrite.PATCH("/templates/:id/status", h.AdminTemplate.UpdateStatus)
			templateWrite.DELETE("/templates/:id", h.AdminTemplate.Delete)
		}

		// Statistics (existing routes)
		adminGroup.GET("/statistics/overview", h.AdminStatistics.Overview)
		adminGroup.GET("/statistics/trend", h.AdminStatistics.Trend)

		// Browse templates (consumer-facing)
		adminGroup.GET("/browse/templates", h.AdminTemplate.Browse)
		adminGroup.GET("/browse/templates/:id", h.AdminTemplate.BrowseDetail)

		// App management (member's own stores)
		appsGroup := adminGroup.Group("/apps")
		appsGroup.Use(middleware.RequireApproval())
		{
			appsGroup.GET("", h.AdminStore.List)
			appsGroup.POST("", h.AdminStore.Create)
			appsGroup.GET("/:id", h.AdminStore.Get)
			appsGroup.PUT("/:id", h.AdminStore.Update)
			appsGroup.DELETE("/:id", h.AdminStore.DeleteStore)
			appsGroup.POST("/:id/credentials", h.AdminStore.GenerateCredentials)
		}

		// Reports
		reportsGroup := adminGroup.Group("/reports")
		reportsGroup.Use(middleware.RequireApproval())
		{
			reportsGroup.GET("/overview", h.AdminReport.Overview)
			reportsGroup.GET("/trend", h.AdminReport.Trend)
			reportsGroup.GET("/export/coupons", h.AdminReport.ExportCoupons)
			reportsGroup.GET("/export/usage", h.AdminReport.ExportUsage)
		}

		// Admin coupon operations
		couponGroup := adminGroup.Group("/coupons")
		couponGroup.Use(middleware.RequireApproval())
		{
			couponGroup.POST("/issue", h.AdminCoupon.Issue)
			couponGroup.GET("/records", h.AdminCoupon.ListRecords)
			couponGroup.GET("/records/:id", h.AdminCoupon.GetRecord)
			couponGroup.POST("/consume", h.AdminCoupon.Consume)
			couponGroup.GET("/consume-records", h.AdminCoupon.ListConsumeRecords)
		}
	}

	// --- Open API (HMAC store auth + rate limiting, unchanged) ---
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
		openGroup.GET("/templates",
			middleware.RateLimit("rl:query", 300, 1*time.Minute),
			h.OpenCoupon.ListTemplates,
		)

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

	// Serve React SPA static files in production
	if cfg.Server.StaticDir != "" {
		staticDir := cfg.Server.StaticDir
		r.NoRoute(func(c *gin.Context) {
			// API paths return 404 as JSON
			if len(c.Request.URL.Path) >= 4 && c.Request.URL.Path[:4] == "/api" {
				c.JSON(http.StatusNotFound, gin.H{"code": 40400, "message": "not found"})
				return
			}
			// Try to serve exact file
			filePath := filepath.Join(staticDir, c.Request.URL.Path)
			if _, err := os.Stat(filePath); err == nil {
				c.File(filePath)
				return
			}
			// SPA fallback
			c.File(filepath.Join(staticDir, "index.html"))
		})
	}
}
