# Board Service Optimization Plan

> 분석 날짜: 2025-11-07
> 목표: User Service 호출 최소화, 성능 최적화

## 📊 현재 상황 분석

### 1. User Service 호출 패턴

#### Workspace 검증 (빈번한 호출)
- `CheckWorkspaceExists` - workspace 존재 확인 (1회)
- `ValidateWorkspaceMembership` - 멤버십 검증 (5회)
  - CreateProject
  - GetProjectsByWorkspaceID
  - SearchProjects
  - CreateJoinRequest
  - 기타

#### 사용자 정보 조회
- `GetUser` - 단일 사용자 조회 (3회)
  - project owner 정보
  - member 정보
  - join request user 정보
- `GetUsersBatch` - 배치 조회 (1회 - board service)
- `GetSimpleUser` - 간단한 사용자 정보 (2회 - comment service)
- `GetSimpleUsers` - 배치 간단 정보 (1회 - comment service)

### 2. 문제점

#### 🔴 문제 1: API 엔드포인트 경로 오류
```go
// 현재 (잘못됨)
url := fmt.Sprintf("%s/api/workspace/%s", c.baseURL, workspaceID)

// 수정 필요 (User Service API 문서 기준)
url := fmt.Sprintf("%s/api/workspaces/%s", c.baseURL, workspaceID)
```

**영향:**
- User Service API와 불일치
- 404 에러 발생 가능
- workspace 검증 실패

**위치:**
- `board-service/internal/client/user_client.go:212` - CheckWorkspaceExists
- `board-service/internal/client/user_client.go:245` - ValidateWorkspaceMembership
- `board-service/internal/client/user_client.go:278` - GetWorkspace

#### 🔴 문제 2: 매 요청마다 Workspace 멤버십 검증
```go
// CreateProject, GetProjectsByWorkspaceID 등에서 매번 호출
isMember, err := s.userClient.ValidateWorkspaceMembership(ctx, workspaceID, userID, token)
```

**영향:**
- User Service에 불필요한 부하
- API 응답 시간 증가 (네트워크 왕복)
- 동일한 사용자의 반복 요청에도 매번 검증

#### 🔴 문제 3: 사용자 정보 중복 조회
```go
// toProjectResponse에서 매번 User Service 호출
userInfo, err := s.userClient.GetUser(ctx, project.OwnerID.String())
```

**영향:**
- N+1 문제 발생 가능
- 동일 사용자 정보를 여러 번 조회
- API 응답 시간 증가

---

## 🎯 최적화 전략

### 전략 1: API 엔드포인트 수정 (필수)

**작업:**
- `/api/workspace` → `/api/workspaces` 경로 수정
- User Service API 문서와 일치시키기

**우선순위:** 🔥 High (즉시 수정 필요)

**파일:**
- `board-service/internal/client/user_client.go`

---

### 전략 2: Workspace 멤버십 캐싱

**Redis 캐싱 구조:**
```
Key: workspace_member:{workspace_id}:{user_id}
Value: "true" | "false"
TTL: 5분 (300초)
```

**로직:**
1. 요청 들어옴 → Redis 캐시 확인
2. 캐시 있음 → 바로 반환 (User Service 호출 X)
3. 캐시 없음 → User Service 호출 → Redis에 저장

**장점:**
- User Service 호출 95% 이상 감소
- 응답 시간 10배 이상 개선
- User Service 부하 대폭 감소

**우선순위:** 🔥 High

**구현 위치:**
- `board-service/internal/cache/workspace_cache.go` (신규 생성)
- `board-service/internal/service/project_service.go` (수정)

---

### 전략 3: 사용자 정보 캐싱

**Redis 캐싱 구조:**
```
Key: user_info:{user_id}
Value: JSON { "userId": "...", "name": "...", "email": "...", "avatarUrl": "..." }
TTL: 10분 (600초)
```

**로직:**
1. 사용자 정보 필요 → Redis 확인
2. 캐시 있음 → 바로 사용
3. 캐시 없음 → User Service 호출 → Redis 저장

**장점:**
- 중복 조회 제거
- N+1 문제 해결
- 응답 속도 향상

**우선순위:** 🟡 Medium

---

### 전략 4: 배치 API 활용

**현재:**
```go
// 각 project마다 owner 정보 조회 (N번 호출)
for _, proj := range projects {
    userInfo, _ := s.userClient.GetUser(ctx, proj.OwnerID.String())
}
```

**개선:**
```go
// 모든 owner ID 수집 후 한 번에 조회
ownerIDs := extractOwnerIDs(projects)
usersMap, _ := s.userClient.GetUsersBatch(ctx, ownerIDs)
```

**장점:**
- N번 호출 → 1번 호출
- User Service 부하 감소
- 응답 시간 단축

**우선순위:** 🟡 Medium

---

### 전략 5: Workspace 정보 비정규화 (선택적)

