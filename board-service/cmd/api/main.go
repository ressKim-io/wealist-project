package main

import (
	"board-service/internal/cache"
	"board-service/internal/config"
	"board-service/internal/database"
	"board-service/internal/middleware"
	"board-service/pkg/logger"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"

	_ "board-service/docs" // Swagger docs
)

// @title Board Service API
// @version 1.0
// @description Board management API for weAlist project management platform
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@wealist.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8000
// @BasePath /api

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	// 1. Load configuration
	cfg, err := config.Load()
	if err != nil {
		panic("Failed to load config: " + err.Error())
	}

	// 2. Initialize logger
	log, err := logger.Init(cfg.Log.Level, cfg.Server.Env)
	if err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}
	defer log.Sync()

	log.Info("Starting board-service",
		zap.String("env", cfg.Server.Env),
		zap.String("port", cfg.Server.Port),
	)

	// 3. Connect to database
	db, err := database.Connect(cfg.Database.URL, log, cfg.Server.UseAutoMigrate)
	if err != nil {
		log.Fatal("Failed to connect to database", zap.Error(err))
	}

	// 4. Connect to Redis
	rdb, err := cache.Connect(cfg.Redis.URL, log)
	if err != nil {
		log.Fatal("Failed to connect to Redis", zap.Error(err))
	}

	// 5. Initialize application with dependency injection
	app, err := InitializeApplication(cfg, log, db, rdb)
	if err != nil {
		log.Fatal("Failed to initialize application", zap.Error(err))
	}

	// 6. Configure Gin mode
	if cfg.Server.Env == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 7. Create router and register middleware
	r := gin.New()
	r.Use(middleware.RequestIDMiddleware())
	r.Use(middleware.LoggerMiddleware(log))
	r.Use(middleware.RecoveryMiddleware(log))
	r.Use(middleware.CORSMiddleware(cfg.CORS.Origins))

	// 8. Register Swagger (development only)
	if cfg.Server.Env == "dev" {
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
		log.Info("Swagger UI enabled", zap.String("url", "http://localhost:"+cfg.Server.Port+"/swagger/index.html"))
	}

	// 9. Register Prometheus metrics endpoint
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// 10. Register all routes through the application
	app.RegisterRoutes(r, cfg)

	// 11. Start server
	addr := ":" + cfg.Server.Port
	log.Info("Server starting", zap.String("address", addr))

	if err := r.Run(addr); err != nil {
		log.Fatal("Server failed to start", zap.Error(err))
		os.Exit(1)
	}
}
