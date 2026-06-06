// Package main is the entry point for the jiufang application.
// This file initializes all components and starts the HTTP server.
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"jiufang/internal/agent"
	v1 "jiufang/internal/api/v1"
	"jiufang/internal/infrastructure/cache"
	"jiufang/internal/infrastructure/database"
	"jiufang/internal/infrastructure/erp"
	"jiufang/internal/infrastructure/llm"
	"jiufang/internal/middleware"
	"jiufang/internal/migration"
	"jiufang/internal/pkg/config"
	"jiufang/internal/pkg/id"
	"jiufang/internal/pkg/jwt"
	"jiufang/internal/pkg/logger"
	"jiufang/internal/repository"
	"jiufang/internal/service"
)

func main() {
	// Step 1: Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Step 2: Initialize logger
	log, err := logger.InitializeDefault()
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Sync()

	log.Info("Starting jiufang application...",
		zap.Int("port", cfg.Server.Port),
		zap.String("mode", cfg.Server.Mode),
	)

	// Step 3: Connect to PostgreSQL application database
	dbConfig := &database.DatabaseConfig{
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		DBName:   cfg.Database.DBName,
		SSLMode:  cfg.Database.SSLMode,
	}
	db, err := database.Connect(dbConfig)
	if err != nil {
		log.Fatal("Failed to connect to PostgreSQL database", zap.Error(err))
	}
	log.Info("PostgreSQL database connected successfully")

	// Step 3.1: Run database migrations
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.DBName,
	)
	if err := migration.RunMigrations(dsn); err != nil {
		log.Fatal("Failed to run database migrations", zap.Error(err))
	}
	log.Info("Database migrations executed successfully")

	// Step 4: Connect to Redis cache
	redisConfig := &cache.RedisConfig{
		Address:      cfg.Redis.Address,
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		PoolSize:     cfg.Redis.PoolSize,
		MinIdleConns: cfg.Redis.MinIdleConns,
		MaxRetries:   cfg.Redis.MaxRetries,
		DialTimeout:  cfg.Redis.DialTimeout,
		ReadTimeout:  cfg.Redis.ReadTimeout,
		WriteTimeout: cfg.Redis.WriteTimeout,
	}
	redisClient, err := cache.NewRedisClient(redisConfig, log)
	if err != nil {
		log.Fatal("Failed to connect to Redis", zap.Error(err))
	}
	log.Info("Redis connected successfully")

	// Step 5: Initialize LLM client (DeepSeek)
	llmConfig := &llm.LLMConfig{
		Provider: cfg.LLM.Provider,
		Model:    cfg.LLM.Model,
		APIKey:   cfg.LLM.APIKey,
		Endpoint: cfg.LLM.Endpoint,
	}
	llmFactory := llm.NewFactory()
	llmClient, err := llmFactory.GetClient(llmConfig)
	if err != nil {
		log.Fatal("Failed to initialize LLM client", zap.Error(err))
	}
	log.Info("LLM client initialized successfully",
		zap.String("provider", cfg.LLM.Provider),
		zap.String("model", cfg.LLM.Model),
	)

	// Step 6: Connect to ERP database (MySQL, read-only)
	erpConfig := &erp.ERPConfig{
		Driver:          cfg.ERP.Driver,
		Host:            cfg.ERP.Host,
		Port:            cfg.ERP.Port,
		Database:        cfg.ERP.Database,
		Username:        cfg.ERP.Username,
		Password:        cfg.ERP.Password,
		MaxOpenConns:    cfg.ERP.MaxOpenConns,
		MaxIdleConns:    cfg.ERP.MaxIdleConns,
		ConnMaxLifetime: cfg.ERP.ConnMaxLifetime,
		QueryTimeout:    cfg.ERP.QueryTimeout,
		MaxResultRows:   cfg.ERP.MaxResultRows,
	}
	erpReader, err := erp.NewReader(erpConfig)
	if err != nil {
		log.Fatal("Failed to connect to ERP database", zap.Error(err))
	}
	log.Info("ERP database connected successfully")

	// Step 7: Initialize Snowflake ID generator
	idGenerator, err := id.NewSnowflakeGenerator(1) // Node ID = 1
	if err != nil {
		log.Fatal("Failed to initialize Snowflake ID generator", zap.Error(err))
	}
	// Also initialize global snowflake for id.Generate() calls
	if err := id.Init(1); err != nil {
		log.Fatal("Failed to initialize global Snowflake ID generator", zap.Error(err))
	}
	log.Info("Snowflake ID generator initialized successfully")

	// Step 8: Initialize repositories
	userRepo := repository.NewUserRepository(db)
	userGroupRepo := repository.NewUserGroupRepository(db)
	permissionRepo := repository.NewPermissionRepository(db)
	dialogRepo := repository.NewDialogRepository(db)
	queryRepo := repository.NewQueryRepository(db)
	dialogFavoriteRepo := repository.NewDialogFavoriteRepository(db)
	exportRecordRepo := repository.NewExportRecordRepository(db)
	reportRepo := repository.NewReportRepository(db)
	feedbackRepo := repository.NewFeedbackRepository(db)
	alertRepo := repository.NewAlertRepository(db, log)
	operationLogRepo := repository.NewOperationLogRepository(db, log)

	log.Info("Repositories initialized successfully")

	// Step 9: Initialize JWT manager
	jwtManager := jwt.NewJWTManager(&cfg.JWT)

	// Step 10: Initialize SQL validator
	sqlValidator := agent.NewSQLValidator()

	// Step 11: Initialize services
	authService := service.NewAuthAppService(userRepo, userGroupRepo, jwtManager)
	queryService := service.NewQueryAppServiceWithHistory(llmClient, erpReader, queryRepo, idGenerator, log)
	userService := service.NewUserAppService(userRepo, userGroupRepo, idGenerator, log)
	groupService := service.NewGroupAppService(userGroupRepo, userRepo, permissionRepo)
	permissionService := service.NewPermissionAppService(userGroupRepo, permissionRepo)
	historyService := service.NewHistoryAppService(queryRepo, idGenerator)
	dialogFavoriteService := service.NewDialogFavoriteAppService(dialogFavoriteRepo, dialogRepo, idGenerator)
	exportAppService := service.NewExportAppService(exportRecordRepo, queryRepo, idGenerator, log, "./exports", 10000)
	reportService := service.NewReportService(reportRepo, idGenerator, log)
	feedbackService := service.NewFeedbackService(feedbackRepo, queryRepo, idGenerator, log)
	alertService := service.NewAlertService(alertRepo, idGenerator, sqlValidator, log)
	operationLogService := service.NewOperationLogService(operationLogRepo, userRepo, log)
	profileAppService := service.NewProfileAppService(userRepo, userGroupRepo)

	log.Info("Services initialized successfully")

	// Step 12: Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(jwtManager)
	operationLogger := middleware.NewOperationLogger(operationLogService)

	// Step 12: Create Gin engine
	gin.SetMode(cfg.Server.Mode)
	router := gin.New()

	// Step 13: Register global middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.CORS(&middleware.CORSConfig{
		AllowOrigins: cfg.CORS.AllowOrigins,
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders: []string{
			"Origin", "Content-Type", "Content-Length", "Accept-Encoding",
			"X-CSRF-Token", "Authorization", "accept", "origin", "Cache-Control", "X-Requested-With",
		},
		ExposeHeaders:     []string{"Content-Length", "Content-Type"},
		AllowCredentials:  true,
		MaxAge:            12 * 3600,
	}))
	router.Use(operationLogger.Log())

	// Step 14: Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "jiufang application is running",
			"time":    time.Now().Format(time.RFC3339),
		})
	})

	// Step 15: Register API routes
	apiV1 := router.Group("/api/v1")
	v1.RegisterRoutes(
		apiV1,
		authMiddleware,
		profileAppService,
		groupService,
		permissionService,
		authService,
		historyService,
		dialogFavoriteService,
		exportAppService,
		reportService,
		feedbackService,
		alertService,
		operationLogService,
		queryService,
		userService,
		erpReader,
		log,
	)

	log.Info("API routes registered successfully")

	// Step 16: Start HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 120 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Info("HTTP server starting...", zap.String("address", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start HTTP server", zap.Error(err))
		}
	}()

	log.Info("jiufang application started successfully",
		zap.String("address", srv.Addr),
		zap.String("health_check", fmt.Sprintf("http://localhost:%d/health", cfg.Server.Port)),
	)

	// Step 17: Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit

	log.Info("Shutting down server...", zap.String("signal", sig.String()))

	// Step 18: Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("Server forced to shutdown", zap.Error(err))
	}

	// Step 19: Close database connections
	if err := database.Close(db); err != nil {
		log.Error("Failed to close PostgreSQL database", zap.Error(err))
	}

	if err := redisClient.Close(); err != nil {
		log.Error("Failed to close Redis connection", zap.Error(err))
	}

	// Note: ERP reader doesn't have a Close method in the current implementation
	// If needed, add a Close method to erp.Reader

	log.Info("Server exited properly")
}
