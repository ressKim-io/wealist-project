import axios, { AxiosError, AxiosInstance, InternalAxiosRequestConfig } from 'axios';

// 환경 변수 가져오기
const INJECTED_API_BASE_URL = import.meta.env.VITE_API_BASE_URL;

// ============================================================================
// 💡 [핵심 수정]: Context Path를 환경에 따라 조건부로 붙입니다.
// ============================================================================

const getApiBaseUrl = (path: string): string => {
  // 1. 환경 변수 주입 확인
  if (INJECTED_API_BASE_URL) {
    // 쉘 스크립트에서 VITE_API_BASE_URL='http://localhost'가 주입된 경우
    const isLocalDevelopment = INJECTED_API_BASE_URL.includes('localhost');

    if (isLocalDevelopment)
      return `${INJECTED_API_BASE_URL}${path === '/api/users' ? ':8080' : ':8000/api'}`;

    return `${INJECTED_API_BASE_URL}${path}`;
  }

  // 환경 변수가 없을 경우 (Fallback, CI/CD 실패 대비)
  return `https://api.wealist.co.kr${path}`;
};

export const USER_REPO_API_URL = getApiBaseUrl('/api/users');
export const BOARD_SERVICE_API_URL = getApiBaseUrl('/api/boards/api');
// export const BOARD_WS_URL = getApiBaseUrl('/api/ws/project');
// ============================================================================
// 인증 갱신 관련 변수
// ============================================================================

let isRefreshing = false;
let failedQueue: Array<{
  resolve: (value?: unknown) => void;
  reject: (reason?: unknown) => void;
}> = [];

// ============================================================================
// 네트워크 재시도 설정
// ============================================================================

const MAX_RETRIES = 5; // 최대 5회 재시도 (총 6회 요청)
const RETRY_DELAY_MS = 1000; // 재시도 간격 (1초)

// ============================================================================
// Axios 인스턴스 생성
// ============================================================================

/**
 * User Repo API (Java)를 위한 Axios 인스턴스
 */
export const userRepoClient = axios.create({
  baseURL: USER_REPO_API_URL,
  headers: { 'Content-Type': 'application/json' },
  withCredentials: true, // CORS 인증 정보 포함
});

/**
 * Board Service API (Go)를 위한 Axios 인스턴스
 */
export const boardServiceClient = axios.create({
  baseURL: BOARD_SERVICE_API_URL,
  headers: { 'Content-Type': 'application/json' },
  withCredentials: true, // CORS 인증 정보 포함
});

// ============================================================================
// 인증 갱신 헬퍼 함수
// ============================================================================

const processQueue = (error: Error | null, token: string | null = null) => {
  failedQueue.forEach((prom) => {
    if (error) {
      prom.reject(error);
    } else {
      prom.resolve(token);
    }
  });
  failedQueue = [];
};

/**
 * localStorage를 정리하고 로그인 페이지로 리다이렉트합니다.
 */
const performLogout = () => {
  localStorage.removeItem('accessToken');
  localStorage.removeItem('refreshToken');
  localStorage.removeItem('nickName');
  localStorage.removeItem('userEmail');
  window.location.href = '/';
};

/**
 * Refresh Token을 사용하여 새로운 Access Token을 발급받습니다.
 */
const refreshAccessToken = async (): Promise<string> => {
  const refreshToken = localStorage.getItem('refreshToken');
  // Refresh Token이 없으면 즉시 로그아웃
  if (!refreshToken) {
    console.warn('⚠️ Refresh token not found. Logging out...');
    performLogout();
    throw new Error('No refresh token available');
  }

  try {
    // 💡 토큰 갱신 API는 인증 헤더 없이 리프레시 토큰만으로 호출되어야 함
    const response = await axios.post(`${USER_REPO_API_URL}/api/auth/refresh`, {
      refreshToken,
    });

    const { accessToken, refreshToken: newRefreshToken } = response.data;

    // 새로운 토큰들을 localStorage에 저장
    localStorage.setItem('accessToken', accessToken);
    if (newRefreshToken) {
      localStorage.setItem('refreshToken', newRefreshToken);
    }

    return accessToken;
  } catch (error) {
    // Refresh token도 만료된 경우 로그아웃 처리
    localStorage.removeItem('accessToken');
    localStorage.removeItem('refreshToken');
    localStorage.removeItem('nickName');
    localStorage.removeItem('userEmail');

    // 로그인 페이지로 리다이렉트
    window.location.href = '/'; // 💡 React Router를 사용하는 경우 navigate('/') 등으로 대체될 수 있음

    throw error;
  }
};

