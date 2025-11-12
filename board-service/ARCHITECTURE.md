# Board Service Architecture (2025-11-11)

## 📋 개요

Board Service는 **Clean Architecture + DDD 패턴**을 기반으로 설계된 백엔드 서비스입니다.
**7단계 점진적 리팩토링**을 통해 확장성, 테스트 용이성, 유지보수성을 개선했습니다.

---

## 🏗️ 레이어 구조

```
┌─────────────────────────────────────────────────────┐
│  Handler Layer (HTTP)                               │
│  - Gin 기반 REST API                                │
│  - 요청/응답 변환 (DTO ↔ Domain)                    │
└─────────────────────────────────────────────────────┘
                        ↓
┌─────────────────────────────────────────────────────┐
│  Service Layer (Business Logic)                     │
│  - 비즈니스 로직 조율                               │
│  - 트랜잭션 관리 (UnitOfWork)                       │
│  - 권한 체크 (Authorizer)                           │
└─────────────────────────────────────────────────────┘
                        ↓
┌─────────────────────────────────────────────────────┐
│  Repository Layer (Data Access)                     │
│  - Generic Base Repository                          │
│  - 도메인별 독립 Repository                         │
│  - Soft Delete 지원                                 │
└─────────────────────────────────────────────────────┘
                        ↓
┌─────────────────────────────────────────────────────┐
│  Domain Layer (Business Rules)                      │
│  - Rich Domain Model                                │
│  - 비즈니스 규칙 캡슐화                             │
│  - 도메인 메서드 (Assign, MarkAsDeleted 등)         │
└─────────────────────────────────────────────────────┘
```

---

## 📁 디렉토리 구조

```
board-service/
├── cmd/api/                    # 애플리케이션 진입점
│   ├── main.go                 # 부팅 로직
│   ├── injector.go             # 의존성 주입 (DI)
│   └── wire.go                 # Google Wire 설정 (준비)
│
├── internal/
│   ├── domain/                 # 도메인 모델 (13 files)
│   │   ├── base.go             # BaseModel + Entity Interface
│   │   ├── board.go            # Rich Domain Model (10 메서드)
│   │   ├── project.go          # Rich Domain Model (9 메서드)
│   │   └── comment.go          # Rich Domain Model (7 메서드)
│   │
│   ├── repository/             # 데이터 접근 (13 files)
│   │   ├── base/               # Generic Base Repository
│   │   │   └── base_repository.go
│   │   ├── board_repository.go
│   │   ├── project_repository.go
│   │   ├── field_repository.go         # Adapter Pattern
│   │   ├── project_field_repository.go # 독립 Repository (Stage 4)
│   │   ├── field_option_repository.go  # 독립 Repository
│   │   ├── field_value_repository.go   # 독립 Repository
│   │   ├── view_repository.go          # 독립 Repository
│   │   └── board_order_repository.go   # 독립 Repository
│   │
│   ├── service/                # 비즈니스 로직 (13 files)
│   │   ├── board_service.go
│   │   ├── board_service_with_uow.go   # UoW 예제
│   │   ├── project_service.go
│   │   └── ...
│   │
│   ├── handler/                # HTTP 핸들러 (6 files)
│   │   ├── board_handler.go
│   │   ├── project_handler.go
│   │   └── ...
│   │
│   ├── common/                 # 공통 유틸리티
│   │   ├── auth/               # 권한 관리 (중앙화)
│   │   ├── parser/             # UUID 파싱
│   │   ├── validator/          # 입력 검증
│   │   └── pagination/         # 페이지네이션
│   │
│   ├── uow/                    # Unit of Work 패턴
│   │   ├── unit_of_work.go
│   │   └── example.go          # 사용 예제
│   │
│   ├── dto/                    # Data Transfer Objects
│   ├── middleware/             # HTTP 미들웨어
│   ├── cache/                  # Redis 캐싱
│   ├── apperrors/              # 에러 처리
│   └── testutil/               # 테스트 헬퍼
│
└── pkg/                        # 공용 패키지
    ├── logger/                 # Zap 로거
    └── jwt/                    # JWT 토큰
```

---

## 🎯 7단계 리팩토링 결과

