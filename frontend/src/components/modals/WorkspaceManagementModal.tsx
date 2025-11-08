/**
 * 워크스페이스 관리 모달 컴포넌트 (OWNER/ADMIN용)
 *
 * [백엔드 개발자 참고사항]
 *
 * 모든 API 로직은 src/api/user/userService.ts에 구현되어 있습니다.
 * userService.ts 파일에서 USE_MOCK_DATA 플래그를 false로 변경하면
 * 자동으로 실제 API를 호출합니다.
 *
 * 필요한 백엔드 API:
 * 1. GET  /api/workspaces/{workspaceId}/settings           - 워크스페이스 설정 조회
 * 2. PUT  /api/workspaces/{workspaceId}/settings           - 워크스페이스 설정 업데이트
 * 3. GET  /api/workspaces/{workspaceId}/members            - 회원 목록 조회
 * 4. GET  /api/workspaces/{workspaceId}/pending-members    - 승인 대기 목록 조회
 * 5. POST /api/workspaces/{workspaceId}/members/{userId}/approve - 회원 승인
 * 6. POST /api/workspaces/{workspaceId}/members/{userId}/reject  - 회원 거절
 * 7. PUT  /api/workspaces/{workspaceId}/members/{userId}/role    - 회원 역할 변경
 * 8. DELETE /api/workspaces/{workspaceId}/members/{userId}       - 회원 퇴출
 * 9. GET  /api/workspaces/{workspaceId}/invitable-users?query=   - 초대 가능 회원 검색
 * 10. POST /api/workspaces/{workspaceId}/invite/{userId}         - 회원 초대
 */

import React, { useState, useEffect, useMemo } from 'react';
import { X, Search } from 'lucide-react';
import { useTheme } from '../../contexts/ThemeContext';
import { useAuth } from '../../contexts/AuthContext';
import {
  WorkspaceSettings,
  WorkspaceMember,
  PendingMember as PendingMemberType,
  InvitableUser,
  WorkspaceMemberRole,
  getWorkspaceSettings,
  updateWorkspaceSettings,
  getWorkspaceMembers,
  getPendingMembers,
  approveMember,
  rejectMember,
  updateMemberRole,
  removeMember,
  searchInvitableUsers,
  inviteUser,
} from '../../api/user/userService';

interface WorkspaceManagementModalProps {
  workspaceId: string;
  workspaceName: string;
  onClose: () => void;
}

