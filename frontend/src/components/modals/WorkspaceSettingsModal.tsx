import React, { useState, useEffect, useCallback, useMemo } from 'react';
import { Plus, X, Search, Users, Briefcase, AlertTriangle } from 'lucide-react';
import { useTheme } from '../../contexts/ThemeContext';
import { useAuth } from '../../contexts/AuthContext'; // ✅ useAuth 임포트
import {
  WorkspaceMember,
  getWorkspaceMembers,
  inviteMemberByEmail,
  updateMemberRole,
  removeMember,
} from '../../api/user/workspaceService'; // ✅ API 경로 수정 (../api/user/workspaceService)

// 💡 Mock 데이터: 프로젝트 현황 상세 (Workspace GENERAL 탭에서 사용)
interface ProjectStatus {
  id: string;
  name: string;
  memberCount: number;
  taskCount: number;
  lastUpdated: string;
}
const getMockProjectStatus = (): ProjectStatus[] => {
  return [
    {
      id: 'prj-1',
      name: 'Wealist 서비스 개발',
      memberCount: 4,
      taskCount: 22,
      lastUpdated: '2025-10-31',
    },
    {
      id: 'prj-2',
      name: 'Orange Cloud 디자인 시스템',
      memberCount: 2,
      taskCount: 15,
      lastUpdated: '2025-10-28',
    },
    {
      id: 'prj-3',
      name: '내부 인프라 구축 (EKS)',
      memberCount: 3,
      taskCount: 8,
      lastUpdated: '2025-11-01',
    },
  ];
};

interface WorkspaceSettingsModalProps {
  workspaceId: string;
  workspaceName: string;
  onClose: () => void;
}

