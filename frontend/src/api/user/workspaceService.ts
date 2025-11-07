import { userRepoClient, getAuthHeaders } from '../apiConfig';
import { AxiosResponse } from 'axios';

// --- DTO Interfaces ---

export interface WorkspaceMember {
  id: string; // WorkspaceMember ID (not userId)
  workspaceId: string;
  userId: string;
  userName: string;
  userEmail: string;
  roleName: 'OWNER' | 'ADMIN' | 'MEMBER';
  isDefault: boolean;
  joinedAt: string;
}

// 멤버 역할 변경 요청 DTO
interface UpdateMemberRoleRequest {
  roleName: 'ADMIN' | 'MEMBER';
}

// 멤버 초대 요청 DTO (기능 요구사항에 따라 POST 요청을 가정)
interface InviteMemberRequest {
  email: string;
  roleName: 'ADMIN' | 'MEMBER';
}

// 백엔드 역할을 프론트엔드 역할로 매핑하는 함수 (OWNER -> MASTER, ADMIN -> ORGANIZER)
const mapRole = (role: string): WorkspaceMember['roleName'] => {
  if (role === 'OWNER') return 'OWNER';
  if (role === 'ADMIN') return 'ADMIN';
  // API 명세에 MEMBER가 아닌 다른 역할이 있다면 여기에 추가
  return 'MEMBER';
};

// --- API Service Functions ---

/**
 * 특정 워크스페이스의 모든 멤버 목록을 조회합니다.
 * GET /api/workspaces/{workspaceId}/members
 */
export async function getWorkspaceMembers(
  workspaceId: string,
  token: string,
): Promise<WorkspaceMember[]> {
  const response: AxiosResponse<any[]> = await userRepoClient.get(
    `/api/workspaces/${workspaceId}/members`,
    {
      headers: getAuthHeaders(token),
    },
  );

  return response.data.map((member) => ({
    ...member,
    roleName: mapRole(member.roleName), // 역할 매핑 적용
  })) as WorkspaceMember[];
}

/**
 * 워크스페이스 멤버를 이메일로 초대합니다.
 * POST /api/workspaces/{workspaceId}/members (API 명세에는 없지만, 기능을 위해 가정)
 */
export async function inviteMemberByEmail(
  workspaceId: string,
  email: string,
  roleName: 'ADMIN' | 'MEMBER',
  token: string,
): Promise<void> {
  const requestBody: InviteMemberRequest = { email, roleName };

  await userRepoClient.post(`/api/workspaces/${workspaceId}/members`, requestBody, {
    headers: getAuthHeaders(token),
  });
}

/**
 * 워크스페이스 멤버의 역할을 변경합니다.
 * PUT /api/workspaces/{workspaceId}/members/{memberId}/role
 */
export async function updateMemberRole(
  workspaceId: string,
  memberId: string, // WorkspaceMember ID
  newRoleName: 'ADMIN' | 'MEMBER',
  token: string,
): Promise<WorkspaceMember> {
  const requestBody: UpdateMemberRoleRequest = { roleName: newRoleName };

  const response: AxiosResponse<any> = await userRepoClient.put(
    `/api/workspaces/${workspaceId}/members/${memberId}/role`,
    requestBody,
    {
      headers: getAuthHeaders(token),
    },
  );

  const data = response.data;
  return {
    ...data,
    roleName: mapRole(data.roleName), // 역할 매핑 적용
  } as WorkspaceMember;
}

/**
 * 워크스페이스 멤버를 제거합니다.
 * DELETE /api/workspaces/{workspaceId}/members/{memberId}
 */
export async function removeMember(
  workspaceId: string,
  memberId: string, // WorkspaceMember ID
  token: string,
): Promise<void> {
  await userRepoClient.delete(`/api/workspaces/${workspaceId}/members/${memberId}`, {
    headers: getAuthHeaders(token),
  });
}
