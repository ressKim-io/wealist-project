# Board Service Test Scripts

board-service를 테스트하기 위한 스크립트 모음입니다.

## 📋 스크립트 목록

### 1. integration-test.sh
**완전한 통합 테스트 스크립트 (권장)**

실제 user-service 로그인 API를 사용하여 board-service의 주요 엔드포인트를 테스트합니다.

```bash
./board-service/scripts/integration-test.sh
```

**테스트 항목:**
- ✅ user-service와 board-service health check
- ✅ 실제 로그인 API를 통한 JWT 토큰 발급
- ✅ 사용자 workspace 정보 조회
- ✅ GET /api/projects (프로젝트 목록 조회)
- ✅ POST /api/projects (프로젝트 생성)
- ✅ GET /api/projects/{id} (프로젝트 상세 조회)
- ✅ GET /api/projects/{id}/boards (보드 목록 조회)
- ✅ 서비스 간 통신 및 인증 플로우 검증

**환경 변수:**
```bash
BOARD_API_URL=http://localhost:8000 \
USER_API_URL=http://localhost:8080 \
TEST_USER_EMAIL=test@example.com \
TEST_USER_PASSWORD=password123 \
./board-service/scripts/integration-test.sh
```

> **참고**: 이전 버전의 mock 기반 테스트 스크립트(`test-with-mock-user.sh`)는 참고용으로 `.deprecated` 확장자로 보관되어 있습니다.

---

### 2. quick-test.sh
**빠른 단일 엔드포인트 테스트**

특정 엔드포인트만 빠르게 테스트할 때 사용합니다.

```bash
# 프로젝트 목록 조회
./board-service/scripts/quick-test.sh projects

# 프로젝트 생성
./board-service/scripts/quick-test.sh create-project

# 보드 목록 조회
./board-service/scripts/quick-test.sh boards

# user-service 연결 확인
./board-service/scripts/quick-test.sh user-check
```

---

### 3. test-500-error-fix.sh
**500 에러 수정 검증 스크립트**

특정 버그 수정을 검증하기 위한 스크립트입니다.

```bash
./board-service/scripts/test-500-error-fix.sh
```

---

### 4. test-network-connectivity.sh
**네트워크 연결 테스트**

Docker 네트워크 설정과 서비스 간 통신을 확인합니다.

```bash
./board-service/scripts/test-network-connectivity.sh
```

---

## 🧪 테스트 유저 정보

user-service의 `DataInitializer`가 생성하는 테스트 데이터:

- **유저**: user1@example.com ~ user50@example.com (총 50명)
- **워크스페이스**: 10개 (각 5명씩 배정)
- **기본 테스트 유저**: user1@example.com

### 테스트 유저 구조
```
Workspace 1 (테스트 워크스페이스 1)
├── user1@example.com (OWNER)
├── user2@example.com (MEMBER)
├── user3@example.com (MEMBER)
├── user4@example.com (MEMBER)
└── user5@example.com (MEMBER)

Workspace 2 (테스트 워크스페이스 2)
├── user6@example.com (OWNER)
├── user7@example.com (MEMBER)
...
```

---

## 🔧 문제 해결

### 1. "Failed to connect to user-service"
```bash
# user-service가 실행 중인지 확인
docker ps | grep user-service

# user-service 시작
docker-compose up -d user-service

# 로그 확인
docker logs user-service -f
```

### 2. "Failed to get user ID"
```bash
# 테스트 데이터가 초기화되었는지 확인
curl http://localhost:8080/api/users/email/user1@example.com

# 데이터가 없다면 user-service 재시작 (dev/local 프로파일에서 자동 초기화)
docker-compose restart user-service
```

### 3. board-service가 user-service를 호출할 수 없음
```bash
# Docker 네트워크 확인
docker network inspect wealist-network

# board-service에서 user-service로 ping 테스트
docker exec board-service ping user-service

# 네트워크 연결 테스트 스크립트 실행
./board-service/scripts/test-network-connectivity.sh
```

### 4. 500 에러 발생
```bash
# board-service 로그 확인
docker logs board-service -f

# user-service 로그 확인
docker logs user-service -f

# 설정 파일 확인
./board-service/scripts/verify-config.sh
```

---

## 📝 수동 테스트 예제

### 1. 테스트 유저 정보 가져오기
```bash
curl http://localhost:8080/api/users/email/user1@example.com
```

### 2. 유저의 워크스페이스 목록
```bash
USER_ID="<user-id-from-step-1>"
curl http://localhost:8080/api/users/${USER_ID}/workspaces
```

### 3. 프로젝트 목록 조회
```bash
USER_ID="<user-id>"
WORKSPACE_ID="<workspace-id>"
curl "http://localhost:8000/api/projects?workspaceId=${WORKSPACE_ID}" \
  -H "X-User-ID: ${USER_ID}"
```

### 4. 프로젝트 생성
```bash
USER_ID="<user-id>"
WORKSPACE_ID="<workspace-id>"
curl -X POST http://localhost:8000/api/projects \
  -H "Content-Type: application/json" \
  -H "X-User-ID: ${USER_ID}" \
  -d '{
    "workspaceId": "'${WORKSPACE_ID}'",
    "projectName": "My Test Project",
    "projectDescription": "Test project description",
    "projectColor": "#3B82F6"
  }'
```

---

## 🚀 CI/CD에서 사용

GitHub Actions에서 테스트 스크립트를 실행하는 예제:

```yaml
- name: Run integration tests
  run: |
    # Wait for services to be ready
    sleep 10
    
    # Run integration test
    ./board-service/scripts/test-with-mock-user.sh
```

---

## 📚 관련 문서

- [board-service README](../README.md)
- [User Service API](../../_README/USER_SERVICE.md)
- [Network Connectivity Test](../docs/NETWORK_CONNECTIVITY_TEST.md)
- [Manual Test Guide](../MANUAL_TEST_GUIDE.md)
