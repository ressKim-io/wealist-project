/**
 * 사용자 프로필 모달 컴포넌트
 *
 * [백엔드 개발자 참고사항]
 *
 * 모든 API 로직은 src/api/user/userService.ts에 구현되어 있습니다.
 * userService.ts 파일에서 USE_MOCK_DATA 플래그를 false로 변경하면
 * 자동으로 실제 API를 호출합니다.
 *
 * 필요한 백엔드 API:
 * 1. GET  /api/profiles/me                         - 기본 프로필 조회
 * 2. PUT  /api/profiles/me                         - 기본 프로필 업데이트
 * 3. GET  /api/profiles/workspace/{workspaceId}    - 워크스페이스 프로필 조회
 * 4. PUT  /api/profiles/workspace/{workspaceId}    - 워크스페이스 프로필 생성/수정
 * 5. GET  /api/workspaces                          - 내가 속한 워크스페이스 목록
 */

import React, { useState, useRef, ChangeEvent, useEffect } from 'react';
import { X, Camera } from 'lucide-react';
import { useTheme } from '../../contexts/ThemeContext';
import { useAuth } from '../../contexts/AuthContext';
import { UserProfile } from '../../types';
import {
  getMyProfile,
  updateMyProfile,
  getWorkspaceProfile,
  updateWorkspaceProfile,
  getWorkspaces,
  WorkspaceResponse,
  UserProfileResponse,
} from '../../api/user/userService';

interface UserProfileModalProps {
  user: UserProfile;
  onClose: () => void;
}

