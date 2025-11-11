# Board Service 개선 사항 (2025-11-11)

## 📋 분석 결과 요약

7단계 리팩토링을 통해 구조가 크게 개선되었으나, 다음 영역에서 추가 개선이 필요합니다:

---

## 🔴 높은 우선순위 (High Priority)

### 1. ❌ Service에 UnitOfWork 미적용
**현재 상태**:
- `board_service_with_uow.go`에 예제만 존재
- 실제 Service에는 UoW가 주입되지 않음
- 복잡한 트랜잭션 로직이 수동 관리됨

**문제점**:
```go
// 현재: 트랜잭션 관리가 수동
func (s *boardService) DeleteBoard() error {
    // 1. 보드 삭제
    s.repo.Delete(boardID)

    // 2. 댓글 삭제 (별도 트랜잭션)
    // ⚠️ 보드 삭제 성공 후 댓글 삭제 실패 시 불일치 발생
    s.commentRepo.DeleteByBoard(boardID)
}
```

**개선 방안**:
```go
// 개선: UnitOfWork 주입 및 적용
type boardService struct {
    uow uow.UnitOfWork  // 추가
    // ... 기타 의존성
}

func (s *boardService) DeleteBoard() error {
    return s.uow.Do(func(repos *uow.Repositories) error {
        board, _ := repos.Board.FindByID(boardID)
        board.MarkAsDeleted()
        repos.Board.Update(board)

        // 관련 댓글도 트랜잭션으로 삭제
        comments, _ := repos.Comment.FindByBoard(boardID)
        for _, c := range comments {
            repos.Comment.Delete(c.ID)
        }
        return nil  // 원자성 보장
    })
}
```

**영향 범위**:
- `BoardService`, `ProjectService`, `CommentService`
- `cmd/api/injector.go` (UoW 주입 추가)

**예상 작업 시간**: 4-6시간

---

### 2. ❌ Comment 도메인이 BaseModel을 사용하지 않음
**현재 상태**:
```go
// Comment는 BaseModel 사용 안함
type Comment struct {
    ID        uuid.UUID      `gorm:"type:uuid;primaryKey"`
    Content   string         `gorm:"type:text;not null"`
    UserID    uuid.UUID      `gorm:"type:uuid;not null;index"`
    BoardID   uuid.UUID      `gorm:"type:uuid;not null;index"`
    CreatedAt time.Time      `gorm:"autoCreateTime"`
    UpdatedAt time.Time      `gorm:"autoUpdateTime"`
    DeletedAt gorm.DeletedAt `gorm:"index"`  // ⚠️ Hard Delete 사용
}
```

**문제점**:
1. **일관성 부족**: 다른 엔티티는 BaseModel 사용 (Soft Delete)
2. **Generic Repository 미사용**: BaseModel이 없어 Base Repository 활용 불가
3. **Entity 인터페이스 미구현**: GetID(), SetIsDeleted() 없음

**개선 방안**:
```go
// 개선: BaseModel 사용으로 일관성 확보
type Comment struct {
    BaseModel  // ID, CreatedAt, UpdatedAt, IsDeleted 포함
    Content   string         `gorm:"type:text;not null"`
    UserID    uuid.UUID      `gorm:"type:uuid;not null;index"`
    BoardID   uuid.UUID      `gorm:"type:uuid;not null;index"`
    Board     Board          `gorm:"foreignKey:BoardID"`
}

// Entity 인터페이스 자동 구현 (BaseModel에서 상속)
```

**마이그레이션 필요**:
```sql
-- DeletedAt (gorm.DeletedAt) → IsDeleted (bool) 변환
ALTER TABLE comments ADD COLUMN is_deleted BOOLEAN DEFAULT FALSE;
UPDATE comments SET is_deleted = TRUE WHERE deleted_at IS NOT NULL;
ALTER TABLE comments DROP COLUMN deleted_at;
```

**예상 작업 시간**: 2-3시간

---

### 3. ⚠️ 에러 처리 일관성 부족
**현재 상태**:
```go
// Service마다 에러 처리 방식이 다름
return apperrors.New(apperrors.ErrCodeNotFound, "보드를 찾을 수 없습니다", 404)
return apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "보드 조회 실패", 500)
return errors.New("invalid input")  // ⚠️ 일반 에러
```

