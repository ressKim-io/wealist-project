# Board Service API Reference

> 분석 날짜: 2025-11-08
> 기술 스택: Go + Gin + GORM + PostgreSQL + Redis

## 📋 목차
1. [개요](#개요)
2. [공통 응답 형식](#공통-응답-형식)
3. [Project API](#project-api)
4. [Custom Fields API](#custom-fields-api)
5. [Board API](#board-api)
6. [Comment API](#comment-api)
7. [User Order API](#user-order-api)
8. [주의사항](#주의사항)

---

## 개요

Board Service는 프로젝트 관리, 보드(칸반 카드), 커스텀 필드, 댓글 관리를 담당하는 마이크로서비스입니다.

### 주요 기능
- 프로젝트 생성 및 관리 (Workspace 기반)
- 커스텀 필드 (Role, Stage, Importance)
- 보드 카드 (Task/Issue) 관리
- 댓글 시스템
- 사용자별 Drag & Drop 순서 저장
- JWT 인증

### 포트
- 기본 포트: 8000

---

## 공통 응답 형식

### 성공 응답
```json
{
  "data": { ... },
  "request_id": "uuid"
}
```

### 에러 응답
```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "에러 메시지"
  },
  "request_id": "uuid"
}
```

---

## Project API

### 1. 프로젝트 생성
```http
POST /api/projects
Authorization: Bearer <token>
Content-Type: application/json

Request:
{
  "workspaceId": "uuid",
  "name": "string",
  "description": "string"
}

Response 201:
{
  "data": {
    "id": "uuid",
    "workspaceId": "uuid",
    "name": "string",
    "description": "string",
    "ownerId": "uuid",
    "ownerName": "string",
    "ownerEmail": "string",
    "createdAt": "timestamp",
    "updatedAt": "timestamp"
  }
}
```

**참고:**
- Workspace 멤버만 프로젝트 생성 가능
- 생성 시 자동으로 기본 커스텀 필드 생성됨
  - Role: "없음"
  - Stage: "없음", "대기", "진행중", "완료"
  - Importance: "없음", "낮음", "보통", "높음", "긴급"

### 2. 프로젝트 조회
```http
GET /api/projects/{id}
Authorization: Bearer <token>

Response 200: ProjectResponse
```

### 3. Workspace의 프로젝트 목록
```http
GET /api/projects?workspace_id={workspaceId}
Authorization: Bearer <token>

Response 200:
{
  "data": {
    "projects": [ProjectResponse, ...]
  }
}
```

### 4. 프로젝트 검색
```http
GET /api/projects/search?workspaceId={id}&query={text}&page=1&limit=10
Authorization: Bearer <token>

Response 200:
{
  "data": {
    "projects": [...],
    "total": 100,
    "page": 1,
    "limit": 10
  }
}
```

### 5. 프로젝트 수정
```http
PUT /api/projects/{id}
Authorization: Bearer <token>
Content-Type: application/json

Request:
{
  "name": "string",
  "description": "string"
}

Response 200: ProjectResponse
```
**권한:** OWNER만

### 6. 프로젝트 삭제
```http
DELETE /api/projects/{id}
Authorization: Bearer <token>

Response 200:
{
  "data": {
    "message": "프로젝트가 삭제되었습니다"
  }
}
```
**권한:** OWNER만
**방식:** Soft Delete

---

## Custom Fields API

### Custom Roles

#### 1. Role 목록 조회
```http
GET /api/custom-fields/projects/{projectId}/roles
Authorization: Bearer <token>

Response 200:
{
  "data": [
    {
      "id": "uuid",
      "projectId": "uuid",
      "name": "string",
      "color": "#RRGGBB",
      "isSystemDefault": false,
      "displayOrder": 0,
      "createdAt": "timestamp",
      "updatedAt": "timestamp"
    }
  ]
}
```

#### 2. Role 생성
```http
POST /api/custom-fields/roles
Authorization: Bearer <token>
Content-Type: application/json

Request:
{
  "projectId": "uuid",
  "name": "string",
  "color": "#RRGGBB"
}

Response 201: CustomRoleResponse
```

#### 3. Role 수정
```http
PUT /api/custom-fields/roles/{id}
Authorization: Bearer <token>
Content-Type: application/json

Request:
{
  "name": "string",
  "color": "#RRGGBB"
}

Response 200: CustomRoleResponse
```

#### 4. Role 삭제
```http
DELETE /api/custom-fields/roles/{id}
Authorization: Bearer <token>

Response 200: { "message": "..." }
```

**참고:** 시스템 기본값(`isSystemDefault: true`)은 삭제 불가

### Custom Stages

동일한 구조 (엔드포인트만 `/stages`로 변경)

### Custom Importance

```http
POST /api/custom-fields/importance
Content-Type: application/json

Request:
{
  "projectId": "uuid",
  "name": "string",
  "color": "#RRGGBB",
  "level": 1-5
}
```

---

## Board API

### ⚠️ 중요: 쿼리 파라미터는 camelCase 사용

### 1. Board 생성
```http
POST /api/boards
Authorization: Bearer <token>
Content-Type: application/json

Request:
{
  "projectId": "uuid",
  "title": "string",              // required, max 200
  "content": "string",             // max 5000
  "roleIds": ["uuid"],             // required, 최소 1개
  "stageId": "uuid",               // required
  "importanceId": "uuid",          // optional
  "assigneeId": "uuid",            // optional
  "dueDate": "ISO 8601 string"     // optional
}

Response 201:
{
  "data": {
    "id": "uuid",
    "projectId": "uuid",
    "title": "string",
    "content": "string",
    "stage": CustomStageResponse,
    "importance": CustomImportanceResponse | null,
    "roles": [CustomRoleResponse],
    "assignee": UserInfo | null,
    "author": UserInfo,
    "dueDate": "timestamp" | null,
    "createdAt": "timestamp",
    "updatedAt": "timestamp"
  }
}
```

**필드 설명:**
- `content`: 설명 필드 (~~description~~이 아님!)
- `roleIds`: 배열 필수 (~~roleId~~가 아님!)

### 2. Board 조회
```http
GET /api/boards/{id}
Authorization: Bearer <token>

Response 200: BoardResponse
```

### 3. Board 목록 조회 (필터링)
```http
GET /api/boards?projectId={id}&stageId={id}&roleId={id}&importanceId={id}&assigneeId={id}&authorId={id}&page=1&limit=20
Authorization: Bearer <token>

Query Parameters:
- projectId: uuid (required)     ← camelCase!
- stageId: uuid (optional)
- roleId: uuid (optional)
- importanceId: uuid (optional)
- assigneeId: uuid (optional)
- authorId: uuid (optional)
- page: int (default: 1)
- limit: int (default: 20, max: 100)

Response 200:
{
  "data": {
    "boards": [BoardResponse, ...],
    "total": 100,
    "page": 1,
    "limit": 20
  }
}
```

**⚠️ 중요:** `project_id`가 아닌 **`projectId`** 사용!

### 4. Board 수정
```http
PUT /api/boards/{id}
Authorization: Bearer <token>
Content-Type: application/json

Request:
{
  "title": "string",
  "content": "string",           ← content!
  "stageId": "uuid",
  "importanceId": "uuid",
  "roleIds": ["uuid"],
  "assigneeId": "uuid",
  "dueDate": "ISO 8601"
}

Response 200: BoardResponse
```

**권한:** 작성자 또는 ADMIN+

### 5. Board 삭제
```http
DELETE /api/boards/{id}
Authorization: Bearer <token>

Response 200:
{
  "data": {
    "message": "보드가 삭제되었습니다"
  }
}
```

**권한:** 작성자 또는 ADMIN+
**방식:** Soft Delete

---

## Comment API

### ⚠️ 중요: 쿼리 파라미터는 camelCase 사용

### 1. Comment 생성
```http
POST /api/comments
Authorization: Bearer <token>
Content-Type: application/json

Request:
{
  "boardId": "uuid",
  "content": "string"
}

Response 201:
{
  "data": {
    "id": "uuid",
    "boardId": "uuid",
    "userId": "uuid",
    "content": "string",
    "author": {
      "id": "uuid",
      "name": "string",
      "avatarUrl": "string"
    },
    "createdAt": "timestamp",
    "updatedAt": "timestamp"
  }
}
```

### 2. Board의 Comment 목록
```http
GET /api/comments?boardId={boardId}
Authorization: Bearer <token>

Query Parameters:
- boardId: uuid (required)     ← camelCase!

Response 200:
{
  "data": [CommentResponse, ...]
}
```

**⚠️ 중요:** `board_id`가 아닌 **`boardId`** 사용!

### 3. Comment 수정
```http
PUT /api/comments/{id}
Authorization: Bearer <token>
Content-Type: application/json

Request:
{
  "content": "string"
}

Response 200: CommentResponse
```

**권한:** 작성자만

### 4. Comment 삭제
```http
DELETE /api/comments/{id}
Authorization: Bearer <token>

Response 204 No Content
```

**⚠️ 중요:** 성공 시 **204 반환** (200이 아님!)
**권한:** 작성자만

---

## User Order API

사용자별 Drag & Drop 순서 저장

### 1. Role 기반 보드 뷰
```http
GET /api/projects/{id}/orders/role-board
Authorization: Bearer <token>

Response 200:
{
  "data": {
    "columnOrder": ["roleId1", "roleId2"],
    "columns": {
      "roleId1": {
        "role": CustomRoleResponse,
        "boards": [BoardResponse, ...]
      }
    }
  }
}
```

### 2. Stage 기반 보드 뷰
```http
GET /api/projects/{id}/orders/stage-board
Authorization: Bearer <token>

Response 200: (구조 동일)
```

### 3. Role Column 순서 업데이트
```http
PUT /api/projects/{id}/orders/role-columns
Authorization: Bearer <token>
Content-Type: application/json

Request:
{
  "columnIds": ["roleId1", "roleId2", ...]
}

Response 200: { "message": "..." }
```

### 4. Stage Column 순서 업데이트
```http
PUT /api/projects/{id}/orders/stage-columns
(동일)
```

### 5. Role 내 Board 순서 업데이트
```http
PUT /api/projects/{id}/orders/role-boards/{roleId}
Authorization: Bearer <token>
Content-Type: application/json

Request:
{
  "boardIds": ["boardId1", "boardId2", ...]
}

Response 200: { "message": "..." }
```

### 6. Stage 내 Board 순서 업데이트
```http
PUT /api/projects/{id}/orders/stage-boards/{stageId}
(동일)
```

---

## 주의사항

### 🔴 Critical

1. **쿼리 파라미터는 camelCase**
   - ✅ `projectId`, `boardId`
   - ❌ `project_id`, `board_id`

2. **Board 필드명**
   - ✅ `content` (설명)
   - ❌ `description`
   - ✅ `roleIds` (배열)
   - ❌ `roleId` (단일값)

3. **Comment DELETE 응답**
   - ✅ 204 No Content
   - ❌ 200 OK

4. **UUID 검증**
   - 모든 ID는 UUID v4 형식
   - 잘못된 형식 시 400 에러

### 🟡 주의

1. **Workspace 검증**
   - 프로젝트 생성 전 User Service에서 Workspace 멤버십 확인
   - 비멤버는 403 에러

2. **기본 커스텀 필드**
   - `isSystemDefault: true`인 필드는 삭제 불가
   - 프로젝트 생성 시 자동 생성됨

3. **Soft Delete**
   - 프로젝트, 보드 삭제는 논리 삭제
   - `is_deleted` 플래그 사용

4. **Author 정보**
   - Board/Comment의 author는 User Service 호출로 채움
   - User Service 실패 시 "Unknown User" 표시

### 📊 페이지네이션

```
기본값:
- page: 1
- limit: 20
- max limit: 100
```

### 🔒 권한 체계

**Project:**
- OWNER: 모든 권한
- ADMIN: 멤버 관리, 설정 변경
- MEMBER: 읽기, 보드 생성

**Board/Comment:**
- 작성자: 수정/삭제
- ADMIN+: 모든 보드/댓글 수정/삭제 가능

---

## 예제 시나리오

### 1. 새 프로젝트에서 Board 생성

```bash
# 1. 프로젝트 생성
POST /api/projects
{
  "workspaceId": "...",
  "name": "My Project"
}
→ projectId 획득

# 2. 커스텀 필드 확인 (자동 생성됨)
GET /api/custom-fields/projects/{projectId}/roles
→ 기본 "없음" Role ID 획득

GET /api/custom-fields/projects/{projectId}/stages
→ 기본 "대기" Stage ID 획득

# 3. Board 생성
POST /api/boards
{
  "projectId": "...",
  "title": "첫 번째 작업",
  "content": "작업 설명",
  "roleIds": ["없음 Role ID"],
  "stageId": "대기 Stage ID"
}
```

### 2. Board 필터링 및 정렬

```bash
# 1. "진행중" 상태의 Board만 조회
GET /api/boards?projectId={id}&stageId={진행중_id}

# 2. 특정 사용자에게 할당된 Board
GET /api/boards?projectId={id}&assigneeId={userId}

# 3. 사용자별 순서로 정렬된 뷰
GET /api/projects/{id}/orders/stage-board
```

---

## 참고 자료
- User Service API: `/home/user/wealist-project/USER_SERVICE_API_REFERENCE.md`
- Board Service 코드: `/home/user/wealist-project/board-service/`
- 테스트 스크립트: `/home/user/wealist-project/board-service/test-board-api.sh`
