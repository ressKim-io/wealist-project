// --- DTO Interfaces ---

export interface AuthResponse {
  accessToken: string;
  refreshToken: string;
  userId: string; // (format: uuid)
  name: string; // Google OAuth에서 받은 사용자 이름 (UserProfile.nickName 값)
  email: string;
  tokenType: string; // e.g., "bearer"
}

export interface WorkspaceResponse {
  workspaceId: string;
  workspaceName: string;
  workspaceDescription: string;
  ownerId: string;
  ownerName: string;
  ownerEmail: string;
  createdAt: string;
  updatedAt: string;
}

export interface CreateWorkspaceRequest {
  workspaceName: string;
  workspaceDescription?: string;
}

export interface UserProfileResponse {
  profileId: string;
  userId: string;
  workspaceId?: string | null; // [추가] 워크스페이스별 프로필용
  nickName: string;
  email: string | null;
  profileImageUrl: string | null;
  createdAt: string;
  updatedAt: string;
}

export interface UpdateProfileRequest {
  nickName?: string;
  email?: string;
  profileImageUrl?: string;
}

// --- Workspace Management Interfaces ---

export type WorkspaceMemberRole = 'OWNER' | 'ADMIN' | 'MEMBER';

export interface WorkspaceMember {
  userId: string;
  userName: string; // Changed from 'name' to match backend DTO
  userEmail: string; // Changed from 'email' to match backend DTO
  roleName: WorkspaceMemberRole; // Changed from 'role' to match backend DTO
  profileImageUrl?: string | null;
  joinedAt: string;
}

export interface PendingMember {
  userId: string;
  nickName: string;
  email: string;
  requestedAt: string;
}

export interface InvitableUser {
  userId: string;
  nickName: string;
  email: string;
}

export interface WorkspaceSettings {
  workspaceId: string;
  workspaceName: string;
  workspaceDescription: string;
  isPublic: boolean; // 공개/비공개
  requiresApproval: boolean; // 승인제/비승인제
  onlyOwnerCanInvite: boolean; // OWNER만 초대 가능
}

export interface UpdateWorkspaceSettingsRequest {
  workspaceName?: string;
  workspaceDescription?: string;
  isPublic?: boolean;
  requiresApproval?: boolean;
  onlyOwnerCanInvite?: boolean;
}

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
export interface UpdateMemberRoleRequest {
  roleName: 'ADMIN' | 'MEMBER';
}

// 멤버 초대 요청 DTO (기능 요구사항에 따라 POST 요청을 가정)
export interface InviteMemberRequest {
  email: string;
  roleName: 'ADMIN' | 'MEMBER';
}
