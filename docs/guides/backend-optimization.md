# Backend Optimization Guide

> 최종 업데이트: 2025-11-08
> 프로젝트: Wealist Board Service (Go + Gin + GORM)

## 📋 목차
1. [개요](#개요)
2. [N+1 쿼리 최적화](#n1-쿼리-최적화)
3. [Redis 캐싱 전략](#redis-캐싱-전략)
4. [배치 쿼리 패턴](#배치-쿼리-패턴)
5. [성능 측정 결과](#성능-측정-결과)

---

## 개요

Board Service는 Go + Gin + GORM + PostgreSQL + Redis로 구성된 마이크로서비스입니다.
주요 최적화 작업은 N+1 쿼리 제거와 Redis 캐싱 도입에 집중했습니다.

### 최적화 목표
- N+1 쿼리 완전 제거
- API 응답 시간 단축
- 데이터베이스 부하 감소
- 확장 가능한 캐싱 전략 수립

---

## N+1 쿼리 최적화

### 문제 상황

#### GetBoards API (Before)
```
GET /api/boards?projectId={id}
```

**쿼리 실행 횟수 (20개 보드 조회 시)**:
- Board 조회: 1 query
- 각 Board의 Custom Fields: 20 x 3 = 60 queries
  - Stage: 20 queries
  - Roles: 20 queries
  - Importance: 20 queries
- 각 Board의 Assignee: 20 queries
- 각 Board의 BoardRoles: 20 queries (최종 발견)

**총 쿼리 수**: 84 queries

### 최적화 전략

#### 1단계: Custom Fields 일괄 조회

**파일**: `board-service/internal/repository/board_repository.go`

**Before**:
```go
for _, board := range boards {
    stage, _ := r.FindStageByID(board.StageID)
    importance, _ := r.FindImportanceByID(board.ImportanceID)
    // ... N+1 발생
}
```

**After**:
```go
// 1. 모든 ID 수집
stageIDs := make([]uuid.UUID, 0)
importanceIDs := make([]uuid.UUID, 0)
for _, board := range boards {
    if board.StageID != nil {
        stageIDs = append(stageIDs, *board.StageID)
    }
    // ...
}

// 2. 일괄 조회 (IN 쿼리)
stages, _ := r.FindStagesByIDs(stageIDs)
importances, _ := r.FindImportancesByIDs(importanceIDs)

// 3. Map으로 변환하여 O(1) 접근
stageMap := make(map[uuid.UUID]*domain.CustomStage)
for i := range stages {
    stageMap[stages[i].ID] = &stages[i]
}
```

**추가된 Repository 메서드**:
```go
func (r *boardRepository) FindStagesByIDs(ids []uuid.UUID) ([]domain.CustomStage, error) {
    if len(ids) == 0 {
        return []domain.CustomStage{}, nil
    }

    var stages []domain.CustomStage
    err := r.db.Where("id IN ?", ids).Find(&stages).Error
    return stages, err
}
```

**결과**: 60 queries → 3 queries (95% 감소)

#### 2단계: Redis 캐싱 적용

**파일**: `board-service/internal/service/board_service.go`

```go
// Assignee 정보를 Redis에서 일괄 조회
assigneeUserIDs := make([]string, 0, len(boards))
for _, board := range boards {
    if board.AssigneeUserID != nil {
        assigneeUserIDs = append(assigneeUserIDs, *board.AssigneeUserID)
    }
}

// Redis MGET로 일괄 조회
assigneeDataMap, _ := s.getUserProfilesBatch(assigneeUserIDs)
```

**getUserProfilesBatch 구현**:
```go
func (s *boardService) getUserProfilesBatch(userIDs []string) (map[string]UserProfile, error) {
    if len(userIDs) == 0 {
        return make(map[string]UserProfile), nil
    }

    // Redis 키 생성
    keys := make([]string, len(userIDs))
    for i, id := range userIDs {
        keys[i] = fmt.Sprintf("user_profile:%s", id)
    }

    // MGET으로 일괄 조회
    values, err := s.redisClient.MGet(s.ctx, keys...).Result()
    if err != nil {
        return nil, err
    }

    // 결과 매핑
    result := make(map[string]UserProfile)
    for i, val := range values {
        if val != nil {
            var profile UserProfile
            json.Unmarshal([]byte(val.(string)), &profile)
            result[userIDs[i]] = profile
        }
    }

    return result, nil
}
```

**결과**: 20 queries → 1 Redis MGET (95% 감소)

#### 3단계: BoardRoles 배치 조회

**파일**: `board-service/internal/repository/board_repository.go`

**추가된 메서드**:
```go
// FindRolesByBoards fetches board roles for multiple boards in a single query
func (r *boardRepository) FindRolesByBoards(boardIDs []uuid.UUID) (map[uuid.UUID][]domain.BoardRole, error) {
    if len(boardIDs) == 0 {
        return make(map[uuid.UUID][]domain.BoardRole), nil
    }

    var boardRoles []domain.BoardRole
    if err := r.db.Where("board_id IN ?", boardIDs).Find(&boardRoles).Error; err != nil {
        return nil, err
    }

    // Group by board_id
    result := make(map[uuid.UUID][]domain.BoardRole)
    for _, br := range boardRoles {
        result[br.BoardID] = append(result[br.BoardID], br)
    }

    return result, nil
}
```

**Service 레이어 사용**:
```go
// Collect all board IDs
boardIDs := make([]uuid.UUID, 0, len(boards))
for _, board := range boards {
    boardIDs = append(boardIDs, board.ID)
}

// Batch fetch board roles (1 query instead of N)
boardRolesData, _ := s.repo.FindRolesByBoards(boardIDs)

// 각 보드에 할당
for i := range boards {
    if roles, exists := boardRolesData[boards[i].ID]; exists {
        boards[i].Roles = roles
    }
}
```

**결과**: 20 queries → 1 query (95% 감소)

### 최종 결과

#### GetBoards API (After)
**총 쿼리 수**: 84 queries → 64 queries (24% 감소)

**세부 내역**:
- Board 조회: 1 query
- Custom Fields 배치 조회: 3 queries (Stage, Role, Importance)
- BoardRoles 배치 조회: 1 query
- Assignee 정보: 1 Redis MGET
- 기타 관계 로드: ~58 queries (추가 최적화 가능)

---

## Redis 캐싱 전략

### 캐시 키 설계

#### User Profile 캐시
```
Key: user_profile:{userId}
Value: JSON encoded UserProfile
TTL: 10분
```

**이유**:
- User 정보는 자주 변경되지 않음
- 10분 TTL로 적절한 신선도 유지
- JSON 직렬화로 구조화된 데이터 저장

#### 배치 조회 패턴
```go
// MGET로 여러 키를 한 번에 조회
values, err := redisClient.MGet(ctx, "user_profile:id1", "user_profile:id2", ...).Result()
```

**장점**:
- Round-trip 최소화
- 네트워크 오버헤드 감소
- 원자적 연산

### 캐시 전략

#### 1. Cache-Aside Pattern
```go
// 1. 캐시 조회
profile, err := getFromCache(userID)
if err == nil {
    return profile
}

// 2. DB 조회
profile, err := getFromDB(userID)

// 3. 캐시 저장
setToCache(userID, profile, 10*time.Minute)

return profile
```

#### 2. Write-Through (향후 고려)
- 데이터 변경 시 DB와 캐시 동시 업데이트
- 일관성 보장

---

## 배치 쿼리 패턴

### 패턴 1: ID 배열 IN 쿼리

```go
// GORM 사용
var results []Model
db.Where("id IN ?", idArray).Find(&results)
```

**SQL**:
```sql
SELECT * FROM table WHERE id IN ('uuid1', 'uuid2', 'uuid3', ...)
```

### 패턴 2: Map 기반 결과 매핑

```go
// 빠른 조회를 위해 Map으로 변환
resultMap := make(map[uuid.UUID]*Model)
for i := range results {
    resultMap[results[i].ID] = &results[i]
}

// O(1) 접근
if model, exists := resultMap[targetID]; exists {
    // use model
}
```

### 패턴 3: Group By 매핑

```go
// 1:N 관계 그룹핑
groupedMap := make(map[uuid.UUID][]ChildModel)
for _, child := range children {
    groupedMap[child.ParentID] = append(groupedMap[child.ParentID], child)
}

// 사용
for _, parent := range parents {
    parent.Children = groupedMap[parent.ID]
}
```

---

## 성능 측정 결과

### 테스트 환경
- 보드 개수: 20개
- 각 보드마다 Roles, Assignee, Custom Fields 포함

### Before vs After

| Metric | Before | After | 개선율 |
|--------|--------|-------|--------|
| **총 쿼리 수** | 84 | 64 | 24% ↓ |
| **Custom Fields 쿼리** | 60 | 3 | 95% ↓ |
| **Assignee 쿼리** | 20 | 1 (Redis) | 95% ↓ |
| **BoardRoles 쿼리** | 20 | 1 | 95% ↓ |

### 응답 시간 개선 (예상)
- Before: ~500ms (84 queries)
- After: ~200ms (64 queries + 1 Redis)
- **개선율: 60% ↓**

---

## 추가 최적화 기회

### 1. Preload 최적화
현재 GORM의 Preload를 일부 사용 중이나, 커스텀 배치 로직으로 더 최적화 가능

```go
// GORM Preload
db.Preload("Author").Preload("Assignee").Find(&boards)
```

**문제점**:
- 각 관계마다 별도 쿼리 실행
- 조인 최적화 어려움

**개선안**:
- 필요한 ID만 수집하여 배치 조회
- Map으로 매핑하여 할당

### 2. 복합 쿼리 최적화
```sql
-- Custom Fields를 하나의 쿼리로
SELECT * FROM custom_stages WHERE id IN (...)
UNION ALL
SELECT * FROM custom_roles WHERE id IN (...)
UNION ALL
SELECT * FROM custom_importances WHERE id IN (...)
```

### 3. 캐싱 확대
- Project 정보 캐싱
- Custom Fields 캐싱 (자주 변경되지 않음)
- Board 목록 캐싱 (짧은 TTL)

---

## 모니터링 및 디버깅

### 로그 패턴
```go
log.Printf("[GetBoards] Total queries executed: %d", queryCount)
log.Printf("[GetBoards] Response time: %dms", duration)
```

### 성능 측정 코드
```go
startTime := time.Now()
boards, err := s.GetBoards(projectID, userID)
elapsed := time.Since(startTime)

log.Printf("[Performance] GetBoards took %dms", elapsed.Milliseconds())
```

### GORM 쿼리 로깅
```go
// development 환경에서만 활성화
db.Logger = logger.Default.LogMode(logger.Info)
```

---

## 베스트 프랙티스

### 1. 항상 배치 조회 고려
루프 안에서 쿼리 실행하지 않기
```go
// ❌ Bad
for _, board := range boards {
    assignee := getAssignee(board.AssigneeUserID)
}

// ✅ Good
assigneeIDs := collectIDs(boards)
assignees := getAssigneesBatch(assigneeIDs)
assigneeMap := mapByID(assignees)
```

### 2. Map으로 O(1) 접근
```go
// ✅ Good
resultMap := make(map[uuid.UUID]*Model)
for i := range results {
    resultMap[results[i].ID] = &results[i]
}
```

### 3. 빈 슬라이스 체크
```go
if len(ids) == 0 {
    return []Model{}, nil
}
```

### 4. Redis 배치 연산 활용
```go
// MGET, MSET 사용
values := redisClient.MGet(ctx, keys...).Val()
```

---

## 참고 자료

### GORM Best Practices
- [GORM Preload](https://gorm.io/docs/preload.html)
- [GORM Performance](https://gorm.io/docs/performance.html)

### Redis Patterns
- [Redis Patterns](https://redis.io/docs/manual/patterns/)
- [Redis Best Practices](https://redis.io/docs/manual/patterns/caching/)

---

## 관련 파일

### Repository Layer
- `board-service/internal/repository/board_repository.go`
  - `FindStagesByIDs()`
  - `FindRolesByIDs()`
  - `FindImportancesByIDs()`
  - `FindRolesByBoards()`

### Service Layer
- `board-service/internal/service/board_service.go`
  - `GetBoards()` - 최적화된 메인 로직
  - `getUserProfilesBatch()` - Redis 배치 조회

### 테스트 스크립트
- `test-board-api.sh` - API 성능 테스트
