package handler

import (
	"board-service/internal/apperrors"
	"board-service/internal/dto"
	"board-service/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type CustomFieldHandler struct {
	service service.CustomFieldService
}

func NewCustomFieldHandler(service service.CustomFieldService) *CustomFieldHandler {
	return &CustomFieldHandler{service: service}
}

// ==================== Custom Roles ====================

// CreateCustomRole godoc
// @Summary      Create custom role
// @Description  Create a new custom role for a project (ADMIN+ only)
// @Tags         custom-fields
// @Accept       json
// @Produce      json
// @Param        request body dto.CreateCustomRoleRequest true "Role details"
// @Success      201 {object} dto.SuccessResponse{data=dto.CustomRoleResponse}
// @Failure      400 {object} dto.ErrorResponse
// @Failure      403 {object} dto.ErrorResponse
// @Failure      409 {object} dto.ErrorResponse
// @Router       /api/custom-fields/roles [post]
// @Security     BearerAuth
func (h *CustomFieldHandler) CreateCustomRole(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		dto.Error(c, apperrors.ErrUnauthorized)
		return
	}

	var req dto.CreateCustomRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.Error(c, apperrors.Wrap(err, apperrors.ErrCodeValidation, "입력값 검증 실패", 400))
		return
	}

	role, err := h.service.CreateCustomRole(userID, &req)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			dto.Error(c, appErr)
		} else {
			dto.Error(c, apperrors.ErrInternalServer)
		}
		return
	}

	dto.SuccessWithStatus(c, http.StatusCreated, role)
}

// GetCustomRoles godoc
// @Summary      Get custom roles
// @Description  Get all custom roles for a project
// @Tags         custom-fields
// @Accept       json
// @Produce      json
// @Param        project_id path string true "Project ID"
// @Success      200 {object} dto.SuccessResponse{data=[]dto.CustomRoleResponse}
// @Failure      403 {object} dto.ErrorResponse
// @Router       /api/custom-fields/projects/{project_id}/roles [get]
// @Security     BearerAuth
func (h *CustomFieldHandler) GetCustomRoles(c *gin.Context) {
	userID := c.GetString("user_id")
	projectID := c.Param("project_id")

	roles, err := h.service.GetCustomRoles(projectID, userID)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			dto.Error(c, appErr)
		} else {
			dto.Error(c, apperrors.ErrInternalServer)
		}
		return
	}

	dto.Success(c, roles)
}

// GetCustomRole godoc
// @Summary      Get custom role
// @Description  Get a specific custom role by ID
// @Tags         custom-fields
// @Accept       json
// @Produce      json
// @Param        role_id path string true "Role ID"
// @Success      200 {object} dto.SuccessResponse{data=dto.CustomRoleResponse}
// @Failure      404 {object} dto.ErrorResponse
// @Router       /api/custom-fields/roles/{role_id} [get]
// @Security     BearerAuth
func (h *CustomFieldHandler) GetCustomRole(c *gin.Context) {
	userID := c.GetString("user_id")
	roleID := c.Param("role_id")

	role, err := h.service.GetCustomRole(roleID, userID)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			dto.Error(c, appErr)
		} else {
			dto.Error(c, apperrors.ErrInternalServer)
		}
		return
	}

	dto.Success(c, role)
}

