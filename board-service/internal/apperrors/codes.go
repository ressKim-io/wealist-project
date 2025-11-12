package apperrors

// Error codes
const (
	ErrCodeUnauthorized              = "UNAUTHORIZED"
	ErrCodeForbidden                 = "FORBIDDEN"
	ErrCodeNotFound                  = "NOT_FOUND"
	ErrCodeBadRequest                = "BAD_REQUEST"
	ErrCodeInternalServer            = "INTERNAL_SERVER_ERROR"
	ErrCodeInvalidToken              = "INVALID_TOKEN"
	ErrCodeTokenExpired              = "TOKEN_EXPIRED"
	ErrCodeMissingToken              = "MISSING_TOKEN"
	ErrCodeConflict                  = "CONFLICT"
	ErrCodeValidation                = "VALIDATION_ERROR"
	ErrCodeWorkspaceValidationFailed = "WORKSPACE_VALIDATION_FAILED"
	ErrCodeWorkspaceAccessDenied     = "WORKSPACE_ACCESS_DENIED"
	ErrCodeWorkspaceNotFound         = "WORKSPACE_NOT_FOUND"
	ErrCodeNotImplemented            = "NOT_IMPLEMENTED"
)

// Predefined errors
var (
	ErrUnauthorized              = New(ErrCodeUnauthorized, "인증이 필요합니다", 401)
	ErrForbidden                 = New(ErrCodeForbidden, "권한이 없습니다", 403)
	ErrNotFound                  = New(ErrCodeNotFound, "리소스를 찾을 수 없습니다", 404)
	ErrBadRequest                = New(ErrCodeBadRequest, "잘못된 요청입니다", 400)
	ErrInternalServer            = New(ErrCodeInternalServer, "서버 에러가 발생했습니다", 500)
	ErrInvalidToken              = New(ErrCodeInvalidToken, "유효하지 않은 토큰입니다", 401)
	ErrTokenExpired              = New(ErrCodeTokenExpired, "토큰이 만료되었습니다", 401)
	ErrMissingToken              = New(ErrCodeMissingToken, "토큰이 필요합니다", 401)
	ErrConflict                  = New(ErrCodeConflict, "이미 존재하는 리소스입니다", 409)
	ErrValidation                = New(ErrCodeValidation, "입력값 검증에 실패했습니다", 400)
	ErrWorkspaceValidationFailed = New(ErrCodeWorkspaceValidationFailed, "워크스페이스 검증에 실패했습니다", 500)
	ErrWorkspaceAccessDenied     = New(ErrCodeWorkspaceAccessDenied, "워크스페이스 접근 권한이 없습니다", 403)
	ErrWorkspaceNotFound         = New(ErrCodeWorkspaceNotFound, "워크스페이스를 찾을 수 없습니다", 404)
)
