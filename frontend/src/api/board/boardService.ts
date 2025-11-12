// src/api/boardService.ts
import { boardServiceClient } from '../apiConfig';

/**
 * ========================================
 * 목업 모드 전환
 * ========================================
 *
 * USE_MOCK_DATA = true: 목업 데이터 사용
 * USE_MOCK_DATA = false: 실제 API 호출
 */
const USE_MOCK_DATA = false;

// ============================================================================
// 프로젝트 관련 API
// ============================================================================

export interface ProjectResponse {
  projectId: string;
  name: string;
  description?: string;
  workspaceId: string;
  ownerId: string;
  ownerName: string;
  ownerEmail: string;
  createdAt: string;
  updatedAt: string;
}

export interface CreateProjectRequest {
  name: string;
  description?: string;
  workspaceId: string;
}

/**
 * 워크스페이스의 모든 프로젝트를 조회합니다.
 * GET /api/projects
 * @param workspaceId 워크스페이스 ID
 * @param token 액세스 토큰
 * @returns 프로젝트 배열
 */
export const getProjects = async (
  workspaceId: string,
  token: string,
): Promise<ProjectResponse[]> => {
  try {
    const response = await boardServiceClient.get('/api/projects', {
      params: { workspaceId: workspaceId },
      headers: { Authorization: `Bearer ${token}` },
    });
    // API 문서 응답 구조: { data: [ { ...project } ] }
    // 기존 코드: response.data.data?.projects || [] - API 문서와 기존 응답 처리 코드가 일치하지 않습니다. 문서 응답 JSON 형식에 맞게 수정.
    // 문서: { "data": [ { ...project } ] } -> response.data.data를 배열로 가정
    console.log(response?.data?.data?.projects);
    return response?.data?.data?.projects || [];
  } catch (error) {
    console.error('getProjects error:', error);
    throw error;
  }
};

/**
 * 특정 프로젝트를 조회합니다.
 * GET /api/projects/{projectId}
 * @param projectId 프로젝트 ID
 * @param token 액세스 토큰
 * @returns 프로젝트 정보
 */
