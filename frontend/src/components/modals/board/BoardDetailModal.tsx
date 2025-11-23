// src/components/modals/board/BoardDetailModal.tsx

import React, { useState, useEffect } from 'react'; // useRef 삭제 (CommentList 내부로 이동됨)
import {
  X,
  AlertCircle,
  Tag,
  CheckSquare,
  MessageSquare,
  Edit2,
  Trash2,
  Paperclip,
  User,
  Users,
  Download,
  Calendar,
} from 'lucide-react';
import { useTheme } from '../../../contexts/ThemeContext';
import {
  BoardDetailResponse,
  FieldOption,
  CommentResponse,
  ParticipantResponse,
  AttachmentResponse,
} from '../../../types/board';
import {
  getBoard,
  deleteBoard,
  getCommentsByBoard,
  // createComment 삭제 (CommentList 내부에서 처리)
} from '../../../api/board/boardService'; // apis/board 경로 확인 필요
import { getWorkspaceMembers } from '../../../api/user/userService'; // apis/user 경로 확인 필요
import { WorkspaceMemberResponse } from '../../../types/user';
import { AvatarStack } from '../../common/AvartarStack';
import { formatDate } from '../../../utils/date';
import Portal from '../../common/Portal';
import CommentList from '../../comment/CommentList'; // 경로 확인 필요
import { useAuth } from '../../../contexts/AuthContext';

// 1. 정적 데이터를 담을 인터페이스 정의
interface BoardState {
  projectId: string;
  title: string;
  content: string;
  selectedStageId: string;
  selectedRoleId: string;
  selectedImportanceId: string;
  selectedAssigneeId: string;
  dueDate: string;
  startDate: string;
  createdAt: string;
  updatedAt: string;
  participants: ParticipantResponse[];
  fileUrl?: string;
  fileName?: string;
  attachments: AttachmentResponse[]; // 💡 첨부파일 배열 호환을 위해 추가
}

const initialBoardState: BoardState = {
  projectId: '',
  title: '',
  content: '',
  selectedStageId: '',
  selectedRoleId: '',
  selectedImportanceId: '',
  selectedAssigneeId: '',
  dueDate: '',
  startDate: '',
  createdAt: '',
  updatedAt: '',
  participants: [],
  fileUrl: undefined,
  fileName: undefined,
  attachments: [],
};

interface BoardDetailModalProps {
  boardId: string;
  workspaceId: string;
  onClose: () => void;
  onBoardUpdated: () => void; // 💡 사용되지 않더라도 인터페이스 유지
  onBoardDeleted: () => void;
  onEdit: (boardData: {
    boardId: string;
    projectId: string;
    title: string;
    content: string;
    stage: string;
    assigneeId?: string;
    role: string;
    importance?: string;
    dueDate?: string;
    startDate?: string;
    attachments?: AttachmentResponse[];
  }) => void;
  fieldOptionsLookup: {
    stages?: FieldOption[];
    roles?: FieldOption[];
    importances?: FieldOption[];
  };
}

