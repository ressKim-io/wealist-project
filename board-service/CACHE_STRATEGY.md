# Cache 전략 (Caching Strategy)

**Last Updated**: 2025-11-12

## 📋 개요

Board Service는 Redis 기반 캐싱을 통해 외부 서비스 호출을 최소화하고 응답 속도를 향상시킵니다.

## 🎯 캐싱 대상 및 정책

### 1. User Information Cache (UserInfoCache)

**목적**: User Service 호출 최소화

**캐싱 데이터**:
- `UserInfo`: 상세 사용자 정보 (ID, Name, Email, IsActive)
- `SimpleUser`: 기본 사용자 정보 (ID, Name, AvatarURL)

**TTL**: 10분

**Invalidation 전략**:
- 사용자 정보 변경 시: `InvalidateUser(userID)`
- 자동 만료: 10분 후

**최적화**:
- Batch 조회 지원: `GetSimpleUsersBatch()` - MGET 사용
- Batch 저장 지원: `SetSimpleUsersBatch()` - Pipeline 사용

**Key 패턴**:
```
user_info:{userID}       // 상세 정보
simple_user:{userID}     // 기본 정보
```

**사용 예시**:
```go
// Service Layer에서 사용
func (s *boardService) buildBoardResponse(board *domain.Board) (*dto.BoardResponse, error) {
    // 1. Cache 확인
    cacheExists, cachedUser, _ := s.userInfoCache.GetSimpleUser(ctx, userID)
    if cacheExists && cachedUser != nil {
        return *cachedUser, nil
    }

    // 2. Cache miss - User Service 호출
    user, err := s.userClient.GetSimpleUser(userID)
    if err != nil {
        return cache.SimpleUser{Name: "Unknown User"}, nil
    }

    // 3. Cache 저장
    s.userInfoCache.SetSimpleUser(ctx, &cache.SimpleUser{
        ID:        user.ID,
        Name:      user.Name,
        AvatarURL: user.AvatarURL,
    })

    return *user, nil
}
```

---

### 2. Workspace Membership Cache (WorkspaceCache)

**목적**: Workspace 멤버십 검증 최적화 (User Service 호출 감소)

**캐싱 데이터**:
- Workspace-User 멤버십 검증 결과 (boolean)

**TTL**: 5분

**Invalidation 전략**:
- 멤버 추가/제거 시: `InvalidateMembership(workspaceID, userID)`
- Workspace 삭제 시: `InvalidateWorkspace(workspaceID)` - SCAN 패턴 사용
- 자동 만료: 5분 후 (권한 변경 반영 시간)

**Key 패턴**:
```
workspace_member:{workspaceID}:{userID}   // boolean ("true" or "false")
```

**사용 예시**:
```go
func (s *projectService) CreateProject(userID, token string, req *dto.CreateProjectRequest) error {
    // 1. Cache 확인
    cacheExists, isMember, _ := s.workspaceCache.GetMembership(ctx, workspaceID, userID)
    if cacheExists {
        if !isMember {
            return errors.New("user not a workspace member")
        }
    } else {
        // 2. Cache miss - User Service로 검증
        isMember, err := s.userClient.ValidateWorkspaceMembership(ctx, workspaceID, userID, token)
        if err != nil {
            return err
        }

        // 3. Cache 저장
        s.workspaceCache.SetMembership(ctx, workspaceID, userID, isMember)

        if !isMember {
            return errors.New("user not a workspace member")
        }
    }

    // ... 비즈니스 로직
}
```

---

### 3. Custom Fields Cache (FieldCache)

**목적**: 빈번하게 조회되는 Custom Fields 데이터 캐싱

**캐싱 데이터**:
- Project Fields: 프로젝트의 커스텀 필드 정의
- Field Options: 필드의 선택 옵션 (SingleSelect, MultiSelect)
- Board Field Values: 보드의 필드 값
- View Results: 뷰 필터링 결과 (with filter hash)

