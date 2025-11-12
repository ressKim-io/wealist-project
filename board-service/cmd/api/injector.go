package main

import (
	"board-service/internal/cache"
	"board-service/internal/client"
	"board-service/internal/config"
	"board-service/internal/handler"
	"board-service/internal/middleware"
	"board-service/internal/repository"
	"board-service/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Application은 모든 핸들러를 포함하는 구조체입니다
// Wire가 설치되면 wire.go의 Application을 사용하게 됩니다
type Application struct {
	HealthHandler  *handler.HealthHandler
	ProjectHandler *handler.ProjectHandler
	BoardHandler   *handler.BoardHandler
	CommentHandler *handler.CommentHandler
	FieldHandler   *handler.FieldHandler
	ViewHandler    *handler.ViewHandler
}

// InitializeApplication은 수동 의존성 주입으로 애플리케이션을 초기화합니다
// Wire가 설치되면 wire_gen.go의 InitializeApplication이 자동 생성됩니다
func InitializeApplication(
	cfg *config.Config,
	log *zap.Logger,
	db *gorm.DB,
	rdb *redis.Client,
) (*Application, error) {
	// ==================== Initialize Caches ====================
	workspaceCache := cache.NewWorkspaceCache(rdb)
	userInfoCache := cache.NewUserInfoCache(rdb)
	fieldCache := cache.NewFieldCache(rdb)

	// ==================== Initialize Client ====================
	userClient := client.NewUserClient(cfg.UserService.URL)
	log.Info("User Service client initialized")

	// ==================== Initialize Repositories ====================
	roleRepo := repository.NewRoleRepository(db)
	projectRepo := repository.NewProjectRepository(db)
	boardRepo := repository.NewBoardRepository(db)
	commentRepo := repository.NewCommentRepository(db)
	fieldRepo := repository.NewFieldRepository(db)

	// ==================== Initialize Services ====================
	boardService := service.NewBoardService(
		boardRepo,
		projectRepo,
		roleRepo,
		fieldRepo,
		commentRepo, // UnitOfWork를 위해 추가
		userClient,
		userInfoCache,
		log,
		db,
	)

	projectService := service.NewProjectService(
		projectRepo,
		roleRepo,
		fieldRepo,
		userClient,
		workspaceCache,
		userInfoCache,
		log,
		db,
	)

	commentService := service.NewCommentService(
		commentRepo,
		boardRepo,
		projectRepo,
		userClient,
		userInfoCache,
		log,
		db,
	)

	fieldService := service.NewFieldService(
		fieldRepo,
		projectRepo,
		fieldCache,
		log,
		db,
	)

	fieldValueService := service.NewFieldValueService(
		fieldRepo,
		boardRepo,
		projectRepo,
		fieldCache,
		log,
		db,
	)

	viewService := service.NewViewService(
		fieldRepo,
		boardRepo,
		projectRepo,
		fieldCache,
		log,
		db,
	)

	// ==================== Initialize Handlers ====================
	healthHandler := handler.NewHealthHandler(db, rdb)
	projectHandler := handler.NewProjectHandler(projectService)
	boardHandler := handler.NewBoardHandler(boardService)
	commentHandler := handler.NewCommentHandler(commentService)
	fieldHandler := handler.NewFieldHandler(fieldService, fieldValueService)
	viewHandler := handler.NewViewHandler(viewService)

	return &Application{
		HealthHandler:  healthHandler,
		ProjectHandler: projectHandler,
		BoardHandler:   boardHandler,
		CommentHandler: commentHandler,
		FieldHandler:   fieldHandler,
		ViewHandler:    viewHandler,
	}, nil
}

// RegisterRoutes는 모든 라우트를 등록합니다
func (app *Application) RegisterRoutes(r *gin.Engine, cfg *config.Config) {
	// Health check (no authentication required)
	handler.RegisterRoutes(r, app.HealthHandler)

	// API routes group (authentication required)
	api := r.Group("/api")
	api.Use(middleware.AuthMiddleware(cfg.JWT.Secret))
	{
		// Project routes
		projects := api.Group("/projects")
		{
			// Project CRUD
			projects.POST("", app.ProjectHandler.CreateProject)
			projects.GET("", app.ProjectHandler.GetProjects)
			projects.GET("/search", app.ProjectHandler.SearchProjects)
			projects.GET("/:projectId", app.ProjectHandler.GetProject)
			projects.PUT("/:projectId", app.ProjectHandler.UpdateProject)
			projects.DELETE("/:projectId", app.ProjectHandler.DeleteProject)

			// Join Requests
			projects.POST("/join-requests", app.ProjectHandler.CreateJoinRequest)
			projects.GET("/:projectId/join-requests", app.ProjectHandler.GetJoinRequests)
			projects.PUT("/join-requests/:joinRequestId", app.ProjectHandler.UpdateJoinRequest)

			// Members
			projects.GET("/:projectId/members", app.ProjectHandler.GetProjectMembers)
			projects.PUT("/:projectId/members/:memberId/role", app.ProjectHandler.UpdateMemberRole)
			projects.DELETE("/:projectId/members/:memberId", app.ProjectHandler.RemoveMember)

			// Project fields
			projects.GET("/:projectId/fields", app.FieldHandler.GetFieldsByProject)
			projects.PUT("/:projectId/fields/order", app.FieldHandler.UpdateFieldOrder)

			// Project views
			projects.GET("/:projectId/views", app.ViewHandler.GetViewsByProject)
		}

		// Board routes
		boards := api.Group("/boards")
		{
			boards.POST("", app.BoardHandler.CreateBoard)
			boards.GET("/:boardId", app.BoardHandler.GetBoard)
			boards.GET("", app.BoardHandler.GetBoards)
			boards.PUT("/:boardId", app.BoardHandler.UpdateBoard)
			boards.DELETE("/:boardId", app.BoardHandler.DeleteBoard)
			boards.PUT("/:boardId/move", app.BoardHandler.MoveBoard)

			// Board field values
			boards.GET("/:boardId/field-values", app.FieldHandler.GetBoardFieldValues)
			api.DELETE("/boards/:boardId/field-values/:fieldId", app.FieldHandler.DeleteFieldValue)
		}

		// Comment routes
		comments := api.Group("/comments")
		{
			comments.POST("", app.CommentHandler.CreateComment)
			comments.GET("", app.CommentHandler.GetCommentsByBoardID)
			comments.PUT("/:commentId", app.CommentHandler.UpdateComment)
			comments.DELETE("/:commentId", app.CommentHandler.DeleteComment)
		}

		// Custom Fields routes
		api.POST("/fields", app.FieldHandler.CreateField)
		api.GET("/fields/:fieldId", app.FieldHandler.GetField)
		api.PATCH("/fields/:fieldId", app.FieldHandler.UpdateField)
		api.DELETE("/fields/:fieldId", app.FieldHandler.DeleteField)

		// Field Options
		api.POST("/field-options", app.FieldHandler.CreateOption)
		api.GET("/fields/:fieldId/options", app.FieldHandler.GetOptionsByField)
		api.PATCH("/field-options/:optionId", app.FieldHandler.UpdateOption)
		api.DELETE("/field-options/:optionId", app.FieldHandler.DeleteOption)
		api.PUT("/fields/:fieldId/options/order", app.FieldHandler.UpdateOptionOrder)

		// Board Field Values
		api.POST("/board-field-values", app.FieldHandler.SetFieldValue)
		api.POST("/board-field-values/multi-select", app.FieldHandler.SetMultiSelectValue)

		// Saved Views
		api.POST("/views", app.ViewHandler.CreateView)
		api.GET("/views/:viewId", app.ViewHandler.GetView)
		api.PATCH("/views/:viewId", app.ViewHandler.UpdateView)
		api.DELETE("/views/:viewId", app.ViewHandler.DeleteView)
		api.GET("/views/:viewId/boards", app.ViewHandler.ApplyView)
		api.PUT("/view-board-orders", app.ViewHandler.UpdateBoardOrder)
	}
}
