# Frontend API 호출 vs 백엔드 구현 상태 분석

**분석 날짜**: 2025-11-09
**목적**: 프론트엔드에서 호출하는 API들이 백엔드에 실제로 구현되어 있는지 확인하고, 누락된 API 파악

---

## 🔴 **흰 화면 발생 원인**

프론트엔드는 **2개의 백엔드 서비스**를 사용하는 아키텍처이지만, **일부 핵심 API가 구현되지 않아** 초기 로딩 시 실패합니다.

| 서비스 | 포트 | 용도 | 상태 |
|--------|------|------|------|
| User Service (Java/Spring Boot) | 8080 | 사용자, 워크스페이스, 프로필 | ⚠️ **일부 API 누락** |
| Board Service (Go/Gin) | 8000 | 프로젝트, 보드, 커스텀 필드 | ✅ **완전 구현** |

---

## 📊 **전체 구현 현황**

### **User Service API (Java - Port 8080)**

#### ✅ **구현 완료 (6개)**
| Method | Endpoint | 프론트엔드 사용 | 백엔드 구현 |
|--------|----------|----------------|-------------|
| GET | `/api/workspaces` | ✅ | ✅ WorkspaceController:30 |
| POST | `/api/workspaces` | ✅ | ✅ WorkspaceController:43 |
| GET | `/api/profiles/me` | ✅ | ✅ ProfileController:41 |
| PUT | `/api/profiles/me` | ✅ | ✅ ProfileController:78 |
| GET | `/api/workspaces/{workspaceId}/members` | ✅ | ✅ WorkspaceController:138 |
| PUT | `/api/workspaces/{workspaceId}/members/{memberId}/role` | ✅ | ✅ WorkspaceController:153 |

#### ❌ **구현 필요 (10개)**
| Method | Endpoint | 프론트엔드 사용처 | 우선순위 | 비고 |
|--------|----------|-------------------|----------|------|
| GET | `/api/profiles/workspace/{workspaceId}` | userService.ts:414 | 🔴 HIGH | 워크스페이스별 프로필 조회 |
| PUT | `/api/profiles/workspace/{workspaceId}` | userService.ts:458 | 🔴 HIGH | 워크스페이스별 프로필 수정 |
| GET | `/api/workspaces/{workspaceId}/settings` | userService.ts:507 | 🟡 MEDIUM | 워크스페이스 설정 조회 |
| PUT | `/api/workspaces/{workspaceId}/settings` | userService.ts:544 | 🟡 MEDIUM | 워크스페이스 설정 수정 |
| GET | `/api/workspaces/{workspaceId}/pending-members` | userService.ts:610 | 🔴 HIGH | 승인 대기 회원 목록 |
| POST | `/api/workspaces/{workspaceId}/members/{userId}/approve` | userService.ts:639 | 🔴 HIGH | 회원 가입 승인 |
| POST | `/api/workspaces/{workspaceId}/members/{userId}/reject` | userService.ts:684 | 🔴 HIGH | 회원 가입 거절 |
| DELETE | `/api/workspaces/{workspaceId}/members/{userId}` | userService.ts:751 | 🟡 MEDIUM | 회원 퇴출 |
| GET | `/api/workspaces/{workspaceId}/invitable-users` | userService.ts:778 | 🟢 LOW | 초대 가능한 사용자 검색 |
| POST | `/api/workspaces/{workspaceId}/invite/{userId}` | userService.ts:815 | 🟢 LOW | 사용자 초대 |

#### ⚠️ **API 구조 불일치 문제**

**문제 1: 회원 승인 시스템**
- 프론트엔드 기대:
  ```
  GET /api/workspaces/{id}/pending-members        (대기 목록)
  POST /api/workspaces/{id}/members/{userId}/approve  (승인)
  POST /api/workspaces/{id}/members/{userId}/reject   (거절)
  ```
- 백엔드 실제:
  ```
  POST /api/workspaces/join-requests               (가입 신청)
  GET /api/workspaces/{id}/join-requests           (신청 목록)
  PUT /api/workspaces/{id}/joinRequests/{requestId} (승인/거절)
  ```
- **해결 필요**: API 경로 통일 또는 프론트엔드 수정

