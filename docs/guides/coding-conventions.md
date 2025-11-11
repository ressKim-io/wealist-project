# 코딩 컨벤션 (Coding Conventions)

## 📌 개요

이 문서는 Wealist 프로젝트의 백엔드 (User Service) 코딩 규칙을 정의합니다.
모든 개발자는 이 규칙을 따라 일관된 코드를 작성해야 합니다.

---

## 🔤 네이밍 규칙 (Naming Conventions)

### 1. 변수명 및 컬럼명

#### ✅ **camelCase 사용**
모든 Java 변수명, 엔티티 필드명, 컬럼명은 **camelCase**를 사용합니다.

**❌ 잘못된 예시 (underscore_case)**:
```java
@Column(name = "user_id")
private UUID user_id;

@Column(name = "created_at")
private LocalDateTime created_at;
```

**✅ 올바른 예시 (camelCase)**:
```java
@Column(name = "userId")
private UUID userId;

@Column(name = "createdAt")
private LocalDateTime createdAt;
```

---

### 2. 공통 변수명에 접두사 규칙

**규칙**: `id`, `name`, `description` 같은 공통 변수명에는 **엔티티명 접두사**를 붙입니다.

#### 이유:
- 코드 가독성 향상
- 변수의 소속을 명확히 표현
- 조인 쿼리에서 컬럼 충돌 방지

#### ✅ **올바른 예시**:

**Workspace 엔티티**:
```java
@Entity
public class Workspace {
    @Column(name = "workspaceId")
    private UUID workspaceId;           // ✅ workspace + Id

    @Column(name = "workspaceName")
    private String workspaceName;       // ✅ workspace + Name

    @Column(name = "workspaceDescription")
    private String workspaceDescription; // ✅ workspace + Description
}
```

**User 엔티티**:
```java
@Entity
public class User {
    @Column(name = "userId")
    private UUID userId;                // ✅ user + Id
}
```

**UserProfile 엔티티**:
```java
@Entity
public class UserProfile {
    @Column(name = "profileId")
    private UUID profileId;             // ✅ profile + Id

    @Column(name = "nickName")
    private String nickName;            // ✅ nick + Name
}
```

#### ❌ **잘못된 예시**:
```java
@Entity
public class Workspace {
    @Column(name = "id")
    private UUID id;                    // ❌ 너무 일반적

    @Column(name = "name")
    private String name;                // ❌ 어떤 name인지 불명확

    @Column(name = "description")
    private String description;         // ❌ 어떤 description인지 불명확
}
```

---

### 3. 특수 필드 네이밍

#### Boolean 필드
- `is` 접두사 사용
- 예: `isActive`, `isPublic`, `isDefault`

```java
@Column(name = "isActive")
private Boolean isActive;

@Column(name = "isPublic")
private Boolean isPublic;
```

#### 날짜/시간 필드
- `At` 접미사 사용
- 예: `createdAt`, `updatedAt`, `deletedAt`, `joinedAt`

```java
@Column(name = "createdAt")
private LocalDateTime createdAt;

@Column(name = "deletedAt")
private LocalDateTime deletedAt;
```

---

## 🗂️ 엔티티 구조 규칙

### 1. Primary Key 네이밍
- 모든 엔티티의 PK는 `{엔티티명}Id` 형식
- 타입은 `UUID` 사용

```java
// User 엔티티
@Column(name = "userId")
private UUID userId;

// Workspace 엔티티
@Column(name = "workspaceId")
private UUID workspaceId;

// WorkspaceMember 엔티티
@Column(name = "workspaceMemberId")
private UUID workspaceMemberId;
```

---

### 2. Foreign Key 네이밍
- FK는 참조하는 엔티티의 PK명과 동일하게 사용
- 예: `userId`, `workspaceId`, `ownerId`

```java
@Entity
public class UserProfile {
    @Column(name = "userId", nullable = false, unique = true)
    private UUID userId;  // User 테이블의 userId 참조
}

@Entity
public class Workspace {
    @Column(name = "ownerId", nullable = false)
    private UUID ownerId;  // User 테이블의 userId 참조
}
```

---

### 3. 테이블명 규칙
- 단수형 또는 복합명사 사용
- camelCase 사용

```java
@Table(name = "users")           // ✅ 복수형 OK
@Table(name = "userProfile")     // ✅ 복합명사 camelCase
@Table(name = "workspaceMembers") // ✅ 복수형 + camelCase
```

---

## 👤 User vs UserProfile 분리 규칙

