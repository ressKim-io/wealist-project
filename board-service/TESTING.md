# Testing Guide - Fractional Indexing

이 문서는 fractional indexing 구현을 테스트하는 방법을 설명합니다.

## 📋 목차

1. [사전 준비](#사전-준비)
2. [기본 테스트](#기본-테스트)
3. [성능 테스트](#성능-테스트)
4. [수동 테스트](#수동-테스트)

---

## 사전 준비

### 1. 서비스 실행

Board Service와 User Service가 모두 실행되어 있어야 합니다:

```bash
# User Service (port 8080)
cd user-service
./gradlew bootRun

# Board Service (port 8000)
cd board-service
go run cmd/api/main.go
```

### 2. 테스트 토큰 확인

User Service의 `/api/auth/test` 엔드포인트가 정상 동작하는지 확인합니다:

```bash
# 테스트 토큰 발급 확인
curl -s http://localhost:8080/api/auth/test | jq '.'

# 응답 예시:
# {
#   "accessToken": "eyJhbGc...",
#   "userId": "test-user-id",
#   "email": "test@example.com"
# }
```

**참고**: 테스트 환경에서는 `/api/auth/test`를 사용하므로 별도의 유저 등록이 필요 없습니다.

---

## 기본 테스트

### Fractional Indexing 통합 테스트

전체 fractional indexing 기능을 테스트하는 스크립트를 실행합니다:

```bash
cd /home/user/wealist-project/board-service
./test-fractional-indexing.sh
```

이 스크립트는 다음을 테스트합니다:

#### ✅ Test 1: 같은 컬럼 내에서 보드 이동
- Board-2를 Board-1과 Board-3 사이로 이동
- 새 position이 두 보드 사이에 올바르게 생성되는지 확인

#### ✅ Test 2: 다른 컬럼으로 보드 이동
- Board-1을 "Todo" → "In Progress"로 이동
- Custom field 값이 올바르게 업데이트되는지 확인

#### ✅ Test 3: 첫 번째 위치로 이동
- Board-5를 맨 앞으로 이동
- `before_position: null, after_position: <first>`

#### ✅ Test 4: 마지막 위치로 이동
- Board-4를 맨 뒤로 이동
- `before_position: <last>, after_position: null`

#### ✅ Test 5: 정렬 검증
- 모든 position이 사전순(lexicographic)으로 정렬되어 있는지 확인

#### ✅ Test 6: 성능 테스트
- 10개 보드를 생성하고 중간 위치로 이동
- 응답 시간 측정 (1개 row만 업데이트되므로 빠름)

---

## 성능 테스트

### Integer vs Fractional Indexing 비교

#### 이전 방식 (Integer-based DisplayOrder)

```
시나리오: 100개 보드 중 50번째 위치에 새 보드 삽입

동작:
1. 새 보드를 position 50에 삽입
2. 기존 position 50-99의 모든 보드를 +1씩 증가
3. 총 51개 row UPDATE 쿼리 실행 (O(N))

결과:
- DB 쿼리: 51개
- 소요 시간: ~200-500ms
- DB 부하: 높음 (N이 커질수록 악화)
```

#### 현재 방식 (Fractional Indexing)

```
시나리오: 100개 보드 중 50번째 위치에 새 보드 삽입

동작:
1. 49번째 보드의 position: "a49"
2. 50번째 보드의 position: "a50"
3. 새 position 생성: "a49V" (a49와 a50 사이)
4. 새 보드만 INSERT 또는 UPDATE (O(1))

결과:
- DB 쿼리: 1개
- 소요 시간: ~10-50ms
- DB 부하: 낮음 (N과 무관)
```

### 실제 성능 측정

테스트 스크립트의 Test 6를 실행하면 실제 응답 시간을 측정할 수 있습니다:

```bash
./test-fractional-indexing.sh
```

출력 예시:
```
========================================
TEST 6: Performance Test - Verify Single Row Update
========================================

>>> Creating 10 additional boards for performance test
✓ Created 10 additional boards

>>> Getting positions in Done column
ℹ Moving board to position 2 (between first and second board)
ℹ Before position: a0
ℹ After position: a1

✓ Move completed in 45ms
✓ New position: a0V

ℹ With fractional indexing, this operation updates only 1 row
ℹ Without it (integer-based), this would update N rows (cascading updates)
```

---

## 수동 테스트

### 1. API 직접 호출하기

#### Step 1: 토큰 얻기

**테스트 환경**에서는 `/api/auth/test` 엔드포인트를 사용합니다:

```bash
export TOKEN=$(curl -s http://localhost:8080/api/auth/test \
  | jq -r '.accessToken')

echo $TOKEN
```

**프로덕션 환경**에서는 실제 로그인을 사용합니다:

```bash
export TOKEN=$(curl -s -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "your-username", "password": "your-password"}' \
  | jq -r '.accessToken')

echo $TOKEN
```

#### Step 2: 워크스페이스 생성 (User Service)

먼저 User Service에서 워크스페이스를 생성합니다:

```bash
WORKSPACE_ID=$(curl -s -X POST http://localhost:8080/api/workspaces \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Workspace",
    "description": "Workspace for testing"
  }' | jq -r '.id')

echo $WORKSPACE_ID
```

#### Step 3: 프로젝트 생성 (Board Service)

워크스페이스 ID를 사용하여 프로젝트를 생성합니다:

```bash
PROJECT_ID=$(curl -s -X POST http://localhost:8000/api/projects \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "workspace_id": "'$WORKSPACE_ID'",
    "name": "Test Project",
    "description": "Testing fractional indexing"
  }' | jq -r '.data.project_id')

echo $PROJECT_ID
```

#### Step 4: Custom Field 생성

```bash
FIELD_ID=$(curl -s -X POST http://localhost:8000/api/custom-fields \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "project_id": "'$PROJECT_ID'",
    "name": "Status",
    "field_type": "single_select"
  }' | jq -r '.data.field_id')

echo $FIELD_ID
```

#### Step 5: Field Options 생성

```bash
# Todo
TODO_ID=$(curl -s -X POST http://localhost:8000/api/custom-fields/$FIELD_ID/options \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"value": "Todo", "color": "#FF0000"}' \
  | jq -r '.data.option_id')

# In Progress
PROGRESS_ID=$(curl -s -X POST http://localhost:8000/api/custom-fields/$FIELD_ID/options \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"value": "In Progress", "color": "#FFA500"}' \
  | jq -r '.data.option_id')

echo "Todo: $TODO_ID"
echo "In Progress: $PROGRESS_ID"
```

#### Step 6: Saved View 생성

```bash
VIEW_ID=$(curl -s -X POST http://localhost:8000/api/views \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "project_id": "'$PROJECT_ID'",
    "name": "Status Board",
    "group_by_field_id": "'$FIELD_ID'"
  }' | jq -r '.data.view_id')

echo $VIEW_ID
```

#### Step 7: 보드 생성

```bash
BOARD_1=$(curl -s -X POST http://localhost:8000/api/boards \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "project_id": "'$PROJECT_ID'",
    "title": "Task 1",
    "description": "First task"
  }' | jq -r '.data.board_id')

BOARD_2=$(curl -s -X POST http://localhost:8000/api/boards \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "project_id": "'$PROJECT_ID'",
    "title": "Task 2",
    "description": "Second task"
  }' | jq -r '.data.board_id')

BOARD_3=$(curl -s -X POST http://localhost:8000/api/boards \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "project_id": "'$PROJECT_ID'",
    "title": "Task 3",
    "description": "Third task"
  }' | jq -r '.data.board_id')

echo "Board 1: $BOARD_1"
echo "Board 2: $BOARD_2"
echo "Board 3: $BOARD_3"
```

#### Step 8: Custom Field 값 설정

```bash
# 모든 보드를 Todo로 설정
for BOARD in $BOARD_1 $BOARD_2 $BOARD_3; do
  curl -s -X PUT http://localhost:8000/api/boards/$BOARD/field-values \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
      "field_values": {
        "'$FIELD_ID'": "'$TODO_ID'"
      }
    }' | jq '.'
done
```

#### Step 9: 보드 순서 확인

```bash
curl -s -X GET "http://localhost:8000/api/views/$VIEW_ID/boards" \
  -H "Authorization: Bearer $TOKEN" \
  | jq -r '.data[] | "\(.title): position=\(.position // "none")"'
```

출력 예시:
```
Task 1: position=a0
Task 2: position=a1
Task 3: position=a2
```

#### Step 10: 보드 이동 테스트

**Task 2를 맨 앞으로 이동** (before Task 1):

```bash
curl -s -X POST http://localhost:8000/api/boards/$BOARD_2/move \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "view_id": "'$VIEW_ID'",
    "group_by_field_id": "'$FIELD_ID'",
    "new_field_value": "'$TODO_ID'",
    "before_position": null,
    "after_position": "a0"
  }' | jq '.'
```

응답:
```json
{
  "success": true,
  "data": {
    "board_id": "...",
    "new_field_value": "...",
    "new_position": "Zz",  // "a0"보다 작은 값
    "message": "Board moved successfully"
  }
}
```

**Task 1을 In Progress로 이동**:

```bash
curl -s -X POST http://localhost:8000/api/boards/$BOARD_1/move \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "view_id": "'$VIEW_ID'",
    "group_by_field_id": "'$FIELD_ID'",
    "new_field_value": "'$PROGRESS_ID'",
    "before_position": null,
    "after_position": null
  }' | jq '.'
```

#### Step 11: 최종 순서 확인

```bash
curl -s -X GET "http://localhost:8000/api/views/$VIEW_ID/boards" \
  -H "Authorization: Bearer $TOKEN" \
  | jq -r '.data[] | "\(.title): column=\(.custom_fields["'$FIELD_ID'"] // "none"), position=\(.position // "none")"'
```

---

## PostgreSQL에서 직접 확인

### 1. DB 접속

```bash
psql -U postgres -d board_service
```

### 2. user_board_order 테이블 확인

```sql
-- 모든 board order 조회
SELECT
    board_id,
    position,
    view_id,
    created_at
FROM user_board_order
ORDER BY position ASC;

-- 특정 view의 board order 조회
SELECT
    ubo.board_id,
    b.title,
    ubo.position
FROM user_board_order ubo
JOIN boards b ON ubo.board_id = b.id
WHERE ubo.view_id = '<YOUR_VIEW_ID>'
ORDER BY ubo.position ASC;

-- Position 분포 확인
SELECT
    position,
    COUNT(*) as count
FROM user_board_order
GROUP BY position
ORDER BY position;
```

### 3. 마이그레이션 확인

```sql
-- display_order 컬럼이 삭제되었는지 확인
\d user_board_order

-- 결과:
-- ✅ position 컬럼이 VARCHAR(255)로 존재
-- ✅ display_order 컬럼이 없음
```

---

## 트러블슈팅

### 문제 1: "column display_order does not exist"

**원인**: 마이그레이션이 아직 실행되지 않음

**해결**:
```bash
cd /home/user/wealist-project/board-service
./scripts/db/apply_migrations.sh dev
```

### 문제 2: 토큰 인증 실패

**원인**: User Service가 실행되지 않았거나 토큰이 만료됨

**해결**:
```bash
# User Service 실행 확인
curl http://localhost:8080/health

# 새 토큰 발급 (테스트 환경)
export TOKEN=$(curl -s http://localhost:8080/api/auth/test \
  | jq -r '.accessToken')
```

### 문제 3: Position이 null로 표시됨

**원인**: SavedView를 통해 MoveBoard를 한 번도 호출하지 않음

**설명**:
- 보드를 생성하면 custom field 값은 설정되지만 position은 null
- SavedView에서 MoveBoard API를 호출해야 position이 생성됨
- 이후부터는 position 기반으로 정렬됨

**해결**: MoveBoard API를 호출하여 초기 position 생성

---

## 성공 기준

다음 조건이 모두 만족되면 테스트 성공:

✅ **기능적 요구사항**
- [ ] 같은 컬럼 내 보드 이동 성공
- [ ] 다른 컬럼으로 보드 이동 성공
- [ ] 첫 번째/마지막 위치로 이동 성공
- [ ] Position이 사전순으로 정렬됨

✅ **성능 요구사항**
- [ ] MoveBoard API 응답 시간 < 100ms
- [ ] DB에 1개 row만 업데이트됨 (N개 아님)
- [ ] 100개 보드가 있어도 성능 저하 없음

✅ **데이터 무결성**
- [ ] Position 값이 중복되지 않음
- [ ] Custom field 값이 올바르게 업데이트됨
- [ ] 다른 보드의 position은 변경되지 않음

---

## 추가 리소스

- **Frontend API Guide**: `docs/FRONTEND_API_GUIDE.md`
- **Migration Files**: `migrations/20250111120000_convert_to_fractional_indexing.{up,down}.sql`
- **Utility Code**: `internal/util/position.go`
- **Figma Blog Post**: https://www.figma.com/blog/realtime-editing-of-ordered-sequences/

---

## 문의

테스트 중 문제가 발생하면 다음을 확인하세요:

1. Board Service와 User Service가 모두 실행 중인가?
2. 마이그레이션이 적용되었는가?
3. 테스트 유저가 생성되었는가?
4. 토큰이 유효한가?

모든 것이 정상이라면 로그를 확인하세요:
```bash
# Board Service 로그
tail -f /var/log/board-service.log

# PostgreSQL 쿼리 로그
tail -f /var/log/postgresql/postgresql-*.log
```