### Stage 1: 테스트 인프라 구축 ✅
- **파일**: `internal/testutil/`
- **성과**: Mock 기반 유닛 테스트, 통합 테스트 환경 구축
- **테스트 파일**: 7개

### Stage 2: 공통 유틸리티 추출 ✅
- **파일**: `internal/common/`
- **성과**: UUID 파싱, 페이지네이션, 검증 로직 중복 제거
- **코드 감소**: 100+ 줄 제거

### Stage 3: 권한 체크 중앙화 ✅
- **파일**: `internal/common/auth/authorizer.go`
- **성과**: ProjectAuthorizer로 일관된 권한 관리
- **코드 감소**: 각 메서드당 18줄 → 4줄

### Stage 4: Repository 책임 분리 ✅
- **파일**: 5개 독립 Repository + Adapter 패턴
- **성과**: 단일 책임 원칙 준수, 테스트 용이성 향상
- **패턴**: Adapter Pattern으로 호환성 유지

### Stage 5: Generic Base Repository 패턴 ✅
- **파일**: `repository/base/base_repository.go`
- **성과**: 공통 CRUD 중복 제거, 타입 안전성
- **Generic**: Go 1.18+ Generics 활용

### Stage 6: DI 구조 개선 ✅
- **파일**: `cmd/api/injector.go`, `wire.go`
- **성과**: main.go 간소화 (225줄 → 117줄)
- **패턴**: Factory Pattern + Wire 준비

### Stage 7: Rich Domain Model + Unit of Work ✅
- **파일**: `domain/*.go`, `uow/`
- **성과**: 비즈니스 로직을 Domain으로 이동
- **메서드**: Board(10), Project(9), Comment(7)

---

## 🔑 핵심 패턴

### 1. Rich Domain Model
**Before (Anemic Domain)**:
```go
// Service가 모든 로직 수행
board.Title = req.Title
board.AssigneeID = assigneeUUID
```

**After (Rich Domain)**:
```go
// Domain이 비즈니스 로직 캡슐화
board.UpdateTitle(req.Title)      // 검증 포함
board.Assign(assigneeUUID)        // 상태 변경
```

**Domain 메서드 예시**:
- `Board`: `IsOverdue()`, `Assign()`, `UpdateTitle()`, `MarkAsDeleted()`
- `Project`: `IsOwnedBy()`, `MakePublic()`, `TransferOwnership()`
- `Comment`: `IsWrittenBy()`, `UpdateContent()`, `WasEdited()`

### 2. Unit of Work 패턴
**복잡한 트랜잭션 관리**:
```go
return s.uow.Do(func(repos *uow.Repositories) error {
    // 1. 보드 삭제
    board, _ := repos.Board.FindByID(boardID)
    board.MarkAsDeleted()
    repos.Board.Update(board)

    // 2. 관련 댓글 삭제
    comments, _ := repos.Comment.FindByBoard(boardID)
    for _, c := range comments {
        repos.Comment.Delete(c.ID)
    }

    // 모두 성공하거나 모두 실패 (원자성)
    return nil
})
```

### 3. Generic Base Repository
**타입 안전한 CRUD**:
```go
type BaseRepository[T Entity] interface {
    Create(entity T) error
    FindByID(id uuid.UUID) (T, error)
    Update(entity T) error
    Delete(id uuid.UUID) error
}

// 사용
boardRepo := base.NewBaseRepository[*domain.Board](db)
```

### 4. Centralized Authorization
**일관된 권한 관리**:
```go
type ProjectAuthorizer interface {
    RequireMember(userID, projectID uuid.UUID) (*domain.ProjectMember, error)
    RequireAdmin(userID, projectID uuid.UUID) (*domain.ProjectMember, error)
    CanEdit(userID, projectID, authorID uuid.UUID) (bool, error)
    CanDelete(userID, projectID, authorID uuid.UUID) (bool, error)
}

// Service에서 사용
canEdit, _ := s.authorizer.CanEdit(userID, projectID, authorID)
if !canEdit {
    return errors.New("권한 없음")
}
```

---

## 🔄 의존성 흐름

```
main.go
  ↓ (InitializeApplication)
Injector
  ↓ (DI)
Handler → Service → Repository → Domain
              ↓
        Authorizer (권한)
        Parser (파싱)
        Validator (검증)
        UnitOfWork (트랜잭션)
```

