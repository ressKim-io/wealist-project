// src/components/SelectGroupPage.tsx (수정 완료된 전체 코드)

import React, { useState, useEffect, useMemo } from 'react';
import { useTheme } from '../contexts/ThemeContext';
import {
  // --- 변경된 Import ---
  WorkspaceResponse,
  CreateWorkspaceRequest,
  getWorkspaces,
  createWorkspace,
  // ❌ createUserInfo 제거
  // ---
} from '../api/userService'; // 경로가 다르다면 수정하세요
import { Search, Plus, X, AlertCircle } from 'lucide-react';

// 1. Props 인터페이스 수정
interface SelectWorkspacePageProps {
  userId: string;
  accessToken: string;
  onWorkspaceSelected: (workspaceId: string) => void; // 이름 변경 (onGroupSelected -> onWorkspaceSelected)
}

type WorkspacePageStep = 'list' | 'create-form' | 'add-members' | 'loading';

interface PendingMember {
  id: string;
  email: string;
}

// 2. Props 이름 변경 (destructuring)
const SelectWorkspacePage: React.FC<SelectWorkspacePageProps> = ({
  userId,
  accessToken,
  onWorkspaceSelected, // 이름 변경 (onGroupSelected -> onWorkspaceSelected)
}) => {
  const { theme } = useTheme();

  // 페이지 상태
  const [step, setStep] = useState<WorkspacePageStep>('list');
  // --- State 이름 변경 ---
  const [workspaces, setWorkspaces] = useState<WorkspaceResponse[] | null>(null);
  // ---
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState('');

  // 조직 생성 폼
  const [newGroupName, setNewGroupName] = useState('');
  const [newCompany, setNewCompany] = useState(''); // ⚠️ API 스펙 참고 (description으로 변경 필요)

  // 멤버 초대 관리
  const [pendingMembers, setPendingMembers] = useState<PendingMember[]>([]);
  const [memberEmail, setMemberEmail] = useState('');
  const [memberEmailError, setMemberEmailError] = useState<string | null>(null);

  // 조직 생성 중 저장된 데이터
  const [_createdGroupId, setCreatedGroupId] = useState<string | null>(null);

  // 1. 초기 워크스페이스 로드
  useEffect(() => {
    const fetchWorkspaces = async () => {
      if (!accessToken) return;

      setIsLoading(true);
      setError(null);

      try {
        // --- API 함수 변경 ---
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
  }, [accessToken]);

  // 2. 검색 필터
  const availableGroups = useMemo(() => {
    if (!workspaces) return []; // 'workspaces' 사용
    const query = searchQuery.toLowerCase().trim();
    if (!query) return workspaces;
    return workspaces.filter(
      (
        ws, // 'ws' (workspace)
      ) =>
        ws.name.toLowerCase().includes(query) ||
        // --- DTO 필드 변경 (companyName -> description) ---
        ws.description.toLowerCase().includes(query),
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

    // 중복 확인
    if (pendingMembers.some((m) => m.email === memberEmail)) {
      setMemberEmailError('이미 추가된 이메일입니다');
      return;
    }

    // 추가
    setPendingMembers([...pendingMembers, { id: Date.now().toString(), email: memberEmail }]);
    setMemberEmail('');
  };

  // 5. 멤버 제거 (동일)
  const handleRemoveMember = (id: string) => {
    setPendingMembers(pendingMembers.filter((m) => m.id !== id));
  };

  // 6. 조직 생성 + 멤버 초대
  const handleCreateGroupWithMembers = async () => {
    if (!newGroupName.trim()) {
      setError('워크스페이스 이름을 입력해주세요');
      return;
    }

    setIsLoading(true);
    setError(null);

    try {
      // --- Step 1: 조직 생성 (API 스펙에 맞게 수정) ---
      const createData: CreateWorkspaceRequest = {
        name: newGroupName,
        // ⚠️ 'companyName'은 CreateWorkspaceRequest DTO에 없습니다.
        //    만약 'description'으로 보내야 한다면:
        // description: newCompany || 'Personal',
      };

      const newWorkspace = await createWorkspace(createData, accessToken);
      const newWorkspaceId = newWorkspace.id; // ❌ groupId 아님
      setCreatedGroupId(newWorkspaceId);

      // --- Step 2: (제거) ---
      // ℹ️ createWorkspace를 호출한 유저(OWNER)는 백엔드에서 자동 등록됩니다.

      // --- Step 3: (기존과 동일 - 콘솔 로그) ---
      for (const member of pendingMembers) {
        // ... (멤버 초대 로직 - 나중에 구현) ...
        console.log(`멤버 초대 예정: ${member.email}`);
      }

      // --- Step 4: 완료 ---
      alert(
        `워크스페이스 '${newGroupName}' 생성 완료! ${pendingMembers.length}명의 멤버 초대 예정입니다.`,
      );

      window.history.pushState(null, '', `/kanban/${newWorkspaceId}`);

      resetCreateForm();
      // 3. Props 콜백 변경
      onWorkspaceSelected(newWorkspaceId);
    } catch (e) {
      const err = e as Error;
      setError(`워크스페이스 생성 실패: ${err.message}`);
      setIsLoading(false);
    }
  };

  // 7. 기존 그룹 선택
  const handleSelectExistingGroup = async (workspace: WorkspaceResponse) => {
    setIsLoading(true);
    setError(null);

    try {
      alert(`워크스페이스 '${workspace.name}'에 참여 완료!`);

      // --- ❌ group.groupId -> workspace.id ---
      window.history.pushState(null, '', `/kanban/${workspace.id}`);

      // App.tsx의 상태 업데이트
      // 3. Props 콜백 변경
      onWorkspaceSelected(workspace.id); // ❌ group.groupId
    } catch (e) {
      const err = e as Error;
      setError(`워크스페이스 참여 실패: ${err.message}`);
      setIsLoading(false);
    }
  };

  // 8. 폼 초기화 (동일)
  const resetCreateForm = () => {
    setNewGroupName('');
    setNewCompany('');
    setPendingMembers([]);
    setMemberEmail('');
    setMemberEmailError(null);
    setStep('list');
  };

  // --- 로딩 화면 ---
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

  // --- 메인 렌더링 (UI 복구) ---
  return (
    <div className={`min-h-screen ${theme.colors.background} flex items-center justify-center p-4`}>
      <div
        className={`${theme.colors.card} ${theme.effects.borderRadius} p-6 sm:p-8 w-full max-w-2xl relative z-10 shadow-xl ${theme.effects.cardBorderWidth} ${theme.colors.border}`}
      >
        {/* Step 1: 워크스페이스 목록 & 선택 */}
        {step === 'list' && (
          <>
            <h2 className={`${theme.font.size.xl} font-extrabold ${theme.colors.text} mb-2`}>
              워크스페이스 조직 선택
            </h2>
            <p className={`text-center mb-6 ${theme.font.size.sm} ${theme.colors.subText}`}>
              기존 조직에 참여하거나 새로운 조직을 생성하세요.
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
              {availableGroups.length > 0 ? (
                availableGroups.map((ws) => (
                  <button
                    key={ws.id} // ❌ group.groupId
                    onClick={() => handleSelectExistingGroup(ws)} // ❌ group
                    disabled={isLoading}
                    className={`w-full text-left p-4 hover:bg-blue-50 border-b border-gray-100 ${theme.colors.text} ${theme.font.size.sm} transition flex justify-between items-center last:border-b-0`}
                  >
                    <div>
                      <span className="font-semibold">{ws.name}</span>
                      <p className={`${theme.colors.subText} ${theme.font.size.xs}`}>
                        {ws.description} {/* ❌ group.companyName */}
                      </p>
                    </div>
                    <span
                      className={`${theme.colors.info} ${theme.font.size.xs} px-2 py-1 border border-blue-200 rounded`}
                    >
                      선택
                    </span>
                  </button>
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
            <h2 className={`${theme.font.size.xl} font-extrabold ${theme.colors.text} mb-2`}>
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
                  value={newGroupName}
                  onChange={(e) => setNewGroupName(e.target.value)}
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
                disabled={isLoading || !newGroupName.trim()}
                className={`flex-1 ${theme.colors.primary} text-white py-3 font-bold rounded-lg ${theme.colors.primaryHover} transition disabled:opacity-50`}
              >
                다음: 멤버 초대 →
              </button>
            </div>
          </>
        )}

        {/* Step 3: 멤버 초대 */}
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

            {/* 추가된 멤버 목록 */}
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

            {/* 액션 버튼 */}
            <div className="flex gap-3">
              <button
                onClick={() => setStep('create-form')}
                disabled={isLoading}
                className={`flex-1 ${theme.colors.secondary} text-gray-700 py-3 font-bold rounded-lg border-2 ${theme.colors.border} hover:bg-gray-100 transition disabled:opacity-50`}
              >
                ← 이전
              </button>
              <button
                onClick={handleCreateGroupWithMembers}
                disabled={isLoading || !newGroupName.trim()}
                className={`flex-1 ${theme.colors.success} text-white py-3 font-bold rounded-lg hover:bg-green-600 transition disabled:opacity-50`}
              >
                {isLoading ? '생성 중...' : '워크스페이스 생성 완료'}
              </button>
            </div>
          </>
        )}
      </div>
    </div>
  );
};

export default SelectWorkspacePage;
