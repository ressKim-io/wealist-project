# Go 1.25 Migration Guide

이 문서는 Go 1.25로 업그레이드하면서 발생한 이슈와 해결 방법을 설명합니다.

## 📋 변경 사항 요약

### 1. 타입 체크 강화
Go 1.25는 타입 안전성이 크게 강화되었습니다.

#### ❌ 이전 코드 (Go 1.24 이하)
```go
// UUID를 정수와 직접 비교 (허용됨)
if member.RoleID != 1 && member.RoleID != 2 {
    return errors.New("권한 없음")
}
```

#### ✅ 새 코드 (Go 1.25)
```go
// Role 관계를 로드하여 Level로 비교
if member.Role == nil || member.Role.Level < 50 {
    return errors.New("권한 없음")
}
```

**변경 이유**: `RoleID`는 `uuid.UUID` 타입(`[16]byte`)이므로 정수와 직접 비교 불가

### 2. 미사용 Import 체크 강화
Go 1.25는 미사용 import를 더 엄격히 체크합니다.

#### ❌ 컴파일 에러
```go
import (
    "github.com/redis/go-redis/v9"  // 사용하지 않으면 에러
)
```

#### ✅ 해결
사용하지 않는 import는 제거하거나, 실제로 사용해야 합니다.

## 🔧 수정된 파일들

### Domain 계층
**`internal/domain/project_member.go`**
```go
type ProjectMember struct {
    // ... 기존 필드들
    Role *Role `gorm:"foreignKey:RoleID;references:ID" json:"role,omitempty"`  // 추가됨
}
```

### Repository 계층
**`internal/repository/project_repository.go`**
```go
func (r *projectRepository) FindMemberByUserAndProject(userID, projectID uuid.UUID) (*domain.ProjectMember, error) {
    var member domain.ProjectMember
    if err := r.db.Preload("Role").  // Role 정보 함께 로드
        Where("user_id = ? AND project_id = ?", userID, projectID).
        First(&member).Error; err != nil {
        return nil, err
    }
    return &member, nil
}
```

### Service 계층
**`internal/service/field_service.go`** (8곳 수정)
```go
// 변경 전
if member.RoleID != 1 && member.RoleID != 2 {
    return apperrors.New(apperrors.ErrCodeForbidden, "권한 없음", 403)
}

// 변경 후
if member.Role == nil || member.Role.Level < 50 {
    return apperrors.New(apperrors.ErrCodeForbidden, "권한 없음 (ADMIN 이상)", 403)
}
```

**`internal/service/field_value_service.go`**
- 미사용 redis import 제거

## 🎯 Role Level 정의

데이터베이스에 정의된 Role 레벨:

| Role | Level | 설명 |
|------|-------|------|
| OWNER | 100 | 프로젝트 소유자 (모든 권한) |
| ADMIN | 50 | 관리자 (관리 권한) |
| MEMBER | 10 | 일반 멤버 (기본 권한) |

### 권한 체크 패턴
```go
// ADMIN 이상 (ADMIN + OWNER)
if member.Role == nil || member.Role.Level < 50 {
    return errors.New("ADMIN 권한 필요")
}

// OWNER만
if member.Role == nil || member.Role.Level < 100 {
    return errors.New("OWNER 권한 필요")
}
```

## 🚀 Docker 빌드 프로세스 개선

### Dockerfile 최적화
```dockerfile
# go.mod/go.sum 동기화를 먼저 수행
COPY go.mod go.sum ./
RUN go mod tidy

# 그 다음 의존성 다운로드
RUN go mod download

# 마지막으로 소스 복사 후 빌드
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o board-api ./cmd/api
```

**순서가 중요한 이유**:
1. `go mod tidy` → Go 버전에 맞게 go.mod/go.sum 동기화
2. `go mod download` → 동기화된 의존성 다운로드
3. `go build` → 소스 빌드

## 📝 사전 체크 스크립트

Docker 빌드 전에 로컬에서 미리 점검:

```bash
./check-go1.25-compat.sh
```

### 체크 항목 (10가지)
1. ✅ Go 버전 확인
2. ✅ 미사용 import 검사
3. ✅ 타입 비교 이슈 검사
4. ✅ 컴파일 테스트
5. ✅ 코드 포맷팅 (gofmt)
6. ✅ 정적 분석 (go vet)
7. ✅ 모듈 의존성 검증
8. ✅ Go 1.25 breaking changes
9. ✅ 테스트 파일 컴파일
10. ✅ Dockerfile 버전 일관성

### 출력 예시
```
╔═══════════════════════════════════════════════════════════════╗
║  ✓ All checks passed! Ready for Docker build                 ║
╚═══════════════════════════════════════════════════════════════╝

You can now run:
  docker-compose build board-service
```

## 🐛 발생했던 주요 에러들

### 1. Type Conversion Error
```
cannot convert 2 (untyped int constant) to type [16]byte
```
**원인**: UUID와 정수 직접 비교
**해결**: Role.Level 기반 비교로 변경

### 2. Unused Import Error
```
"github.com/redis/go-redis/v9" imported as redis and not used
```
**원인**: Go 1.25의 엄격한 import 체크
**해결**: 미사용 import 제거

### 3. Module Updates Needed
```
updates to go.mod needed; to update it: go mod tidy
```
**원인**: go.mod/go.sum이 Go 1.25와 동기화되지 않음
**해결**: Dockerfile에서 `go mod tidy` 먼저 실행

## ✅ 검증 체크리스트

Go 1.25 마이그레이션 완료 전 확인 사항:

- [ ] `check-go1.25-compat.sh` 실행 → 모든 체크 통과
- [ ] Docker 빌드 성공
- [ ] 기존 테스트 통과
- [ ] API 엔드포인트 정상 작동
- [ ] 권한 체크 로직 정상 작동 확인

## 📚 참고 자료

- [Go 1.25 Release Notes](https://tip.golang.org/doc/go1.25)
- [Go 1.25 Blog Post](https://go.dev/blog/go1.25)
- [Official Download](https://go.dev/dl/)

## 💡 추가 참고사항

### 로컬 개발 환경
로컬에서 Go 1.25가 없어도 Docker에서 빌드 가능합니다:
```bash
# Docker가 Go 1.25를 사용하므로 로컬 Go 버전 무관
docker-compose build board-service
```

### 새로운 기능 활용
Go 1.25의 새로운 기능들:
- Container-aware GOMAXPROCS (cgroup CPU 제한 인식)
- DWARF v5 디버그 정보 (공간 절약)
- JSON v2 experimental 패키지 (`encoding/json/v2`)

---

**마이그레이션 완료일**: 2025-01-09
**Go 버전**: 1.25.0 → 1.25.4 (현재 최신)
