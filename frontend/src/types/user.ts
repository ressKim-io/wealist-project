// --- 1. 인증/사용자 기본 DTO ---

/**
 * @summary 토큰 갱신 응답 DTO (AuthResponse)
 * [API: POST /api/auth/refresh]
 */
export interface AuthResponse {
  accessToken: string;
  refreshToken: string;
  userId: string; // (format: uuid)
  name: string;
  email: string;
  tokenType: string;
}

/**
 * @summary 기본 프로필 조회/수정 응답 DTO (UserProfileResponse)
 * [API: GET/PUT /api/profiles/me]
 */
export interface UserProfileResponse {
  profileId: string;
  userId: string;
  workspaceId?: string | null; // null이면 기본 프로필
  nickName: string;
  email: string | null;
  profileImageUrl: string | null;
  // createdAt과 updatedAt은 명세 DTO에는 없으나, 기존 정의와 일관성을 위해 유지
  createdAt?: string;
  updatedAt?: string;
}

/**
 * @summary 프로필 정보 통합 업데이트 요청 DTO (UpdateProfileRequest)
 * [API: PUT /api/profiles/me]
 */
export interface UpdateProfileRequest {
  workspaceId: string; // ✅ 필수로 변경 (Swagger 명세 기준)
  userId: string; // ✅ 필수로 변경 (Swagger 명세 기준)
  nickName?: string;
  email?: string;
  profileImageUrl?: string;
}

// --- 2. 워크스페이스 DTO ---

/**
 * @summary 워크스페이스 조회/생성 응답 DTO (UserWorkspaceResponse)
 * [API: GET /api/workspaces/all]
 */
export interface UserWorkspaceResponse {
  workspaceId: string;
  workspaceName: string;
  workspaceDescription: string;
  owner: boolean;
  role: string;
  createdAt: string;
}

/**
 * @summary 워크스페이스 생성 응답 DTO (WorkspaceResponse)
 * [ POST /api/workspaces]
 */
export interface WorkspaceResponse {
  workspaceId: string;
  workspaceName: string;
  workspaceDescription: string;
  ownerId: string;
  ownerName: string;
  ownerEmail: string;
  isPublic: boolean; // OpenAPI 명세에 추가된 필드
  needApproved: boolean; // OpenAPI 명세에 추가된 필드 (requiresApproval과 동일 목적)
  createdAt: string;
  // OpenAPI 명세 DTO에는 없으나, 기존 정의의 일관성을 위해 유지
  updatedAt?: string;
}

/**
 * @summary 워크스페이스 생성 요청 DTO (CreateWorkspaceRequest)
 * [API: POST /api/workspaces]
 */
export interface CreateWorkspaceRequest {
  workspaceName: string;
  workspaceDescription?: string;
  isPublic?: boolean; // OpenAPI 명세 DTO에 추가됨
}

// --- 3. 워크스페이스 설정 DTO ---

/**
 * @summary 워크스페이스 설정 조회 응답 DTO (WorkspaceSettingsResponse)
 * [API: GET /api/workspaces/{workspaceId}/settings]
 */
export interface WorkspaceSettingsResponse {
  workspaceId: string;
  workspaceName: string;
  workspaceDescription: string;
  isPublic: boolean;
  requiresApproval: boolean; // DTO 명세: requiresApproval
  onlyOwnerCanInvite: boolean;
}

/**
 * @summary 워크스페이스 설정 수정 요청 DTO (UpdateWorkspaceSettingsRequest)
 * [API: PUT /api/workspaces/{workspaceId}/settings]
 */
export interface UpdateWorkspaceSettingsRequest {
  workspaceName?: string;
  workspaceDescription?: string;
  isPublic?: boolean;
  requiresApproval?: boolean; // DTO 명세: requiresApproval
  onlyOwnerCanInvite?: boolean;
}

// 이전 WorkspaceSettings 인터페이스는 WorkspaceSettingsResponse로 대체됩니다.
// export interface WorkspaceSettings { ... } // 제거됨

// --- 4. 멤버/가입 관리 DTO ---

export type WorkspaceMemberRole = 'OWNER' | 'ADMIN' | 'MEMBER';

