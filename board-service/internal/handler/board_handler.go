package handler

import (
	"board-service/internal/apperrors"
	"board-service/internal/dto"
	"board-service/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type BoardHandler struct {
	service service.BoardService
}

func NewBoardHandler(service service.BoardService) *BoardHandler {
	return &BoardHandler{service: service}
}

// CreateBoard godoc
// @Summary      Create board
// @Description  Create a new board card (task/issue) in a project
// @Tags         boards
// @Accept       json
// @Produce      json
// @Param        request body dto.CreateBoardRequest true "Board details"
// @Success      201 {object} dto.SuccessResponse{data=dto.BoardResponse}
// @Failure      400 {object} dto.ErrorResponse
// @Failure      403 {object} dto.ErrorResponse
// @Failure      404 {object} dto.ErrorResponse
// @Router       /api/boards [post]
// @Security     BearerAuth
func (h *BoardHandler) CreateBoard(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		dto.Error(c, apperrors.ErrUnauthorized)
		return
	}

	var req dto.CreateBoardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.Error(c, apperrors.Wrap(err, apperrors.ErrCodeValidation, "입력값 검증 실패", 400))
		return
	}

	board, err := h.service.CreateBoard(userID, &req)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			dto.Error(c, appErr)
		} else {
			dto.Error(c, apperrors.ErrInternalServer)
		}
		return
	}

	dto.SuccessWithStatus(c, http.StatusCreated, board)
}

// GetBoard godoc
// @Summary      Get board
// @Description  Get a specific board by ID (project member only)
// @Tags         boards
// @Accept       json
// @Produce      json
// @Param        boardId path string true "Board ID"
// @Success      200 {object} dto.SuccessResponse{data=dto.BoardResponse}
// @Failure      403 {object} dto.ErrorResponse
// @Failure      404 {object} dto.ErrorResponse
// @Router       /api/boards/{boardId} [get]
// @Security     BearerAuth
func (h *BoardHandler) GetBoard(c *gin.Context) {
	userID := c.GetString("user_id")
	boardID := c.Param("boardId")

	board, err := h.service.GetBoard(boardID, userID)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			dto.Error(c, appErr)
		} else {
			dto.Error(c, apperrors.ErrInternalServer)
		}
		return
	}

	dto.Success(c, board)
}

// GetBoards godoc
// @Summary      Get boards
// @Description  Get boards for a project with optional filters
// @Tags         boards
// @Accept       json
// @Produce      json
// @Param        projectId query string true "Project ID"
// @Param        stageId query string false "Filter by Stage ID"
// @Param        roleId query string false "Filter by Role ID"
// @Param        importanceId query string false "Filter by Importance ID"
// @Param        assigneeId query string false "Filter by Assignee ID"
// @Param        authorId query string false "Filter by Author ID"
// @Param        page query int false "Page number (default: 1)"
// @Param        limit query int false "Items per page (default: 20, max: 100)"
// @Success      200 {object} dto.SuccessResponse{data=dto.PaginatedBoardsResponse}
// @Failure      400 {object} dto.ErrorResponse
// @Failure      403 {object} dto.ErrorResponse
// @Router       /api/boards [get]
// @Security     BearerAuth
func (h *BoardHandler) GetBoards(c *gin.Context) {
	userID := c.GetString("user_id")

	var req dto.GetBoardsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		dto.Error(c, apperrors.Wrap(err, apperrors.ErrCodeValidation, "입력값 검증 실패", 400))
		return
	}

	// Default pagination
	if req.Page == 0 {
		req.Page = 1
	}
	if req.Limit == 0 {
		req.Limit = 20
	}

	boards, err := h.service.GetBoards(userID, &req)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			dto.Error(c, appErr)
		} else {
			dto.Error(c, apperrors.ErrInternalServer)
		}
		return
	}

	dto.Success(c, boards)
}

// UpdateBoard godoc
// @Summary      Update board
// @Description  Update a board (author or ADMIN+ only)
// @Tags         boards
// @Accept       json
// @Produce      json
// @Param        boardId path string true "Board ID"
// @Param        request body dto.UpdateBoardRequest true "Board updates"
// @Success      200 {object} dto.SuccessResponse{data=dto.BoardResponse}
// @Failure      400 {object} dto.ErrorResponse
// @Failure      403 {object} dto.ErrorResponse
// @Failure      404 {object} dto.ErrorResponse
// @Router       /api/boards/{boardId} [put]
// @Security     BearerAuth
func (h *BoardHandler) UpdateBoard(c *gin.Context) {
	userID := c.GetString("user_id")
	boardID := c.Param("boardId")

	var req dto.UpdateBoardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.Error(c, apperrors.Wrap(err, apperrors.ErrCodeValidation, "입력값 검증 실패", 400))
		return
	}

	board, err := h.service.UpdateBoard(boardID, userID, &req)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			dto.Error(c, appErr)
		} else {
			dto.Error(c, apperrors.ErrInternalServer)
		}
		return
	}

	dto.Success(c, board)
}

// DeleteBoard godoc
// @Summary      Delete board
// @Description  Delete a board (soft delete, author or ADMIN+ only)
// @Tags         boards
// @Accept       json
// @Produce      json
// @Param        boardId path string true "Board ID"
// @Success      200 {object} dto.SuccessResponse
// @Failure      403 {object} dto.ErrorResponse
// @Failure      404 {object} dto.ErrorResponse
// @Router       /api/boards/{boardId} [delete]
// @Security     BearerAuth
func (h *BoardHandler) DeleteBoard(c *gin.Context) {
	userID := c.GetString("user_id")
	boardID := c.Param("boardId")

	if err := h.service.DeleteBoard(boardID, userID); err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			dto.Error(c, appErr)
		} else {
			dto.Error(c, apperrors.ErrInternalServer)
		}
		return
	}

	dto.Success(c, gin.H{"message": "보드가 삭제되었습니다"})
}

// MoveBoard godoc
// @Summary      Move board to different column
// @Description  Move a board to a different column/group in a view (integrated API: field value change + order update in single transaction)
// @Tags         boards
// @Accept       json
// @Produce      json
// @Param        boardId path string true "Board ID"
// @Param        request body dto.MoveBoardRequest true "Move board request"
// @Success      200 {object} dto.SuccessResponse{data=dto.MoveBoardResponse}
// @Failure      400 {object} dto.ErrorResponse
// @Failure      403 {object} dto.ErrorResponse
// @Failure      404 {object} dto.ErrorResponse
// @Router       /api/boards/{boardId}/move [put]
// @Security     BearerAuth
func (h *BoardHandler) MoveBoard(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		dto.Error(c, apperrors.ErrUnauthorized)
		return
	}

	boardID := c.Param("boardId")
	if boardID == "" {
		dto.Error(c, apperrors.Wrap(nil, apperrors.ErrCodeBadRequest, "보드 ID가 필요합니다", 400))
		return
	}

	var req dto.MoveBoardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.Error(c, apperrors.Wrap(err, apperrors.ErrCodeValidation, "입력값 검증 실패", 400))
		return
	}

	response, err := h.service.MoveBoard(userID, boardID, &req)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			dto.Error(c, appErr)
		} else {
			dto.Error(c, apperrors.ErrInternalServer)
		}
		return
	}

	dto.Success(c, response)
}
