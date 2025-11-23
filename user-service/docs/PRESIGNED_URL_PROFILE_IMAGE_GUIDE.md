# Presigned URL 기반 프로필 이미지 업로드 API 가이드

## 개요

이 문서는 user-service의 Presigned URL 기반 프로필 이미지 업로드 API 사용 방법을 설명합니다.

Presigned URL 방식을 사용하면 클라이언트가 백엔드 서버를 거치지 않고 S3에 직접 파일을 업로드할 수 있어, 서버 부하를 줄이고 업로드 속도를 향상시킬 수 있습니다.

## 지원 파일 형식

### 이미지만 허용
- jpg, jpeg, png, gif, webp

### 제한사항
- 최대 파일 크기: **20MB**
- 이미지 파일만 업로드 가능 (문서, 음성, 영상 파일 불가)

## API 엔드포인트

### 1. Presigned URL 생성

**Endpoint:** `POST /api/profiles/me/image/presigned-url`

**설명:** S3에 직접 업로드할 수 있는 Presigned URL을 생성합니다.

**Headers:**
```
Authorization: Bearer {access_token}
Content-Type: application/json
```

**Request Body:**
```json
{
  "workspaceId": "abc123",
  "fileName": "avatar.jpg",
  "fileSize": 512000,
  "contentType": "image/jpeg"
}
```

**Request Fields:**
- `workspaceId` (required): 워크스페이스 ID (UUID)
- `fileName` (required): 업로드할 파일 이름
- `fileSize` (required): 파일 크기 (bytes, 최대 20MB)
- `contentType` (required): 파일의 MIME 타입 (이미지만 허용)

**Response (200 OK):**
```json
{
  "uploadUrl": "https://wealist-dev-files.s3.ap-northeast-2.amazonaws.com/user/abc123/2024/01/880e8400-e29b-41d4-a716-446655440003_1704326400.jpg?X-Amz-Algorithm=...",
  "fileKey": "user/abc123/2024/01/880e8400-e29b-41d4-a716-446655440003_1704326400.jpg",
  "expiresIn": 300
}
```

**Response Fields:**
- `uploadUrl`: S3 업로드를 위한 Presigned URL (5분간 유효)
- `fileKey`: S3에 저장될 파일의 키 (프로필 업데이트 시 사용)
- `expiresIn`: URL 만료 시간 (초)

**Error Responses:**
- `400 Bad Request`: 잘못된 요청 또는 파일 검증 실패
  ```json
  {
    "code": "FILE_TOO_LARGE",
    "message": "파일 크기는 20MB를 초과할 수 없습니다."
  }
  ```
  ```json
  {
    "code": "INVALID_FILE_TYPE",
    "message": "이미지 파일만 업로드 가능합니다. 지원 형식: jpg, jpeg, png, gif, webp"
  }
  ```
- `401 Unauthorized`: 인증되지 않은 사용자
- `500 Internal Server Error`: Presigned URL 생성 실패

### 2. S3에 파일 업로드

**Endpoint:** `PUT {uploadUrl}`

**설명:** 1단계에서 받은 Presigned URL을 사용하여 S3에 직접 파일을 업로드합니다.

**Headers:**
```
Content-Type: {contentType}
```

**Body:** 파일의 바이너리 데이터

**Example (JavaScript):**
```javascript
const file = document.getElementById('fileInput').files[0];

const response = await fetch(uploadUrl, {
  method: 'PUT',
  headers: {
    'Content-Type': file.type
  },
  body: file
});

if (response.ok) {
  console.log('Upload successful');
}
```

**Example (cURL):**
```bash
curl -X PUT \
  -H "Content-Type: image/jpeg" \
  --data-binary @avatar.jpg \
  "{uploadUrl}"
```

### 3. 프로필 이미지 업데이트

**Endpoint:** `PUT /api/profiles/me/image`

**설명:** S3 업로드 완료 후, fileKey를 사용하여 프로필 이미지를 업데이트합니다.

**Headers:**
```
Authorization: Bearer {access_token}
Content-Type: application/json
```

**Request Body:**
```json
{
  "workspaceId": "abc123",
  "fileKey": "user/abc123/2024/01/880e8400-e29b-41d4-a716-446655440003_1704326400.jpg"
}
```

**Request Fields:**
- `workspaceId` (required): 워크스페이스 ID (UUID)
- `fileKey` (required): 1단계에서 받은 파일 키

