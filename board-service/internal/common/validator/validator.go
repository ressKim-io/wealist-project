package validator

import (
	"board-service/internal/apperrors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// ValidateRequired checks if a string field is not empty
func ValidateRequired(value, fieldName string) error {
	if strings.TrimSpace(value) == "" {
		return apperrors.New(
			apperrors.ErrCodeBadRequest,
			fmt.Sprintf("%s는 필수 입력 항목입니다", fieldName),
			400,
		)
	}
	return nil
}

// ValidateMaxLength checks if a string does not exceed max length
func ValidateMaxLength(value, fieldName string, maxLength int) error {
	if len(value) > maxLength {
		return apperrors.New(
			apperrors.ErrCodeBadRequest,
			fmt.Sprintf("%s는 최대 %d자까지 입력 가능합니다", fieldName, maxLength),
			400,
		)
	}
	return nil
}

// ValidateMinLength checks if a string meets minimum length
func ValidateMinLength(value, fieldName string, minLength int) error {
	if len(value) < minLength {
		return apperrors.New(
			apperrors.ErrCodeBadRequest,
			fmt.Sprintf("%s는 최소 %d자 이상이어야 합니다", fieldName, minLength),
			400,
		)
	}
	return nil
}

// ValidateStringLength checks if a string length is within range
func ValidateStringLength(value, fieldName string, minLength, maxLength int) error {
	length := len(value)
	if length < minLength || length > maxLength {
		return apperrors.New(
			apperrors.ErrCodeBadRequest,
			fmt.Sprintf("%s는 %d~%d자 사이여야 합니다", fieldName, minLength, maxLength),
			400,
		)
	}
	return nil
}

// ValidateEmail checks if a string is a valid email format
func ValidateEmail(email string) error {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return apperrors.New(
			apperrors.ErrCodeBadRequest,
			"유효하지 않은 이메일 형식입니다",
			400,
		)
	}
	return nil
}

// ValidateURL checks if a string is a valid URL
func ValidateURL(url string) error {
	urlRegex := regexp.MustCompile(`^https?://[^\s/$.?#].[^\s]*$`)
	if !urlRegex.MatchString(url) {
		return apperrors.New(
			apperrors.ErrCodeBadRequest,
			"유효하지 않은 URL 형식입니다",
			400,
		)
	}
	return nil
}

// ValidateEnum checks if a value is in the allowed list
func ValidateEnum(value, fieldName string, allowedValues []string) error {
	for _, allowed := range allowedValues {
		if value == allowed {
			return nil
		}
	}
	return apperrors.New(
		apperrors.ErrCodeBadRequest,
		fmt.Sprintf("%s는 다음 값 중 하나여야 합니다: %s", fieldName, strings.Join(allowedValues, ", ")),
		400,
	)
}

// ValidateRange checks if a number is within range (inclusive)
func ValidateRange(value int, fieldName string, min, max int) error {
	if value < min || value > max {
		return apperrors.New(
			apperrors.ErrCodeBadRequest,
			fmt.Sprintf("%s는 %d~%d 사이의 값이어야 합니다", fieldName, min, max),
			400,
		)
	}
	return nil
}

// ValidatePositive checks if a number is positive
func ValidatePositive(value int, fieldName string) error {
	if value <= 0 {
		return apperrors.New(
			apperrors.ErrCodeBadRequest,
			fmt.Sprintf("%s는 양수여야 합니다", fieldName),
			400,
		)
	}
	return nil
}

// ValidateNonNegative checks if a number is non-negative
func ValidateNonNegative(value int, fieldName string) error {
	if value < 0 {
		return apperrors.New(
			apperrors.ErrCodeBadRequest,
			fmt.Sprintf("%s는 0 이상이어야 합니다", fieldName),
			400,
		)
	}
	return nil
}

// ValidateDateFormat checks if a date string is in RFC3339 format
func ValidateDateFormat(dateStr, fieldName string) (*time.Time, error) {
	parsed, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		return nil, apperrors.Wrap(
			err,
			apperrors.ErrCodeBadRequest,
			fmt.Sprintf("%s의 날짜 형식이 올바르지 않습니다 (ISO 8601 required)", fieldName),
			400,
		)
	}
	return &parsed, nil
}

// ValidateFutureDate checks if a date is in the future
func ValidateFutureDate(date time.Time, fieldName string) error {
	if date.Before(time.Now()) {
		return apperrors.New(
			apperrors.ErrCodeBadRequest,
			fmt.Sprintf("%s는 미래 날짜여야 합니다", fieldName),
			400,
		)
	}
	return nil
}

// ValidatePastDate checks if a date is in the past
func ValidatePastDate(date time.Time, fieldName string) error {
	if date.After(time.Now()) {
		return apperrors.New(
			apperrors.ErrCodeBadRequest,
			fmt.Sprintf("%s는 과거 날짜여야 합니다", fieldName),
			400,
		)
	}
	return nil
}

// ValidateArrayNotEmpty checks if an array is not empty
func ValidateArrayNotEmpty[T any](arr []T, fieldName string) error {
	if len(arr) == 0 {
		return apperrors.New(
			apperrors.ErrCodeBadRequest,
			fmt.Sprintf("%s는 비어있을 수 없습니다", fieldName),
			400,
		)
	}
	return nil
}

// ValidateArrayMaxLength checks if an array does not exceed max length
func ValidateArrayMaxLength[T any](arr []T, fieldName string, maxLength int) error {
	if len(arr) > maxLength {
		return apperrors.New(
			apperrors.ErrCodeBadRequest,
			fmt.Sprintf("%s는 최대 %d개까지 가능합니다", fieldName, maxLength),
			400,
		)
	}
	return nil
}

// ValidateArrayUnique checks if all elements in array are unique
func ValidateArrayUnique[T comparable](arr []T, fieldName string) error {
	seen := make(map[T]bool)
	for _, item := range arr {
		if seen[item] {
			return apperrors.New(
				apperrors.ErrCodeBadRequest,
				fmt.Sprintf("%s에 중복된 값이 있습니다", fieldName),
				400,
			)
		}
		seen[item] = true
	}
	return nil
}

// Validator is a builder pattern for chaining validations
type Validator struct {
	errors []error
}

// New creates a new Validator
func New() *Validator {
	return &Validator{
		errors: make([]error, 0),
	}
}

// Check adds an error if the condition is false
func (v *Validator) Check(condition bool, err error) *Validator {
	if !condition {
		v.errors = append(v.errors, err)
	}
	return v
}

// Validate runs a validation function and collects errors
func (v *Validator) Validate(fn func() error) *Validator {
	if err := fn(); err != nil {
		v.errors = append(v.errors, err)
	}
	return v
}

// HasErrors returns true if any validation failed
func (v *Validator) HasErrors() bool {
	return len(v.errors) > 0
}

// Errors returns all collected errors
func (v *Validator) Errors() []error {
	return v.errors
}

// FirstError returns the first error, or nil if no errors
func (v *Validator) FirstError() error {
	if len(v.errors) == 0 {
		return nil
	}
	return v.errors[0]
}
