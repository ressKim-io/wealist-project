// src/App.tsx (수정본)

import React, { useState, Suspense, lazy } from 'react';
import { ThemeProvider } from './contexts/ThemeContext';
import { AuthResponse } from './api/userService';
import SelectWorkspacePage from './components/SelectWorkspacePage';
// import { createWorkspace, WorkspaceCreate } from './api/KanbanService'; // 주석처리: 에러 방지

// --- 변경 ---
type AppState = 'AUTH' | 'SELECT_WORKSPACE' | 'CREATE_WORKSPACE' | 'KANBAN';

// Lazy load 페이지들
const AuthPage = lazy(() => import('./pages/Authpage'));
// 💡 컴포넌트 파일 이름이 SelectGroupPage이더라도,
//    이 컴포넌트는 이제 Workspace를 선택하는 역할을 합니다.
const SelectGroupPage = lazy(() => import('./components/SelectWorkspacePage'));
const MainDashboard = lazy(() => import('./pages/Dashboard'));
const OAuthRedirectPage = lazy(() => import('./pages/OAuthRedirectPage'));

const LoadingScreen = ({ msg = '로딩 중..' }) => (
  // ... (로딩 스크린 코드는 동일)
  <div className="text-center min-h-screen flex items-center justify-center bg-gray-50">
    <div className="p-8 bg-white rounded-xl shadow-lg">
      <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500 mx-auto mb-4"></div>
      <h1 className="text-xl font-medium text-gray-800">{msg}</h1>
    </div>
  </div>
);

const App: React.FC = () => {
  // --- 변경 ---
  const [appState, setAppState] = useState<AppState>('AUTH');
  const [accessToken, setAccessToken] = useState<string | null>(null);
  const [userId, setUserId] = useState<string | null>(null);
  // --- 변경 ---
  const [currentWorkspaceId, setCurrentWorkspaceId] = useState<string | null>(null);
  // const [loadingMessage, setLoadingMessage] = useState<string | null>(null);

  // handleAuthSuccess는 OAuthRedirectPage에서 호출됩니다.
  const handleAuthSuccess = (authData: AuthResponse) => {
    if (authData.accessToken && authData.userId) {
      setAccessToken(authData.accessToken);
      setUserId(authData.userId);
      localStorage.setItem('access_token', authData.accessToken);
      localStorage.setItem('user_id', authData.userId);
      // --- 변경 ---
      setAppState('SELECT_WORKSPACE');
    } else {
      handleLogout();
    }
  };

  const handleLogout = () => {
    localStorage.removeItem('access_token');
    localStorage.removeItem('user_id');
    setAccessToken(null);
    setUserId(null);
    // --- 변경 ---
    setCurrentWorkspaceId(null);
    setAppState('AUTH');
  };

  // --- 함수 이름 및 인자 변경 ---
  const handleWorkspaceSelectionSuccess = (workspaceId: string) => {
    if (!accessToken || !userId) {
      handleLogout();
      return;
    }
    // --- 변경 ---
    setCurrentWorkspaceId(workspaceId);
    setAppState('KANBAN');
  };
  /*
  // (참고) 기존 handleGroupSelectionSuccess 로직
  const handleGroupSelectionSuccess = (groupId: string) => {
    if (!accessToken || !userId) {
      handleLogout();
      return;
    }
    setCurrentGroupId(groupId);
    setAppState('KANBAN');
  };
  */

  const renderContent = () => {
    // 1. OAuth Redirect Check (동일)
    const urlParams = new URLSearchParams(window.location.search);
    const hasOAuthTokens = urlParams.has('accessToken') && urlParams.has('refreshToken');
    const isRedirectPath = window.location.pathname.includes('/oauth/redirect');

    if (isRedirectPath || hasOAuthTokens) {
      return <OAuthRedirectPage onAuthSuccess={handleAuthSuccess} />;
    }

    // 2. Standard State Routing
    if (appState === 'AUTH') {
      return <AuthPage />;
    }

    // --- 변경 ---
    if (appState === 'SELECT_WORKSPACE' && userId && accessToken) {
      return (
        <SelectWorkspacePage
          userId={userId}
          accessToken={accessToken}
          // --- prop 이름 및 핸들러 변경 ---
          onWorkspaceSelected={handleWorkspaceSelectionSuccess}
        />
      );
    }

    // if (appState === 'CREATE_WORKSPACE') { ... }

    // --- 변경 ---
    if (appState === 'KANBAN' && currentWorkspaceId && accessToken) {
      return (
        <MainDashboard
          onLogout={handleLogout}
          // --- prop 이름 및 값 변경 ---
          currentGroupId={currentWorkspaceId}
          accessToken={accessToken}
        />
      );
    }

    // 3. Fallback (동일)
    return <AuthPage />;
  };

  return (
    <ThemeProvider>
      <Suspense fallback={<LoadingScreen />}>{renderContent()}</Suspense>
    </ThemeProvider>
  );
};

export default App;
