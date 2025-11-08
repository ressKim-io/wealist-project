/**
 * 사용자 프로필 모달 컴포넌트 (UI 목업 버전)
 *
 * [백엔드 개발자 참고사항]
 *
 * 1. 프로필 구조:
 *    - 기본 프로필: userId 단일 키로 관리 (workspaceId = null)
 *    - 워크스페이스별 프로필: userId + workspaceId 복합 키로 관리
 *
 * 2. API 엔드포인트 (예상):
 *    - GET  /api/profiles/me                    : 기본 프로필 조회
 *    - PUT  /api/profiles/me                    : 기본 프로필 업데이트
 *    - GET  /api/profiles/workspace/{workspaceId} : 특정 워크스페이스 프로필 조회
 *    - PUT  /api/profiles/workspace/{workspaceId} : 워크스페이스 프로필 생성/업데이트
 *
 * 3. UserProfile DTO 구조:
 *    {
 *      profileId: string (UUID)
 *      userId: string (UUID)
 *      workspaceId?: string | null (UUID, 기본 프로필은 null)
 *      name: string
 *      email: string | null
 *      profileImageUrl: string | null
 *      createdAt: string (ISO-8601)
 *      updatedAt: string (ISO-8601)
 *    }
 *
 * 4. 저장 로직:
 *    - 기본 프로필: workspaceId 없이 저장
 *    - 워크스페이스 프로필: 선택된 workspaceId와 함께 저장
 *    - 같은 userId + workspaceId 조합이 있으면 UPDATE, 없으면 INSERT
 */

import React, { useState, useRef, ChangeEvent } from 'react';
import { X, Camera } from 'lucide-react';
import { useTheme } from '../../contexts/ThemeContext';
import { UserProfile } from '../../types';

interface UserProfileModalProps {
  user: UserProfile;
  onClose: () => void;
}

// ========================================
// 목업 데이터 (백엔드 연동 전까지 사용)
// ========================================

// 워크스페이스 목업 데이터
const MOCK_WORKSPACES = [
  { id: 'workspace-1', name: '오렌지클라우드' },
  { id: 'workspace-2', name: '데이터랩' },
  { id: 'workspace-3', name: '마케팅팀' },
];

// 사용자 프로필 목업 데이터
const MOCK_USER_PROFILE: UserProfile = {
  profileId: 'profile-default-001',
  userId: 'user-123',
  name: '김개발',
  email: 'dev.kim@example.com',
  profileImageUrl: null,
  createdAt: '2024-01-01T00:00:00Z',
  updatedAt: '2024-01-01T00:00:00Z',
};

// 워크스페이스별 프로필 목업 데이터 (userId + workspaceId가 키)
const MOCK_WORKSPACE_PROFILES: Record<string, UserProfile> = {
  'workspace-1': {
    profileId: 'profile-ws-001',
    userId: 'user-123',
    name: '김개발 (오렌지클라우드)',
    email: 'dev.kim@orangecloud.com',
    profileImageUrl: null,
    createdAt: '2024-01-02T00:00:00Z',
    updatedAt: '2024-01-02T00:00:00Z',
  },
};

