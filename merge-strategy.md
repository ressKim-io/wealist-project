# Merge Strategy: Keep Current Version

## 1. 머지 시작
```bash
git checkout claude/check-merge-safety-011CV1X7tHJ9zyFJ8rSoVTbj
git merge upstream/main
# 충돌 발생 (예상됨)
```

## 2. 충돌 해결 - 현재 버전 우선

### Board-service 전체 (우리 버전으로 덮기)
```bash
git checkout --ours board-service/
git add board-service/
```

### User-service 전체 (우리 버전으로 덮기)  
```bash
git checkout --ours user-service/
git add user-service/
```

### Frontend 전체 (우리 버전으로 덮기)
```bash
git checkout --ours frontend/
git add frontend/
```

### Docker 설정 (우리 버전으로 덮기)
```bash
git checkout --ours docker-compose.*
git add docker-compose.*
```

## 3. 삭제된 파일 처리
```bash
# upstream에서 수정했지만 우리가 삭제한 파일들 확인
git status | grep "deleted by us"

# 삭제 확정
git rm board-service/internal/dto/custom_field.go
git rm board-service/internal/dto/user_order.go
git rm board-service/internal/handler/custom_field_handler.go
git rm board-service/internal/handler/user_order_handler.go
git rm board-service/internal/repository/custom_field_repository.go
git rm board-service/migrations/20250106120000_baseline_v1.0.0.*
git rm docker-compose.base.yml
git rm docker-compose.yaml
```

## 4. 머지 커밋
```bash
git commit -m "Merge upstream/main (keep current implementation)"
```

## 5. 검증
```bash
# 주요 파일들이 우리 버전인지 확인
git show HEAD:user-service/src/main/java/OrangeCloud/UserRepo/entity/Workspace.java | head -30
git show HEAD:frontend/src/api/apiConfig.ts | head -20
git show HEAD:frontend/src/contexts/AuthContext.tsx | grep "userId"
```
