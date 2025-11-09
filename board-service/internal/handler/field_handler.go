package handler

import (
	"board-service/internal/apperrors"
	"board-service/internal/dto"
	"board-service/internal/middleware"
	"board-service/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type FieldHandler struct {
	fieldService      service.FieldService
	fieldValueService service.FieldValueService
}

func NewFieldHandler(fieldService service.FieldService, fieldValueService service.FieldValueService) *FieldHandler {
	return &FieldHandler{
		fieldService:      fieldService,
		fieldValueService: fieldValueService,
	}
}

// ==================== Field CRUD ====================

// CreateField godoc
// @Summary Create a custom field
// @Description Create a new custom field for a project
// @Tags Fields
// @Accept json
// @Produce json
// @Param request body dto.CreateFieldRequest true "Field creation request"
// @Success 201 {object} dto.Success{data=dto.FieldResponse}
// @Failure 400 {object} dto.Error
// @Failure 401 {object} dto.Error
// @Failure 403 {object} dto.Error
// @Failure 500 {object} dto.Error
// @Router /fields [post]
// @Security BearerAuth
func (h *FieldHandler) CreateField(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)

	var req dto.CreateFieldRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		appErr := apperrors.Wrap(err, apperrors.ErrCodeValidation, "입력값이 유효하지 않습니다", 400)
		dto.Error(c, appErr)
		return
	}

	field, err := h.fieldService.CreateField(userID, &req)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			dto.Error(c, appErr)
		} else {
			dto.Error(c, apperrors.New(apperrors.ErrCodeInternalServer, "필드 생성 실패", 500))
		}
		return
	}

	dto.SuccessWithStatus(c, http.StatusCreated, field)
}

// GetFieldsByProject godoc
// @Summary Get fields by project
// @Description Get all custom fields for a project
// @Tags Fields
// @Accept json
// @Produce json
// @Param project_id path string true "Project ID"
// @Success 200 {object} dto.Success{data=[]dto.FieldResponse}
// @Failure 400 {object} dto.Error
// @Failure 401 {object} dto.Error
// @Failure 403 {object} dto.Error
// @Failure 500 {object} dto.Error
// @Router /projects/{project_id}/fields [get]
// @Security BearerAuth
func (h *FieldHandler) GetFieldsByProject(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	projectID := c.Param("project_id")

	fields, err := h.fieldService.GetFieldsByProject(userID, projectID)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			dto.Error(c, appErr)
		} else {
			dto.Error(c, apperrors.New(apperrors.ErrCodeInternalServer, "필드 조회 실패", 500))
		}
		return
	}

	dto.Success(c, fields)
}

// GetField godoc
// @Summary Get field by ID
// @Description Get a custom field by its ID
// @Tags Fields
// @Accept json
// @Produce json
// @Param field_id path string true "Field ID"
// @Success 200 {object} dto.Success{data=dto.FieldResponse}
// @Failure 400 {object} dto.Error
// @Failure 401 {object} dto.Error
// @Failure 403 {object} dto.Error
// @Failure 404 {object} dto.Error
// @Failure 500 {object} dto.Error
// @Router /fields/{field_id} [get]
// @Security BearerAuth
func (h *FieldHandler) GetField(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	fieldID := c.Param("field_id")

	field, err := h.fieldService.GetField(userID, fieldID)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			dto.Error(c, appErr)
		} else {
			dto.Error(c, apperrors.New(apperrors.ErrCodeInternalServer, "필드 조회 실패", 500))
		}
		return
	}

	dto.Success(c, field)
}

