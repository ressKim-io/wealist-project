package dto

import (
	"board-service/internal/apperrors"

	"github.com/gin-gonic/gin"
)

// SuccessResponse represents a successful API response
type SuccessResponse struct {
	Data      interface{} `json:"data"`
	RequestID string      `json:"requestId,omitempty"`
}

// ErrorResponse represents an error API response
type ErrorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
	RequestID string `json:"requestId,omitempty"`
}

// PaginationMeta contains pagination metadata
type PaginationMeta struct {
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	PerPage    int   `json:"perPage"`
	TotalPages int   `json:"totalPages"`
}

// PaginatedResponse represents a paginated API response
type PaginatedResponse struct {
	Data       interface{}    `json:"data"`
	Pagination PaginationMeta `json:"pagination"`
	RequestID  string         `json:"requestId,omitempty"`
}

// Success sends a successful JSON response
func Success(c *gin.Context, data interface{}) {
	requestID := getRequestID(c)
	c.JSON(200, SuccessResponse{
		Data:      data,
		RequestID: requestID,
	})
}

// SuccessWithStatus sends a successful JSON response with custom status code
func SuccessWithStatus(c *gin.Context, status int, data interface{}) {
	requestID := getRequestID(c)
	c.JSON(status, SuccessResponse{
		Data:      data,
		RequestID: requestID,
	})
}

// Error sends an error JSON response
func Error(c *gin.Context, err *apperrors.AppError) {
	requestID := getRequestID(c)

	response := ErrorResponse{
		RequestID: requestID,
	}
	response.Error.Code = err.Code
	response.Error.Message = err.Message

	c.JSON(err.HTTPStatus, response)
}

// Paginated sends a paginated JSON response
func Paginated(c *gin.Context, data interface{}, meta PaginationMeta) {
	requestID := getRequestID(c)
	c.JSON(200, PaginatedResponse{
		Data:       data,
		Pagination: meta,
		RequestID:  requestID,
	})
}

// getRequestID extracts request ID from gin context
func getRequestID(c *gin.Context) string {
	if requestID, exists := c.Get("request_id"); exists {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return ""
}
