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

// --- Workspace Management Interfaces ---

export type WorkspaceMemberRole = 'OWNER' | 'ADMIN' | 'MEMBER';

export interface WorkspaceMember {
  userId: string;
  name: string;
  email: string;
  role: WorkspaceMemberRole;
  profileImageUrl?: string | null;
  joinedAt: string;
}

export interface PendingMember {
  userId: string;
  name: string;
  email: string;
  requestedAt: string;
}

export interface InvitableUser {
  userId: string;
  name: string;
  email: string;
}

export interface WorkspaceSettings {
  workspaceId: string;
  name: string;
  description: string;
  isPublic: boolean; // 공개/비공개
  requiresApproval: boolean; // 승인제/비승인제
  onlyOwnerCanInvite: boolean; // OWNER만 초대 가능
}

export interface UpdateWorkspaceSettingsRequest {
  name?: string;
  description?: string;
  isPublic?: boolean;
  requiresApproval?: boolean;
  onlyOwnerCanInvite?: boolean;
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

// 목업: 워크스페이스 설정
let MOCK_WORKSPACE_SETTINGS: Record<string, WorkspaceSettings> = {
  'workspace-1': {
    workspaceId: 'workspace-1',
    name: '오렌지클라우드',
    description: '메인 워크스페이스',
    isPublic: true,
    requiresApproval: true,
    onlyOwnerCanInvite: false,
  },
};

// 목업: 워크스페이스 회원 목록
let MOCK_WORKSPACE_MEMBERS: Record<string, WorkspaceMember[]> = {
  'workspace-1': [
    {
      userId: 'user-123',
      name: '김개발',
      email: 'dev.kim@orangecloud.com',
      role: 'OWNER',
      profileImageUrl: 'https://i.pravatar.cc/150?img=12',
      joinedAt: '2024-01-01T00:00:00Z',
    },
    {
      userId: 'user-456',
      name: '이디자인',
      email: 'design.lee@orangecloud.com',
      role: 'ADMIN',
      profileImageUrl: 'https://i.pravatar.cc/150?img=5',
      joinedAt: '2024-01-05T00:00:00Z',
    },
    {
      userId: 'user-789',
      name: '박프론트',
      email: 'frontend.park@orangecloud.com',
      role: 'MEMBER',
      profileImageUrl: null,
      joinedAt: '2024-01-10T00:00:00Z',
    },
    {
      userId: 'user-101',
      name: '정백엔드',
      email: 'backend.jung@orangecloud.com',
      role: 'MEMBER',
      profileImageUrl: 'https://i.pravatar.cc/150?img=33',
      joinedAt: '2024-01-12T00:00:00Z',
    },
    {
      userId: 'user-202',
      name: '최데브옵스',
      email: 'devops.choi@orangecloud.com',
      role: 'MEMBER',
      profileImageUrl: null,
      joinedAt: '2024-01-15T00:00:00Z',
    },
  ],
};

// 목업: 승인 대기 회원
let MOCK_PENDING_MEMBERS: Record<string, PendingMember[]> = {
  'workspace-1': [
    {
      userId: 'user-pending-1',
      name: '한신입',
      email: 'newbie.han@gmail.com',
      requestedAt: '2024-01-20T09:30:00Z',
    },
    {
      userId: 'user-pending-2',
      name: '강인턴',
      email: 'intern.kang@naver.com',
      requestedAt: '2024-01-21T14:20:00Z',
    },
  ],
};

// 목업: 초대 가능 회원 (전체 사용자 목록에서 현재 워크스페이스에 없는 사람들)
const MOCK_INVITABLE_USERS: InvitableUser[] = [
  {
    userId: 'user-inv-1',
    name: '정마케팅',
    email: 'marketing.jung@example.com',
  },
  {
    userId: 'user-inv-2',
    name: '송영업',
    email: 'sales.song@example.com',
  },
  {
    userId: 'user-inv-3',
    name: '한재무',
    email: 'finance.han@example.com',
  },
];

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

// ========================================
// Workspace Management API Functions
// ========================================

/**
 * 워크스페이스 설정 조회
 *
 * [백엔드 API]
 * - GET /api/workspaces/{workspaceId}/settings
 * - Headers: Authorization: Bearer {accessToken}
 * - Response: WorkspaceSettings
 */
export const getWorkspaceSettings = async (
  workspaceId: string,
  accessToken: string,
): Promise<WorkspaceSettings> => {
  if (USE_MOCK_DATA) {
    console.log('[MOCK] getWorkspaceSettings 호출:', workspaceId);
    const settings = MOCK_WORKSPACE_SETTINGS[workspaceId] || {
      workspaceId,
      name: '워크스페이스',
      description: '',
      isPublic: false,
      requiresApproval: false,
      onlyOwnerCanInvite: true,
    };
    return new Promise((resolve) => {
      setTimeout(() => resolve(settings), 300);
    });
  }

  const response: AxiosResponse<WorkspaceSettings> = await userRepoClient.get(
    `/api/workspaces/${workspaceId}/settings`,
    {
      headers: { Authorization: `Bearer ${accessToken}` },
    },
  );
  return response.data;
};

/**
 * 워크스페이스 설정 업데이트
 *
 * [백엔드 API]
 * - PUT /api/workspaces/{workspaceId}/settings
 * - Headers: Authorization: Bearer {accessToken}
 * - Body: UpdateWorkspaceSettingsRequest
 * - Response: WorkspaceSettings
 */
export const updateWorkspaceSettings = async (
  workspaceId: string,
  data: UpdateWorkspaceSettingsRequest,
  accessToken: string,
): Promise<WorkspaceSettings> => {
  if (USE_MOCK_DATA) {
    console.log('[MOCK] updateWorkspaceSettings 호출:', workspaceId, data);
    const current = MOCK_WORKSPACE_SETTINGS[workspaceId];
    const updated = {
      ...current,
      ...data,
      workspaceId,
    };
    MOCK_WORKSPACE_SETTINGS[workspaceId] = updated;
    return new Promise((resolve) => {
      setTimeout(() => resolve(updated), 300);
    });
  }

  const response: AxiosResponse<WorkspaceSettings> = await userRepoClient.put(
    `/api/workspaces/${workspaceId}/settings`,
    data,
    {
      headers: { Authorization: `Bearer ${accessToken}` },
    },
  );
  return response.data;
};

/**
 * 워크스페이스 회원 목록 조회
 *
 * [백엔드 API]
 * - GET /api/workspaces/{workspaceId}/members
 * - Headers: Authorization: Bearer {accessToken}
 * - Response: WorkspaceMember[]
 */
export const getWorkspaceMembers = async (
  workspaceId: string,
  accessToken: string,
): Promise<WorkspaceMember[]> => {
  if (USE_MOCK_DATA) {
    console.log('[MOCK] getWorkspaceMembers 호출:', workspaceId);
    const members = MOCK_WORKSPACE_MEMBERS[workspaceId] || [];
    return new Promise((resolve) => {
      setTimeout(() => resolve(members), 300);
    });
  }

  const response: AxiosResponse<WorkspaceMember[]> = await userRepoClient.get(
    `/api/workspaces/${workspaceId}/members`,
    {
      headers: { Authorization: `Bearer ${accessToken}` },
    },
  );
  return response.data;
};

/**
 * 승인 대기 회원 목록 조회
 *
 * [백엔드 API]
 * - GET /api/workspaces/{workspaceId}/pending-members
 * - Headers: Authorization: Bearer {accessToken}
 * - Response: PendingMember[]
 */
export const getPendingMembers = async (
  workspaceId: string,
  accessToken: string,
): Promise<PendingMember[]> => {
  if (USE_MOCK_DATA) {
    console.log('[MOCK] getPendingMembers 호출:', workspaceId);
    const pending = MOCK_PENDING_MEMBERS[workspaceId] || [];
    return new Promise((resolve) => {
      setTimeout(() => resolve(pending), 300);
    });
  }

  const response: AxiosResponse<PendingMember[]> = await userRepoClient.get(
    `/api/workspaces/${workspaceId}/pending-members`,
    {
      headers: { Authorization: `Bearer ${accessToken}` },
    },
  );
  return response.data;
};

/**
 * 회원 승인
 *
 * [백엔드 API]
 * - POST /api/workspaces/{workspaceId}/members/{userId}/approve
 * - Headers: Authorization: Bearer {accessToken}
 * - Response: void
 */
export const approveMember = async (
  workspaceId: string,
  userId: string,
  accessToken: string,
): Promise<void> => {
  if (USE_MOCK_DATA) {
    console.log('[MOCK] approveMember 호출:', workspaceId, userId);
    // 승인 대기 목록에서 제거
    const pending = MOCK_PENDING_MEMBERS[workspaceId] || [];
    const member = pending.find((m) => m.userId === userId);
    if (member) {
      MOCK_PENDING_MEMBERS[workspaceId] = pending.filter((m) => m.userId !== userId);
      // 회원 목록에 추가
      const members = MOCK_WORKSPACE_MEMBERS[workspaceId] || [];
      members.push({
        userId: member.userId,
        name: member.name,
        email: member.email,
        role: 'MEMBER',
        joinedAt: new Date().toISOString(),
      });
      MOCK_WORKSPACE_MEMBERS[workspaceId] = members;
    }
    return new Promise((resolve) => {
      setTimeout(() => resolve(), 300);
    });
  }

  await userRepoClient.post(
    `/api/workspaces/${workspaceId}/members/${userId}/approve`,
    {},
    {
      headers: { Authorization: `Bearer ${accessToken}` },
    },
  );
};

/**
 * 회원 거절
 *
 * [백엔드 API]
 * - POST /api/workspaces/{workspaceId}/members/{userId}/reject
 * - Headers: Authorization: Bearer {accessToken}
 * - Response: void
 */
export const rejectMember = async (
  workspaceId: string,
  userId: string,
  accessToken: string,
): Promise<void> => {
  if (USE_MOCK_DATA) {
    console.log('[MOCK] rejectMember 호출:', workspaceId, userId);
    const pending = MOCK_PENDING_MEMBERS[workspaceId] || [];
    MOCK_PENDING_MEMBERS[workspaceId] = pending.filter((m) => m.userId !== userId);
    return new Promise((resolve) => {
      setTimeout(() => resolve(), 300);
    });
  }

  await userRepoClient.post(
    `/api/workspaces/${workspaceId}/members/${userId}/reject`,
    {},
    {
      headers: { Authorization: `Bearer ${accessToken}` },
    },
  );
};

/**
 * 회원 역할 변경
 *
 * [백엔드 API]
 * - PUT /api/workspaces/{workspaceId}/members/{userId}/role
 * - Headers: Authorization: Bearer {accessToken}
 * - Body: { role: WorkspaceMemberRole }
 * - Response: void
 */
export const updateMemberRole = async (
  workspaceId: string,
  userId: string,
  role: WorkspaceMemberRole,
  accessToken: string,
): Promise<void> => {
  if (USE_MOCK_DATA) {
    console.log('[MOCK] updateMemberRole 호출:', workspaceId, userId, role);
    const members = MOCK_WORKSPACE_MEMBERS[workspaceId] || [];
    const member = members.find((m) => m.userId === userId);
    if (member) {
      member.role = role;
    }
    return new Promise((resolve) => {
      setTimeout(() => resolve(), 300);
    });
  }

  await userRepoClient.put(
    `/api/workspaces/${workspaceId}/members/${userId}/role`,
    { role },
    {
      headers: { Authorization: `Bearer ${accessToken}` },
    },
  );
};

/**
 * 회원 퇴출
 *
 * [백엔드 API]
 * - DELETE /api/workspaces/{workspaceId}/members/{userId}
 * - Headers: Authorization: Bearer {accessToken}
 * - Response: void
 */
export const removeMember = async (
  workspaceId: string,
  userId: string,
  accessToken: string,
): Promise<void> => {
  if (USE_MOCK_DATA) {
    console.log('[MOCK] removeMember 호출:', workspaceId, userId);
    const members = MOCK_WORKSPACE_MEMBERS[workspaceId] || [];
    MOCK_WORKSPACE_MEMBERS[workspaceId] = members.filter((m) => m.userId !== userId);
    return new Promise((resolve) => {
      setTimeout(() => resolve(), 300);
    });
  }

  await userRepoClient.delete(`/api/workspaces/${workspaceId}/members/${userId}`, {
    headers: { Authorization: `Bearer ${accessToken}` },
  });
};

/**
 * 초대 가능 회원 검색
 *
 * [백엔드 API]
 * - GET /api/workspaces/{workspaceId}/invitable-users?query={query}
 * - Headers: Authorization: Bearer {accessToken}
 * - Response: InvitableUser[]
 */
export const searchInvitableUsers = async (
  workspaceId: string,
  query: string,
  accessToken: string,
): Promise<InvitableUser[]> => {
  if (USE_MOCK_DATA) {
    console.log('[MOCK] searchInvitableUsers 호출:', workspaceId, query);
    const filtered = query.trim()
      ? MOCK_INVITABLE_USERS.filter(
          (u) =>
            u.name.toLowerCase().includes(query.toLowerCase()) ||
            u.email.toLowerCase().includes(query.toLowerCase()),
        )
      : MOCK_INVITABLE_USERS;
    return new Promise((resolve) => {
      setTimeout(() => resolve(filtered), 300);
    });
  }

  const response: AxiosResponse<InvitableUser[]> = await userRepoClient.get(
    `/api/workspaces/${workspaceId}/invitable-users`,
    {
      params: { query },
      headers: { Authorization: `Bearer ${accessToken}` },
    },
  );
  return response.data;
};

/**
 * 회원 초대
 *
 * [백엔드 API]
 * - POST /api/workspaces/{workspaceId}/invite/{userId}
 * - Headers: Authorization: Bearer {accessToken}
 * - Response: void
 */
export const inviteUser = async (
  workspaceId: string,
  userId: string,
  accessToken: string,
): Promise<void> => {
  if (USE_MOCK_DATA) {
    console.log('[MOCK] inviteUser 호출:', workspaceId, userId);
    const user = MOCK_INVITABLE_USERS.find((u) => u.userId === userId);
    if (user) {
      // 승인제라면 승인 대기 목록에 추가, 아니면 바로 회원으로 추가
      const settings = MOCK_WORKSPACE_SETTINGS[workspaceId];
      if (settings?.requiresApproval) {
        const pending = MOCK_PENDING_MEMBERS[workspaceId] || [];
        pending.push({
          userId: user.userId,
          name: user.name,
          email: user.email,
          requestedAt: new Date().toISOString(),
        });
        MOCK_PENDING_MEMBERS[workspaceId] = pending;
      } else {
        const members = MOCK_WORKSPACE_MEMBERS[workspaceId] || [];
        members.push({
          userId: user.userId,
          name: user.name,
          email: user.email,
          role: 'MEMBER',
          joinedAt: new Date().toISOString(),
        });
        MOCK_WORKSPACE_MEMBERS[workspaceId] = members;
      }
    }
    return new Promise((resolve) => {
      setTimeout(() => resolve(), 300);
    });
  }

  await userRepoClient.post(
    `/api/workspaces/${workspaceId}/invite/${userId}`,
    {},
    {
      headers: { Authorization: `Bearer ${accessToken}` },
    },
  );
};

// 필요한 경우, getAuthInfo 등 인증 관련 API도 여기에 추가할 수 있습니다.
