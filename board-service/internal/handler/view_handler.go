package handler

import (
	"board-service/internal/apperrors"
	"board-service/internal/dto"
	"board-service/internal/middleware"
	"board-service/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ViewHandler struct {
	viewService service.ViewService
}

func NewViewHandler(viewService service.ViewService) *ViewHandler {
	return &ViewHandler{viewService: viewService}
}

// ==================== View CRUD ====================

// CreateView godoc
// @Summary Create a saved view
// @Description Create a new saved view with filters, sorting, and grouping
// @Tags Views
// @Accept json
// @Produce json
// @Param request body dto.CreateViewRequest true "View creation request"
// @Success 201 {object} dto.SuccessResponse{data=dto.ViewResponse}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /views [post]
// @Security BearerAuth
func (h *ViewHandler) CreateView(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)

	var req dto.CreateViewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		appErr := apperrors.Wrap(err, apperrors.ErrCodeValidation, "입력값이 유효하지 않습니다", 400)
		dto.Error(c, appErr)
		return
	}

	view, err := h.viewService.CreateView(userID, &req)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			dto.Error(c, appErr)
		} else {
			dto.Error(c, apperrors.New(apperrors.ErrCodeInternalServer, "뷰 생성 실패", 500))
		}
		return
	}

	dto.SuccessWithStatus(c, http.StatusCreated, view)
}

// GetViewsByProject godoc
// @Summary Get views by project
// @Description Get all saved views for a project
// @Tags Views
// @Accept json
// @Produce json
// @Param projectId path string true "Project ID"
// @Success 200 {object} dto.SuccessResponse{data=[]dto.ViewResponse}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /projects/{projectId}/views [get]
// @Security BearerAuth
func (h *ViewHandler) GetViewsByProject(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	projectID := c.Param("projectId")

	views, err := h.viewService.GetViewsByProject(userID, projectID)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			dto.Error(c, appErr)
		} else {
			dto.Error(c, apperrors.New(apperrors.ErrCodeInternalServer, "뷰 조회 실패", 500))
		}
		return
	}

	dto.Success(c, views)
}

// GetView godoc
// @Summary Get view by ID
// @Description Get a saved view by its ID
// @Tags Views
// @Accept json
// @Produce json
// @Param viewId path string true "View ID"
// @Success 200 {object} dto.SuccessResponse{data=dto.ViewResponse}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /views/{viewId} [get]
// @Security BearerAuth
func (h *ViewHandler) GetView(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	viewID := c.Param("viewId")

	view, err := h.viewService.GetView(userID, viewID)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			dto.Error(c, appErr)
		} else {
			dto.Error(c, apperrors.New(apperrors.ErrCodeInternalServer, "뷰 조회 실패", 500))
		}
		return
	}

	dto.Success(c, view)
}

// UpdateView godoc
// @Summary Update view
// @Description Update a saved view
// @Tags Views
// @Accept json
// @Produce json
// @Param viewId path string true "View ID"
// @Param request body dto.UpdateViewRequest true "View update request"
// @Success 200 {object} dto.SuccessResponse{data=dto.ViewResponse}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /views/{viewId} [patch]
// @Security BearerAuth
func (h *ViewHandler) UpdateView(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	viewID := c.Param("viewId")

	var req dto.UpdateViewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		appErr := apperrors.Wrap(err, apperrors.ErrCodeValidation, "입력값이 유효하지 않습니다", 400)
		dto.Error(c, appErr)
		return
	}

	view, err := h.viewService.UpdateView(userID, viewID, &req)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			dto.Error(c, appErr)
		} else {
			dto.Error(c, apperrors.New(apperrors.ErrCodeInternalServer, "뷰 수정 실패", 500))
		}
		return
	}

	dto.Success(c, view)
}

// DeleteView godoc
// @Summary Delete view
// @Description Delete a saved view
// @Tags Views
// @Accept json
// @Produce json
// @Param viewId path string true "View ID"
// @Success 204
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /views/{viewId} [delete]
// @Security BearerAuth
func (h *ViewHandler) DeleteView(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	viewID := c.Param("viewId")

	if err := h.viewService.DeleteView(userID, viewID); err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			dto.Error(c, appErr)
		} else {
			dto.Error(c, apperrors.New(apperrors.ErrCodeInternalServer, "뷰 삭제 실패", 500))
		}
		return
	}

	c.Status(http.StatusNoContent)
}

// ==================== Apply View ====================

// ApplyView godoc
// @Summary Apply view
// @Description Apply a saved view to get filtered/sorted/grouped boards
// @Tags Views
// @Accept json
// @Produce json
// @Param viewId path string true "View ID"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} dto.SuccessResponse{data=object}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /views/{viewId}/boards [get]
// @Security BearerAuth
func (h *ViewHandler) ApplyView(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	viewID := c.Param("viewId")

	// Get pagination params
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	result, err := h.viewService.ApplyView(userID, viewID, page, limit)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			dto.Error(c, appErr)
		} else {
			dto.Error(c, apperrors.New(apperrors.ErrCodeInternalServer, "뷰 적용 실패", 500))
		}
		return
	}

	dto.Success(c, result)
}

// ==================== Board Order ====================

// UpdateBoardOrder godoc
// @Summary Update board order in view
// @Description Update manual board ordering within a view
// @Tags Views
// @Accept json
// @Produce json
// @Param request body dto.UpdateBoardOrderRequest true "Board order update request"
// @Success 204
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /view-board-orders [put]
// @Security BearerAuth
func (h *ViewHandler) UpdateBoardOrder(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)

	var req dto.UpdateBoardOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		appErr := apperrors.Wrap(err, apperrors.ErrCodeValidation, "입력값이 유효하지 않습니다", 400)
		dto.Error(c, appErr)
		return
	}

	if err := h.viewService.UpdateBoardOrder(userID, &req); err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			dto.Error(c, appErr)
		} else {
			dto.Error(c, apperrors.New(apperrors.ErrCodeInternalServer, "보드 순서 업데이트 실패", 500))
		}
		return
	}

	c.Status(http.StatusNoContent)
}