// UpdateField godoc
// @Summary Update field
// @Description Update a custom field
// @Tags Fields
// @Accept json
// @Produce json
// @Param field_id path string true "Field ID"
// @Param request body dto.UpdateFieldRequest true "Field update request"
// @Success 200 {object} dto.Success{data=dto.FieldResponse}
// @Failure 400 {object} dto.Error
// @Failure 401 {object} dto.Error
// @Failure 403 {object} dto.Error
// @Failure 404 {object} dto.Error
// @Failure 500 {object} dto.Error
// @Router /fields/{field_id} [patch]
// @Security BearerAuth
func (h *FieldHandler) UpdateField(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	fieldID := c.Param("field_id")

	var req dto.UpdateFieldRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		appErr := apperrors.Wrap(err, apperrors.ErrCodeValidation, "입력값이 유효하지 않습니다", 400)
		dto.Error(c, appErr)
		return
	}

	field, err := h.fieldService.UpdateField(userID, fieldID, &req)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			dto.Error(c, appErr)
		} else {
			dto.Error(c, apperrors.New(apperrors.ErrCodeInternalServer, "필드 수정 실패", 500))
		}
		return
	}

	dto.Success(c, field)
}

// DeleteField godoc
// @Summary Delete field
// @Description Delete a custom field
// @Tags Fields
// @Accept json
// @Produce json
// @Param field_id path string true "Field ID"
// @Success 204
// @Failure 400 {object} dto.Error
// @Failure 401 {object} dto.Error
// @Failure 403 {object} dto.Error
// @Failure 404 {object} dto.Error
// @Failure 500 {object} dto.Error
// @Router /fields/{field_id} [delete]
// @Security BearerAuth
func (h *FieldHandler) DeleteField(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	fieldID := c.Param("field_id")

	if err := h.fieldService.DeleteField(userID, fieldID); err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			dto.Error(c, appErr)
		} else {
			dto.Error(c, apperrors.New(apperrors.ErrCodeInternalServer, "필드 삭제 실패", 500))
		}
		return
	}

	c.Status(http.StatusNoContent)
}

// UpdateFieldOrder godoc
// @Summary Update field order
// @Description Update display order of fields in a project
// @Tags Fields
// @Accept json
// @Produce json
// @Param project_id path string true "Project ID"
// @Param request body dto.UpdateFieldOrderRequest true "Field order update request"
// @Success 204
// @Failure 400 {object} dto.Error
// @Failure 401 {object} dto.Error
// @Failure 403 {object} dto.Error
// @Failure 500 {object} dto.Error
// @Router /projects/{project_id}/fields/order [put]
// @Security BearerAuth
func (h *FieldHandler) UpdateFieldOrder(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	projectID := c.Param("project_id")

	var req dto.UpdateFieldOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		appErr := apperrors.Wrap(err, apperrors.ErrCodeValidation, "입력값이 유효하지 않습니다", 400)
		dto.Error(c, appErr)
		return
	}

	if err := h.fieldService.UpdateFieldOrder(userID, projectID, &req); err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			dto.Error(c, appErr)
		} else {
			dto.Error(c, apperrors.New(apperrors.ErrCodeInternalServer, "필드 순서 업데이트 실패", 500))
		}
		return
	}

	c.Status(http.StatusNoContent)
}

// ==================== Option CRUD ====================

// CreateOption godoc
// @Summary Create field option
// @Description Create a new option for a select field
// @Tags Field Options
// @Accept json
// @Produce json
// @Param request body dto.CreateOptionRequest true "Option creation request"
// @Success 201 {object} dto.Success{data=dto.OptionResponse}
// @Failure 400 {object} dto.Error
// @Failure 401 {object} dto.Error
// @Failure 403 {object} dto.Error
// @Failure 500 {object} dto.Error
// @Router /field-options [post]
// @Security BearerAuth
func (h *FieldHandler) CreateOption(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)

	var req dto.CreateOptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		appErr := apperrors.Wrap(err, apperrors.ErrCodeValidation, "입력값이 유효하지 않습니다", 400)
		dto.Error(c, appErr)
		return
	}

	option, err := h.fieldService.CreateOption(userID, &req)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			dto.Error(c, appErr)
		} else {
			dto.Error(c, apperrors.New(apperrors.ErrCodeInternalServer, "옵션 생성 실패", 500))
		}
		return
	}

	dto.SuccessWithStatus(c, http.StatusCreated, option)
}