**Response (200 OK):**
```json
{
  "profileId": "profile-uuid",
  "userId": "user-uuid",
  "workspaceId": "workspace-uuid",
  "nickName": "사용자 이름",
  "email": "user@example.com",
  "profileImageUrl": "https://wealist-dev-files.s3.ap-northeast-2.amazonaws.com/user/abc123/2024/01/880e8400-e29b-41d4-a716-446655440003_1704326400.jpg",
  "createdAt": "2024-01-15T10:30:00Z",
  "updatedAt": "2024-01-15T11:00:00Z"
}
```

**Response Fields:**
- `profileId`: 프로필 ID
- `userId`: 사용자 ID
- `workspaceId`: 워크스페이스 ID
- `nickName`: 사용자 닉네임
- `email`: 사용자 이메일
- `profileImageUrl`: 업데이트된 프로필 이미지 URL
- `createdAt`: 프로필 생성 시간
- `updatedAt`: 프로필 수정 시간

**Error Responses:**
- `400 Bad Request`: 잘못된 요청 또는 fileKey 검증 실패
- `401 Unauthorized`: 인증되지 않은 사용자
- `404 Not Found`: 프로필을 찾을 수 없음
- `500 Internal Server Error`: 프로필 업데이트 실패

## 전체 업로드 플로우

### JavaScript 예제

```javascript
async function uploadProfileImage(file, workspaceId, accessToken) {
  try {
    // Step 1: Presigned URL 요청
    const presignedResponse = await fetch('/api/profiles/me/image/presigned-url', {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${accessToken}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        workspaceId: workspaceId,
        fileName: file.name,
        fileSize: file.size,
        contentType: file.type
      })
    });

    if (!presignedResponse.ok) {
      throw new Error('Failed to get presigned URL');
    }

    const { uploadUrl, fileKey } = await presignedResponse.json();

    // Step 2: S3에 직접 업로드
    const uploadResponse = await fetch(uploadUrl, {
      method: 'PUT',
      headers: {
        'Content-Type': file.type
      },
      body: file
    });

    if (!uploadResponse.ok) {
      throw new Error('Failed to upload to S3');
    }

    // Step 3: 프로필 이미지 업데이트
    const updateResponse = await fetch('/api/profiles/me/image', {
      method: 'PUT',
      headers: {
        'Authorization': `Bearer ${accessToken}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        workspaceId: workspaceId,
        fileKey: fileKey
      })
    });

    if (!updateResponse.ok) {
      throw new Error('Failed to update profile');
    }

    const profile = await updateResponse.json();
    console.log('Profile updated:', profile);
    return profile;

  } catch (error) {
    console.error('Upload failed:', error);
    throw error;
  }
}
```

### React 컴포넌트 예제

```jsx
import React, { useState } from 'react';

