// src/api/user/userService.ts

import {
  CreateWorkspaceRequest,
  UpdateProfileRequest,
  UpdateWorkspaceSettingsRequest,
  UserProfileResponse,
  WorkspaceResponse,
  WorkspaceMemberResponse,
  WorkspaceMemberRole, // DTO 타입에서 역할 타입을 가져옴
  WorkspaceSettingsResponse,
  JoinRequestResponse,
  InviteUserRequest,
  UserWorkspaceResponse,
  SaveAttachmentRequest,
  UpdateProfileImageByKeyRequest,
  PresignedUrlRequest,
  PresignedUrlResponse,
  AttachmentResponse,
  // UpdateWorkspaceRequest DTO가 명시되지 않아 임시로 구조를 정의함
} from '../../types/user';
import { userRepoClient } from '../apiConfig';
import { AxiosResponse } from 'axios';

// ========================================
// Workspace API Functions (워크스페이스 전체 관리)
// ========================================

/**
 * 워크스페이스 목록 조회 (현재 사용자가 속한 모든 워크스페이스)
 * [API] GET /api/workspaces/all
 */
export const getMyWorkspaces = async (): Promise<UserWorkspaceResponse[]> => {
  const response: AxiosResponse<UserWorkspaceResponse[]> = await userRepoClient.get(
    '/api/workspaces/all',
  );
  console.log(response.data);
  return response.data;
};

/**
 * 퍼블릭 워크스페이스 목록 조회 (GET /api/workspaces/public/{workspaceName})
 * (참고: API 설명에 따르면 '퍼블릭 워크스페이스 검색'이 목적이므로,
 * path parameter인 workspaceName을 검색어로 사용하여 필터링하는 것으로 해석됩니다.)
 *
 * @param workspaceName 검색/필터링할 워크스페이스 이름
 * @returns WorkspaceResponse 배열을 담은 Promise
 */
export const getPublicWorkspaces = async (workspaceName: string): Promise<WorkspaceResponse[]> => {
  const response: AxiosResponse<WorkspaceResponse[]> = await userRepoClient.get(
    `/api/workspaces/public/${workspaceName}`,
  );
  return response.data;
};

/**
 * 특정 워크스페이스 조회
 * [API] GET /api/workspaces/{workspaceId}
 * * Response: { data: WorkspaceResponse }
 */
export const getWorkspace = async (workspaceId: string): Promise<WorkspaceResponse> => {
  const response: AxiosResponse<{ data: WorkspaceResponse }> = await userRepoClient.get(
    `/api/workspaces/${workspaceId}`,
  );
  return response.data.data; // data 필드 추출
};

/**
 * 워크스페이스 생성
 * [API] POST /api/workspaces/create
 * * Response: WorkspaceResponse (API 스펙에 따라 { data: WorkspaceResponse }가 아닐 수 있음)
 */
export const createWorkspace = async (data: CreateWorkspaceRequest): Promise<WorkspaceResponse> => {
  const response: AxiosResponse<WorkspaceResponse> = await userRepoClient.post(
    '/api/workspaces/create',
    data,
  );
  return response.data;
};

/**
 * 워크스페이스 수정
 * [API] PUT /api/workspaces/{workspaceId}
 * * Response: { data: WorkspaceResponse }
 */
export const updateWorkspace = async (
  workspaceId: string,
  data: { workspaceName?: string; workspaceDescription?: string },
): Promise<WorkspaceResponse> => {
  const response: AxiosResponse<{ data: WorkspaceResponse }> = await userRepoClient.put(
    `/api/workspaces/${workspaceId}`,
    data,
  );
  return response.data.data; // data 필드 추출
};

/**
 * 워크스페이스 삭제 (소프트 삭제)
 * [API] DELETE /api/workspaces/{workspaceId}
 */
export const deleteWorkspace = async (workspaceId: string): Promise<void> => {
  await userRepoClient.delete(`/api/workspaces/${workspaceId}`);
};

/**
 * 워크스페이스 설정 조회
 * [API] GET /api/workspaces/{workspaceId}/settings
 */
export const getWorkspaceSettings = async (
  workspaceId: string,
): Promise<WorkspaceSettingsResponse> => {
  const response: AxiosResponse<WorkspaceSettingsResponse> = await userRepoClient.get(
    `/api/workspaces/${workspaceId}/settings`,
  );
  return response.data;
};

/**
 * 워크스페이스 설정 수정
 * [API] PUT /api/workspaces/{workspaceId}/settings
 */
export const updateWorkspaceSettings = async (
  workspaceId: string,
  data: UpdateWorkspaceSettingsRequest,
): Promise<WorkspaceSettingsResponse> => {
  const response: AxiosResponse<WorkspaceSettingsResponse> = await userRepoClient.put(
    `/api/workspaces/${workspaceId}/settings`,
    data,
  );
  return response.data;
};

// ========================================
// Member & Join Request API Functions (회원/가입 요청 관리)
// * WorkspaceMembersTab.tsx 에서 사용되는 주요 함수 그룹
// ========================================

