// src/api/userService.ts

// 1. axios 인스턴스 (아마도 별도 파일에 있겠지만 예시로 포함)
import axios from 'axios';

const apiClient = axios.create({
  baseURL: 'http://localhost:8080', // API Docs 서버 URL 기준
});

/**
 * 인증 응답 DTO (로그인, 토큰 갱신 시 사용)
 * (GET /api/auth/test, POST /api/auth/refresh)
 */
export interface AuthResponse {
  accessToken: string;
  refreshToken: string;
  userId: string; // (format: uuid)
  name: string;
  email: string;
  tokenType: string; // e.g., "bearer"
}
// 2. DTO (Interface) 정의 - API Docs 기준

// GET /api/workspaces 응답
export interface WorkspaceResponse {
  id: string; // ❌ groupId 아님
  name: string;
  description: string; // ❌ companyName 아님
  ownerId: string;
  ownerName: string;
  ownerEmail: string;
  createdAt: string;
  updatedAt: string;
}

// POST /api/workspaces 요청
export interface CreateWorkspaceRequest {
  name: string;
  // ⚠️ API 스펙에 companyName이 없습니다.
  // description?: string; // (추측) 만약 description을 받는다면
}

// 3. API 함수 수정

/**
 * 워크스페이스 목록 조회 (GET /api/workspaces)
 */
export const getWorkspaces = async (accessToken: string): Promise<WorkspaceResponse[]> => {
  const response = await apiClient.get('/api/workspaces', {
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
  const response = await apiClient.post('/api/workspaces', data, {
    headers: { Authorization: `Bearer ${accessToken}` },
  });
  return response.data;
};

// ❌ createUserInfo 함수는 제거합니다.
// 백엔드 리포트에 따르면
// 워크스페이스 생성 시 (createWorkspace) 호출한 유저가 자동으로 'OWNER'가 됩니다.