**문제점**:
1. `apperrors.New()`, `apperrors.Wrap()`, `errors.New()` 혼용
2. 에러 코드와 HTTP 상태 코드가 분리되어 관리
3. Domain 에러와 Infrastructure 에러 구분 부족

**개선 방안**:
```go
// 1. Domain 레벨 에러
type DomainError struct {
    Code    string
    Message string
    Field   string  // Validation 에러용
}

// 2. Infrastructure 레벨 에러
type AppError struct {
    Code       string
    Message    string
    HTTPStatus int
    Err        error  // Original error
}

// 3. 사용 예시
// Domain
func (b *Board) UpdateTitle(title string) error {
    if title == "" {
        return &DomainError{
            Code:    "BOARD_TITLE_REQUIRED",
            Message: "제목은 필수입니다",
            Field:   "title",
        }
    }
    return nil
}

// Service
func (s *boardService) UpdateBoard() error {
    if err := board.UpdateTitle(req.Title); err != nil {
        // Domain 에러를 App 에러로 변환
        return &AppError{
            Code:       "VALIDATION_ERROR",
            Message:    err.Error(),
            HTTPStatus: 400,
            Err:        err,
        }
    }
}
```

**예상 작업 시간**: 3-4시간

---

### 4. ⚠️ Repository 인터페이스가 구현체와 같은 파일에 있음
**현재 상태**:
```go
// board_repository.go
type BoardRepository interface {  // 인터페이스
    Create(board *domain.Board) error
    FindByID(id uuid.UUID) (*domain.Board, error)
}

type boardRepository struct {  // 구현체
    db *gorm.DB
}
```

**문제점**:
1. 테스트 시 Mock 구현이 어려움
2. 인터페이스와 구현체 분리 원칙 위배
3. 순환 의존성 발생 가능

**개선 방안**:
```
repository/
├── interfaces.go          # 모든 Repository 인터페이스
│   ├── BoardRepository
│   ├── ProjectRepository
│   └── CommentRepository
│
└── impl/                  # 구현체
    ├── board_repository.go
    ├── project_repository.go
    └── comment_repository.go
```

```go
// repository/interfaces.go
package repository

type BoardRepository interface {
    Create(board *domain.Board) error
    FindByID(id uuid.UUID) (*domain.Board, error)
    // ...
}

// repository/impl/board_repository.go
package impl

type boardRepository struct {
    db *gorm.DB
}

func NewBoardRepository(db *gorm.DB) repository.BoardRepository {
    return &boardRepository{db: db}
}
```

**예상 작업 시간**: 2-3시간

---

## 🟡 중간 우선순위 (Medium Priority)

### 5. 📦 DTO와 Domain 변환 로직이 여러 곳에 분산
**현재 상태**:
```go
// Service 곳곳에 변환 로직 산재
func (s *boardService) GetBoard() (*dto.BoardResponse, error) {
    board, _ := s.repo.FindByID(boardID)

    // 변환 로직이 Service에 있음
    return &dto.BoardResponse{
        ID:          board.ID.String(),
        Title:       board.Title,
        Description: board.Description,
        // ... 20줄 이상
    }, nil
}
```

**문제점**:
1. 변환 로직 중복 (GetBoard, GetBoards, CreateBoard 등)
2. Service가 DTO 상세 구조에 의존
3. DTO 필드 변경 시 여러 Service 수정 필요

**개선 방안**:
```go
// dto/mapper.go 생성
package dto

type BoardMapper struct{}

func (m *BoardMapper) ToResponse(board *domain.Board) *BoardResponse {
    return &BoardResponse{
        ID:          board.ID.String(),
        Title:       board.Title,
        Description: board.Description,
        IsOverdue:   board.IsOverdue(),  // Domain 메서드 활용
        // ...
    }
}

func (m *BoardMapper) ToResponseList(boards []domain.Board) []BoardResponse {
    responses := make([]BoardResponse, len(boards))
    for i, board := range boards {
        responses[i] = *m.ToResponse(&board)
    }
    return responses
}

// Service에서 사용
func (s *boardService) GetBoard() (*dto.BoardResponse, error) {
    board, _ := s.repo.FindByID(boardID)
    return s.mapper.ToResponse(board), nil  // 간결
}
```

**예상 작업 시간**: 3-4시간

---

