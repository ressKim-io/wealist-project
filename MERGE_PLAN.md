# Merge Plan: Upstream User-Service + Current Board-Service

## 전략
- **board-service**: 현재 버전 유지 (ours)
- **user-service**: upstream 버전 채택 (theirs)
- **frontend**: 필드명만 수정하면 OK

## 변경 사항

### User-Service (자동 - upstream 채택)
```
groupId → workspaceId
name → workspaceName  
companyName → workspaceDescription
+ isPublic
+ needApproved
```

### Frontend (수동 수정 필요)
**파일 6개, 예상 30-40곳**

#### 1. frontend/src/api/user/userService.ts
```typescript
interface WorkspaceResponse {
  workspaceId: string;  // was: id
  workspaceName: string; // was: name
  workspaceDescription: string; // was: description
  isPublic: boolean; // NEW
  needApproved: boolean; // NEW
  ownerId: string;
  ownerName: string;
  ownerEmail: string;
  createdAt: string;
}
```

#### 2. frontend/src/types/index.ts
```typescript
export interface Workspace {
  workspaceId: number;
  workspaceName: string;
  workspaceDescription?: string;
  // ...
}
```

#### 3-6. 컴포넌트 수정
- SelectWorkspacePage.tsx: workspace.name → workspace.workspaceName
- Dashboard.tsx: 같은 패턴
- WorkspaceManagementModal.tsx: 같은 패턴  
- ProjectModal.tsx: workspace_id는 그대로 (API 파라미터)

## Merge 실행 순서

```bash
# 1. upstream 머지 (user-service는 theirs, board-service는 ours)
git merge upstream/main --no-commit --no-ff

# 2. board-service 우리 버전으로 덮기
git checkout --ours board-service/
git add board-service/

# 3. 삭제된 파일 정리
git rm board-service/internal/dto/custom_field.go
git rm board-service/internal/dto/user_order.go
git rm board-service/internal/handler/custom_field_handler.go
git rm board-service/internal/handler/user_order_handler.go  
git rm board-service/internal/repository/custom_field_repository.go
git rm board-service/migrations/20250106120000_baseline_v1.0.0.*
git rm docker-compose.base.yml docker-compose.yaml

# 4. Frontend 필드명 수정 (찾아바꾸기)
# - workspace.id → workspace.workspaceId
# - workspace.name → workspace.workspaceName
# - workspace.description → workspace.workspaceDescription

# 5. 커밋
git commit -m "Merge upstream/main: adopt user-service, keep board-service"

# 6. 푸시
git push -u origin claude/check-merge-safety-011CV1X7tHJ9zyFJ8rSoVTbj
```

## 예상 작업 시간
- Merge 실행: 10분
- Frontend 수정: 2-3시간
- 테스트: 1-2시간
- **총: 반나절 이내**

## 장점
✅ User-Service의 최신 기능 활용 (isPublic, needApproved)
✅ Board-Service는 현재 구조 유지
✅ Frontend 수정 범위 작음
✅ 데이터베이스는 upstream 스키마 사용
