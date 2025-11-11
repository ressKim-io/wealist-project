package parser

import (
	"board-service/internal/apperrors"
	"fmt"

	"github.com/google/uuid"
)

// ParseUUID parses a UUID string and returns an AppError on failure
func ParseUUID(idStr, fieldName string) (uuid.UUID, error) {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return uuid.Nil, apperrors.Wrap(
			err,
			apperrors.ErrCodeBadRequest,
			fmt.Sprintf("잘못된 %s ID", fieldName),
			400,
		)
	}
	return id, nil
}

// ParseUserID is a convenience wrapper for ParseUUID with "사용자" as field name
func ParseUserID(userID string) (uuid.UUID, error) {
	return ParseUUID(userID, "사용자")
}

// ParseProjectID is a convenience wrapper for ParseUUID with "프로젝트" as field name
func ParseProjectID(projectID string) (uuid.UUID, error) {
	return ParseUUID(projectID, "프로젝트")
}

// ParseBoardID is a convenience wrapper for ParseUUID with "보드" as field name
func ParseBoardID(boardID string) (uuid.UUID, error) {
	return ParseUUID(boardID, "보드")
}

// ParseFieldID is a convenience wrapper for ParseUUID with "필드" as field name
func ParseFieldID(fieldID string) (uuid.UUID, error) {
	return ParseUUID(fieldID, "필드")
}

// ParseOptionalUUID parses an optional UUID (nullable pointer)
// Returns nil UUID pointer if input is nil or empty
func ParseOptionalUUID(idStr *string, fieldName string) (*uuid.UUID, error) {
	if idStr == nil || *idStr == "" {
		return nil, nil
	}

	id, err := uuid.Parse(*idStr)
	if err != nil {
		return nil, apperrors.Wrap(
			err,
			apperrors.ErrCodeBadRequest,
			fmt.Sprintf("잘못된 %s ID", fieldName),
			400,
		)
	}
	return &id, nil
}

// MustParseUUID parses a UUID and panics on error (use only in tests or initialization)
func MustParseUUID(idStr string) uuid.UUID {
	id, err := uuid.Parse(idStr)
	if err != nil {
		panic(fmt.Sprintf("invalid UUID: %s", idStr))
	}
	return id
}

// UUIDsToStrings converts a slice of UUIDs to strings
func UUIDsToStrings(ids []uuid.UUID) []string {
	strs := make([]string, len(ids))
	for i, id := range ids {
		strs[i] = id.String()
	}
	return strs
}

// StringsToUUIDs converts a slice of UUID strings to UUIDs
// Returns error on first invalid UUID
func StringsToUUIDs(strs []string) ([]uuid.UUID, error) {
	ids := make([]uuid.UUID, len(strs))
	for i, str := range strs {
		id, err := uuid.Parse(str)
		if err != nil {
			return nil, apperrors.Wrap(
				err,
				apperrors.ErrCodeBadRequest,
				fmt.Sprintf("잘못된 UUID: %s", str),
				400,
			)
		}
		ids[i] = id
	}
	return ids, nil
}

// ValidateUUID checks if a string is a valid UUID without parsing
func ValidateUUID(idStr string) bool {
	_, err := uuid.Parse(idStr)
	return err == nil
}

// IsNilOrEmpty checks if a UUID is nil or empty (all zeros)
func IsNilOrEmpty(id uuid.UUID) bool {
	return id == uuid.Nil
}

// UUIDPtrToString converts *uuid.UUID to *string
func UUIDPtrToString(id *uuid.UUID) *string {
	if id == nil {
		return nil
	}
	str := id.String()
	return &str
}

// StringPtrToUUID converts *string to *uuid.UUID
func StringPtrToUUID(str *string) (*uuid.UUID, error) {
	if str == nil || *str == "" {
		return nil, nil
	}
	id, err := uuid.Parse(*str)
	if err != nil {
		return nil, err
	}
	return &id, nil
}
