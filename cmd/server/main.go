package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/mobarok/devflow/internal/auth"
	"github.com/mobarok/devflow/internal/middleware"
	"github.com/mobarok/devflow/internal/organization"
	"github.com/mobarok/devflow/pkg/config"
	"github.com/mobarok/devflow/pkg/database"
	"github.com/mobarok/devflow/pkg/logger"
	"github.com/mobarok/devflow/pkg/response"
)

const version = "0.1.0"

func main() {
	cfg := config.Load()

	logger.Setup(cfg.Environment, cfg.LogLevel)

	slog.Info("starting DevFlow server",
		"version", version,
		"environment", cfg.Environment,
		"port", cfg.ServerPort,
	)

	// Connect to PostgreSQL
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Initialize Auth module dependencies
	authRepo := auth.NewRepository(db)
	authService := auth.NewService(authRepo, cfg)
	authHandler := auth.NewHandler(authService)

	// Initialize Organization module dependencies
	orgRepo := organization.NewRepository(db)
	orgService := organization.NewService(orgRepo)
	orgHandler := organization.NewHandler(orgService)

	// Create Gin router
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())

	registerRoutes(router, cfg, authHandler, orgHandler, orgService)

	// Start server with graceful shutdown
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.ServerPort),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("server listening", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("server forced to shutdown", "error", err)
		os.Exit(1)
	}

	slog.Info("server exited cleanly")
}

func registerRoutes(
	router *gin.Engine,
	cfg *config.Config,
	authHandler *auth.Handler,
	orgHandler *organization.Handler,
	orgChecker middleware.MemberChecker,
) {
	v1 := router.Group("/api/v1")

	v1.GET("/health", func(c *gin.Context) {
		response.Success(c, http.StatusOK, gin.H{
			"service":     "devflow",
			"version":     version,
			"environment": cfg.Environment,
		})
	})

	authMiddleware := middleware.AuthRequired(cfg.JWTSecret)
	auth.RegisterRoutes(v1, authHandler, authMiddleware)
	organization.RegisterRoutes(v1, orgHandler, authMiddleware, orgChecker)
}
