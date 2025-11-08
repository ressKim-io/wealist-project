import { userRepoClient } from '../apiConfig';
import { AxiosResponse } from 'axios';

/**
 * ========================================
 * [백엔드 개발자 참고]
 * ========================================
 *
 * 목업 모드 전환:
 * - USE_MOCK_DATA = true: 목업 데이터 사용 (백엔드 없이 프론트엔드 개발)
 * - USE_MOCK_DATA = false: 실제 API 호출 (백엔드 연동)
 *
 * 백엔드 API 구현 후 아래 플래그를 false로 변경하세요.
 */
const USE_MOCK_DATA = true;

// --- DTO Interfaces ---

export interface AuthResponse {
  accessToken: string;
  refreshToken: string;
  userId: string; // (format: uuid)
  name: string;
  email: string;
  tokenType: string; // e.g., "bearer"
}

export interface WorkspaceResponse {
  id: string;
  name: string;
  description: string;
  ownerId: string;
  ownerName: string;
  ownerEmail: string;
  createdAt: string;
  updatedAt: string;
}

export interface CreateWorkspaceRequest {
  name: string;
  // description?: string; // API 스펙에 name만 required이므로 일단 제외
}

export interface UserProfileResponse {
  profileId: string;
  userId: string;
  workspaceId?: string | null; // [추가] 워크스페이스별 프로필용
  name: string;
  email: string | null;
  profileImageUrl: string | null;
  createdAt: string;
  updatedAt: string;
}

export interface UpdateProfileRequest {
  name?: string;
  email?: string;
  profileImageUrl?: string;
}

// ========================================
// 목업 데이터 (백엔드 개발자가 수정 가능)
// ========================================

/**
 * [백엔드 개발자 참고]
 *
 * 아래 목업 데이터는 프론트엔드 개발용입니다.
 * 실제 백엔드 API를 구현하면 이 데이터는 사용되지 않습니다.
 * (USE_MOCK_DATA = false일 때)
 */

// 목업: 워크스페이스 목록
const MOCK_WORKSPACES: WorkspaceResponse[] = [
  {
    id: 'workspace-1',
    name: '오렌지클라우드',
    description: '메인 워크스페이스',
    ownerId: 'user-123',
    ownerName: '김개발',
    ownerEmail: 'dev.kim@example.com',
    createdAt: '2024-01-01T00:00:00Z',
    updatedAt: '2024-01-01T00:00:00Z',
  },
  {
    id: 'workspace-2',
    name: '데이터랩',
    description: '데이터 분석 팀',
    ownerId: 'user-123',
    ownerName: '김개발',
    ownerEmail: 'dev.kim@example.com',
    createdAt: '2024-01-02T00:00:00Z',
    updatedAt: '2024-01-02T00:00:00Z',
  },
  {
    id: 'workspace-3',
    name: '마케팅팀',
    description: '마케팅 전략팀',
    ownerId: 'user-123',
    ownerName: '김개발',
    ownerEmail: 'dev.kim@example.com',
    createdAt: '2024-01-03T00:00:00Z',
    updatedAt: '2024-01-03T00:00:00Z',
  },
];

// 목업: 기본 프로필 (workspaceId = null)
let MOCK_DEFAULT_PROFILE: UserProfileResponse = {
  profileId: 'profile-default-001',
  userId: 'user-123',
  workspaceId: null,
  name: '김개발',
  email: 'dev.kim@example.com',
  profileImageUrl: null,
  createdAt: '2024-01-01T00:00:00Z',
  updatedAt: '2024-01-01T00:00:00Z',
};

// 목업: 워크스페이스별 프로필 (userId + workspaceId)
let MOCK_WORKSPACE_PROFILES: Record<string, UserProfileResponse> = {
  'workspace-1': {
    profileId: 'profile-ws-001',
    userId: 'user-123',
    workspaceId: 'workspace-1',
    name: '김개발 (오렌지클라우드)',
    email: 'dev.kim@orangecloud.com',
    profileImageUrl: null,
    createdAt: '2024-01-02T00:00:00Z',
    updatedAt: '2024-01-02T00:00:00Z',
  },
};

