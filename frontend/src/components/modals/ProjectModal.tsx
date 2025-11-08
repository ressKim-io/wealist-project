import React, { useState, useEffect } from 'react';
import { X } from 'lucide-react';
import { useTheme } from '../../contexts/ThemeContext';

/**
 * ProjectModal - 프로젝트 생성 및 편집을 위한 통합 모달
 * - project prop이 있으면 편집 모드, 없으면 생성 모드
 */
interface ProjectData {
  id: string;
  name: string;
  description?: string;
  workspaceId: string;
  ownerId: string;
  ownerName: string;
  ownerEmail: string;
  createdAt: string;
  updatedAt: string;
}

interface ProjectModalProps {
  workspaceId: string;
  project?: ProjectData; // 편집 모드일 때만 전달
  onClose: () => void;
  onProjectSaved: () => void; // 생성 또는 수정 후 호출
}

export const ProjectModal: React.FC<ProjectModalProps> = ({
  workspaceId,
  project,
  onClose,
  onProjectSaved,
}) => {
  const { theme } = useTheme();
  const isEditMode = !!project;

  const [name, setName] = useState(project?.name || '');
  const [description, setDescription] = useState(project?.description || '');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // project prop이 변경되면 폼 리셋
  useEffect(() => {
    if (project) {
      setName(project.name);
      setDescription(project.description || '');
    } else {
      setName('');
      setDescription('');
    }
    setError(null);
  }, [project]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!name.trim()) {
      setError('프로젝트 이름은 필수입니다.');
      return;
    }

    setIsLoading(true);
    setError(null);

    try {
      const accessToken = localStorage.getItem('access_token') || '';

      if (isEditMode) {
        // 편집 모드
        const { updateProject } = await import('../../api/board/boardService');
        await updateProject(
          project.id,
          {
            name: name.trim(),
            description: description.trim() || undefined,
          },
          accessToken,
        );
        console.log('✅ 프로젝트 수정 성공:', name);
      } else {
        // 생성 모드
        const { createProject } = await import('../../api/board/boardService');
        await createProject(
          {
            workspaceId,
            name: name.trim(),
            description: description.trim() || undefined,
          },
          accessToken,
        );
        console.log('✅ 프로젝트 생성 성공:', name);
      }

      onProjectSaved();
      onClose();
    } catch (err) {
      const error = err as Error;
      console.error(
        isEditMode ? '❌ 프로젝트 수정 실패:' : '❌ 프로젝트 생성 실패:',
        error,
      );
      setError(
        error.message ||
          (isEditMode ? '프로젝트 수정에 실패했습니다.' : '프로젝트 생성에 실패했습니다.'),
      );
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div
      className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-[90]"
      onClick={onClose}
    >
      <div
        className={`relative w-full max-w-md ${theme.colors.card} p-6 ${theme.effects.borderRadius} shadow-xl`}
        onClick={(e) => e.stopPropagation()}
      >
        {/* Header */}
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-xl font-bold text-gray-800">
            {isEditMode ? '프로젝트 설정' : '새 프로젝트 만들기'}
          </h2>
          <button
            onClick={onClose}
            className="p-2 rounded-full hover:bg-gray-100 text-gray-500 hover:text-gray-700 transition"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Project Owner Info (편집 모드일 때만 표시) */}
        {isEditMode && project && (
          <div className="mb-4 p-3 bg-gray-50 rounded-lg">
            <div className="text-xs text-gray-500 mb-1">프로젝트 소유자</div>
            <div className="text-sm font-medium text-gray-700">
              {project.ownerName} ({project.ownerEmail})
            </div>
          </div>
        )}

        {/* Error Message */}
        {error && (
          <div className="mb-4 p-3 bg-red-50 border border-red-300 rounded-lg text-red-700 text-sm">
            {error}
          </div>
        )}

        {/* Form */}
        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-sm font-semibold text-gray-700 mb-2">
              프로젝트 이름 <span className="text-red-500">*</span>
            </label>
            <input
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="예: Wealist 서비스 개발"
              className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm"
              disabled={isLoading}
              maxLength={100}
              autoFocus
            />
          </div>

          <div>
            <label className="block text-sm font-semibold text-gray-700 mb-2">
              프로젝트 설명 (선택)
            </label>
            <textarea
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="프로젝트에 대한 간단한 설명을 입력하세요"
              className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm resize-none"
              rows={3}
              disabled={isLoading}
              maxLength={500}
            />
          </div>

          {/* Timestamps (편집 모드일 때만 표시) */}
          {isEditMode && project && (
            <div className="text-xs text-gray-500 space-y-1">
              <div>생성일: {new Date(project.createdAt).toLocaleString('ko-KR')}</div>
              <div>수정일: {new Date(project.updatedAt).toLocaleString('ko-KR')}</div>
            </div>
          )}

          {/* Actions */}
          <div className="flex gap-3 pt-2">
            <button
              type="button"
              onClick={onClose}
              className="flex-1 px-4 py-2 border border-gray-300 text-gray-700 font-semibold rounded-lg hover:bg-gray-50 transition"
              disabled={isLoading}
            >
              취소
            </button>
            <button
              type="submit"
              className={`flex-1 px-4 py-2 bg-blue-500 text-white font-semibold rounded-lg hover:bg-blue-600 transition ${
                isLoading ? 'opacity-50 cursor-not-allowed' : ''
              }`}
              disabled={isLoading}
            >
              {isLoading
                ? isEditMode
                  ? '저장 중...'
                  : '생성 중...'
                : isEditMode
                  ? '저장'
                  : '프로젝트 만들기'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};
