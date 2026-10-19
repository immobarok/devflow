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

	"github.com/mobarok/devflow/internal/activity"
	"github.com/mobarok/devflow/internal/auth"
	"github.com/mobarok/devflow/internal/comments"
	"github.com/mobarok/devflow/internal/middleware"
	"github.com/mobarok/devflow/internal/notifications"
	"github.com/mobarok/devflow/internal/organization"
	"github.com/mobarok/devflow/internal/projects"
	"github.com/mobarok/devflow/internal/tasks"
	"github.com/mobarok/devflow/pkg/cache"
	"github.com/mobarok/devflow/pkg/config"
	"github.com/mobarok/devflow/pkg/database"
	"github.com/mobarok/devflow/pkg/email"
	"github.com/mobarok/devflow/pkg/logger"
	"github.com/mobarok/devflow/pkg/oauth"
	"github.com/mobarok/devflow/pkg/pubsub"
	appRedis "github.com/mobarok/devflow/pkg/redis"
	"github.com/mobarok/devflow/pkg/response"
	appWS "github.com/mobarok/devflow/pkg/websocket"
	"github.com/mobarok/devflow/pkg/worker"
	goredis "github.com/redis/go-redis/v9"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/mobarok/devflow/docs"
)

const version = "0.1.0"

// @title DevFlow API
// @version 1.0
// @description Production-grade project management SaaS REST API in Go with multi-tenancy, RBAC, background workers, and real-time WebSocket notifications.
// @contact.name DevFlow Engineering
// @contact.email support@devflow.io
// @license.name MIT
// @license.url https://opensource.org/licenses/MIT
// @host localhost:8080
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and your JWT token. Example: "Bearer eyJhbGciOi..."
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

	// Initialize Redis client, Cache, Token Blacklist, and Pub/Sub Bus
	var redisClient *goredis.Client
	var appCache cache.Cache
	var tokenBlacklist auth.TokenBlacklist
	var pubSubBus pubsub.Bus

	if cfg.RedisURL != "" {
		rc, err := appRedis.Connect(cfg.RedisURL)
		if err != nil {
			slog.Warn("failed to connect to Redis; proceeding without cache/rate-limiting/pubsub", "error", err)
		} else {
			redisClient = rc
			defer redisClient.Close()
			appCache = cache.New(redisClient)
			tokenBlacklist = auth.NewTokenBlacklist(redisClient)
			pubSubBus = pubsub.NewBus(redisClient)
			slog.Info("Redis initialized: cache, rate limiting, token blacklist, and Pub/Sub active")
		}
	}

	oauthMgr := oauth.NewManager(cfg)
	authRepo := auth.NewRepository(db)
	authService := auth.NewService(authRepo, cfg, oauthMgr)
	authHandler := auth.NewHandler(authService, cfg)

	orgRepo := organization.NewRepository(db)
	orgService := organization.NewService(orgRepo)
	orgHandler := organization.NewHandler(orgService)

	projectRepo := projects.NewRepository(db)
	projectService := projects.NewService(projectRepo, appCache)
	projectHandler := projects.NewHandler(projectService)

	// Initialize Activity module dependencies
	activityRepo := activity.NewRepository(db)
	activityService := activity.NewService(activityRepo)
	activityHandler := activity.NewHandler(activityService)

	// Initialize Tasks module dependencies
	taskRepo := tasks.NewRepository(db)
	taskService := tasks.NewService(taskRepo, activityService)
	taskHandler := tasks.NewHandler(taskService)

	// Initialize Comments module dependencies
	commentRepo := comments.NewRepository(db)
	commentService := comments.NewService(commentRepo, activityService)
	commentHandler := comments.NewHandler(commentService)

	// Initialize and start background worker pool
	workerPool := worker.NewPool(5, 100)
	workerPool.Start()
	defer workerPool.Stop()

	// Initialize Email service
	emailService := email.NewService(workerPool)
	_ = emailService

	// Initialize WebSocket Hub and start event loop
	wsHub := appWS.NewHub(pubSubBus)
	hubCtx, hubCancel := context.WithCancel(context.Background())
	defer hubCancel()
	go wsHub.Run(hubCtx)

	wsHandler := appWS.NewHandler(wsHub, cfg.JWTSecret)

	// Initialize Notifications module dependencies (with pubSubBus for real-time delivery)
	notifRepo := notifications.NewRepository(db)
	notifService := notifications.NewService(notifRepo, workerPool, pubSubBus)
	notifHandler := notifications.NewHandler(notifService)

	// Create Gin router
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())

	// Attach Redis-backed global rate limiter (100 req/min)
	if redisClient != nil {
		router.Use(middleware.RateLimiter(redisClient, 100, time.Minute))
	}

	registerRoutes(router, cfg, authHandler, orgHandler, projectHandler, taskHandler, commentHandler, activityHandler, notifHandler, wsHandler, tokenBlacklist, orgService)

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
	projectHandler *projects.Handler,
	taskHandler *tasks.Handler,
	commentHandler *comments.Handler,
	activityHandler *activity.Handler,
	notifHandler *notifications.Handler,
	wsHandler *appWS.Handler,
	tokenBlacklist auth.TokenBlacklist,
	orgChecker middleware.MemberChecker,
) {
	// Swagger interactive documentation UI
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	v1 := router.Group("/api/v1")

	v1.GET("/health", func(c *gin.Context) {
		response.Success(c, http.StatusOK, gin.H{
			"service":     "devflow",
			"version":     version,
			"environment": cfg.Environment,
		})
	})

	// WebSocket endpoint for real-time notifications and updates
	v1.GET("/ws", wsHandler.HandleWebSocket)

	authMiddleware := middleware.AuthRequired(cfg.JWTSecret, tokenBlacklist)
	auth.RegisterRoutes(v1, authHandler, authMiddleware)
	organization.RegisterRoutes(v1, orgHandler, authMiddleware, orgChecker)
	projects.RegisterRoutes(v1, projectHandler, authMiddleware, orgChecker)
	tasks.RegisterRoutes(v1, taskHandler, authMiddleware)
	comments.RegisterRoutes(v1, commentHandler, authMiddleware)
	activity.RegisterRoutes(v1, activityHandler, authMiddleware)
	notifications.RegisterRoutes(v1, notifHandler, authMiddleware)
}
