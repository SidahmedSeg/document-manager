package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/SidahmedSeg/document-manager/backend/pkg/cache"
	"github.com/SidahmedSeg/document-manager/backend/pkg/config"
	"github.com/SidahmedSeg/document-manager/backend/pkg/database"
	"github.com/SidahmedSeg/document-manager/backend/pkg/logger"
	"github.com/SidahmedSeg/document-manager/backend/pkg/middleware"
	"github.com/SidahmedSeg/document-manager/backend/services/quota-service/internal/handler"
	"github.com/SidahmedSeg/document-manager/backend/services/quota-service/internal/repository"
	"github.com/SidahmedSeg/document-manager/backend/services/quota-service/internal/service"
	"go.uber.org/zap"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}

	// Override port for quota service
	cfg.Server.Port = 10006

	// Initialize logger
	log, err := logger.New(cfg.Environment, cfg.Logger.Level, cfg.Logger.Format)
	if err != nil {
		panic(fmt.Sprintf("failed to initialize logger: %v", err))
	}
	defer log.Sync()
	logger.SetGlobal(log)

	log.Info("starting quota service",
		zap.String("environment", cfg.Environment),
		zap.String("version", cfg.AppVersion),
		zap.Int("port", cfg.Server.Port),
	)

	// Connect to database
	db, err := database.NewPostgresDB(cfg.Database, log.Logger)
	if err != nil {
		log.Fatal("failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// Verify database health
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.HealthCheck(ctx); err != nil {
		log.Fatal("database health check failed", zap.Error(err))
	}
	log.Info("database connection established")

	// Connect to Redis cache
	cacheClient, err := cache.NewRedisCache(cfg.Redis, log.Logger)
	if err != nil {
		log.Fatal("failed to connect to cache", zap.Error(err))
	}
	defer cacheClient.Close()

	// Verify cache health
	if err := cacheClient.HealthCheck(ctx); err != nil {
		log.Fatal("cache health check failed", zap.Error(err))
	}
	log.Info("cache connection established")

	// Initialize layers
	repo := repository.NewRepository(db, log.Logger)
	svc := service.NewService(repo, cacheClient, log.Logger)
	h := handler.NewHandler(svc, log.Logger)

	// Setup HTTP router
	mux := http.NewServeMux()

	// Health check endpoints (no auth required)
	mux.HandleFunc("GET /health", h.HealthCheck)
	mux.HandleFunc("GET /health/ready", h.ReadyCheck)

	// Public endpoints
	mux.HandleFunc("GET /api/quotas/plans", h.GetPredefinedPlans)

	// Quota check endpoint (internal use)
	mux.HandleFunc("POST /api/quotas/check", h.CheckQuota)

	// Quota endpoints (auth required)
	mux.HandleFunc("POST /api/quotas", h.CreateQuota)
	mux.HandleFunc("GET /api/quotas/me", h.GetQuota)
	mux.HandleFunc("PUT /api/quotas/me", h.UpdateQuota)

	// Usage endpoints (auth required)
	mux.HandleFunc("GET /api/quotas/usage", h.GetUsage)
	mux.HandleFunc("GET /api/quotas/overview", h.GetOverview)
	mux.HandleFunc("POST /api/quotas/usage/increment", h.IncrementUsage)
	mux.HandleFunc("POST /api/quotas/usage/decrement", h.DecrementUsage)

	// Stats and logs endpoints (auth required)
	mux.HandleFunc("GET /api/quotas/stats", h.GetUsageStats)
	mux.HandleFunc("GET /api/quotas/logs", h.GetUsageLogs)

	// Apply middleware chain
	var httpHandler http.Handler = mux
	httpHandler = middleware.RequestID()(httpHandler)
	httpHandler = middleware.ExtractAuthHeaders(log)(httpHandler)
	httpHandler = middleware.Logging(log)(httpHandler)
	httpHandler = middleware.Recovery(log)(httpHandler)
	httpHandler = middleware.Timeout(30 * time.Second)(httpHandler)

	// Create HTTP server
	srv := &http.Server{
		Addr:         cfg.Server.GetServerAddr(),
		Handler:      httpHandler,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Start server in goroutine
	go func() {
		log.Info("quota service listening",
			zap.String("addr", srv.Addr),
		)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down quota service...")

	// Graceful shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("server shutdown error", zap.Error(err))
	}

	log.Info("quota service stopped")
}
