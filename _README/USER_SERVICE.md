# User Service API Documentation

## 📋 Overview

User Service는 사용자 인증, 프로필 관리, 워크스페이스 관리를 담당하는 마이크로서비스입니다.

- **Base URL**: `/api`
- **인증 방식**: JWT Bearer Token
- **OAuth Provider**: Google OAuth 2.0

---

## 🗂️ Database Entities

### 1. User (사용자)

사용자 인증 정보를 저장하는 엔티티

| 필드        | 타입          | 설명                            |
| ----------- | ------------- | ------------------------------- |
| `userId`    | UUID          | 사용자 고유 ID (PK)             |
| `email`     | String        | 이메일 (unique, not null)       |
| `provider`  | String        | OAuth 제공자 (기본값: "google") |
| `googleId`  | String        | Google OAuth ID (unique)        |
| `createdAt` | LocalDateTime | 생성 시간                       |
| `updatedAt` | LocalDateTime | 수정 시간                       |
| `isActive`  | Boolean       | 활성화 상태 (기본값: true)      |
| `deletedAt` | LocalDateTime | 삭제 시간 (소프트 삭제)         |

**테이블명**: `users`

---

### 2. UserProfile (사용자 프로필)

사용자 프로필 정보를 저장하는 엔티티

| 필드              | 타입          | 설명                             |
| ----------------- | ------------- | -------------------------------- |
| `profileId`       | UUID          | 프로필 고유 ID (PK)              |
| `userId`          | UUID          | 사용자 ID (FK, unique, not null) |
| `nickName`        | String        | 닉네임 (최대 50자)               |
| `email`           | String        | 이메일 (최대 100자)              |
| `profileImageUrl` | String        | 프로필 이미지 URL                |
| `createdAt`       | LocalDateTime | 생성 시간                        |
| `updatedAt`       | LocalDateTime | 수정 시간                        |

**테이블명**: `userProfile`

**참고**:

- name 필드는 제거되었으며, nickName만 사용합니다
- Google OAuth 로그인 시 Google의 name이 nickName으로 저장됩니다

---

### 3. Workspace (워크스페이스)

프로젝트 워크스페이스를 관리하는 엔티티

| 필드                   | 타입          | 설명                               |
| ---------------------- | ------------- | ---------------------------------- |
| `workspaceId`          | UUID          | 워크스페이스 고유 ID (PK)          |
| `ownerId`              | UUID          | 소유자 ID (not null)               |
| `workspaceName`        | String        | 워크스페이스 이름 (not null)       |
| `workspaceDescription` | String        | 워크스페이스 설명 (not null)       |
| `isPublic`             | Boolean       | 공개 여부 (기본값: false)          |
| `needApproved`         | Boolean       | 가입 승인 필요 여부 (기본값: true) |
| `createdAt`            | LocalDateTime | 생성 시간                          |
| `deletedAt`            | LocalDateTime | 삭제 시간 (소프트 삭제)            |
| `isActive`             | Boolean       | 활성화 상태 (기본값: true)         |

**테이블명**: `workspaces`

**참고**:

- isPublic: 검색 및 초대 가능성 여부
- needApproved: workspace 가입 시 승인 필요 여부

---

### 4. WorkspaceMember (워크스페이스 멤버)

워크스페이스 멤버십 및 역할을 관리하는 엔티티

| 필드                | 타입          | 설명                                   |
| ------------------- | ------------- | -------------------------------------- |
| `workspaceMemberId` | UUID          | 멤버 고유 ID (PK)                      |
| `workspaceId`       | UUID          | 워크스페이스 ID (not null)             |
| `userId`            | UUID          | 사용자 ID (not null)                   |
| `role`              | Enum          | 역할 (OWNER, ADMIN, MEMBER)            |
| `isDefault`         | boolean       | 기본 워크스페이스 여부 (기본값: false) |
| `joinedAt`          | LocalDateTime | 가입 시간                              |
| `updatedAt`         | LocalDateTime | 수정 시간                              |
| `isActive`          | Boolean       | 활성화 상태 (기본값: true)             |

**테이블명**: `workspaceMembers`

**역할 (WorkspaceRole)**:

- `OWNER`: 워크스페이스 소유자 (모든 권한)
- `ADMIN`: 관리자 (멤버 관리, 가입 승인/거절)
- `MEMBER`: 일반 멤버

**참고**:

- User 엔티티와 ManyToOne 관계 (@JoinColumn으로 연결)
- isDefault는 primitive boolean 타입 (Lombok이 isDefault() getter 생성)

---

### 5. WorkspaceJoinRequest (워크스페이스 가입 요청)

