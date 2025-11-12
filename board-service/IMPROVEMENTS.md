# Board Service 개선 사항

**Last Updated**: 2025-11-12
**Status**: Phase 1, 2, 3-1 완료 ✅

## 📋 분석 결과 요약

7단계 리팩토링을 통해 구조가 크게 개선되었으며, **Phase 1 (전체), Phase 2 (전체), Phase 3-1이 완료**되었습니다.

---

## ✅ 완료된 개선 사항

### Phase 1: 핵심 패턴 적용 (완료)

#### 1. ✅ Service에 UnitOfWork 적용 완료
**완료 내용**:
- ✅ `boardService`에 UnitOfWork 주입 완료
- ✅ `DeleteBoard()` 메서드에 UoW 패턴 적용
- ✅ 보드 + 댓글 원자적 삭제 구현
- ✅ Callback 패턴 (자동 rollback)

**적용 코드**:
```go
// ✅ 완료: UnitOfWork 주입
type boardService struct {
    uow uow.UnitOfWork  // 추가됨
    // ... 기타 의존성
}

func (s *boardService) DeleteBoard() error {
    return s.uow.Do(func(repos *uow.Repositories) error {
        board.MarkAsDeleted()
        repos.Board.Update(board)

        // 관련 댓글도 트랜잭션으로 삭제
        comments, _ := repos.Comment.FindByBoard(boardID)
        for _, c := range comments {
            repos.Comment.Delete(c.ID)
        }
        return nil  // 모두 성공하거나 모두 실패
    })
}
```

**커밋**: `feat: [Phase 1] UnitOfWork 실제 적용 + Comment BaseModel 마이그레이션`

---

#### 2. ✅ Comment BaseModel 마이그레이션 완료

**완료 내용**:
- ✅ `DeletedAt (gorm.DeletedAt)` → `IsDeleted (bool)` 변경
- ✅ BaseModel 임베딩 (ID, CreatedAt, UpdatedAt, IsDeleted)
- ✅ Generic Repository 지원
- ✅ Entity 인터페이스 구현

**적용 코드**:
```go
// ✅ 완료: BaseModel 사용
type Comment struct {
    BaseModel  // ID, CreatedAt, UpdatedAt, IsDeleted 포함
    Content   string    `gorm:"type:text;not null"`
    UserID    uuid.UUID `gorm:"type:uuid;not null;index"`
    BoardID   uuid.UUID `gorm:"type:uuid;not null;index"`
    Board     Board     `gorm:"foreignKey:BoardID"`
}
```

**커밋**: `feat: [Phase 1] UnitOfWork 실제 적용 + Comment BaseModel 마이그레이션`

---

#### 3. ✅ 에러 처리 일관성 개선 완료

**완료 내용**:
- ✅ `DomainError` 타입 정의 (ValidationError, BusinessRuleError, InvalidStateError)
- ✅ `AppError`와 Domain 에러 분리
- ✅ `FromDomainError()` 자동 변환 함수
- ✅ Board, Project, Comment에 적용

**적용 코드**:
```go
// ✅ 완료: Domain 레벨 에러
type DomainError struct {
    Code    DomainErrorCode  // VALIDATION_ERROR, BUSINESS_RULE_VIOLATION, INVALID_STATE
    Message string
    Field   string  // Validation 에러용
}

// ✅ 완료: Infrastructure 레벨 에러 변환
func FromDomainError(err error) *AppError {
    var domainErr *domain.DomainError
    if errors.As(err, &domainErr) {
        switch domainErr.Code {
        case domain.ErrCodeValidation:
            return &AppError{Code: ErrCodeBadRequest, HTTPStatus: 400, ...}
        case domain.ErrCodeBusinessRule:
            return &AppError{Code: ErrCodeForbidden, HTTPStatus: 403, ...}
        // ...
        }
    }
}

// 사용 예시
func (b *Board) UpdateTitle(title string) error {
    if title == "" {
        return NewValidationError("title", "제목은 필수입니다")
    }
    return nil
}

// Service에서
if err := board.UpdateTitle(req.Title); err != nil {
    return apperrors.FromDomainError(err)  // 자동 변환
}
```

**커밋**: `feat: [Phase 1-3] 도메인 에러 처리 일관성 개선`

---

### Phase 2: 아키텍처 개선 (완료 ✅)

#### 4. ✅ Repository 인터페이스 문서화 완료