**문제 2: Path Parameter 불일치**
| 프론트엔드 | 백엔드 | 문제점 |
|-----------|--------|--------|
| `...members/{userId}/role` | `...members/{memberId}/role` | 파라미터 이름 다름 |
| `...members/{userId}` | `...members/{memberId}` | 파라미터 이름 다름 |

---

### **Board Service API (Go - Port 8000)**

#### ✅ **완전 구현 (31개)**

**Projects (7개)**
- ✅ POST `/api/projects` - 프로젝트 생성
- ✅ GET `/api/projects` - 프로젝트 목록
- ✅ GET `/api/projects/search` - 프로젝트 검색
- ✅ GET `/api/projects/{project_id}` - 프로젝트 조회
- ✅ PUT `/api/projects/{project_id}` - 프로젝트 수정
- ✅ DELETE `/api/projects/{project_id}` - 프로젝트 삭제
- ✅ GET `/api/projects/{project_id}/members` - 프로젝트 멤버 조회

**Boards (5개)**
- ✅ POST `/api/boards` - 보드 생성
- ✅ GET `/api/boards` - 보드 목록
- ✅ GET `/api/boards/{board_id}` - 보드 조회
- ✅ PUT `/api/boards/{board_id}` - 보드 수정
- ✅ DELETE `/api/boards/{board_id}` - 보드 삭제

**Custom Fields (15개)**
- ✅ Roles: POST, GET (list/single), PUT, DELETE, PUT order
- ✅ Stages: POST, GET (list/single), PUT, DELETE, PUT order
- ✅ Importance: POST, GET (list/single), PUT, DELETE, PUT order

**Comments (4개)**
- ✅ POST `/api/comments` - 댓글 생성
- ✅ GET `/api/comments` - 댓글 목록
- ✅ PUT `/api/comments/{id}` - 댓글 수정
- ✅ DELETE `/api/comments/{id}` - 댓글 삭제

---

## 🚨 **흰 화면 발생 시나리오**

### **시나리오 1: 워크스페이스 페이지 로딩 실패**
```
1. 사용자가 OAuth 로그인 후 /workspaces 접속
2. 프론트엔드가 getWorkspaces() 호출
   → GET http://localhost:8080/api/workspaces
3. ✅ API 정상 구현 (WorkspaceController:30)
4. 워크스페이스 목록 표시 성공
```
**결과**: ✅ 정상 작동 예상

---

### **시나리오 2: 대시보드 로딩 실패**
```
1. 사용자가 워크스페이스 선택 후 /kanban/{workspaceId} 접속
2. 프론트엔드가 3개 API 동시 호출:

   [API 1] getProjects(workspace_id)
   → GET http://localhost:8000/api/projects?workspace_id=xxx
   ✅ 성공 (board-service)

   [API 2] getWorkspaceMembers(workspaceId)
   → GET http://localhost:8080/api/workspaces/{id}/members
   ✅ 성공 (user-service)

   [API 3] fetchBoards() → getProjectStages(project_id)
   → GET http://localhost:8000/api/custom-fields/projects/{id}/stages
   ✅ 성공 (board-service)

3. 모든 API 성공 → 대시보드 표시
```
**결과**: ✅ 정상 작동 예상

---

### **시나리오 3: 워크스페이스 설정 페이지 실패**
```
1. 사용자가 워크스페이스 설정 버튼 클릭
2. 프론트엔드가 getWorkspaceSettings() 호출
   → GET http://localhost:8080/api/workspaces/{id}/settings
3. ❌ 404 Not Found (API 미구현)
4. 오류 발생 → 흰 화면 또는 오류 메시지
```
**결과**: ❌ **실패 - API 미구현**

---

### **시나리오 4: 회원 승인 기능 실패**
```
1. OWNER가 승인 대기 회원 보려고 시도
2. 프론트엔드가 getPendingMembers() 호출
   → GET http://localhost:8080/api/workspaces/{id}/pending-members
3. ❌ 404 Not Found (API 미구현)
4. 승인 대기 목록이 비어있거나 오류 발생
```
**결과**: ❌ **실패 - API 미구현**

---

## 📋 **구현 우선순위별 API 목록**