/**
 * 워크스페이스 회원 목록 조회
 * [API] GET /api/workspaces/{workspaceId}/members
 */
export const getWorkspaceMembers = async (
  workspaceId: string,
): Promise<WorkspaceMemberResponse[]> => {
  const response: AxiosResponse<WorkspaceMemberResponse[]> = await userRepoClient.get(
    `/api/workspaces/${workspaceId}/members`,
  );
  return response.data;
};

/**
 * 승인 대기 회원 목록 조회
 * [API] GET /api/workspaces/{workspaceId}/pendingMembers
 * * Response: { data: JoinRequestResponse[] }
 */
export const getPendingMembers = async (workspaceId: string): Promise<JoinRequestResponse[]> => {
  const response: AxiosResponse<JoinRequestResponse[]> = await userRepoClient.get(
    `/api/workspaces/${workspaceId}/pendingMembers`,
  );
  return response.data; // ✅ response.data.data가 아님!
};

/**
 * 가입 신청 승인
 * [API] POST /api/workspaces/{workspaceId}/members/{userId}/approve
 */
export const approveMember = async (workspaceId: string, userId: string): Promise<void> => {
  await userRepoClient.post(`/api/workspaces/${workspaceId}/members/${userId}/approve`, {});
};

/**
 * 가입 신청 거절
 * [API] POST /api/workspaces/{workspaceId}/members/{userId}/reject
 */
export const rejectMember = async (workspaceId: string, userId: string): Promise<void> => {
  await userRepoClient.post(`/api/workspaces/${workspaceId}/members/${userId}/reject`, {});
};

/**
 * 워크스페이스에 사용자 초대 (userId 기준)
 * [API] POST /api/workspaces/{workspaceId}/members/invite
 * * Response: { data: WorkspaceMemberResponse }
 */
export const inviteUser = async (
  workspaceId: string,
  query: string,
): Promise<WorkspaceMemberResponse> => {
  const data: InviteUserRequest = { query };

  const response: AxiosResponse<{ data: WorkspaceMemberResponse }> = await userRepoClient.post(
    `/api/workspaces/${workspaceId}/members/invite`,
    data,
  );
  return response.data.data; // data 필드 추출
};

/**
 * 멤버 역할 변경
 * [API] PUT /api/workspaces/{workspaceId}/members/{memberId}/role
 * * Response: { data: WorkspaceMemberResponse }
 */
export const updateMemberRole = async (
  workspaceId: string,
  memberId: string,
  roleName: WorkspaceMemberRole, // DTO 타입을 사용
): Promise<WorkspaceMemberResponse> => {
  const data = { roleName };

  const response: AxiosResponse<{ data: WorkspaceMemberResponse }> = await userRepoClient.put(
    `/api/workspaces/${workspaceId}/members/${memberId}/role`,
    data,
  );
  return response.data.data; // data 필드 추출
};

/**
 * 멤버 제거
 * [API] DELETE /api/workspaces/{workspaceId}/members/{memberId}
 */
export const removeMember = async (workspaceId: string, memberId: string): Promise<void> => {
  await userRepoClient.delete(`/api/workspaces/${workspaceId}/members/${memberId}`);
};

/**
 * 가입 신청 목록 조회 (status 필터 가능)
 * [API] GET /api/workspaces/{workspaceId}/joinRequests
 * * Response: { data: JoinRequestResponse[] }
 */
export const getJoinRequests = async (
  workspaceId: string,
  status?: string, // 'PENDING', 'APPROVED', 'REJECTED'
): Promise<JoinRequestResponse[]> => {
  const response: AxiosResponse<{ data: JoinRequestResponse[] }> = await userRepoClient.get(
    `/api/workspaces/${workspaceId}/joinRequests`,
    { params: { status } },
  );
  return response.data.data; // data 필드 추출
};

/**
 * 워크스페이스 가입 신청
 * [API] POST /api/workspaces/join-requests
 * * Response: { data: JoinRequestResponse }
 */
export const createJoinRequest = async (workspaceId: string): Promise<JoinRequestResponse> => {
  const data = { workspaceId };
  const response: AxiosResponse<JoinRequestResponse> = await userRepoClient.post(
    '/api/workspaces/join-requests',
    data,
  );
  console.log(response.data);
  return response.data; // data 필드 추출
};

// ========================================
// UserProfile API Functions
// ========================================

/**
 * 내 프로필 조회 (기본 프로필)
 * [API] GET /api/profiles/me
 * * Response: { data: UserProfileResponse }
 */
export const getMyProfile = async (): Promise<UserProfileResponse> => {
  const response: AxiosResponse<{ data: UserProfileResponse }> = await userRepoClient.get(
    '/api/profiles/me',
  );
  return response.data.data; // data 필드 추출
};

/**
 * 내 모든 프로필 조회 (기본 프로필 + 워크스페이스별 프로필)
 * [API] GET /api/profiles/all/me
 * * Response: { data: UserProfileResponse[] }
 */