**아이디어:**
Project 테이블에 workspace 기본 정보 저장
```go
type Project struct {
    // 기존 필드
    WorkspaceID   uuid.UUID

    // 추가 필드 (비정규화)
    WorkspaceName string  // User Service에서 가져와서 저장
}
```

**장점:**
- Workspace 이름을 위해 User Service 호출 불필요
- JOIN 없이 조회 가능

**단점:**
- 데이터 동기화 문제
- Workspace 이름 변경 시 업데이트 필요

**우선순위:** 🟢 Low (검토 후 결정)

---

## 📝 구현 계획

### Phase 1: 긴급 수정 (즉시)
1. ✅ **API 엔드포인트 수정**
   - `/api/workspace` → `/api/workspaces`
   - 테스트 후 배포

### Phase 2: 캐싱 구현 (1일)
2. ✅ **Workspace 멤버십 캐싱**
   - `WorkspaceCache` 구현
   - `project_service.go` 수정
   - 테스트

3. ✅ **사용자 정보 캐싱**
   - `UserInfoCache` 구현
   - `project_service.go`, `board_service.go` 수정

### Phase 3: 배치 최적화 (1일)
4. ✅ **배치 API 활용**
   - `toProjectResponse` 최적화
   - 멤버 목록 조회 최적화

### Phase 4: 모니터링 및 튜닝
5. ✅ **성능 측정**
   - Before/After 비교
   - 캐시 히트율 모니터링
   - User Service 호출 횟수 추적

---

## 🔢 예상 효과

### Before (현재)
- **GetProjectsByWorkspaceID** (workspace에 10개 프로젝트)
  - User Service 호출: 12회
    - Workspace 멤버십 검증: 1회
    - Owner 정보 조회: 10회
    - Role 정보: 1회
  - 응답 시간: ~500-1000ms

### After (최적화 후)
- **GetProjectsByWorkspaceID** (동일 조건)
  - User Service 호출: 1-2회
    - Workspace 멤버십 검증: 0회 (캐시)
    - Owner 정보 조회: 1회 (배치)
  - 응답 시간: ~50-100ms

**개선율:**
- User Service 호출: 83-92% 감소
- 응답 시간: 80-90% 개선

---

## 🛠️ 구현 상세

### 1. API 엔드포인트 수정

**파일:** `board-service/internal/client/user_client.go`

#### Before:
```go
func (c *userClient) CheckWorkspaceExists(ctx context.Context, workspaceID string, token string) (bool, error) {
    url := fmt.Sprintf("%s/api/workspace/%s", c.baseURL, workspaceID)
    // ...
}

func (c *userClient) ValidateWorkspaceMembership(ctx context.Context, workspaceID string, userID string, token string) (bool, error) {
    url := fmt.Sprintf("%s/api/workspace/%s/members/%s", c.baseURL, workspaceID, userID)
    // ...
}

func (c *userClient) GetWorkspace(ctx context.Context, workspaceID string, token string) (*WorkspaceInfo, error) {
    url := fmt.Sprintf("%s/api/workspace/%s", c.baseURL, workspaceID)
    // ...
}
```

#### After:
```go
func (c *userClient) CheckWorkspaceExists(ctx context.Context, workspaceID string, token string) (bool, error) {
    url := fmt.Sprintf("%s/api/workspaces/%s", c.baseURL, workspaceID)
    // ...
}

func (c *userClient) ValidateWorkspaceMembership(ctx context.Context, workspaceID string, userID string, token string) (bool, error) {
    url := fmt.Sprintf("%s/api/workspaces/%s/members/%s", c.baseURL, workspaceID, userID)
    // ...
}

func (c *userClient) GetWorkspace(ctx context.Context, workspaceID string, token string) (*WorkspaceInfo, error) {
    url := fmt.Sprintf("%s/api/workspaces/%s", c.baseURL, workspaceID)
    // ...
}
```

---

### 2. Workspace 멤버십 캐싱

**신규 파일:** `board-service/internal/cache/workspace_cache.go`

```go
package cache

import (
    "context"
    "fmt"
    "time"

    "github.com/redis/go-redis/v9"
)

type WorkspaceCache interface {
    GetMembership(ctx context.Context, workspaceID, userID string) (bool, bool, error)
    SetMembership(ctx context.Context, workspaceID, userID string, isMember bool) error
    InvalidateMembership(ctx context.Context, workspaceID, userID string) error
}

type workspaceCache struct {
    client *redis.Client
    ttl    time.Duration
}

func NewWorkspaceCache(client *redis.Client) WorkspaceCache {
    return &workspaceCache{
        client: client,
        ttl:    5 * time.Minute, // 5분 TTL
    }
}

func (c *workspaceCache) GetMembership(ctx context.Context, workspaceID, userID string) (bool, bool, error) {
    key := fmt.Sprintf("workspace_member:%s:%s", workspaceID, userID)
    val, err := c.client.Get(ctx, key).Result()

    if err == redis.Nil {
        return false, false, nil // 캐시 없음
    }
    if err != nil {
        return false, false, err
    }

    return true, val == "true", nil // 캐시 있음, 멤버 여부
}

func (c *workspaceCache) SetMembership(ctx context.Context, workspaceID, userID string, isMember bool) error {
    key := fmt.Sprintf("workspace_member:%s:%s", workspaceID, userID)
    val := "false"
    if isMember {
        val = "true"
    }
    return c.client.Set(ctx, key, val, c.ttl).Err()
}

func (c *workspaceCache) InvalidateMembership(ctx context.Context, workspaceID, userID string) error {
    key := fmt.Sprintf("workspace_member:%s:%s", workspaceID, userID)
    return c.client.Del(ctx, key).Err()
}
```