const WorkspaceManagementModal: React.FC<WorkspaceManagementModalProps> = ({
  workspaceId,
  workspaceName,
  onClose,
}) => {
  const { theme } = useTheme();
  const { token } = useAuth();

  // ========================================
  // 상태 관리
  // ========================================

  // 탭 상태: 'settings' (기본정보) | 'members' (회원관리)
  const [activeTab, setActiveTab] = useState<'settings' | 'members'>('settings');

  // 워크스페이스 설정
  const [settings, setSettings] = useState<WorkspaceSettings | null>(null);
  const [settingsForm, setSettingsForm] = useState({
    name: '',
    description: '',
    isPublic: false,
    requiresApproval: false,
    onlyOwnerCanInvite: false,
  });

  // 회원 관리
  const [members, setMembers] = useState<WorkspaceMember[]>([]);
  const [pendingMembers, setPendingMembers] = useState<PendingMemberType[]>([]);
  const [invitableUsers, setInvitableUsers] = useState<InvitableUser[]>([]);
  const [searchQuery, setSearchQuery] = useState('');
  const [memberSearchQuery, setMemberSearchQuery] = useState('');

  // 로딩 및 에러
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // ========================================
  // 초기 데이터 로드
  // ========================================

  useEffect(() => {
    const loadData = async () => {
      if (!token) {
        setError('인증 토큰이 없습니다.');
        return;
      }

      try {
        setLoading(true);
        const [settingsData, membersData, pendingData] = await Promise.all([
          getWorkspaceSettings(workspaceId, token),
          getWorkspaceMembers(workspaceId, token),
          getPendingMembers(workspaceId, token),
        ]);

        setSettings(settingsData);
        setSettingsForm({
          name: settingsData.name,
          description: settingsData.description,
          isPublic: settingsData.isPublic,
          requiresApproval: settingsData.requiresApproval,
          onlyOwnerCanInvite: settingsData.onlyOwnerCanInvite,
        });
        setMembers(membersData);
        setPendingMembers(pendingData);
      } catch (err) {
        console.error('[WorkspaceManagement] 데이터 로드 실패:', err);
        setError('워크스페이스 정보를 불러오는데 실패했습니다.');
      } finally {
        setLoading(false);
      }
    };

    loadData();
  }, [workspaceId, token]);

  // 회원 검색
  useEffect(() => {
    const loadInvitableUsers = async () => {
      if (activeTab !== 'members' || !token) return;

      try {
        const users = await searchInvitableUsers(workspaceId, searchQuery, token);
        setInvitableUsers(users);
      } catch (err) {
        console.error('[WorkspaceManagement] 초대 가능 회원 검색 실패:', err);
      }
    };

    const debounce = setTimeout(() => {
      loadInvitableUsers();
    }, 300);

    return () => clearTimeout(debounce);
  }, [searchQuery, activeTab, workspaceId, token]);

  // ========================================
  // 기본정보 핸들러
  // ========================================

  const handleSaveSettings = async () => {
    if (!token) return;

    try {
      setLoading(true);
      setError(null);
      const updated = await updateWorkspaceSettings(workspaceId, settingsForm, token);
      setSettings(updated);
      alert('워크스페이스 설정이 저장되었습니다.');
    } catch (err) {
      console.error('[WorkspaceManagement] 설정 저장 실패:', err);
      setError('설정 저장에 실패했습니다.');
    } finally {
      setLoading(false);
    }
  };

  // ========================================
  // 회원관리 핸들러
  // ========================================

  const handleInviteUser = async (userId: string) => {
    if (!token) return;

    try {
      setLoading(true);
      await inviteUser(workspaceId, userId, token);

      // 목록 새로고침
      const [membersData, pendingData] = await Promise.all([
        getWorkspaceMembers(workspaceId, token),
        getPendingMembers(workspaceId, token),
      ]);
      setMembers(membersData);
      setPendingMembers(pendingData);

      alert('회원 초대가 완료되었습니다.');
    } catch (err) {
      console.error('[WorkspaceManagement] 회원 초대 실패:', err);
      setError('회원 초대에 실패했습니다.');
    } finally {
      setLoading(false);
    }
  };

  const handleApproveMember = async (userId: string) => {
    if (!token) return;

    try {
      setLoading(true);
      await approveMember(workspaceId, userId, token);

      // 목록 새로고침
      const [membersData, pendingData] = await Promise.all([
        getWorkspaceMembers(workspaceId, token),
        getPendingMembers(workspaceId, token),
      ]);
      setMembers(membersData);
      setPendingMembers(pendingData);

      alert('회원 승인이 완료되었습니다.');
    } catch (err) {
      console.error('[WorkspaceManagement] 회원 승인 실패:', err);
      setError('회원 승인에 실패했습니다.');
    } finally {
      setLoading(false);
    }
  };

  const handleRejectMember = async (userId: string) => {
    if (!token) return;

    try {
      setLoading(true);
      await rejectMember(workspaceId, userId, token);

      // 목록 새로고침
      const pendingData = await getPendingMembers(workspaceId, token);
      setPendingMembers(pendingData);

      alert('회원 요청을 거절했습니다.');
    } catch (err) {
      console.error('[WorkspaceManagement] 회원 거절 실패:', err);
      setError('회원 거절에 실패했습니다.');
    } finally {
      setLoading(false);
    }
  };

  const handleUpdateRole = async (userId: string, role: WorkspaceMemberRole) => {
    if (!token) return;

    try {
      setLoading(true);
      await updateMemberRole(workspaceId, userId, role, token);

      // 목록 새로고침
      const membersData = await getWorkspaceMembers(workspaceId, token);
      setMembers(membersData);

      alert(`회원 역할이 ${role}로 변경되었습니다.`);
    } catch (err) {
      console.error('[WorkspaceManagement] 역할 변경 실패:', err);
      setError('역할 변경에 실패했습니다.');
    } finally {
      setLoading(false);
    }
  };

  const handleRemoveMember = async (userId: string, userName: string) => {
    if (!confirm(`정말 ${userName}님을 퇴출하시겠습니까?`)) return;
    if (!token) return;

    try {
      setLoading(true);
      await removeMember(workspaceId, userId, token);

      // 목록 새로고침
      const membersData = await getWorkspaceMembers(workspaceId, token);
      setMembers(membersData);

      alert(`${userName}님이 퇴출되었습니다.`);
    } catch (err) {
      console.error('[WorkspaceManagement] 회원 퇴출 실패:', err);
      setError('회원 퇴출에 실패했습니다.');
    } finally {
      setLoading(false);
    }
  };

  // ========================================
  // 회원 목록 필터링
  // ========================================

  const filteredMembers = useMemo(() => {
    if (!memberSearchQuery.trim()) return members;
    const query = memberSearchQuery.toLowerCase();
    return members.filter(
      (member) =>
        member.name.toLowerCase().includes(query) ||
        member.email.toLowerCase().includes(query) ||
        member.role.toLowerCase().includes(query),
    );
  }, [members, memberSearchQuery]);

  // ========================================
  // 렌더링
  // ========================================

  if (!settings) {
    return (
      <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
        <div className="bg-white p-8 rounded-xl shadow-lg">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500 mx-auto mb-4"></div>
          <p className="text-gray-700">워크스페이스 정보를 불러오는 중...</p>
        </div>
      </div>
    );
  }

  return (
    <div
      className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50"
      onClick={onClose}
    >
      <div className="relative w-full max-w-2xl" onClick={(e) => e.stopPropagation()}>
        <div
          className={`relative ${theme.colors.card} ${theme.effects.borderWidth} ${theme.colors.border} ${theme.effects.borderRadius} shadow-xl max-h-[90vh] overflow-y-auto`}
        >
          {/* 헤더 */}
          <div className="flex items-center justify-between p-4 pb-3">
            <h2 className={`${theme.font.size.base} font-bold text-gray-800`}>워크스페이스 관리</h2>
            <button
              onClick={onClose}
              className="p-2 hover:bg-gray-100 rounded-lg transition"
              title="닫기"
            >
              <X className="w-4 h-4 text-gray-600" />
            </button>
          </div>

          {/* 탭 메뉴 */}
          <div className="flex border-b border-gray-200 px-6">
            <button
              onClick={() => setActiveTab('settings')}
              className={`flex-1 py-3 text-sm font-medium transition-colors relative ${
                activeTab === 'settings' ? 'text-blue-600' : 'text-gray-500 hover:text-gray-700'
              }`}
            >
              기본정보
              {activeTab === 'settings' && (
                <div className="absolute bottom-0 left-0 right-0 h-0.5 bg-blue-600"></div>
              )}
            </button>
            <button
              onClick={() => setActiveTab('members')}
              className={`flex-1 py-3 text-sm font-medium transition-colors relative ${
                activeTab === 'members' ? 'text-blue-600' : 'text-gray-500 hover:text-gray-700'
              }`}
            >
              회원관리
              {activeTab === 'members' && (
                <div className="absolute bottom-0 left-0 right-0 h-0.5 bg-blue-600"></div>
              )}
            </button>
          </div>

          {/* 탭 컨텐츠 */}
          <div className="p-6 space-y-5">
            {/* 에러 메시지 */}
            {error && (
              <div className="p-3 bg-red-100 border border-red-400 text-red-700 rounded-md text-sm">
                {error}
              </div>
            )}

            {/* 기본정보 탭 */}
            {activeTab === 'settings' && (
              <div className="space-y-4">
                <div>
                  <label className={`block ${theme.font.size.xs} mb-2 text-gray-500 font-medium`}>
                    워크스페이스 이름:
                  </label>
                  <input
                    type="text"
                    value={settingsForm.name}
                    onChange={(e) => setSettingsForm({ ...settingsForm, name: e.target.value })}
                    className={`w-full px-3 py-2 ${theme.effects.cardBorderWidth} ${theme.colors.border} ${theme.colors.card} ${theme.font.size.xs} ${theme.effects.borderRadius} focus:outline-none focus:ring-2 focus:ring-blue-500`}
                    placeholder="워크스페이스 이름"
                  />
                </div>

                <div>
                  <label className={`block ${theme.font.size.xs} mb-2 text-gray-500 font-medium`}>
                    설명:
                  </label>
                  <textarea
                    value={settingsForm.description}
                    onChange={(e) =>
                      setSettingsForm({ ...settingsForm, description: e.target.value })
                    }
                    rows={3}
                    className={`w-full px-3 py-2 ${theme.effects.cardBorderWidth} ${theme.colors.border} ${theme.colors.card} ${theme.font.size.xs} ${theme.effects.borderRadius} focus:outline-none focus:ring-2 focus:ring-blue-500`}
                    placeholder="워크스페이스 설명"
                  />
                </div>

                <div className="space-y-3">
                  {/* 공개/비공개 토글 */}
                  <div className="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
                    <div>
                      <p className="text-sm font-medium text-gray-700">공개 워크스페이스</p>
                      <p className="text-xs text-gray-500">
                        공개 시 다른 사용자가 검색하여 가입 신청할 수 있습니다
                      </p>
                    </div>
                    <button
                      onClick={() =>
                        setSettingsForm({ ...settingsForm, isPublic: !settingsForm.isPublic })
                      }
                      className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors ${
                        settingsForm.isPublic ? 'bg-blue-600' : 'bg-gray-300'
                      }`}
                    >
                      <span
                        className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform ${
                          settingsForm.isPublic ? 'translate-x-6' : 'translate-x-1'
                        }`}
                      />
                    </button>
                  </div>

                  {/* 승인제/비승인제 토글 */}
                  <div className="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
                    <div>
                      <p className="text-sm font-medium text-gray-700">가입 승인제</p>
                      <p className="text-xs text-gray-500">
                        활성화 시 가입 신청을 승인해야 회원이 됩니다
                      </p>
                    </div>
                    <button
                      onClick={() =>
                        setSettingsForm({
                          ...settingsForm,
                          requiresApproval: !settingsForm.requiresApproval,
                        })
                      }
                      className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors ${
                        settingsForm.requiresApproval ? 'bg-blue-600' : 'bg-gray-300'
                      }`}
                    >
                      <span
                        className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform ${
                          settingsForm.requiresApproval ? 'translate-x-6' : 'translate-x-1'
                        }`}
                      />
                    </button>
                  </div>

                  {/* OWNER 전용 초대 토글 */}
                  <div className="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
                    <div>
                      <p className="text-sm font-medium text-gray-700">OWNER만 초대 가능</p>
                      <p className="text-xs text-gray-500">
                        활성화 시 OWNER만 새 회원을 초대할 수 있습니다
                      </p>
                    </div>
                    <button
                      onClick={() =>
                        setSettingsForm({
                          ...settingsForm,
                          onlyOwnerCanInvite: !settingsForm.onlyOwnerCanInvite,
                        })
                      }
                      className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors ${
                        settingsForm.onlyOwnerCanInvite ? 'bg-blue-600' : 'bg-gray-300'
                      }`}
                    >
                      <span
                        className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform ${
                          settingsForm.onlyOwnerCanInvite ? 'translate-x-6' : 'translate-x-1'
                        }`}
                      />
                    </button>
                  </div>
                </div>

                <button
                  onClick={handleSaveSettings}
                  disabled={loading}
                  className={`w-full ${theme.colors.primary} text-white py-3 ${
                    theme.effects.borderRadius
                  } font-semibold transition ${
                    loading ? 'opacity-50 cursor-not-allowed' : 'hover:opacity-90'
                  }`}
                >
                  {loading ? '저장 중...' : '설정 저장'}
                </button>
              </div>
            )}

            {/* 회원관리 탭 */}
            {activeTab === 'members' && (
              <div className="space-y-6">
                {/* 회원 검색 */}
                <div>
                  <label className={`block ${theme.font.size.xs} mb-2 text-gray-500 font-medium`}>
                    회원 검색 및 초대:
                  </label>
                  <div className="relative">
                    <input
                      type="text"
                      placeholder="이름 또는 이메일로 검색"
                      value={searchQuery}
                      onChange={(e) => setSearchQuery(e.target.value)}
                      className={`w-full px-3 pl-10 py-2 ${theme.effects.cardBorderWidth} ${theme.colors.border} ${theme.colors.card} ${theme.font.size.xs} ${theme.effects.borderRadius} focus:outline-none focus:ring-2 focus:ring-blue-500`}
                    />
                    <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-gray-400" />
                  </div>
                </div>

                {/* 초대 가능 회원 목록 */}
                {searchQuery && invitableUsers.length > 0 && (
                  <div className="border border-gray-200 rounded-lg p-4 bg-gray-50">
                    <p className="text-sm font-semibold text-gray-700 mb-3">초대 가능 회원</p>
                    <div className="space-y-2 max-h-40 overflow-y-auto">
                      {invitableUsers.map((user) => (
                        <div
                          key={user.userId}
                          className="flex items-center justify-between bg-white p-2 rounded border border-gray-200"
                        >
                          <div>
                            <p className="text-sm font-medium text-gray-700">{user.name}</p>
                            <p className="text-xs text-gray-500">{user.email}</p>
                          </div>
                          <button
                            onClick={() => handleInviteUser(user.userId)}
                            disabled={loading}
                            className="px-3 py-1 bg-blue-500 text-white text-xs rounded hover:bg-blue-600 transition disabled:opacity-50"
                          >
                            초대
                          </button>
                        </div>
                      ))}
                    </div>
                  </div>
                )}

                {/* 승인 대기 목록 */}
                {pendingMembers.length > 0 && (
                  <div className="border border-gray-200 rounded-lg p-4">
                    <p className="text-sm font-semibold text-gray-700 mb-3">
                      승인 대기 목록 ({pendingMembers.length}명)
                    </p>
                    <div className="space-y-2">
                      {pendingMembers.map((member) => (
                        <div
                          key={member.userId}
                          className="flex items-center justify-between bg-gray-50 p-3 rounded border border-gray-200"
                        >
                          <div>
                            <p className="text-sm font-medium text-gray-700">{member.name}</p>
                            <p className="text-xs text-gray-500">{member.email}</p>
                          </div>
                          <div className="flex gap-2">
                            <button
                              onClick={() => handleApproveMember(member.userId)}
                              disabled={loading}
                              className="px-3 py-1 bg-green-500 text-white text-xs rounded hover:bg-green-600 transition disabled:opacity-50"
                            >
                              승인
                            </button>
                            <button
                              onClick={() => handleRejectMember(member.userId)}
                              disabled={loading}
                              className="px-3 py-1 bg-red-500 text-white text-xs rounded hover:bg-red-600 transition disabled:opacity-50"
                            >
                              거절
                            </button>
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>
                )}

                {/* 조직 회원 목록 */}
                <div className="border border-gray-200 rounded-lg p-4">
                  <div className="flex items-center justify-between mb-3 gap-3">
                    <p className="text-sm font-semibold text-gray-700 whitespace-nowrap">
                      조직 회원 목록 ({members.length}명)
                    </p>
                    <div className="relative flex-1 max-w-xs">
                      <input
                        type="text"
                        placeholder="이름, 이메일, 역할 검색"
                        value={memberSearchQuery}
                        onChange={(e) => setMemberSearchQuery(e.target.value)}
                        className={`w-full px-3 pl-8 py-1.5 ${theme.effects.cardBorderWidth} ${theme.colors.border} ${theme.colors.card} text-xs ${theme.effects.borderRadius} focus:outline-none focus:ring-2 focus:ring-blue-500`}
                      />
                      <Search className="absolute left-2.5 top-1/2 transform -translate-y-1/2 w-3.5 h-3.5 text-gray-400" />
                    </div>
                  </div>
                  <div className="space-y-2">
                    {filteredMembers.map((member) => (
                      <div
                        key={member.userId}
                        className="flex items-center justify-between bg-gray-50 p-3 rounded border border-gray-200"
                      >
                        <div>
                          <p className="text-sm font-medium text-gray-700">
                            {member.name}{' '}
                            <span className="text-xs text-blue-600 font-semibold">
                              ({member.role})
                            </span>
                          </p>
                          <p className="text-xs text-gray-500">{member.email}</p>
                        </div>
                        {member.role !== 'OWNER' && (
                          <div className="flex gap-2">
                            {member.role === 'MEMBER' && (
                              <button
                                onClick={() => handleUpdateRole(member.userId, 'ADMIN')}
                                disabled={loading}
                                className="px-3 py-1 bg-blue-500 text-white text-xs rounded hover:bg-blue-600 transition disabled:opacity-50"
                              >
                                ADMIN
                              </button>
                            )}
                            {member.role === 'ADMIN' && (
                              <>
                                <button
                                  onClick={() => handleUpdateRole(member.userId, 'MEMBER')}
                                  disabled={loading}
                                  className="px-3 py-1 bg-gray-500 text-white text-xs rounded hover:bg-gray-600 transition disabled:opacity-50"
                                >
                                  MEMBER
                                </button>
                                <button
                                  onClick={() => handleUpdateRole(member.userId, 'OWNER')}
                                  disabled={loading}
                                  className="px-3 py-1 bg-purple-500 text-white text-xs rounded hover:bg-purple-600 transition disabled:opacity-50"
                                >
                                  OWNER
                                </button>
                              </>
                            )}
                            <button
                              onClick={() => handleRemoveMember(member.userId, member.name)}
                              disabled={loading}
                              className="px-3 py-1 bg-red-500 text-white text-xs rounded hover:bg-red-600 transition disabled:opacity-50"
                            >
                              퇴출
                            </button>
                          </div>
                        )}
                      </div>
                    ))}
                    {filteredMembers.length === 0 && (
                      <p className="text-center text-sm text-gray-500 py-4">
                        {memberSearchQuery.trim()
                          ? '검색 결과가 없습니다.'
                          : '조직 회원이 없습니다.'}
                      </p>
                    )}
                  </div>
                </div>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
};

export default WorkspaceManagementModal;