### 🔴 **Priority 1: CRITICAL (앱 핵심 기능)**
즉시 구현하지 않으면 워크스페이스 관리 기능 사용 불가

1. **GET** `/api/workspaces/{workspaceId}/pending-members`
   - **목적**: 승인 대기 회원 목록 조회
   - **대체 방안**: 기존 `/api/workspaces/{id}/join-requests`를 프론트에 맞게 수정
   - **프론트 사용**: Dashboard.tsx, WorkspaceSettingsModal.tsx

2. **POST** `/api/workspaces/{workspaceId}/members/{userId}/approve`
   - **목적**: 회원 가입 승인
   - **대체 방안**: 기존 `PUT /api/workspaces/{id}/joinRequests/{requestId}`를 사용하도록 프론트 수정
   - **프론트 사용**: WorkspaceMemberManagement.tsx

3. **POST** `/api/workspaces/{workspaceId}/members/{userId}/reject`
   - **목적**: 회원 가입 거절
   - **프론트 사용**: WorkspaceMemberManagement.tsx

4. **GET** `/api/profiles/workspace/{workspaceId}`
   - **목적**: 워크스페이스별 사용자 프로필 조회
   - **비고**: 현재는 전역 프로필만 존재 (`/api/profiles/me`)
   - **프론트 사용**: Dashboard.tsx, ProfileModal.tsx

5. **PUT** `/api/profiles/workspace/{workspaceId}`
   - **목적**: 워크스페이스별 사용자 프로필 수정
   - **프론트 사용**: ProfileModal.tsx

---

### 🟡 **Priority 2: IMPORTANT (주요 기능)**
사용자 경험 향상을 위해 필요

6. **GET** `/api/workspaces/{workspaceId}/settings`
   - **목적**: 워크스페이스 설정 조회
   - **필드**: isPublic, requiresApproval, onlyOwnerCanInvite
   - **프론트 사용**: WorkspaceSettingsModal.tsx

7. **PUT** `/api/workspaces/{workspaceId}/settings`
   - **목적**: 워크스페이스 설정 수정
   - **프론트 사용**: WorkspaceSettingsModal.tsx

8. **DELETE** `/api/workspaces/{workspaceId}/members/{userId}`
   - **목적**: 회원 퇴출
   - **현재 상태**: 비슷한 API 있음 (/{memberId})
   - **필요 작업**: 파라미터 이름 통일

---

### 🟢 **Priority 3: NICE-TO-HAVE (편의 기능)**
없어도 핵심 기능 사용 가능

9. **GET** `/api/workspaces/{workspaceId}/invitable-users?query={query}`
   - **목적**: 초대 가능한 사용자 검색
   - **프론트 사용**: InviteMemberModal.tsx

10. **POST** `/api/workspaces/{workspaceId}/invite/{userId}`
    - **목적**: 사용자를 워크스페이스에 초대
    - **프론트 사용**: InviteMemberModal.tsx

---

## 🛠️ **권장 해결 방안**

### **방안 1: 백엔드 API 추가 구현 (권장)**
**장점**:
- 프론트엔드 코드 수정 불필요
- 설계된 대로 기능 구현

**작업량**:
```java
// 1. WorkspaceController.java에 추가
@GetMapping("/{workspaceId}/pending-members")
@GetMapping("/{workspaceId}/settings")
@PutMapping("/{workspaceId}/settings")
@PostMapping("/{workspaceId}/members/{userId}/approve")
@PostMapping("/{workspaceId}/members/{userId}/reject")
@DeleteMapping("/{workspaceId}/members/{userId}")
@GetMapping("/{workspaceId}/invitable-users")
@PostMapping("/{workspaceId}/invite/{userId}")

// 2. ProfileController.java에 추가
@GetMapping("/workspace/{workspaceId}")
@PutMapping("/workspace/{workspaceId}")
```

---

### **방안 2: 프론트엔드 코드 수정**
**장점**:
- 기존 백엔드 API 활용
- 빠른 임시 해결