// ========================================
// API Service Functions (목업/실제 API 자동 전환)
// ========================================

/**
 * 워크스페이스 목록 조회
 *
 * [백엔드 API]
 * - GET /api/workspaces
 * - Headers: Authorization: Bearer {accessToken}
 * - Response: WorkspaceResponse[]
 */
export const getWorkspaces = async (accessToken: string): Promise<WorkspaceResponse[]> => {
  if (USE_MOCK_DATA) {
    // 목업 모드
    console.log('[MOCK] getWorkspaces 호출');
    return new Promise((resolve) => {
      setTimeout(() => resolve(MOCK_WORKSPACES), 300); // 네트워크 딜레이 시뮬레이션
    });
  }

  // 실제 API 호출
  const response: AxiosResponse<WorkspaceResponse[]> = await userRepoClient.get('/api/workspaces', {
    headers: { Authorization: `Bearer ${accessToken}` },
  });
  return response.data;
};

/**
 * 워크스페이스 생성
 *
 * [백엔드 API]
 * - POST /api/workspaces
 * - Headers: Authorization: Bearer {accessToken}
 * - Body: CreateWorkspaceRequest
 * - Response: WorkspaceResponse
 */
export const createWorkspace = async (
  data: CreateWorkspaceRequest,
  accessToken: string,
): Promise<WorkspaceResponse> => {
  if (USE_MOCK_DATA) {
    // 목업 모드
    console.log('[MOCK] createWorkspace 호출:', data);
    const newWorkspace: WorkspaceResponse = {
      id: `workspace-${Date.now()}`,
      name: data.name,
      description: '',
      ownerId: 'user-123',
      ownerName: '김개발',
      ownerEmail: 'dev.kim@example.com',
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    };
    return new Promise((resolve) => {
      setTimeout(() => resolve(newWorkspace), 300);
    });
  }

  // 실제 API 호출
  const response: AxiosResponse<WorkspaceResponse> = await userRepoClient.post(
    '/api/workspaces',
    data,
    {
      headers: { Authorization: `Bearer ${accessToken}` },
    },
  );
  return response.data;
};

/**
 * 기본 프로필 조회 (workspaceId = null)
 *
 * [백엔드 API]
 * - GET /api/profiles/me
 * - Headers: Authorization: Bearer {accessToken}
 * - Response: UserProfileResponse (workspaceId = null)
 */
export const getMyProfile = async (accessToken: string): Promise<UserProfileResponse> => {
  if (USE_MOCK_DATA) {
    // 목업 모드
    console.log('[MOCK] getMyProfile 호출 - 기본 프로필');
    return new Promise((resolve) => {
      setTimeout(() => resolve(MOCK_DEFAULT_PROFILE), 300);
    });
  }

  // 실제 API 호출
  const response: AxiosResponse<UserProfileResponse> = await userRepoClient.get('/api/profiles/me', {
    headers: { Authorization: `Bearer ${accessToken}` },
  });
  return response.data;
};

/**
 * 기본 프로필 업데이트 (workspaceId = null)
 *
 * [백엔드 API]
 * - PUT /api/profiles/me
 * - Headers: Authorization: Bearer {accessToken}
 * - Body: UpdateProfileRequest
 * - Response: UserProfileResponse (workspaceId = null)
 */
