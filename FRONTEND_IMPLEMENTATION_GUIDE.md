# Frontend Implementation Guide

> 최종 업데이트: 2025-11-08
> 프로젝트: Wealist Board Service Frontend

## 📋 목차
1. [개요](#개요)
2. [API 연동 현황](#api-연동-현황)
3. [드래그 앤 드롭 시스템](#드래그-앤-드롭-시스템)
4. [커스텀 필드 관리](#커스텀-필드-관리)
5. [필터 및 검색](#필터-및-검색)
6. [색상 시스템](#색상-시스템)
7. [향후 작업](#향후-작업)

---

## 개요

Wealist Board Service의 Frontend는 React + TypeScript + Vite로 구성되어 있으며,
Board Service API와 User Service API를 활용하여 칸반 보드 시스템을 구현합니다.

### 기술 스택
- **Framework**: React 19
- **Language**: TypeScript
- **Build Tool**: Vite
- **HTTP Client**: Axios
- **Styling**: Tailwind CSS
- **Icons**: Lucide React

---

## API 연동 현황

### ✅ 완료된 API 연동

#### 1. Project API
- `GET /api/projects?workspace_id={id}` - 프로젝트 목록 조회
- `POST /api/projects` - 프로젝트 생성
- **구현 위치**: `frontend/src/api/board/boardService.ts`
- **사용 컴포넌트**: `Dashboard.tsx`, `CreateProjectModal.tsx`

**주의사항**:
- 쿼리 파라미터는 snake_case (`workspace_id`)
- 응답 구조: `response.data.data.projects`

#### 2. Board API
- `GET /api/boards?projectId={id}` - 보드 목록 조회
- `POST /api/boards` - 보드 생성
- `GET /api/boards/{id}` - 보드 상세 조회
- `PUT /api/boards/{id}` - 보드 수정
- `DELETE /api/boards/{id}` - 보드 삭제
- **구현 위치**: `frontend/src/api/board/boardService.ts`
- **사용 컴포넌트**: `Dashboard.tsx`, `CreateBoardModal.tsx`, `BoardDetailModal.tsx`

#### 3. Custom Fields API
- `GET /api/custom-fields/projects/{projectId}/stages` - Stage 목록
- `GET /api/custom-fields/projects/{projectId}/roles` - Role 목록
- `GET /api/custom-fields/projects/{projectId}/importance` - Importance 목록
- `POST /api/custom-fields/stages` - Stage 생성
- `PUT /api/custom-fields/stages/{id}` - Stage 수정
- `DELETE /api/custom-fields/stages/{id}` - Stage 삭제
- *(Role, Importance도 동일한 CRUD 패턴)*
- **구현 위치**: `frontend/src/api/board/boardService.ts`
- **사용 컴포넌트**: `CustomFieldManageModal.tsx`, `Dashboard.tsx`

#### 4. Comment API
- `GET /api/comments?boardId={id}` - 댓글 목록
- `POST /api/comments` - 댓글 작성
- `PUT /api/comments/{id}` - 댓글 수정
- `DELETE /api/comments/{id}` - 댓글 삭제
- **구현 위치**: `frontend/src/api/board/boardService.ts`
- **사용 컴포넌트**: `BoardDetailModal.tsx`

#### 5. User Order API
- `PUT /api/projects/{id}/orders/stage-columns` - Stage 컬럼 순서 저장
- `PUT /api/projects/{id}/orders/stage-boards/{stageId}` - Stage 내 보드 순서 저장
- **구현 위치**: `frontend/src/api/board/boardService.ts`
- **사용 컴포넌트**: `Dashboard.tsx` (드래그 앤 드롭)

### 🔄 부분 구현

#### Role 기반 뷰
- `GET /api/projects/{id}/orders/role-board` - Role 기준 보드 뷰 (API만 정의됨)
- `PUT /api/projects/{id}/orders/role-columns` - Role 컬럼 순서 (미구현)
- `PUT /api/projects/{id}/orders/role-boards/{roleId}` - Role 내 보드 순서 (미구현)

---

## 드래그 앤 드롭 시스템

### 구현된 기능

#### 1. 보드 카드 드래그 (Cross-Column)
- **기능**: 보드를 다른 Stage 컬럼으로 이동
- **API 호출**: `PUT /api/boards/{id}` (stageId 변경)
- **구현**: `Dashboard.tsx:298-401`

**작동 방식**:
```typescript
// 1. Optimistic UI Update
const newColumns = columns.map((col) => {
  if (col.id === draggedFromColumn) {
    return { ...col, boards: col.boards.filter((t) => t.id !== draggedBoard.id) };
  }
  if (col.id === targetColumnId) {
    return { ...col, boards: [...col.boards, updatedBoard] };
  }
  return col;
});
setColumns(newColumns);

// 2. API 호출
await updateBoard(boardId, { stageId: targetColumnId, ... }, token);

// 3. 에러 시 롤백
catch (error) {
  setColumns(columns); // 이전 상태로 복구
}
```

#### 2. 보드 카드 드래그 (Same-Column)
- **기능**: 같은 컬럼 내에서 보드 순서 변경
- **API 호출**: `PUT /api/projects/{id}/orders/stage-boards/{stageId}`
- **구현**: `Dashboard.tsx:295-356`

**작동 방식**:
```typescript
// 1. 배열 재정렬
const newBoards = [...targetColumn.boards];
const [removed] = newBoards.splice(draggedIndex, 1);
newBoards.splice(targetIndex, 0, removed);

// 2. User Order API로 순서 저장
const boardIds = newBoards.map((b) => b.id);
await updateStageBoardOrder(projectId, stageId, boardIds, token);
```

#### 3. 컬럼 드래그 (Stage 순서 변경)
- **기능**: Stage 컬럼 자체의 순서 변경
- **API 호출**: `PUT /api/projects/{id}/orders/stage-columns`
- **구현**: `Dashboard.tsx:403-443`

### 시각적 피드백

#### 드래그 중인 항목
```css
opacity-80 scale-95
```

#### 드롭 대상 컬럼
```css
border-blue-500 bg-blue-50 dark:bg-blue-900/20
```

#### 드롭 위치 인디케이터
- **세로 드래그**: 대상 보드 위에 파란색 펄스 라인 + `mt-3` 여백
- **가로 드래그**: 컬럼 하단에 "여기에 추가됩니다" 텍스트 + 파란색 라인

**구현**:
```tsx
{dragOverBoardId === board.id && draggedBoard && draggedBoard.id !== board.id && (
  <div className="absolute -top-1 left-0 right-0 h-1 bg-blue-500 rounded-full shadow-lg shadow-blue-500/50 z-10 animate-pulse"></div>
)}
```

---

## 커스텀 필드 관리

### CustomFieldManageModal

**파일**: `frontend/src/components/modals/CustomFieldManageModal.tsx`

#### 기능
- Stage, Role, Importance를 탭으로 구분하여 관리
- 생성, 수정, 삭제 기능
- 12가지 색상 팔레트
- Importance는 Level(1-5) 설정 가능

#### 색상 선택 UI
```tsx
<div className="grid grid-cols-6 gap-2 mt-2">
  {CUSTOM_FIELD_COLORS.map((color) => (
    <button
      style={{ backgroundColor: color.hex }}
      className={selectedColor === color.hex ? 'ring-2 ring-blue-500' : ''}
    />
  ))}
</div>
```

#### 삭제 제한
- `isSystemDefault: true` 인 항목은 삭제 불가
- UI에서 비활성화 + 툴팁 표시

#### 사용 방법
1. Dashboard에서 FilterBar의 "관리" 버튼 클릭
2. 원하는 탭(Stage/Role/Importance) 선택
3. "새로운 항목 추가" 버튼으로 생성
4. 연필 아이콘으로 수정, 휴지통 아이콘으로 삭제

---

## 필터 및 검색

### FilterBar

**파일**: `frontend/src/components/FilterBar.tsx`

#### 구성 요소

```
┌──────────────────────────────────────────────────────────────────┐
│  🔍 [검색...]   [뷰: Stage ▼]   [필터: 전체 ▼]   [⚙️ 관리]     │
└──────────────────────────────────────────────────────────────────┘
```

#### 1. 검색바
- 보드 제목/내용 검색
- **상태**: UI 완료, 검색 로직은 TODO

#### 2. 뷰 전환
- **Stage 기준**: 진행 단계별로 컬럼 구성 (현재 구현됨)
- **Role 기준**: 역할별로 컬럼 구성 (미구현)

#### 3. 필터 옵션
- 전체
- 내가 담당한 것만
- 중요도 높음
- 긴급
- 완료된 것 숨기기

**상태**: UI 완료, 필터링 로직은 TODO

#### 4. 관리 버튼
- 클릭 시 `CustomFieldManageModal` 열림

---

## 색상 시스템

### 색상 팔레트

**파일**: `frontend/src/constants/colors.ts`

#### 정의된 색상 (12가지)
```typescript
export const CUSTOM_FIELD_COLORS: ColorOption[] = [
  { name: '파란색', hex: '#3B82F6', bgClass: 'bg-blue-500', textClass: 'text-blue-500' },
  { name: '하늘색', hex: '#06B6D4', bgClass: 'bg-cyan-500', textClass: 'text-cyan-500' },
  { name: '청록색', hex: '#14B8A6', bgClass: 'bg-teal-500', textClass: 'text-teal-500' },
  { name: '초록색', hex: '#22C55E', bgClass: 'bg-green-500', textClass: 'text-green-500' },
  { name: '연두색', hex: '#84CC16', bgClass: 'bg-lime-500', textClass: 'text-lime-500' },
  { name: '노란색', hex: '#EAB308', bgClass: 'bg-yellow-500', textClass: 'text-yellow-500' },
  { name: '주황색', hex: '#F97316', bgClass: 'bg-orange-500', textClass: 'text-orange-500' },
  { name: '빨간색', hex: '#EF4444', bgClass: 'bg-red-500', textClass: 'text-red-500' },
  { name: '분홍색', hex: '#EC4899', bgClass: 'bg-pink-500', textClass: 'text-pink-500' },
  { name: '보라색', hex: '#A855F7', bgClass: 'bg-purple-500', textClass: 'text-purple-500' },
  { name: '남색', hex: '#6366F1', bgClass: 'bg-indigo-500', textClass: 'text-indigo-500' },
  { name: '회색', hex: '#6B7280', bgClass: 'bg-gray-500', textClass: 'text-gray-500' },
];
```

### 색상 사용 방식

#### Before (위치 기반)
```typescript
const columnColors = ['bg-blue-500', 'bg-yellow-500', 'bg-purple-500'];
<span className={columnColors[idx % columnColors.length]} />
```
**문제**: 컬럼 순서가 바뀌면 색깔도 바뀜

#### After (API 기반)
```typescript
<span style={{ backgroundColor: column.color || getDefaultColorByIndex(idx).hex }} />
```
**장점**:
- API에서 받은 색상을 사용 (persistence)
- 드래그로 순서를 바꿔도 색깔 유지
- API에 색상이 없으면 기본 팔레트에서 자동 할당

### Column 인터페이스
```typescript
interface Column {
  id: string;
  title: string;
  color?: string; // hex color from API
  boards: BoardResponse[];
}
```

---

## 향후 작업

### 1. 검색 및 필터링 로직 구현
**현재 상태**: UI만 완성, 로직은 TODO

**필요한 작업**:
```typescript
// Dashboard.tsx에서 구현 필요
const filteredColumns = columns.map(col => ({
  ...col,
  boards: col.boards.filter(board => {
    // 검색어 필터
    if (searchQuery && !board.title.includes(searchQuery)) return false;

    // 필터 옵션
    if (filterOption === 'my' && board.assignee?.userId !== currentUserId) return false;
    if (filterOption === 'high' && board.importance?.level < 4) return false;
    // ...

    return true;
  })
}));
```

### 2. Role 기반 뷰 구현
**필요한 작업**:
- `fetchBoardsByRole` 함수 생성
- Role 컬럼 렌더링 로직
- Role 기반 드래그 앤 드롭 핸들러
- User Order API 통합 (Role 버전)

### 3. Project CRUD 확장
**현재**: 생성만 가능
**추가 필요**:
- `PUT /api/projects/{id}` - 프로젝트 수정 (이름, 설명)
- `DELETE /api/projects/{id}` - 프로젝트 삭제

### 4. Assignee 및 Due Date
**현재**: 표시만 가능
**추가 필요**:
- Assignee 선택 UI (User Service API 연동)
- Due Date 선택 달력 UI
- 기한 임박 필터링

### 5. 뷰 저장 기능
- 사용자별 필터/정렬 설정 저장
- "내 뷰" 기능 (즐겨찾기)

---

## 주요 컴포넌트 구조

```
src/
├── components/
│   ├── modals/
│   │   ├── BoardDetailModal.tsx          # 보드 상세/수정/삭제
│   │   ├── CreateProjectModal.tsx        # 프로젝트 생성
│   │   ├── CreateBoardModal.tsx          # 보드 생성
│   │   ├── CustomFieldManageModal.tsx    # 커스텀 필드 관리
│   │   └── UserProfileModal.tsx          # 사용자 프로필
│   └── FilterBar.tsx                      # 필터/검색/뷰 전환
├── pages/
│   └── Dashboard.tsx                      # 메인 대시보드
├── api/
│   └── board/
│       └── boardService.ts                # Board Service API
├── constants/
│   └── colors.ts                          # 색상 팔레트
└── types/
    └── board.ts                           # 보드 관련 타입 정의
```

---

## 코드 컨벤션

### API 호출 패턴
```typescript
export const apiFunction = async (params, token: string): Promise<ResponseType> => {
  try {
    const response = await boardService.method('/api/endpoint', data, {
      headers: { Authorization: `Bearer ${token}` },
    });
    return response.data.data;
  } catch (error) {
    console.error('apiFunction error:', error);
    throw error;
  }
};
```

### 에러 핸들링
- Optimistic UI Update 사용
- 에러 발생 시 이전 상태로 롤백
- 사용자에게 alert로 에러 메시지 표시

### 상태 관리
- useState for local component state
- useCallback for memoized functions
- useEffect for side effects

---

## 성능 최적화

### 1. API 호출 최소화
- useCallback으로 fetch 함수 메모이제이션
- 필요할 때만 re-fetch

### 2. Optimistic UI Updates
- API 응답 대기 없이 UI 먼저 업데이트
- 사용자 경험 개선

### 3. 드래그 앤 드롭 최적화
- 드래그 중 불필요한 리렌더링 방지
- CSS transition으로 부드러운 애니메이션

---

## 테스트 가이드

### 수동 테스트 시나리오

#### 1. 프로젝트 생성
1. 대시보드 헤더의 프로젝트 선택기 클릭
2. "새 프로젝트" 버튼
3. 이름/설명 입력 후 생성
4. 프로젝트 목록에 추가됨 확인

#### 2. 보드 생성
1. 프로젝트 선택
2. 컬럼의 "보드 추가" 버튼
3. Stage, Role 선택 후 생성
4. 해당 컬럼에 보드 추가됨 확인

#### 3. 드래그 앤 드롭
1. 보드 카드 드래그
2. 드롭 위치 인디케이터 확인
3. 드롭 후 위치 변경 확인
4. 페이지 새로고침 후 순서 유지 확인

#### 4. 커스텀 필드 관리
1. FilterBar의 "관리" 버튼 클릭
2. Stage 탭에서 새 항목 추가
3. 색상 선택 확인
4. Dashboard에서 새 컬럼 생성 시 해당 Stage 선택 가능 확인
5. 컬럼 색상이 선택한 색상으로 표시됨 확인

---

## 문제 해결 가이드

### Q1. 프로젝트 목록이 안 불러와져요
**원인**: Query parameter mismatch
**해결**: `params: { workspace_id: workspaceId }` (snake_case 사용)

### Q2. 빈 프로젝트에서 보드 생성 버튼이 안 보여요
**원인**: Stages가 없으면 컬럼이 안 생성됨
**해결**: 모든 Stage를 먼저 fetch하여 빈 컬럼 생성

### Q3. 드래그 후 색깔이 바뀌어요
**원인**: 위치 기반 색상 사용
**해결**: API의 `color` 필드 사용 (이미 수정됨)

### Q4. Custom Field 삭제가 안 돼요
**원인**: 시스템 기본값
**해결**: `isSystemDefault: false`인 항목만 삭제 가능

---

## 참고 문서
- [Board Service API Reference](./BOARD_SERVICE_API_REFERENCE.md)
- [Backend 최적화 가이드](./BACKEND_OPTIMIZATION_GUIDE.md)