**작업량**:
```typescript
// userService.ts 수정
// 기존 join-requests API 사용하도록 변경
export const getPendingMembers = async (workspaceId, token) => {
  // GET /api/workspaces/{id}/join-requests?status=PENDING 사용
}

export const approveMember = async (workspaceId, requestId, token) => {
  // PUT /api/workspaces/{id}/joinRequests/{requestId} 사용
  // Body: { status: "APPROVED" }
}

// 워크스페이스별 프로필 기능 제거 또는 전역 프로필로 대체
```

---

### **방안 3: Hybrid (단계적 구현)**
1. **즉시**: 프론트엔드에서 기존 API 사용하도록 임시 수정
2. **1주 내**: Priority 1 API 구현 (회원 승인, 프로필)
3. **2주 내**: Priority 2 API 구현 (설정)
4. **장기**: Priority 3 API 구현 (초대 기능)

---

## 📝 **API 구현 체크리스트**

### **User Service (Java) - 구현 필요**

#### **ProfileController.java**
- [ ] `GET /api/profiles/workspace/{workspaceId}` - 워크스페이스별 프로필 조회
- [ ] `PUT /api/profiles/workspace/{workspaceId}` - 워크스페이스별 프로필 수정

#### **WorkspaceController.java**
- [ ] `GET /api/workspaces/{workspaceId}/settings` - 워크스페이스 설정 조회
- [ ] `PUT /api/workspaces/{workspaceId}/settings` - 워크스페이스 설정 수정
- [ ] `GET /api/workspaces/{workspaceId}/pending-members` - 승인 대기 회원 조회
  - 또는 기존 `GET .../join-requests?status=PENDING` 사용하도록 프론트 수정
- [ ] `POST /api/workspaces/{workspaceId}/members/{userId}/approve` - 회원 승인
  - 또는 기존 `PUT .../joinRequests/{requestId}` 사용하도록 프론트 수정
- [ ] `POST /api/workspaces/{workspaceId}/members/{userId}/reject` - 회원 거절
- [ ] `DELETE /api/workspaces/{workspaceId}/members/{userId}` - 회원 퇴출
  - 기존: `DELETE .../members/{memberId}` (파라미터 이름만 다름)
- [ ] `GET /api/workspaces/{workspaceId}/invitable-users` - 초대 가능 사용자 검색
- [ ] `POST /api/workspaces/{workspaceId}/invite/{userId}` - 사용자 초대

---

## 🔍 **디버깅 가이드**

### **흰 화면 발생 시 확인 사항**

1. **브라우저 콘솔 확인**
   ```javascript
   // F12 → Console 탭
   // 404, 500 에러 확인
   ```

2. **Network 탭 확인**
   ```
   F12 → Network 탭
   - 실패한 API 요청 확인
   - Response 내용 확인
   ```

3. **Backend 서버 상태 확인**
   ```bash
   # User Service (Java)
   curl http://localhost:8080/actuator/health

   # Board Service (Go)
   curl http://localhost:8000/health
   ```

4. **CORS 오류 확인**
   ```
   Access-Control-Allow-Origin 오류 확인
   → backend CORS 설정 확인
   ```

5. **Token 확인**
   ```javascript
   // 브라우저 콘솔에서
   localStorage.getItem('access_token')
   // null이면 로그인 필요
   ```

---

## 📌 **결론**

### **현재 상태**:
- Board Service (Go): ✅ **100% 구현 완료**
- User Service (Java): ⚠️ **60% 구현 완료 (10개 API 누락)**

### **흰 화면 원인**:
핵심 기능인 **회원 승인 시스템 API**와 **워크스페이스별 프로필 API**가 누락되어, 해당 기능 사용 시 오류 발생

### **해결 방안**:
1. **임시**: 프론트엔드에서 기존 join-requests API 사용하도록 수정
2. **장기**: 위 체크리스트의 10개 API를 User Service에 구현

---

**작성자**: Claude Code
**참고 파일**:
- `/home/user/wealist-project/frontend/src/api/user/userService.ts`
- `/home/user/wealist-project/frontend/src/api/board/boardService.ts`
- `/home/user/wealist-project/user-service/src/main/java/OrangeCloud/UserRepo/controller/WorkspaceController.java`
- `/home/user/wealist-project/user-service/src/main/java/OrangeCloud/UserRepo/controller/ProfileController.java`
- `/home/user/wealist-project/board-service/cmd/api/main.go`
