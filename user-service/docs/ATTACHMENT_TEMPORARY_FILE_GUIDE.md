# User Service Attachment 임시 파일 관리 가이드

## 개요

User Service에 프로필 이미지를 위한 임시 파일 관리 시스템이 구현되었습니다. 이 시스템은 Board Service와 동일한 방식으로 작동하며, 업로드된 파일을 임시로 저장한 후 프로필 업데이트 시 확정하는 방식입니다.

## 주요 기능

### 1. 임시 첨부파일 관리
- 파일 업로드 후 임시 상태(TEMP)로 저장
- 1시간 후 자동 만료
- 프로필 업데이트 시 확정 상태(CONFIRMED)로 변경

### 2. 자동 정리 작업
- 매 시간 정각에 만료된 임시 파일 자동 삭제
- S3 파일과 DB 레코드 모두 삭제

## 데이터 모델

### Attachment Entity

```java
@Entity
@Table(name = "attachments")
public class Attachment {
    private UUID id;
    private EntityType entityType;      // USER_PROFILE
    private UUID entityId;              // nullable - 임시 파일은 null
    private AttachmentStatus status;    // TEMP or CONFIRMED
    private String fileName;
    private String fileUrl;
    private Long fileSize;
    private String contentType;
    private UUID uploadedBy;
    private LocalDateTime expiresAt;    // 만료 시간 (생성 시간 + 1시간)
    private LocalDateTime createdAt;
    private LocalDateTime updatedAt;
    private LocalDateTime deletedAt;    // Soft delete
}
```

### Status 흐름

```
TEMP (임시) → CONFIRMED (확정)
    ↓
만료 시 자동 삭제
```

## API 엔드포인트

### 1. Presigned URL 생성
```http
POST /api/profiles/me/image/presigned-url
Content-Type: application/json

Request:
{
  "workspaceId": "abc123",
  "fileName": "profile.jpg",
  "fileSize": 512000,
  "contentType": "image/jpeg"
}

Response:
{
  "uploadUrl": "https://wealist-dev-files.s3.ap-northeast-2.amazonaws.com/...",
  "fileKey": "user/abc123/2024/01/user-id_timestamp.jpg",
  "expiresIn": 300
}
```

### 2. S3에 직접 업로드
```http
PUT {uploadUrl}
Content-Type: image/jpeg
Body: {binary file data}
```

### 3. 임시 첨부파일 메타데이터 저장
```http
POST /api/profiles/me/image/attachment
Content-Type: application/json

Request:
{
  "fileKey": "user/abc123/2024/01/user-id_timestamp.jpg",
  "fileName": "profile.jpg",
  "fileSize": 512000,
  "contentType": "image/jpeg"
}

Response:
{
  "id": "attachment-uuid",
  "entityType": "USER_PROFILE",
  "entityId": null,
  "status": "TEMP",
  "fileName": "profile.jpg",
  "fileUrl": "https://wealist-dev-files.s3.ap-northeast-2.amazonaws.com/...",
  "fileSize": 512000,
  "contentType": "image/jpeg",
  "uploadedBy": "user-uuid",
  "uploadedAt": "2024-01-15T10:30:00",
  "expiresAt": "2024-01-15T11:30:00"
}
```

### 4. 프로필 이미지 업데이트 (첨부파일 확정)
```http
PUT /api/profiles/me/image
Content-Type: application/json

Request:
{
  "workspaceId": "abc123",
  "fileKey": "user/abc123/2024/01/user-id_timestamp.jpg"
}

Response:
{
  "profileId": "uuid",
  "userId": "uuid",
  "workspaceId": "uuid",
  "nickName": "사용자 이름",
  "email": "user@example.com",
  "profileImageUrl": "https://wealist-dev-files.s3.ap-northeast-2.amazonaws.com/...",
  ...
}
```

## 클라이언트 플로우

### 전체 프로세스