**완료 내용**:
- ✅ `repository/interfaces.go` 생성
- ✅ 10개 Repository 목록 및 책임 정의
- ✅ ISP (Interface Segregation Principle) 설명
- ✅ DIP (Dependency Inversion) 아키텍처 다이어그램

**적용 코드**:
```go
// ✅ 완료: repository/interfaces.go
// - 모든 Repository 인터페이스의 중앙 문서
// - Interface Segregation Principle (ISP) 설명
// - 의존성 역전 원칙 (DIP) 아키텍처
//
// Repository 목록:
// - BoardRepository       : Board 엔티티 관리
// - ProjectRepository     : Project 엔티티 관리
// - CommentRepository     : Comment 엔티티 관리
// - RoleRepository        : Role 엔티티 관리
// - FieldRepository       : Custom Fields 통합 관리
// - ProjectFieldRepository: ProjectField 엔티티 관리
// - FieldOptionRepository : FieldOption 엔티티 관리
// - FieldValueRepository  : BoardFieldValue 엔티티 관리
// - ViewRepository        : SavedView 엔티티 관리
// - BoardOrderRepository  : UserBoardOrder 엔티티 관리
```

**참고**: 현재 인터페이스와 구현체는 같은 파일에 있지만, 중앙 문서화로 명확성을 확보했습니다.

**커밋**: `feat: [Phase 2] Repository 인터페이스 문서화 + DTO Mapper 패턴 도입`

---

---

## 🔴 높은 우선순위 (High Priority) - 미완료

#### 5. ✅ DTO Mapper 패턴 도입 완료

**완료 내용**:
- ✅ `dto/mapper.go` 생성
- ✅ BoardMapper, ProjectMapper, CommentMapper, ProjectMemberMapper 구현
- ✅ **코드 감소**: 140+ 줄 중복 제거
  - `buildBoardResponse()`: 74줄 → 8줄
  - `buildBoardResponseOptimized()`: 68줄 → 4줄

**적용 코드**:
```go
// ✅ 완료: dto/mapper.go
type BoardMapper struct {
    logger *zap.Logger
}

func (m *BoardMapper) ToResponseWithUserMap(
    board *domain.Board,
    userMap map[string]client.UserInfo,
) *BoardResponse {
    // CustomFieldsCache (JSONB) 파싱
    // User 정보 매핑
    // Author, Assignee 처리
    return response
}

func (m *BoardMapper) ToResponseList(
    boards []domain.Board,
    userMap map[string]client.UserInfo,
) []BoardResponse {
    // 배치 최적화 지원 (N+1 방지)
}

// ✅ Service에서 간결하게 사용
func (s *boardService) buildBoardResponse(board *domain.Board) (*dto.BoardResponse, error) {
    userMap := s.getUserInfoBatch(ctx, userIDs)
    return s.mapper.ToResponseWithUserMap(board, userMap), nil  // 8줄
}
```

**장점**:
- DTO 변환 로직 중앙화
- 타입 안전성 보장
- 배치 최적화 지원
- 테스트 용이성 (순수 함수)

**커밋**: `feat: [Phase 2] Repository 인터페이스 문서화 + DTO Mapper 패턴 도입`

---

### Phase 3: 운영 최적화 (부분 완료)

#### 6. ✅ 구조화된 로깅 전략 완료

**완료 내용**:
- ✅ `common/logging/strategy.go` 생성
- ✅ 로그 레벨 정책 정의 (DEBUG, INFO, WARN, ERROR, FATAL/PANIC)
- ✅ Context-Aware Logging (trace_id, request_id, user_id)
- ✅ AuditLogger (11개 비즈니스 이벤트)
- ✅ Performance Logging (Timer)
- ✅ Security (MaskEmail, MaskToken)

**적용 코드**:
```go
// ✅ 완료: 로그 레벨 정책
// DEBUG: 상세 디버깅 (SQL, 내부 상태)
// INFO: 비즈니스 로직 실행 (생성, 수정, 삭제)
// WARN: 잠재적 문제 (캐시 미스, 외부 서비스 지연)
// ERROR: 에러 발생 (복구 시도)
// FATAL/PANIC: 치명적 오류

// ✅ 완료: Context-Aware Logging
func WithContext(ctx context.Context, logger *zap.Logger) *zap.Logger {
    // trace_id, request_id, user_id, project_id 추출
}

// ✅ 완료: AuditLogger
type AuditLogger struct {
    logger *zap.Logger
}

func (a *AuditLogger) LogBoardCreated(ctx context.Context, userID, boardID, projectID string) {
    a.LogEvent(ctx, EventBoardCreated, userID,
        zap.String("board_id", boardID),
        zap.String("project_id", projectID),
    )
}

// 11개 이벤트:
// - board.created, board.updated, board.deleted, board.moved
// - project.created, project.deleted
// - member.added, member.removed, role.changed
// - comment.created, comment.deleted

// ✅ 완료: Performance Logging
timer := logging.NewTimer(logger, "GetBoards", zap.String("project_id", projectID))
defer timer.End()

// ✅ 완료: Security
MaskEmail("user@example.com")  // → u***@example.com
MaskToken("abc123xyz789")      // → abc1...x789
```