워크스페이스 가입 신청을 관리하는 엔티티

| 필드            | 타입          | 설명                               |
| --------------- | ------------- | ---------------------------------- |
| `joinRequestId` | UUID          | 가입 요청 고유 ID (PK)             |
| `workspaceId`   | UUID          | 워크스페이스 ID (not null)         |
| `userId`        | UUID          | 사용자 ID (not null)               |
| `status`        | Enum          | 상태 (PENDING, APPROVED, REJECTED) |
| `requestedAt`   | LocalDateTime | 요청 시간                          |
| `updatedAt`     | LocalDateTime | 수정 시간                          |

**테이블명**: `workspaceJoinRequests`

**상태 (JoinRequestStatus)**:

- `PENDING`: 승인 대기 중
- `APPROVED`: 승인됨
- `REJECTED`: 거절됨

---

## 🔌 API Endpoints

### Authentication API (`/api/auth`)

#### 1. 로그아웃

```
POST /api/auth/logout
```

**설명**: 현재 세션을 종료하고 토큰을 무효화합니다.

**Headers**:

```
Authorization: Bearer {accessToken}
```

**Response**: `200 OK`

```json
{
  "success": true,
  "message": "로그아웃 성공"
}
```

---

#### 2. 토큰 갱신

```
POST /api/auth/refresh
```

**설명**: Refresh Token을 사용하여 새로운 Access Token을 발급받습니다.

**Request Body**:

```json
{
  "refreshToken": "string"
}
```

**Response**: `200 OK`

```json
{
  "accessToken": "string",
  "refreshToken": "string",
  "userId": "uuid",
  "name": "string",
  "email": "string",
  "tokenType": "Bearer"
}
```

---

#### 3. 현재 사용자 정보 조회

```
GET /api/auth/me
```

**설명**: 현재 인증된 사용자의 정보를 조회합니다.

**Headers**:

```
Authorization: Bearer {accessToken}
```

**Response**: `200 OK`

```json
{
  "userId": "uuid",
  "email": "string",
  "provider": "google",
  "googleId": "string",
  "createdAt": "2025-01-01T00:00:00",
  "updatedAt": "2025-01-01T00:00:00",
  "isActive": true
}
```

---

#### 4. 테스트 로그인

```
POST /api/auth/test-login
```

**설명**: 테스트용 Google OAuth 사용자를 생성하고 로그인합니다.

**Response**: `200 OK`

```json
{
  "accessToken": "string",
  "refreshToken": "string",
  "userId": "uuid",
  "name": "string",
  "email": "string",
  "tokenType": "Bearer"
}
```

---

### User API (`/api/users`)

#### 1. 내 정보 조회

```
GET /api/users/me
```

**설명**: 현재 인증된 사용자의 정보를 조회합니다.

**Headers**:

```
Authorization: Bearer {accessToken}
```

**Response**: `200 OK`

```json
{
  "userId": "uuid",
  "email": "string",
  "provider": "google",
  "googleId": "string",
  "createdAt": "2025-01-01T00:00:00",
  "updatedAt": "2025-01-01T00:00:00",
  "isActive": true
}
```

---

#### 2. 특정 사용자 정보 조회

```
GET /api/users/{userId}
```

**설명**: 특정 사용자의 정보를 조회합니다.

**Path Parameters**:

- `userId` (UUID): 사용자 ID

**Response**: `200 OK`

---

#### 3. 계정 삭제

```
DELETE /api/users/me
```

**설명**: 현재 사용자 계정을 삭제합니다 (소프트 삭제).

**Response**: `200 OK`

```json
{
  "success": true,
  "message": "계정이 삭제되었습니다."
}
```

---

#### 4. 사용자 삭제 (관리자용)

```
DELETE /api/users/{userId}
```

**설명**: 특정 사용자를 삭제합니다 (관리자용).

**Response**: `200 OK`

---

#### 5. 사용자 복구

```
PUT /api/users/{userId}/restore
```

**설명**: 삭제된 사용자를 복구합니다.

**Response**: `200 OK`

---

### Profile API (`/api/profiles`)

#### 1. 내 프로필 조회

```
GET /api/profiles/me
```

**설명**: 현재 인증된 사용자의 프로필을 조회합니다.

**Headers**:

```
Authorization: Bearer {accessToken}
```

**Response**: `200 OK`

```json
{
  "profileId": "uuid",
  "userId": "uuid",
  "nickName": "코딩왕",
  "email": "user@example.com",
  "profileImageUrl": "https://example.com/image.jpg",
  "createdAt": "2025-01-01T00:00:00",
  "updatedAt": "2025-01-01T00:00:00"
}
```

---

