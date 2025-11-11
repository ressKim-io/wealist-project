//go:build wireinject
// +build wireinject

package main

import (
	"board-service/internal/cache"
	"board-service/internal/client"
	"board-service/internal/config"
	"board-service/internal/handler"
	"board-service/internal/repository"
	"board-service/internal/service"
	"board-service/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ==================== Wire Provider Sets ====================

// repositorySet은 모든 repository providers를 포함합니다
var repositorySet = wire.NewSet(
	repository.NewRoleRepository,
	repository.NewProjectRepository,
	repository.NewBoardRepository,
	repository.NewCommentRepository,
	repository.NewFieldRepository,
)

// cacheSet은 모든 cache providers를 포함합니다
var cacheSet = wire.NewSet(
	cache.NewWorkspaceCache,
	cache.NewUserInfoCache,
	cache.NewFieldCache,
)

// clientSet은 모든 외부 service client providers를 포함합니다
var clientSet = wire.NewSet(
	provideUserClient,
)

// serviceSet은 모든 service providers를 포함합니다
var serviceSet = wire.NewSet(
	service.NewBoardService,
	service.NewProjectService,
	service.NewCommentService,
	service.NewFieldService,
	service.NewFieldValueService,
	service.NewViewService,
)

// handlerSet은 모든 handler providers를 포함합니다
var handlerSet = wire.NewSet(
	handler.NewHealthHandler,
	handler.NewProjectHandler,
	handler.NewBoardHandler,
	handler.NewCommentHandler,
	handler.NewFieldHandler,
	handler.NewViewHandler,
)

// ==================== Provider Functions ====================

// provideUserClient는 UserClient를 생성합니다
func provideUserClient(cfg *config.Config) client.UserClient {
	return client.NewUserClient(cfg.UserService.URL)
}

// ==================== Wire Injectors ====================

// InitializeApplication은 전체 애플리케이션을 초기화합니다
// Wire가 이 함수의 구현을 자동 생성합니다
func InitializeApplication(
	cfg *config.Config,
	log *zap.Logger,
	db *gorm.DB,
	rdb *redis.Client,
) (*Application, error) {
	wire.Build(
		// All provider sets
		repositorySet,
		cacheSet,
		clientSet,
		serviceSet,
		handlerSet,

		// Application constructor
		NewApplication,
	)
	return nil, nil
}

// ==================== Application Struct ====================

// Application은 모든 핸들러를 포함하는 구조체입니다
type Application struct {
	HealthHandler  *handler.HealthHandler
	ProjectHandler *handler.ProjectHandler
	BoardHandler   *handler.BoardHandler
	CommentHandler *handler.CommentHandler
	FieldHandler   *handler.FieldHandler
	ViewHandler    *handler.ViewHandler
}

// NewApplication은 Application을 생성합니다
// Wire가 필요한 모든 의존성을 자동으로 주입합니다
func NewApplication(
	healthHandler *handler.HealthHandler,
	projectHandler *handler.ProjectHandler,
	boardHandler *handler.BoardHandler,
	commentHandler *handler.CommentHandler,
	fieldHandler *handler.FieldHandler,
	viewHandler *handler.ViewHandler,
) *Application {
	return &Application{
		HealthHandler:  healthHandler,
		ProjectHandler: projectHandler,
		BoardHandler:   boardHandler,
		CommentHandler: commentHandler,
		FieldHandler:   fieldHandler,
		ViewHandler:    viewHandler,
	}
}

// RegisterRoutes는 모든 라우트를 등록합니다
func (app *Application) RegisterRoutes(r *gin.Engine, cfg *config.Config) {
	// Health check (no authentication required)
	handler.RegisterRoutes(r, app.HealthHandler)

	// API routes group (authentication required)
	api := r.Group("/api")
	// Uncomment when needed: api.Use(middleware.AuthMiddleware(cfg.JWT.Secret))
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
