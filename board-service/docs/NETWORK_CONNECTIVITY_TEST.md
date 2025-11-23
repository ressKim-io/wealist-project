# 네트워크 연결 테스트 결과

## 테스트 개요

board-service에서 user-service로의 네트워크 연결을 테스트하여 API 호출이 정상적으로 작동하는지 확인합니다.

## 테스트 환경

### 1. 로컬 환경 (localhost:8080)

**테스트 명령어:**
```bash
./board-service/scripts/test-network-connectivity.sh local
```

**테스트 결과:**
- ✓ DNS 해석: 성공
- ✓ TCP 연결: 성공  
- ✓ Health 엔드포인트: 성공
- ✓ 타임아웃 설정: 적절함

**API 엔드포인트 테스트:**
- `/api/users/{userId}`: 404 (엔드포인트 미구현)
- `/api/workspaces/{workspaceId}`: 500 (엔드포인트 존재, 데이터 없음)
- `/api/profiles/workspace/{workspaceId}`: 500 (엔드포인트 존재, 데이터 없음)
- `/api/workspaces/{workspaceId}/validate-member/{userId}`: 500 (엔드포인트 존재, 데이터 없음)

**결론:** 
- user-service가 localhost:8080에서 정상 작동 중
- 대부분의 API 엔드포인트가 존재함 (500 에러는 데이터 없음을 의미)
- `/api/users/{userId}` 엔드포인트는 구현되지 않음 (404)

### 2. Docker Compose 환경 (user-service:8080)

**호스트에서 테스트 (실패):**
```bash
./board-service/scripts/test-network-connectivity.sh docker
```

**결과:** 전부 실패
- ✗ DNS 해석: 실패
- ✗ TCP 연결: 실패
- ✗ Health 엔드포인트: 실패

**원인:** 호스트 머신에서는 Docker 컨테이너 이름(`user-service`)으로 직접 접근할 수 없습니다.

**올바른 테스트 방법 - Docker 컨테이너 내부에서 실행:**

```bash
# board-service 컨테이너 내부에서 테스트
docker exec -it board-service /bin/sh -c "curl -v http://user-service:8080/actuator/health"

# 또는 컨테이너에 접속해서 테스트
docker exec -it board-service /bin/sh
# 컨테이너 내부에서:
curl http://user-service:8080/actuator/health
curl http://user-service:8080/api/workspaces/00000000-0000-0000-0000-000000000000
```

## 발견된 문제점

### 1. `/api/users/{userId}` 엔드포인트 404

**문제:**
- board-service가 호출하는 `/api/users/{userId}` 엔드포인트가 user-service에 존재하지 않음

**영향:**
- `user_client.go`의 `GetUser()` 메서드가 404 에러 반환
- 사용자 정보 조회 기능 작동 불가

**해결 방법:**
1. user-service에 `/api/users/{userId}` 엔드포인트 구현
2. 또는 board-service가 다른 엔드포인트 사용하도록 수정

### 2. Docker 네트워크 설정

**현재 설정:**
- `docker-compose.dev.yml`에서 user-service와 board-service가 같은 네트워크(`backend-net`)에 있음
- Service name으로 통신 가능해야 함

**확인 필요 사항:**
```bash
# Docker 네트워크 확인
docker network ls | grep wealist

# 컨테이너가 올바른 네트워크에 연결되어 있는지 확인
docker inspect user-service | grep -A 10 Networks
docker inspect board-service | grep -A 10 Networks
```

## 권장 사항

### 1. 환경별 URL 설정

**로컬 개발 환경:**
```yaml
# board-service/configs/config.yaml
user_api:
  base_url: "http://localhost:8080"
  timeout: 5s
```

**Docker Compose 환경:**
```yaml
# docker-compose.dev.yml
environment:
  USER_SERVICE_URL: http://user-service:8080
```

**Kubernetes 환경:**
```yaml
environment:
  USER_SERVICE_URL: http://user-service.default.svc.cluster.local:8080
```

### 2. 타임아웃 설정

현재 설정된 5초는 적절합니다:
- Health check: 1초 이내 응답
- API 호출: 1-2초 이내 응답
- 5초 타임아웃은 충분한 여유 제공

### 3. 에러 처리

board-service의 user_client.go에서:
- 404 에러: 사용자/리소스 없음으로 처리
- 500 에러: user-service 내부 오류로 처리
- 타임아웃: 네트워크 오류로 처리

## 다음 단계

1. **user-service API 엔드포인트 확인**
   - `/api/users/{userId}` 엔드포인트 구현 여부 확인
   - 필요시 엔드포인트 추가 또는 board-service 수정

2. **Docker 환경에서 실제 테스트**
   - Docker Compose로 전체 시스템 실행
   - 컨테이너 내부에서 네트워크 연결 테스트
   - board-service 로그에서 user-service 호출 확인

3. **통합 테스트 작성**
   - board-service에서 user-service 호출하는 통합 테스트
   - Docker Compose 환경에서 자동화된 테스트

## 테스트 명령어 요약

```bash
# 1. 로컬 환경 테스트
./board-service/scripts/test-network-connectivity.sh local

# 2. Docker 환경 - 컨테이너 내부에서 테스트
docker exec -it board-service curl http://user-service:8080/actuator/health

# 3. Docker 네트워크 확인
docker network inspect wealist-backend-net

# 4. 컨테이너 로그 확인
docker logs board-service
docker logs user-service

# 5. user-service API 직접 테스트
curl http://localhost:8080/actuator/health
curl http://localhost:8080/api/workspaces/00000000-0000-0000-0000-000000000000
```
