package apperrors

import (
	"board-service/internal/domain"
	"errors"
	"fmt"
)

// ==================== Infrastructure Error (AppError) ====================
// AppError represents an infrastructure-level error with HTTP status
// This wraps domain errors and adds HTTP context

// AppError represents an application-specific error
type AppError struct {
	Code       string // Error code (e.g., "UNAUTHORIZED", "NOT_FOUND")
	Message    string // User-friendly message
	HTTPStatus int    // HTTP status code
	Err        error  // Original error (for logging)
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap returns the underlying error
func (e *AppError) Unwrap() error {
	return e.Err
}

// New creates a new AppError
func New(code, message string, status int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: status,
	}
}

// Wrap creates a new AppError wrapping an existing error
func Wrap(err error, code, message string, status int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: status,
		Err:        err,
	}
}

// ==================== Domain Error Conversion ====================
// Convert domain errors to infrastructure errors with appropriate HTTP status

// FromDomainError converts a domain.DomainError to an AppError
func FromDomainError(err error) *AppError {
	var domainErr *domain.DomainError
	if errors.As(err, &domainErr) {
		switch domainErr.Code {
		case domain.ErrCodeValidation:
			return &AppError{
				Code:       ErrCodeBadRequest,
				Message:    domainErr.Message,
				HTTPStatus: 400,
				Err:        domainErr,
			}
		case domain.ErrCodeBusinessRule:
			return &AppError{
				Code:       ErrCodeForbidden,
				Message:    domainErr.Message,
				HTTPStatus: 403,
				Err:        domainErr,
			}
		case domain.ErrCodeInvalidState:
			return &AppError{
				Code:       ErrCodeConflict,
				Message:    domainErr.Message,
				HTTPStatus: 409,
				Err:        domainErr,
			}
		default:
			return &AppError{
				Code:       ErrCodeBadRequest,
				Message:    domainErr.Message,
				HTTPStatus: 400,
				Err:        domainErr,
			}
		}
	}

	// Legacy ValidationError support
	var validationErr *domain.ValidationError
	if errors.As(err, &validationErr) {
		return &AppError{
			Code:       ErrCodeBadRequest,
			Message:    validationErr.Message,
			HTTPStatus: 400,
			Err:        validationErr,
		}
	}

	// Unknown error
	return &AppError{
		Code:       ErrCodeInternalServer,
		Message:    "내부 서버 오류",
		HTTPStatus: 500,
		Err:        err,
	}
}
