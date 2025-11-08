import { userRepoClient } from '../apiConfig';
import { AxiosResponse } from 'axios';

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

// --- API Service Functions ---

/**
 * 워크스페이스 목록 조회 (GET /api/workspaces)
 */
export const getWorkspaces = async (accessToken: string): Promise<WorkspaceResponse[]> => {
  const response: AxiosResponse<WorkspaceResponse[]> = await userRepoClient.get('/api/workspaces', {
    headers: { Authorization: `Bearer ${accessToken}` },
  });
  return response.data;
};

/**
 * 워크스페이스 생성 (POST /api/workspaces)
 */
export const createWorkspace = async (
  data: CreateWorkspaceRequest,
  accessToken: string,
): Promise<WorkspaceResponse> => {
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
 * 내 프로필 조회 (GET /api/profiles/me)
 */
export const getMyProfile = async (accessToken: string): Promise<UserProfileResponse> => {
  const response: AxiosResponse<UserProfileResponse> = await userRepoClient.get('/api/profiles/me', {
    headers: { Authorization: `Bearer ${accessToken}` },
  });
  return response.data;
};

/**
 * 내 프로필 업데이트 (PUT /api/profiles/me)
 */
export const updateMyProfile = async (
  data: UpdateProfileRequest,
  accessToken: string,
): Promise<UserProfileResponse> => {
  const response: AxiosResponse<UserProfileResponse> = await userRepoClient.put(
    '/api/profiles/me',
    data,
    {
      headers: { Authorization: `Bearer ${accessToken}` },
    },
  );
  return response.data;
};

// 필요한 경우, getAuthInfo 등 인증 관련 API도 여기에 추가할 수 있습니다.
