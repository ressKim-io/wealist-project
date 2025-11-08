// src/api/boardService.ts
import axios from 'axios';

const BOARD_API_URL = import.meta.env.VITE_REACT_APP_GO_API_URL || 'http://localhost:8000';

const boardService = axios.create({
  baseURL: BOARD_API_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

// ============================================================================
// 프로젝트 관련 API
// ============================================================================

export interface ProjectResponse {
  id: string;
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
    const response = await boardService.get('/api/projects', {
      params: { workspace_id: workspaceId },
      headers: { Authorization: `Bearer ${token}` },
    });
    return response.data.data?.projects || [];
  } catch (error) {
    console.error('getProjects error:', error);
    throw error;
  }
};

/**
 * 특정 프로젝트를 조회합니다.
 * GET /api/projects/{id}
 * @param projectId 프로젝트 ID
 * @param token 액세스 토큰
 * @returns 프로젝트 정보
 */
export const getProject = async (projectId: string, token: string): Promise<ProjectResponse> => {
  try {
    const response = await boardService.get(`/api/projects/${projectId}`, {
      headers: { Authorization: `Bearer ${token}` },
    });
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
  try {
    const response = await boardService.post('/api/projects', data, {
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
 * PUT /api/projects/{id}
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
    const response = await boardService.put(`/api/projects/${projectId}`, data, {
      headers: { Authorization: `Bearer ${token}` },
    });
    return response.data.data;
  } catch (error) {
    console.error('updateProject error:', error);
    throw error;
  }
};

/**
 * 프로젝트를 삭제합니다.
 * DELETE /api/projects/{id}
 * @param projectId 프로젝트 ID
 * @param token 액세스 토큰
 * @returns 응답 메시지
 */
export const deleteProject = async (projectId: string, token: string): Promise<any> => {
  try {
    const response = await boardService.delete(`/api/projects/${projectId}`, {
      headers: { Authorization: `Bearer ${token}` },
    });
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
    const response = await boardService.get('/api/projects/search', {
      params: { workspaceId, query },
      headers: { Authorization: `Bearer ${token}` },
    });
    return response.data.data?.projects || [];
  } catch (error) {
    console.error('searchProjects error:', error);
    throw error;
  }
};

// ============================================================================
// 보드 관련 API
// ============================================================================

export interface BoardResponse {
  id: string;
  projectId: string;
  title: string;
  content?: string;
  stage?: any;
  roles?: any[];
  importance?: any;
  assignee?: any;
  author?: any;
  dueDate?: string;
  createdAt: string;
  updatedAt: string;
}

export interface CreateBoardRequest {
  projectId: string;
  title: string;
  content?: string;
  stageId: string;
  roleIds: string[];
  importanceId?: string;
  assigneeId?: string;
  dueDate?: string;
}

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
 * @param filters 필터 옵션 (stageId, roleId, importanceId, assigneeId, authorId, page, limit)
 * @returns 보드 배열
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
  try {
    const params = { projectId, ...filters };
    const response = await boardService.get('/api/boards', {
      params,
      headers: { Authorization: `Bearer ${token}` },
    });
    return response.data.data || { boards: [], total: 0, page: 1, limit: 20 };
  } catch (error) {
    console.error('getBoards error:', error);
    throw error;
  }
};

/**
 * 특정 보드를 조회합니다.
 * GET /api/boards/{id}
 * @param boardId 보드 ID
 * @param token 액세스 토큰
 * @returns 보드 정보
 */
export const getBoard = async (boardId: string, token: string): Promise<BoardResponse> => {
  try {
    const response = await boardService.get(`/api/boards/${boardId}`, {
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
    const response = await boardService.post('/api/boards', data, {
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
 * PUT /api/boards/{id}
 * @param boardId 보드 ID
 * @param data 업데이트 정보
 * @param token 액세스 토큰
 * @returns 업데이트된 보드
 */
export const updateBoard = async (
  boardId: string,
  data: Partial<CreateBoardRequest>,
  token: string,
): Promise<BoardResponse> => {
  try {
    const response = await boardService.put(`/api/boards/${boardId}`, data, {
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
 * DELETE /api/boards/{id}
 * @param boardId 보드 ID
 * @param token 액세스 토큰
 * @returns 응답 메시지
 */
export const deleteBoard = async (boardId: string, token: string): Promise<any> => {
  try {
    const response = await boardService.delete(`/api/boards/${boardId}`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    return response.data;
  } catch (error) {
    console.error('deleteBoard error:', error);
    throw error;
  }
};

// ============================================================================
// 커스텀 필드 API
// ============================================================================

export interface CustomStageResponse {
  id: string;
  projectId: string;
  name: string;
  color?: string;
  displayOrder: number;
  isSystemDefault: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface CustomRoleResponse {
  id: string;
  projectId: string;
  name: string;
  color?: string;
  displayOrder: number;
  isSystemDefault: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface CustomImportanceResponse {
  id: string;
  projectId: string;
  name: string;
  color?: string;
  displayOrder: number;
  isSystemDefault: boolean;
  createdAt: string;
  updatedAt: string;
}

/**
 * 프로젝트의 모든 Stage를 조회합니다.
 * GET /api/custom-fields/projects/{projectId}/stages
 * @param projectId 프로젝트 ID
 * @param token 액세스 토큰
 * @returns Stage 배열
 */
export const getProjectStages = async (
  projectId: string,
  token: string,
): Promise<CustomStageResponse[]> => {
  try {
    const response = await boardService.get(`/api/custom-fields/projects/${projectId}/stages`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    return response.data.data || [];
  } catch (error) {
    console.error('getProjectStages error:', error);
    throw error;
  }
};

/**
 * 프로젝트의 모든 Role을 조회합니다.
 * GET /api/custom-fields/projects/{projectId}/roles
 * @param projectId 프로젝트 ID
 * @param token 액세스 토큰
 * @returns Role 배열
 */
export const getProjectRoles = async (
  projectId: string,
  token: string,
): Promise<CustomRoleResponse[]> => {
  try {
    const response = await boardService.get(`/api/custom-fields/projects/${projectId}/roles`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    return response.data.data || [];
  } catch (error) {
    console.error('getProjectRoles error:', error);
    throw error;
  }
};

/**
 * 프로젝트의 모든 Importance를 조회합니다.
 * GET /api/custom-fields/projects/{projectId}/importance
 * @param projectId 프로젝트 ID
 * @param token 액세스 토큰
 * @returns Importance 배열
 */
export const getProjectImportances = async (
  projectId: string,
  token: string,
): Promise<CustomImportanceResponse[]> => {
  try {
    const response = await boardService.get(`/api/custom-fields/projects/${projectId}/importance`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    return response.data.data || [];
  } catch (error) {
    console.error('getProjectImportances error:', error);
    throw error;
  }
};

// ============================================================================
// Custom Fields CRUD API
// ============================================================================

export interface CreateCustomStageRequest {
  projectId: string;
  name: string;
  color: string;
}

export interface UpdateCustomStageRequest {
  name: string;
  color: string;
}

export interface CreateCustomRoleRequest {
  projectId: string;
  name: string;
  color: string;
}

export interface UpdateCustomRoleRequest {
  name: string;
  color: string;
}

export interface CreateCustomImportanceRequest {
  projectId: string;
  name: string;
  color: string;
  level: number; // 1-5
}

export interface UpdateCustomImportanceRequest {
  name: string;
  color: string;
  level: number;
}

/**
 * Stage를 생성합니다.
 * POST /api/custom-fields/stages
 */
export const createStage = async (
  data: CreateCustomStageRequest,
  token: string,
): Promise<CustomStageResponse> => {
  try {
    const response = await boardService.post('/api/custom-fields/stages', data, {
      headers: { Authorization: `Bearer ${token}` },
    });
    return response.data.data;
  } catch (error) {
    console.error('createStage error:', error);
    throw error;
  }
};

/**
 * Stage를 수정합니다.
 * PUT /api/custom-fields/stages/{id}
 */
export const updateStage = async (
  stageId: string,
  data: UpdateCustomStageRequest,
  token: string,
): Promise<CustomStageResponse> => {
  try {
    const response = await boardService.put(`/api/custom-fields/stages/${stageId}`, data, {
      headers: { Authorization: `Bearer ${token}` },
    });
    return response.data.data;
  } catch (error) {
    console.error('updateStage error:', error);
    throw error;
  }
};

/**
 * Stage를 삭제합니다.
 * DELETE /api/custom-fields/stages/{id}
 */
export const deleteStage = async (stageId: string, token: string): Promise<void> => {
  try {
    await boardService.delete(`/api/custom-fields/stages/${stageId}`, {
      headers: { Authorization: `Bearer ${token}` },
    });
  } catch (error) {
    console.error('deleteStage error:', error);
    throw error;
  }
};

/**
 * Role을 생성합니다.
 * POST /api/custom-fields/roles
 */
export const createRole = async (
  data: CreateCustomRoleRequest,
  token: string,
): Promise<CustomRoleResponse> => {
  try {
    const response = await boardService.post('/api/custom-fields/roles', data, {
      headers: { Authorization: `Bearer ${token}` },
    });
    return response.data.data;
  } catch (error) {
    console.error('createRole error:', error);
    throw error;
  }
};

/**
 * Role을 수정합니다.
 * PUT /api/custom-fields/roles/{id}
 */
export const updateRole = async (
  roleId: string,
  data: UpdateCustomRoleRequest,
  token: string,
): Promise<CustomRoleResponse> => {
  try {
    const response = await boardService.put(`/api/custom-fields/roles/${roleId}`, data, {
      headers: { Authorization: `Bearer ${token}` },
    });
    return response.data.data;
  } catch (error) {
    console.error('updateRole error:', error);
    throw error;
  }
};

/**
 * Role을 삭제합니다.
 * DELETE /api/custom-fields/roles/{id}
 */
export const deleteRole = async (roleId: string, token: string): Promise<void> => {
  try {
    await boardService.delete(`/api/custom-fields/roles/${roleId}`, {
      headers: { Authorization: `Bearer ${token}` },
    });
  } catch (error) {
    console.error('deleteRole error:', error);
    throw error;
  }
};

/**
 * Importance를 생성합니다.
 * POST /api/custom-fields/importance
 */
export const createImportance = async (
  data: CreateCustomImportanceRequest,
  token: string,
): Promise<CustomImportanceResponse> => {
  try {
    const response = await boardService.post('/api/custom-fields/importance', data, {
      headers: { Authorization: `Bearer ${token}` },
    });
    return response.data.data;
  } catch (error) {
    console.error('createImportance error:', error);
    throw error;
  }
};

/**
 * Importance를 수정합니다.
 * PUT /api/custom-fields/importance/{id}
 */
export const updateImportance = async (
  importanceId: string,
  data: UpdateCustomImportanceRequest,
  token: string,
): Promise<CustomImportanceResponse> => {
  try {
    const response = await boardService.put(`/api/custom-fields/importance/${importanceId}`, data, {
      headers: { Authorization: `Bearer ${token}` },
    });
    return response.data.data;
  } catch (error) {
    console.error('updateImportance error:', error);
    throw error;
  }
};

/**
 * Importance를 삭제합니다.
 * DELETE /api/custom-fields/importance/{id}
 */
export const deleteImportance = async (importanceId: string, token: string): Promise<void> => {
  try {
    await boardService.delete(`/api/custom-fields/importance/${importanceId}`, {
      headers: { Authorization: `Bearer ${token}` },
    });
  } catch (error) {
    console.error('deleteImportance error:', error);
    throw error;
  }
};

// ============================================================================
// Comment API
// ============================================================================

export interface CommentResponse {
  id: string;
  userId: string;
  userName: string;
  userAvatar: string;
  content: string;
  createdAt: string;
  updatedAt: string;
}

export interface CreateCommentRequest {
  boardId: string;
  content: string;
}

export interface UpdateCommentRequest {
  content: string;
}

/**
 * 보드의 모든 댓글을 조회합니다.
 * GET /api/comments
 * @param boardId 보드 ID
 * @param token 액세스 토큰
 * @returns 댓글 배열
 */
export const getComments = async (
  boardId: string,
  token: string,
): Promise<CommentResponse[]> => {
  try {
    const response = await boardService.get('/api/comments', {
      params: { boardId },
      headers: { Authorization: `Bearer ${token}` },
    });
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
    const response = await boardService.post('/api/comments', data, {
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
 * PUT /api/comments/{id}
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
    const response = await boardService.put(`/api/comments/${commentId}`, data, {
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
 * DELETE /api/comments/{id}
 * @param commentId 댓글 ID
 * @param token 액세스 토큰
 */
export const deleteComment = async (commentId: string, token: string): Promise<void> => {
  try {
    await boardService.delete(`/api/comments/${commentId}`, {
      headers: { Authorization: `Bearer ${token}` },
    });
  } catch (error) {
    console.error('deleteComment error:', error);
    throw error;
  }
};

// ============================================================================
// 보드 뷰 API (Stage/Role 기반)
// ============================================================================

export interface RoleBasedBoardView {
  projectId: string;
  viewType: 'role';
  columns: Array<{
    customRoleId: string;
    roleName: string;
    roleColor: string;
    displayOrder: number;
    boards: Array<{
      boardId: string;
      title: string;
      displayOrder: number;
    }>;
  }>;
}

export interface StageBasedBoardView {
  projectId: string;
  viewType: 'stage';
  columns: Array<{
    customStageId: string;
    stageName: string;
    stageColor: string;
    displayOrder: number;
    boards: Array<{
      boardId: string;
      title: string;
      displayOrder: number;
    }>;
  }>;
}

/**
 * Role 기반 보드 뷰를 조회합니다.
 * GET /api/projects/{id}/orders/role-board
 * @param projectId 프로젝트 ID
 * @param token 액세스 토큰
 * @returns Role 기반 보드 뷰
 */
export const getRoleBasedBoardView = async (
  projectId: string,
  token: string,
): Promise<RoleBasedBoardView> => {
  try {
    const response = await boardService.get(`/api/projects/${projectId}/orders/role-board`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    return response.data.data;
  } catch (error) {
    console.error('getRoleBasedBoardView error:', error);
    throw error;
  }
};

/**
 * Stage 기반 보드 뷰를 조회합니다.
 * GET /api/projects/{id}/orders/stage-board
 * @param projectId 프로젝트 ID
 * @param token 액세스 토큰
 * @returns Stage 기반 보드 뷰
 */
export const getStageBasedBoardView = async (
  projectId: string,
  token: string,
): Promise<StageBasedBoardView> => {
  try {
    const response = await boardService.get(`/api/projects/${projectId}/orders/stage-board`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    return response.data.data;
  } catch (error) {
    console.error('getStageBasedBoardView error:', error);
    throw error;
  }
};

/**
 * Stage 컬럼 순서를 업데이트합니다.
 * PUT /api/projects/{id}/orders/stage-columns
 * @param projectId 프로젝트 ID
 * @param stageIds Stage ID 배열 (순서대로)
 * @param token 액세스 토큰
 */
export const updateStageColumnOrder = async (
  projectId: string,
  stageIds: string[],
  token: string,
): Promise<void> => {
  try {
    await boardService.put(
      `/api/projects/${projectId}/orders/stage-columns`,
      { itemIds: stageIds },
      { headers: { Authorization: `Bearer ${token}` } },
    );
  } catch (error) {
    console.error('updateStageColumnOrder error:', error);
    throw error;
  }
};

/**
 * Stage 내 Board 순서를 업데이트합니다.
 * PUT /api/projects/{id}/orders/stage-boards/{stageId}
 * @param projectId 프로젝트 ID
 * @param stageId Stage ID
 * @param boardIds Board ID 배열 (순서대로)
 * @param token 액세스 토큰
 */
export const updateStageBoardOrder = async (
  projectId: string,
  stageId: string,
  boardIds: string[],
  token: string,
): Promise<void> => {
  try {
    await boardService.put(
      `/api/projects/${projectId}/orders/stage-boards/${stageId}`,
      { itemIds: boardIds },
      { headers: { Authorization: `Bearer ${token}` } },
    );
  } catch (error) {
    console.error('updateStageBoardOrder error:', error);
    throw error;
  }
};
