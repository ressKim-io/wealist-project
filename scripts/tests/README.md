# 테스트 스크립트

## 통합 테스트

### Board Service 통합 테스트
```bash
./scripts/tests/test-board-integration.sh
```

전체 Board Service 기능을 테스트합니다:
- Workspace, Project 생성
- Custom Fields (Status, Priority, Tags)
- Board CRUD
- Field Values
- Comments
- 필터링

### User Service 테스트
```bash
./scripts/tests/test-user-service.sh
```

User Service API를 테스트합니다.

## 단위 테스트

### Board API 테스트
```bash
./scripts/tests/test-board-api.sh
```

### Fractional Indexing 테스트
```bash
./scripts/tests/test-fractional-indexing.sh
```

Board 순서 관리 알고리즘을 테스트합니다.

### User API 테스트
```bash
./scripts/tests/test-user-api.sh
```

## 전제조건

모든 테스트 실행 전:
1. Docker 컨테이너가 실행 중이어야 합니다
   ```bash
   ./docker/scripts/dev.sh up-d
   ```

2. 서비스가 healthy 상태여야 합니다
   ```bash
   docker ps
   ```

## 테스트 토큰 획득

```bash
./scripts/get_user_token.sh
```
