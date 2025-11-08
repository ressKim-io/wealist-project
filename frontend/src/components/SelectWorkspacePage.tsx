// src/components/SelectWorkspacePage.tsx (라우터 적용 수정본)

import React, { useState, useEffect, useMemo } from 'react';
import { useNavigate } from 'react-router-dom'; // 1. useNavigate 임포트
import { useTheme } from '../contexts/ThemeContext';
import {
  WorkspaceResponse,
  CreateWorkspaceRequest,
  getWorkspaces,
  createWorkspace,
} from '../api/user/userService';
import { Search, Plus, X, AlertCircle, Settings } from 'lucide-react';
import WorkspaceManagementModal from './modals/WorkspaceManagementModal';

// 2. Props 인터페이스 제거 (더 이상 App.tsx에서 props를 받지 않음)
/*
interface SelectWorkspacePageProps {
  userId: string;
  accessToken: string;
  onWorkspaceSelected: (workspaceId: string) => void;
}
*/

type WorkspacePageStep = 'list' | 'create-form' | 'add-members' | 'loading';

interface PendingMember {
  id: string;
  email: string;
}

// 3. props 제거
const SelectWorkspacePage: React.FC = () => {
  const navigate = useNavigate(); // 4. navigate 훅 사용
  const { theme } = useTheme();

  // 5. localStorage에서 토큰 및 ID 직접 조회
  const accessToken = localStorage.getItem('access_token') || '';
  const userId = localStorage.getItem('user_id') || ''; // OWNER 확인용

  // 페이지 상태
  const [step, setStep] = useState<WorkspacePageStep>('list');
  const [workspaces, setWorkspaces] = useState<WorkspaceResponse[] | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState('');

  // 폼 상태
  const [newWorkspaceName, setNewWorkspaceName] = useState('');
  const [newCompany, setNewCompany] = useState(''); // (Description)

  // 멤버 초대
  const [pendingMembers, setPendingMembers] = useState<PendingMember[]>([]);
  const [memberEmail, setMemberEmail] = useState('');
  const [memberEmailError, setMemberEmailError] = useState<string | null>(null);

  const [_createdWorkspaceId, setCreatedWorkspaceId] = useState<string | null>(null);

  // 워크스페이스 관리 모달
  const [managingWorkspace, setManagingWorkspace] = useState<WorkspaceResponse | null>(null);

  // 1. 초기 워크스페이스 로드
  useEffect(() => {
    const fetchWorkspaces = async () => {
      if (!accessToken) {
        // 토큰이 없으면 로그인 페이지로 (방어 코드)
        navigate('/');
        return;
      }

      setIsLoading(true);
      setError(null);
      try {
        const fetchedWorkspaces = await getWorkspaces(accessToken);
        setWorkspaces(fetchedWorkspaces);
      } catch (e) {
        const err = e as Error;
        setError(`워크스페이스 목록 조회 실패: ${err.message}`);
        setWorkspaces([]);
      } finally {
        setIsLoading(false);
      }
    };

    fetchWorkspaces();
  }, [accessToken, navigate]); // 의존성에 navigate 추가

  // 2. 검색 필터
  const availableWorkspaces = useMemo(() => {
    if (!workspaces) return [];
    const query = searchQuery.toLowerCase().trim();
    if (!query) return workspaces;
    return workspaces.filter(
      (ws) => ws.name.toLowerCase().includes(query) || ws.description.toLowerCase().includes(query),
    );
  }, [searchQuery, workspaces]);

  // 3. 이메일 유효성 검사 (동일)
  const isValidEmail = (email: string) => {
    return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email);
  };

  // 4. 멤버 추가 (동일)
  const handleAddMember = () => {
    setMemberEmailError(null);
    if (!memberEmail.trim()) {
      setMemberEmailError('이메일을 입력해주세요');
      return;
    }
    if (!isValidEmail(memberEmail)) {
      setMemberEmailError('유효한 이메일 주소를 입력해주세요');
      return;
    }
    if (pendingMembers.some((m) => m.email === memberEmail)) {
      setMemberEmailError('이미 추가된 이메일입니다');
      return;
    }
    setPendingMembers([...pendingMembers, { id: Date.now().toString(), email: memberEmail }]);
    setMemberEmail('');
  };

  // 5. 멤버 제거 (동일)
  const handleRemoveMember = (id: string) => {
    setPendingMembers(pendingMembers.filter((m) => m.id !== id));
  };

  // 6. 워크스페이스 생성 (onWorkspaceSelected -> navigate)
  const handleCreateWorkspaceWithMembers = async () => {
    if (!newWorkspaceName.trim()) {
      setError('워크스페이스 이름을 입력해주세요');
      return;
    }
    setIsLoading(true);
    setError(null);
    try {
      const createData: CreateWorkspaceRequest = {
        name: newWorkspaceName,
        // description: newCompany || 'Personal', // (필요시 API DTO 수정 후)
      };
      const newWorkspace = await createWorkspace(createData, accessToken);
      const newWorkspaceId = newWorkspace.id;
      setCreatedWorkspaceId(newWorkspaceId);

      for (const member of pendingMembers) {
        console.log(`멤버 초대 예정: ${member.email}`);
      }

      alert(
        `워크스페이스 '${newWorkspaceName}' 생성 완료! ${pendingMembers.length}명의 멤버 초대 예정입니다.`,
      );

      resetCreateForm();
      // 6. [수정] props 콜백 대신 navigate로 페이지 이동
      navigate(`/kanban/${newWorkspaceId}`);
    } catch (e) {
      const err = e as Error;
      setError(`워크스페이스 생성 실패: ${err.message}`);
      setIsLoading(false);
    }
  };

  // 7. 기존 워크스페이스 선택 (onWorkspaceSelected -> navigate)
  const handleSelectExistingWorkspace = async (workspace: WorkspaceResponse) => {
    setIsLoading(true);
    setError(null);
    try {
      alert(`워크스페이스 '${workspace.name}'에 참여 완료!`);

      // 6. [수정] props 콜백 대신 navigate로 페이지 이동
      navigate(`/kanban/${workspace.id}`);
    } catch (e) {
      const err = e as Error;
      setError(`워크스페이스 참여 실패: ${err.message}`);
      setIsLoading(false);
    }
  };

  // 워크스페이스 관리 모달 열기
  const handleManageWorkspace = (workspace: WorkspaceResponse) => {
    setManagingWorkspace(workspace);
  };

  // 8. 폼 초기화 (동일)
  const resetCreateForm = () => {
    setNewWorkspaceName('');
    setNewCompany('');
    setPendingMembers([]);
    setMemberEmail('');
    setMemberEmailError(null);
    setStep('list');
  };

  // --- 로딩 화면 (동일) ---
  if (isLoading && workspaces === null) {
    return (
      <div
        className={`min-h-screen ${theme.colors.background} flex items-center justify-center p-4`}
      >
        <div className="p-8">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500 mx-auto mb-4"></div>
          <p className={`${theme.font.size.lg} ${theme.colors.text}`}>
            워크스페이스 정보를 확인 중...
          </p>
        </div>
      </div>
    );
  }

  // --- 메인 렌더링 (동일) ---
  return (
    <div className={`min-h-screen ${theme.colors.background} flex items-center justify-center p-4`}>
      <div
        className={`${theme.colors.card} ${theme.effects.borderRadius} p-6 sm:p-8 w-full max-w-2xl relative z-10 shadow-xl ${theme.effects.cardBorderWidth} ${theme.colors.border}`}
      >
        {/* Step 1: 워크스페이스 목록 & 선택 */}
        {step === 'list' && (
          <>
            <h2
              className={`text-center ${theme.font.size.xl} font-extrabold ${theme.colors.text} mb-2`}
            >
              워크스페이스 선택
            </h2>
            <p className={`text-center mb-6 ${theme.font.size.sm} ${theme.colors.subText}`}>
              기존 워크스페이스에 참여하거나 새로운 워크스페이스를 생성하세요.
            </p>

            {error && (
              <div
                className={`${theme.colors.danger} text-center mb-4 ${theme.font.size.sm} border border-red-300 p-2 rounded-md bg-red-50 flex items-center gap-2`}
              >
                <AlertCircle className="w-4 h-4 flex-shrink-0" />
                {error}
              </div>
            )}

            {/* 검색 */}
            <div className="relative mb-4">
              <input
                type="text"
                placeholder="워크스페이스 이름 또는 설명으로 검색"
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className={`w-full px-4 pl-10 py-3 ${theme.colors.secondary} ${theme.font.size.sm} rounded-lg border-2 border-gray-300 focus:border-blue-500 focus:ring-2 focus:ring-blue-200 transition`}
              />
              <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-gray-400" />
            </div>

            {/* 워크스페이스 목록 */}
            <div
              className={`max-h-60 overflow-y-auto border-2 ${theme.colors.border} rounded-lg mb-4`}
            >
              {availableWorkspaces.length > 0 ? (
                availableWorkspaces.map((ws) => (
                  <div
                    key={ws.id}
                    onClick={() => !isLoading && handleSelectExistingWorkspace(ws)}
                    className={`w-full text-left p-4 hover:bg-blue-50 border-b border-gray-100 ${
                      theme.colors.text
                    } ${
                      theme.font.size.sm
                    } transition flex justify-between items-center last:border-b-0 ${
                      isLoading ? 'opacity-50 cursor-not-allowed' : 'cursor-pointer'
                    }`}
                  >
                    <div>
                      <span className="font-semibold">{ws.name}</span>
                      <p className={`${theme.colors.subText} ${theme.font.size.xs}`}>
                        {ws.description}
                      </p>
                    </div>
                    <div className="flex items-center gap-2">
                      {/* {ws.ownerId === userId && ( */}
                      <button
                        onClick={(e) => {
                          e.stopPropagation();
                          handleManageWorkspace(ws);
                        }}
                        className="p-2 hover:bg-gray-200 rounded-lg transition"
                        title="워크스페이스 관리"
                      >
                        <Settings className="w-4 h-4 text-gray-600" />
                      </button>
                      {/* )} */}
                      <span
                        className={`${theme.colors.info} ${theme.font.size.xs} px-2 py-1 border border-blue-200 rounded`}
                      >
                        선택
                      </span>
                    </div>
                  </div>
                ))
              ) : (
                <p className={`p-4 text-center ${theme.colors.subText} ${theme.font.size.sm}`}>
                  {searchQuery.trim() ? '검색 결과가 없습니다.' : '소속된 워크스페이스가 없습니다.'}
                </p>
              )}
            </div>

            {/* 새 워크스페이스 생성 버튼 */}
            <button
              onClick={() => setStep('create-form')}
              disabled={isLoading}
              className={`w-full ${theme.colors.primary} text-white py-3 font-bold rounded-lg ${theme.colors.primaryHover} transition disabled:opacity-50 shadow-lg flex items-center justify-center gap-2`}
            >
              <Plus className="w-5 h-5" /> 새 워크스페이스 생성
            </button>
          </>
        )}

        {/* Step 2: 워크스페이스 정보 입력 */}
        {step === 'create-form' && (
          <>
            <h2
              className={`text-center ${theme.font.size.xl} font-extrabold ${theme.colors.text} mb-2`}
            >
              새로운 워크스페이스 생성
            </h2>
            <p className={`text-center mb-6 ${theme.font.size.sm} ${theme.colors.subText}`}>
              워크스페이스 정보를 입력하세요.
            </p>

            {error && (
              <div
                className={`${theme.colors.danger} text-center mb-4 ${theme.font.size.sm} border border-red-300 p-2 rounded-md bg-red-50`}
              >
                {error}
              </div>
            )}

            <div className="space-y-4 mb-6">
              <div>
                <label
                  className={`block ${theme.font.size.sm} font-semibold ${theme.colors.text} mb-2`}
                >
                  워크스페이스 이름 *
                </label>
                <input
                  type="text"
                  placeholder="예: Orange Cloud 팀"
                  value={newWorkspaceName}
                  onChange={(e) => setNewWorkspaceName(e.target.value)}
                  className={`w-full px-4 py-3 ${theme.colors.secondary} ${theme.font.size.sm} rounded-lg border-2 border-gray-300 focus:border-blue-500 focus:ring-2 focus:ring-blue-200 transition`}
                  disabled={isLoading}
                />
              </div>

              <div>
                <label
                  className={`block ${theme.font.size.sm} font-semibold ${theme.colors.text} mb-2`}
                >
                  설명 (선택)
                </label>
                <input
                  type="text"
                  placeholder="예: Orange Cloud 프로젝트"
                  value={newCompany}
                  onChange={(e) => setNewCompany(e.target.value)}
                  className={`w-full px-4 py-3 ${theme.colors.secondary} ${theme.font.size.sm} rounded-lg border-2 border-gray-300 focus:border-blue-500 focus:ring-2 focus:ring-blue-200 transition`}
                  disabled={isLoading}
                />
              </div>
            </div>

            <div className="flex gap-3">
              <button
                onClick={() => resetCreateForm()}
                disabled={isLoading}
                className={`flex-1 ${theme.colors.secondary} text-gray-700 py-3 font-bold rounded-lg border-2 ${theme.colors.border} hover:bg-gray-100 transition disabled:opacity-50`}
              >
                ← 돌아가기
              </button>
              <button
                onClick={() => setStep('add-members')}
                disabled={isLoading || !newWorkspaceName.trim()}
                className={`flex-1 ${theme.colors.primary} text-white py-3 font-bold rounded-lg ${theme.colors.primaryHover} transition disabled:opacity-50`}
              >
                다음: 멤버 초대 →
              </button>
            </div>
          </>
        )}

        {/* Step 3: 멤버 초대 (동일) */}
        {step === 'add-members' && (
          <>
            <h2
              className={`${theme.font.size.xl} font-extrabold ${theme.colors.text} mb-2 text-center`}
            >
              멤버 초대 (선택사항)
            </h2>
            <p className={`text-center mb-6 ${theme.font.size.sm} ${theme.colors.subText}`}>
              초대할 멤버의 이메일을 입력하세요. 나중에 추가할 수도 있습니다.
            </p>

            {error && (
              <div
                className={`${theme.colors.danger} text-center mb-4 ${theme.font.size.sm} border border-red-300 p-2 rounded-md bg-red-50`}
              >
                {error}
              </div>
            )}

            {/* 이메일 입력 폼 */}
            <div className="mb-4 space-y-2 w-full">
              <div className="flex gap-2 w-full">
                <input
                  type="email"
                  placeholder="멤버 이메일"
                  value={memberEmail}
                  onChange={(e) => setMemberEmail(e.target.value)}
                  onKeyPress={(e) => e.key === 'Enter' && handleAddMember()}
                  className={`w-full px-4 py-3 ${theme.colors.secondary} ${theme.font.size.sm} rounded-lg border-2 border-gray-300 focus:border-blue-500 focus:ring-2 focus:ring-blue-200 transition`}
                  disabled={isLoading}
                />
                <button
                  onClick={handleAddMember}
                  disabled={isLoading || !memberEmail.trim()}
                  className={`px-4 py-3 font-bold rounded-lg flex-shrink-0 transition ${
                    isLoading || !memberEmail.trim()
                      ? 'bg-blue-300 text-white cursor-not-allowed opacity-60'
                      : 'bg-blue-500 text-white hover:bg-blue-600'
                  }`}
                >
                  <Plus className="w-5 h-5" />
                </button>
              </div>
              {memberEmailError && (
                <p className={`${theme.font.size.xs} ${theme.colors.danger}`}>{memberEmailError}</p>
              )}
            </div>

            {/* 추가된 멤버 목록 (동일) */}
            {pendingMembers.length > 0 && (
              <div className="mb-6 p-4 bg-blue-50 rounded-lg border border-blue-200">
                <p className={`${theme.font.size.sm} font-semibold ${theme.colors.text} mb-3`}>
                  초대 예정 멤버 ({pendingMembers.length}명)
                </p>
                <div className="space-y-2">
                  {pendingMembers.map((member) => (
                    <div
                      key={member.id}
                      className="flex items-center justify-between bg-white p-2 rounded border border-gray-200"
                    >
                      <span className={`${theme.font.size.sm} ${theme.colors.text}`}>
                        {member.email}
                      </span>
                      <button
                        onClick={() => handleRemoveMember(member.id)}
                        className="text-red-500 hover:text-red-700 transition"
                      >
                        <X className="w-4 h-4" />
                      </button>
                    </div>
                  ))}
                </div>
              </div>
            )}

            {/* 액션 버튼 (동일) */}
            <div className="flex gap-3">
              <button
                onClick={() => setStep('create-form')}
                disabled={isLoading}
                className={`flex-1 ${theme.colors.secondary} text-gray-700 py-3 font-bold rounded-lg border-2 ${theme.colors.border} hover:bg-gray-100 transition disabled:opacity-50`}
              >
                ← 이전
              </button>
              <button
                onClick={handleCreateWorkspaceWithMembers}
                disabled={isLoading || !newWorkspaceName.trim()}
                className={`flex-1 ${theme.colors.success} text-white py-3 font-bold rounded-lg hover:bg-green-600 transition disabled:opacity-50`}
              >
                {isLoading ? '생성 중...' : '워크스페이스 생성 완료'}
              </button>
            </div>
          </>
        )}
      </div>

      {/* 워크스페이스 관리 모달 */}
      {managingWorkspace && (
        <WorkspaceManagementModal
          workspaceId={managingWorkspace.id}
          workspaceName={managingWorkspace.name}
          onClose={() => setManagingWorkspace(null)}
        />
      )}
    </div>
  );
};

export default SelectWorkspacePage;