**TTL**: Flexible (호출 시 지정)
- Project Fields: 30분 권장
- Field Options: 30분 권장
- Board Field Values: 5분 권장
- View Results: 2분 권장

**Invalidation 전략**:
- Field 생성/수정/삭제 시: `InvalidateProjectFields(projectID)`
- Option 변경 시: `InvalidateFieldOptions(fieldID)`
- Board 필드 값 변경 시: `InvalidateBoardFieldValues(boardID)`
- View 변경 시: `InvalidateViewResults(viewID)` - Tracking Set 사용

**Key 패턴**:
```
project:{projectID}:fields                  // Project fields JSON
field:{fieldID}:options                     // Field options JSON
board:{boardID}:field_values                // Field values map
view:{viewID}:results:{filterHash}          // View results JSON
view:{viewID}:result_keys                   // Tracking set for invalidation
```

**Smart Invalidation (View Results)**:
```go
// SetViewResults: Tracking set에 키 추가
func (c *fieldCache) SetViewResults(viewID, filterHash string, resultsJSON []byte, ttl time.Duration) error {
    resultKey := fmt.Sprintf("view:%s:results:%s", viewID, filterHash)
    trackingSetKey := fmt.Sprintf("view:%s:result_keys", viewID)

    pipe := c.client.Pipeline()
    pipe.Set(ctx, resultKey, resultsJSON, ttl)
    pipe.SAdd(ctx, trackingSetKey, resultKey)  // Track this key
    pipe.Expire(ctx, trackingSetKey, ttl+time.Hour)
    _, err := pipe.Exec(ctx)
    return err
}

// InvalidateViewResults: Tracking set의 모든 키 삭제
func (c *fieldCache) InvalidateViewResults(viewID string) error {
    trackingSetKey := fmt.Sprintf("view:%s:result_keys", viewID)
    keys, _ := c.client.SMembers(ctx, trackingSetKey).Result()
    keysToDelete := append(keys, trackingSetKey)
    return c.client.Del(ctx, keysToDelete...).Err()
}
```

---

## 🔄 Cache Invalidation 전략

### Write-Through Pattern
데이터 변경 시 즉시 캐시 무효화:

```go
// Example: 프로젝트 멤버 추가
func (s *projectService) AddMember(projectID, userID string) error {
    // 1. DB에 저장
    err := s.projectRepo.CreateMember(member)
    if err != nil {
        return err
    }

    // 2. 관련 캐시 무효화
    s.workspaceCache.InvalidateMembership(ctx, project.WorkspaceID, userID)

    return nil
}
```

### TTL-Based Expiration
모든 캐시 항목은 TTL을 가지며, 자동 만료됩니다:

| Cache Type | TTL | 이유 |
|-----------|-----|------|
| UserInfo | 10분 | 사용자 정보 변경 빈도 낮음 |
| Workspace Membership | 5분 | 권한 변경 빠른 반영 필요 |
| Project Fields | 30분 | 필드 정의 변경 드묾 |
| Board Field Values | 5분 | 보드 값 변경 빈번 |
| View Results | 2분 | 필터 결과 변경 빈번 |

---

## ⚡ 성능 최적화

### 1. Batch Operations (N+1 방지)

**Bad**: N+1 쿼리
```go
// ❌ 나쁜 예: 각 사용자마다 개별 조회
for _, board := range boards {
    user, _ := s.userInfoCache.GetSimpleUser(ctx, board.CreatedBy)
    // ...
}
```

**Good**: Batch 조회
```go
// ✅ 좋은 예: 한 번에 모든 사용자 조회
userIDs := extractUserIDs(boards)
userMap := s.getSimpleUsersBatch(ctx, userIDs)  // MGET 사용

for _, board := range boards {
    user := userMap[board.CreatedBy]
    // ...
}
```

### 2. Pipeline 사용

