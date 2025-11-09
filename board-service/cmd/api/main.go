package main

import (
	"board-service/internal/cache"
	"board-service/internal/client"
	"board-service/internal/config"
	"board-service/internal/database"
	"board-service/internal/handler"
	"board-service/internal/middleware"
	"board-service/internal/repository"
	"board-service/internal/service"
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

	// 5. Initialize User Service client
	userClient := client.NewUserClient(cfg.UserService.URL)
	log.Info("User Service client initialized", zap.String("url", cfg.UserService.URL))

	// 5.5. Initialize caches
	userOrderCache := cache.NewUserOrderCache(rdb)
	workspaceCache := cache.NewWorkspaceCache(rdb)
	userInfoCache := cache.NewUserInfoCache(rdb)

	// 5.6. Initialize repositories
	roleRepo := repository.NewRoleRepository(db)
	projectRepo := repository.NewProjectRepository(db)
	customFieldRepo := repository.NewCustomFieldRepository(db)
	boardRepo := repository.NewBoardRepository(db)
	userOrderRepo := repository.NewUserOrderRepository(db)
	commentRepo := repository.NewCommentRepository(db) // Add CommentRepository

	// 5.7. Initialize services
	// Note: customFieldService needs boardRepo (for Phase 4 TODO), then injected into projectService
	customFieldService := service.NewCustomFieldService(customFieldRepo, projectRepo, roleRepo, boardRepo, log, db)
	boardService := service.NewBoardService(boardRepo, projectRepo, customFieldRepo, roleRepo, userClient, userInfoCache, log, db)
	projectService := service.NewProjectService(projectRepo, roleRepo, userOrderRepo, customFieldService, userClient, workspaceCache, userInfoCache, log, db)
	userOrderService := service.NewUserOrderService(userOrderRepo, projectRepo, customFieldRepo, boardRepo, userOrderCache, log)
	commentService := service.NewCommentService(commentRepo, boardRepo, projectRepo, userClient, userInfoCache, log, db) // Add CommentService

	// 6. Configure Gin mode
	if cfg.Server.Env == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 7. Create router
	r := gin.New()

	

	// 8. Register middleware (order is important)
	r.Use(middleware.RequestIDMiddleware())
	r.Use(middleware.LoggerMiddleware(log))
	r.Use(middleware.RecoveryMiddleware(log))
	r.Use(middleware.CORSMiddleware(cfg.CORS.Origins))

	// 9. Register Swagger (development only)
	if cfg.Server.Env == "dev" {
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
		log.Info("Swagger UI enabled", zap.String("url", "http://localhost:"+cfg.Server.Port+"/swagger/index.html"))
	}

	// 10. Register Prometheus metrics endpoint
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// 11. Register health check (no authentication required)
	healthHandler := handler.NewHealthHandler(db, rdb)
	handler.RegisterRoutes(r, healthHandler)

	// 12. API routes group (authentication required)
	api := r.Group("/api")
	api.Use(middleware.AuthMiddleware(cfg.JWT.Secret))
	{
		// Initialize handlers
		projectHandler := handler.NewProjectHandler(projectService)
		customFieldHandler := handler.NewCustomFieldHandler(customFieldService)
		boardHandler := handler.NewBoardHandler(boardService)
		userOrderHandler := handler.NewUserOrderHandler(userOrderService)
		commentHandler := handler.NewCommentHandler(commentService) // Add CommentHandler

		// Project routes
		projects := api.Group("/projects")
		{
			// Project CRUD
			projects.POST("", projectHandler.CreateProject)
			projects.GET("", projectHandler.GetProjects)           // Get all projects in workspace
			projects.GET("/search", projectHandler.SearchProjects) // Must be before /:project_id
			projects.GET("/:project_id", projectHandler.GetProject)
			projects.PUT("/:project_id", projectHandler.UpdateProject)
			projects.DELETE("/:project_id", projectHandler.DeleteProject)

			// Join Requests
			projects.POST("/join-requests", projectHandler.CreateJoinRequest)
			projects.GET("/:project_id/join-requests", projectHandler.GetJoinRequests)
			projects.PUT("/join-requests/:join_request_id", projectHandler.UpdateJoinRequest)

			// Members
			projects.GET("/:project_id/members", projectHandler.GetProjectMembers)
			projects.PUT("/:project_id/members/:member_id/role", projectHandler.UpdateMemberRole)
			projects.DELETE("/:project_id/members/:member_id", projectHandler.RemoveMember)

			// User Order Management (Drag-and-Drop)
			projects.GET("/:project_id/orders/role-board", userOrderHandler.GetRoleBasedBoardView)
			projects.GET("/:project_id/orders/stage-board", userOrderHandler.GetStageBasedBoardView)
			projects.PUT("/:project_id/orders/role-columns", userOrderHandler.UpdateRoleColumnOrder)
			projects.PUT("/:project_id/orders/stage-columns", userOrderHandler.UpdateStageColumnOrder)
			projects.PUT("/:project_id/orders/role-boards/:role_id", userOrderHandler.UpdateBoardOrderInRole)
			projects.PUT("/:project_id/orders/stage-boards/:stage_id", userOrderHandler.UpdateBoardOrderInStage)
		}

		// Custom Fields routes
		customFields := api.Group("/custom-fields")
		{
			// Custom Roles
			customFields.POST("/roles", customFieldHandler.CreateCustomRole)
			customFields.GET("/projects/:project_id/roles", customFieldHandler.GetCustomRoles)
			customFields.GET("/roles/:role_id", customFieldHandler.GetCustomRole)
			customFields.PUT("/roles/:role_id", customFieldHandler.UpdateCustomRole)
			customFields.DELETE("/roles/:role_id", customFieldHandler.DeleteCustomRole)
			customFields.PUT("/projects/:project_id/roles/order", customFieldHandler.UpdateCustomRoleOrder)

			// Custom Stages
			customFields.POST("/stages", customFieldHandler.CreateCustomStage)
			customFields.GET("/projects/:project_id/stages", customFieldHandler.GetCustomStages)
			customFields.GET("/stages/:stage_id", customFieldHandler.GetCustomStage)
			customFields.PUT("/stages/:stage_id", customFieldHandler.UpdateCustomStage)
			customFields.DELETE("/stages/:stage_id", customFieldHandler.DeleteCustomStage)
			customFields.PUT("/projects/:project_id/stages/order", customFieldHandler.UpdateCustomStageOrder)

			// Custom Importance
			customFields.POST("/importance", customFieldHandler.CreateCustomImportance)
			customFields.GET("/projects/:project_id/importance", customFieldHandler.GetCustomImportances)
			customFields.GET("/importance/:importance_id", customFieldHandler.GetCustomImportance)
			customFields.PUT("/importance/:importance_id", customFieldHandler.UpdateCustomImportance)
			customFields.DELETE("/importance/:importance_id", customFieldHandler.DeleteCustomImportance)
			customFields.PUT("/projects/:project_id/importance/order", customFieldHandler.UpdateCustomImportanceOrder)
		}

		// Board routes
		boards := api.Group("/boards")
		{
			boards.POST("", boardHandler.CreateBoard)
			boards.GET("/:board_id", boardHandler.GetBoard)
			boards.GET("", boardHandler.GetBoards)
			boards.PUT("/:board_id", boardHandler.UpdateBoard)
			boards.DELETE("/:board_id", boardHandler.DeleteBoard)
		}

		// Comment routes
		comments := api.Group("/comments")
		{
			comments.POST("", commentHandler.CreateComment)
			comments.GET("", commentHandler.GetCommentsByBoardID) // Changed from nested route
			comments.PUT("/:id", commentHandler.UpdateComment)
			comments.DELETE("/:id", commentHandler.DeleteComment)
		}
	}

	// 12. Start server
	addr := ":" + cfg.Server.Port
	log.Info("Server starting", zap.String("address", addr))

	if err := r.Run(addr); err != nil {
		log.Fatal("Server failed to start", zap.Error(err))
		os.Exit(1)
	}
}