function ProfileImageUpload({ workspaceId, accessToken, onUploadComplete }) {
  const [uploading, setUploading] = useState(false);
  const [error, setError] = useState(null);
  const [preview, setPreview] = useState(null);

  const validateFile = (file) => {
    const maxSize = 20 * 1024 * 1024; // 20MB
    const allowedTypes = ['image/jpeg', 'image/png', 'image/gif', 'image/webp'];

    if (file.size > maxSize) {
      throw new Error('파일 크기는 20MB를 초과할 수 없습니다.');
    }

    if (!allowedTypes.includes(file.type)) {
      throw new Error('이미지 파일만 업로드 가능합니다. (jpg, jpeg, png, gif, webp)');
    }
  };

  const handleFileChange = async (event) => {
    const file = event.target.files[0];
    if (!file) return;

    try {
      // 파일 검증
      validateFile(file);

      // 미리보기 생성
      const reader = new FileReader();
      reader.onloadend = () => setPreview(reader.result);
      reader.readAsDataURL(file);

      setUploading(true);
      setError(null);

      // 1. Presigned URL 요청
      const presignedRes = await fetch('/api/profiles/me/image/presigned-url', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${accessToken}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          workspaceId,
          fileName: file.name,
          fileSize: file.size,
          contentType: file.type
        })
      });

      if (!presignedRes.ok) {
        const errorData = await presignedRes.json();
        throw new Error(errorData.message || 'Presigned URL 생성 실패');
      }

      const { uploadUrl, fileKey } = await presignedRes.json();

      // 2. S3 업로드
      const uploadRes = await fetch(uploadUrl, {
        method: 'PUT',
        headers: { 'Content-Type': file.type },
        body: file
      });

      if (!uploadRes.ok) {
        throw new Error('S3 업로드 실패');
      }

      // 3. 프로필 업데이트
      const updateRes = await fetch('/api/profiles/me/image', {
        method: 'PUT',
        headers: {
          'Authorization': `Bearer ${accessToken}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          workspaceId,
          fileKey
        })
      });

      if (!updateRes.ok) {
        const errorData = await updateRes.json();
        throw new Error(errorData.message || '프로필 업데이트 실패');
      }

      const profile = await updateRes.json();
      onUploadComplete(profile);

    } catch (err) {
      setError(err.message);
      console.error('Upload error:', err);
    } finally {
      setUploading(false);
    }
  };

  return (
    <div className="profile-image-upload">
      <div className="preview">
        {preview && <img src={preview} alt="Preview" />}
      </div>
      
      <input
        type="file"
        accept="image/jpeg,image/png,image/gif,image/webp"
        onChange={handleFileChange}
        disabled={uploading}
      />
      
      {uploading && <p>업로드 중...</p>}
      {error && <p className="error">{error}</p>}
    </div>
  );
}

export default ProfileImageUpload;
```

### cURL 예제

```bash
#!/bin/bash

# 환경 변수 설정
ACCESS_TOKEN="your_access_token"
WORKSPACE_ID="abc123"
FILE_PATH="avatar.jpg"
FILE_SIZE=$(wc -c < "$FILE_PATH")
CONTENT_TYPE="image/jpeg"

# 1. Presigned URL 요청
echo "Step 1: Requesting presigned URL..."
PRESIGNED_RESPONSE=$(curl -s -X POST http://localhost:8081/api/profiles/me/image/presigned-url \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"workspaceId\": \"$WORKSPACE_ID\",
    \"fileName\": \"avatar.jpg\",
    \"fileSize\": $FILE_SIZE,
    \"contentType\": \"$CONTENT_TYPE\"
  }")

echo "Presigned response: $PRESIGNED_RESPONSE"

UPLOAD_URL=$(echo $PRESIGNED_RESPONSE | jq -r '.uploadUrl')
FILE_KEY=$(echo $PRESIGNED_RESPONSE | jq -r '.fileKey')

# 2. S3 업로드
echo "Step 2: Uploading to S3..."
curl -X PUT "$UPLOAD_URL" \
  -H "Content-Type: $CONTENT_TYPE" \
  --data-binary @"$FILE_PATH"

# 3. 프로필 업데이트
echo "Step 3: Updating profile..."
UPDATE_RESPONSE=$(curl -s -X PUT http://localhost:8081/api/profiles/me/image \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"workspaceId\": \"$WORKSPACE_ID\",
    \"fileKey\": \"$FILE_KEY\"
  }")

echo "Update response: $UPDATE_RESPONSE"
```

## 에러 처리

### 클라이언트 측 에러 처리

```javascript
async function uploadWithErrorHandling(file, workspaceId, accessToken) {
  try {
    // 파일 검증
    if (file.size > 20 * 1024 * 1024) {
      throw new Error('파일 크기는 20MB를 초과할 수 없습니다.');
    }

    const allowedTypes = ['image/jpeg', 'image/png', 'image/gif', 'image/webp'];
    if (!allowedTypes.includes(file.type)) {
      throw new Error('이미지 파일만 업로드 가능합니다.');
    }

    // Presigned URL 요청
    const presignedRes = await fetch('/api/profiles/me/image/presigned-url', {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${accessToken}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        workspaceId,
        fileName: file.name,
        fileSize: file.size,
        contentType: file.type
      })
    });

    if (!presignedRes.ok) {
      const error = await presignedRes.json();
      throw new Error(error.message || 'Presigned URL 생성 실패');
    }

    const { uploadUrl, fileKey } = await presignedRes.json();

    // S3 업로드
    const uploadRes = await fetch(uploadUrl, {
      method: 'PUT',
      headers: { 'Content-Type': file.type },
      body: file
    });

    if (!uploadRes.ok) {
      throw new Error('S3 업로드 실패');
    }

    // 프로필 업데이트
    const updateRes = await fetch('/api/profiles/me/image', {
      method: 'PUT',
      headers: {
        'Authorization': `Bearer ${accessToken}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({ workspaceId, fileKey })
    });

    if (!updateRes.ok) {
      const error = await updateRes.json();
      throw new Error(error.message || '프로필 업데이트 실패');
    }

    return await updateRes.json();

  } catch (error) {
    console.error('Upload failed:', error);
    
    // 사용자에게 친화적인 에러 메시지 표시
    if (error.message.includes('20MB')) {
      alert('파일이 너무 큽니다. 20MB 이하의 파일을 선택해주세요.');
    } else if (error.message.includes('이미지')) {
      alert('이미지 파일만 업로드 가능합니다.');
    } else if (error.message.includes('인증')) {
      alert('로그인이 필요합니다.');
    } else {
      alert('업로드에 실패했습니다. 다시 시도해주세요.');
    }
    
    throw error;
  }
}
```

## 에러 코드

| 에러 코드 | HTTP 상태 | 설명 |
|----------|----------|------|
| `FILE_TOO_LARGE` | 400 | 파일 크기가 20MB 초과 |
| `INVALID_FILE_TYPE` | 400 | 이미지 파일이 아님 |
| `VALIDATION_ERROR` | 400 | 요청 데이터 검증 실패 |
| `UNAUTHORIZED` | 401 | 인증되지 않은 사용자 |
| `NOT_FOUND` | 404 | 프로필을 찾을 수 없음 |
| `INTERNAL_ERROR` | 500 | 서버 내부 오류 |

## 보안 고려사항

1. **Presigned URL 만료**: URL은 5분 후 자동으로 만료됩니다
2. **파일 검증**: 파일 크기, 타입, 확장자를 모두 검증합니다
3. **인증 필수**: 모든 API 호출 시 JWT 토큰이 필요합니다
4. **HTTPS 사용**: 프로덕션 환경에서는 반드시 HTTPS를 사용해야 합니다

## 성능 최적화

### 이미지 압축

업로드 전에 클라이언트에서 이미지를 압축하면 업로드 속도를 향상시킬 수 있습니다:

```javascript
async function compressImage(file, maxWidth = 1024, quality = 0.8) {
  return new Promise((resolve) => {
    const reader = new FileReader();
    reader.onload = (e) => {
      const img = new Image();
      img.onload = () => {
        const canvas = document.createElement('canvas');
        let width = img.width;
        let height = img.height;

        if (width > maxWidth) {
          height = (height * maxWidth) / width;
          width = maxWidth;
        }

        canvas.width = width;
        canvas.height = height;

        const ctx = canvas.getContext('2d');
        ctx.drawImage(img, 0, 0, width, height);

        canvas.toBlob(
          (blob) => {
            resolve(new File([blob], file.name, {
              type: 'image/jpeg',
              lastModified: Date.now()
            }));
          },
          'image/jpeg',
          quality
        );
      };
      img.src = e.target.result;
    };
    reader.readAsDataURL(file);
  });
}