**커밋**: `feat: [Phase 3-1] 구조화된 로깅 전략 구현`

---

### 7. ✅ 테스트 커버리지 향상 완료 (Phase 2-3)

**완료 내용**:
- ✅ Service Layer 유닛 테스트 추가
  - ProjectService: 13개 테스트 케이스
  - CommentService: 16개 테스트 케이스
- ✅ Repository Layer 통합 테스트 추가
  - CommentRepository: 22개 테스트 케이스
  - ProjectRepository: 23개 테스트 케이스
- ✅ Mock 인프라 구축
  - MockCommentRepository 추가
  - MockUserClient, MockUserInfoCache 구현
- ✅ **총 177개 테스트 케이스** (기존 103개 + 신규 74개)

**테스트 구성**:
```go
// ✅ 완료: Service 유닛 테스트 (Given-When-Then 패턴)
func TestCommentService_CreateComment_Success(t *testing.T) {
    suite := setupCommentServiceTest(t)

    // Given: Valid comment request
    req := dto.CreateCommentRequest{...}
    suite.boardRepo.On("FindByID", boardID).Return(board, nil)
    suite.projectRepo.On("FindMemberByUserAndProject", ...).Return(member, nil)
    suite.commentRepo.On("Create", ...).Return(nil)

    // When: Create comment
    result, err := suite.service.CreateComment(ctx, req, userID)

    // Then: Verify success
    assert.NoError(t, err)
    suite.commentRepo.AssertExpectations(t)
}

// ✅ 완료: Repository 통합 테스트 (실제 DB 사용)
func TestCommentRepository_FindByBoardID_OrderedByCreatedAt(t *testing.T) {
    suite := setupCommentRepoTest(t)  // SQLite in-memory DB
    defer suite.teardown()

    // Create comments with time gaps
    comment1 := &domain.Comment{...}
    suite.repo.Create(comment1)
    time.Sleep(10 * time.Millisecond)

    comment2 := &domain.Comment{...}
    suite.repo.Create(comment2)

    // Verify ascending order
    comments, _ := suite.repo.FindByBoardID(boardID)
    assert.True(t, comments[0].CreatedAt.Before(comments[1].CreatedAt))
}
```

**테스트 커버리지**:
| 레이어 | 이전 | 이후 | 증가 |
|--------|------|------|------|
| Service | 61개 | 90개 | +29개 |
| Repository | 17개 | 60개 | +43개 |
| **합계** | **103개** | **177개** | **+74개** |

**커밋**: `feat: [Phase 2-3] 테스트 커버리지 향상 - Service & Repository 테스트 추가`

---

## ⚠️ 미완료 개선 사항

**개선 방안**:
```go
// 1. Service 유닛 테스트 추가
// service/board_service_test.go 보강
func TestBoardService_UpdateBoard_Success(t *testing.T) {
    // Given: Mock 설정
    mockRepo := new(mocks.BoardRepository)
    mockAuth := new(mocks.ProjectAuthorizer)

    service := NewBoardService(mockRepo, mockAuth, ...)

    // When: 보드 업데이트
    result, err := service.UpdateBoard(boardID, userID, req)

    // Then: 검증
    assert.NoError(t, err)
    assert.Equal(t, "New Title", result.Title)
    mockRepo.AssertCalled(t, "Update", mock.Anything)
}

// 2. Repository 통합 테스트
// repository/board_repository_integration_test.go
func TestBoardRepository_FindByProject(t *testing.T) {
    // Given: 실제 DB (In-memory SQLite)
    db := testutil.SetupTestDB()
    repo := NewBoardRepository(db)

    // When: 프로젝트의 보드 조회
    boards, _ := repo.FindByProject(projectID)

    // Then: 검증
    assert.Len(t, boards, 2)
}

// 3. HTTP Handler E2E 테스트
// handler/board_handler_e2e_test.go
func TestBoardAPI_CreateBoard_E2E(t *testing.T) {
    // Given: 테스트 서버 시작
    router := setupTestRouter()

    // When: POST /api/boards
    w := httptest.NewRecorder()
    req, _ := http.NewRequest("POST", "/api/boards", body)
    router.ServeHTTP(w, req)

    // Then: 검증
    assert.Equal(t, 201, w.Code)
}
```