/**
 * @summary 워크스페이스 멤버 응답 DTO (WorkspaceMemberResponse)
 * [API: GET /api/workspaces/{workspaceId}/members, PUT /api/workspaces/{id}/role]
 * @description 워크스페이스 멤버 목록 조회 및 역할 변경 응답
 */
export interface WorkspaceMemberResponse {
  id: string; // WorkspaceMember ID (format: uuid)
  workspaceId: string;
  userId: string;
  profileImageUrl?: string;
  userName: string;
  userEmail: string;
  roleName: WorkspaceMemberRole;
  isDefault: boolean;
  joinedAt: string;
  role: string;
}

/**
 * @summary 멤버 역할 변경 요청 DTO (UpdateMemberRoleRequest)
 * [API: PUT /api/workspaces/{workspaceId}/members/{memberId}/role]
 */
export interface UpdateMemberRoleRequest {
  roleName: 'ADMIN' | 'MEMBER'; // OWNER는 경로 변수에서 판단될 가능성이 높으므로 ADMIN/MEMBER만 남김
}

/**
 * @summary 가입/초대 요청 응답 DTO (JoinRequestResponse)
 * [API: GET /api/workspaces/{id}/pendingMembers]
 * @description 승인 대기 목록 조회 응답 (이전 PendingMember 대체)
 */
export interface JoinRequestResponse {
  id: string; // JoinRequest ID (format: uuid)
  workspaceId: string;
  userId: string;
  userName: string;
  userEmail: string;
  status: string; // e.g., "PENDING"
  requestedAt: string;
  updatedAt: string;
}

/**
 * @summary 워크스페이스 가입 신청 요청 DTO (CreateJoinRequestRequest)
 * [API: POST /api/workspaces/join-requests]
 */
export interface CreateJoinRequestRequest {
  workspaceId: string;
}

/**
 * @summary 워크스페이스 멤버 초대 요청 DTO (InviteUserRequest)
 * [API: POST /api/workspaces/{workspaceId}/members/invite]
 */
export interface InviteUserRequest {
  query: string;
}

/**
 * @summary 기본 워크스페이스 설정 요청 DTO (SetDefaultWorkspaceRequest)
 * [API: POST /api/workspaces/default]
 */
export interface SetDefaultWorkspaceRequest {
  workspaceId: string;
}

// --- 5. 제거된 불필요/구 버전 타입 ---

// // WorkspaceMember (구 버전): WorkspaceMemberResponse로 대체됨
// // PendingMember (구 버전): JoinRequestResponse로 대체됨
// // InvitableUser (명세에서 검색 API가 사라짐): 제거함
// // InviteMemberRequest (구 버전): InviteUserRequest로 대체됨
// --- 6. 프로필 이미지 업로드 관련 DTO (추가 필요) ---

/**
 * @summary Presigned URL 생성 요청 DTO
 * [API: POST /api/profiles/me/image/presigned-url]
 */
export interface PresignedUrlRequest {
  workspaceId: string; // format: uuid
  fileName: string;
  fileSize: number; // bytes
  contentType: string; // e.g., "image/jpeg"
}

/**
 * @summary Presigned URL 생성 응답 DTO
 * [API: POST /api/profiles/me/image/presigned-url]
 */
export interface PresignedUrlResponse {
  uploadUrl: string;
  fileKey: string;
  expiresIn: number; // 초 단위
}

/**
 * @summary 첨부파일 메타데이터 저장 요청 DTO
 * [API: POST /api/profiles/me/image/attachment]
 */
export interface SaveAttachmentRequest {
  fileKey: string;
  fileName: string;
  fileSize: number;
  contentType: string;
}

/**
 * @summary 첨부파일 메타데이터 저장 응답 DTO
 * [API: POST /api/profiles/me/image/attachment]
 */
export interface AttachmentResponse {
  id: string; // format: uuid
  entityType: string;
  entityId: string | null; // format: uuid
  status: string;
  fileName: string;
  fileUrl: string;
  fileSize: number;
  contentType: string;
  uploadedBy: string; // format: uuid
  uploadedAt: string; // format: date-time
  expiresAt: string | null; // format: date-time
}

/**
 * @summary 프로필 이미지 업데이트 요청 DTO (fileKey 기반)
 * [API: PUT /api/profiles/me/image]
 */
export interface UpdateProfileImageByKeyRequest {
  workspaceId: string; // format: uuid
  fileKey: string;
}