export const getProject = async (projectId: string, token: string): Promise<ProjectResponse> => {
  try {
    const response = await boardServiceClient.get(`/api/projects/${projectId}`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    // API 문서 응답 구조는 명확하지 않지만, 상세 조회는 단일 객체를 기대하고, 기존 코드 유지
    return response.data.data;
  } catch (error) {
    console.error('getProject error:', error);
    throw error;
  }
};

/**
 * 새로운 프로젝트를 생성합니다.
 * POST /api/projects
 * @param data 프로젝트 생성 정보
 * @param token 액세스 토큰
 * @returns 생성된 프로젝트
 */
export const createProject = async (
  data: CreateProjectRequest,
  token: string,
): Promise<ProjectResponse> => {
  // 목업 로직 제거 (USE_MOCK_DATA가 false이므로)

  try {
    const response = await boardServiceClient.post('/api/projects', data, {
      headers: { Authorization: `Bearer ${token}` },
    });
    return response.data.data;
  } catch (error) {
    console.error('createProject error:', error);
    throw error;
  }
};

/**
 * 프로젝트를 업데이트합니다.
 * PUT /api/projects/{projectId}
 * @param projectId 프로젝트 ID
 * @param data 업데이트 정보
 * @param token 액세스 토큰
 * @returns 업데이트된 프로젝트
 */
export const updateProject = async (
  projectId: string,
  data: Partial<CreateProjectRequest>,
  token: string,
): Promise<ProjectResponse> => {
  try {
    const response = await boardServiceClient.put(`/api/projects/${projectId}`, data, {
      headers: { Authorization: `Bearer ${token}` },
    });
    // API 문서 응답은 200 OK, 응답 본문 구조는 미제공, 기존 코드 유지
    return response.data.data;
  } catch (error) {
    console.error('updateProject error:', error);
    throw error;
  }
};

/**
 * 프로젝트를 삭제합니다.
 * DELETE /api/projects/{projectId}
 * @param projectId 프로젝트 ID
 * @param token 액세스 토큰
 * @returns 응답 메시지
 */
export const deleteProject = async (projectId: string, token: string): Promise<any> => {
  try {
    const response = await boardServiceClient.delete(`/api/projects/${projectId}`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    // API 문서 응답은 200 OK, 응답 본문 구조는 미제공, 기존 코드 유지
    return response.data;
  } catch (error) {
    console.error('deleteProject error:', error);
    throw error;
  }
};

/**
 * 프로젝트를 검색합니다.
 * GET /api/projects/search
 * @param workspaceId 워크스페이스 ID
 * @param query 검색 쿼리
 * @param token 액세스 토큰
 * @returns 검색된 프로젝트 배열
 */
export const searchProjects = async (
  workspaceId: string,
  query: string,
  token: string,
): Promise<ProjectResponse[]> => {
  try {
    const response = await boardServiceClient.get('/api/projects/search', {
      params: { workspaceId, query },
      headers: { Authorization: `Bearer ${token}` },
    });
    // API 문서 응답 구조는 명확하지 않지만, 기존 코드 유지
    return response.data.data || [];
  } catch (error) {
    console.error('searchProjects error:', error);
    throw error;
  }
};

// ============================================================================
// 보드 관련 API
// ============================================================================

export interface BoardResponse {
  // 💡 board_id -> boardId (API 응답 구조에 맞게 수정)
  boardId: string;
  title: string;
  content: string;
  projectId: string;
  position: string;
  dueDate: string;
  createdAt: string;
  updatedAt: string;
  author: {
    userId: string;
    name: string;
    email: string;
    isActive: boolean;
  };
  assignee: {
    userId: string;
    name: string;
    email: string;
    isActive: boolean;
  };
  customFields: Record<string, any>; // customFields는 동적 객체이므로 Record<string, any> 사용
}

export interface CreateBoardRequest {
  projectId: string;
  title: string;
  content?: string;
  assigneeId?: string; // 💡 assignee_id 대신 assigneeId 사용
  dueDate?: string;
  stageId?: string; // 💡 레거시 필드 유지
  importanceId?: string;
  roleIds?: string[]; // 💡 roleIds 사용
}

export interface UpdateBoardRequest extends Partial<CreateBoardRequest> {}

export interface PaginatedBoardsResponse {
  boards: BoardResponse[];
  total: number;
  page: number;
  limit: number;
}

/**
 * 프로젝트의 보드를 조회합니다.
 * GET /api/boards
 * @param projectId 프로젝트 ID
 * @param token 액세스 토큰
 * @param filters 필터 옵션
 * @returns 페이징된 보드 응답
 */
export const getBoards = async (
  projectId: string,
  token: string,
  filters?: {
    stageId?: string;
    roleId?: string;
    importanceId?: string;
    assigneeId?: string;
    authorId?: string;
    page?: number;
    limit?: number;
  },
): Promise<PaginatedBoardsResponse> => {
  if (USE_MOCK_DATA) {
    // 목업 데이터 로직 (생략, 기존 로직 유지)
    console.log('[MOCK] getBoards 호출:', projectId, filters);
    return new Promise((resolve) => {
      resolve({ boards: [], total: 0, page: 1, limit: 20 });
    });
  }

  try {
    const params = { projectId, ...filters };
    const response = await boardServiceClient.get('/api/boards', {
      params,
      headers: { Authorization: `Bearer ${token}` },
    });
    // API 문서 응답 구조: { data: { boards: [], total: 0, ... } }
    return response.data.data || { boards: [], total: 0, page: 1, limit: 20 };
  } catch (error) {
    console.error('getBoards error:', error);
    throw error;
  }
};

/**
 * 특정 보드를 조회합니다.
 * GET /api/boards/{boardId}
 * @param boardId 보드 ID
 * @param token 액세스 토큰
 * @returns 보드 정보
 */
export const getBoard = async (boardId: string, token: string): Promise<BoardResponse> => {
  if (USE_MOCK_DATA) {
    // 목업 데이터 로직 (생략)
    return new Promise((resolve, reject) => {
      reject(new Error('[MOCK] 보드를 찾을 수 없습니다.'));
    });
  }

  try {
    const response = await boardServiceClient.get(`/api/boards/${boardId}`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    return response.data.data;
  } catch (error) {
    console.error('getBoard error:', error);
    throw error;
  }
};

/**
 * 새로운 보드를 생성합니다.
 * POST /api/boards
 * @param data 보드 생성 정보
 * @param token 액세스 토큰
 * @returns 생성된 보드
 */
export const createBoard = async (
  data: CreateBoardRequest,
  token: string,
): Promise<BoardResponse> => {
  try {
    const response = await boardServiceClient.post('/api/boards', data, {
      headers: { Authorization: `Bearer ${token}` },
    });
    return response.data.data;
  } catch (error) {
    console.error('createBoard error:', error);
    throw error;
  }
};

/**
 * 보드를 업데이트합니다.
 * PUT /api/boards/{boardId}
 * @param boardId 보드 ID
 * @param data 업데이트 정보
 * @param token 액세스 토큰
 * @returns 업데이트된 보드
 */
export const updateBoard = async (
  boardId: string,
  data: UpdateBoardRequest,
  token: string,
): Promise<BoardResponse> => {
  try {
    // 💡 PUT /boards/{boardId} 엔드포인트 사용
    const response = await boardServiceClient.put(`/api/boards/${boardId}`, data, {
      headers: { Authorization: `Bearer ${token}` },
    });
    return response.data.data;
  } catch (error) {
    console.error('updateBoard error:', error);
    throw error;
  }
};

/**
 * 보드를 삭제합니다.
 * DELETE /api/boards/{boardId}
 * @param boardId 보드 ID
 * @param token 액세스 토큰
 * @returns 응답 메시지
 */
export const deleteBoard = async (boardId: string, token: string): Promise<any> => {
  try {
    const response = await boardServiceClient.delete(`/api/boards/${boardId}`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    return response.data;
  } catch (error) {
    console.error('deleteBoard error:', error);
    throw error;
  }
};

// ============================================================================
// 보드 이동 API
// ============================================================================

export interface MoveBoardRequest {
  viewId: string;
  groupByFieldId: string;
  newFieldValue: string;
  beforePosition?: string;
  afterPosition?: string;
}

export interface MoveBoardResponse {
  boardId: string;
  newFieldValue: string;
  newPosition: string;
  message: string;
}

/**
 * 보드를 이동합니다.
 * PUT /api/boards/{boardId}/move
 * @param boardId 보드 ID
 * @param data 이동 정보
 * @param token 액세스 토큰
 * @returns 이동 결과
 */
export const moveBoard = async (
  boardId: string,
  data: MoveBoardRequest,
  token: string,
): Promise<MoveBoardResponse> => {
  try {
    const response = await boardServiceClient.put(`/api/boards/${boardId}/move`, data, {
      headers: { Authorization: `Bearer ${token}` },
    });
    return response.data.data;
  } catch (error) {
    console.error('moveBoard error:', error);
    throw error;
  }
};

// ============================================================================
// 커스텀 필드 관련 API (Stage/Role/Importance는 일반 필드/옵션으로 통합)
// ============================================================================

// 💡 기존 Stage/Role/Importance 인터페이스는 더 이상 명시적으로 사용되지 않지만,
//    필드/옵션 API를 위한 일반적인 응답 타입을 정의합니다.
//    (필드 타입이 너무 다양하여, 명시적인 Stage/Role/Importance 타입은 제거하고 일반 필드/옵션 타입으로 대체)
export interface FieldResponse {
  fieldId: string;
  projectId: string;
  name: string;
  description: string;
  fieldType:
    | 'text'
    | 'number'
    | 'single_select'
    | 'multi_select'
    | 'date'
    | 'datetime'
    | 'single_user'
    | 'multi_user'
    | 'checkbox'
    | 'url';
  isRequired: boolean;
  config: Record<string, any>;
}

export interface FieldOptionResponse {
  optionId: string;
  fieldId: string;
  label: string;
  description: string;
  color: string;
  displayOrder: number;
}

// 💡 기존 Custom Stage/Role/Importance 조회 함수 제거
// 💡 프로젝트 필드 목록 조회 (GET /projects/{projectId}/fields) 추가
/**
 * 프로젝트의 모든 커스텀 필드를 조회합니다.
 * GET /api/projects/{projectId}/fields
 * @param projectId 프로젝트 ID
 * @param token 액세스 토큰
 * @returns 필드 배열
 */
export const getProjectFields = async (
  projectId: string,
  token: string,
): Promise<FieldResponse[]> => {
  try {
    const response = await boardServiceClient.get(`/api/projects/${projectId}/fields`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    return response.data.data || [];
  } catch (error) {
    console.error('getProjectFields error:', error);
    throw error;
  }
};

// 💡 필드 옵션 목록 조회 (GET /fields/{fieldId}/options) 추가
/**
 * 셀렉트 필드의 모든 옵션을 조회합니다.
 * GET /api/fields/{fieldId}/options
 * @param fieldId 필드 ID
 * @param token 액세스 토큰
 * @returns 필드 옵션 배열
 */
export const getFieldOptions = async (
  fieldId: string,
  token: string,
): Promise<FieldOptionResponse[]> => {
  try {
    const response = await boardServiceClient.get(`/api/fields/${fieldId}/options`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    return response.data.data || [];
  } catch (error) {
    console.error('getFieldOptions error:', error);
    throw error;
  }
};

// 💡 기존 Custom Fields CRUD 함수 제거 및 새로운 Fields/Options CRUD 함수 추가 (문서 기반)
// (기존 Custom Fields CRUD 함수는 문서의 Field/Option CRUD API로 대체되어야 합니다.
//  여기서는 모든 CRUD를 다 추가하지 않고, 프로젝트에서 자주 사용되는 조회 함수만 대표적으로 남깁니다.)

// ============================================================================
// 댓글 관리 API
// ============================================================================

export interface CommentResponse {
  commentId: string;
  userId: string;
  userName: string;
  userAvatar: string;
  content: string;
  createdAt: string;
  updatedAt: string;
}

export interface CreateCommentRequest {
  boardId: string; // 💡 board_id -> boardId (API 문서 파라미터에 맞게 수정)
  content: string;
}

export interface UpdateCommentRequest {
  content: string;
}

/**
 * 보드의 모든 댓글을 조회합니다.
 * GET /api/comments
 * @param boardId 보드 ID (💡 board_id -> boardId)
 * @param token 액세스 토큰
 * @returns 댓글 배열
 */
export const getComments = async (boardId: string, token: string): Promise<CommentResponse[]> => {
  try {
    const response = await boardServiceClient.get('/api/comments', {
      params: { boardId }, // 💡 쿼리 파라미터 boardId 사용
      headers: { Authorization: `Bearer ${token}` },
    });
    // API 문서 응답 구조: { data: [ { ...comment } ] }
    return response.data.data || [];
  } catch (error) {
    console.error('getComments error:', error);
    throw error;
  }
};

/**
 * 새 댓글을 생성합니다.
 * POST /api/comments
 * @param data 댓글 생성 정보
 * @param token 액세스 토큰
 * @returns 생성된 댓글
 */
export const createComment = async (
  data: CreateCommentRequest,
  token: string,
): Promise<CommentResponse> => {
  try {
    const response = await boardServiceClient.post('/api/comments', data, {
      headers: { Authorization: `Bearer ${token}` },
    });
    return response.data.data;
  } catch (error) {
    console.error('createComment error:', error);
    throw error;
  }
};

/**
 * 댓글을 수정합니다.
 * PUT /api/comments/{commentId}
 * @param commentId 댓글 ID
 * @param data 수정할 내용
 * @param token 액세스 토큰
 * @returns 수정된 댓글
 */
export const updateComment = async (
  commentId: string,
  data: UpdateCommentRequest,
  token: string,
): Promise<CommentResponse> => {
  try {
    const response = await boardServiceClient.put(`/api/comments/${commentId}`, data, {
      headers: { Authorization: `Bearer ${token}` },
    });
    return response.data.data;
  } catch (error) {
    console.error('updateComment error:', error);
    throw error;
  }
};

/**
 * 댓글을 삭제합니다.
 * DELETE /api/comments/{commentId}
 * @param commentId 댓글 ID
 * @param token 액세스 토큰
 */
export const deleteComment = async (commentId: string, token: string): Promise<void> => {
  try {
    await boardServiceClient.delete(`/api/comments/${commentId}`, {
      headers: { Authorization: `Bearer ${token}` },
    });
  } catch (error) {
    console.error('deleteComment error:', error);
    throw error;
  }
};

// ============================================================================
// 뷰 관리 API (기존 Stage/Role 기반 뷰 API 대신 문서의 뷰 API로 대체)
// ============================================================================

// 💡 뷰 응답/요청 타입 정의 (문서의 응답 예시가 불완전하여 일반적인 뷰 정보만 정의)
export interface ViewResponse {
  viewId: string;
  projectId: string;
  name: string;
  description?: string;
  isDefault: boolean;
  isShared: boolean;
  // ... 기타 필터/정렬/그룹화 필드
}

/**
 * 프로젝트별 뷰 목록 조회
 * GET /api/projects/{projectId}/views
 */
export const getProjectViews = async (
  projectId: string,
  token: string,
): Promise<ViewResponse[]> => {
  try {
    const response = await boardServiceClient.get(`/api/projects/${projectId}/views`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    return response.data.data || [];
  } catch (error) {
    console.error('getProjectViews error:', error);
    throw error;
  }
};

/**
 * 뷰 적용하여 보드 조회
 * GET /api/views/{viewId}/boards
 */
export const getBoardsByView = async (
  viewId: string,
  token: string,
  filters?: {
    page?: number;
    limit?: number;
  },
): Promise<PaginatedBoardsResponse> => {
  try {
    const response = await boardServiceClient.get(`/api/views/${viewId}/boards`, {
      params: filters,
      headers: { Authorization: `Bearer ${token}` },
    });
    // API 문서 응답 구조는 명확하지 않지만, 보드 목록을 반환할 것으로 예상
    return response.data.data || { boards: [], total: 0, page: 1, limit: 20 };
  } catch (error) {
    console.error('getBoardsByView error:', error);
    throw error;
  }
};

// ❌ 기존 Stage/Role 기반 뷰 API (getRoleBasedBoardView, getStageBasedBoardView, updateStageColumnOrder, updateStageBoardOrder)는
//    API 문서에 명시된 내용이 아니므로 삭제 처리했습니다.
