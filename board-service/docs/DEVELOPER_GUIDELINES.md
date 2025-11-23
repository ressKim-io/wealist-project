# Board Service 개발자 가이드

## 목차

- [개요](#개요)
- [개발 환경 설정](#개발-환경-설정)
- [코딩 표준](#코딩-표준)
- [API 개발 워크플로우](#api-개발-워크플로우)
- [Swagger 문서 유지보수](#swagger-문서-유지보수)
- [테스트 작성 가이드](#테스트-작성-가이드)
- [Git 워크플로우](#git-워크플로우)
- [코드 리뷰 체크리스트](#코드-리뷰-체크리스트)

## 개요

이 문서는 Board Service 프로젝트에 기여하는 개발자를 위한 가이드입니다. 코딩 표준, 개발 워크플로우, 문서화 규칙 등을 다룹니다.

## 개발 환경 설정

### 필수 도구

1. **Go 1.21+**
   ```bash
   go version
   ```

2. **PostgreSQL 14+**
   ```bash
   psql --version
   ```

3. **Make**
   ```bash
   make --version
   ```

4. **Docker & Docker Compose** (선택사항)
   ```bash
   docker --version
   docker-compose --version
   ```

### 개발 도구 설치

```bash
# 프로젝트 의존성
make deps

# 개발 도구 (air, golangci-lint, swag)
make install-tools
```

### 로컬 개발 환경 설정

1. **환경 변수 설정**
   ```bash
   cp .env.example .env
   # .env 파일 편집
   ```

2. **데이터베이스 설정**
   ```bash
   make db-create
   make migrate-up
   ```

3. **서버 실행**
   ```bash
   # Hot reload 개발 모드
   make dev
   
   # 또는 일반 실행
   make run
   ```

## 코딩 표준

### Go 코드 스타일

1. **공식 Go 스타일 가이드 준수**
   - [Effective Go](https://golang.org/doc/effective_go)
   - [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)

2. **코드 포맷팅**
   ```bash
   # 자동 포맷팅
   make fmt
   
   # 코드 검사
   make vet
   
   # Lint 실행
   make lint
   ```

3. **네이밍 컨벤션**
   - 패키지명: 소문자, 단수형 (예: `handler`, `service`)
   - 인터페이스: 동사 + er (예: `BoardRepository`, `UserService`)
   - 구조체: PascalCase (예: `BoardHandler`, `CreateBoardRequest`)
   - 함수/메서드: PascalCase (public), camelCase (private)
   - 상수: PascalCase 또는 UPPER_SNAKE_CASE

### 프로젝트 구조

```
internal/
├── handler/      # HTTP 핸들러 (요청/응답 처리)
├── service/      # 비즈니스 로직
├── repository/   # 데이터 접근 계층
├── domain/       # 도메인 모델
├── dto/          # 데이터 전송 객체
├── middleware/   # HTTP 미들웨어
├── response/     # 공통 응답 헬퍼
├── config/       # 설정 관리
├── database/     # 데이터베이스 연결
└── logger/       # 로깅 설정
```

### 계층별 책임

#### Handler Layer
- HTTP 요청 파싱 및 검증
- DTO 바인딩
- Service 호출
- HTTP 응답 생성
- **비즈니스 로직 포함 금지**

```go
func (h *BoardHandler) CreateBoard(c *gin.Context) {
    var req dto.CreateBoardRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Error(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
        return
    }
    
    board, err := h.service.CreateBoard(c.Request.Context(), &req)
    if err != nil {
        handler.HandleError(c, err)
        return
    }
    
    response.Success(c, http.StatusCreated, board, "Board created successfully")
}
```

#### Service Layer
- 비즈니스 로직 구현
- 트랜잭션 관리
- Repository 호출
- 도메인 규칙 적용

```go
func (s *BoardService) CreateBoard(ctx context.Context, req *dto.CreateBoardRequest) (*dto.BoardResponse, error) {
    // 비즈니스 검증
    if err := s.validateProject(ctx, req.ProjectID); err != nil {
        return nil, err
    }
    
    // 도메인 객체 생성
    board := &domain.Board{
        ProjectID:  req.ProjectID,
        Title:      req.Title,
        Content:    req.Content,
        Stage:      req.Stage,
        Importance: req.Importance,
        Role:       req.Role,
    }
    
    // 저장
    if err := s.repo.Create(ctx, board); err != nil {
        return nil, err
    }
    
    return dto.ToBoardResponse(board), nil
}
```

#### Repository Layer
- 데이터베이스 쿼리 실행
- GORM 작업
- 도메인 모델 반환

```go
func (r *BoardRepository) Create(ctx context.Context, board *domain.Board) error {
    return r.db.WithContext(ctx).Create(board).Error
}

func (r *BoardRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Board, error) {
    var board domain.Board
    err := r.db.WithContext(ctx).
        Where("id = ? AND deleted_at IS NULL", id).
        First(&board).Error
    if err != nil {
        return nil, err
    }
    return &board, nil
}
```

### 에러 처리

1. **커스텀 에러 타입 사용**
   ```go
   var (
       ErrBoardNotFound = errors.New("board not found")
       ErrInvalidStage  = errors.New("invalid stage value")
   )
   ```

2. **에러 래핑**
   ```go
   if err := repo.Create(ctx, board); err != nil {
       return nil, fmt.Errorf("failed to create board: %w", err)
   }
   ```

3. **에러 핸들링**
   ```go
   if errors.Is(err, ErrBoardNotFound) {
       response.Error(c, http.StatusNotFound, "BOARD_NOT_FOUND", "Board not found")
       return
   }
   ```

## API 개발 워크플로우

### 1. 요구사항 분석

- API 스펙 정의 (엔드포인트, 메서드, 요청/응답)
- 비즈니스 규칙 파악
- 데이터 모델 설계

### 2. 도메인 모델 작성

```go
// internal/domain/board.go
type Board struct {
    Base
    ProjectID  uuid.UUID `gorm:"type:uuid;not null;index"`
    Title      string    `gorm:"type:varchar(200);not null"`
    Content    string    `gorm:"type:text"`
    Stage      string    `gorm:"type:varchar(50);not null"`
    Importance string    `gorm:"type:varchar(50);not null"`
    Role       string    `gorm:"type:varchar(50);not null"`
}
```

### 3. DTO 작성

```go
// internal/dto/board_dto.go
type CreateBoardRequest struct {
    ProjectID  uuid.UUID `json:"projectId" binding:"required"`
    Title      string    `json:"title" binding:"required,max=200"`
    Content    string    `json:"content" binding:"max=5000"`
    Stage      string    `json:"stage" binding:"required,oneof=in_progress pending approved review"`
    Importance string    `json:"importance" binding:"required,oneof=urgent normal"`
    Role       string    `json:"role" binding:"required,oneof=developer planner"`
}

type BoardResponse struct {
    BoardID    uuid.UUID `json:"boardId"`
    ProjectID  uuid.UUID `json:"projectId"`
    Title      string    `json:"title"`
    Content    string    `json:"content"`
    Stage      string    `json:"stage"`
    Importance string    `json:"importance"`
    Role       string    `json:"role"`
    CreatedAt  time.Time `json:"createdAt"`
    UpdatedAt  time.Time `json:"updatedAt"`
}
```

### 4. Repository 구현

```go
// internal/repository/board_repository.go
type BoardRepository interface {
    Create(ctx context.Context, board *domain.Board) error
    FindByID(ctx context.Context, id uuid.UUID) (*domain.Board, error)
    Update(ctx context.Context, board *domain.Board) error
    Delete(ctx context.Context, id uuid.UUID) error
}

type boardRepository struct {
    db *gorm.DB
}

func NewBoardRepository(db *gorm.DB) BoardRepository {
    return &boardRepository{db: db}
}
```

### 5. Service 구현

```go
// internal/service/board_service.go
type BoardService interface {
    CreateBoard(ctx context.Context, req *dto.CreateBoardRequest) (*dto.BoardResponse, error)
    GetBoard(ctx context.Context, id uuid.UUID) (*dto.BoardDetailResponse, error)
    UpdateBoard(ctx context.Context, id uuid.UUID, req *dto.UpdateBoardRequest) (*dto.BoardResponse, error)
    DeleteBoard(ctx context.Context, id uuid.UUID) error
}

type boardService struct {
    repo BoardRepository
}

func NewBoardService(repo BoardRepository) BoardService {
    return &boardService{repo: repo}
}
```

### 6. Handler 구현 (Godoc 주석 포함)

```go
// internal/handler/board_handler.go
type BoardHandler struct {
    service service.BoardService
}

func NewBoardHandler(service service.BoardService) *BoardHandler {
    return &BoardHandler{service: service}
}

// CreateBoard godoc
// @Summary      Board 생성
// @Description  새로운 Board를 생성합니다
// @Tags         boards
// @Accept       json
// @Produce      json
// @Param        request body dto.CreateBoardRequest true "Board 생성 요청"
// @Success      201 {object} response.SuccessResponse{data=dto.BoardResponse}
// @Failure      400 {object} response.ErrorResponse
// @Failure      404 {object} response.ErrorResponse
// @Failure      500 {object} response.ErrorResponse
// @Router       /boards [post]
func (h *BoardHandler) CreateBoard(c *gin.Context) {
    // 구현
}
```

### 7. 라우터 등록

```go
// internal/router/router.go
func SetupRouter(handler *handler.BoardHandler) *gin.Engine {
    router := gin.Default()
    
    api := router.Group("/api")
    {
        boards := api.Group("/boards")
        {
            boards.POST("", handler.CreateBoard)
            boards.GET("/:boardId", handler.GetBoard)
            boards.PUT("/:boardId", handler.UpdateBoard)
            boards.DELETE("/:boardId", handler.DeleteBoard)
        }
    }
    
    return router
}
```

### 8. Swagger 문서 생성 및 검증

```bash
# Swagger 문서 생성
make swagger

# 검증
./scripts/validate-swagger.sh
./scripts/validate-dto-schemas.sh
./scripts/validate-godoc.sh
```

### 9. 테스트 작성

```bash
# 테스트 실행
make test

# 커버리지 확인
make test-coverage
```

### 10. 수동 테스트

```bash
# 서버 실행
make run

# Swagger UI 접속
open http://localhost:8000/swagger/index.html

# API 테스트
curl -X POST http://localhost:8000/api/boards \
  -H "Content-Type: application/json" \
  -d '{"projectId":"...","title":"Test Board",...}'
```

## Swagger 문서 유지보수

### Godoc 주석 작성 규칙

1. **모든 핸들러에 주석 필수**
   - @Summary: 필수
   - @Description: 권장
   - @Tags: 필수
   - @Router: 필수

2. **일관된 한글 용어 사용**
   - "생성" (Create)
   - "조회" (Get/Read)
   - "수정" (Update)
   - "삭제" (Delete)

3. **에러 응답 문서화**
   - 모든 가능한 HTTP 상태 코드 포함
   - 명확한 에러 메시지

4. **파라미터 설명 명확히**
   - 타입 명시 (UUID, string, int 등)
   - 필수/선택 구분
   - Enum 값 나열

### Swagger 문서 업데이트 프로세스

1. **코드 변경**
   - 핸들러 추가/수정
   - DTO 추가/수정

2. **Godoc 주석 업데이트**
   - 변경사항 반영
   - 새 엔드포인트 주석 추가

3. **문서 재생성**
   ```bash
   make swagger
   ```

4. **검증**
   ```bash
   ./scripts/validate-swagger.sh
   ./scripts/validate-dto-schemas.sh
   ./scripts/validate-godoc.sh
   ```

5. **수동 확인**
   - Swagger UI에서 확인
   - API 테스트

6. **커밋**
   - docs/ 디렉토리 포함하여 커밋

### 검증 스크립트 사용

#### 엔드포인트 커버리지 검증
```bash
./scripts/validate-swagger.sh
```
- 100% 커버리지 확인
- 누락된 엔드포인트 리포트

#### DTO 스키마 검증
```bash
./scripts/validate-dto-schemas.sh
```
- 모든 DTO가 swagger에 정의되어 있는지 확인

#### Godoc 주석 품질 검증
```bash
./scripts/validate-godoc.sh
```
- 필수 주석 태그 확인
- 불완전한 주석 리포트

## 테스트 작성 가이드

### 테스트 구조

```
internal/
├── handler/
│   ├── board_handler.go
│   └── board_handler_test.go
├── service/
│   ├── board_service.go
│   └── board_service_test.go
└── repository/
    ├── board_repository.go
    └── board_repository_test.go
```

### Unit Test 예시

```go
// internal/service/board_service_test.go
func TestBoardService_CreateBoard(t *testing.T) {
    // Setup
    mockRepo := &MockBoardRepository{}
    service := NewBoardService(mockRepo)
    
    req := &dto.CreateBoardRequest{
        ProjectID:  uuid.New(),
        Title:      "Test Board",
        Stage:      "in_progress",
        Importance: "urgent",
        Role:       "developer",
    }
    
    // Mock 설정
    mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
    
    // 실행
    result, err := service.CreateBoard(context.Background(), req)
    
    // 검증
    assert.NoError(t, err)
    assert.NotNil(t, result)
    assert.Equal(t, req.Title, result.Title)
    mockRepo.AssertExpectations(t)
}
```

### Integration Test 예시

```go
// internal/handler/board_handler_test.go
func TestBoardHandler_CreateBoard(t *testing.T) {
    // Setup
    router := setupTestRouter()
    
    reqBody := `{
        "projectId": "550e8400-e29b-41d4-a716-446655440000",
        "title": "Test Board",
        "stage": "in_progress",
        "importance": "urgent",
        "role": "developer"
    }`
    
    // 요청 생성
    req, _ := http.NewRequest("POST", "/api/boards", strings.NewReader(reqBody))
    req.Header.Set("Content-Type", "application/json")
    
    // 응답 기록
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)
    
    // 검증
    assert.Equal(t, http.StatusCreated, w.Code)
    
    var response map[string]interface{}
    json.Unmarshal(w.Body.Bytes(), &response)
    assert.NotNil(t, response["data"])
}
```

### 테스트 실행

```bash
# 모든 테스트
make test

# 특정 패키지
go test ./internal/service/...

# 커버리지
make test-coverage

# Verbose 모드
go test -v ./...
```

## Git 워크플로우

### 브랜치 전략

- `main`: 프로덕션 코드
- `develop`: 개발 브랜치
- `feature/*`: 기능 개발
- `bugfix/*`: 버그 수정
- `hotfix/*`: 긴급 수정

### 커밋 메시지 규칙

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Type:**
- `feat`: 새 기능
- `fix`: 버그 수정
- `docs`: 문서 변경
- `style`: 코드 포맷팅
- `refactor`: 리팩토링
- `test`: 테스트 추가/수정
- `chore`: 빌드/설정 변경

**예시:**
```
feat(board): add custom fields filtering

- Add customFields query parameter to GetBoardsByProject
- Update swagger documentation
- Add validation for custom field filters

Closes #123
```

### Pull Request 프로세스

1. **브랜치 생성**
   ```bash
   git checkout -b feature/custom-fields-filter
   ```

2. **개발 및 커밋**
   ```bash
   git add .
   git commit -m "feat(board): add custom fields filtering"
   ```

3. **Swagger 문서 업데이트**
   ```bash
   make swagger
   git add docs/
   git commit -m "docs(swagger): update board API documentation"
   ```

4. **검증**
   ```bash
   make check
   make test
   ./scripts/validate-swagger.sh
   ```

5. **Push 및 PR 생성**
   ```bash
   git push origin feature/custom-fields-filter
   ```

6. **코드 리뷰 대응**

7. **Merge**

## 코드 리뷰 체크리스트

### 기능 구현

- [ ] 요구사항을 정확히 구현했는가?
- [ ] 에지 케이스를 고려했는가?
- [ ] 에러 처리가 적절한가?

### 코드 품질

- [ ] Clean Architecture 원칙을 따르는가?
- [ ] 계층별 책임이 명확한가?
- [ ] 코드가 읽기 쉽고 이해하기 쉬운가?
- [ ] 중복 코드가 없는가?

### 테스트

- [ ] Unit test가 작성되었는가?
- [ ] 테스트 커버리지가 충분한가?
- [ ] 모든 테스트가 통과하는가?

### 문서화

- [ ] Godoc 주석이 완전한가?
- [ ] Swagger 문서가 업데이트되었는가?
- [ ] 검증 스크립트가 통과하는가?
- [ ] README가 업데이트되었는가? (필요시)

### 성능 및 보안

- [ ] N+1 쿼리 문제가 없는가?
- [ ] 입력 검증이 적절한가?
- [ ] SQL Injection 취약점이 없는가?
- [ ] 민감한 정보가 로그에 노출되지 않는가?

### Git

- [ ] 커밋 메시지가 규칙을 따르는가?
- [ ] 불필요한 파일이 커밋되지 않았는가?
- [ ] docs/ 디렉토리가 포함되었는가?

## 참고 자료

- [Effective Go](https://golang.org/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [Swagger Documentation](docs/SWAGGER.md)

## 문의

개발 가이드 관련 문의사항이나 개선 제안은 GitHub Issues를 이용해주세요.
