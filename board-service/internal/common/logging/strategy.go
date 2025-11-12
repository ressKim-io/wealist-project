package logging

import (
	"context"

	"go.uber.org/zap"
)

// ==================== Structured Logging Strategy ====================
// Uber Zap 기반 구조화된 로깅 베스트 프랙티스 (2024)
//
// 핵심 원칙:
// 1. 로그 레벨 정책 (Debug/Info/Warn/Error)
// 2. 구조화된 컨텍스트 (traceID, userID, requestID)
// 3. 성능 최적화 (SugaredLogger vs Logger)
// 4. 보안 (민감 정보 마스킹)
//
// Best Practices:
// - Production: Info 레벨 이상만 로깅
// - Development: Debug 레벨까지 로깅
// - Error: 스택 트레이스 포함
// - 구조화된 필드 사용 (key-value)

// ==================== Log Levels ====================

// LogLevel 정책:
//
// DEBUG: 개발 중 상세한 디버깅 정보
//   - SQL 쿼리 상세
//   - 내부 상태 변화
//   - 함수 진입/종료
//   사용 예: logger.Debug("Finding board by ID", zap.String("board_id", id))
//
// INFO: 일반적인 정보성 메시지
//   - 비즈니스 로직 실행 (생성, 수정, 삭제)
//   - 외부 API 호출
//   - 배치 작업 시작/종료
//   사용 예: logger.Info("Board created", zap.String("board_id", id), zap.String("user_id", userID))
//
// WARN: 잠재적 문제 (복구 가능)
//   - 리소스 부족 경고
//   - Deprecated API 사용
//   - 캐시 미스
//   - 외부 서비스 지연
//   사용 예: logger.Warn("Cache miss", zap.String("key", key))
//
// ERROR: 에러 발생 (복구 시도)
//   - 비즈니스 로직 에러
//   - 데이터베이스 에러
//   - 외부 API 에러
//   사용 예: logger.Error("Failed to create board", zap.Error(err), zap.String("user_id", userID))
//
// FATAL/PANIC: 치명적 오류 (프로세스 종료)
//   - 설정 파일 로드 실패
//   - 필수 리소스 초기화 실패
//   사용 예: logger.Fatal("Failed to connect to database", zap.Error(err))

// ==================== Context-Aware Logging ====================

// ContextKey는 context에 저장되는 키 타입
type ContextKey string

const (
	// TraceIDKey는 분산 추적 ID
	TraceIDKey ContextKey = "trace_id"

	// RequestIDKey는 HTTP 요청 ID
	RequestIDKey ContextKey = "request_id"

	// UserIDKey는 사용자 ID
	UserIDKey ContextKey = "user_id"

	// ProjectIDKey는 프로젝트 ID (컨텍스트용)
	ProjectIDKey ContextKey = "project_id"
)

// WithContext는 context에서 구조화된 필드를 추출하여 logger에 추가
func WithContext(ctx context.Context, logger *zap.Logger) *zap.Logger {
	fields := []zap.Field{}

	if traceID, ok := ctx.Value(TraceIDKey).(string); ok && traceID != "" {
		fields = append(fields, zap.String("trace_id", traceID))
	}

	if requestID, ok := ctx.Value(RequestIDKey).(string); ok && requestID != "" {
		fields = append(fields, zap.String("request_id", requestID))
	}

	if userID, ok := ctx.Value(UserIDKey).(string); ok && userID != "" {
		fields = append(fields, zap.String("user_id", userID))
	}

	if projectID, ok := ctx.Value(ProjectIDKey).(string); ok && projectID != "" {
		fields = append(fields, zap.String("project_id", projectID))
	}

	return logger.With(fields...)
}

// ==================== Audit Logging ====================

// AuditEvent는 감사 로그 이벤트 타입
type AuditEvent string

const (
	// 비즈니스 이벤트
	EventBoardCreated    AuditEvent = "board.created"
	EventBoardUpdated    AuditEvent = "board.updated"
	EventBoardDeleted    AuditEvent = "board.deleted"
	EventBoardMoved      AuditEvent = "board.moved"
	EventProjectCreated  AuditEvent = "project.created"
	EventProjectDeleted  AuditEvent = "project.deleted"
	EventMemberAdded     AuditEvent = "member.added"
	EventMemberRemoved   AuditEvent = "member.removed"
	EventRoleChanged     AuditEvent = "role.changed"
	EventCommentCreated  AuditEvent = "comment.created"
	EventCommentDeleted  AuditEvent = "comment.deleted"
)