---

### 3. ProjectService 수정

**파일:** `board-service/internal/service/project_service.go`

```go
type projectService struct {
    repo               repository.ProjectRepository
    roleRepo           repository.RoleRepository
    userOrderRepo      repository.UserOrderRepository
    customFieldService CustomFieldService
    userClient         client.UserClient
    workspaceCache     cache.WorkspaceCache  // 추가
    logger             *zap.Logger
    db                 *gorm.DB
}

// CreateProject에서 캐싱 활용
func (s *projectService) CreateProject(userID string, token string, req *dto.CreateProjectRequest) (*dto.ProjectResponse, error) {
    // ... validation ...

    ctx := context.Background()

    // 1. 캐시 확인
    cached, isMember, err := s.workspaceCache.GetMembership(ctx, req.WorkspaceID, userID)
    if err != nil {
        s.logger.Warn("Failed to get workspace membership from cache", zap.Error(err))
    }

    if cached {
        // 캐시에서 멤버십 확인
        if !isMember {
            return nil, apperrors.New(apperrors.ErrCodeWorkspaceAccessDenied, "워크스페이스 멤버가 아닙니다", 403)
        }
    } else {
        // 캐시 없음 → User Service 호출
        isMember, err := s.userClient.ValidateWorkspaceMembership(ctx, req.WorkspaceID, userID, token)
        if err != nil {
            s.logger.Error("Failed to validate workspace membership", zap.Error(err))
            return nil, apperrors.Wrap(err, apperrors.ErrCodeWorkspaceValidationFailed, "워크스페이스 멤버십 확인 실패", 500)
        }

        if !isMember {
            return nil, apperrors.New(apperrors.ErrCodeWorkspaceAccessDenied, "워크스페이스 멤버가 아닙니다", 403)
        }

        // 캐시 저장
        _ = s.workspaceCache.SetMembership(ctx, req.WorkspaceID, userID, true)
    }

    // ... 나머지 로직 ...
}
```

---

## 📊 모니터링 지표

### 추적할 메트릭
1. **User Service 호출 횟수**
   - 엔드포인트별 호출 횟수
   - 시간대별 호출 패턴

2. **캐시 성능**
   - 캐시 히트율 (Hit Rate)
   - 캐시 미스율 (Miss Rate)
   - 평균 응답 시간 (캐시 vs 실제 호출)

3. **API 응답 시간**
   - P50, P95, P99 latency
   - 엔드포인트별 성능 변화

### Prometheus 메트릭 추가
```go
var (
    userServiceCalls = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "board_service_user_service_calls_total",
            Help: "Total number of User Service calls",
        },
        []string{"method", "endpoint"},
    )

    workspaceCacheHits = promauto.NewCounter(
        prometheus.CounterOpts{
            Name: "board_service_workspace_cache_hits_total",
            Help: "Total number of workspace cache hits",
        },
    )

    workspaceCacheMisses = promauto.NewCounter(
        prometheus.CounterOpts{
            Name: "board_service_workspace_cache_misses_total",
            Help: "Total number of workspace cache misses",
        },
    )
)
```

---

## ✅ 체크리스트

### Phase 1: API 엔드포인트 수정
- [ ] `user_client.go` 수정 (`/api/workspace` → `/api/workspaces`)
- [ ] 로컬 테스트
- [ ] 통합 테스트
- [ ] 배포

### Phase 2: 캐싱 구현
- [ ] `workspace_cache.go` 구현
- [ ] `user_info_cache.go` 구현
- [ ] `main.go`에서 캐시 초기화
- [ ] Service 레이어에 캐시 주입
- [ ] 테스트

### Phase 3: 배치 최적화
- [ ] `toProjectResponse` 배치 최적화
- [ ] `GetProjectMembers` 배치 최적화
- [ ] 테스트

### Phase 4: 모니터링
- [ ] Prometheus 메트릭 추가
- [ ] Grafana 대시보드 생성
- [ ] 성능 비교 분석

---

## 📌 참고 자료
- User Service API 문서: `/home/user/wealist-project/USER_SERVICE_API_REFERENCE.md`
- Board Service 코드: `/home/user/wealist-project/board-service/`
- Redis 캐싱 패턴: https://redis.io/docs/manual/patterns/