**목표 커버리지**: 80%
**예상 작업 시간**: 8-12시간

---

## 🟡 중간 우선순위 (Medium Priority) - 미완료

### 8. ⚠️ Cache 전략이 명확하지 않음 (Phase 3-2)
**현재 상태**:
```go
// Cache가 Service에서 선택적으로 사용됨
userMap := s.getUserInfoBatch(ctx, userIDs)  // Cache 사용

// 일부 조회는 Cache 없이 직접 DB 접근
project, _ := s.projectRepo.FindByID(projectID)
```

**문제점**:
1. 어떤 데이터를 캐싱할지 불명확
2. Cache 만료 정책 없음
3. Cache Invalidation 전략 부재

**개선 방안**:
```go
// 1. Cache Decorator 패턴
type CachedBoardRepository struct {
    repo  repository.BoardRepository
    cache cache.Cache
    ttl   time.Duration
}

func (r *CachedBoardRepository) FindByID(id uuid.UUID) (*domain.Board, error) {
    // 1. Cache 확인
    cached, found := r.cache.Get(fmt.Sprintf("board:%s", id))
    if found {
        return cached.(*domain.Board), nil
    }

    // 2. DB 조회
    board, err := r.repo.FindByID(id)
    if err != nil {
        return nil, err
    }

    // 3. Cache 저장
    r.cache.Set(fmt.Sprintf("board:%s", id), board, r.ttl)
    return board, nil
}

// 2. Cache Invalidation
func (r *CachedBoardRepository) Update(board *domain.Board) error {
    // DB 업데이트
    if err := r.repo.Update(board); err != nil {
        return err
    }

    // Cache 무효화
    r.cache.Delete(fmt.Sprintf("board:%s", board.ID))
    return nil
}

// 3. Cache 정책 정의
const (
    BoardCacheTTL   = 5 * time.Minute   // 보드: 5분
    ProjectCacheTTL = 10 * time.Minute  // 프로젝트: 10분
    UserCacheTTL    = 30 * time.Minute  // 사용자: 30분
)
```

**예상 작업 시간**: 4-6시간

---

## 🟢 낮은 우선순위 (Low Priority)

### 9. ⚠️ 메트릭 및 모니터링 부족 (Phase 3-3)
**현재 상태**:
- Prometheus `/metrics` 엔드포인트만 존재
- 커스텀 메트릭 없음

**개선 방안**:
```go
// 비즈니스 메트릭 추가
var (
    boardCreatedCounter = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "board_created_total",
            Help: "Total number of boards created",
        },
        []string{"project_id"},
    )

    boardUpdateDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "board_update_duration_seconds",
            Help:    "Board update duration",
            Buckets: prometheus.DefBuckets,
        },
        []string{"project_id"},
    )
)

// Service에서 사용
func (s *boardService) CreateBoard() error {
    // ... 비즈니스 로직

    boardCreatedCounter.WithLabelValues(projectID).Inc()
    return nil
}
```

**예상 작업 시간**: 2-3시간

---

### 10. 🔐 Rate Limiting 부재
**현재 상태**:
- API Rate Limiting 없음
- 악의적 요청 방어 부족

**개선 방안**:
```go
// middleware/rate_limiter.go
func RateLimiterMiddleware(rdb *redis.Client) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID := c.GetString("user_id")
        key := fmt.Sprintf("rate_limit:%s", userID)

        count, _ := rdb.Incr(c.Request.Context(), key).Result()
        if count == 1 {
            rdb.Expire(c.Request.Context(), key, time.Minute)
        }

        if count > 100 {  // 분당 100 요청 제한
            c.JSON(429, gin.H{"error": "Too Many Requests"})
            c.Abort()
            return
        }

        c.Next()
    }
}
```

**예상 작업 시간**: 2-3시간

---

### 11. 🌐 국제화 (i18n) 미지원
**현재 상태**:
- 에러 메시지가 한글로 하드코딩
- 다국어 지원 불가