export const WorkspaceSettingsModal: React.FC<WorkspaceSettingsModalProps> = ({
  workspaceId,
  workspaceName,
  onClose,
}) => {
  const { theme } = useTheme();
  const { token, userId: currentUserId } = useAuth(); // ✅ useAuth 훅 사용

  const [activeTab, setActiveTab] = useState<'MEMBERSHIP' | 'GENERAL'>('GENERAL');

  // --- 멤버십 관리 상태 ---
  const [members, setMembers] = useState<WorkspaceMember[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState('');

  // 초대 폼 상태
  const [inviteEmail, setInviteEmail] = useState('');
  const [inviteRole, setInviteRole] = useState<'ORGANIZER' | 'MEMBER'>('MEMBER');
  const [isInviting, setIsInviting] = useState(false);
  const [inviteMessage, setInviteMessage] = useState<string | null>(null);

  // 현재 로그인한 사용자의 역할 확인 (권한 확인용)
  const currentUserRole = useMemo(() => {
    return members.find((m) => m.userId === currentUserId)?.roleName || 'MEMBER';
  }, [members, currentUserId]);

  // 조직장(MASTER)이거나 운영자(ORGANIZER)인지 확인
  const isManager = currentUserRole === 'MASTER' || currentUserRole === 'ORGANIZER';
  const isMaster = currentUserRole === 'MASTER'; // 조직장은 MASTER만 해당

  // 워크스페이스 이름 상태 (General 탭에서 사용)
  const [name, setName] = useState(workspaceName);
  const [description, setDescription] = useState('칸반 보드를 위한 설정');

  // --- 데이터 로딩 함수 ---
  const fetchMembers = useCallback(async () => {
    if (!token) return;
    setIsLoading(true);
    setError(null);
    try {
      // ✅ API 함수 사용
      const fetchedMembers = await getWorkspaceMembers(workspaceId, token);
      setMembers(fetchedMembers);
    } catch (err) {
      console.error('워크스페이스 멤버 조회 실패:', err);
      setError('워크스페이스 멤버 정보를 불러오는데 실패했습니다.');
    } finally {
      setIsLoading(false);
    }
  }, [workspaceId, token]);

  useEffect(() => {
    fetchMembers();
  }, [fetchMembers]);

  // --- 멤버 관리 로직 ---

  // 1. 역할 변경 (MASTER는 MASTER만 변경 가능)
  const handleChangeRole = async (memberId: string, currentRole: WorkspaceMember['roleName']) => {
    if (!isMaster || !token) return; // MASTER만 역할 변경 가능

    const newRole = currentRole === 'ORGANIZER' ? 'MEMBER' : 'ORGANIZER';
    if (
      !window.confirm(
        `${members.find((m) => m.id === memberId)?.userName} 님의 역할을 ${
          newRole === 'ORGANIZER' ? '운영자' : '팀원'
        }으로 변경하시겠습니까?`,
      )
    ) {
      return;
    }
    try {
      setIsLoading(true);
      // ✅ API 함수 사용
      await updateMemberRole(workspaceId, memberId, newRole, token);
      await fetchMembers(); // 변경 후 목록 새로고침
    } catch (err: any) {
      console.error('역할 변경 실패:', err);
      alert(`역할 변경 실패: ${err.message}`);
    } finally {
      setIsLoading(false);
    }
  };

  // 2. 멤버 제거 (MASTER는 MASTER가 아닌 멤버만 제거 가능)
  const handleRemoveMember = async (memberId: string) => {
    if (!isMaster || !token) return; // MASTER만 멤버 제거 가능

    const member = members.find((m) => m.id === memberId);
    if (!member || member.roleName === 'MASTER') return; // 자기 자신(MASTER) 또는 다른 MASTER 제거 불가

    if (!window.confirm(`${member.userName} 님을 워크스페이스에서 제거하시겠습니까?`)) {
      return;
    }
    try {
      setIsLoading(true);
      // ✅ API 함수 사용
      await removeMember(workspaceId, memberId, token);
      await fetchMembers(); // 제거 후 목록 새로고침
    } catch (err: any) {
      console.error('멤버 제거 실패:', err);
      alert(`멤버 제거 실패: ${err.message}`);
    } finally {
      setIsLoading(false);
    }
  };

  // 3. 멤버 초대 (이메일)
  const handleInviteMember = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!isManager || !token || isInviting) return;
    if (!inviteEmail || !inviteRole) {
      setInviteMessage('이메일과 역할을 선택해주세요.');
      return;
    }

    setIsInviting(true);
    setInviteMessage(null);
    try {
      // ✅ API 함수 사용
      await inviteMemberByEmail(workspaceId, inviteEmail, inviteRole, token);
      setInviteMessage(`✅ ${inviteEmail} 님을 성공적으로 초대했습니다.`);
      setInviteEmail('');
      setInviteRole('MEMBER');

      await fetchMembers();
    } catch (err: any) {
      console.error('멤버 초대 실패:', err);
      setInviteMessage(`❌ 초대 실패: ${err.message}`);
    } finally {
      setIsInviting(false);
      setTimeout(() => setInviteMessage(null), 5000); // 5초 후 메시지 제거
    }
  };

  // --- 유틸리티 및 렌더링 (이하 코드는 수정 없음) ---

  const filteredMembers = members.filter(
    (member) =>
      member.userName.toLowerCase().includes(searchQuery.toLowerCase()) ||
      member.userEmail.toLowerCase().includes(searchQuery.toLowerCase()),
  );

  const getRoleLabel = (role: WorkspaceMember['roleName']) => {
    switch (role) {
      case 'MASTER':
        return { text: '조직장 (MASTER)', color: 'bg-red-500 text-white font-semibold' };
      case 'ORGANIZER':
        return { text: '운영자 (ORGANIZER)', color: 'bg-yellow-300 text-yellow-900 font-medium' };
      case 'MEMBER':
      default:
        return { text: '팀원 (MEMBER)', color: 'bg-blue-100 text-blue-700 font-medium' };
    }
  };

  // 💡 멤버 목록 렌더링 컴포넌트 (MEMBERSHIP 탭)
  const MemberListContent = () => {
    if (isLoading) {
      return <div className="text-center py-8 text-gray-500">멤버 목록을 불러오는 중...</div>;
    }
    if (error) {
      return (
        <div className="text-center py-8 text-red-500 bg-red-50 rounded-lg">
          <AlertTriangle className="w-5 h-5 inline mr-2" /> {error}
        </div>
      );
    }

    return (
      <>
        <h3 className="text-sm font-semibold text-gray-600 mb-2">
          전체 조직원 ({filteredMembers.length}명)
        </h3>
        <div className="max-h-80 overflow-y-auto space-y-2 p-1 -m-1">
          {filteredMembers.length > 0 ? (
            filteredMembers.map((member) => {
              const isSelf = member.userId === currentUserId;
              // MASTER만 역할 변경/제거 권한 가짐. 자기 자신은 역할 변경 불가.
              const canChange = isMaster && !isSelf && member.roleName !== 'MASTER';
              const canRemove = isMaster && member.roleName !== 'MASTER' && !isSelf;

              return (
                <div
                  key={member.id}
                  className="flex items-center justify-between p-3 bg-gray-50 hover:bg-gray-100 rounded-lg transition"
                >
                  <div className="flex items-center gap-3">
                    {/* 아바타 */}
                    <div className="w-8 h-8 rounded-full bg-blue-500 flex items-center justify-center text-white text-xs font-bold">
                      {member.userName ? member.userName[0] : member.userEmail[0]}
                    </div>

                    {/* 이름 및 역할 */}
                    <div>
                      <span className="text-sm font-medium text-gray-800">
                        {member.userName || '이름 없음'} {isSelf && '(나)'}
                      </span>
                      <p className="text-xs text-gray-500 truncate">{member.userEmail}</p>
                      <div className="flex items-center mt-0.5 space-x-2">
                        <span
                          className={`text-xs px-2 py-0.5 rounded-full ${
                            getRoleLabel(member.roleName).color
                          }`}
                        >
                          {getRoleLabel(member.roleName).text}
                        </span>
                        {member.isDefault && (
                          <span className="text-xs font-medium px-2 py-0.5 rounded-full bg-green-100 text-green-800">
                            기본 워크스페이스
                          </span>
                        )}
                      </div>
                    </div>
                  </div>

                  {/* 액션 버튼 영역 */}
                  {(canChange || canRemove) && (
                    <div className="flex items-center gap-2">
                      {/* 역할 변경 버튼 */}
                      {canChange && (
                        <button
                          onClick={() => handleChangeRole(member.id, member.roleName)}
                          className={`text-xs px-3 py-1 rounded-full transition ${
                            member.roleName === 'ORGANIZER'
                              ? 'bg-yellow-500 text-white hover:bg-yellow-600'
                              : 'bg-blue-500 text-white hover:bg-blue-600'
                          }`}
                        >
                          {member.roleName === 'ORGANIZER' ? '팀원 지정' : '운영자 지정'}
                        </button>
                      )}

                      {/* 멤버 제거 버튼 */}
                      {canRemove && (
                        <button
                          onClick={() => handleRemoveMember(member.id)}
                          className="text-xs px-3 py-1 rounded-full transition bg-red-500 text-white hover:bg-red-600"
                        >
                          제거
                        </button>
                      )}
                    </div>
                  )}
                </div>
              );
            })
          ) : (
            <p className="text-center py-4 text-gray-500">
              {searchQuery ? '검색 결과가 없습니다.' : '워크스페이스에 멤버가 없습니다.'}
            </p>
          )}
        </div>
      </>
    );
  };

  // 💡 조직원 초대 폼 (MEMBERSHIP 탭)
  const MemberInvitationForm = () => (
    <form
      onSubmit={handleInviteMember}
      className="space-y-3 p-4 bg-white border border-dashed border-gray-300 rounded-lg"
    >
      <h4 className="text-md font-bold text-gray-800">새 조직원 초대</h4>
      <div className="flex gap-2">
        <input
          type="email"
          placeholder="초대할 조직원의 이메일 주소"
          value={inviteEmail}
          onChange={(e) => setInviteEmail(e.target.value)}
          required
          className="flex-grow px-3 py-2 border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
          disabled={isInviting || !isManager}
        />
        <select
          value={inviteRole}
          onChange={(e) => setInviteRole(e.target.value as 'ORGANIZER' | 'MEMBER')}
          className="px-3 py-2 border rounded-lg text-sm bg-white focus:outline-none focus:ring-2 focus:ring-blue-500"
          disabled={isInviting || !isManager}
        >
          <option value="MEMBER">팀원 (MEMBER)</option>
          <option value="ORGANIZER">운영자 (ORGANIZER)</option>
        </select>
        <button
          type="submit"
          className={`flex items-center gap-1 px-4 py-2 text-sm font-semibold rounded-lg transition shadow-md ${
            isInviting || !isManager
              ? 'bg-gray-400 cursor-not-allowed'
              : 'bg-blue-500 hover:bg-blue-600'
          } text-white`}
          disabled={isInviting || !isManager}
        >
          {isInviting ? (
            '초대 중...'
          ) : (
            <>
              <Plus className="w-4 h-4" />
              초대
            </>
          )}
        </button>
      </div>
      {inviteMessage && (
        <p
          className={`text-xs p-2 rounded-lg ${
            inviteMessage.startsWith('✅')
              ? 'bg-green-100 text-green-800'
              : 'bg-red-100 text-red-800'
          }`}
        >
          {inviteMessage}
        </p>
      )}
      {!isManager && (
        <p className="text-xs text-red-500">
          조직원 초대 및 관리는 조직장(MASTER) 또는 운영자(ORGANIZER)만 가능합니다.
        </p>
      )}
    </form>
  );

  // 💡 일반 설정 탭 내용 (GENERAL 탭)
  const GeneralSettingsContent = () => {
    const projectStatus = getMockProjectStatus();

    // (저장 로직은 API 호출이 필요하므로 현재는 UI만 구성)

    return (
      <div className="space-y-6">
        {/* 1. 기본 정보 섹션 */}
        <div className="p-4 bg-gray-50 rounded-lg border space-y-4">
          <label className="block text-sm font-medium text-gray-700 mb-1">워크스페이스 이름</label>
          <input
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            className="w-full px-3 py-2 border rounded-lg text-sm"
            disabled={!isMaster}
          />
          <label className="block text-sm font-medium text-gray-700 mb-1">워크스페이스 설명</label>
          <textarea
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            className="w-full px-3 py-2 border rounded-lg text-sm min-h-20"
            disabled={!isMaster}
          />
          {!isMaster && (
            <p className="text-xs text-red-500">
              워크스페이스 기본 정보 수정은 조직장(MASTER)만 가능합니다.
            </p>
          )}
        </div>

        {/* 2. 현황 정보 섹션 (프로젝트 현황 목록) */}
        <div className="pt-4">
          <h3 className="text-md font-bold text-gray-800 mb-3">
            <Briefcase className="w-5 h-5 inline mr-2 text-blue-500" />
            프로젝트 현황 (총 {projectStatus.length}개)
          </h3>
          <div className="max-h-80 overflow-y-auto space-y-3 p-1 -m-1">
            {projectStatus.map((project) => (
              <div
                key={project.id}
                className="p-3 bg-white border border-gray-200 rounded-lg shadow-sm"
              >
                <div className="flex items-center justify-between">
                  <h4 className="text-sm font-semibold text-gray-800 truncate">{project.name}</h4>
                  <span className="text-xs text-gray-500">
                    {project.lastUpdated.slice(5)} 업데이트
                  </span>
                </div>

                <div className="flex gap-4 mt-2 text-sm">
                  <span className="text-gray-700 font-medium">팀원: {project.memberCount}명</span>
                  <span className="text-gray-700 font-medium">태스크: {project.taskCount}개</span>
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* 3. 저장 버튼 */}
        <div className="pt-6 border-t border-gray-200">
          <button
            className={`w-full py-2 font-semibold rounded-lg transition ${
              isMaster
                ? 'bg-blue-500 text-white hover:bg-blue-600'
                : 'bg-gray-300 text-gray-600 cursor-not-allowed'
            }`}
            disabled={!isMaster}
          >
            워크스페이스 저장
          </button>
        </div>
      </div>
    );
  };

  return (
    <div
      className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-[90]"
      onClick={onClose}
    >
      <div className="relative w-full max-w-lg" onClick={(e) => e.stopPropagation()}>
        <div
          className={`relative ${theme.colors.card} ${theme.effects.borderWidth} ${theme.colors.border} p-6 ${theme.effects.borderRadius} shadow-xl`}
        >
          {/* 헤더/탭 구조 */}
          <div className="flex items-center justify-between mb-4 border-b border-gray-200 -mt-4 -mx-6 px-6 pt-4">
            <div className="flex">
              {/* 탭 버튼 */}
              <button
                onClick={() => setActiveTab('GENERAL')}
                className={`py-2 px-4 text-sm font-semibold transition ${
                  activeTab === 'GENERAL'
                    ? 'text-blue-600 border-b-2 border-blue-600'
                    : 'text-gray-500 hover:text-gray-700'
                }`}
              >
                일반 설정 &amp; 현황
              </button>
              <button
                onClick={() => setActiveTab('MEMBERSHIP')}
                className={`py-2 px-4 text-sm font-semibold transition ${
                  activeTab === 'MEMBERSHIP'
                    ? 'text-blue-600 border-b-2 border-blue-600'
                    : 'text-gray-500 hover:text-gray-700'
                }`}
              >
                조직원/역할 관리
              </button>
            </div>
            {/* 닫기 버튼 */}
            <button
              onClick={onClose}
              className="p-2 rounded-full hover:bg-gray-100 text-gray-500 hover:text-gray-700"
            >
              <X className="w-5 h-5" />
            </button>
          </div>

          <div className="space-y-4">
            {/* 일반 설정 및 현황 탭 내용 */}
            {activeTab === 'GENERAL' && <GeneralSettingsContent />}

            {/* 멤버십 탭 (조직원 관리) */}
            {activeTab === 'MEMBERSHIP' && (
              <div className="space-y-4">
                {/* 1. 조직원 초대 폼 */}
                <MemberInvitationForm />

                {/* 2. 검색 및 멤버 목록 */}
                <div className="flex gap-3">
                  {/* 검색 필드 */}
                  <div className="relative flex-grow">
                    <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-gray-400" />
                    <input
                      type="text"
                      placeholder="이름 또는 이메일로 조직원 검색..."
                      value={searchQuery}
                      onChange={(e) => setSearchQuery(e.target.value)}
                      className="w-full px-4 py-2 pl-10 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm"
                      disabled={isLoading}
                    />
                  </div>
                </div>

                {/* 멤버 목록 */}
                <MemberListContent />

                {/* 권한 설명 메시지 */}
                <p className="text-sm text-gray-500 mt-4 p-3 bg-gray-100 rounded-lg border border-gray-200">
                  <Users className="w-4 h-4 inline mr-1 text-blue-500" />
                  현재 당신의 역할은 {getRoleLabel(currentUserRole).text}이며, 역할 변경/제거 권한은
                  조직장(MASTER)에게 있습니다.
                </p>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
};