### User 엔티티
**역할**: 인증 정보만 저장

**포함 필드**:
- `userId`, `email`, `googleId`, `provider`
- `createdAt`, `updatedAt`, `isActive`, `deletedAt`

```java
@Entity
@Table(name = "users")
public class User {
    private UUID userId;
    private String email;
    private String googleId;
    private String provider;
    // ... 타임스탬프 필드
}
```

---

### UserProfile 엔티티
**역할**: 사용자 프로필 정보 저장

**포함 필드**:
- `profileId`, `userId` (FK)
- `nickName`, `email`, `profileImageUrl`
- `createdAt`, `updatedAt`

**⚠️ 중요**: `name` 필드는 사용하지 않습니다. **`nickName`만 사용**합니다.

```java
@Entity
@Table(name = "userProfile")
public class UserProfile {
    private UUID profileId;
    private UUID userId;
    private String nickName;      // ✅ nickName 사용
    private String email;
    private String profileImageUrl;
    // ... 타임스탬프 필드
}
```

#### ❌ 금지:
```java
private String name;           // ❌ name 필드 사용 금지
private String userNickName;   // ❌ 중복 접두사
```

---

## 🔄 소프트 삭제 패턴

삭제가 필요한 엔티티는 **소프트 삭제 패턴**을 사용합니다.

### 필수 필드
```java
@Column(name = "isActive", nullable = false)
@Builder.Default
private Boolean isActive = true;

@Column(name = "deletedAt")
private LocalDateTime deletedAt;
```

### 삭제 메서드
```java
public void softDelete() {
    this.isActive = false;
    this.deletedAt = LocalDateTime.now();
}

public void restore() {
    this.isActive = true;
    this.deletedAt = null;
}
```

### 적용 대상
- ✅ User
- ✅ Workspace
- ❌ UserProfile (User 삭제 시 cascade)
- ❌ WorkspaceMember (활성화 여부만 관리)

---

## ⏰ 타임스탬프 규칙

### 자동 생성 어노테이션 사용
```java
@CreationTimestamp
@Column(name = "createdAt", updatable = false)
private LocalDateTime createdAt;

@UpdateTimestamp
@Column(name = "updatedAt")
private LocalDateTime updatedAt;
```

### 수동 설정이 필요한 경우
```java
@Column(name = "deletedAt")
private LocalDateTime deletedAt;  // 삭제 시점에 수동 설정

@Column(name = "joinedAt", updatable = false)
private LocalDateTime joinedAt;   // 가입 시점에 수동 설정
```

---

## 📊 DTO 네이밍 규칙

### Request DTO
- 동사 + 엔티티명 + `Request`
- 예: `CreateWorkspaceRequest`, `UpdateProfileRequest`

```java
public class CreateWorkspaceRequest {
    private String workspaceName;
    private String workspaceDescription;
}

public class UpdateProfileRequest {
    private String nickName;
    private String email;
    private String profileImageUrl;
}
```

---

### Response DTO
- 엔티티명 + `Response`
- 예: `UserProfileResponse`, `WorkspaceResponse`

```java
public record UserProfileResponse(
    UUID profileId,
    UUID userId,
    String nickName,
    String email,
    String profileImageUrl,
    LocalDateTime createdAt,
    LocalDateTime updatedAt
) {}
```

---

## 🔧 메서드 네이밍 규칙

### 업데이트 메서드
- `update` + 필드명 (camelCase)

```java
// UserProfile 엔티티
public void updateNickName(String nickName) {
    this.nickName = nickName;
}

public void updateEmail(String email) {
    this.email = email;
}

public void updateProfileImageUrl(String profileImageUrl) {
    this.profileImageUrl = profileImageUrl;
}
```

### ❌ 잘못된 예시:
```java
public void setNickName(String nickName) { }  // ❌ set 대신 update 사용
public void changeName(String name) { }       // ❌ name 필드는 존재하지 않음
```

---

## 🎯 필드명 예시 정리

### 공통 패턴

| 원래 이름 | 변경 후 (camelCase + 접두사) | 엔티티 |
|----------|---------------------------|--------|
| `id` | `userId` | User |
| `id` | `workspaceId` | Workspace |
| `id` | `profileId` | UserProfile |
| `name` | `workspaceName` | Workspace |
| `name` | `nickName` | UserProfile |
| `description` | `workspaceDescription` | Workspace |
| `user_id` | `userId` | 모든 FK |
| `workspace_id` | `workspaceId` | 모든 FK |
| `created_at` | `createdAt` | 모든 엔티티 |
| `updated_at` | `updatedAt` | 모든 엔티티 |
| `is_active` | `isActive` | User, Workspace 등 |
| `deleted_at` | `deletedAt` | User, Workspace 등 |