export const updateMyProfile = async (
  data: UpdateProfileRequest,
  accessToken: string,
): Promise<UserProfileResponse> => {
  if (USE_MOCK_DATA) {
    // 목업 모드
    console.log('[MOCK] updateMyProfile 호출 - 기본 프로필:', data);
    MOCK_DEFAULT_PROFILE = {
      ...MOCK_DEFAULT_PROFILE,
      ...data,
      updatedAt: new Date().toISOString(),
    };
    return new Promise((resolve) => {
      setTimeout(() => resolve(MOCK_DEFAULT_PROFILE), 300);
    });
  }

  // 실제 API 호출
  const response: AxiosResponse<UserProfileResponse> = await userRepoClient.put(
    '/api/profiles/me',
    data,
    {
      headers: { Authorization: `Bearer ${accessToken}` },
    },
  );
  return response.data;
};

/**
 * 워크스페이스별 프로필 조회
 *
 * [백엔드 API]
 * - GET /api/profiles/workspace/{workspaceId}
 * - Headers: Authorization: Bearer {accessToken}
 * - Response: UserProfileResponse (workspaceId = {workspaceId})
 *
 * [비즈니스 로직]
 * - userId + workspaceId 조합으로 프로필 조회
 * - 프로필이 없으면 404 또는 null 반환 (프론트에서 기본값 사용)
 */
export const getWorkspaceProfile = async (
  workspaceId: string,
  accessToken: string,
): Promise<UserProfileResponse | null> => {
  if (USE_MOCK_DATA) {
    // 목업 모드
    console.log('[MOCK] getWorkspaceProfile 호출:', workspaceId);
    const profile = MOCK_WORKSPACE_PROFILES[workspaceId];
    return new Promise((resolve) => {
      setTimeout(() => resolve(profile || null), 300);
    });
  }

  // 실제 API 호출
  try {
    const response: AxiosResponse<UserProfileResponse> = await userRepoClient.get(
      `/api/profiles/workspace/${workspaceId}`,
      {
        headers: { Authorization: `Bearer ${accessToken}` },
      },
    );
    return response.data;
  } catch (error: any) {
    // 404 에러면 프로필이 없는 것으로 처리
    if (error.response?.status === 404) {
      return null;
    }
    throw error;
  }
};

/**
 * 워크스페이스별 프로필 생성/업데이트
 *
 * [백엔드 API]
 * - PUT /api/profiles/workspace/{workspaceId}
 * - Headers: Authorization: Bearer {accessToken}
 * - Body: UpdateProfileRequest
 * - Response: UserProfileResponse (workspaceId = {workspaceId})
 *
 * [비즈니스 로직]
 * - userId + workspaceId 조합으로 프로필 검색
 * - 있으면 UPDATE, 없으면 INSERT (UPSERT)
 */
export const updateWorkspaceProfile = async (
  workspaceId: string,
  data: UpdateProfileRequest,
  accessToken: string,
): Promise<UserProfileResponse> => {
  if (USE_MOCK_DATA) {
    // 목업 모드
    console.log('[MOCK] updateWorkspaceProfile 호출:', workspaceId, data);
    const existingProfile = MOCK_WORKSPACE_PROFILES[workspaceId];
    const updatedProfile: UserProfileResponse = {
      profileId: existingProfile?.profileId || `profile-ws-${Date.now()}`,
      userId: 'user-123',
      workspaceId: workspaceId,
      name: data.name || existingProfile?.name || '김개발',
      email: data.email !== undefined ? data.email : existingProfile?.email || null,
      profileImageUrl:
        data.profileImageUrl !== undefined ? data.profileImageUrl : existingProfile?.profileImageUrl || null,
      createdAt: existingProfile?.createdAt || new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    };
    MOCK_WORKSPACE_PROFILES[workspaceId] = updatedProfile;
    return new Promise((resolve) => {
      setTimeout(() => resolve(updatedProfile), 300);
    });
  }

  // 실제 API 호출
  const response: AxiosResponse<UserProfileResponse> = await userRepoClient.put(
    `/api/profiles/workspace/${workspaceId}`,
    data,
    {
      headers: { Authorization: `Bearer ${accessToken}` },
    },
  );
  return response.data;
};

// 필요한 경우, getAuthInfo 등 인증 관련 API도 여기에 추가할 수 있습니다.
