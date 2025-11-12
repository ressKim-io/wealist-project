package domain

import "fmt"

// ==================== Domain Error Types ====================
// Domain 레벨 에러는 비즈니스 규칙 위반을 나타냅니다.
// Infrastructure 레벨 에러(AppError)와 명확히 분리됩니다.

// DomainErrorCode represents domain-specific error codes
type DomainErrorCode string

const (
	// Validation errors
	ErrCodeValidation DomainErrorCode = "VALIDATION_ERROR"

	// Business rule errors
	ErrCodeBusinessRule DomainErrorCode = "BUSINESS_RULE_VIOLATION"

	// State errors
	ErrCodeInvalidState DomainErrorCode = "INVALID_STATE"
)

// DomainError represents a domain-level error (business logic violation)
type DomainError struct {
	Code    DomainErrorCode
	Message string
	Field   string // Optional: specific field that caused the error
}

// Error implements the error interface
func (e *DomainError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("[%s] %s: %s", e.Code, e.Field, e.Message)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// NewDomainError creates a new domain error
func NewDomainError(code DomainErrorCode, message string) *DomainError {
	return &DomainError{
		Code:    code,
		Message: message,
	}
}

// NewValidationError creates a validation error with field information
func NewValidationError(field, message string) *DomainError {
	return &DomainError{
		Code:    ErrCodeValidation,
		Message: message,
		Field:   field,
	}
}

// NewBusinessRuleError creates a business rule violation error
func NewBusinessRuleError(message string) *DomainError {
	return &DomainError{
		Code:    ErrCodeBusinessRule,
		Message: message,
	}
}

// NewInvalidStateError creates an invalid state error
func NewInvalidStateError(message string) *DomainError {
	return &DomainError{
		Code:    ErrCodeInvalidState,
		Message: message,
	}
}

// ==================== Legacy Support ====================
// ValidationError는 기존 코드와의 호환성을 위해 유지하지만
// 내부적으로 DomainError를 사용합니다

// ValidationError is a legacy type for backward compatibility
// Use NewValidationError instead
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

// ToDomainError converts ValidationError to DomainError
func (e *ValidationError) ToDomainError() *DomainError {
	return NewValidationError(e.Field, e.Message)
}
