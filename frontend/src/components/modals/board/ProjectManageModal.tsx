// src/components/modals/board/ProjectManageModal.tsx

import React, { useState, useEffect, useMemo, useCallback } from 'react';
import {
  X,
  Paperclip,
  Download,
  Edit2,
  BarChart3,
  Lock,
  Loader2,
  User as UserIcon,
} from 'lucide-react';
import { useTheme } from '../../../contexts/ThemeContext';
import { createProject, updateProject, getBoardsByProject } from '../../../api/board/boardService';
import {
  ProjectResponse,
  BoardResponse,
  CreateProjectRequest,
  UpdateProjectRequest,
} from '../../../types/board';
import { IROLES } from '../../../types/common';
import Portal from '../../common/Portal';
import { WorkspaceMemberResponse } from '../../../types/user';

// 💡 [추가] 파일 업로드 관련 Import (경로를 실제 구조에 맞게 확인해주세요)
import { useFileUpload } from '../../../hooks/useFileUpload';
import { FileUploader } from '../../common/FileUploader';

/**
 * 모달 모드 타입 정의
 */
type ProjectModalMode = 'create' | 'detail' | 'edit';

interface ProjectManageModalProps {
  workspaceId: string;
  project?: ProjectResponse;
  onClose: () => void;
  onProjectSaved: () => void;
  onProjectCreated?: (createObj: ProjectResponse) => void;
  userRole: IROLES;
  initialMode: ProjectModalMode;
  members?: WorkspaceMemberResponse[] | undefined;
}

// 파일 다운로드 핸들러 (Detail 모드용)
const handleFileDownload = (fileUrl: string, fileName: string) => {
  if (!fileUrl) return;

  // ✅ fileUrl은 이미 full URL이므로 바로 사용 가능
  const link = document.createElement('a');
  link.href = fileUrl;
  link.setAttribute('download', fileName);
  link.setAttribute('target', '_blank'); // 새 탭에서 열기
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
};