// ============================================================================
// 요청 인터셉터 설정 함수
// ============================================================================

/**
 * localStorage에서 accessToken을 자동으로 가져와 Authorization 헤더에 추가합니다.
 */
const setupRequestInterceptor = (client: AxiosInstance) => {
  client.interceptors.request.use(
    (config) => {
      // localStorage에서 accessToken 가져오기
      const accessToken = localStorage.getItem('accessToken');

      // Authorization 헤더가 이미 설정되어 있지 않고, accessToken이 있으면 추가
      if (accessToken && !config.headers.Authorization) {
        config.headers.Authorization = `Bearer ${accessToken}`;
      }

      return config;
    },
    (error) => {
      return Promise.reject(error);
    },
  );
};

// ============================================================================
// 통합 응답 인터셉터 설정 함수
// ============================================================================

/**
 * 두 가지 응답 인터셉터를 하나의 함수로 통합 설정합니다.
 */
const setupUnifiedResponseInterceptor = (client: AxiosInstance) => {
  client.interceptors.response.use(
    (response) => response,
    async (error: AxiosError) => {
      const originalRequest = error.config as InternalAxiosRequestConfig & {
        _retry?: boolean;
        retryCount?: number;
      };
      const status = error.response?.status;

      // ----------------------------------------
      // 1. 401 에러 (토큰 갱신) 처리
      // ----------------------------------------
      if (status === 401 && !originalRequest._retry) {
        if (isRefreshing) {
          // 이미 갱신 중이면 대기열에 추가하고 대기
          return new Promise((resolve, reject) => {
            failedQueue.push({ resolve, reject });
          })
            .then((token) => {
              originalRequest.headers.Authorization = `Bearer ${token}`;
              return client(originalRequest);
            })
            .catch((err) => {
              return Promise.reject(err);
            });
        }

        originalRequest._retry = true; // 토큰 갱신 시도 플래그
        isRefreshing = true;

        try {
          const newAccessToken = await refreshAccessToken();
          processQueue(null, newAccessToken);

          originalRequest.headers.Authorization = `Bearer ${newAccessToken}`;
          return client(originalRequest); // 새 토큰으로 원래 요청 재시도
        } catch (refreshError) {
          processQueue(refreshError as Error, null);
          return Promise.reject(refreshError);
        } finally {
          isRefreshing = false;
        }
      }

      // ----------------------------------------
      // 2. 네트워크 오류 재시도 방지 처리 (401 처리가 끝난 후 실행)
      // ----------------------------------------

      // 4xx, 5xx 에러는 백엔드 비즈니스 로직 오류이므로 재시도하지 않고 바로 에러 반환
      // (status가 없거나 599보다 큰 경우: 네트워크 단절 등)
      if (status && status >= 400 && status < 599) {
        return Promise.reject(error);
      }
      // 💡 네트워크 단절 오류 처리: response가 없고, 오류가 AbortError가 아닌 경우
      // 이는 서버가 완전히 꺼졌을 때 발생하는 오류(ERR_CONNECTION_REFUSED)를 포함합니다.
      if (!status && error.code !== 'ERR_CANCELED') {
        // Axios의 기본 취소 에러는 무시

        // 네트워크 단절 오류 처리:
        originalRequest.retryCount = originalRequest.retryCount || 0;

        if (originalRequest.retryCount >= MAX_RETRIES) {
          console.error(
            `[Axios Interceptor] 최대 재시도 횟수(${MAX_RETRIES}회) 초과. 요청 중단: ${originalRequest.url}`,
          );
          return Promise.reject(error);
        }

        originalRequest.retryCount += 1;

        const delay = new Promise((resolve) => {
          setTimeout(resolve, RETRY_DELAY_MS);
        });

        console.warn(
          `[Axios Interceptor] 요청 실패(${originalRequest.retryCount}회 재시도 중): ${originalRequest.url}`,
        );

        await delay;
        return client(originalRequest);
      }
    },
  );
};

// 💡 두 클라이언트 인스턴스에 인터셉터 적용
// 1. Request Interceptor: 자동으로 accessToken을 헤더에 추가
setupRequestInterceptor(userRepoClient);
setupRequestInterceptor(boardServiceClient);

// 2. Response Interceptor: 토큰 갱신 및 네트워크 오류 재시도
setupUnifiedResponseInterceptor(userRepoClient);
setupUnifiedResponseInterceptor(boardServiceClient);

/**
 * JWT 토큰을 포함하는 인증 헤더를 반환합니다.
 */
export const getAuthHeaders = (token: string) => ({
  Authorization: `Bearer ${token}`,
  Accept: 'application/json',
});