// GetOptionsByField godoc
// @Summary Get options by field
// @Description Get all options for a select field
// @Tags Field Options
// @Accept json
// @Produce json
// @Param field_id path string true "Field ID"
// @Success 200 {object} dto.Success{data=[]dto.OptionResponse}
// @Failure 400 {object} dto.Error
// @Failure 401 {object} dto.Error
// @Failure 403 {object} dto.Error
// @Failure 500 {object} dto.Error
// @Router /fields/{field_id}/options [get]
// @Security BearerAuth
func (h *FieldHandler) GetOptionsByField(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	fieldID := c.Param("field_id")

	options, err := h.fieldService.GetOptionsByField(userID, fieldID)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			dto.Error(c, appErr)
		} else {
			dto.Error(c, apperrors.New(apperrors.ErrCodeInternalServer, "옵션 조회 실패", 500))
		}
		return
	}

	dto.Success(c, options)
}

// UpdateOption godoc
// @Summary Update field option
// @Description Update a field option
// @Tags Field Options
// @Accept json
// @Produce json
// @Param option_id path string true "Option ID"
// @Param request body dto.UpdateOptionRequest true "Option update request"
// @Success 200 {object} dto.Success{data=dto.OptionResponse}
// @Failure 400 {object} dto.Error
// @Failure 401 {object} dto.Error
// @Failure 403 {object} dto.Error
// @Failure 404 {object} dto.Error
// @Failure 500 {object} dto.Error
// @Router /field-options/{option_id} [patch]
// @Security BearerAuth
func (h *FieldHandler) UpdateOption(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	optionID := c.Param("option_id")

	var req dto.UpdateOptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		appErr := apperrors.Wrap(err, apperrors.ErrCodeValidation, "입력값이 유효하지 않습니다", 400)
		dto.Error(c, appErr)
		return
	}

	option, err := h.fieldService.UpdateOption(userID, optionID, &req)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			dto.Error(c, appErr)
		} else {
			dto.Error(c, apperrors.New(apperrors.ErrCodeInternalServer, "옵션 수정 실패", 500))
		}
		return
	}

	dto.Success(c, option)
}

// DeleteOption godoc
// @Summary Delete field option
// @Description Delete a field option
// @Tags Field Options
// @Accept json
// @Produce json
// @Param option_id path string true "Option ID"
// @Success 204
// @Failure 400 {object} dto.Error
// @Failure 401 {object} dto.Error
// @Failure 403 {object} dto.Error
// @Failure 404 {object} dto.Error
// @Failure 500 {object} dto.Error
// @Router /field-options/{option_id} [delete]
// @Security BearerAuth
func (h *FieldHandler) DeleteOption(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	optionID := c.Param("option_id")

	if err := h.fieldService.DeleteOption(userID, optionID); err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			dto.Error(c, appErr)
		} else {
			dto.Error(c, apperrors.New(apperrors.ErrCodeInternalServer, "옵션 삭제 실패", 500))
		}
		return
	}

	c.Status(http.StatusNoContent)
}

// UpdateOptionOrder godoc
// @Summary Update option order
// @Description Update display order of options for a field
// @Tags Field Options
// @Accept json
// @Produce json
// @Param field_id path string true "Field ID"
// @Param request body dto.UpdateOptionOrderRequest true "Option order update request"
// @Success 204
// @Failure 400 {object} dto.Error
// @Failure 401 {object} dto.Error
// @Failure 403 {object} dto.Error
// @Failure 500 {object} dto.Error
// @Router /fields/{field_id}/options/order [put]
// @Security BearerAuth
func (h *FieldHandler) UpdateOptionOrder(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	fieldID := c.Param("field_id")

	var req dto.UpdateOptionOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		appErr := apperrors.Wrap(err, apperrors.ErrCodeValidation, "입력값이 유효하지 않습니다", 400)
		dto.Error(c, appErr)
		return
	}

	if err := h.fieldService.UpdateOptionOrder(userID, fieldID, &req); err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			dto.Error(c, appErr)
		} else {
			dto.Error(c, apperrors.New(apperrors.ErrCodeInternalServer, "옵션 순서 업데이트 실패", 500))
		}
		return
	}

	c.Status(http.StatusNoContent)
}

// ==================== Field Value Operations ====================

