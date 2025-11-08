# Changelog

> Wealist Board Service 변경 이력
> 시작일: 2025-11-08

---

## [v0.3.0] - 2025-11-08

### ✨ Added - Custom Field Management System

#### FilterBar Component
- 검색, 뷰 전환, 필터, 관리 기능이 통합된 상단 바 추가
- Stage/Role 기준 뷰 전환 UI
- 필터 옵션: 전체, 내가 담당한 것, 중요도 높음, 긴급, 완료된 것 숨기기

#### CustomFieldManageModal
- Stage, Role, Importance 관리를 위한 탭 형식 모달
- 12가지 색상 팔레트 제공
- 생성, 수정, 삭제 기능 완전 구현
- 시스템 기본값 삭제 방지
- Importance Level(1-5) 설정 지원

#### Custom Fields CRUD API
**boardService.ts에 추가**:
- `createStage`, `updateStage`, `deleteStage`
- `createRole`, `updateRole`, `deleteRole`
- `createImportance`, `updateImportance`, `deleteImportance`

**파일 변경**:
- `frontend/src/components/FilterBar.tsx` (NEW)
- `frontend/src/components/modals/CustomFieldManageModal.tsx` (NEW)
- `frontend/src/api/board/boardService.ts` (+204 lines)
- `frontend/src/pages/Dashboard.tsx` (FilterBar 통합)

---

## [v0.2.0] - 2025-11-08

### ✨ Added - Color Persistence System

#### 색상 팔레트 시스템
- 12가지 미리 정의된 색상 팔레트 생성
- `getDefaultColorByIndex()` - 인덱스 기반 기본 색상 할당
- `getColorByHex()` - Hex 값으로 색상 검색

#### Column 색상 관리 개선
- **Before**: 위치(idx) 기반으로 색상 할당 → 순서 바뀌면 색상도 변경
- **After**: API의 `stage.color` 사용 → 순서 바뀌어도 색상 유지
- Column 인터페이스에 `color?: string` 필드 추가

**파일 변경**:
- `frontend/src/constants/colors.ts` (NEW)
- `frontend/src/pages/Dashboard.tsx`

---

## [v0.1.0] - 2025-11-08

### ✨ Added - Drag & Drop Visual Feedback

#### Cross-Column Drop Indicator
- 보드를 다른 컬럼으로 드래그할 때 컬럼 하단에 드롭 인디케이터 표시
- "여기에 추가됩니다" 텍스트 + 파란색 펄스 라인
- 다른 컬럼으로 이동하는 경우에만 표시

#### Same-Column Drop Indicator
- 같은 컬럼 내에서 보드 순서 변경 시 드롭 위치 표시
- 대상 보드 위에 파란색 펄스 라인
- `mt-3` 여백으로 삽입 위치 명확하게 표시

#### Dragged Item Opacity
- 드래그 중인 항목: `opacity-80` (기존 `opacity-50`에서 개선)
- 더 선명하게 보여 사용자 경험 개선

**파일 변경**:
- `frontend/src/pages/Dashboard.tsx`
  - `dragOverColumn` state 추가
  - Drop position indicator 렌더링

---

## [v0.0.5] - 2025-11-08

### ✨ Added - User Order API Integration

#### Drag & Drop Persistence
- Stage 컬럼 순서 변경 저장 (드래그로 컬럼 이동)
- 같은 컬럼 내 보드 순서 변경 저장 (세로 드래그)
- 다른 컬럼으로 보드 이동 시 Stage 변경 저장

#### API 추가 (boardService.ts)
- `updateStageColumnOrder` - Stage 컬럼 순서 저장
- `updateStageBoardOrder` - Stage 내 보드 순서 저장

#### Drag Handlers
- `handleColumnDragStart`, `handleColumnDrop` - 컬럼 드래그
- `handleDrop` 개선 - 같은 컬럼 / 다른 컬럼 분기 처리

**파일 변경**:
- `frontend/src/api/board/boardService.ts` (+62 lines)
- `frontend/src/pages/Dashboard.tsx`
  - Column drag handlers 추가
  - Same-column board reordering 로직 추가

---

## [v0.0.4] - 2025-11-08

### ✨ Added - Board Detail & Comment Integration

#### BoardDetailModal (완전 재작성)
- 보드 상세 정보 표시
- 인라인 편집 모드 (제목, 내용, Custom Fields)
- 보드 삭제 (확인 다이얼로그)
- 실시간 댓글 표시 및 작성
- Custom Fields 선택 (Stages, Roles, Importances)
- 2열 레이아웃 (2/3 메인 + 1/3 사이드바)

#### Comment API
**boardService.ts에 추가**:
- `getComments` - 댓글 목록 조회
- `createComment` - 댓글 작성
- `updateComment` - 댓글 수정
- `deleteComment` - 댓글 삭제

#### Dashboard 개선
- `selectedBoard` → `selectedBoardId` (ID만 전달)
- `onBoardUpdated`, `onBoardDeleted` 콜백으로 목록 새로고침

**파일 변경**:
- `frontend/src/components/modals/BoardDetailModal.tsx` (완전 재작성)
- `frontend/src/api/board/boardService.ts` (+90 lines)
- `frontend/src/pages/Dashboard.tsx`

---

## [v0.0.3] - 2025-11-08

### 🐛 Fixed - Empty Project Display