export const BoardDetailModal: React.FC<BoardDetailModalProps> = ({
  boardId,
  workspaceId,
  onClose,
  onBoardDeleted,
  onEdit,
  fieldOptionsLookup,
}) => {
  const { theme } = useTheme();
  const { userId } = useAuth();

  const [boardData, setBoardData] = useState<BoardState>(initialBoardState);
  const [workspaceMembers, setWorkspaceMembers] = useState<WorkspaceMemberResponse[]>([]);

  // UI state
  const [isLoading, setIsLoading] = useState(false);
  const [isLoadingBoard, setIsLoadingBoard] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Comment state
  const [comments, setComments] = useState<CommentResponse[]>([]);
  const [previewImage, setPreviewImage] = useState<string | null>(null);
  // 💡 삭제: newComment, selectedFile, fileInputRef (CommentList가 담당함)
  // 이미지 파일 여부 확인
  const isImageFile = (contentType?: string, fileName?: string): boolean => {
    if (contentType) {
      return contentType.startsWith('image/');
    }
    if (fileName) {
      const ext = fileName.split('.').pop()?.toLowerCase();
      return ['jpg', 'jpeg', 'png', 'gif', 'webp', 'svg'].includes(ext || '');
    }
    return false;
  };
  // 보드 데이터 조회
  useEffect(() => {
    const fetchBoard = async () => {
      setIsLoadingBoard(true);
      try {
        const data: BoardDetailResponse = await getBoard(boardId);
        const customFields = data.customFields || {};

        setBoardData({
          projectId: data.projectId || '',
          title: data.title || '',
          content: data.content || '',
          selectedStageId: customFields.stage || '',
          selectedRoleId: customFields.role || '',
          selectedImportanceId: customFields.importance || '',
          selectedAssigneeId: data.assigneeId || '',
          dueDate: (data as any).dueDate || '',
          startDate: (data as any).startDate || '',
          createdAt: data.createdAt,
          updatedAt: data.updatedAt,
          participants: data.participants || [],
          attachments: data.attachments || [],
          // 💡 단일 fileUrl 지원을 위해 첫 번째 첨부파일 매핑 (필요 시)
          fileUrl: data.attachments?.[0]?.fileUrl,
          fileName: data.attachments?.[0]?.fileName,
        });

        setComments(data.comments || []);
      } catch (err) {
        console.error('❌ 보드 데이터 로드 실패:', err);
        setError('보드 정보를 불러오는데 실패했습니다.');
      } finally {
        setIsLoadingBoard(false);
      }
    };

    fetchBoard();
  }, [boardId]);

  const { stages = [], roles = [], importances = [] } = fieldOptionsLookup;

  useEffect(() => {
    const fetchMembers = async () => {
      try {
        const members = await getWorkspaceMembers(workspaceId);
        setWorkspaceMembers(members);
      } catch (err) {
        console.error('❌ 워크스페이스 멤버 로드 실패:', err);
      }
    };

    if (workspaceId) {
      fetchMembers();
    }
  }, [workspaceId]);

  // 댓글 목록 갱신 함수 (CommentList에 전달)
  const fetchComments = async () => {
    try {
      const res = await getCommentsByBoard(boardId);
      setComments(res);
    } catch (error) {
      console.error('댓글 갱신 실패:', error);
    }
  };

  const handleDelete = async () => {
    if (!window.confirm('정말 이 보드를 삭제하시겠습니까?')) return;

    setIsLoading(true);
    try {
      await deleteBoard(boardId);
      alert('보드가 삭제되었습니다.');
      onBoardDeleted();
      onClose();
    } catch (err: any) {
      const errorMsg = err.response?.data?.error?.message || err.message;
      console.error('❌ 보드 삭제 실패:', errorMsg);
      setError(errorMsg || '보드 삭제에 실패했습니다.');
    } finally {
      setIsLoading(false);
    }
  };

  const handleFileDownload = (fileUrl: string, fileName: string) => {
    if (!fileUrl) return;
    const link = document.createElement('a');
    link.href = fileUrl;
    link.setAttribute('download', fileName);
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
  };

  // 💡 삭제: handleFileChange, handleAddComment (CommentList로 이동됨)

  const getFieldOption = (options: FieldOption[], id: string) => {
    return options.find((opt) => opt.optionValue === id);
  };

  if (isLoadingBoard) {
    return (
      <div
        className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-[200]"
        onClick={onClose}
      >
        <div
          className={`relative w-full max-w-2xl ${theme.colors.card} p-6 ${theme.effects.borderRadius} shadow-xl`}
          onClick={(e) => e.stopPropagation()}
        >
          <div className="flex items-center justify-center py-12">
            <div className="text-center">
              <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500 mx-auto mb-4"></div>
              <p className="text-gray-600">보드 정보를 불러오는 중...</p>
            </div>
          </div>
        </div>
      </div>
    );
  }

  const currentStage = getFieldOption(stages, boardData.selectedStageId);
  const currentRole = getFieldOption(roles, boardData.selectedRoleId);
  const currentImportance = getFieldOption(importances, boardData.selectedImportanceId);

  const assigneeMember = workspaceMembers.find((m) => m.userId === boardData.selectedAssigneeId);
  const participantMembers = workspaceMembers.filter((m) =>
    boardData.participants.some((p) => p.userId === m.userId),
  );

  return (
    <Portal>
      <div
        className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-[9999]"
        onClick={onClose}
      >
        <div
          className={`relative w-full max-w-2xl ${theme.colors.card} p-6 ${theme.effects.borderRadius} shadow-xl max-h-[90vh] overflow-y-auto custom-scrollbar`}
          onClick={(e) => e.stopPropagation()}
        >
          {/* Header */}
          <div className="flex items-start justify-between mb-4 pb-4">
            <div className="flex-1 pr-4">
              <h2 className="text-xl font-bold text-gray-800 mb-2">
                {boardData.title || '제목 없음'}
              </h2>
            </div>
            <div className="flex gap-2">
              <button
                onClick={onClose}
                className="p-2 rounded-full hover:bg-gray-100 text-gray-500 hover:text-gray-700 transition"
              >
                <X className="w-5 h-5" />
              </button>
            </div>
          </div>

          {/* Error Message */}
          {error && (
            <div className="mb-4 p-3 bg-red-50 border border-red-300 rounded-lg text-red-700 text-sm">
              {error}
            </div>
          )}

          {/* Content */}
          <div className="space-y-4 mb-6">
            {/* Description */}
            <div className="relative">
              <div className="absolute top-0 right-0 text-right space-y-1 text-xs text-gray-500 pt-0">
                <p>
                  <span className="font-medium text-gray-700">생성일:</span>{' '}
                  {formatDate(boardData.createdAt)}
                </p>
                <p>
                  <span className="font-medium text-gray-700">수정일:</span>{' '}
                  {formatDate(boardData.updatedAt)}
                </p>
              </div>
              <label className="block text-sm font-semibold text-gray-700 mb-2">설명</label>
              <p className="text-sm text-gray-600 whitespace-pre-wrap">
                {boardData.content || '설명이 없습니다.'}
              </p>
              {/* 보드 첨부파일 다운로드 (기존 UI 개선) */}
              <div
                className="mt-4 p-2 bg-gray-50 border border-gray-200 rounded-lg flex items-center justify-between text-sm relative"
                onMouseEnter={() => {
                  if (boardData.fileUrl && isImageFile(undefined, boardData.fileName)) {
                    setPreviewImage(boardData.fileUrl);
                  }
                }}
                onMouseLeave={() => setPreviewImage(null)}
              >
                <span className="text-gray-700 truncate flex items-center gap-1">
                  <Paperclip className="w-4 h-4 text-gray-500 flex-shrink-0" />
                  {boardData.fileUrl ? (
                    <span className="text-gray-700">
                      {boardData.fileName || '첨부된 보드 파일'}
                    </span>
                  ) : (
                    <span className="text-gray-500">첨부 파일 없음</span>
                  )}
                </span>

                {boardData.fileUrl ? (
                  <button
                    type="button"
                    onClick={() => {
                      if (boardData?.fileUrl)
                        handleFileDownload(boardData.fileUrl, boardData.fileName || 'board_file');
                    }}
                    className="flex items-center gap-1 text-blue-600 hover:text-blue-700 transition font-medium ml-2 flex-shrink-0"
                  >
                    <Download className="w-4 h-4" />
                    <span className="text-xs">다운로드</span>
                  </button>
                ) : (
                  <span className="text-gray-400 text-xs flex-shrink-0">첨부 가능</span>
                )}

                {/* 💡 이미지 미리보기 툴팁 */}
                {previewImage && (
                  <div className="absolute left-0 top-full mt-2 z-50 pointer-events-none">
                    <div className="bg-white border-2 border-gray-300 rounded-lg shadow-2xl p-2">
                      <img
                        src={previewImage}
                        alt="미리보기"
                        className="max-w-xs max-h-64 rounded"
                        style={{ objectFit: 'contain' }}
                      />
                    </div>
                  </div>
                )}
              </div>
            </div>

            <hr className="mt-4 border-gray-100" />

            {/* Stage / Role / Importance */}
            <div className="grid grid-cols-2 gap-4">
              <div className="col-span-1">
                <label className="block text-sm font-semibold text-gray-700 mb-2">
                  <CheckSquare className="w-4 h-4 inline mr-1 text-blue-500" />
                  진행 단계
                </label>
                {currentStage ? (
                  <div className="flex items-center gap-2">
                    <span
                      className="w-3 h-3 rounded-full flex-shrink-0"
                      style={{
                        backgroundColor: (currentStage as any).color || '#6B7280',
                      }}
                    />
                    <span className="text-sm truncate">{currentStage.optionLabel}</span>
                  </div>
                ) : (
                  <span className="text-sm text-gray-500">미정</span>
                )}
              </div>

              <div className="col-span-1">
                <label className="block text-sm font-semibold text-gray-700 mb-2">
                  <Tag className="w-4 h-4 inline mr-1 text-purple-500" />
                  역할
                </label>
                {currentRole ? (
                  <div className="flex items-center gap-2">
                    <span
                      className="w-3 h-3 rounded-full flex-shrink-0"
                      style={{
                        backgroundColor: (currentRole as any).color || '#6B7280',
                      }}
                    />
                    <span className="text-sm truncate">{currentRole.optionLabel}</span>
                  </div>
                ) : (
                  <span className="text-sm text-gray-500">미정</span>
                )}
              </div>

              <div className="col-span-2">
                <label className="block text-sm font-semibold text-gray-700 mb-2">
                  <AlertCircle className="w-4 h-4 inline mr-1 text-red-500" />
                  중요도
                </label>
                {currentImportance ? (
                  <div className="flex items-center gap-2">
                    <span
                      className="w-3 h-3 rounded-full flex-shrink-0"
                      style={{
                        backgroundColor: (currentImportance as any).color || '#6B7280',
                      }}
                    />
                    <span className="text-sm truncate">{currentImportance.optionLabel}</span>
                  </div>
                ) : (
                  <span className="text-sm text-gray-500">없음</span>
                )}
              </div>
            </div>

            {/* Assignee and Participants */}
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-semibold text-gray-700 mb-2">
                  <User className="w-4 h-4 inline mr-1 text-green-500" />
                  작업 할당자
                </label>
                {assigneeMember ? (
                  <div className="flex items-center gap-2">
                    <AvatarStack members={[assigneeMember]} />
                    <span className="text-sm">{assigneeMember.userName}</span>
                  </div>
                ) : (
                  <span className="text-sm text-gray-500">할당되지 않음</span>
                )}
              </div>

              <div>
                <label className="block text-sm font-semibold text-gray-700 mb-2">
                  <Users className="w-4 h-4 inline mr-1 text-orange-500" />
                  작업자 ({participantMembers.length}명)
                </label>
                {participantMembers.length > 0 ? (
                  <div className="flex items-center gap-2">
                    <AvatarStack members={participantMembers} />
                    <span className="text-sm text-gray-700">
                      {participantMembers
                        .slice(0, 3)
                        .map((m) => m.userName)
                        .join(', ')}{' '}
                      {participantMembers.length > 3 && (
                        <span className="text-gray-500"> 외 {participantMembers.length - 3}명</span>
                      )}
                    </span>
                  </div>
                ) : (
                  <span className="text-sm text-gray-500">없음</span>
                )}
              </div>
            </div>

            {/* Dates */}
            <div className="grid grid-cols-2 gap-4">
              <div className="col-span-1">
                <label className="block text-sm font-semibold text-gray-700 mb-2">
                  <Calendar className="w-4 h-4 inline mr-1 text-gray-500" />
                  시작일
                </label>
                <div className="flex items-center gap-2">
                  <span className="text-sm">
                    {boardData.startDate ? formatDate(boardData.startDate) : '미정'}
                  </span>
                </div>
              </div>

              <div className="col-span-1">
                <label className="block text-sm font-semibold text-gray-700 mb-2">
                  <Calendar className="w-4 h-4 inline mr-1 text-red-500" />
                  마감일
                </label>
                <div className="flex items-center gap-2">
                  <span
                    className={`text-sm ${
                      boardData.dueDate ? 'text-red-600 font-semibold' : 'text-gray-500'
                    }`}
                  >
                    {boardData.dueDate ? formatDate(boardData.dueDate) : '미정'}
                  </span>
                </div>
              </div>
            </div>
          </div>

          {/* Comments Section */}
          <div className="pt-4 border-t border-gray-200">
            <div className="flex items-center gap-2 pb-2">
              <MessageSquare className="w-5 h-5 text-gray-700" />
              <h3 className="text-base font-bold text-gray-800">댓글 ({comments.length}개)</h3>
            </div>

            {/* 💡 CommentList 호출
              - boardId, workspaceId를 추가로 전달합니다.
              - 기존에 있던 input, file 로직을 모두 제거하고 리스트만 렌더링합니다.
            */}
            <CommentList
              boardId={boardId}
              workspaceId={workspaceId}
              comments={comments}
              members={workspaceMembers}
              currentUserId={userId || ''}
              onRefresh={fetchComments}
            />
          </div>

          {/* Actions */}
          <div className="flex gap-3 mt-6 pt-4 border-t border-gray-300">
            <button
              onClick={handleDelete}
              className="flex-1 px-4 py-2 bg-red-500 text-white font-semibold rounded-lg hover:bg-red-600 transition disabled:opacity-50 flex items-center justify-center gap-2"
              disabled={isLoading}
            >
              <Trash2 className="w-4 h-4" />
              보드 삭제
            </button>
            <button
              onClick={() => {
                onEdit({
                  boardId,
                  projectId: boardData.projectId,
                  title: boardData.title || '',
                  content: boardData.content || '',
                  stage: boardData.selectedStageId,
                  role: boardData.selectedRoleId,
                  importance: boardData.selectedImportanceId,
                  assigneeId: boardData.selectedAssigneeId,
                  dueDate: boardData.dueDate,
                  startDate: boardData.startDate,
                  attachments: boardData.attachments,
                });
              }}
              className="flex-1 px-4 py-2 bg-blue-500 text-white font-semibold rounded-lg hover:bg-blue-600 transition disabled:opacity-50 flex items-center justify-center gap-2"
              disabled={isLoading}
            >
              <Edit2 className="w-4 h-4" />
              보드 수정
            </button>
          </div>
        </div>
      </div>
    </Portal>
  );
};
