import axios from 'axios';

// 1. User/Workspace 서비스 (Java 백엔드) 기본 URL
// (API 문서의 servers[0] URL을 사용)
export const USER_REPO_API_URL = 'http://localhost:8080';

// 2. Board/Project 서비스 (Go 백엔드) 기본 URL
// (boardService.ts에서 사용하는 기본 URL)
export const BOARD_SERVICE_API_URL = 'http://localhost:8000';

/**
 * User Repo API (Java)를 위한 Axios 인스턴스
 */
export const userRepoClient = axios.create({
  baseURL: USER_REPO_API_URL,
  headers: { 'Content-Type': 'application/json' },
});

/**
 * Board Service API (Go)를 위한 Axios 인스턴스
 */
export const boardServiceClient = axios.create({
  baseURL: BOARD_SERVICE_API_URL,
  headers: { 'Content-Type': 'application/json' },
});

/**
 * JWT 토큰을 포함하는 인증 헤더를 반환합니다.
 */
export const getAuthHeaders = (token: string) => ({
  Authorization: `Bearer ${token}`,
  Accept: 'application/json',
});