```
1. Presigned URL 요청
   ↓
2. S3에 직접 업로드
   ↓
3. 임시 첨부파일 메타데이터 저장
   ↓
4. 프로필 업데이트 (첨부파일 확정)
```

### 예제 코드 (JavaScript)

```javascript
// 1. Presigned URL 요청
const presignedResponse = await fetch('/api/profiles/me/image/presigned-url', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${token}`
  },
  body: JSON.stringify({
    workspaceId: 'abc123',
    fileName: file.name,
    fileSize: file.size,
    contentType: file.type
  })
});

const { uploadUrl, fileKey } = await presignedResponse.json();

// 2. S3에 직접 업로드
await fetch(uploadUrl, {
  method: 'PUT',
  headers: {
    'Content-Type': file.type
  },
  body: file
});

// 3. 임시 첨부파일 메타데이터 저장
const attachmentResponse = await fetch('/api/profiles/me/image/attachment', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${token}`
  },
  body: JSON.stringify({
    fileKey: fileKey,
    fileName: file.name,
    fileSize: file.size,
    contentType: file.type
  })
});

const attachment = await attachmentResponse.json();

// 4. 프로필 업데이트
const profileResponse = await fetch('/api/profiles/me/image', {
  method: 'PUT',
  headers: {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${token}`
  },
  body: JSON.stringify({
    workspaceId: 'abc123',
    fileKey: fileKey
  })
});

const profile = await profileResponse.json();
console.log('프로필 업데이트 완료:', profile);
```

## 자동 정리 작업 (Cleanup Job)

### 설정

```java
@Component
public class AttachmentCleanupJob {
    
    @Scheduled(cron = "0 0 * * * *")  // 매 시간 정각
    public void cleanupExpiredAttachments() {
        // 만료된 임시 파일 조회 및 삭제
    }
}
```

### 실행 주기
- **매 시간 정각** (0분 0초)에 자동 실행
- 예: 01:00:00, 02:00:00, 03:00:00, ...

### 동작 방식
1. `status = TEMP` AND `expiresAt < NOW()` 조건으로 만료된 파일 조회
2. S3에서 파일 삭제
3. DB에서 레코드 삭제 (hard delete)

### 로그 예시

```
2024-01-15 10:00:00 INFO  - 첨부파일 정리 작업 시작
2024-01-15 10:00:01 INFO  - 만료된 첨부파일 발견 - count: 5
2024-01-15 10:00:02 DEBUG - S3 파일 삭제 완료 - fileKey: user/abc123/2024/01/file1.jpg
2024-01-15 10:00:02 DEBUG - S3 파일 삭제 완료 - fileKey: user/abc123/2024/01/file2.jpg
...
2024-01-15 10:00:03 INFO  - 첨부파일 일괄 삭제 완료 - count: 5
2024-01-15 10:00:03 INFO  - 첨부파일 정리 작업 완료 - 삭제된 파일 수: 5
```

## 에러 처리

### 에러 코드

| 코드 | HTTP 상태 | 설명 |
|------|-----------|------|
| A001 | 404 | Attachment not found |
| A002 | 400 | Invalid attachment status |

### 예외 상황

1. **첨부파일을 찾을 수 없음**
   ```json
   {
     "code": "A001",
     "message": "Attachment not found"
   }
   ```

2. **이미 확정된 첨부파일**
   ```json
   {
     "code": "A002",
     "message": "이미 확정된 첨부파일입니다."
   }
   ```

3. **S3 삭제 실패**
   - S3 삭제 실패 시에도 DB 레코드는 삭제됨
   - 로그에 에러 기록

## 데이터베이스 스키마

### attachments 테이블

```sql
CREATE TABLE attachments (
    id UUID PRIMARY KEY,
    entity_type VARCHAR(50) NOT NULL,
    entity_id UUID,
    status VARCHAR(20) NOT NULL DEFAULT 'TEMP',
    file_name VARCHAR(255) NOT NULL,
    file_url TEXT NOT NULL,
    file_size BIGINT NOT NULL,
    content_type VARCHAR(100) NOT NULL,
    uploaded_by UUID NOT NULL,
    expires_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    deleted_at TIMESTAMP
);