// UpdateCustomRole godoc
// @Summary      Update custom role
// @Description  Update a custom role (ADMIN+ only, system defaults cannot be updated)
// @Tags         custom-fields
// @Accept       json
// @Produce      json
// @Param        role_id path string true "Role ID"
// @Param        request body dto.UpdateCustomRoleRequest true "Role updates"
// @Success      200 {object} dto.SuccessResponse{data=dto.CustomRoleResponse}
// @Failure      400 {object} dto.ErrorResponse
// @Failure      403 {object} dto.ErrorResponse
// @Failure      404 {object} dto.ErrorResponse
// @Router       /api/custom-fields/roles/{role_id} [put]
// @Security     BearerAuth
func (h *CustomFieldHandler) UpdateCustomRole(c *gin.Context) {
	userID := c.GetString("user_id")
	roleID := c.Param("role_id")

	var req dto.UpdateCustomRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.Error(c, apperrors.Wrap(err, apperrors.ErrCodeValidation, "입력값 검증 실패", 400))
		return
	}

	role, err := h.service.UpdateCustomRole(roleID, userID, &req)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			dto.Error(c, appErr)
		} else {
			dto.Error(c, apperrors.ErrInternalServer)
		}
		return
	}

	dto.Success(c, role)
}

// DeleteCustomRole godoc
// @Summary      Delete custom role
// @Description  Delete a custom role (ADMIN+ only, system defaults cannot be deleted)
// @Tags         custom-fields
// @Accept       json
// @Produce      json
// @Param        role_id path string true "Role ID"
// @Success      200 {object} dto.SuccessResponse
// @Failure      403 {object} dto.ErrorResponse
// @Failure      404 {object} dto.ErrorResponse
// @Router       /api/custom-fields/roles/{role_id} [delete]
// @Security     BearerAuth
func (h *CustomFieldHandler) DeleteCustomRole(c *gin.Context) {
	userID := c.GetString("user_id")
	roleID := c.Param("role_id")

	if err := h.service.DeleteCustomRole(roleID, userID); err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			dto.Error(c, appErr)
		} else {
			dto.Error(c, apperrors.ErrInternalServer)
		}
		return
	}

	dto.Success(c, gin.H{"message": "역할이 삭제되었습니다"})
}

// UpdateCustomRoleOrder godoc
// @Summary      Update role display order
// @Description  Update the display order of custom roles (ADMIN+ only)
// @Tags         custom-fields
// @Accept       json
// @Produce      json
// @Param        project_id path string true "Project ID"
// @Param        request body dto.UpdateCustomRoleOrderRequest true "Role orders"
// @Success      200 {object} dto.SuccessResponse
// @Failure      400 {object} dto.ErrorResponse
// @Failure      403 {object} dto.ErrorResponse
// @Router       /api/custom-fields/projects/{project_id}/roles/order [put]
// @Security     BearerAuth
func (h *CustomFieldHandler) UpdateCustomRoleOrder(c *gin.Context) {
	userID := c.GetString("user_id")
	projectID := c.Param("project_id")

	var req dto.UpdateCustomRoleOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.Error(c, apperrors.Wrap(err, apperrors.ErrCodeValidation, "입력값 검증 실패", 400))
		return
	}

	if err := h.service.UpdateCustomRoleOrder(projectID, userID, &req); err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			dto.Error(c, appErr)
		} else {
			dto.Error(c, apperrors.ErrInternalServer)
		}
		return
	}

	dto.Success(c, gin.H{"message": "역할 순서가 변경되었습니다"})
}

// ==================== Custom Stages ====================

// CreateCustomStage godoc
// @Summary      Create custom stage
// @Description  Create a new custom stage for a project (ADMIN+ only)
// @Tags         custom-fields
// @Accept       json
// @Produce      json
// @Param        request body dto.CreateCustomStageRequest true "Stage details"
// @Success      201 {object} dto.SuccessResponse{data=dto.CustomStageResponse}
// @Failure      400 {object} dto.ErrorResponse
// @Failure      403 {object} dto.ErrorResponse
// @Router       /api/custom-fields/stages [post]
// @Security     BearerAuth
func (h *CustomFieldHandler) CreateCustomStage(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		dto.Error(c, apperrors.ErrUnauthorized)
		return
	}

	var req dto.CreateCustomStageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.Error(c, apperrors.Wrap(err, apperrors.ErrCodeValidation, "입력값 검증 실패", 400))
		return
	}

	stage, err := h.service.CreateCustomStage(userID, &req)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			dto.Error(c, appErr)
		} else {
			dto.Error(c, apperrors.ErrInternalServer)
		}
		return
	}

	dto.SuccessWithStatus(c, http.StatusCreated, stage)
}