**배치 쓰기 최적화**:
```go
func (c *userInfoCache) SetSimpleUsersBatch(ctx context.Context, simpleUsers []SimpleUser) error {
    pipe := c.client.Pipeline()

    for _, user := range simpleUsers {
        key := c.simpleUserKey(user.ID)
        data, _ := json.Marshal(user)
        pipe.Set(ctx, key, data, c.ttl)
    }

    _, err := pipe.Exec(ctx)  // 한 번에 실행
    return err
}
```

### 3. Cache-Aside Pattern (Lazy Loading)

```go
func (s *boardService) GetBoard(boardID string) (*dto.BoardResponse, error) {
    // 1. Cache 확인
    cached, found := s.cache.Get(boardID)
    if found {
        return cached, nil
    }

    // 2. Cache miss - DB 조회
    board, err := s.repo.FindByID(boardID)
    if err != nil {
        return nil, err
    }

    // 3. Cache 저장
    s.cache.Set(boardID, board, ttl)

    return board, nil
}
```

---

## 🛠️ Cache 에러 처리

**원칙**: Cache 에러는 비즈니스 로직에 영향을 주지 않음

```go
func (s *commentService) getSimpleUserWithCache(ctx context.Context, userID string) cache.SimpleUser {
    // 1. Cache 조회 시도
    cacheExists, cachedUser, err := s.userInfoCache.GetSimpleUser(ctx, userID)
    if err != nil {
        s.logger.Warn("Failed to get simple user from cache", zap.Error(err))
        // ⚠️ 에러 무시하고 계속 진행
    }

    if cacheExists && cachedUser != nil {
        return *cachedUser
    }

    // 2. Cache miss - User Service 호출
    user, err := s.userClient.GetSimpleUser(userID)
    if err != nil {
        s.logger.Error("Failed to get user info", zap.Error(err))
        return cache.SimpleUser{Name: "Unknown User", AvatarURL: ""}
    }

    // 3. Cache 저장 시도
    cacheUser := &cache.SimpleUser{
        ID:        user.ID,
        Name:      user.Name,
        AvatarURL: user.AvatarURL,
    }
    if cacheErr := s.userInfoCache.SetSimpleUser(ctx, cacheUser); cacheErr != nil {
        s.logger.Warn("Failed to cache simple user", zap.Error(cacheErr))
        // ⚠️ 에러 무시 - 비즈니스 로직은 계속
    }

    return *cacheUser
}
```

---

## 📊 모니터링

### Cache Hit Rate 측정 (권장)

```go
// Prometheus 메트릭 추가 권장
var (
    cacheHitTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "cache_hit_total",
            Help: "Total number of cache hits",
        },
        []string{"cache_type"},
    )

    cacheMissTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "cache_miss_total",
            Help: "Total number of cache misses",
        },
        []string{"cache_type"},
    )
)

// Hit Rate = Hits / (Hits + Misses)
```

---

## ✅ Cache 사용 체크리스트

Service Layer에서 Cache를 사용할 때:

- [ ] Cache miss 시 원본 데이터 소스 호출
- [ ] Cache 조회 결과 저장
- [ ] Cache 에러 로깅 (비즈니스 로직 중단 X)
- [ ] Batch 조회 가능 시 Batch API 사용
- [ ] 데이터 변경 시 관련 캐시 무효화
- [ ] 적절한 TTL 설정 (데이터 변경 빈도 고려)

---

## 🔮 향후 개선 방안

### 1. Cache Warming (선택적)
- 서버 시작 시 자주 사용되는 데이터 미리 캐싱
- 예: 기본 Role 데이터, 시스템 필드 정의

### 2. Distributed Cache Invalidation
- Redis Pub/Sub을 통한 다중 인스턴스 캐시 동기화
- 현재는 단일 Redis 인스턴스이므로 불필요

### 3. Cache Metrics Dashboard
- Grafana 대시보드: Hit Rate, Miss Rate, Latency
- Cache 성능 모니터링 및 튜닝

---

**참고**: 현재 구현은 모든 Best Practice를 따르고 있으며, 추가 코드 변경 없이 문서화만으로 충분합니다.
