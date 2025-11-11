# User Service API Reference

> 분석 날짜: 2025-11-07
> 브랜치: main-user
> 기술 스택: Java Spring Boot + PostgreSQL + Redis

## 목차
1. [개요](#개요)
2. [인증](#인증)
3. [엔티티 구조](#엔티티-구조)
4. [API 엔드포인트](#api-엔드포인트)

---

## 개요

User Service는 사용자 관리, 인증, 프로필, 워크스페이스 관리를 담당하는 마이크로서비스입니다.

### 주요 기능
- Google OAuth2 기반 소셜 로그인
- JWT 토큰 기반 인증 (Access Token + Refresh Token)
- 사용자 프로필 관리
- 워크스페이스(그룹) 관리
- 워크스페이스 멤버 및 역할 관리
- 가입 신청 관리

### 포트
- 기본 포트: 8080 (추정)

---

## 인증

### 인증 방식
- **Bearer Token (JWT)**
- Header: `Authorization: Bearer <access_token>`

### 토큰 종류
1. **Access Token**: API 요청 시 사용 (짧은 만료 시간)
2. **Refresh Token**: Access Token 갱신용 (긴 만료 시간)

### 인증 흐름
1. Google OAuth2 로그인 → `/oauth2/authorization/google`
2. 성공 시 Access Token + Refresh Token 발급
3. API 요청 시 Access Token 사용
4. Access Token 만료 시 Refresh Token으로 갱신 → `POST /api/auth/refresh`

---

## 엔티티 구조

### User (사용자)
```java
{
  "user_id": "UUID",              // 사용자 ID (PK)
  "email": "string",             // 이메일 (unique)
  "provider": "google",          // OAuth 제공자
  "google_id": "string",          // Google ID (unique)
  "createdAt": "timestamp",      // 생성 시간
  "updatedAt": "timestamp",      // 수정 시간
  "isActive": "boolean",         // 활성화 상태 (기본: true)
  "deletedAt": "timestamp"       // 소프트 삭제 시간
}
```

### UserProfile (사용자 프로필)
```java
{
  "profile_id": "UUID",           // 프로필 ID (PK)
  "user_id": "UUID",              // 사용자 ID (FK, unique)
  "name": "string",              // 이름 (최대 50자)
  "profileImageUrl": "string",   // 프로필 이미지 URL (nullable)
  "createdAt": "timestamp",      // 생성 시간
  "updatedAt": "timestamp"       // 수정 시간
}
```

### Workspace (워크스페이스)
테이블명: `groups`
```java
{
  "group_id": "UUID",             // 워크스페이스 ID (PK)
  "name": "string",              // 워크스페이스 이름
  "companyName": "string",       // 설명 (description 역할)
  "createdAt": "timestamp",      // 생성 시간
  "updatedAt": "timestamp",      // 수정 시간
  "isActive": "boolean",         // 활성화 상태
  "deletedAt": "timestamp"       // 소프트 삭제 시간
}
```

### WorkspaceMember (워크스페이스 멤버)
```java
{
  "memberId": "UUID",            // 멤버 ID (PK)
  "workspace_id": "UUID",         // 워크스페이스 ID (FK)
  "user_id": "UUID",              // 사용자 ID (FK)
  "role": "string",              // 역할: OWNER, ADMIN, MEMBER
  "joinedAt": "timestamp"        // 가입 시간
}
```

### WorkspaceJoinRequest (가입 신청)
```java
{
  "requestId": "UUID",           // 신청 ID (PK)
  "workspace_id": "UUID",         // 워크스페이스 ID (FK)
  "user_id": "UUID",              // 신청자 ID (FK)
  "status": "string",            // 상태: PENDING, APPROVED, REJECTED
  "createdAt": "timestamp",      // 신청 시간
  "processedAt": "timestamp"     // 처리 시간
}
```

---

## API 엔드포인트

### 1. 인증 API (`/api/auth`)

#### 1.1 로그아웃
```http
POST /api/auth/logout
Authorization: Bearer <access_token>

Response 200:
{
  "success": true,
  "message": "로그아웃 성공"
}
```

#### 1.2 토큰 갱신
```http
POST /api/auth/refresh
Content-Type: application/json

Request:
{
  "refreshToken": "string"
}

Response 200:
{
  "accessToken": "string",
  "refreshToken": "string",
  "user_id": "UUID",
  "name": "string",
  "email": "string",
  "tokenType": "Bearer"
}
```

#### 1.3 내 정보 조회
```http
GET /api/auth/me
Authorization: Bearer <access_token>

Response 200: User 엔티티
```

#### 1.4 테스트 로그인
```http
GET /api/auth/test

Response 200: AuthResponse (테스트용 토큰 발급)
```

---

### 2. 사용자 API (`/api/users`)

#### 2.1 내 정보 조회
```http
GET /api/users/me
Authorization: Bearer <access_token>

Response 200: User 엔티티
```

#### 2.2 특정 사용자 정보 조회
```http
GET /api/users/{userId}
Authorization: Bearer <access_token>

Response 200: User 엔티티
```

#### 2.3 내 계정 삭제 (소프트 삭제)
```http
DELETE /api/users/me
Authorization: Bearer <access_token>

Response 200:
{
  "success": true,
  "message": "계정이 삭제되었습니다."
}
```

#### 2.4 사용자 삭제 (관리자용)
```http
DELETE /api/users/{userId}
Authorization: Bearer <access_token>

Response 200:
{
  "success": true,
  "message": "사용자가 삭제되었습니다."
}
```

#### 2.5 사용자 복구
```http
PUT /api/users/{userId}/restore
Authorization: Bearer <access_token>

Response 200:
{
  "success": true,
  "message": "사용자가 복구되었습니다."
}
```

---

### 3. 프로필 API (`/api/profiles`)

#### 3.1 내 프로필 조회
```http
GET /api/profiles/me
Authorization: Bearer <access_token>

Response 200:
{
  "profile_id": "UUID",
  "user_id": "UUID",
  "name": "string",
  "profileImageUrl": "string",
  "createdAt": "timestamp",
  "updatedAt": "timestamp"
}
```

#### 3.2 프로필 정보 통합 업데이트
```http
PUT /api/profiles/me
Authorization: Bearer <access_token>
Content-Type: application/json

Request:
{
  "name": "string",              // 선택적
  "profileImageUrl": "string"    // 선택적
}

Response 200: UserProfileResponse
```

#### 3.3 프로필 사진만 업데이트
```http
PUT /api/profiles/me/image
Authorization: Bearer <access_token>
Content-Type: application/json

Request:
{
  "profileImageUrl": "string"
}

Response 200: UserProfileResponse
```

---

### 4. 워크스페이스 API (`/api/workspaces`)

#### 4.1 워크스페이스 목록 조회
```http
GET /api/workspaces
Authorization: Bearer <access_token>

Response 200: WorkspaceResponse[]
```

#### 4.2 워크스페이스 생성
```http
POST /api/workspaces
Authorization: Bearer <access_token>
Content-Type: application/json

Request:
{
  "name": "string",
  "description": "string"
}

Response 200:
{
  "id": "UUID",
  "name": "string",
  "description": "string",
  "ownerId": "UUID",
  "ownerName": "string",
  "ownerEmail": "string",
  "createdAt": "timestamp",
  "updatedAt": "timestamp"
}
```

#### 4.3 기본 워크스페이스 설정
```http
POST /api/workspaces/default
Authorization: Bearer <access_token>
Content-Type: application/json

Request:
{
  "workspace_id": "UUID"
}

Response 200: (empty body)
```

#### 4.4 워크스페이스 조회
```http
GET /api/workspaces/{workspaceId}
Authorization: Bearer <access_token>

Response 200: WorkspaceResponse
```

#### 4.5 워크스페이스 수정 (OWNER만)
```http
PUT /api/workspaces/{workspaceId}
Authorization: Bearer <access_token>
Content-Type: application/json

Request:
{
  "name": "string",
  "description": "string"
}

Response 200: WorkspaceResponse
```

#### 4.6 워크스페이스 삭제 (OWNER만)
```http
DELETE /api/workspaces/{workspaceId}
Authorization: Bearer <access_token>

Response 200: (empty body)
```

#### 4.7 워크스페이스 검색
```http
GET /api/workspaces/search?query={검색어}
Authorization: Bearer <access_token>

Response 200: WorkspaceResponse[]
Note: 현재 구현되지 않음 (빈 배열 반환)
```

---

### 5. 워크스페이스 멤버 관리

#### 5.1 멤버 목록 조회
```http
GET /api/workspaces/{workspaceId}/members
Authorization: Bearer <access_token>

Response 200:
[
  {
    "memberId": "UUID",
    "user_id": "UUID",
    "userName": "string",
    "userEmail": "string",
    "role": "OWNER|ADMIN|MEMBER",
    "joinedAt": "timestamp"
  }
]
```

#### 5.2 멤버 역할 변경 (OWNER만)
```http
PUT /api/workspaces/{workspaceId}/members/{memberId}/role
Authorization: Bearer <access_token>
Content-Type: application/json

Request:
{
  "roleName": "ADMIN|MEMBER"
}

Response 200: WorkspaceMemberResponse
```

#### 5.3 멤버 제거 (OWNER/ADMIN만)
```http
DELETE /api/workspaces/{workspaceId}/members/{memberId}
Authorization: Bearer <access_token>

Response 200: (empty body)
```

---

### 6. 가입 신청 관리

#### 6.1 워크스페이스 가입 신청
```http
POST /api/workspaces/join-requests
Authorization: Bearer <access_token>
Content-Type: application/json

Request:
{
  "workspace_id": "UUID"
}

Response 200:
{
  "requestId": "UUID",
  "workspace_id": "UUID",
  "user_id": "UUID",
  "userName": "string",
  "status": "PENDING",
  "createdAt": "timestamp"
}
```

#### 6.2 가입 신청 처리 (OWNER/ADMIN만)
```http
PUT /api/workspaces/{workspaceId}/joinRequests/{requestId}
Authorization: Bearer <access_token>
Content-Type: application/json

Request:
{
  "status": "APPROVED|REJECTED"
}

Response 200: JoinRequestResponse
```

#### 6.3 가입 신청 목록 조회 (OWNER/ADMIN만)
```http
GET /api/workspaces/{workspaceId}/join-requests?status=PENDING
Authorization: Bearer <access_token>

Response 200: JoinRequestResponse[]
```

---

## 주요 특징

### 보안
- JWT 기반 인증
- Redis를 활용한 토큰 관리
- Rate Limiting (RateLimitingService)
- Spring Security 적용

### 데이터 관리
- 소프트 삭제 (Soft Delete) 지원
- Timestamp 자동 관리 (@CreationTimestamp, @UpdateTimestamp)
- UUID 기반 ID 사용

### 역할 기반 접근 제어
- **OWNER**: 워크스페이스 소유자 (모든 권한)
- **ADMIN**: 관리자 (멤버 관리, 가입 신청 처리)
- **MEMBER**: 일반 멤버 (읽기 권한)

### 로깅
- MDC를 활용한 요청 추적
- LoggingAspect로 AOP 기반 로깅
- 구조화된 로그 출력

---

## 참고사항

### 엔티티 매핑
- `Workspace.groupId` ↔ DB 테이블 `groups.group_id`
- `Workspace.companyName` → 실제로는 description 역할
- User와 UserProfile은 1:1 관계

### 테스트
- `/api/auth/test` 엔드포인트로 테스트용 사용자 생성 가능
- `user-service/test-api.sh` 스크립트로 API 테스트 가능

### Swagger UI
- SwaggerConfig 설정되어 있음
- 접근 경로: `/swagger-ui.html` (추정)