// GetCustomStages godoc
// @Summary      Get custom stages
// @Description  Get all custom stages for a project
// @Tags         custom-fields
// @Accept       json
// @Produce      json
// @Param        project_id path string true "Project ID"
// @Success      200 {object} dto.SuccessResponse{data=[]dto.CustomStageResponse}
// @Failure      403 {object} dto.ErrorResponse
// @Router       /api/custom-fields/projects/{project_id}/stages [get]
// @Security     BearerAuth
func (h *CustomFieldHandler) GetCustomStages(c *gin.Context) {
	userID := c.GetString("user_id")
	projectID := c.Param("project_id")

	stages, err := h.service.GetCustomStages(projectID, userID)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			dto.Error(c, appErr)
		} else {
			dto.Error(c, apperrors.ErrInternalServer)
		}
		return
	}

	dto.Success(c, stages)
}

// GetCustomStage godoc
// @Summary      Get custom stage
// @Description  Get a specific custom stage by ID
// @Tags         custom-fields
// @Accept       json
// @Produce      json
// @Param        stage_id path string true "Stage ID"
// @Success      200 {object} dto.SuccessResponse{data=dto.CustomStageResponse}
// @Failure      404 {object} dto.ErrorResponse
// @Router       /api/custom-fields/stages/{stage_id} [get]
// @Security     BearerAuth
func (h *CustomFieldHandler) GetCustomStage(c *gin.Context) {
	userID := c.GetString("user_id")
	stageID := c.Param("stage_id")

	stage, err := h.service.GetCustomStage(stageID, userID)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			dto.Error(c, appErr)
		} else {
			dto.Error(c, apperrors.ErrInternalServer)
		}
		return
	}

	dto.Success(c, stage)
}

// UpdateCustomStage godoc
// @Summary      Update custom stage
// @Description  Update a custom stage (ADMIN+ only, system defaults cannot be updated)
// @Tags         custom-fields
// @Accept       json
// @Produce      json
// @Param        stage_id path string true "Stage ID"
// @Param        request body dto.UpdateCustomStageRequest true "Stage updates"
// @Success      200 {object} dto.SuccessResponse{data=dto.CustomStageResponse}
// @Failure      400 {object} dto.ErrorResponse
// @Failure      403 {object} dto.ErrorResponse
// @Router       /api/custom-fields/stages/{stage_id} [put]
// @Security     BearerAuth
func (h *CustomFieldHandler) UpdateCustomStage(c *gin.Context) {
	userID := c.GetString("user_id")
	stageID := c.Param("stage_id")

	var req dto.UpdateCustomStageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.Error(c, apperrors.Wrap(err, apperrors.ErrCodeValidation, "입력값 검증 실패", 400))
		return
	}

	stage, err := h.service.UpdateCustomStage(stageID, userID, &req)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			dto.Error(c, appErr)
		} else {
			dto.Error(c, apperrors.ErrInternalServer)
		}
		return
	}

	dto.Success(c, stage)
}

// DeleteCustomStage godoc
// @Summary      Delete custom stage
// @Description  Delete a custom stage (ADMIN+ only, system defaults cannot be deleted)
// @Tags         custom-fields
// @Accept       json
// @Produce      json
// @Param        stage_id path string true "Stage ID"
// @Success      200 {object} dto.SuccessResponse
// @Failure      403 {object} dto.ErrorResponse
// @Router       /api/custom-fields/stages/{stage_id} [delete]
// @Security     BearerAuth
func (h *CustomFieldHandler) DeleteCustomStage(c *gin.Context) {
	userID := c.GetString("user_id")
	stageID := c.Param("stage_id")

	if err := h.service.DeleteCustomStage(stageID, userID); err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			dto.Error(c, appErr)
		} else {
			dto.Error(c, apperrors.ErrInternalServer)
		}
		return
	}

	dto.Success(c, gin.H{"message": "단계가 삭제되었습니다"})
}

