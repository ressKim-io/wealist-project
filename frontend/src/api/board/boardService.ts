// src/api/boardService.ts
import axios from 'axios';

const BOARD_API_URL = import.meta.env.VITE_REACT_APP_GO_API_URL || 'http://localhost:8000';

const boardService = axios.create({
  baseURL: BOARD_API_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

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

// 목업: 프로젝트 목록
let MOCK_PROJECTS: ProjectResponse[] = [
  {
    projectId: 'project-1',
    name: 'Wealist 서비스 개발',
    description: '칸반보드 기반 협업 툴 개발',
    workspaceId: 'workspace-1',
    ownerId: 'user-123',
    ownerName: '김개발',
    ownerEmail: 'dev.kim@orangecloud.com',
    createdAt: '2024-01-01T00:00:00Z',
    updatedAt: '2024-01-15T00:00:00Z',
  },
  {
    projectId: 'project-2',
    name: 'Orange Cloud 디자인 시스템',
    description: 'UI/UX 컴포넌트 라이브러리 구축',
    workspaceId: 'workspace-1',
    ownerId: 'user-456',
    ownerName: '이디자인',
    ownerEmail: 'design.lee@orangecloud.com',
    createdAt: '2024-01-05T00:00:00Z',
    updatedAt: '2024-01-20T00:00:00Z',
  },
  {
    projectId: 'project-3',
    name: '인프라 자동화',
    description: 'EKS 기반 CI/CD 파이프라인 구축',
    workspaceId: 'workspace-1',
    ownerId: 'user-202',
    ownerName: '최데브옵스',
    ownerEmail: 'devops.choi@orangecloud.com',
    createdAt: '2024-01-10T00:00:00Z',
    updatedAt: '2024-01-18T00:00:00Z',
  },
];

// 목업: Stage 데이터
const MOCK_STAGES: CustomStageResponse[] = [
  {
    stage_id: 'stage-none',
    project_id: 'project-1',
    name: '없음',
    color: '#94A3B8',
    displayOrder: 0,
    isSystemDefault: true,
    createdAt: '2024-01-01T00:00:00Z',
    updatedAt: '2024-01-01T00:00:00Z',
  },
  {
    stage_id: 'stage-waiting',
    project_id: 'project-1',
    name: '대기',
    color: '#F59E0B',
    displayOrder: 1,
    isSystemDefault: true,
    createdAt: '2024-01-01T00:00:00Z',
    updatedAt: '2024-01-01T00:00:00Z',
  },
  {
    stage_id: 'stage-progress',
    project_id: 'project-1',
    name: '진행중',
    color: '#3B82F6',
    displayOrder: 2,
    isSystemDefault: true,
    createdAt: '2024-01-01T00:00:00Z',
    updatedAt: '2024-01-01T00:00:00Z',
  },
  {
    stage_id: 'stage-done',
    project_id: 'project-1',
    name: '완료',
    color: '#10B981',
    displayOrder: 3,
    isSystemDefault: true,
    createdAt: '2024-01-01T00:00:00Z',
    updatedAt: '2024-01-01T00:00:00Z',
  },
];

// 목업: Role 데이터
const MOCK_ROLES: CustomRoleResponse[] = [
  {
    role_id: 'role-none',
    project_id: 'project-1',
    name: '없음',
    color: '#94A3B8',
    displayOrder: 0,
    isSystemDefault: true,
    createdAt: '2024-01-01T00:00:00Z',
    updatedAt: '2024-01-01T00:00:00Z',
  },
  {
    role_id: 'role-frontend',
    project_id: 'project-1',
    name: '프론트엔드',
    color: '#8B5CF6',
    displayOrder: 1,
    isSystemDefault: false,
    createdAt: '2024-01-01T00:00:00Z',
    updatedAt: '2024-01-01T00:00:00Z',
  },
  {
    role_id: 'role-backend',
    project_id: 'project-1',
    name: '백엔드',
    color: '#EC4899',
    displayOrder: 2,
    isSystemDefault: false,
    createdAt: '2024-01-01T00:00:00Z',
    updatedAt: '2024-01-01T00:00:00Z',
  },
  {
    role_id: 'role-design',
    project_id: 'project-1',
    name: '디자인',
    color: '#F59E0B',
    displayOrder: 3,
    isSystemDefault: false,
    createdAt: '2024-01-01T00:00:00Z',
    updatedAt: '2024-01-01T00:00:00Z',
  },
];

// 목업: Importance 데이터
const MOCK_IMPORTANCES: CustomImportanceResponse[] = [
  {
    importance_id: 'importance-none',
    project_id: 'project-1',
    name: '없음',
    color: '#94A3B8',
    displayOrder: 0,
    isSystemDefault: true,
    createdAt: '2024-01-01T00:00:00Z',
    updatedAt: '2024-01-01T00:00:00Z',
  },
  {
    importance_id: 'importance-low',
    project_id: 'project-1',
    name: '낮음',
    color: '#10B981',
    displayOrder: 1,
    isSystemDefault: true,
    createdAt: '2024-01-01T00:00:00Z',
    updatedAt: '2024-01-01T00:00:00Z',
  },
  {
    importance_id: 'importance-medium',
    project_id: 'project-1',
    name: '보통',
    color: '#3B82F6',
    displayOrder: 2,
    isSystemDefault: true,
    createdAt: '2024-01-01T00:00:00Z',
    updatedAt: '2024-01-01T00:00:00Z',
  },
  {
    importance_id: 'importance-high',
    project_id: 'project-1',
    name: '높음',
    color: '#F59E0B',
    displayOrder: 3,
    isSystemDefault: true,
    createdAt: '2024-01-01T00:00:00Z',
    updatedAt: '2024-01-01T00:00:00Z',
  },
  {
    importance_id: 'importance-urgent',
    project_id: 'project-1',
    name: '긴급',
    color: '#EF4444',
    displayOrder: 4,
    isSystemDefault: true,
    createdAt: '2024-01-01T00:00:00Z',
    updatedAt: '2024-01-01T00:00:00Z',
  },
];

// 목업: Board(카드) 데이터
const MOCK_BOARDS: BoardResponse[] = [
  {
    board_id: 'board-1',
    project_id: 'project-1',
    title: '로그인 페이지 구현',
    content: 'JWT 인증 방식으로 로그인/로그아웃 기능 구현',
    stage: MOCK_STAGES.find((s) => s.name === '진행중'),
    roles: [MOCK_ROLES.find((r) => r.name === '프론트엔드')],
    importance: MOCK_IMPORTANCES.find((i) => i.name === '높음'),
    assignee: {
      user_id: 'user-123',
      name: '김개발',
      email: 'dev.kim@orangecloud.com',
      isActive: true,
    },
    author: {
      user_id: 'user-123',
      name: '김개발',
      email: 'dev.kim@orangecloud.com',
      isActive: true,
    },
    dueDate: '2024-02-15T00:00:00Z',
    createdAt: '2024-01-20T00:00:00Z',
    updatedAt: '2024-01-25T00:00:00Z',
  },
  {
    board_id: 'board-2',
    project_id: 'project-1',
    title: 'API 엔드포인트 설계',
    content: 'RESTful API 설계 및 Swagger 문서 작성',
    stage: MOCK_STAGES.find((s) => s.name === '완료'),
    roles: [MOCK_ROLES.find((r) => r.name === '백엔드')],
    importance: MOCK_IMPORTANCES.find((i) => i.name === '높음'),
    assignee: {
      user_id: 'user-456',
      name: '이디자인',
      email: 'design.lee@orangecloud.com',
      isActive: true,
    },
    author: {
      user_id: 'user-456',
      name: '이디자인',
      email: 'design.lee@orangecloud.com',
      isActive: true,
    },
    createdAt: '2024-01-15T00:00:00Z',
    updatedAt: '2024-01-22T00:00:00Z',
  },
  {
    board_id: 'board-3',
    project_id: 'project-1',
    title: 'UI 컴포넌트 디자인',
    content: '버튼, 인풋, 모달 등 기본 컴포넌트 디자인',
    stage: MOCK_STAGES.find((s) => s.name === '대기'),
    roles: [MOCK_ROLES.find((r) => r.name === '디자인')],
    importance: MOCK_IMPORTANCES.find((i) => i.name === '보통'),
    assignee: {
      user_id: 'user-789',
      name: '박프론트',
      email: 'front.park@orangecloud.com',
      isActive: true,
    },
    author: {
      user_id: 'user-123',
      name: '김개발',
      email: 'dev.kim@orangecloud.com',
      isActive: true,
    },
    dueDate: '2024-02-20T00:00:00Z',
    createdAt: '2024-01-18T00:00:00Z',
    updatedAt: '2024-01-18T00:00:00Z',
  },
  {
    board_id: 'board-4',
    project_id: 'project-1',
    title: '데이터베이스 스키마 설계',
    content: 'PostgreSQL 테이블 구조 및 관계 정의',
    stage: MOCK_STAGES.find((s) => s.name === '완료'),
    roles: [MOCK_ROLES.find((r) => r.name === '백엔드')],
    importance: MOCK_IMPORTANCES.find((i) => i.name === '긴급'),
    assignee: {
      user_id: 'user-456',
      name: '이디자인',
      email: 'design.lee@orangecloud.com',
      isActive: true,
    },
    author: {
      user_id: 'user-456',
      name: '이디자인',
      email: 'design.lee@orangecloud.com',
      isActive: true,
    },
    createdAt: '2024-01-10T00:00:00Z',
    updatedAt: '2024-01-20T00:00:00Z',
  },
  {
    board_id: 'board-5',
    project_id: 'project-1',
    title: 'CI/CD 파이프라인 구축',
    content: 'GitHub Actions를 이용한 자동 배포 설정',
    stage: MOCK_STAGES.find((s) => s.name === '진행중'),
    roles: [MOCK_ROLES.find((r) => r.name === '백엔드')],
    importance: MOCK_IMPORTANCES.find((i) => i.name === '보통'),
    assignee: {
      user_id: 'user-202',
      name: '최데브옵스',
      email: 'devops.choi@orangecloud.com',
      isActive: true,
    },
    author: {
      user_id: 'user-202',
      name: '최데브옵스',
      email: 'devops.choi@orangecloud.com',
      isActive: true,
    },
    dueDate: '2024-02-10T00:00:00Z',
    createdAt: '2024-01-12T00:00:00Z',
    updatedAt: '2024-01-26T00:00:00Z',
  },
  {
    board_id: 'board-6',
    project_id: 'project-1',
    title: '사용자 피드백 수집',
    content: '베타 테스트 사용자 의견 정리 및 분석',
    stage: MOCK_STAGES.find((s) => s.name === '대기'),
    roles: [MOCK_ROLES.find((r) => r.name === '디자인')],
    importance: MOCK_IMPORTANCES.find((i) => i.name === '낮음'),
    author: {
      user_id: 'user-789',
      name: '박프론트',
      email: 'front.park@orangecloud.com',
      isActive: true,
    },
    createdAt: '2024-01-22T00:00:00Z',
    updatedAt: '2024-01-22T00:00:00Z',
  },
];

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
  workspace_id: string,
  token: string,
): Promise<ProjectResponse[]> => {
  if (USE_MOCK_DATA) {
    console.log('[MOCK] getProjects 호출:', workspace_id);
    return new Promise((resolve) => {
      setTimeout(() => {
        const filtered = MOCK_PROJECTS.filter((p) => p.workspace_id === workspace_id);
        resolve(filtered);
      }, 300);
    });
  }

  try {
    const response = await boardService.get('/api/projects', {
      params: { workspaceId: workspace_id },
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
 * GET /api/projects/{project_id}
 * @param project_id 프로젝트 ID
 * @param token 액세스 토큰
 * @returns 프로젝트 정보
 */
export const getProject = async (project_id: string, token: string): Promise<ProjectResponse> => {
  try {
    const response = await boardService.get(`/api/projects/${project_id}`, {
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
  if (USE_MOCK_DATA) {
    console.log('[MOCK] createProject 호출:', data);
    return new Promise((resolve) => {
      setTimeout(() => {
        const newProject: ProjectResponse = {
          project_id: `project-${Date.now()}`,
          name: data.name,
          description: data.description,
          workspace_id: data.workspace_id,
          owner_id: 'user-123',
          ownerName: '김개발',
          ownerEmail: 'dev.kim@orangecloud.com',
          createdAt: new Date().toISOString(),
          updatedAt: new Date().toISOString(),
        };
        MOCK_PROJECTS.push(newProject);
        resolve(newProject);
      }, 300);
    });
  }

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
 * PUT /api/projects/{project_id}
 * @param project_id 프로젝트 ID
 * @param data 업데이트 정보
 * @param token 액세스 토큰
 * @returns 업데이트된 프로젝트
 */
export const updateProject = async (
  project_id: string,
  data: Partial<CreateProjectRequest>,
  token: string,
): Promise<ProjectResponse> => {
  try {
    const response = await boardService.put(`/api/projects/${project_id}`, data, {
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
 * DELETE /api/projects/{project_id}
 * @param project_id 프로젝트 ID
 * @param token 액세스 토큰
 * @returns 응답 메시지
 */
export const deleteProject = async (project_id: string, token: string): Promise<any> => {
  try {
    const response = await boardService.delete(`/api/projects/${project_id}`, {
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
  boardId: string;
  projectId: string;
  title: string;
  content?: string;
  assignee?: any;
  author?: any;
  dueDate?: string;
  createdAt: string;
  updatedAt: string;
  customFields?: Record<string, any>; // 통합된 커스텀 필드
  position?: string; // Board position in view

  // 하위 호환성을 위한 레거시 필드 (deprecated)
  stage?: any;
  roles?: any[];
  importance?: any;
  assignees?: any[];
}

export interface CreateBoardRequest {
  projectId: string;
  title: string;
  content?: string;

  // Legacy fields (deprecated - use custom fields instead)
  stageId?: string;
  importanceId?: string;
  roleIds?: string[];

  assigneeId?: string;
  dueDate?: string; // ISO 8601 format
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
 * @param project_id 프로젝트 ID
 * @param token 액세스 토큰
 * @param filters 필터 옵션 (stageId, roleId, importanceId, assigneeId, authorId, page, limit)
 * @returns 보드 배열
 */
export const getBoards = async (
  project_id: string,
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
    console.log('[MOCK] getBoards 호출:', project_id, filters);
    return new Promise((resolve) => {
      setTimeout(() => {
        let filtered = MOCK_BOARDS.filter((b) => b.project_id === project_id);

        // 필터 적용
        if (filters?.stageId) {
          filtered = filtered.filter((b) => b.stage?.stage_id === filters.stageId);
        }
        if (filters?.roleId) {
          filtered = filtered.filter((b) => b.roles?.some((r) => r?.role_id === filters.roleId));
        }
        if (filters?.importanceId) {
          filtered = filtered.filter((b) => b.importance?.importance_id === filters.importanceId);
        }
        if (filters?.assigneeId) {
          filtered = filtered.filter((b) => b.assignee?.user_id === filters.assigneeId);
        }
        if (filters?.authorId) {
          filtered = filtered.filter((b) => b.author?.user_id === filters.authorId);
        }

        resolve({
          boards: filtered,
          total: filtered.length,
          page: filters?.page || 1,
          limit: filters?.limit || 20,
        });
      }, 300);
    });
  }

  try {
    const params = { projectId: project_id, ...filters };
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
 * GET /api/boards/{board_id}
 * @param board_id 보드 ID
 * @param token 액세스 토큰
 * @returns 보드 정보
 */
export const getBoard = async (board_id: string, token: string): Promise<BoardResponse> => {
  if (USE_MOCK_DATA) {
    console.log('[MOCK] getBoard 호출:', board_id);
    return new Promise((resolve, reject) => {
      setTimeout(() => {
        const board = MOCK_BOARDS.find((b) => b.board_id === board_id);
        if (board) {
          resolve(board);
        } else {
          reject(new Error('보드를 찾을 수 없습니다.'));
        }
      }, 300);
    });
  }

  try {
    const response = await boardService.get(`/api/boards/${board_id}`, {
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
 * PUT /api/boards/{board_id}
 * @param board_id 보드 ID
 * @param data 업데이트 정보
 * @param token 액세스 토큰
 * @returns 업데이트된 보드
 */
export const updateBoard = async (
  board_id: string,
  data: Partial<CreateBoardRequest>,
  token: string,
): Promise<BoardResponse> => {
  try {
    const response = await boardService.put(`/api/boards/${board_id}`, data, {
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
 * DELETE /api/boards/{board_id}
 * @param board_id 보드 ID
 * @param token 액세스 토큰
 * @returns 응답 메시지
 */
export const deleteBoard = async (board_id: string, token: string): Promise<any> => {
  try {
    const response = await boardService.delete(`/api/boards/${board_id}`, {
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

// ==================== 새로운 Field/Option API 타입 ====================

export interface FieldResponse {
  field_id: string;
  project_id: string;
  name: string;
  fieldType: string;
  description?: string;
  displayOrder: number;
  isRequired: boolean;
  isSystemDefault: boolean;
  config: any;
  canEditRoles: string[];
  createdAt: string;
  updatedAt: string;
}

export interface OptionResponse {
  option_id: string;
  field_id: string;
  label: string;
  color?: string;
  description?: string;
  displayOrder: number;
  createdAt: string;
  updatedAt: string;
}

export interface CreateFieldRequest {
  project_id: string;
  name: string;
  field_type: string;
  description?: string;
  is_required?: boolean;
  config?: any;
}

export interface CreateOptionRequest {
  field_id: string;
  label: string;
  color?: string;
  description?: string;
}

export interface UpdateOptionRequest {
  label?: string;
  color?: string;
  description?: string;
}

// ==================== 기존 Custom Fields 호환 타입 (유지) ====================

export interface CustomStageResponse {
  stage_id: string;
  project_id: string;
  name: string;
  color?: string;
  displayOrder: number;
  isSystemDefault: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface CustomRoleResponse {
  role_id: string;
  project_id: string;
  name: string;
  color?: string;
  displayOrder: number;
  isSystemDefault: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface CustomImportanceResponse {
  importance_id: string;
  project_id: string;
  name: string;
  color?: string;
  displayOrder: number;
  isSystemDefault: boolean;
  createdAt: string;
  updatedAt: string;
}

// ==================== 새로운 Field/Option API 헬퍼 함수 ====================

/**
 * 프로젝트의 모든 Field를 조회합니다.
 * GET /api/projects/{project_id}/fields
 */
const getProjectFields = async (
  project_id: string,
  token: string,
): Promise<FieldResponse[]> => {
  try {
    const response = await boardService.get(`/api/projects/${project_id}/fields`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    return response.data.data || [];
  } catch (error) {
    console.error('getProjectFields error:', error);
    throw error;
  }
};

/**
 * 특정 Field의 모든 Option을 조회합니다.
 * GET /api/fields/{field_id}/options
 */
const getFieldOptions = async (
  field_id: string,
  token: string,
): Promise<OptionResponse[]> => {
  try {
    const response = await boardService.get(`/api/fields/${field_id}/options`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    return response.data.data || [];
  } catch (error) {
    console.error('getFieldOptions error:', error);
    throw error;
  }
};

/**
 * Field를 생성합니다.
 * POST /api/fields
 */
const createFieldInternal = async (
  data: CreateFieldRequest,
  token: string,
): Promise<FieldResponse> => {
  try {
    const response = await boardService.post('/api/fields', data, {
      headers: { Authorization: `Bearer ${token}` },
    });
    return response.data.data;
  } catch (error) {
    console.error('createField error:', error);
    throw error;
  }
};

/**
 * Field Option을 생성합니다.
 * POST /api/field-options
 */
const createFieldOption = async (
  data: CreateOptionRequest,
  token: string,
): Promise<OptionResponse> => {
  try {
    const response = await boardService.post('/api/field-options', data, {
      headers: { Authorization: `Bearer ${token}` },
    });
    return response.data.data;
  } catch (error) {
    console.error('createFieldOption error:', error);
    throw error;
  }
};

/**
 * Field Option을 수정합니다.
 * PATCH /api/field-options/{option_id}
 */
const updateFieldOption = async (
  option_id: string,
  data: UpdateOptionRequest,
  token: string,
): Promise<OptionResponse> => {
  try {
    const response = await boardService.patch(`/api/field-options/${option_id}`, data, {
      headers: { Authorization: `Bearer ${token}` },
    });
    return response.data.data;
  } catch (error) {
    console.error('updateFieldOption error:', error);
    throw error;
  }
};

/**
 * Field Option을 삭제합니다.
 * DELETE /api/field-options/{option_id}
 */
const deleteFieldOption = async (
  option_id: string,
  token: string,
): Promise<void> => {
  try {
    await boardService.delete(`/api/field-options/${option_id}`, {
      headers: { Authorization: `Bearer ${token}` },
    });
  } catch (error) {
    console.error('deleteFieldOption error:', error);
    throw error;
  }
};

// ==================== 기존 인터페이스를 유지하면서 새 API 사용 ====================

/**
 * 프로젝트의 모든 Stage를 조회합니다.
 * (내부적으로 새로운 Field/Option API를 사용하여 변환)
 * @param project_id 프로젝트 ID
 * @param token 액세스 토큰
 * @returns Stage 배열
 */
export const getProjectStages = async (
  project_id: string,
  token: string,
): Promise<CustomStageResponse[]> => {
  if (USE_MOCK_DATA) {
    console.log('[MOCK] getProjectStages 호출:', project_id);
    return new Promise((resolve) => {
      setTimeout(() => {
        const filtered = MOCK_STAGES.filter((s) => s.project_id === project_id);
        resolve(filtered);
      }, 200);
    });
  }

  try {
    // 1. 프로젝트의 모든 Field 조회
    const fields = await getProjectFields(project_id, token);

    // 2. Stage Field 찾기 (이름으로 구분)
    const stageField = fields.find(
      f => f.name === 'Stage' || f.name === '진행 단계' || f.name === 'stage'
    );

    if (!stageField) {
      console.log('[getProjectStages] Stage Field가 없습니다. 빈 배열 반환');
      return [];
    }

    // 3. Stage Field의 Options 조회
    const options = await getFieldOptions(stageField.field_id, token);

    // 4. Backend 응답을 CustomStageResponse 형식으로 변환
    return options.map(opt => ({
      stage_id: opt.option_id,
      project_id: project_id,
      name: opt.label,
      color: opt.color,
      displayOrder: opt.displayOrder,
      isSystemDefault: stageField.isSystemDefault,
      createdAt: opt.createdAt,
      updatedAt: opt.updatedAt,
    }));
  } catch (error) {
    console.error('getProjectStages error:', error);
    throw error;
  }
};

/**
 * 프로젝트의 모든 Role을 조회합니다.
 * (내부적으로 새로운 Field/Option API를 사용하여 변환)
 * @param project_id 프로젝트 ID
 * @param token 액세스 토큰
 * @returns Role 배열
 */
export const getProjectRoles = async (
  project_id: string,
  token: string,
): Promise<CustomRoleResponse[]> => {
  if (USE_MOCK_DATA) {
    console.log('[MOCK] getProjectRoles 호출:', project_id);
    return new Promise((resolve) => {
      setTimeout(() => {
        const filtered = MOCK_ROLES.filter((r) => r.project_id === project_id);
        resolve(filtered);
      }, 200);
    });
  }

  try {
    const fields = await getProjectFields(project_id, token);
    const roleField = fields.find(
      f => f.name === 'Role' || f.name === '역할' || f.name === 'role'
    );

    if (!roleField) {
      console.log('[getProjectRoles] Role Field가 없습니다. 빈 배열 반환');
      return [];
    }

    const options = await getFieldOptions(roleField.field_id, token);

    return options.map(opt => ({
      role_id: opt.option_id,
      project_id: project_id,
      name: opt.label,
      color: opt.color,
      displayOrder: opt.displayOrder,
      isSystemDefault: roleField.isSystemDefault,
      createdAt: opt.createdAt,
      updatedAt: opt.updatedAt,
    }));
  } catch (error) {
    console.error('getProjectRoles error:', error);
    throw error;
  }
};

/**
 * 프로젝트의 모든 Importance를 조회합니다.
 * (내부적으로 새로운 Field/Option API를 사용하여 변환)
 * @param project_id 프로젝트 ID
 * @param token 액세스 토큰
 * @returns Importance 배열
 */
export const getProjectImportances = async (
  project_id: string,
  token: string,
): Promise<CustomImportanceResponse[]> => {
  if (USE_MOCK_DATA) {
    console.log('[MOCK] getProjectImportances 호출:', project_id);
    return new Promise((resolve) => {
      setTimeout(() => {
        const filtered = MOCK_IMPORTANCES.filter((i) => i.project_id === project_id);
        resolve(filtered);
      }, 200);
    });
  }

  try {
    const fields = await getProjectFields(project_id, token);
    const importanceField = fields.find(
      f => f.name === 'Importance' || f.name === '중요도' || f.name === 'importance' || f.name === 'Priority'
    );

    if (!importanceField) {
      console.log('[getProjectImportances] Importance Field가 없습니다. 빈 배열 반환');
      return [];
    }

    const options = await getFieldOptions(importanceField.field_id, token);

    return options.map(opt => ({
      importance_id: opt.option_id,
      project_id: project_id,
      name: opt.label,
      color: opt.color,
      displayOrder: opt.displayOrder,
      isSystemDefault: importanceField.isSystemDefault,
      createdAt: opt.createdAt,
      updatedAt: opt.updatedAt,
    }));
  } catch (error) {
    console.error('getProjectImportances error:', error);
    throw error;
  }
};

// ============================================================================
// Custom Fields CRUD API
// ============================================================================

export interface CreateCustomStageRequest {
  project_id: string;
  name: string;
  color: string;
}

export interface UpdateCustomStageRequest {
  name: string;
  color: string;
}

export interface CreateCustomRoleRequest {
  project_id: string;
  name: string;
  color: string;
}

export interface UpdateCustomRoleRequest {
  name: string;
  color: string;
}

export interface CreateCustomImportanceRequest {
  project_id: string;
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
 * (내부적으로 Field Option을 생성)
 */
export const createStage = async (
  data: CreateCustomStageRequest,
  token: string,
): Promise<CustomStageResponse> => {
  try {
    // 1. Stage Field 찾기 또는 생성
    const fields = await getProjectFields(data.project_id, token);
    let stageField = fields.find(
      f => f.name === 'Stage' || f.name === '진행 단계' || f.name === 'stage'
    );

    if (!stageField) {
      // Stage Field가 없으면 생성
      stageField = await createFieldInternal({
        project_id: data.project_id,
        name: 'Stage',
        field_type: 'single_select',
        description: '작업 진행 단계',
        is_required: true,
      }, token);
    }

    // 2. Option 생성
    const option = await createFieldOption({
      field_id: stageField.field_id,
      label: data.name,
      color: data.color,
    }, token);

    // 3. CustomStageResponse 형식으로 변환
    return {
      stage_id: option.option_id,
      project_id: data.project_id,
      name: option.label,
      color: option.color,
      displayOrder: option.displayOrder,
      isSystemDefault: false,
      createdAt: option.createdAt,
      updatedAt: option.updatedAt,
    };
  } catch (error) {
    console.error('createStage error:', error);
    throw error;
  }
};

/**
 * Stage를 수정합니다.
 * (내부적으로 Field Option을 수정)
 */
export const updateStage = async (
  stage_id: string,
  data: UpdateCustomStageRequest,
  token: string,
): Promise<CustomStageResponse> => {
  try {
    // stage_id는 option_id
    const option = await updateFieldOption(stage_id, {
      label: data.name,
      color: data.color,
    }, token);

    // CustomStageResponse 형식으로 변환
    // project_id는 option에 없으므로 추가 조회 필요하거나, 기존 값 유지
    return {
      stage_id: option.option_id,
      project_id: '', // Option에는 project_id가 없음 (필요시 Field 조회)
      name: option.label,
      color: option.color,
      displayOrder: option.displayOrder,
      isSystemDefault: false,
      createdAt: option.createdAt,
      updatedAt: option.updatedAt,
    };
  } catch (error) {
    console.error('updateStage error:', error);
    throw error;
  }
};

/**
 * Stage를 삭제합니다.
 * (내부적으로 Field Option을 삭제)
 */
export const deleteStage = async (stage_id: string, token: string): Promise<void> => {
  try {
    await deleteFieldOption(stage_id, token);
  } catch (error) {
    console.error('deleteStage error:', error);
    throw error;
  }
};

/**
 * Role을 생성합니다.
 * (내부적으로 Field Option을 생성)
 */
export const createRole = async (
  data: CreateCustomRoleRequest,
  token: string,
): Promise<CustomRoleResponse> => {
  try {
    const fields = await getProjectFields(data.project_id, token);
    let roleField = fields.find(
      f => f.name === 'Role' || f.name === '역할' || f.name === 'role'
    );

    if (!roleField) {
      roleField = await createFieldInternal({
        project_id: data.project_id,
        name: 'Role',
        field_type: 'single_select',
        description: '담당 역할',
        is_required: true,
      }, token);
    }

    const option = await createFieldOption({
      field_id: roleField.field_id,
      label: data.name,
      color: data.color,
    }, token);

    return {
      role_id: option.option_id,
      project_id: data.project_id,
      name: option.label,
      color: option.color,
      displayOrder: option.displayOrder,
      isSystemDefault: false,
      createdAt: option.createdAt,
      updatedAt: option.updatedAt,
    };
  } catch (error) {
    console.error('createRole error:', error);
    throw error;
  }
};

/**
 * Role을 수정합니다.
 * (내부적으로 Field Option을 수정)
 */
export const updateRole = async (
  role_id: string,
  data: UpdateCustomRoleRequest,
  token: string,
): Promise<CustomRoleResponse> => {
  try {
    const option = await updateFieldOption(role_id, {
      label: data.name,
      color: data.color,
    }, token);

    return {
      role_id: option.option_id,
      project_id: '',
      name: option.label,
      color: option.color,
      displayOrder: option.displayOrder,
      isSystemDefault: false,
      createdAt: option.createdAt,
      updatedAt: option.updatedAt,
    };
  } catch (error) {
    console.error('updateRole error:', error);
    throw error;
  }
};

/**
 * Role을 삭제합니다.
 * (내부적으로 Field Option을 삭제)
 */
export const deleteRole = async (role_id: string, token: string): Promise<void> => {
  try {
    await deleteFieldOption(role_id, token);
  } catch (error) {
    console.error('deleteRole error:', error);
    throw error;
  }
};

/**
 * Importance를 생성합니다.
 * (내부적으로 Field Option을 생성)
 */
export const createImportance = async (
  data: CreateCustomImportanceRequest,
  token: string,
): Promise<CustomImportanceResponse> => {
  try {
    const fields = await getProjectFields(data.project_id, token);
    let importanceField = fields.find(
      f => f.name === 'Importance' || f.name === '중요도' || f.name === 'importance' || f.name === 'Priority'
    );

    if (!importanceField) {
      importanceField = await createFieldInternal({
        project_id: data.project_id,
        name: 'Importance',
        field_type: 'single_select',
        description: '작업 중요도',
        is_required: false,
      }, token);
    }

    const option = await createFieldOption({
      field_id: importanceField.field_id,
      label: data.name,
      color: data.color,
      description: `Level: ${data.level}`,
    }, token);

    return {
      importance_id: option.option_id,
      project_id: data.project_id,
      name: option.label,
      color: option.color,
      displayOrder: option.displayOrder,
      isSystemDefault: false,
      createdAt: option.createdAt,
      updatedAt: option.updatedAt,
    };
  } catch (error) {
    console.error('createImportance error:', error);
    throw error;
  }
};

/**
 * Importance를 수정합니다.
 * (내부적으로 Field Option을 수정)
 */
export const updateImportance = async (
  importance_id: string,
  data: UpdateCustomImportanceRequest,
  token: string,
): Promise<CustomImportanceResponse> => {
  try {
    const option = await updateFieldOption(importance_id, {
      label: data.name,
      color: data.color,
      description: `Level: ${data.level}`,
    }, token);

    return {
      importance_id: option.option_id,
      project_id: '',
      name: option.label,
      color: option.color,
      displayOrder: option.displayOrder,
      isSystemDefault: false,
      createdAt: option.createdAt,
      updatedAt: option.updatedAt,
    };
  } catch (error) {
    console.error('updateImportance error:', error);
    throw error;
  }
};

/**
 * Importance를 삭제합니다.
 * (내부적으로 Field Option을 삭제)
 */
const deleteImportance = async (importance_id: string, token: string): Promise<void> => {
  try {
    await deleteFieldOption(importance_id, token);
  } catch (error) {
    console.error('deleteImportance error:', error);
    throw error;
  }
};

// Export deleteImportance
export { deleteImportance };

// ============================================================================
// 기본 필드 초기화 함수
// ============================================================================

/**
 * 프로젝트의 기본 Custom Fields (Stage, Role)를 자동 생성합니다.
 * Dashboard 첫 로드 시 호출됩니다.
 */
export const initializeDefaultFields = async (
  project_id: string,
  token: string,
): Promise<void> => {
  try {
    console.log('[initializeDefaultFields] 기본 필드 초기화 시작:', project_id);

    // 1. Stage Field 생성
    const stageField = await createFieldInternal({
      project_id,
      name: 'Stage',
      field_type: 'single_select',
      description: '작업 진행 단계',
      is_required: true,
    }, token);

    // 기본 Stage Options 생성
    await createFieldOption({
      field_id: stageField.field_id,
      label: '대기',
      color: '#F59E0B',
    }, token);

    await createFieldOption({
      field_id: stageField.field_id,
      label: '진행중',
      color: '#3B82F6',
    }, token);

    await createFieldOption({
      field_id: stageField.field_id,
      label: '완료',
      color: '#10B981',
    }, token);

    // 2. Role Field 생성
    const roleField = await createFieldInternal({
      project_id,
      name: 'Role',
      field_type: 'single_select',
      description: '담당 역할',
      is_required: true,
    }, token);

    // 기본 Role Options 생성
    await createFieldOption({
      field_id: roleField.field_id,
      label: '개발',
      color: '#3B82F6',
    }, token);

    await createFieldOption({
      field_id: roleField.field_id,
      label: '디자인',
      color: '#8B5CF6',
    }, token);

    await createFieldOption({
      field_id: roleField.field_id,
      label: '기획',
      color: '#EC4899',
    }, token);

    // 3. Importance Field 생성 (선택사항)
    const importanceField = await createFieldInternal({
      project_id,
      name: 'Importance',
      field_type: 'single_select',
      description: '작업 중요도',
      is_required: false,
    }, token);

    await createFieldOption({
      field_id: importanceField.field_id,
      label: '낮음',
      color: '#94A3B8',
    }, token);

    await createFieldOption({
      field_id: importanceField.field_id,
      label: '보통',
      color: '#F59E0B',
    }, token);

    await createFieldOption({
      field_id: importanceField.field_id,
      label: '높음',
      color: '#EF4444',
    }, token);

    console.log('[initializeDefaultFields] 기본 필드 초기화 완료');
  } catch (error) {
    console.error('[initializeDefaultFields] 기본 필드 초기화 실패:', error);
    throw error;
  }
};

// ============================================================================
// Orders API (제거됨 - 사용하지 않음)
// ============================================================================

/**
 * 아래 Orders API 함수들은 더 이상 사용되지 않습니다.
 * Backend에서 제거되었고, Frontend에서도 사용하지 않습니다.
 *
 * - getRoleBasedBoardView
 * - getStageBasedBoardView
 * - updateStageColumnOrder
 * - updateStageBoardOrder
 */

/* Orders API 제거 완료 - 더 이상 사용하지 않습니다. */

// ============================================================================
// Comment API
// ============================================================================

export interface CommentResponse {
  comment_id: string;
  user_id: string;
  userName: string;
  userAvatar: string;
  content: string;
  createdAt: string;
  updatedAt: string;
}

export interface CreateCommentRequest {
  board_id: string;
  content: string;
}

export interface UpdateCommentRequest {
  content: string;
}

/**
 * 보드의 모든 댓글을 조회합니다.
 * GET /api/comments
 * @param board_id 보드 ID
 * @param token 액세스 토큰
 * @returns 댓글 배열
 */
export const getComments = async (
  board_id: string,
  token: string,
): Promise<CommentResponse[]> => {
  try {
    const response = await boardService.get('/api/comments', {
      params: { boardId: board_id },
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
 * PUT /api/comments/{comment_id}
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
 * DELETE /api/comments/{comment_id}
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
