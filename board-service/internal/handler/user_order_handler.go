package handler

import (
	"board-service/internal/apperrors"
	"board-service/internal/dto"
	"board-service/internal/service"

	"github.com/gin-gonic/gin"
)

type UserOrderHandler struct {
	service service.UserOrderService
}

func NewUserOrderHandler(service service.UserOrderService) *UserOrderHandler {
	return &UserOrderHandler{service: service}
}

// GetRoleBasedBoardView godoc
// @Summary      Get role-based board view
// @Description  Get the board organized by roles with user-specific column and board orders
// @Tags         user-orders
// @Accept       json
// @Produce      json
// @Param        project_id path string true "Project ID"
// @Success      200 {object} dto.SuccessResponse{data=dto.RoleBasedBoardView}
// @Failure      400 {object} dto.ErrorResponse
// @Failure      403 {object} dto.ErrorResponse
// @Router       /api/projects/{project_id}/orders/role-board [get]
// @Security     BearerAuth
func (h *UserOrderHandler) GetRoleBasedBoardView(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		dto.Error(c, apperrors.ErrUnauthorized)
		return
	}

	projectID := c.Param("project_id")
	if projectID == "" {
		dto.Error(c, apperrors.Wrap(nil, apperrors.ErrCodeBadRequest, "프로젝트 ID가 필요합니다", 400))
		return
	}

	view, err := h.service.GetRoleBasedBoardView(c.Request.Context(), userID, projectID)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			dto.Error(c, appErr)
		} else {
			dto.Error(c, apperrors.ErrInternalServer)
		}
		return
	}

	dto.Success(c, view)
}

// GetStageBasedBoardView godoc
// @Summary      Get stage-based board view
// @Description  Get the board organized by stages with user-specific column and board orders
// @Tags         user-orders
// @Accept       json
// @Produce      json
// @Param        project_id path string true "Project ID"
// @Success      200 {object} dto.SuccessResponse{data=dto.StageBasedBoardView}
// @Failure      400 {object} dto.ErrorResponse
// @Failure      403 {object} dto.ErrorResponse
// @Router       /api/projects/{project_id}/orders/stage-board [get]
// @Security     BearerAuth
func (h *UserOrderHandler) GetStageBasedBoardView(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		dto.Error(c, apperrors.ErrUnauthorized)
		return
	}

	projectID := c.Param("project_id")
	if projectID == "" {
		dto.Error(c, apperrors.Wrap(nil, apperrors.ErrCodeBadRequest, "프로젝트 ID가 필요합니다", 400))
		return
	}

	view, err := h.service.GetStageBasedBoardView(c.Request.Context(), userID, projectID)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			dto.Error(c, appErr)
		} else {
			dto.Error(c, apperrors.ErrInternalServer)
		}
		return
	}

	dto.Success(c, view)
}

// UpdateRoleColumnOrder godoc
// @Summary      Update role column order
// @Description  Update the display order of role columns (drag-and-drop)
// @Tags         user-orders
// @Accept       json
// @Produce      json
// @Param        project_id path string true "Project ID"
// @Param        request body dto.UpdateOrderRequest true "Order update details"
// @Success      200 {object} dto.SuccessResponse{data=string}
// @Failure      400 {object} dto.ErrorResponse
// @Failure      403 {object} dto.ErrorResponse
// @Router       /api/projects/{project_id}/orders/role-columns [put]
// @Security     BearerAuth
func (h *UserOrderHandler) UpdateRoleColumnOrder(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		dto.Error(c, apperrors.ErrUnauthorized)
		return
	}

	projectID := c.Param("project_id")
	if projectID == "" {
		dto.Error(c, apperrors.Wrap(nil, apperrors.ErrCodeBadRequest, "프로젝트 ID가 필요합니다", 400))
		return
	}

	var req dto.UpdateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.Error(c, apperrors.Wrap(err, apperrors.ErrCodeValidation, "입력값 검증 실패", 400))
		return
	}

	err := h.service.UpdateRoleColumnOrder(c.Request.Context(), userID, projectID, &req)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			dto.Error(c, appErr)
		} else {
			dto.Error(c, apperrors.ErrInternalServer)
		}
		return
	}

	dto.Success(c, "역할 컬럼 순서가 업데이트되었습니다")
}

