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
	workspaceCache := cache.NewWorkspaceCache(rdb)
	userInfoCache := cache.NewUserInfoCache(rdb)
	fieldCache := cache.NewFieldCache(rdb) // Custom fields cache

	// 5.6. Initialize repositories
	roleRepo := repository.NewRoleRepository(db)
	projectRepo := repository.NewProjectRepository(db)
	boardRepo := repository.NewBoardRepository(db)
	commentRepo := repository.NewCommentRepository(db)
	fieldRepo := repository.NewFieldRepository(db)

	// 5.7. Initialize services
	boardService := service.NewBoardService(boardRepo, projectRepo, roleRepo, fieldRepo, userClient, userInfoCache, log, db)
	projectService := service.NewProjectService(projectRepo, roleRepo, fieldRepo, userClient, workspaceCache, userInfoCache, log, db)
	commentService := service.NewCommentService(commentRepo, boardRepo, projectRepo, userClient, userInfoCache, log, db)
	// Custom fields services (new ProjectField system)
	fieldService := service.NewFieldService(fieldRepo, projectRepo, fieldCache, log, db)
	fieldValueService := service.NewFieldValueService(fieldRepo, boardRepo, projectRepo, fieldCache, log, db)
	viewService := service.NewViewService(fieldRepo, boardRepo, projectRepo, fieldCache, log, db)

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
		boardHandler := handler.NewBoardHandler(boardService)
		commentHandler := handler.NewCommentHandler(commentService)
		fieldHandler := handler.NewFieldHandler(fieldService, fieldValueService) // Custom fields (new ProjectField system)
		viewHandler := handler.NewViewHandler(viewService) // Saved views (filters/sorting/grouping)

		// Project routes
		projects := api.Group("/projects")
		{
			// Project CRUD
			projects.POST("", projectHandler.CreateProject)
			projects.GET("", projectHandler.GetProjects)           // Get all projects in workspace
			projects.GET("/search", projectHandler.SearchProjects) // Must be before /:projectId
			projects.GET("/:projectId", projectHandler.GetProject)
			projects.PUT("/:projectId", projectHandler.UpdateProject)
			projects.DELETE("/:projectId", projectHandler.DeleteProject)

			// Join Requests
			projects.POST("/join-requests", projectHandler.CreateJoinRequest)
			projects.GET("/:projectId/join-requests", projectHandler.GetJoinRequests)
			projects.PUT("/join-requests/:joinRequestId", projectHandler.UpdateJoinRequest)

			// Members
			projects.GET("/:projectId/members", projectHandler.GetProjectMembers)
			projects.PUT("/:projectId/members/:memberId/role", projectHandler.UpdateMemberRole)
			projects.DELETE("/:projectId/members/:memberId", projectHandler.RemoveMember)
		}

		// Custom Fields routes removed - use new ProjectField system instead
		// See /fields, /field-values, and /views endpoints

		// Board routes
		boards := api.Group("/boards")
		{
			boards.POST("", boardHandler.CreateBoard)
			boards.GET("/:boardId", boardHandler.GetBoard)
			boards.GET("", boardHandler.GetBoards)
			boards.PUT("/:boardId", boardHandler.UpdateBoard)
			boards.DELETE("/:boardId", boardHandler.DeleteBoard)
			boards.PUT("/:boardId/move", boardHandler.MoveBoard) // Integrated API: field value change + order update
		}

		// Comment routes
		comments := api.Group("/comments")
		{
			comments.POST("", commentHandler.CreateComment)
			comments.GET("", commentHandler.GetCommentsByBoardID) // Changed from nested route
			comments.PUT("/:commentId", commentHandler.UpdateComment)
			comments.DELETE("/:commentId", commentHandler.DeleteComment)
		}

		// Custom Fields routes (Jira-style) - NEW SYSTEM
		// Field CRUD
		api.POST("/fields", fieldHandler.CreateField)
		api.GET("/fields/:fieldId", fieldHandler.GetField)
		api.PATCH("/fields/:fieldId", fieldHandler.UpdateField)
		api.DELETE("/fields/:fieldId", fieldHandler.DeleteField)
		projects.GET("/:projectId/fields", fieldHandler.GetFieldsByProject) // Under projects
		projects.PUT("/:projectId/fields/order", fieldHandler.UpdateFieldOrder)

		// Field Options
		api.POST("/field-options", fieldHandler.CreateOption)
		api.GET("/fields/:fieldId/options", fieldHandler.GetOptionsByField)
		api.PATCH("/field-options/:optionId", fieldHandler.UpdateOption)
		api.DELETE("/field-options/:optionId", fieldHandler.DeleteOption)
		api.PUT("/fields/:fieldId/options/order", fieldHandler.UpdateOptionOrder)

		// Board Field Values
		api.POST("/board-field-values", fieldHandler.SetFieldValue)
		api.POST("/board-field-values/multi-select", fieldHandler.SetMultiSelectValue)
		api.GET("/boards/:boardId/field-values", fieldHandler.GetBoardFieldValues)
		api.DELETE("/boards/:boardId/field-values/:fieldId", fieldHandler.DeleteFieldValue)

		// Saved Views (filters/sorting/grouping)
		api.POST("/views", viewHandler.CreateView)
		api.GET("/views/:viewId", viewHandler.GetView)
		api.PATCH("/views/:viewId", viewHandler.UpdateView)
		api.DELETE("/views/:viewId", viewHandler.DeleteView)
		api.GET("/views/:viewId/boards", viewHandler.ApplyView) // Apply view and get boards
		projects.GET("/:projectId/views", viewHandler.GetViewsByProject) // Under projects
		api.PUT("/view-board-orders", viewHandler.UpdateBoardOrder) // Manual board ordering in views
	}

	// 12. Start server
	addr := ":" + cfg.Server.Port
	log.Info("Server starting", zap.String("address", addr))

	if err := r.Run(addr); err != nil {
		log.Fatal("Server failed to start", zap.Error(err))
		os.Exit(1)
	}
}