const UserProfileModal: React.FC<UserProfileModalProps> = ({ user, onClose }) => {
  const { theme } = useTheme();
  const { token } = useAuth();

  // ========================================
  // 상태 관리
  // ========================================

  // 탭 상태: 'default' (기본 프로필) | 'workspace' (워크스페이스별 프로필)
  const [activeTab, setActiveTab] = useState<'default' | 'workspace'>('default');

  // 워크스페이스 목록
  const [workspaces, setWorkspaces] = useState<WorkspaceResponse[]>([]);
  const [selectedWorkspaceId, setSelectedWorkspaceId] = useState<string>('');

  // 파일 입력 Ref
  const fileInputRef = useRef<HTMLInputElement>(null);

  // 기본 프로필 상태
  const [defaultProfile, setDefaultProfile] = useState<UserProfileResponse | null>(null);
  const [defaultName, setDefaultName] = useState('');
  const [defaultEmail, setDefaultEmail] = useState('');

  // 워크스페이스 프로필 상태
  const [workspaceProfile, setWorkspaceProfile] = useState<UserProfileResponse | null>(null);
  const [workspaceName, setWorkspaceName] = useState('');
  const [workspaceEmail, setWorkspaceEmail] = useState('');

  // 프로필 이미지 미리보기 URL
  const [avatarPreviewUrl, setAvatarPreviewUrl] = useState<string | null>(null);

  // 로딩 및 에러 상태
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // ========================================
  // 초기 데이터 로드
  // ========================================

  useEffect(() => {
    const loadInitialData = async () => {
      if (!token) {
        setError('인증 토큰이 없습니다. 다시 로그인해주세요.');
        return;
      }

      try {
        setLoading(true);

        // 기본 프로필과 워크스페이스 목록 동시 로드
        const [profile, workspaceList] = await Promise.all([
          getMyProfile(token),
          getWorkspaces(token),
        ]);

        setDefaultProfile(profile);
        setDefaultName(profile.name);
        setDefaultEmail(profile.email || '');

        setWorkspaces(workspaceList);
        if (workspaceList.length > 0) {
          setSelectedWorkspaceId(workspaceList[0].id);
        }
      } catch (err) {
        console.error('[Initial Data Load Error]', err);
        setError('프로필 정보를 불러오는데 실패했습니다.');
      } finally {
        setLoading(false);
      }
    };

    loadInitialData();
  }, [token]);

  // ========================================
  // 워크스페이스 프로필 로드
  // ========================================

  useEffect(() => {
    const loadWorkspaceProfile = async () => {
      if (!token || !selectedWorkspaceId || activeTab !== 'workspace') {
        return;
      }

      try {
        const profile = await getWorkspaceProfile(selectedWorkspaceId, token);

        if (profile) {
          setWorkspaceProfile(profile);
          setWorkspaceName(profile.name);
          setWorkspaceEmail(profile.email || '');
        } else {
          // 프로필이 없으면 기본 프로필 정보로 초기화
          setWorkspaceProfile(null);
          const workspace = workspaces.find((ws) => ws.id === selectedWorkspaceId);
          setWorkspaceName(`${defaultProfile?.name || ''} (${workspace?.name || ''})`);
          setWorkspaceEmail(defaultProfile?.email || '');
        }
      } catch (err) {
        console.error('[Workspace Profile Load Error]', err);
        // 에러 발생 시 기본값으로 설정
        setWorkspaceProfile(null);
        const workspace = workspaces.find((ws) => ws.id === selectedWorkspaceId);
        setWorkspaceName(`${defaultProfile?.name || ''} (${workspace?.name || ''})`);
        setWorkspaceEmail(defaultProfile?.email || '');
      }
    };

    loadWorkspaceProfile();
  }, [selectedWorkspaceId, activeTab, token, workspaces, defaultProfile]);

  // ========================================
  // 이미지 업로드 핸들러
  // ========================================

  const handleAvatarChangeClick = () => {
    fileInputRef.current?.click();
  };

  const handleFileChange = (event: ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (file) {
      if (avatarPreviewUrl) {
        URL.revokeObjectURL(avatarPreviewUrl);
      }
      setAvatarPreviewUrl(URL.createObjectURL(file));
      console.log(`[File] 새 프로필 사진 선택: ${file.name}`);
    }
  };

  // ========================================
  // 워크스페이스 변경 핸들러
  // ========================================

  const handleWorkspaceChange = (workspaceId: string) => {
    setSelectedWorkspaceId(workspaceId);
    setAvatarPreviewUrl(null); // 워크스페이스 변경 시 미리보기 초기화
  };

  // ========================================
  // 저장 핸들러
  // ========================================

  const handleSave = async () => {
    if (!token) {
      setError('인증 토큰이 없습니다. 다시 로그인해주세요.');
      return;
    }

    try {
      setLoading(true);
      setError(null);

      if (activeTab === 'default') {
        // 기본 프로필 저장
        const updatedProfile = await updateMyProfile(
          {
            name: defaultName,
            email: defaultEmail || undefined,
            profileImageUrl: avatarPreviewUrl || undefined,
          },
          token,
        );

        setDefaultProfile(updatedProfile);
        alert('기본 프로필이 저장되었습니다.');
      } else {
        // 워크스페이스 프로필 저장
        const updatedProfile = await updateWorkspaceProfile(
          selectedWorkspaceId,
          {
            name: workspaceName,
            email: workspaceEmail || undefined,
            profileImageUrl: avatarPreviewUrl || undefined,
          },
          token,
        );

        setWorkspaceProfile(updatedProfile);
        const workspaceName_display = workspaces.find((ws) => ws.id === selectedWorkspaceId)?.name;
        alert(`${workspaceName_display} 프로필이 저장되었습니다.`);
      }

      setAvatarPreviewUrl(null);
    } catch (err) {
      console.error('[Profile Save Error]', err);
      setError('프로필 저장에 실패했습니다.');
    } finally {
      setLoading(false);
    }
  };

  // ========================================
  // 모달 닫기 핸들러
  // ========================================

  const handleClose = () => {
    if (avatarPreviewUrl) {
      URL.revokeObjectURL(avatarPreviewUrl);
    }
    onClose();
  };

  // ========================================
  // 현재 활성 탭의 프로필 정보 가져오기
  // ========================================

  const currentProfile = activeTab === 'default' ? defaultProfile : (workspaceProfile || defaultProfile);
  const currentName = activeTab === 'default' ? defaultName : workspaceName;
  const currentEmail = activeTab === 'default' ? defaultEmail : workspaceEmail;
  const setCurrentName = activeTab === 'default' ? setDefaultName : setWorkspaceName;
  const setCurrentEmail = activeTab === 'default' ? setDefaultEmail : setWorkspaceEmail;

  // ========================================
  // 렌더링
  // ========================================

  if (!defaultProfile) {
    return (
      <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
        <div className="bg-white p-8 rounded-xl shadow-lg">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500 mx-auto mb-4"></div>
          <p className="text-gray-700">프로필 정보를 불러오는 중...</p>
        </div>
      </div>
    );
  }

  return (
    <div
      className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50"
      onClick={handleClose}
    >
      <div className="relative w-full max-w-md" onClick={(e) => e.stopPropagation()}>
        <div
          className={`relative ${theme.colors.card} ${theme.effects.borderWidth} ${theme.colors.border} ${theme.effects.borderRadius} shadow-xl`}
        >
          {/* 헤더 */}
          <div className="flex items-center justify-between p-4 pb-3">
            <h2 className={`${theme.font.size.base} font-bold text-gray-800`}>사용자 프로필 설정</h2>
            <button
              onClick={handleClose}
              className="p-2 hover:bg-gray-100 rounded-lg transition"
              title="닫기"
            >
              <X className="w-4 h-4 text-gray-600" />
            </button>
          </div>

          {/* 탭 메뉴 */}
          <div className="flex border-b border-gray-200 px-6">
            <button
              onClick={() => setActiveTab('default')}
              className={`flex-1 py-3 text-sm font-medium transition-colors relative ${
                activeTab === 'default' ? 'text-blue-600' : 'text-gray-500 hover:text-gray-700'
              }`}
            >
              기본 프로필
              {activeTab === 'default' && (
                <div className="absolute bottom-0 left-0 right-0 h-0.5 bg-blue-600"></div>
              )}
            </button>
            <button
              onClick={() => setActiveTab('workspace')}
              className={`flex-1 py-3 text-sm font-medium transition-colors relative ${
                activeTab === 'workspace' ? 'text-blue-600' : 'text-gray-500 hover:text-gray-700'
              }`}
            >
              워크스페이스별 프로필
              {activeTab === 'workspace' && (
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

            {/* 워크스페이스 선택 - 탭 전환 시 높이 유지를 위해 항상 렌더링 */}
            <div className={activeTab === 'default' ? 'invisible pointer-events-none' : ''}>
              <label className={`block ${theme.font.size.xs} mb-2 text-gray-500 font-medium`}>
                워크스페이스 선택:
              </label>
              <select
                value={selectedWorkspaceId}
                onChange={(e) => handleWorkspaceChange(e.target.value)}
                className={`w-full px-3 py-2 ${theme.effects.cardBorderWidth} ${theme.colors.border} ${theme.colors.card} ${theme.font.size.xs} ${theme.effects.borderRadius} focus:outline-none focus:ring-2 focus:ring-blue-500`}
              >
                {workspaces.map((workspace) => (
                  <option key={workspace.id} value={workspace.id}>
                    {workspace.name}
                  </option>
                ))}
              </select>
              <p className="mt-1 text-xs text-gray-500">
                워크스페이스마다 다른 프로필을 설정할 수 있습니다
              </p>
            </div>

            {/* 프로필 이미지 */}
            <div className="flex flex-col items-center mb-4">
              <div className="relative">
                {avatarPreviewUrl ? (
                  <img
                    src={avatarPreviewUrl}
                    alt="프로필 미리보기"
                    className="w-24 h-24 object-cover border-2 border-gray-300 rounded-full"
                  />
                ) : currentProfile?.profileImageUrl ? (
                  <img
                    src={currentProfile.profileImageUrl}
                    alt="프로필 이미지"
                    className="w-24 h-24 object-cover border-2 border-gray-300 rounded-full"
                  />
                ) : (
                  <div className="w-24 h-24 bg-blue-500 border-2 border-gray-300 flex items-center justify-center text-white text-3xl font-bold rounded-full">
                    {currentName[0] || 'U'}
                  </div>
                )}

                <input
                  type="file"
                  ref={fileInputRef}
                  onChange={handleFileChange}
                  accept="image/*"
                  className="hidden"
                />

                <button
                  onClick={handleAvatarChangeClick}
                  className="absolute bottom-0 right-0 p-2 bg-gray-700 hover:bg-gray-800 text-white rounded-full transition shadow-md"
                  title="프로필 사진 변경"
                >
                  <Camera className="w-4 h-4" />
                </button>
              </div>
            </div>

            {/* 사용자 ID */}
            <div>
              <label className={`block ${theme.font.size.xs} mb-2 text-gray-500 font-medium`}>
                사용자 ID:
              </label>
              <input
                type="text"
                readOnly
                disabled
                value={currentProfile?.userId || ''}
                className="w-full px-3 py-2 border border-gray-300 text-gray-700 text-xs rounded-md disabled:bg-gray-100 disabled:text-gray-500 disabled:cursor-not-allowed"
              />
            </div>

            {/* 이름 */}
            <div>
              <label className={`block ${theme.font.size.xs} mb-2 text-gray-500 font-medium`}>
                이름:
              </label>
              <input
                type="text"
                value={currentName}
                onChange={(e) => setCurrentName(e.target.value)}
                className={`w-full px-3 py-2 ${theme.effects.cardBorderWidth} ${theme.colors.border} ${theme.colors.card} ${theme.font.size.xs} ${theme.effects.borderRadius} focus:outline-none focus:ring-2 focus:ring-blue-500`}
                placeholder="이름을 입력하세요"
              />
            </div>

            {/* 이메일 */}
            <div>
              <label className={`block ${theme.font.size.xs} mb-2 text-gray-500 font-medium`}>
                이메일:
              </label>
              <input
                type="email"
                value={currentEmail}
                onChange={(e) => setCurrentEmail(e.target.value)}
                className={`w-full px-3 py-2 ${theme.effects.cardBorderWidth} ${theme.colors.border} ${theme.colors.card} ${theme.font.size.xs} ${theme.effects.borderRadius} focus:outline-none focus:ring-2 focus:ring-blue-500`}
                placeholder="이메일을 입력하세요"
              />
            </div>

            {/* 버튼 영역 */}
            <div className="flex gap-2 pt-4">
              <button
                onClick={handleSave}
                disabled={loading}
                className={`flex-1 ${theme.colors.primary} text-white py-3 ${theme.effects.borderRadius} font-semibold transition ${
                  loading ? 'opacity-50 cursor-not-allowed' : 'hover:opacity-90'
                }`}
              >
                {loading ? '저장 중...' : '저장'}
              </button>
              <button
                onClick={handleClose}
                className="flex-1 bg-gray-300 text-gray-800 py-3 rounded-lg font-semibold hover:bg-gray-400 transition"
              >
                취소
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default UserProfileModal;