// UpdateStageColumnOrder godoc
// @Summary      Update stage column order
// @Description  Update the display order of stage columns (drag-and-drop)
// @Tags         user-orders
// @Accept       json
// @Produce      json
// @Param        project_id path string true "Project ID"
// @Param        request body dto.UpdateOrderRequest true "Order update details"
// @Success      200 {object} dto.SuccessResponse{data=string}
// @Failure      400 {object} dto.ErrorResponse
// @Failure      403 {object} dto.ErrorResponse
// @Router       /api/projects/{project_id}/orders/stage-columns [put]
// @Security     BearerAuth
func (h *UserOrderHandler) UpdateStageColumnOrder(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		dto.Error(c, apperrors.ErrUnauthorized)
		return
	}

	projectID := c.Param("project_id")
	if projectID == "" {
		dto.Error(c, apperrors.Wrap(nil, apperrors.ErrCodeBadRequest, "프로젝트 ID가 필요합니다", 400))
		return
	}

	var req dto.UpdateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.Error(c, apperrors.Wrap(err, apperrors.ErrCodeValidation, "입력값 검증 실패", 400))
		return
	}

	err := h.service.UpdateStageColumnOrder(c.Request.Context(), userID, projectID, &req)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			dto.Error(c, appErr)
		} else {
			dto.Error(c, apperrors.ErrInternalServer)
		}
		return
	}

	dto.Success(c, "진행단계 컬럼 순서가 업데이트되었습니다")
}

// UpdateBoardOrderInRole godoc
// @Summary      Update board order in role column
// @Description  Update the display order of boards within a specific role column (drag-and-drop)
// @Tags         user-orders
// @Accept       json
// @Produce      json
// @Param        project_id path string true "Project ID"
// @Param        role_id path string true "Role ID"
// @Param        request body dto.UpdateOrderRequest true "Order update details"
// @Success      200 {object} dto.SuccessResponse{data=string}
// @Failure      400 {object} dto.ErrorResponse
// @Failure      403 {object} dto.ErrorResponse
// @Router       /api/projects/{project_id}/orders/role-boards/{role_id} [put]
// @Security     BearerAuth
func (h *UserOrderHandler) UpdateBoardOrderInRole(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		dto.Error(c, apperrors.ErrUnauthorized)
		return
	}

	projectID := c.Param("project_id")
	roleID := c.Param("role_id")

	if projectID == "" || roleID == "" {
		dto.Error(c, apperrors.Wrap(nil, apperrors.ErrCodeBadRequest, "프로젝트 ID와 역할 ID가 필요합니다", 400))
		return
	}

	var req dto.UpdateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.Error(c, apperrors.Wrap(err, apperrors.ErrCodeValidation, "입력값 검증 실패", 400))
		return
	}

	err := h.service.UpdateBoardOrderInRole(c.Request.Context(), userID, projectID, roleID, &req)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			dto.Error(c, appErr)
		} else {
			dto.Error(c, apperrors.ErrInternalServer)
		}
		return
	}

	dto.Success(c, "역할별 칸반 순서가 업데이트되었습니다")
}

// UpdateBoardOrderInStage godoc
// @Summary      Update board order in stage column
// @Description  Update the display order of boards within a specific stage column (drag-and-drop)
// @Tags         user-orders
// @Accept       json
// @Produce      json
// @Param        project_id path string true "Project ID"
// @Param        stage_id path string true "Stage ID"
// @Param        request body dto.UpdateOrderRequest true "Order update details"
// @Success      200 {object} dto.SuccessResponse{data=string}
// @Failure      400 {object} dto.ErrorResponse
// @Failure      403 {object} dto.ErrorResponse
// @Router       /api/projects/{project_id}/orders/stage-boards/{stage_id} [put]
// @Security     BearerAuth
func (h *UserOrderHandler) UpdateBoardOrderInStage(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		dto.Error(c, apperrors.ErrUnauthorized)
		return
	}

	projectID := c.Param("project_id")
	stageID := c.Param("stage_id")

	if projectID == "" || stageID == "" {
		dto.Error(c, apperrors.Wrap(nil, apperrors.ErrCodeBadRequest, "프로젝트 ID와 진행단계 ID가 필요합니다", 400))
		return
	}

	var req dto.UpdateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.Error(c, apperrors.Wrap(err, apperrors.ErrCodeValidation, "입력값 검증 실패", 400))
		return
	}

	err := h.service.UpdateBoardOrderInStage(c.Request.Context(), userID, projectID, stageID, &req)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			dto.Error(c, appErr)
		} else {
			dto.Error(c, apperrors.ErrInternalServer)
		}
		return
	}

	dto.Success(c, "진행단계별 칸반 순서가 업데이트되었습니다")
}