const UserProfileModal: React.FC<UserProfileModalProps> = ({ user, onClose }) => {
  const { theme } = useTheme();

  // ========================================
  // 상태 관리
  // ========================================

  // 탭 상태: 'default' (기본 프로필) | 'workspace' (워크스페이스별 프로필)
  const [activeTab, setActiveTab] = useState<'default' | 'workspace'>('default');

  // 선택된 워크스페이스 ID
  const [selectedWorkspaceId, setSelectedWorkspaceId] = useState<string>(MOCK_WORKSPACES[0].id);

  // 파일 입력 Ref
  const fileInputRef = useRef<HTMLInputElement>(null);

  // 기본 프로필 상태
  const [defaultProfile, setDefaultProfile] = useState<UserProfile>(MOCK_USER_PROFILE);
  const [defaultName, setDefaultName] = useState(MOCK_USER_PROFILE.name);
  const [defaultEmail, setDefaultEmail] = useState(MOCK_USER_PROFILE.email || '');

  // 워크스페이스 프로필 상태 (선택된 워크스페이스에 따라 동적으로 변경)
  const [workspaceProfiles, setWorkspaceProfiles] = useState(MOCK_WORKSPACE_PROFILES);
  const [workspaceName, setWorkspaceName] = useState('');
  const [workspaceEmail, setWorkspaceEmail] = useState('');

  // 프로필 이미지 미리보기 URL
  const [avatarPreviewUrl, setAvatarPreviewUrl] = useState<string | null>(null);

  const [loading, setLoading] = useState(false);

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

    // 해당 워크스페이스의 프로필이 있으면 불러오기, 없으면 기본값 설정
    const wsProfile = workspaceProfiles[workspaceId];
    if (wsProfile) {
      setWorkspaceName(wsProfile.name);
      setWorkspaceEmail(wsProfile.email || '');
    } else {
      // 프로필이 없으면 기본 프로필 정보로 초기화
      const workspace = MOCK_WORKSPACES.find((ws) => ws.id === workspaceId);
      setWorkspaceName(`${defaultProfile.name} (${workspace?.name || ''})`);
      setWorkspaceEmail(defaultProfile.email || '');
    }
  };

  // ========================================
  // 저장 핸들러
  // ========================================

  /**
   * [백엔드 개발자 참고]
   *
   * 기본 프로필 저장:
   * - PUT /api/profiles/me
   * - Body: { name, email, profileImageUrl }
   * - workspaceId는 전송하지 않음 (또는 null)
   */
  const handleSaveDefaultProfile = () => {
    console.log('[목업] 기본 프로필 저장:', {
      userId: defaultProfile.userId,
      workspaceId: null,
      name: defaultName,
      email: defaultEmail,
      profileImageUrl: avatarPreviewUrl,
    });

    setDefaultProfile({
      ...defaultProfile,
      name: defaultName,
      email: defaultEmail,
      profileImageUrl: avatarPreviewUrl,
    });

    alert('기본 프로필이 저장되었습니다 (목업)');
  };

  /**
   * [백엔드 개발자 참고]
   *
   * 워크스페이스 프로필 저장:
   * - PUT /api/profiles/workspace/{workspaceId}
   * - Body: { name, email, profileImageUrl }
   * - 같은 userId + workspaceId가 있으면 UPDATE, 없으면 INSERT
   */
  const handleSaveWorkspaceProfile = () => {
    const newProfile: UserProfile = {
      profileId: `profile-ws-${Date.now()}`, // 실제로는 백엔드에서 생성
      userId: defaultProfile.userId,
      name: workspaceName,
      email: workspaceEmail,
      profileImageUrl: avatarPreviewUrl,
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    };

    setWorkspaceProfiles({
      ...workspaceProfiles,
      [selectedWorkspaceId]: newProfile,
    });

    console.log('[목업] 워크스페이스 프로필 저장:', {
      userId: defaultProfile.userId,
      workspaceId: selectedWorkspaceId,
      name: workspaceName,
      email: workspaceEmail,
      profileImageUrl: avatarPreviewUrl,
    });

    alert(`${MOCK_WORKSPACES.find((ws) => ws.id === selectedWorkspaceId)?.name} 프로필이 저장되었습니다 (목업)`);
  };

  const handleSave = () => {
    setLoading(true);
    setTimeout(() => {
      if (activeTab === 'default') {
        handleSaveDefaultProfile();
      } else {
        handleSaveWorkspaceProfile();
      }
      setLoading(false);
    }, 500); // 목업 딜레이
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

  const currentProfile =
    activeTab === 'default' ? defaultProfile : workspaceProfiles[selectedWorkspaceId] || defaultProfile;

  const currentName = activeTab === 'default' ? defaultName : workspaceName;
  const currentEmail = activeTab === 'default' ? defaultEmail : workspaceEmail;

  const setCurrentName = activeTab === 'default' ? setDefaultName : setWorkspaceName;
  const setCurrentEmail = activeTab === 'default' ? setDefaultEmail : setWorkspaceEmail;

  // ========================================
  // 렌더링
  // ========================================

  return (
    <div
      className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50"
      onClick={handleClose}
    >
      <div className="relative w-full max-w-md" onClick={(e) => e.stopPropagation()}>
        <div
          className={`relative ${theme.colors.card} ${theme.effects.borderWidth} ${theme.colors.border} ${theme.effects.borderRadius} shadow-xl`}
        >
          {/* ========================================
              헤더: 제목 + 닫기 버튼
              ======================================== */}
          <div
            className={`flex items-center justify-between p-6 pb-4 ${theme.effects.borderWidth} ${theme.colors.border} border-t-0 border-l-0 border-r-0`}
          >
            <h2 className={`${theme.font.size.base} font-bold text-gray-800`}>사용자 프로필 설정</h2>
            <button
              onClick={handleClose}
              className="bg-red-500 p-2 hover:bg-red-600 rounded-lg transition"
              title="닫기"
            >
              <X className="w-4 h-4 text-white" />
            </button>
          </div>

          {/* ========================================
              탭 메뉴: 기본 프로필 / 워크스페이스별 프로필
              ======================================== */}
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

          {/* ========================================
              탭 컨텐츠
              ======================================== */}
          <div className="p-6 space-y-5">
            {/* 워크스페이스 선택 (워크스페이스별 프로필 탭에서만 표시) */}
            {activeTab === 'workspace' && (
              <div>
                <label className={`block ${theme.font.size.xs} mb-2 text-gray-500 font-medium`}>
                  워크스페이스 선택:
                </label>
                <select
                  value={selectedWorkspaceId}
                  onChange={(e) => handleWorkspaceChange(e.target.value)}
                  className={`w-full px-3 py-2 ${theme.effects.cardBorderWidth} ${theme.colors.border} ${theme.colors.card} ${theme.font.size.xs} ${theme.effects.borderRadius} focus:outline-none focus:ring-2 focus:ring-blue-500`}
                >
                  {MOCK_WORKSPACES.map((workspace) => (
                    <option key={workspace.id} value={workspace.id}>
                      {workspace.name}
                    </option>
                  ))}
                </select>
                <p className="mt-1 text-xs text-gray-500">
                  워크스페이스마다 다른 프로필을 설정할 수 있습니다
                </p>
              </div>
            )}

            {/* 프로필 이미지 */}
            <div className="flex flex-col items-center mb-4">
              <div className="relative">
                {avatarPreviewUrl ? (
                  <img
                    src={avatarPreviewUrl}
                    alt="프로필 미리보기"
                    className="w-24 h-24 object-cover border-2 border-gray-300 rounded-full"
                  />
                ) : currentProfile.profileImageUrl ? (
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

            {/* 사용자 ID (읽기 전용) */}
            <div>
              <label className={`block ${theme.font.size.xs} mb-2 text-gray-500 font-medium`}>
                사용자 ID:
              </label>
              <input
                type="text"
                readOnly
                disabled
                value={currentProfile.userId}
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
