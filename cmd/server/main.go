package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"hxcoupon/config"
	"hxcoupon/internal/model"
	"hxcoupon/internal/repository"
	"hxcoupon/internal/router"
	"hxcoupon/internal/service"
	redisutil "hxcoupon/internal/pkg/redis"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	configPath := flag.String("config", "config/config.yaml", "Path to config file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	logger := initLogger(cfg.Log)
	defer logger.Sync()

	// Set Gin mode early
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	db, err := initDB(cfg.Database)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	if err := redisutil.Init(cfg.Redis.Addr(), cfg.Redis.Password, cfg.Redis.DB, cfg.Redis.PoolSize); err != nil {
		logger.Warn("Redis connection failed, running with degraded mode", zap.Error(err))
	} else {
		logger.Info("Redis connected")
	}
	defer redisutil.Close()

	// Auto-migrate in debug mode
	if cfg.Server.Mode == "debug" {
		db.AutoMigrate(
			&model.Store{},
			&model.StoreAPICredential{},
			&model.CouponTemplate{},
			&model.CouponTemplateStore{},
			&model.CouponInstance{},
			&model.CouponUsageRecord{},
			&model.AdminUser{},
			&model.ApprovalRecord{},
		)
		logger.Info("Auto-migration completed")
	}

	// Init repositories
	storeRepo := repository.NewStoreRepo(db)
	credRepo := repository.NewCredentialRepo(db)
	templateRepo := repository.NewTemplateRepo(db)
	templateStoreRepo := repository.NewTemplateStoreRepo(db)
	instanceRepo := repository.NewInstanceRepo(db)
	usageRecordRepo := repository.NewUsageRecordRepo(db)
	adminUserRepo := repository.NewAdminUserRepo(db)
	approvalRecordRepo := repository.NewApprovalRecordRepo(db)

	// Init services
	authSvc := service.NewAuthService(adminUserRepo, cfg.JWT)
	storeSvc := service.NewStoreService(db, storeRepo, credRepo, cfg.OSS)
	templateSvc := service.NewTemplateService(db, templateRepo, templateStoreRepo, storeRepo)
	couponSvc := service.NewCouponService(db, instanceRepo, templateRepo, templateStoreRepo, usageRecordRepo, storeRepo, credRepo)
	statisticsSvc := service.NewStatisticsService(storeRepo, templateRepo, instanceRepo)
	adminUserSvc := service.NewAdminUserService(db, adminUserRepo, storeRepo, credRepo, approvalRecordRepo)
	reportSvc := service.NewReportService(db, instanceRepo, usageRecordRepo, templateRepo, storeRepo)

	svc := &router.Services{
		Auth:       authSvc,
		Store:      storeSvc,
		Template:   templateSvc,
		Coupon:     couponSvc,
		Statistics: statisticsSvc,
		AdminUser:  adminUserSvc,
		Report:     reportSvc,
	}

	handlers := router.NewHandlers(svc)

	r := gin.New()
	router.Setup(r, handlers, svc, cfg, logger)

	// Start expiry background job
	go runExpiryJob(instanceRepo, logger)

	// Start HTTP server
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	logger.Info("Server starting", zap.String("addr", addr))

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server failed", zap.Error(err))
		}
	}()

	<-quit
	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}
	logger.Info("Server stopped")
}

func initDB(cfg config.DatabaseConfig) (*gorm.DB, error) {
	dbLogLevel := logger.Warn
	if gin.Mode() == gin.DebugMode {
		dbLogLevel = logger.Info
	}

	db, err := gorm.Open(mysql.Open(cfg.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(dbLogLevel),
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	return db, nil
}

func initLogger(cfg config.LogConfig) *zap.Logger {
	var level zapcore.Level
	switch cfg.Level {
	case "debug":
		level = zapcore.DebugLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	default:
		level = zapcore.InfoLevel
	}

	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "timestamp"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	var encoder zapcore.Encoder
	if cfg.Format == "json" {
		encoder = zapcore.NewJSONEncoder(encoderCfg)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderCfg)
	}

	var writeSyncer zapcore.WriteSyncer
	if cfg.Output == "stdout" || cfg.Output == "" {
		writeSyncer = zapcore.AddSync(os.Stdout)
	} else {
		file, _ := os.OpenFile(cfg.Output, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		writeSyncer = zapcore.AddSync(file)
	}

	core := zapcore.NewCore(encoder, writeSyncer, level)
	return zap.New(core, zap.AddCaller())
}

func runExpiryJob(instanceRepo *repository.InstanceRepo, logger *zap.Logger) {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		ctx := context.Background()

		// Distributed lock via Redis SETNX
		lockKey := "lock:expiry_job"
		locked := true
		if redisutil.Client != nil {
			ok, err := redisutil.Client.SetNX(ctx, lockKey, "1", 9*time.Minute).Result()
			if err != nil || !ok {
				locked = false
			}
		}

		if !locked {
			continue
		}

		total := int64(0)
		for {
			affected, err := instanceRepo.ExpireBatch(ctx, 5000)
			if err != nil {
				logger.Error("Expiry job error", zap.Error(err))
				break
			}
			total += affected
			if affected < 5000 {
				break
			}
		}
		if total > 0 {
			logger.Info("Expiry job completed", zap.Int64("expired", total))
		}
	}
}