---

## 📊 주요 지표

| 항목 | Before | After | 개선 |
|-----|--------|-------|------|
| main.go 라인수 | 225 | 117 | -48% |
| 권한 체크 중복 | 18줄/메서드 | 4줄/메서드 | -78% |
| Repository 책임 | 1개 (349줄) | 5개 독립 | SRP |
| Domain 메서드 | 0개 | 26개 | Rich |
| 테스트 파일 | 0개 | 7개 | ✅ |

---

## 🚀 기술 스택

| 레이어 | 기술 |
|-------|------|
| **Web Framework** | Gin |
| **ORM** | GORM |
| **Database** | PostgreSQL |
| **Cache** | Redis |
| **Logger** | Zap |
| **Auth** | JWT |
| **DI** | Manual (Wire 준비) |
| **Testing** | testify |

---

## 📝 API 엔드포인트

### Projects
- `POST /api/projects` - 프로젝트 생성
- `GET /api/projects` - 프로젝트 목록
- `GET /api/projects/:id` - 프로젝트 조회
- `PUT /api/projects/:id` - 프로젝트 수정
- `DELETE /api/projects/:id` - 프로젝트 삭제

### Boards
- `POST /api/boards` - 보드 생성
- `GET /api/boards` - 보드 목록
- `GET /api/boards/:id` - 보드 조회
- `PUT /api/boards/:id` - 보드 수정
- `DELETE /api/boards/:id` - 보드 삭제
- `PUT /api/boards/:id/move` - 보드 이동

### Comments
- `POST /api/comments` - 댓글 생성
- `GET /api/comments` - 댓글 목록
- `PUT /api/comments/:id` - 댓글 수정
- `DELETE /api/comments/:id` - 댓글 삭제

### Custom Fields
- `POST /api/fields` - 필드 생성
- `GET /api/fields/:id` - 필드 조회
- `PATCH /api/fields/:id` - 필드 수정
- `DELETE /api/fields/:id` - 필드 삭제

---

## 🔐 보안

### 인증
- **JWT 토큰** 기반 인증
- Authorization 헤더: `Bearer <token>`

### 권한
- **OWNER** (100): 모든 권한
- **ADMIN** (50): 콘텐츠 편집/삭제
- **MEMBER** (10): 콘텐츠 생성/자신의 콘텐츠 수정

### 권한 체크 흐름
```
1. AuthMiddleware (JWT 검증)
   ↓
2. Handler (요청 파싱)
   ↓
3. Service (비즈니스 로직)
   ↓
4. ProjectAuthorizer (권한 확인)
   ↓
5. Repository (데이터 접근)
```

---

## 🧪 테스트 전략

### Unit Test
- **Mock Repository** 사용
- **testify/mock** 프레임워크
- Service 레이어 집중

### Integration Test
- **In-memory SQLite** 또는 **Postgres**
- 실제 DB 사용
- Repository 레이어 검증

### 테스트 파일 위치
```
internal/
├── service/
│   ├── board_service_unit_test.go
│   ├── field_service_test.go
│   └── view_service_test.go
├── repository/
│   └── board_repository_test.go
└── common/auth/
    └── authorizer_test.go
```

---

## 📈 향후 계획

### 단기 (1-2주)
- [ ] 나머지 Service에 Domain 메서드 적용
- [ ] UnitOfWork를 Service에 실제 주입
- [ ] 테스트 커버리지 80% 달성

### 중기 (1-2개월)
- [ ] Google Wire 완전 도입
- [ ] CQRS 패턴 적용 (읽기/쓰기 분리)
- [ ] Event Sourcing (히스토리 추적)

### 장기 (3-6개월)
- [ ] Saga 패턴 (분산 트랜잭션)
- [ ] Microservices 분리
- [ ] gRPC 도입

---

## 📚 참고 문서

- [Unit of Work 사용 가이드](internal/service/board_service_with_uow.go)
- [Generic Repository 패턴](internal/repository/base/base_repository.go)
- [Domain Model 메서드](internal/domain/)
- [권한 관리 가이드](internal/common/auth/authorizer.go)

---

**Last Updated**: 2025-11-11
**Architecture Version**: 2.0 (7-Stage Refactoring Complete)