-- 인덱스
CREATE INDEX idx_attachments_entity ON attachments(entity_type, entity_id);
CREATE INDEX idx_attachments_entity_id ON attachments(entity_id);
CREATE INDEX idx_attachments_status ON attachments(status);
CREATE INDEX idx_attachments_uploaded_by ON attachments(uploaded_by);
CREATE INDEX idx_attachments_expires_at ON attachments(expires_at);
```

### 쿼리 예시

```sql
-- 만료된 임시 파일 조회
SELECT * FROM attachments 
WHERE status = 'TEMP' 
  AND expires_at < NOW()
  AND deleted_at IS NULL;

-- 사용자의 프로필 이미지 첨부파일 조회
SELECT * FROM attachments
WHERE entity_type = 'USER_PROFILE'
  AND entity_id = 'user-profile-uuid'
  AND status = 'CONFIRMED'
  AND deleted_at IS NULL;
```

## 모니터링

### 주요 메트릭

1. **임시 파일 수**
   - `SELECT COUNT(*) FROM attachments WHERE status = 'TEMP'`

2. **만료 예정 파일 수**
   - `SELECT COUNT(*) FROM attachments WHERE status = 'TEMP' AND expires_at < NOW() + INTERVAL '10 minutes'`

3. **정리 작업 실행 횟수**
   - 로그 분석: `grep "첨부파일 정리 작업 완료" application.log | wc -l`

4. **삭제된 파일 수**
   - 로그 분석: `grep "삭제된 파일 수" application.log`

### 알림 설정

- 만료된 파일이 100개 이상인 경우 알림
- 정리 작업 실패 시 알림
- S3 삭제 실패율이 10% 이상인 경우 알림

## 테스트

### 단위 테스트

```bash
./gradlew test --tests "AttachmentServiceTest"
./gradlew test --tests "AttachmentCleanupJobTest"
```

### 통합 테스트

```bash
# 전체 프로필 이미지 업로드 플로우 테스트
./gradlew test --tests "ProfileImageIntegrationTest"
```

## 트러블슈팅

### 문제: 임시 파일이 삭제되지 않음

**원인:**
- Cleanup Job이 실행되지 않음
- `@EnableScheduling` 누락

**해결:**
```java
@SpringBootApplication
@EnableScheduling  // 추가
public class UserRepoApplication {
    // ...
}
```

### 문제: S3 파일 삭제 실패

**원인:**
- S3 권한 부족
- 잘못된 fileKey

**해결:**
1. IAM 정책 확인
2. fileKey 형식 검증
3. 로그 확인

### 문제: 프로필 업데이트 시 첨부파일 확정 안 됨

**원인:**
- attachmentId를 전달하지 않음
- 현재 구현에서는 fileKey만 사용

**해결:**
- 향후 attachmentId를 사용하도록 개선 필요
- 현재는 fileKey로 S3 URL 생성 후 직접 프로필 업데이트

## 향후 개선 사항

1. **첨부파일 확정 자동화**
   - 프로필 업데이트 시 attachmentId를 받아 자동으로 확정
   - 현재는 fileKey만 사용하여 직접 URL 생성

2. **배치 크기 제한**
   - 한 번에 삭제하는 파일 수 제한 (예: 100개)
   - 대량 삭제 시 성능 개선

3. **재시도 로직**
   - S3 삭제 실패 시 재시도
   - Dead Letter Queue 활용

4. **메트릭 수집**
   - Prometheus 메트릭 추가
   - Grafana 대시보드 구성

## 참고 문서

- [Presigned URL Profile Image Guide](./PRESIGNED_URL_PROFILE_IMAGE_GUIDE.md)
- [Board Service Attachment Guide](../../board-service/docs/PRESIGNED_URL_API_GUIDE.md)
- [S3 Configuration Guide](../README.md)