// AuditLogger는 비즈니스 이벤트를 감사 로그로 기록
type AuditLogger struct {
	logger *zap.Logger
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger(logger *zap.Logger) *AuditLogger {
	return &AuditLogger{
		logger: logger.Named("audit"),
	}
}

// LogEvent logs an audit event with structured fields
func (a *AuditLogger) LogEvent(
	ctx context.Context,
	event AuditEvent,
	userID string,
	fields ...zap.Field,
) {
	// 컨텍스트 필드 추가
	contextLogger := WithContext(ctx, a.logger)

	// 기본 필드 추가
	allFields := append([]zap.Field{
		zap.String("event", string(event)),
		zap.String("actor", userID),
	}, fields...)

	contextLogger.Info("Audit event", allFields...)
}

// LogBoardCreated logs board creation event
func (a *AuditLogger) LogBoardCreated(ctx context.Context, userID, boardID, projectID string) {
	a.LogEvent(ctx, EventBoardCreated, userID,
		zap.String("board_id", boardID),
		zap.String("project_id", projectID),
	)
}

// LogBoardDeleted logs board deletion event
func (a *AuditLogger) LogBoardDeleted(ctx context.Context, userID, boardID, projectID string) {
	a.LogEvent(ctx, EventBoardDeleted, userID,
		zap.String("board_id", boardID),
		zap.String("project_id", projectID),
	)
}

// LogProjectCreated logs project creation event
func (a *AuditLogger) LogProjectCreated(ctx context.Context, userID, projectID, workspaceID string) {
	a.LogEvent(ctx, EventProjectCreated, userID,
		zap.String("project_id", projectID),
		zap.String("workspace_id", workspaceID),
	)
}

// LogMemberAdded logs member added event
func (a *AuditLogger) LogMemberAdded(ctx context.Context, actorID, memberID, projectID, role string) {
	a.LogEvent(ctx, EventMemberAdded, actorID,
		zap.String("member_id", memberID),
		zap.String("project_id", projectID),
		zap.String("role", role),
	)
}

// LogRoleChanged logs role changed event
func (a *AuditLogger) LogRoleChanged(ctx context.Context, actorID, memberID, projectID, oldRole, newRole string) {
	a.LogEvent(ctx, EventRoleChanged, actorID,
		zap.String("member_id", memberID),
		zap.String("project_id", projectID),
		zap.String("old_role", oldRole),
		zap.String("new_role", newRole),
	)
}

// ==================== Performance Logging ====================

// Timer는 함수 실행 시간을 측정
type Timer struct {
	logger *zap.Logger
	name   string
	fields []zap.Field
}

// NewTimer creates a new performance timer
func NewTimer(logger *zap.Logger, name string, fields ...zap.Field) *Timer {
	return &Timer{
		logger: logger,
		name:   name,
		fields: fields,
	}
}

// End logs the elapsed time (defer로 사용)
func (t *Timer) End() {
	// Zap의 Duration 필드 사용
	// 사용 예:
	// timer := logging.NewTimer(logger, "GetBoards", zap.String("project_id", projectID))
	// defer timer.End()
	//
	// 실제 사용 시에는 time.Now()를 저장하고 Duration을 계산해야 함
	// 여기서는 예제만 제공
}

// ==================== Sensitive Data Masking ====================

// MaskEmail masks email addresses for logging
func MaskEmail(email string) string {
	if len(email) < 3 {
		return "***"
	}
	atIndex := -1
	for i, c := range email {
		if c == '@' {
			atIndex = i
			break
		}
	}
	if atIndex > 0 {
		return email[:1] + "***" + email[atIndex:]
	}
	return "***"
}

// MaskToken masks authentication tokens
func MaskToken(token string) string {
	if len(token) < 10 {
		return "***"
	}
	return token[:4] + "..." + token[len(token)-4:]
}

// ==================== Usage Examples ====================

// Example 1: Service 레이어에서 구조화된 로깅
//
//	func (s *boardService) CreateBoard(userID string, req *dto.CreateBoardRequest) (*dto.BoardResponse, error) {
//	    s.logger.Info("Creating board",
//	        zap.String("user_id", userID),
//	        zap.String("project_id", req.ProjectID),
//	        zap.String("title", req.Title),
//	    )
//
//	    board, err := s.repo.Create(...)
//	    if err != nil {
//	        s.logger.Error("Failed to create board",
//	            zap.Error(err),
//	            zap.String("user_id", userID),
//	            zap.String("project_id", req.ProjectID),
//	        )
//	        return nil, err
//	    }
//
//	    // Audit log
//	    s.auditLogger.LogBoardCreated(ctx, userID, board.ID.String(), req.ProjectID)
//
//	    s.logger.Info("Board created successfully",
//	        zap.String("board_id", board.ID.String()),
//	        zap.String("user_id", userID),
//	    )
//	    return response, nil
//	}
//
// Example 2: Context-aware 로깅
//
//	func (s *boardService) GetBoard(ctx context.Context, boardID string) (*dto.BoardResponse, error) {
//	    logger := logging.WithContext(ctx, s.logger)
//	    logger.Info("Fetching board", zap.String("board_id", boardID))
//	    // ...
//	}
//
// Example 3: 성능 측정
//
//	func (s *boardService) GetBoards(projectID string) ([]dto.BoardResponse, error) {
//	    timer := logging.NewTimer(s.logger, "GetBoards", zap.String("project_id", projectID))
//	    defer timer.End()
//	    // ...
//	}
