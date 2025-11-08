import React, { useState, useEffect } from 'react';
import {
  X,
  Calendar,
  User,
  AlertCircle,
  Tag,
  CheckSquare,
  MessageSquare,
  Send,
  Edit2,
  Trash2,
} from 'lucide-react';
import { useTheme } from '../../contexts/ThemeContext';
import {
  CustomStageResponse,
  CustomRoleResponse,
  CustomImportanceResponse,
  getProjectStages,
  getProjectRoles,
  getProjectImportances,
  getBoard,
  deleteBoard,
} from '../../api/board/boardService';
import { WorkspaceMember, getWorkspaceMembers } from '../../api/user/userService';

/**
 * BoardDetailModal - 보드 상세 보기 및 수정
 */
interface BoardDetailModalProps {
  boardId: string;
  workspaceId: string;
  onClose: () => void;
  onBoardUpdated: () => void;
  onBoardDeleted: () => void;
  onEdit: (boardData: {
    boardId: string;
    projectId: string;
    title: string;
    content: string;
    stageId: string;
    roleId: string;
    importanceId: string;
    assigneeIds: string[];
    dueDate: string;
  }) => void;
}

export const BoardDetailModal: React.FC<BoardDetailModalProps> = ({
  boardId,
  workspaceId,
  onClose,
  onBoardUpdated,
  onBoardDeleted,
  onEdit,
}) => {
  const { theme } = useTheme();
  const accessToken = localStorage.getItem('access_token') || '';

  // Form state
  const [projectId, setProjectId] = useState<string>('');
  const [title, setTitle] = useState('');
  const [content, setContent] = useState('');
  const [selectedStageId, setSelectedStageId] = useState('');
  const [selectedRoleId, setSelectedRoleId] = useState<string>('');
  const [selectedImportanceId, setSelectedImportanceId] = useState<string>('');
  const [assigneeIds, setAssigneeIds] = useState<string[]>([]);
  const [dueDate, setDueDate] = useState<string>('');

  // Data state
  const [stages, setStages] = useState<CustomStageResponse[]>([]);
  const [roles, setRoles] = useState<CustomRoleResponse[]>([]);
  const [importances, setImportances] = useState<CustomImportanceResponse[]>([]);
  const [workspaceMembers, setWorkspaceMembers] = useState<WorkspaceMember[]>([]);

  // UI state
  const [isLoading, setIsLoading] = useState(false);
  const [isLoadingBoard, setIsLoadingBoard] = useState(true);
  const [isLoadingFields, setIsLoadingFields] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Comment state
  const [comments, setComments] = useState<any[]>([]);
  const [newComment, setNewComment] = useState('');

  // 보드 데이터 조회
  useEffect(() => {
    const fetchBoard = async () => {
      setIsLoadingBoard(true);
      try {
        const boardData = await getBoard(boardId, accessToken);

        // 보드 데이터로 상태 초기화
        setProjectId(boardData.projectId);
        setTitle(boardData.title);
        setContent(boardData.content || '');
        setSelectedStageId(boardData.stage?.id || '');
        // roles가 배열이므로 첫 번째 역할만 선택 (단일 선택으로 변경)
        setSelectedRoleId(boardData.roles?.[0]?.id || '');
        setSelectedImportanceId(boardData.importance?.id || '');
        // assignees가 배열이면 userId 추출, 단일 객체면 변환
        const assignees = Array.isArray(boardData.assignees)
          ? boardData.assignees.map((a: any) => a.userId).filter(Boolean)
          : boardData.assignee?.userId
          ? [boardData.assignee.userId]
          : [];
        setAssigneeIds(assignees);
        setDueDate(boardData.dueDate || '');

        console.log('✅ 보드 데이터 로드 성공:', boardData);
      } catch (err) {
        console.error('❌ 보드 데이터 로드 실패:', err);
        setError('보드 정보를 불러오는데 실패했습니다.');
      } finally {
        setIsLoadingBoard(false);
      }
    };

    fetchBoard();
  }, [boardId, accessToken]);

  // Custom Fields 조회 (projectId가 설정된 후)
  useEffect(() => {
    if (!projectId) return;

    const fetchCustomFields = async () => {
      setIsLoadingFields(true);
      try {
        const [stagesData, rolesData, importancesData] = await Promise.all([
          getProjectStages(projectId, accessToken),
          getProjectRoles(projectId, accessToken),
          getProjectImportances(projectId, accessToken),
        ]);

        setStages(stagesData);
        setRoles(rolesData);
        setImportances(importancesData);
      } catch (err) {
        console.error('❌ Custom Fields 로드 실패:', err);
        setError('커스텀 필드를 불러오는데 실패했습니다.');
      } finally {
        setIsLoadingFields(false);
      }
    };

    fetchCustomFields();
  }, [projectId, accessToken]);

  // 워크스페이스 멤버 조회
  useEffect(() => {
    const fetchMembers = async () => {
      try {
        const members = await getWorkspaceMembers(workspaceId, accessToken);
        setWorkspaceMembers(members);
        console.log('✅ 워크스페이스 멤버 로드:', members.length);
      } catch (err) {
        console.error('❌ 워크스페이스 멤버 로드 실패:', err);
      }
    };

    if (workspaceId) {
      fetchMembers();
    }
  }, [workspaceId, accessToken]);

  const handleDelete = async () => {
    if (!confirm('정말로 이 보드를 삭제하시겠습니까?')) return;

    setIsLoading(true);
    try {
      await deleteBoard(boardId, accessToken);
      console.log('✅ 보드 삭제 성공');
      onBoardDeleted();
      onClose();
    } catch (err) {
      const error = err as Error;
      console.error('❌ 보드 삭제 실패:', error);
      setError(error.message || '보드 삭제에 실패했습니다.');
    } finally {
      setIsLoading(false);
    }
  };

  const handleAddComment = () => {
    if (newComment.trim()) {
      setComments([
        ...comments,
        {
          id: comments.length + 1,
          author: '사용자',
          content: newComment,
          timestamp: '방금 전',
        },
      ]);
      setNewComment('');
    }
  };

  // 로딩 중이면 로딩 UI 표시
  if (isLoadingBoard || isLoadingFields) {
    return (
      <div
        className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-[90]"
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

  return (
    <div
      className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-[90]"
      onClick={onClose}
    >
      <div
        className={`relative w-full max-w-2xl ${theme.colors.card} p-6 ${theme.effects.borderRadius} shadow-xl max-h-[90vh] overflow-y-auto`}
        onClick={(e) => e.stopPropagation()}
      >
        {/* Header */}
        <div className="flex items-start justify-between mb-4 pb-4 border-b border-gray-200">
          <div className="flex-1 pr-4">
            <h2 className="text-xl font-bold text-gray-800 mb-2">{title}</h2>
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
          <div>
            <label className="block text-sm font-semibold text-gray-700 mb-2">설명</label>
            <p className="text-sm text-gray-600 whitespace-pre-wrap">
              {content || '설명이 없습니다.'}
            </p>
          </div>

          {/* Stage and Role - 2 columns */}
          <div className="grid grid-cols-2 gap-4">
            {/* Stage */}
            <div>
              <label className="block text-sm font-semibold text-gray-700 mb-2">
                <CheckSquare className="w-4 h-4 inline mr-1" />
                진행 단계
              </label>
              <div className="flex items-center gap-2">
                <span
                  className="w-3 h-3 rounded-full"
                  style={{
                    backgroundColor:
                      stages.find((s) => s.id === selectedStageId)?.color || '#6B7280',
                  }}
                />
                <span className="text-sm">
                  {stages.find((s) => s.id === selectedStageId)?.name || '알 수 없음'}
                </span>
              </div>
            </div>

            {/* Role (단일 선택) */}
            <div>
              <label className="block text-sm font-semibold text-gray-700 mb-2">
                <Tag className="w-4 h-4 inline mr-1" />
                역할
              </label>
              <div className="flex items-center gap-2">
                <span
                  className="w-3 h-3 rounded-full"
                  style={{
                    backgroundColor: roles.find((r) => r.id === selectedRoleId)?.color || '#6B7280',
                  }}
                />
                <span className="text-sm">
                  {roles.find((r) => r.id === selectedRoleId)?.name || '알 수 없음'}
                </span>
              </div>
            </div>
          </div>

          {/* Importance */}
          <div>
            <label className="block text-sm font-semibold text-gray-700 mb-2">
              <AlertCircle className="w-4 h-4 inline mr-1" />
              중요도
            </label>
            <div className="flex items-center gap-2">
              {selectedImportanceId ? (
                <>
                  <span
                    className="w-3 h-3 rounded-full"
                    style={{
                      backgroundColor:
                        importances.find((i) => i.id === selectedImportanceId)?.color ||
                        '#6B7280',
                    }}
                  />
                  <span className="text-sm">
                    {importances.find((i) => i.id === selectedImportanceId)?.name || '알 수 없음'}
                  </span>
                </>
              ) : (
                <span className="text-sm text-gray-500">없음</span>
              )}
            </div>
          </div>

          {/* Assignee and Due Date */}
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-semibold text-gray-700 mb-2">
                <User className="w-4 h-4 inline mr-1" />
                담당자
              </label>
              {assigneeIds.length > 0 ? (
                <div className="flex flex-wrap gap-1">
                  {assigneeIds.map((userId) => {
                    const member = workspaceMembers.find((m) => m.userId === userId);
                    return (
                      <span
                        key={userId}
                        className="inline-flex items-center px-2 py-1 bg-blue-100 text-blue-700 text-xs rounded-full"
                      >
                        {member?.name || userId}
                      </span>
                    );
                  })}
                </div>
              ) : (
                <p className="text-sm text-gray-600">없음</p>
              )}
            </div>

            <div>
              <label className="block text-sm font-semibold text-gray-700 mb-2">
                <Calendar className="w-4 h-4 inline mr-1" />
                마감일
              </label>
              <p className="text-sm text-gray-600">
                {dueDate ? new Date(dueDate).toLocaleDateString('ko-KR') : '없음'}
              </p>
            </div>
          </div>
        </div>

        {/* Comments Section */}
        <div className="pt-4 border-t border-gray-200">
          <div className="flex items-center gap-2 mb-4">
            <MessageSquare className="w-5 h-5 text-gray-700" />
            <h3 className="text-base font-bold text-gray-800">댓글 ({comments.length}개)</h3>
          </div>

          <div className="space-y-3 mb-4 max-h-40 overflow-y-auto">
            {comments.map((comment) => (
              <div key={comment.id} className="p-3 bg-gray-100 border border-gray-200 rounded-lg">
                <div className="flex items-start gap-2">
                  <div className="w-6 h-6 bg-blue-500 flex items-center justify-center text-white text-xs font-bold rounded-full flex-shrink-0">
                    {comment.author[0]}
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2 mb-1">
                      <span className="text-xs font-bold">{comment.author}</span>
                      <span className="text-[10px] text-gray-500">{comment.timestamp}</span>
                    </div>
                    <p className="text-sm break-words text-gray-700">{comment.content}</p>
                  </div>
                </div>
              </div>
            ))}
          </div>

          <div className="flex gap-2">
            <input
              type="text"
              value={newComment}
              onChange={(e) => setNewComment(e.target.value)}
              onKeyPress={(e) => e.key === 'Enter' && handleAddComment()}
              placeholder="댓글을 입력하세요..."
              className="flex-1 px-3 py-2 border border-gray-300 text-sm rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
              disabled={isLoading}
            />
            <button
              onClick={handleAddComment}
              disabled={isLoading || !newComment.trim()}
              className="bg-blue-500 text-white px-4 py-2 hover:bg-blue-600 transition flex items-center justify-center gap-1 rounded-lg disabled:bg-gray-400"
            >
              <Send className="w-4 h-4" />
              <span className="text-xs">등록</span>
            </button>
          </div>
        </div>

        {/* Actions */}
        <div className="flex gap-3 mt-6 pt-4 border-t border-gray-300">
          <button
            onClick={() => {
              onEdit({
                boardId,
                projectId,
                title,
                content,
                stageId: selectedStageId,
                roleId: selectedRoleId,
                importanceId: selectedImportanceId,
                assigneeIds,
                dueDate,
              });
            }}
            className="flex-1 px-4 py-2 bg-blue-500 text-white font-semibold rounded-lg hover:bg-blue-600 transition disabled:opacity-50 flex items-center justify-center gap-2"
            disabled={isLoading}
          >
            <Edit2 className="w-4 h-4" />
            보드 수정
          </button>
          <button
            onClick={handleDelete}
            className="flex-1 px-4 py-2 bg-red-500 text-white font-semibold rounded-lg hover:bg-red-600 transition disabled:opacity-50 flex items-center justify-center gap-2"
            disabled={isLoading}
          >
            <Trash2 className="w-4 h-4" />
            보드 삭제
          </button>
        </div>
      </div>
    </div>
  );
};