---

## 📝 전체 예시 코드

### Workspace 엔티티 (완전한 예시)
```java
@Entity
@Table(name = "workspaces")
@Getter
@Setter
@NoArgsConstructor
@AllArgsConstructor
@Builder
public class Workspace {

    @Id
    @GeneratedValue(strategy = GenerationType.UUID)
    @Column(name = "workspaceId", updatable = false, nullable = false, columnDefinition = "UUID")
    private UUID workspaceId;

    @Column(name = "ownerId", nullable = false, columnDefinition = "UUID")
    private UUID ownerId;

    @Column(name = "workspaceName", nullable = false)
    private String workspaceName;

    @Column(name = "workspaceDescription", nullable = false)
    private String workspaceDescription;

    @Column(name = "isPublic", nullable = false)
    @Builder.Default
    private Boolean isPublic = false;

    @Column(name = "needApproved", nullable = false)
    @Builder.Default
    private Boolean needApproved = true;

    @CreationTimestamp
    @Column(name = "createdAt", updatable = false)
    private LocalDateTime createdAt;

    @Column(name = "deletedAt")
    private LocalDateTime deletedAt;

    @Column(name = "isActive", nullable = false)
    @Builder.Default
    private Boolean isActive = true;

    // 소프트 삭제
    public void softDelete() {
        this.isActive = false;
        this.deletedAt = LocalDateTime.now();
    }

    public void restore() {
        this.isActive = true;
        this.deletedAt = null;
    }
}
```

---

## ✅ 체크리스트

새로운 엔티티나 필드를 추가할 때 다음을 확인하세요:

- [ ] 모든 변수명이 camelCase인가?
- [ ] `id`, `name`, `description` 같은 공통 변수에 엔티티명 접두사를 붙였는가?
- [ ] PK는 `{엔티티명}Id` 형식인가?
- [ ] Boolean 필드는 `is` 접두사를 사용했는가?
- [ ] 날짜 필드는 `At` 접미사를 사용했는가?
- [ ] UserProfile에 `name` 필드를 사용하지 않았는가? (nickName만 사용)
- [ ] User 엔티티는 인증 정보만, UserProfile은 프로필 정보만 포함하는가?
- [ ] 소프트 삭제가 필요한 엔티티에 `isActive`와 `deletedAt`을 추가했는가?
- [ ] 타임스탬프 필드에 적절한 어노테이션을 사용했는가?
- [ ] DTO 네이밍이 규칙에 맞는가?

---

## 🚫 금지 사항

### ❌ 절대 하지 말 것:

1. **underscore_case 사용 금지**
   ```java
   private String user_name;        // ❌
   @Column(name = "created_at")     // ❌
   ```

2. **일반적인 필드명 사용 금지**
   ```java
   private UUID id;                 // ❌ workspaceId 사용
   private String name;             // ❌ workspaceName 또는 nickName 사용
   private String description;      // ❌ workspaceDescription 사용
   ```

3. **UserProfile에 name 필드 사용 금지**
   ```java
   private String name;             // ❌ 무조건 nickName 사용
   private String userName;         // ❌ nickName 사용
   private String userNickName;     // ❌ nickName 사용 (중복 접두사)
   ```

4. **User 엔티티에 프로필 정보 추가 금지**
   ```java
   // User 엔티티에는 추가하면 안됨
   private String nickName;         // ❌ UserProfile에 위치해야 함
   private String profileImageUrl;  // ❌ UserProfile에 위치해야 함
   ```

---

## 📚 참고 자료

- [user-service-api.md](../api/user-service-api.md) - 전체 API 및 엔티티 문서
- Java Naming Conventions: camelCase for fields, PascalCase for classes
- JPA/Hibernate Best Practices

---

## 🔄 버전 히스토리

- **v1.1** (2025-01-11): ressKim-io 저장소로 병합
  - 문서 경로 업데이트
  - 구조 정리
- **v1.0** (2025-01-09): 초기 버전 작성 (OrangesCloud 저장소)
  - camelCase 규칙 정의
  - 엔티티명 접두사 규칙 추가
  - User/UserProfile 분리 규칙 명시
  - UserProfile.name → nickName 변경 반영

---

## 💬 문의

규칙에 대한 질문이나 개선 제안이 있다면 팀에 문의하세요.