#### 2. 프로필 정보 통합 업데이트

```
PUT /api/profiles/me
```

**설명**: 인증된 사용자의 닉네임, 이메일, 프로필 이미지 URL을 업데이트합니다.

**Request Body** (모든 필드 optional):

```json
{
  "nickName": "새로운닉네임",
  "email": "newemail@example.com",
  "profileImageUrl": "https://new.image.url/avatar.jpg"
}
```

**Response**: `200 OK`

```json
{
  "profileId": "uuid",
  "userId": "uuid",
  "nickName": "새로운닉네임",
  "email": "newemail@example.com",
  "profileImageUrl": "https://new.image.url/avatar.jpg",
  "createdAt": "2025-01-01T00:00:00",
  "updatedAt": "2025-01-01T12:00:00"
}
```

---

#### 3. 프로필 사진 업데이트

```
PUT /api/profiles/me/image
```

**설명**: 프로필 사진 URL만 업데이트합니다.

**Request Body**:

```json
{
  "profileImageUrl": "https://new.image.url/avatar.jpg"
}
```

**Response**: `200 OK`

---

### Workspace API (`/api/workspaces`)

#### 1. 워크스페이스 목록 조회

```
GET /api/workspaces
```

**설명**: 현재 사용자가 속한 모든 워크스페이스를 조회합니다.

**Headers**:

```
Authorization: Bearer {accessToken}
```

**Response**: `200 OK`

```json
[
  {
    "workspaceId": "uuid",
    "workspaceName": "My Workspace",
    "workspaceDescription": "설명",
    "ownerId": "uuid",
    "ownerName": "소유자닉네임",
    "ownerEmail": "owner@example.com",
    "isPublic": false,
    "needApproved": true,
    "createdAt": "2025-01-01T00:00:00"
  }
]
```

---

#### 2. 워크스페이스 생성

```
POST /api/workspaces
```

**설명**: 새로운 워크스페이스를 생성합니다. (생성자가 자동으로 OWNER)

**Request Body**:

```json
{
  "workspaceName": "New Workspace",
  "workspaceDescription": "프로젝트 설명"
}
```

**Response**: `200 OK`

---

#### 3. 기본 워크스페이스 설정

```
POST /api/workspaces/default
```

**설명**: 사용자의 기본 워크스페이스를 설정합니다.

**Request Body**:

```json
{
  "workspaceId": "uuid"
}
```

**Response**: `200 OK`

---

#### 4. 특정 워크스페이스 조회

```
GET /api/workspaces/{workspaceId}
```

**설명**: 특정 워크스페이스의 정보를 조회합니다. (멤버만 가능)

**Path Parameters**:

- `workspaceId` (UUID): 워크스페이스 ID

**Response**: `200 OK`

---

#### 5. 워크스페이스 수정

```
PUT /api/workspaces/{workspaceId}
```

**설명**: 워크스페이스 정보를 수정합니다. (OWNER만 가능)

**Request Body** (모든 필드 optional):

```json
{
  "name": "Updated Name",
  "workspaceDescription": "Updated Description"
}
```

**Response**: `200 OK`

---

#### 6. 워크스페이스 삭제

```
DELETE /api/workspaces/{workspaceId}
```

**설명**: 워크스페이스를 삭제합니다 (소프트 삭제). (OWNER만 가능)

**Response**: `200 OK`

---

#### 7. 워크스페이스 멤버 목록 조회

```
GET /api/workspaces/{workspaceId}/members
```

**설명**: 특정 워크스페이스의 모든 멤버를 조회합니다.

**Response**: `200 OK`

```json
[
  {
    "id": "uuid",
    "workspaceId": "uuid",
    "userId": "uuid",
    "userName": "멤버닉네임",
    "userEmail": "member@example.com",
    "profileImageUrl": "https://example.com/image.jpg",
    "roleName": "OWNER",
    "isDefault": true,
    "joinedAt": "2025-01-01T00:00:00"
  }
]
```

---

#### 8. 멤버 역할 변경

```
PUT /api/workspaces/{workspaceId}/members/{memberId}/role
```

**설명**: 멤버의 역할을 변경합니다. (OWNER만 가능)

**Request Body**:

```json
{
  "roleName": "ADMIN"
}
```

**가능한 값**: `OWNER`, `ADMIN`, `MEMBER`

**Response**: `200 OK`

---

#### 9. 멤버 제거

```
DELETE /api/workspaces/{workspaceId}/members/{memberId}
```

**설명**: 워크스페이스에서 멤버를 제거합니다. (OWNER/ADMIN만 가능, OWNER는 제거 불가)

**Response**: `200 OK`

---