// 사용 예제
const originalFile = document.getElementById('fileInput').files[0];
const compressedFile = await compressImage(originalFile);
await uploadProfileImage(compressedFile, workspaceId, accessToken);
```

## 참고사항

- Presigned URL은 5분간 유효하므로, 생성 후 즉시 업로드해야 합니다
- 파일 업로드 실패 시 재시도하려면 1단계부터 다시 시작해야 합니다
- 프로필 이미지는 워크스페이스별로 관리됩니다
- 이전 프로필 이미지는 자동으로 교체됩니다 (S3에서 삭제되지 않음)

## 문제 해결

### Q: Presigned URL이 만료되었다는 에러가 발생합니다
A: Presigned URL은 5분간만 유효합니다. URL을 받은 후 즉시 업로드를 진행하세요.

### Q: S3 업로드는 성공했지만 프로필 업데이트가 실패합니다
A: fileKey가 올바른지 확인하세요. Presigned URL 응답에서 받은 fileKey를 그대로 사용해야 합니다.

### Q: CORS 에러가 발생합니다
A: S3 버킷의 CORS 설정을 확인하세요. 프론트엔드 도메인이 허용되어 있어야 합니다.

### Q: 파일 타입 에러가 발생합니다
A: 이미지 파일만 업로드 가능합니다. jpg, jpeg, png, gif, webp 형식을 사용하세요.