### 6. 🔍 로깅 전략 부재
**현재 상태**:
```go
// 로깅이 일관되지 않음
s.logger.Info("보드 생성", zap.String("id", board.ID.String()))
// 어떤 곳은 로깅 없음
// 어떤 곳은 Debug, 어떤 곳은 Info
```

**문제점**:
1. 로그 레벨 일관성 부족 (Debug, Info, Warn, Error 혼용)
2. 구조화된 로깅 부족 (context, trace ID)
3. 비즈니스 이벤트 로깅 부족

**개선 방안**:
```go
// 1. 로깅 레벨 정책 수립
// Debug: 개발 시 디버깅
// Info:  중요한 비즈니스 이벤트 (생성, 수정, 삭제)
// Warn:  예상된 에러 (권한 부족, 유효성 검증 실패)
// Error: 예상 못한 에러 (DB 연결 실패, 외부 API 실패)

// 2. 구조화된 로깅
func (s *boardService) CreateBoard() error {
    s.logger.Info("보드 생성 시작",
        zap.String("user_id", userID),
        zap.String("project_id", projectID),
        zap.String("request_id", ctx.Value("request_id")),  // Trace
    )

    // ... 비즈니스 로직

    s.logger.Info("보드 생성 완료",
        zap.String("board_id", board.ID.String()),
        zap.Duration("duration", time.Since(start)),
    )
}

// 3. 비즈니스 이벤트 로깅 (Audit Log)
type AuditLogger struct {
    logger *zap.Logger
}

func (a *AuditLogger) LogBoardCreated(board *domain.Board, userID uuid.UUID) {
    a.logger.Info("BOARD_CREATED",
        zap.String("event", "BOARD_CREATED"),
        zap.String("board_id", board.ID.String()),
        zap.String("user_id", userID.String()),
        zap.Time("timestamp", time.Now()),
    )
}
```

**예상 작업 시간**: 2-3시간

---

### 7. 🧪 테스트 커버리지 부족
**현재 상태**:
- 테스트 파일: 7개
- 테스트 커버리지: 추정 20-30%

**문제점**:
1. Service 레이어 테스트 부족 (3개만 존재)
2. Repository 테스트 부족 (1개만 존재)
3. 통합 테스트 부족
4. E2E 테스트 없음

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

### 8. 🔄 Cache 전략이 명확하지 않음
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

### 9. 📊 메트릭 및 모니터링 부족
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

## 📈 개선 로드맵 (권장 순서)

### Phase 1: 핵심 기능 안정화 (1-2주)
1. **UnitOfWork 실제 적용** (High Priority #1)
2. **Comment BaseModel 마이그레이션** (High Priority #2)
3. **에러 처리 일관성** (High Priority #3)

### Phase 2: 코드 품질 향상 (2-3주)
4. **Repository 인터페이스 분리** (High Priority #4)
5. **DTO Mapper 도입** (Medium Priority #5)
6. **테스트 커버리지 80%** (Medium Priority #7)

### Phase 3: 운영 안정성 (3-4주)
7. **로깅 전략 수립** (Medium Priority #6)
8. **Cache 전략 명확화** (Medium Priority #8)
9. **메트릭 및 모니터링** (Low Priority #9)

### Phase 4: 확장성 (4-8주)
10. **Rate Limiting** (Low Priority #10)
11. **국제화 지원** (Low Priority #11)

---

## 📊 예상 총 작업 시간

| 우선순위 | 작업 수 | 총 시간 |
|---------|--------|---------|
| 🔴 High | 4개 | 11-16시간 |
| 🟡 Medium | 4개 | 17-27시간 |
| 🟢 Low | 3개 | 8-12시간 |
| **총합** | **11개** | **36-55시간** |

---

## ✅ 개선 완료 시 기대 효과

### 코드 품질
- ✅ 트랜잭션 안정성 향상 (UnitOfWork)
- ✅ 일관된 에러 처리
- ✅ 테스트 커버리지 80%+

### 개발 생산성
- ✅ DTO 변환 로직 중복 제거
- ✅ Repository 인터페이스 분리로 Mock 용이
- ✅ 명확한 로깅 전략

### 운영 안정성
- ✅ 구조화된 로깅으로 디버깅 향상
- ✅ Cache 전략으로 성능 개선
- ✅ Rate Limiting으로 서비스 보호

---

**Last Updated**: 2025-11-11
**Next Review**: Phase 1 완료 후