// UpdateCustomStageOrder godoc
// @Summary      Update stage display order
// @Description  Update the display order of custom stages (ADMIN+ only)
// @Tags         custom-fields
// @Accept       json
// @Produce      json
// @Param        project_id path string true "Project ID"
// @Param        request body dto.UpdateCustomStageOrderRequest true "Stage orders"
// @Success      200 {object} dto.SuccessResponse
// @Failure      400 {object} dto.ErrorResponse
// @Router       /api/custom-fields/projects/{project_id}/stages/order [put]
// @Security     BearerAuth
func (h *CustomFieldHandler) UpdateCustomStageOrder(c *gin.Context) {
	userID := c.GetString("user_id")
	projectID := c.Param("project_id")

	var req dto.UpdateCustomStageOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.Error(c, apperrors.Wrap(err, apperrors.ErrCodeValidation, "입력값 검증 실패", 400))
		return
	}

	if err := h.service.UpdateCustomStageOrder(projectID, userID, &req); err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			dto.Error(c, appErr)
		} else {
			dto.Error(c, apperrors.ErrInternalServer)
		}
		return
	}

	dto.Success(c, gin.H{"message": "단계 순서가 변경되었습니다"})
}

// ==================== Custom Importance ====================

// CreateCustomImportance godoc
// @Summary      Create custom importance
// @Description  Create a new custom importance for a project (ADMIN+ only)
// @Tags         custom-fields
// @Accept       json
// @Produce      json
// @Param        request body dto.CreateCustomImportanceRequest true "Importance details"
// @Success      201 {object} dto.SuccessResponse{data=dto.CustomImportanceResponse}
// @Failure      400 {object} dto.ErrorResponse
// @Router       /api/custom-fields/importance [post]
// @Security     BearerAuth
func (h *CustomFieldHandler) CreateCustomImportance(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		dto.Error(c, apperrors.ErrUnauthorized)
		return
	}

	var req dto.CreateCustomImportanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.Error(c, apperrors.Wrap(err, apperrors.ErrCodeValidation, "입력값 검증 실패", 400))
		return
	}

	importance, err := h.service.CreateCustomImportance(userID, &req)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			dto.Error(c, appErr)
		} else {
			dto.Error(c, apperrors.ErrInternalServer)
		}
		return
	}

	dto.SuccessWithStatus(c, http.StatusCreated, importance)
}

// GetCustomImportances godoc
// @Summary      Get custom importances
// @Description  Get all custom importances for a project
// @Tags         custom-fields
// @Accept       json
// @Produce      json
// @Param        project_id path string true "Project ID"
// @Success      200 {object} dto.SuccessResponse{data=[]dto.CustomImportanceResponse}
// @Failure      403 {object} dto.ErrorResponse
// @Router       /api/custom-fields/projects/{project_id}/importance [get]
// @Security     BearerAuth
func (h *CustomFieldHandler) GetCustomImportances(c *gin.Context) {
	userID := c.GetString("user_id")
	projectID := c.Param("project_id")

	importances, err := h.service.GetCustomImportances(projectID, userID)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			dto.Error(c, appErr)
		} else {
			dto.Error(c, apperrors.ErrInternalServer)
		}
		return
	}

	dto.Success(c, importances)
}