// SetFieldValue godoc
// @Summary Set field value
// @Description Set a field value for a board
// @Tags Field Values
// @Accept json
// @Produce json
// @Param request body dto.SetFieldValueRequest true "Field value request"
// @Success 204
// @Failure 400 {object} dto.Error
// @Failure 401 {object} dto.Error
// @Failure 403 {object} dto.Error
// @Failure 500 {object} dto.Error
// @Router /board-field-values [post]
// @Security BearerAuth
func (h *FieldHandler) SetFieldValue(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)

	var req dto.SetFieldValueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		appErr := apperrors.Wrap(err, apperrors.ErrCodeValidation, "입력값이 유효하지 않습니다", 400)
		dto.Error(c, appErr)
		return
	}

	if err := h.fieldValueService.SetFieldValue(userID, &req); err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			dto.Error(c, appErr)
		} else {
			dto.Error(c, apperrors.New(apperrors.ErrCodeInternalServer, "필드 값 설정 실패", 500))
		}
		return
	}

	c.Status(http.StatusNoContent)
}

// GetBoardFieldValues godoc
// @Summary Get board field values
// @Description Get all field values for a board
// @Tags Field Values
// @Accept json
// @Produce json
// @Param board_id path string true "Board ID"
// @Success 200 {object} dto.Success{data=dto.BoardFieldValuesResponse}
// @Failure 400 {object} dto.Error
// @Failure 401 {object} dto.Error
// @Failure 403 {object} dto.Error
// @Failure 404 {object} dto.Error
// @Failure 500 {object} dto.Error
// @Router /boards/{board_id}/field-values [get]
// @Security BearerAuth
func (h *FieldHandler) GetBoardFieldValues(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	boardID := c.Param("board_id")

	values, err := h.fieldValueService.GetBoardFieldValues(userID, boardID)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			dto.Error(c, appErr)
		} else {
			dto.Error(c, apperrors.New(apperrors.ErrCodeInternalServer, "필드 값 조회 실패", 500))
		}
		return
	}

	dto.Success(c, values)
}

// DeleteFieldValue godoc
// @Summary Delete field value
// @Description Delete a field value for a board
// @Tags Field Values
// @Accept json
// @Produce json
// @Param board_id path string true "Board ID"
// @Param field_id path string true "Field ID"
// @Success 204
// @Failure 400 {object} dto.Error
// @Failure 401 {object} dto.Error
// @Failure 403 {object} dto.Error
// @Failure 500 {object} dto.Error
// @Router /boards/{board_id}/field-values/{field_id} [delete]
// @Security BearerAuth
func (h *FieldHandler) DeleteFieldValue(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	boardID := c.Param("board_id")
	fieldID := c.Param("field_id")

	if err := h.fieldValueService.DeleteFieldValue(userID, boardID, fieldID); err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			dto.Error(c, appErr)
		} else {
			dto.Error(c, apperrors.New(apperrors.ErrCodeInternalServer, "필드 값 삭제 실패", 500))
		}
		return
	}

	c.Status(http.StatusNoContent)
}

// SetMultiSelectValue godoc
// @Summary Set multi-select field value
// @Description Set ordered multi-select values for a board field
// @Tags Field Values
// @Accept json
// @Produce json
// @Param request body dto.SetMultiSelectValueRequest true "Multi-select value request"
// @Success 204
// @Failure 400 {object} dto.Error
// @Failure 401 {object} dto.Error
// @Failure 403 {object} dto.Error
// @Failure 500 {object} dto.Error
// @Router /board-field-values/multi-select [post]
// @Security BearerAuth
func (h *FieldHandler) SetMultiSelectValue(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)

	var req dto.SetMultiSelectValueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		appErr := apperrors.Wrap(err, apperrors.ErrCodeValidation, "입력값이 유효하지 않습니다", 400)
		dto.Error(c, appErr)
		return
	}

	if err := h.fieldValueService.SetMultiSelectValue(userID, &req); err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			dto.Error(c, appErr)
		} else {
			dto.Error(c, apperrors.New(apperrors.ErrCodeInternalServer, "Multi-select 값 설정 실패", 500))
		}
		return
	}

	c.Status(http.StatusNoContent)
}