#### 문제
- 보드가 없는 프로젝트 선택 시 빈 화면만 표시
- 컬럼이 없어서 "보드 추가" 버튼도 안 보임

#### 해결
- **Before**: 보드에서 Stage를 추출하여 컬럼 생성 → 보드 없으면 컬럼도 없음
- **After**: 프로젝트의 모든 Stage를 먼저 조회 → 빈 컬럼 먼저 생성 → 보드 추가

#### 로직 변경
```typescript
// 1. 모든 Stages 조회
const stages = await getProjectStages(projectId, token);

// 2. 빈 컬럼 먼저 생성
const stageMap = new Map();
stages.forEach(stage => {
  stageMap.set(stage.id, { stage, boards: [] });
});

// 3. 보드 추가
boards.forEach(board => {
  stageMap.get(board.stage.id).boards.push(board);
});

// 4. displayOrder로 정렬
const sorted = Array.from(stageMap.values())
  .sort((a, b) => a.stage.displayOrder - b.stage.displayOrder);
```

**파일 변경**:
- `frontend/src/pages/Dashboard.tsx`

---

## [v0.0.2] - 2025-11-08

### 🐛 Fixed - Project List Loading

#### 문제 1: Query Parameter Mismatch
- Backend: `workspace_id` (snake_case) 기대
- Frontend: `workspaceId` (camelCase) 전송
- **결과**: 프로젝트 목록 안 불러와짐

#### 문제 2: Response Structure
- Backend: `{ data: { projects: [...] } }`
- Frontend: `response.data.data` (배열 기대, 객체 받음)
- **결과**: 프로젝트 목록 파싱 실패

#### 해결
```typescript
// Before
params: { workspaceId }
return response.data.data || [];

// After
params: { workspace_id: workspaceId }
return response.data.data?.projects || [];
```

**파일 변경**:
- `frontend/src/api/board/boardService.ts`

---

## [v0.0.1] - 2025-11-08 (이전 세션에서 이어짐)

### ✨ Added - Frontend Board API Integration

#### Project Creation
- `CreateProjectModal` 컴포넌트 생성
- `createProject` API 연동
- Dashboard에서 프로젝트 생성 플로우 통합

#### Board Creation
- `CreateBoardModal` 컴포넌트 생성
- Custom Fields (Stages, Roles) 자동 조회
- Stage 및 Role 선택 UI
- "보드 추가" 버튼에서 현재 Stage 자동 선택

#### Dashboard 개선
- `fetchProjects`, `fetchBoards`를 useCallback으로 메모이제이션
- 프로젝트 선택 시 자동으로 보드 로드
- 보드 클릭 시 상세 모달 표시

**파일 변경**:
- `frontend/src/components/modals/CreateProjectModal.tsx` (NEW)
- `frontend/src/components/modals/CreateBoardModal.tsx` (NEW)
- `frontend/src/pages/Dashboard.tsx`
- `frontend/src/api/board/boardService.ts`

---

## [Backend v0.2.0] - 2025-11-08 (이전 세션)

### 🚀 Performance - N+1 Query Optimization

#### GetBoards API 최적화
- **Before**: 84 queries (20개 보드 조회 시)
- **After**: 64 queries (24% 감소)

#### 최적화 내역

##### Custom Fields 배치 조회
- `FindStagesByIDs()` 추가 (20 → 1 query)
- `FindRolesByIDs()` 추가 (20 → 1 query)
- `FindImportancesByIDs()` 추가 (20 → 1 query)

##### Assignee 배치 조회
- Redis MGET로 일괄 조회 (20 → 1 Redis command)
- `getUserProfilesBatch()` 구현

##### BoardRoles 배치 조회
- `FindRolesByBoards()` 추가 (20 → 1 query)

**파일 변경**:
- `board-service/internal/repository/board_repository.go`
- `board-service/internal/service/board_service.go`

---

## 문서화

### 추가된 문서
- `FRONTEND_IMPLEMENTATION_GUIDE.md` - Frontend 구현 가이드
- `BACKEND_OPTIMIZATION_GUIDE.md` - Backend 최적화 가이드
- `BOARD_SERVICE_API_REFERENCE.md` - Board Service API 레퍼런스 (이전 작성)
- `CHANGELOG.md` - 변경 이력 (이 문서)

---

## 향후 계획

### 검색 및 필터링 로직 구현
- [ ] 보드 제목/내용 검색
- [ ] 담당자 필터
- [ ] 중요도 필터
- [ ] 완료된 항목 숨기기

### Role 기반 뷰
- [ ] Role 기준 컬럼 렌더링
- [ ] Role 드래그 앤 드롭
- [ ] Role User Order API 통합

### Project 관리 확장
- [ ] 프로젝트 수정 (PUT /api/projects/{id})
- [ ] 프로젝트 삭제 (DELETE /api/projects/{id})

### Assignee 및 Due Date
- [ ] Assignee 선택 UI (User Service 연동)
- [ ] Due Date 달력 UI
- [ ] 기한 임박 알림

### 추가 최적화
- [ ] 검색 디바운싱
- [ ] 무한 스크롤
- [ ] Virtual List (긴 목록 최적화)
- [ ] Backend 쿼리 추가 최적화

---

## 기여자
- Claude (AI Assistant)
- ressKim (Project Owner)

---

## 라이센스
Private Project