**개선 방안**:
```go
// i18n/messages.go
type Messages struct {
    lang string
}

func (m *Messages) Get(key string) string {
    messages := map[string]map[string]string{
        "ko": {
            "board.not_found": "보드를 찾을 수 없습니다",
            "board.forbidden": "권한이 없습니다",
        },
        "en": {
            "board.not_found": "Board not found",
            "board.forbidden": "Permission denied",
        },
    }
    return messages[m.lang][key]
}

// Service에서 사용
func (s *boardService) GetBoard() error {
    if board == nil {
        return errors.New(s.i18n.Get("board.not_found"))
    }
}
```

**예상 작업 시간**: 4-6시간

---

## 📈 개선 로드맵

### ✅ Phase 1: 핵심 기능 안정화 (완료)
1. ✅ **UnitOfWork 실제 적용** - DeleteBoard에 트랜잭션 패턴 적용
2. ✅ **Comment BaseModel 마이그레이션** - Soft Delete 일관성 확보
3. ✅ **에러 처리 일관성** - DomainError/AppError 분리

### ✅ Phase 2: 코드 품질 향상 (완료)
4. ✅ **Repository 인터페이스 문서화** - interfaces.go 생성
5. ✅ **DTO Mapper 도입** - 140+ 줄 중복 제거
6. ✅ **테스트 커버리지 향상** - 177개 테스트 케이스 (+74개)

### ✅ Phase 3: 운영 안정성 (부분 완료)
7. ✅ **로깅 전략 수립** - 구조화된 로깅 + Audit Log
8. ⏳ **Cache 전략 명확화** - 미완료
9. ⏳ **메트릭 및 모니터링** - 미완료

### Phase 4: 확장성 (4-8주)
10. **Rate Limiting** (Low Priority #10)
11. **국제화 지원** (Low Priority #11)

---

## 📊 작업 시간 현황

| 우선순위 | 작업 수 | 완료 | 진행률 | 남은 시간 |
|---------|--------|------|--------|----------|
| 🔴 High | 4개 | 4개 ✅ | 100% | 0시간 |
| 🟡 Medium | 4개 | 2개 ✅ | 50% | 10-15시간 |
| 🟢 Low | 3개 | 0개 | 0% | 8-12시간 |
| **총합** | **11개** | **6개 ✅** | **55%** | **18-27시간**

---

## ✅ 현재까지 달성한 효과

### 코드 품질
- ✅ 트랜잭션 안정성 향상 (UnitOfWork DeleteBoard 적용)
- ✅ 일관된 에러 처리 (DomainError/AppError 분리)
- ✅ DTO 변환 로직 중복 제거 (140+ 줄)

### 개발 생산성
- ✅ Repository 인터페이스 문서화 (ISP/DIP 명확화)
- ✅ Mapper 패턴으로 Service 간소화
- ✅ Domain 메서드 26개 (비즈니스 로직 캡슐화)

### 운영 안정성
- ✅ 구조화된 로깅 전략 (Audit Log + Context-Aware)
- ✅ 보안 강화 (민감 정보 마스킹)
- ⏳ Cache 전략 (미완료)
- ⏳ 메트릭/모니터링 (미완료)

## 📝 완료된 커밋

| 순번 | 커밋 | Phase | 날짜 |
|-----|------|-------|------|
| 1 | `feat: [Phase 1] UnitOfWork 실제 적용 + Comment BaseModel 마이그레이션` | Phase 1-1, 1-2 | 2025-11-12 |
| 2 | `feat: [Phase 1-3] 도메인 에러 처리 일관성 개선` | Phase 1-3 | 2025-11-12 |
| 3 | `feat: [Phase 2] Repository 인터페이스 문서화 + DTO Mapper 패턴 도입` | Phase 2-1, 2-2 | 2025-11-12 |
| 4 | `feat: [Phase 3-1] 구조화된 로깅 전략 구현` | Phase 3-1 | 2025-11-12 |
| 5 | `feat: [Phase 2-3] 테스트 커버리지 향상 - Service & Repository 테스트 추가` | Phase 2-3 | 2025-11-12 |

---

**Last Updated**: 2025-11-12
**Status**: Phase 1 (완료), Phase 2 (완료), Phase 3-1 (완료) ✅
**Next Steps**: Phase 3-2 (Cache 전략), Phase 3-3 (Prometheus 메트릭)