#### 10. 워크스페이스 가입 신청

```
POST /api/workspaces/{workspaceId}/joinRequests
```

**설명**: 워크스페이스 가입을 신청합니다.

**Response**: `200 OK`

```json
{
  "id": "uuid",
  "workspaceId": "uuid",
  "userId": "uuid",
  "userName": "신청자닉네임",
  "userEmail": "user@example.com",
  "status": "PENDING",
  "requestedAt": "2025-01-01T00:00:00",
  "updatedAt": "2025-01-01T00:00:00"
}
```

---

#### 11. 가입 신청 목록 조회

```
GET /api/workspaces/{workspaceId}/joinRequests
```

**설명**: 워크스페이스의 가입 신청 목록을 조회합니다. (OWNER/ADMIN만 가능)

**Query Parameters** (optional):

- `status`: 상태 필터 (`PENDING`, `APPROVED`, `REJECTED`)

**Response**: `200 OK`

---

#### 12. 가입 신청 승인/거절

```
PUT /api/workspaces/{workspaceId}/joinRequests/{requestId}
```

**설명**: 가입 신청을 승인하거나 거절합니다. (OWNER/ADMIN만 가능)

**Request Body**:

```json
{
  "status": "APPROVED"
}
```

**가능한 값**: `APPROVED`, `REJECTED`

**Response**: `200 OK`

---

## 🔐 Authentication Flow

### Google OAuth 2.0 Login Flow

1. 클라이언트가 `/oauth2/authorization/google`로 리디렉션
2. Google 로그인 완료 후 `/api/oauth2/callback/google`로 콜백
3. 서버가 Google에서 사용자 정보 받아옴 (email, googleId, name)
4. `CustomOAuth2UserService`에서 사용자 생성/조회:
   - User 엔티티 생성 (email, googleId, provider)
   - UserProfile 엔티티 생성 (nickName = Google name)
5. JWT 토큰 생성 (Access Token, Refresh Token)
6. 프론트엔드로 리디렉션 (토큰 포함)

**OAuth Success Handler 리디렉션 URL**:

```
http://localhost:3000/auth/callback?
  token={accessToken}&
  refreshToken={refreshToken}&
  userId={userId}&
  name={nickName}
```

---

## 📝 Important Notes

### Entity 변경 사항

1. **UserProfile**:

   - `name` 필드 제거됨
   - `userNickName` → `nickName`으로 변경
   - Google OAuth의 name이 nickName으로 저장됨

2. **Workspace**:

   - `description` → `workspaceDescription`으로 변경
   - 이전 "groups" 테이블에서 "workspaces"로 변경됨

3. **모든 엔티티**:
   - 컬럼명이 underscore_case에서 camelCase로 변경됨
   - 예: `user_id` → `userId`, `created_at` → `createdAt`

### 권한 체계

- **OWNER**: 모든 작업 가능 (워크스페이스 삭제, 멤버 역할 변경 등)
- **ADMIN**: 멤버 관리, 가입 신청 승인/거절 가능
- **MEMBER**: 워크스페이스 조회 및 일반 작업만 가능

### 소프트 삭제

- User, Workspace 엔티티는 소프트 삭제 방식 사용
- `isActive` 필드가 false로 변경되고 `deletedAt`에 삭제 시간 기록
- 실제 데이터는 DB에 유지됨

---

## 🛠️ Tech Stack

- **Framework**: Spring Boot 3.5.6
- **Language**: Java 21
- **Database**: PostgreSQL
- **ORM**: JPA/Hibernate
- **Authentication**: Spring Security + JWT + OAuth 2.0
- **Cache**: Redis (프로필 캐싱)
- **API Documentation**: Swagger/OpenAPI 3.0
- **Build Tool**: Gradle 8.5.0

---

## 📌 Environment Variables

```env
# Database
SPRING_DATASOURCE_URL=jdbc:postgresql://localhost:5432/wealist
SPRING_DATASOURCE_USERNAME=your_username
SPRING_DATASOURCE_PASSWORD=your_password

# JWT
JWT_SECRET=your_jwt_secret_key
JWT_EXPIRATION=3600000
JWT_REFRESH_EXPIRATION=604800000

# Google OAuth
SPRING_SECURITY_OAUTH2_CLIENT_REGISTRATION_GOOGLE_CLIENT_ID=your_client_id
SPRING_SECURITY_OAUTH2_CLIENT_REGISTRATION_GOOGLE_CLIENT_SECRET=your_client_secret

# Redis
SPRING_REDIS_HOST=localhost
SPRING_REDIS_PORT=6379
```

---

## 📞 Contact

For questions or issues, please contact the development team.
