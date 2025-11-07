# Board Service API Test Scripts

Board Service API를 테스트하기 위한 스크립트 모음입니다.

## 📋 사전 요구사항

- User Service 실행 중 (`http://localhost:8080`)
- Board Service 실행 중 (`http://localhost:8000`)
- PostgreSQL 데이터베이스 실행 중
- `curl` 및 `jq` 설치되어 있어야 함

```bash
# macOS
brew install jq

# Ubuntu/Debian
sudo apt-get install jq

# CentOS/RHEL
sudo yum install jq
```

## 🚀 사용 방법

### 1단계: User Service에서 토큰 받아오기

```bash
# User Service에 로그인하여 JWT 토큰 받기
cd scripts
./get_user_token.sh <your-email> <your-password>

# 예시:
./get_user_token.sh test@example.com password123
```

출력된 환경 변수를 복사해서 실행:

```bash
export JWT_TOKEN='eyJhbGciOiJIUzUxMiJ9...'
export USER_ID='123e4567-e89b-12d3-a456-426614174000'
export WORKSPACE_ID='987fcdeb-51a2-43f7-8b9c-123456789abc'
```

### 2단계: Board Service API 테스트

```bash
# 환경 변수가 설정된 상태에서 테스트 스크립트 실행
./test_board_api.sh
```

## 📝 테스트 시나리오

`test_board_api.sh` 스크립트는 다음 순서로 API를 테스트합니다:

1. ✅ **Health Check** - Board Service 상태 확인
2. ✅ **Create Project** - 새 프로젝트 생성
3. ✅ **Get Projects** - 워크스페이스 내 프로젝트 조회
4. ✅ **Get Project Details** - 프로젝트 상세 정보 조회
5. ✅ **Create Custom Role** - 커스텀 역할 생성 (예: Backend Developer)
6. ✅ **Create Custom Stage** - 커스텀 진행단계 생성 (예: In Progress)
7. ✅ **Create Custom Importance** - 커스텀 중요도 생성 (예: High Priority)
8. ✅ **Get Custom Fields** - 모든 커스텀 필드 조회
9. ✅ **Create Board** - 보드(카드) 생성
10. ✅ **Get Boards** - 프로젝트 내 보드 조회
11. ✅ **Create Comment** - 보드에 댓글 작성
12. ✅ **Get Comments** - 보드의 댓글 조회
13. ✅ **Role-Based Board View** - 역할 기반 보드 뷰 조회
14. ✅ **Stage-Based Board View** - 진행단계 기반 보드 뷰 조회

## 🔧 환경 변수 설정

### 필수 환경 변수

```bash
export JWT_TOKEN='your-jwt-token-from-user-service'
export USER_ID='your-user-uuid'
export WORKSPACE_ID='your-workspace-uuid'
```

### 선택 환경 변수

```bash
# Board Service URL (기본값: http://localhost:8000)
export BOARD_SERVICE_URL='http://localhost:8000'

# User Service URL (기본값: http://localhost:8080)
export USER_SERVICE_URL='http://localhost:8080'
```

## 📊 예시 출력

```
=================================================
1. Health Check (No Auth Required)
=================================================
✓ Health check passed
{
  "status": "healthy",
  "database": "connected",
  "redis": "connected"
}

=================================================
2. Create Project
=================================================
✓ Project created: 123e4567-e89b-12d3-a456-426614174000
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "name": "Test Project 1704096000",
  "description": "Test project created by test script",
  "workspaceId": "987fcdeb-51a2-43f7-8b9c-123456789abc",
  "ownerId": "456e7890-e12b-34d5-a678-901234567890"
}
```

## ⚠️ 중요 사항

### Workspace 검증

Board Service는 프로젝트 생성 시 User Service의 `/api/workspace` API를 호출하여 다음을 검증합니다:

1. **Workspace 존재 여부** - `CheckWorkspaceExists()`
2. **사용자 멤버십** - `ValidateWorkspaceMembership()`

따라서 **User Service가 반드시 실행 중이어야** 합니다. User Service가 중단된 상태에서는 프로젝트 생성이 실패합니다.

### 에러 발생 시

```bash
# Board Service 로그 확인
docker logs board-service

# User Service 로그 확인
docker logs user-service

# 데이터베이스 연결 확인
psql -U board_service -d wealist_board_db -c "SELECT version();"
```

## 🧪 개발 모드에서 테스트

```bash
# Docker Compose로 모든 서비스 실행
cd /home/user/wealist-project
docker-compose up -d

# 서비스 상태 확인
docker-compose ps

# 테스트 실행
cd scripts
./get_user_token.sh test@example.com password123
# ... export 명령어 실행 ...
./test_board_api.sh
```

## 📝 수동 API 테스트

개별 API를 수동으로 테스트하려면:

```bash
# JWT 토큰 설정
TOKEN="your-jwt-token"
WORKSPACE_ID="your-workspace-id"

# Health Check
curl http://localhost:8000/health | jq '.'

# Create Project
curl -X POST http://localhost:8000/api/projects \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"workspaceId\": \"$WORKSPACE_ID\",
    \"name\": \"My Test Project\",
    \"description\": \"Testing Board API\"
  }" | jq '.'

# Get Projects
curl "http://localhost:8000/api/projects?workspace_id=$WORKSPACE_ID" \
  -H "Authorization: Bearer $TOKEN" | jq '.'
```

## 🐛 트러블슈팅

### 1. "JWT_TOKEN is not set" 에러

```bash
# 토큰이 설정되지 않았습니다
./get_user_token.sh <email> <password>
# 출력된 export 명령어를 실행하세요
```

### 2. "Workspace validation failed" 에러

```bash
# User Service가 실행 중인지 확인
curl http://localhost:8080/actuator/health

# Workspace가 존재하는지 확인
curl http://localhost:8080/api/workspace \
  -H "Authorization: Bearer $TOKEN" | jq '.'
```

### 3. "Database connection failed" 에러

```bash
# PostgreSQL 실행 확인
docker ps | grep postgres

# 데이터베이스 연결 확인
psql -U board_service -d wealist_board_db
```

## 📚 참고 자료

- [Board Service API 문서](http://localhost:8000/swagger/index.html) (개발 모드)
- [User Service API 문서](http://localhost:8080/swagger-ui/index.html)
- [프로젝트 README](../README.md)