// GetCustomImportance godoc
// @Summary      Get custom importance
// @Description  Get a specific custom importance by ID
// @Tags         custom-fields
// @Accept       json
// @Produce      json
// @Param        importance_id path string true "Importance ID"
// @Success      200 {object} dto.SuccessResponse{data=dto.CustomImportanceResponse}
// @Failure      404 {object} dto.ErrorResponse
// @Router       /api/custom-fields/importance/{importance_id} [get]
// @Security     BearerAuth
func (h *CustomFieldHandler) GetCustomImportance(c *gin.Context) {
	userID := c.GetString("user_id")
	importanceID := c.Param("importance_id")

	importance, err := h.service.GetCustomImportance(importanceID, userID)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			dto.Error(c, appErr)
		} else {
			dto.Error(c, apperrors.ErrInternalServer)
		}
		return
	}

	dto.Success(c, importance)
}

// UpdateCustomImportance godoc
// @Summary      Update custom importance
// @Description  Update a custom importance (ADMIN+ only, system defaults cannot be updated)
// @Tags         custom-fields
// @Accept       json
// @Produce      json
// @Param        importance_id path string true "Importance ID"
// @Param        request body dto.UpdateCustomImportanceRequest true "Importance updates"
// @Success      200 {object} dto.SuccessResponse{data=dto.CustomImportanceResponse}
// @Failure      400 {object} dto.ErrorResponse
// @Router       /api/custom-fields/importance/{importance_id} [put]
// @Security     BearerAuth
func (h *CustomFieldHandler) UpdateCustomImportance(c *gin.Context) {
	userID := c.GetString("user_id")
	importanceID := c.Param("importance_id")

	var req dto.UpdateCustomImportanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.Error(c, apperrors.Wrap(err, apperrors.ErrCodeValidation, "입력값 검증 실패", 400))
		return
	}

	importance, err := h.service.UpdateCustomImportance(importanceID, userID, &req)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			dto.Error(c, appErr)
		} else {
			dto.Error(c, apperrors.ErrInternalServer)
		}
		return
	}

	dto.Success(c, importance)
}

// DeleteCustomImportance godoc
// @Summary      Delete custom importance
// @Description  Delete a custom importance (ADMIN+ only, system defaults cannot be deleted)
// @Tags         custom-fields
// @Accept       json
// @Produce      json
// @Param        importance_id path string true "Importance ID"
// @Success      200 {object} dto.SuccessResponse
// @Failure      403 {object} dto.ErrorResponse
// @Router       /api/custom-fields/importance/{importance_id} [delete]
// @Security     BearerAuth
func (h *CustomFieldHandler) DeleteCustomImportance(c *gin.Context) {
	userID := c.GetString("user_id")
	importanceID := c.Param("importance_id")

	if err := h.service.DeleteCustomImportance(importanceID, userID); err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			dto.Error(c, appErr)
		} else {
			dto.Error(c, apperrors.ErrInternalServer)
		}
		return
	}

	dto.Success(c, gin.H{"message": "중요도가 삭제되었습니다"})
}

// UpdateCustomImportanceOrder godoc
// @Summary      Update importance display order
// @Description  Update the display order of custom importances (ADMIN+ only)
// @Tags         custom-fields
// @Accept       json
// @Produce      json
// @Param        project_id path string true "Project ID"
// @Param        request body dto.UpdateCustomImportanceOrderRequest true "Importance orders"
// @Success      200 {object} dto.SuccessResponse
// @Failure      400 {object} dto.ErrorResponse
// @Router       /api/custom-fields/projects/{project_id}/importance/order [put]
// @Security     BearerAuth
func (h *CustomFieldHandler) UpdateCustomImportanceOrder(c *gin.Context) {
	userID := c.GetString("user_id")
	projectID := c.Param("project_id")

	var req dto.UpdateCustomImportanceOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.Error(c, apperrors.Wrap(err, apperrors.ErrCodeValidation, "입력값 검증 실패", 400))
		return
	}

	if err := h.service.UpdateCustomImportanceOrder(projectID, userID, &req); err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			dto.Error(c, appErr)
		} else {
			dto.Error(c, apperrors.ErrInternalServer)
		}
		return
	}

	dto.Success(c, gin.H{"message": "중요도 순서가 변경되었습니다"})
}
