package main

import (
	"fmt"
	"log"
	"time"

	"github.com/cachatto/master-slave-server/internal/config"
	"github.com/cachatto/master-slave-server/internal/handler"
	"github.com/cachatto/master-slave-server/internal/middleware"
	"github.com/cachatto/master-slave-server/internal/models"
	"github.com/cachatto/master-slave-server/internal/repository"
	"github.com/cachatto/master-slave-server/internal/service"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// ─── Load Configuration ──────────────────────────────────────────
	cfg := config.Load()

	// ─── Connect to PostgreSQL ───────────────────────────────────────
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}
	log.Println("✅ Connected to PostgreSQL")

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("❌ Failed to get underlying DB: %v", err)
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// ─── Auto-Migrate Models ─────────────────────────────────────────
	if err := db.AutoMigrate(
		&models.User{},
		&models.App{},
		&models.UserAppPermission{},
		&models.OneTimeCode{},
	); err != nil {
		log.Fatalf("❌ Failed to auto-migrate: %v", err)
	}
	log.Println("✅ Database migration complete")

	// ─── Initialize Repositories ─────────────────────────────────────
	userRepo := repository.NewUserRepository(db)
	appRepo := repository.NewAppRepository(db)
	otcRepo := repository.NewOTCRepository(db)

	// ─── Initialize Services ─────────────────────────────────────────
	authService := service.NewAuthService(userRepo, appRepo, cfg)
	otcService := service.NewOTCService(otcRepo, appRepo, authService, cfg)

	// ─── Start OTC Cleanup Ticker ────────────────────────────────────
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			if err := otcService.CleanExpiredCodes(); err != nil {
				log.Printf("⚠️  OTC cleanup error: %v", err)
			}
		}
	}()

	// ─── Initialize Handlers ─────────────────────────────────────────
	authHandler := handler.NewAuthHandler(authService)
	otcHandler := handler.NewOTCHandler(otcService)

	// ─── Setup Gin Router ────────────────────────────────────────────
	router := gin.Default()

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "master-slave-server",
			"time":    time.Now().UTC().Format(time.RFC3339),
		})
	})

	// Auth routes
	auth := router.Group("/auth")
	{
		// Public endpoints (no auth required)
		auth.POST("/login", authHandler.Login)
		auth.POST("/refresh", authHandler.Refresh)
		auth.POST("/claim-token", otcHandler.ClaimToken)

		// Protected endpoints (JWT required)
		protected := auth.Group("")
		protected.Use(middleware.JWTAuth(authService))
		{
			protected.GET("/verify", authHandler.Verify)
			protected.POST("/exchange-code", otcHandler.ExchangeCode)
		}
	}

	// ─── Start Server ────────────────────────────────────────────────
	addr := fmt.Sprintf(":%s", cfg.ServerPort)
	log.Printf("🚀 Server starting on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("❌ Server failed to start: %v", err)
	}
}
