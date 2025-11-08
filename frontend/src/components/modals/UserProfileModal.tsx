import React, { useState, useRef, ChangeEvent, useEffect } from 'react';
import { X, Camera, MessageSquare } from 'lucide-react';
import { useTheme } from '../../contexts/ThemeContext';
import { UserProfile } from '../../types';
import { getMyProfile, updateMyProfile } from '../../api/user/userService';

interface UserProfileModalProps {
  user: UserProfile;
  onClose: () => void;
}

const UserProfileModal: React.FC<UserProfileModalProps> = ({ user, onClose }) => {
  const { theme } = useTheme();

  // Ref for file input
  const fileInputRef = useRef<HTMLInputElement>(null);

  const isGoogleConnected = true;

  // 상태 관리
  const [profile, setProfile] = useState<UserProfile>(user);
  const [name, setName] = useState(user.name);
  const [email, setEmail] = useState(user.email || '');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  // 💡 프로필 이미지 미리보기 URL 상태
  const [avatarPreviewUrl, setAvatarPreviewUrl] = useState<string | null>(null);

  // 💡 모달이 열릴 때 최신 프로필 데이터 fetch
  useEffect(() => {
    const fetchProfile = async () => {
      try {
        setLoading(true);
        setError(null);
        const token = localStorage.getItem('accessToken');
        if (!token) {
          setError('인증 토큰이 없습니다. 다시 로그인해주세요.');
          return;
        }
        const profileData = await getMyProfile(token);
        setProfile(profileData);
        setName(profileData.name);
        setEmail(profileData.email || '');
      } catch (err) {
        console.error('[Profile Fetch Error]', err);
        setError('프로필 정보를 불러오는데 실패했습니다.');
      } finally {
        setLoading(false);
      }
    };

    fetchProfile();
  }, []);

  // --- 프로필 사진 변경 로직 ---

  // 💡 파일 입력 필드 클릭 트리거
  const handleAvatarChangeClick = () => {
    fileInputRef.current?.click();
  };

  // 💡 파일 선택 시 미리보기 처리
  const handleFileChange = (event: ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (file) {
      // 이전에 생성된 URL이 있다면 해제 (메모리 누수 방지)
      if (avatarPreviewUrl) {
        URL.revokeObjectURL(avatarPreviewUrl);
      }
      // 새 파일의 URL 생성 및 상태 업데이트
      setAvatarPreviewUrl(URL.createObjectURL(file));
      console.log(`[File] 새로운 프로필 사진 파일 선택됨: ${file.name}`);
      // TODO: 파일 객체를 상태에 저장하거나, base64로 변환하여 서버에 업로드할 준비를 해야 합니다.
    }
  };

  // --- DM 및 저장 로직 ---

  // DM 버튼 클릭 핸들러 (Mock)
  const handleDmClick = () => {
    console.log(`[DM] ${profile.name} 님에게 DM 보내기 요청`);
    // TODO: 실제 DM 기능(채팅 모듈) 구현 필요
    onClose();
  };

  // 저장 버튼 클릭 핸들러
  const handleSave = async () => {
    try {
      setLoading(true);
      setError(null);
      const token = localStorage.getItem('accessToken');
      if (!token) {
        setError('인증 토큰이 없습니다. 다시 로그인해주세요.');
        return;
      }

      console.log(`[저장] 사용자 정보 업데이트: ${name}, ${email}`);

      // 프로필 업데이트 API 호출
      const updatedProfile = await updateMyProfile(
        {
          name,
          email: email || undefined,
          profileImageUrl: avatarPreviewUrl || undefined,
        },
        token,
      );

      console.log('[저장 성공]', updatedProfile);
      setProfile(updatedProfile);

      if (avatarPreviewUrl) {
        console.log(`[저장] 프로필 사진이 업데이트되었습니다.`);
        // TODO: 실제 파일 업로드가 필요한 경우, 파일 업로드 API 호출 추가
      }

      onClose();
    } catch (err) {
      console.error('[Profile Update Error]', err);
      setError('프로필 업데이트에 실패했습니다.');
    } finally {
      setLoading(false);
    }
  };

  // --- 모달 닫힐 때 정리 ---

  // 💡 모달이 닫힐 때 객체 URL을 해제하여 메모리 정리
  const handleClose = () => {
    if (avatarPreviewUrl) {
      URL.revokeObjectURL(avatarPreviewUrl);
    }
    onClose();
  };

  return (
    <div
      className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50"
      onClick={handleClose} // 💡 닫기 핸들러 변경
    >
      <div className="relative w-full max-w-md" onClick={(e) => e.stopPropagation()}>
        <div
          className={`relative ${theme.colors.card} ${theme.effects.borderWidth} ${theme.colors.border} p-6 ${theme.effects.borderRadius} shadow-xl`}
        >
          <div
            className={`flex items-center justify-between mb-6 pb-4 ${theme.effects.borderWidth} ${theme.colors.border} border-t-0 border-l-0 border-r-0`}
          >
            <h2 className={`${theme.font.size.base} font-bold text-gray-800`}>사용자 정보 수정</h2>
            <button
              onClick={handleClose} // 💡 닫기 핸들러 변경
              className={`bg-red-500 ${theme.effects.cardBorderWidth} ${theme.colors.border} p-2 hover:bg-red-600 ${theme.effects.borderRadius} transition`}
            >
              <X className="w-4 h-4 text-white" />
            </button>
          </div>

          <div className="space-y-5">
            {/* 프로필 이미지 및 변경 버튼 */}
            <div className="flex flex-col items-center mb-6">
              <div className="relative">
                {/* 💡 이미지 미리보기 또는 이니셜 아바타 표시 */}
                {avatarPreviewUrl ? (
                  <img
                    src={avatarPreviewUrl}
                    alt="프로필 미리보기"
                    className={`w-28 h-28 object-cover ${theme.effects.borderWidth} ${theme.colors.border} rounded-full`}
                  />
                ) : profile.profileImageUrl ? (
                  <img
                    src={profile.profileImageUrl}
                    alt="프로필 이미지"
                    className={`w-28 h-28 object-cover ${theme.effects.borderWidth} ${theme.colors.border} rounded-full`}
                  />
                ) : (
                  <div
                    className={`w-28 h-28 ${theme.colors.primary} ${theme.effects.borderWidth} ${theme.colors.border} flex items-center justify-center text-white text-4xl font-bold ${theme.effects.borderRadius} rounded-full`}
                  >
                    {profile.name[0]}
                  </div>
                )}

                {/* 💡 숨겨진 파일 입력 필드 */}
                <input
                  type="file"
                  ref={fileInputRef}
                  onChange={handleFileChange}
                  accept="image/*"
                  className="hidden"
                />

                {/* 프로필 사진 변경 버튼 (클릭 트리거) */}
                <button
                  onClick={handleAvatarChangeClick} // 💡 트리거 연결
                  className="absolute bottom-0 right-0 p-2 bg-gray-700 hover:bg-gray-800 text-white rounded-full transition shadow-md"
                  title="프로필 사진 변경"
                >
                  <Camera className="w-4 h-4" />
                </button>
              </div>

              {/* DM 보내기 버튼 */}
              <button
                onClick={handleDmClick}
                className={`mt-3 flex items-center gap-2 px-4 py-2 bg-blue-500 text-white text-sm font-semibold rounded-lg hover:bg-blue-600 transition shadow-md`}
              >
                <MessageSquare className="w-4 h-4" />
                DM 보내기
              </button>
            </div>

            {/* 에러 메시지 표시 */}
            {error && (
              <div className="p-3 bg-red-100 border border-red-400 text-red-700 rounded-md text-sm">
                {error}
              </div>
            )}

            {/* 로딩 표시 */}
            {loading && (
              <div className="p-3 bg-blue-100 border border-blue-400 text-blue-700 rounded-md text-sm">
                처리 중...
              </div>
            )}

            {/* 사용자 ID (읽기 전용) */}
            <div>
              <label className={`block ${theme.font.size.xs} mb-2 text-gray-500 font-medium`}>
                사용자 ID (고유 식별자):
              </label>
              <input
                type="text"
                readOnly
                disabled
                value={profile.userId}
                className={`w-full px-3 py-2 border border-gray-300 text-gray-700 text-xs rounded-md
              disabled:bg-gray-100 disabled:text-gray-500 disabled:cursor-not-allowed focus:outline-none`}
              />
            </div>

            {/* 이름 수정 */}
            <div>
              <label className={`block ${theme.font.size.xs} mb-2 text-gray-500 font-medium`}>
                이름:
              </label>
              <input
                type="text"
                value={name}
                onChange={(e) => setName(e.target.value)}
                className={`w-full px-3 py-2 ${theme.effects.cardBorderWidth} ${theme.colors.border} ${theme.colors.card} ${theme.font.size.xs} ${theme.effects.borderRadius} focus:outline-none focus:ring-2 focus:ring-blue-500`}
              />
            </div>

            {/* 이메일 (Google 연동 표시) */}
            <div>
              <label className={`block ${theme.font.size.xs} mb-2 text-gray-500 font-medium`}>
                이메일:
              </label>
              <div className="relative">
                <input
                  disabled
                  type="email"
                  value={email}
                  readOnly
                  onChange={(e) => setEmail(e.target.value)}
                  className={`w-full px-3 py-2 border border-gray-300 text-gray-700 text-xs rounded-md 
              disabled:bg-gray-100 disabled:text-gray-500 disabled:cursor-not-allowed focus:outline-none`} // 💡 read-only 대신 disabled 사용 및 클래스 수정
                />
                {isGoogleConnected && (
                  <span className="absolute right-2 top-1/2 transform -translate-y-1/2 flex items-center text-xs text-green-600 font-bold bg-green-100 px-2 py-1 rounded-full">
                    Google 연동
                  </span>
                )}
              </div>
            </div>

            {/* 버튼 영역 */}
            <div className="flex gap-2 pt-4">
              <button
                onClick={handleSave}
                className={`flex-1 ${theme.colors.primary} text-white py-3 ${theme.effects.cardBorderWidth} ${theme.colors.border} ${theme.colors.primaryHover} transition ${theme.font.size.xs} ${theme.effects.borderRadius} font-semibold`}
              >
                저장
              </button>
              <button
                onClick={handleClose} // 💡 닫기 핸들러 변경
                className={`flex-1 bg-gray-300 py-3 ${theme.effects.cardBorderWidth} ${theme.colors.border} text-gray-800 hover:bg-gray-400 transition ${theme.font.size.xs} ${theme.effects.borderRadius} font-semibold`}
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