export const ProjectManageModal: React.FC<ProjectManageModalProps> = ({
  workspaceId,
  project,
  onClose,
  onProjectSaved,
  onProjectCreated,
  userRole,
  initialMode = 'create',
  members = [],
}) => {
  const { theme } = useTheme();
  const isExistingProject = !!project;
  const [mode, setMode] = useState<ProjectModalMode>(isExistingProject ? initialMode : 'create');

  // Form state
  const [name, setName] = useState(project?.name || '');
  const [description, setDescription] = useState(project?.description || '');
  const [startDate, setStartDate] = useState(
    project?.startDate ? project.startDate.substring(0, 10) : '',
  );
  const [dueDate, setDueDate] = useState(project?.dueDate ? project.dueDate.substring(0, 10) : '');

  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const [boards, setBoards] = useState<BoardResponse[]>([]);
  const [isBoardsLoading, setIsBoardsLoading] = useState(false);
  const [projectMembers, setProjectMembers] = useState<WorkspaceMemberResponse[]>();

  // 💡 파일 업로드 훅 사용
  const {
    selectedFile,
    previewUrl,
    handleFileSelect,
    handleRemoveFile,
    upload,
    setInitialFile,
    attachmentId,
    setAttachmentId,
  } = useFileUpload();

  // ✅ 수정: attachmentIds 배열 관리 (백엔드로 전송용)
  const [currentAttachmentIds, setCurrentAttachmentIds] = useState<string[]>([]);

  const canEdit = useMemo(() => {
    return (
      isExistingProject &&
      (userRole === 'OWNER' || userRole === 'ADMIN' || userRole === 'ORGANIZER')
    );
  }, [isExistingProject, userRole]);

  // 💡 [수정] useEffect: 프로젝트 데이터 로드 및 파일 상태 초기화
  useEffect(() => {
    if (project) {
      setName(project.name);
      setDescription(project.description || '');
      setStartDate(project.startDate ? project.startDate.substring(0, 10) : '');
      setDueDate(project.dueDate ? project.dueDate.substring(0, 10) : '');
      setProjectMembers(members);

      // ✅ attachments 배열에서 ID 추출
      const existingAttachmentIds = project.attachments?.map((att) => att.id) || [];
      setCurrentAttachmentIds(existingAttachmentIds);

      // ✅ 첫 번째 첨부파일로 UI 초기화
      if (project.attachments && project.attachments.length > 0) {
        const firstAttachment = project.attachments[0];
        setInitialFile(firstAttachment.fileUrl, firstAttachment.fileName);
        setAttachmentId(firstAttachment.id);
      } else {
        handleRemoveFile();
      }
    } else if (mode === 'create') {
      setName('');
      setDescription('');
      setStartDate('');
      setDueDate('');
      setProjectMembers(undefined);
      setCurrentAttachmentIds([]);
      handleRemoveFile();
    }
    setError(null);
  }, [
    project?.projectId,
    mode,
    setInitialFile,
    handleRemoveFile,
    setAttachmentId,
    members.length,
    project?.attachments, // ✅ attachments 변경 감지
  ]);
  const fetchBoards = useCallback(async () => {
    if (!project || mode !== 'detail') {
      setBoards([]);
      return;
    }
    setIsBoardsLoading(true);
    try {
      const response = await getBoardsByProject(project.projectId);
      setBoards(response || []);
    } catch (err) {
      console.error('❌ Failed to fetch boards for statistics:', err);
      setBoards([]);
    } finally {
      setIsBoardsLoading(false);
    }
  }, [project, mode]);

  useEffect(() => {
    fetchBoards();
  }, [fetchBoards]);

  const projectStats = useMemo(() => {
    const totalBoards = boards.length;
    const inProgressBoards = boards.filter((b) => (b as any).status === 'IN_PROGRESS').length;
    const delayedBoards = boards.filter((b) => (b as any).isDelayed).length;
    return {
      totalBoards,
      inProgressBoards,
      delayedBoards,
    };
  }, [boards]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!name.trim()) {
      setError('프로젝트 이름은 필수입니다.');
      return;
    }

    setIsLoading(true);
    setError(null);

    let attachmentIdsPayload: string[] | undefined = undefined; // ✅ 초기값 undefined

    try {
      // 1. 새 파일이 선택된 경우 - 업로드 후 새 ID 사용
      if (selectedFile) {
        const uploadResult = await upload(workspaceId, 'project');
        if (uploadResult) {
          attachmentIdsPayload = [uploadResult.attachmentId];
        }
      }
      // 2. Edit 모드 + 기존 파일 삭제한 경우 - 빈 배열 전송
      else if (mode === 'edit' && !previewUrl && currentAttachmentIds.length > 0) {
        attachmentIdsPayload = [];
      }
      // 3. Edit 모드 + 기존 파일 유지 - undefined (전송 안 함)
      // 4. Create 모드 + 파일 없음 - undefined (전송 안 함)

      const projectBaseData = {
        name: name.trim(),
        description: description.trim() || undefined,
        startDate: startDate ? `${startDate}T00:00:00Z` : undefined,
        dueDate: dueDate ? `${dueDate}T00:00:00Z` : undefined,
      };

      if (mode === 'edit' && project) {
        const updatePayload: UpdateProjectRequest = {
          ...projectBaseData,
          attachmentIds: attachmentIdsPayload, // undefined면 필드 자체가 전송되지 않음
        };

        await updateProject(project.projectId, updatePayload);
        alert(`✅ ${name} 프로젝트가 수정되었습니다!`);
        onProjectSaved();
        setMode('detail');
      } else if (mode === 'create') {
        const createPayload: CreateProjectRequest = {
          workspaceId: workspaceId,
          ...projectBaseData,
          attachmentIds: attachmentIdsPayload,
        };
        const newProjectResponse: ProjectResponse = await createProject(createPayload);
        alert(`✅ ${name} 프로젝트가 생성되었습니다!`);
        if (newProjectResponse) {
          onProjectCreated?.(newProjectResponse);
        }
        onProjectSaved();
        onClose();
      }
    } catch (err: any) {
      const errorMsg = err.response?.data?.error?.message || err.message;
      console.error(mode === 'create' ? '❌ 생성 실패:' : '❌ 수정 실패:', errorMsg);
      setError(errorMsg || '작업 처리에 실패했습니다.');
    } finally {
      setIsLoading(false);
    }
  };

  const modalTitle = useMemo(() => {
    switch (mode) {
      case 'create':
        return '새 프로젝트 만들기';
      case 'edit':
        return `${project?.name || '프로젝트'} 수정`;
      default:
        return `${project?.name || '프로젝트'} 상세 정보`;
    }
  }, [mode, project?.name]);

  // ========================================
  // 🔧 수정 4: 상세 보기용 파일 정보 (attachments 배열 사용)
  // ========================================
  const firstAttachment = project?.attachments?.[0];
  const detailFileUrl = firstAttachment?.fileUrl || '';
  const detailFileName = firstAttachment?.fileName || 'project_file_attachment';
  const hasAttachments = !!firstAttachment;

  // ----------------------------------------------------
  // 🎨 Detail / Edit Mode 렌더링
  // ----------------------------------------------------
  const renderDetailOrEditContent = () => (
    <form onSubmit={handleSubmit} className="space-y-4">
      <div className="grid grid-cols-3 gap-6">
        <div className="col-span-2 space-y-4">
          {/* Name */}
          <div>
            <label className="block text-sm font-semibold text-gray-700 mb-2">
              프로젝트 이름 <span className="text-red-500">*</span>
            </label>
            <input
              type="text"
              value={name} // ✅ State 연결
              onChange={(e) => setName(e.target.value)} // ✅ Handler 연결
              disabled={mode === 'detail' || isLoading}
              className={`w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm ${
                mode === 'detail' ? 'bg-gray-100 text-gray-700' : 'bg-white'
              }`}
              maxLength={100}
            />
          </div>

          {/* Dates */}
          <div className="grid grid-cols-2 gap-4">
            <div className="col-span-1">
              <label className="block text-sm font-semibold text-gray-700 mb-2">시작일</label>
              <input
                type="date"
                value={startDate} // ✅ State 연결
                onChange={(e) => setStartDate(e.target.value)} // ✅ Handler 연결
                disabled={mode === 'detail' || isLoading}
                className={`w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm ${
                  mode === 'detail' ? 'bg-gray-100 text-gray-700' : 'bg-white'
                }`}
              />
            </div>
            <div className="col-span-1">
              <label className="block text-sm font-semibold text-gray-700 mb-2">마감일</label>
              <input
                type="date"
                value={dueDate} // ✅ State 연결
                onChange={(e) => setDueDate(e.target.value)} // ✅ Handler 연결
                disabled={mode === 'detail' || isLoading}
                className={`w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm ${
                  mode === 'detail' ? 'bg-gray-100 text-gray-700' : 'bg-white'
                }`}
              />
            </div>
          </div>

          {/* Description */}
          <div>
            <label className="block text-sm font-semibold text-gray-700 mb-2">프로젝트 설명</label>
            <textarea
              value={description} // ✅ State 연결
              onChange={(e) => setDescription(e.target.value)} // ✅ Handler 연결
              disabled={mode === 'detail' || isLoading}
              className={`w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm resize-none ${
                mode === 'detail' ? 'bg-gray-100 text-gray-700' : 'bg-white'
              }`}
              rows={10}
              maxLength={800}
            />
          </div>

          {/* 💡 [수정] Files 섹션: Edit 모드에서는 Uploader, Detail 모드에서는 다운로드 버튼 */}
          <div className="pt-0">
            {mode === 'edit' ? (
              // ✏️ 수정 모드: 파일 업로더 표시
              <FileUploader
                selectedFile={selectedFile}
                previewUrl={previewUrl}
                onFileSelect={handleFileSelect}
                onRemoveFile={handleRemoveFile}
                existingFileName={firstAttachment?.fileName}
                disabled={isLoading}
                label="첨부 파일 수정"
              />
            ) : (
              // 📖 상세 보기 모드: 다운로드 UI 표시
              <>
                <label className="block text-sm font-semibold text-gray-700 mb-2 flex items-center gap-1">
                  <Paperclip className="w-4 h-4 text-blue-500" />
                  첨부 파일
                </label>
                <div className="p-2 bg-gray-50 border border-gray-200 rounded-lg flex items-center justify-between text-sm">
                  <span className="text-gray-700 truncate flex items-center gap-1">
                    {hasAttachments ? (
                      <span className="text-gray-700">{detailFileName}</span>
                    ) : (
                      <span className="text-gray-500">첨부 파일 없음</span>
                    )}
                  </span>

                  {hasAttachments ? (
                    <button
                      type="button"
                      onClick={() => handleFileDownload(detailFileUrl, detailFileName)}
                      className="flex items-center gap-1 text-blue-600 hover:text-blue-700 transition font-medium ml-2 flex-shrink-0"
                    >
                      <Download className="w-4 h-4" />
                      <span className="text-xs">다운로드</span>
                    </button>
                  ) : (
                    <span className="text-gray-400 text-xs flex-shrink-0">다운로드 불가</span>
                  )}
                </div>
              </>
            )}
          </div>
        </div>

        {/* Right Section (Same as before) */}
        <div className="col-span-1 space-y-4 divide-y divide-gray-200 pl-4 border-l border-gray-200">
          {project && (
            <div className="pb-4">
              <label className="block text-sm font-semibold text-gray-700 mb-2 flex items-center gap-1">
                <UserIcon className="w-4 h-4 text-gray-500" />
                프로젝트 소유자
              </label>
              <div className="text-sm font-medium text-gray-700 ml-1">{project.ownerName}</div>
            </div>
          )}

          <div className="pt-4">
            <h3 className="text-md font-bold text-gray-800 mb-2">
              소속 멤버 ({projectMembers?.length}명)
            </h3>
            <div className="max-h-56 overflow-y-auto space-y-2">
              {projectMembers?.map((member) => (
                <div
                  key={member?.userId}
                  className="flex items-center justify-between p-2 rounded-lg hover:bg-gray-100 transition"
                >
                  <span className="text-sm">{member?.userName}</span>
                  <span
                    className={`text-xs px-2 py-0.5 rounded-full ${
                      member?.role === 'OWNER'
                        ? 'bg-red-100 text-red-600'
                        : member.role === 'MEMBER'
                        ? 'bg-blue-100 text-blue-600'
                        : 'bg-gray-100 text-gray-500'
                    }`}
                  >
                    {member?.role}
                  </span>
                </div>
              ))}
            </div>
          </div>

          {mode === 'detail' && (
            <div className="pt-4">
              <h3 className="text-md font-bold text-gray-800 flex items-center gap-2 mb-3">
                <BarChart3 className="w-5 h-5 text-indigo-500" /> 프로젝트 현황
              </h3>
              {isBoardsLoading ? (
                <div className="flex justify-center items-center py-4 text-gray-500">
                  <Loader2 className="w-5 h-5 animate-spin mr-2" />
                  통계 데이터 로드 중...
                </div>
              ) : (
                <div className="grid grid-cols-3 gap-3 text-center">
                  <div className="p-3 bg-indigo-50 rounded-lg border border-indigo-200">
                    <p className="text-2xl font-bold text-indigo-700">{projectStats.totalBoards}</p>
                    <p className="text-xs text-indigo-500 mt-1">총 보드 수</p>
                  </div>
                  <div className="p-3 bg-green-50 rounded-lg border border-green-200">
                    <p className="text-2xl font-bold text-green-700">
                      {projectStats.inProgressBoards}
                    </p>
                    <p className="text-xs text-green-500 mt-1">진행 중</p>
                  </div>
                  <div className="p-3 bg-red-50 rounded-lg border border-red-200">
                    <p className="text-2xl font-bold text-red-700">{projectStats.delayedBoards}</p>
                    <p className="text-xs text-red-500 mt-1">지연</p>
                  </div>
                </div>
              )}
            </div>
          )}
        </div>
      </div>

      {/* Actions */}
      <div className="flex gap-3 pt-4 px-6 sticky bottom-0 bg-white border-t border-gray-300">
        {mode === 'edit' && (
          <button
            type="button"
            onClick={() => setMode('detail')}
            className="flex-1 px-4 py-2 border border-gray-300 text-gray-700 font-semibold rounded-lg hover:bg-gray-50 transition"
            disabled={isLoading}
          >
            취소 (상세 보기로)
          </button>
        )}

        {mode === 'detail' && (
          <button
            type="button"
            onClick={onClose}
            className="flex-1 px-4 py-2 bg-gray-200 text-gray-700 font-semibold rounded-lg hover:bg-gray-300 transition"
          >
            닫기
          </button>
        )}

        {(mode === 'edit' || mode === 'create') && (
          <button
            type="submit"
            className={`flex-1 px-4 py-2 bg-blue-500 text-white font-semibold rounded-lg hover:bg-blue-600 transition ${
              isLoading ? 'opacity-50 cursor-not-allowed' : ''
            }`}
            disabled={isLoading}
          >
            {isLoading
              ? mode === 'edit'
                ? '저장 중...'
                : '생성 중...'
              : mode === 'edit'
              ? '수정 내용 저장'
              : '프로젝트 만들기'}
          </button>
        )}
      </div>
    </form>
  );

  // ----------------------------------------------------
  // 🎨 Create Mode 렌더링
  // ----------------------------------------------------
  const renderCreateContent = () => (
    <form onSubmit={handleSubmit} className="space-y-4">
      <div className="grid grid-cols-3 gap-6">
        <div className="col-span-2 space-y-4">
          <div>
            <label className="block text-sm font-semibold text-gray-700 mb-2">
              프로젝트 이름 <span className="text-red-500">*</span>
            </label>
            <input
              type="text"
              value={name} // ✅ State 연결
              onChange={(e) => setName(e.target.value)} // ✅ Handler 연결
              placeholder="예: Wealist 서비스 개발"
              className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm"
              disabled={isLoading}
              maxLength={100}
              autoFocus
            />
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-semibold text-gray-700 mb-2">
                시작일 (선택)
              </label>
              <input
                type="date"
                value={startDate} // ✅ State 연결
                onChange={(e) => setStartDate(e.target.value)} // ✅ Handler 연결
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm"
                disabled={isLoading}
              />
            </div>
            <div>
              <label className="block text-sm font-semibold text-gray-700 mb-2">
                마감일 (선택)
              </label>
              <input
                type="date"
                value={dueDate} // ✅ State 연결
                onChange={(e) => setDueDate(e.target.value)} // ✅ Handler 연결
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm"
                disabled={isLoading}
              />
            </div>
          </div>

          <div>
            <label className="block text-sm font-semibold text-gray-700 mb-2">
              프로젝트 설명 (선택)
            </label>
            <textarea
              value={description} // ✅ State 연결
              onChange={(e) => setDescription(e.target.value)} // ✅ Handler 연결
              placeholder="프로젝트에 대한 간단한 설명을 입력하세요"
              className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm resize-none"
              rows={5}
              disabled={isLoading}
              maxLength={500}
            />
          </div>

          {/* 💡 [추가] Create 모드에도 FileUploader 추가 */}
          <div>
            <FileUploader
              selectedFile={selectedFile}
              previewUrl={previewUrl}
              onFileSelect={handleFileSelect}
              onRemoveFile={handleRemoveFile}
              disabled={isLoading}
              label="첨부 파일 (선택)"
            />
          </div>
        </div>

        {/* Right Section (Instructions) */}
        <div className="col-span-1 space-y-4 divide-y divide-gray-200 pl-4 border-l border-gray-200">
          <div className="pb-4">
            <label className="block text-sm font-semibold text-gray-700 mb-2 flex items-center gap-1">
              <UserIcon className="w-4 h-4 text-gray-500" />
              프로젝트 생성 안내
            </label>
            <p className="text-xs text-gray-500">
              프로젝트 생성 시, 자동으로 소유자(Owner) 역할을 갖게 됩니다. 생성 후 멤버를
              초대하거나, 설정을 변경할 수 있습니다.
            </p>
          </div>

          <div className="pt-4">
            <h3 className="text-md font-bold text-gray-800 mb-2">마감일 설정 Tip</h3>
            <p className="text-xs text-gray-600">
              시작일과 마감일을 명확히 설정하면, 보드 현황판에서 **지연된 보드**를 정확하게 파악할
              수 있습니다.
            </p>
          </div>
        </div>
      </div>

      <div className="flex gap-3 pt-4 px-6 sticky bottom-0 bg-white border-t border-gray-300">
        <button
          type="submit"
          className={`flex-1 px-4 py-2 bg-blue-500 text-white font-semibold rounded-lg hover:bg-blue-600 transition ${
            isLoading ? 'opacity-50 cursor-not-allowed' : ''
          }`}
          disabled={isLoading}
        >
          {isLoading ? '생성 중...' : '프로젝트 만들기'}
        </button>
        <button
          type="button"
          onClick={onClose}
          className="flex-1 px-4 py-2 border border-gray-300 text-gray-700 font-semibold rounded-lg hover:bg-gray-50 transition"
          disabled={isLoading}
        >
          취소
        </button>
      </div>
    </form>
  );

  // Return 구문은 동일 (Portal 등)
  return (
    <Portal>
      <div
        className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-[9999]"
        onClick={onClose}
      >
        <div
          className={`relative w-full max-w-4xl ${theme.colors.card} p-6 ${theme.effects.borderRadius} shadow-xl max-h-[90vh] overflow-y-auto`}
          onClick={(e) => e.stopPropagation()}
        >
          {/* Header */}
          <div className="flex items-center justify-between mb-4 pb-2">
            <div className="flex items-center">
              <h2 className="text-xl font-bold text-gray-800">{modalTitle}</h2>
              {mode !== 'create' && canEdit && (
                <div className="flex items-center gap-3">
                  {mode === 'detail' ? (
                    <button
                      onClick={() => setMode('edit')}
                      title="프로젝트 수정"
                      className="p-2 rounded-full hover:bg-yellow-50 text-yellow-600 transition"
                    >
                      <Edit2 className="w-5 h-5" />
                    </button>
                  ) : (
                    <button
                      onClick={() => setMode('detail')}
                      title="수정 취소"
                      className="p-2 rounded-full hover:bg-gray-100 text-gray-600 transition"
                    >
                      <Lock className="w-5 h-5" />
                    </button>
                  )}
                </div>
              )}
            </div>
            <button
              onClick={onClose}
              className="p-2 rounded-full hover:bg-gray-100 text-gray-500 hover:text-gray-700 transition"
            >
              <X className="w-5 h-5" />
            </button>
          </div>

          {error && (
            <div className="mb-4 p-3 bg-red-50 border border-red-300 rounded-lg text-red-700 text-sm mx-6">
              {error}
            </div>
          )}

          {mode === 'create' ? renderCreateContent() : renderDetailOrEditContent()}
        </div>
      </div>
    </Portal>
  );
};