export const getAllMyProfiles = async (): Promise<UserProfileResponse[]> => {
  const response: AxiosResponse<UserProfileResponse[]> = await userRepoClient.get(
    '/api/profiles/all/me',
  );
  console.log(response.data);
  return response.data; // data 필드 추출
};

/**
 * 내 프로필 정보 통합 업데이트 (기본 프로필)
 * [API] PUT /api/profiles/me
 * * Response: { data: UserProfileResponse }
 */
export const updateMyProfile = async (data: UpdateProfileRequest): Promise<UserProfileResponse> => {
  const response: AxiosResponse<{ data: UserProfileResponse }> = await userRepoClient.put(
    '/api/profiles/me',
    data,
  );
  return response.data.data; // data 필드 추출
};

// ========================================
// New API Functions (기타)
// ========================================

/**
 * 기본 워크스페이스 설정
 * [API] POST /api/workspaces/default
 */
export const setDefaultWorkspace = async (workspaceId: string): Promise<void> => {
  const data = { workspaceId };
  await userRepoClient.post('/api/workspaces/default', data);
};

// ========================================
// [제거/대체됨] 워크스페이스 프로필 관리 함수 (호환성 유지용)
// ========================================

/**
 * [제거됨] 워크스페이스 프로필 조회 (GET /api/profiles/workspace/{workspaceId})
 * @deprecated 프론트엔드에서 `getAllMyProfiles()`를 호출하여 필터링해야 합니다.
 */
// export const getWorkspaceProfile = async (
//   workspaceId: string,
// ): Promise<UserProfileResponse | null> => {
//   return null;
// };

// /**
//  * [제거됨] 워크스페이스 프로필 생성/수정 (PUT /api/profiles/workspace/{workspaceId})
//  */
// export const updateWorkspaceProfile = async (
//   workspaceId: string,
//   data: UpdateProfileRequest,
// ): Promise<UserProfileResponse> => {
//   throw new Error('워크스페이스별 프로필 업데이트 엔드포인트가 제거되었습니다. (백엔드 구현 필요)');
// };
// ========================================
// Profile Image Upload API Functions (추가 필요)
// ========================================

/**
 * 프로필 이미지 업로드를 위한 Presigned URL 생성
 * [API] POST /api/profiles/me/image/presigned-url
 */
export const generateProfilePresignedUrl = async (
  data: PresignedUrlRequest,
): Promise<PresignedUrlResponse> => {
  const response: AxiosResponse<PresignedUrlResponse> = await userRepoClient.post(
    '/api/profiles/me/image/presigned-url',
    data,
  );
  return response.data;
};

/**
 * 프로필 이미지 첨부파일 메타데이터 저장
 * [API] POST /api/profiles/me/image/attachment
 */
export const saveProfileAttachmentMetadata = async (
  data: SaveAttachmentRequest,
): Promise<AttachmentResponse> => {
  const response: AxiosResponse<AttachmentResponse> = await userRepoClient.post(
    '/api/profiles/me/image/attachment',
    data,
  );
  return response.data;
};

/**
 * 프로필 이미지 업데이트 (fileKey 기반)
 * [API] PUT /api/profiles/me/image
 */
export const updateProfileImageByKey = async (
  data: UpdateProfileImageByKeyRequest,
): Promise<UserProfileResponse> => {
  const response: AxiosResponse<{ data: UserProfileResponse }> = await userRepoClient.put(
    '/api/profiles/me/image',
    data,
  );
  return response.data.data;
};

/**
 * 프로필 이미지 업로드 헬퍼 함수 (전체 플로우)
 * 1. Presigned URL 생성
 * 2. S3에 직접 업로드
 * 3. 메타데이터 저장
 * 4. 프로필 업데이트
 */
export const uploadProfileImage = async (
  file: File,
  workspaceId: string,
): Promise<UserProfileResponse> => {
  try {
    // 1. Presigned URL 요청
    const presignedData: PresignedUrlRequest = {
      workspaceId,
      fileName: file.name,
      fileSize: file.size,
      contentType: file.type,
    };

    const { uploadUrl, fileKey } = await generateProfilePresignedUrl(presignedData);

    // 2. S3에 파일 업로드
    await fetch(uploadUrl, {
      method: 'PUT',
      headers: {
        'Content-Type': file.type,
      },
      body: file,
    });

    // 3. 첨부파일 메타데이터 저장
    const attachmentData: SaveAttachmentRequest = {
      fileKey,
      fileName: file.name,
      fileSize: file.size,
      contentType: file.type,
    };

    await saveProfileAttachmentMetadata(attachmentData);

    // 4. 프로필 이미지 업데이트
    const updateData: UpdateProfileImageByKeyRequest = {
      workspaceId,
      fileKey,
    };

    return await updateProfileImageByKey(updateData);
  } catch (error) {
    console.error('프로필 이미지 업로드 실패:', error);
    throw error;
  }
};
