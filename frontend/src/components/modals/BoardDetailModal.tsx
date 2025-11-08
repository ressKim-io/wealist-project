import React, { useState, useEffect } from 'react';
import { X, Trash2, Calendar, Tag, User, MessageSquare, Send } from 'lucide-react';
import { useTheme } from '../../contexts/ThemeContext';
import {
  BoardResponse,
  getBoard,
  updateBoard,
  deleteBoard,
  getComments,
  createComment,
  getProjectStages,
  getProjectRoles,
  getProjectImportances,
  CommentResponse,
  CustomStageResponse,
  CustomRoleResponse,
  CustomImportanceResponse,
} from '../../api/board/boardService';

interface BoardDetailModalProps {
  boardId: string;
  projectId: string;
  onClose: () => void;
  onBoardUpdated: () => void;
  onBoardDeleted: () => void;
}

export const BoardDetailModal: React.FC<BoardDetailModalProps> = ({
  boardId,
  projectId,
  onClose,
  onBoardUpdated,
  onBoardDeleted,
}) => {
  const { theme } = useTheme();
  const accessToken = localStorage.getItem('access_token') || '';

  // Board data
  const [board, setBoard] = useState<BoardResponse | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Edit mode
  const [isEditing, setIsEditing] = useState(false);
  const [editTitle, setEditTitle] = useState('');
  const [editContent, setEditContent] = useState('');
  const [editStageId, setEditStageId] = useState('');
  const [editRoleIds, setEditRoleIds] = useState<string[]>([]);
  const [editImportanceId, setEditImportanceId] = useState('');

  // Custom fields
  const [stages, setStages] = useState<CustomStageResponse[]>([]);
  const [roles, setRoles] = useState<CustomRoleResponse[]>([]);
  const [importances, setImportances] = useState<CustomImportanceResponse[]>([]);

  // Comments
  const [comments, setComments] = useState<CommentResponse[]>([]);
  const [newComment, setNewComment] = useState('');
  const [isSubmittingComment, setIsSubmittingComment] = useState(false);

  // Load board and comments
  useEffect(() => {
    const loadData = async () => {
      setIsLoading(true);
      setError(null);

      try {
        const [boardData, commentsData, stagesData, rolesData, importancesData] =
          await Promise.all([
            getBoard(boardId, accessToken),
            getComments(boardId, accessToken),
            getProjectStages(projectId, accessToken),
            getProjectRoles(projectId, accessToken),
            getProjectImportances(projectId, accessToken),
          ]);

        setBoard(boardData);
        setComments(commentsData);
        setStages(stagesData);
        setRoles(rolesData);
        setImportances(importancesData);

        // Set edit values
        setEditTitle(boardData.title);
        setEditContent(boardData.content || '');
        setEditStageId(boardData.stage?.id || '');
        setEditRoleIds(boardData.roles?.map((r) => r.id) || []);
        setEditImportanceId(boardData.importance?.id || '');
      } catch (err) {
        const error = err as Error;
        console.error('❌ Board 로드 실패:', error);
        setError(error.message || '보드를 불러오는데 실패했습니다.');
      } finally {
        setIsLoading(false);
      }
    };

    loadData();
  }, [boardId, projectId, accessToken]);

  // Handle board update
  const handleSave = async () => {
    if (!editTitle.trim()) {
      setError('제목은 필수입니다.');
      return;
    }

    if (editRoleIds.length === 0) {
      setError('최소 1개의 역할을 선택해야 합니다.');
      return;
    }

    setIsLoading(true);
    setError(null);

    try {
      const updated = await updateBoard(
        boardId,
        {
          title: editTitle.trim(),
          content: editContent.trim() || undefined,
          stageId: editStageId,
          roleIds: editRoleIds,
          importanceId: editImportanceId || undefined,
        },
        accessToken,
      );

      setBoard(updated);
      setIsEditing(false);
      onBoardUpdated();
    } catch (err) {
      const error = err as Error;
      console.error('❌ Board 수정 실패:', error);
      setError(error.message || '보드 수정에 실패했습니다.');
    } finally {
      setIsLoading(false);
    }
  };

  // Handle board delete
  const handleDelete = async () => {
    if (!confirm('정말 이 보드를 삭제하시겠습니까?')) {
      return;
    }

    setIsLoading(true);
    setError(null);

    try {
      await deleteBoard(boardId, accessToken);
      onBoardDeleted();
      onClose();
    } catch (err) {
      const error = err as Error;
      console.error('❌ Board 삭제 실패:', error);
      setError(error.message || '보드 삭제에 실패했습니다.');
      setIsLoading(false);
    }
  };

  // Handle comment submit
  const handleCommentSubmit = async () => {
    if (!newComment.trim()) return;

    setIsSubmittingComment(true);

    try {
      const comment = await createComment(
        {
          boardId,
          content: newComment.trim(),
        },
        accessToken,
      );

      setComments([...comments, comment]);
      setNewComment('');
    } catch (err) {
      const error = err as Error;
      console.error('❌ Comment 작성 실패:', error);
      alert('댓글 작성에 실패했습니다.');
    } finally {
      setIsSubmittingComment(false);
    }
  };

  // Toggle role selection
  const toggleRole = (roleId: string) => {
    setEditRoleIds((prev) => {
      if (prev.includes(roleId)) {
        if (prev.length === 1) {
          setError('최소 1개의 역할을 선택해야 합니다.');
          return prev;
        }
        return prev.filter((id) => id !== roleId);
      } else {
        setError(null);
        return [...prev, roleId];
      }
    });
  };

  if (isLoading && !board) {
    return (
      <div
        className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-[90]"
        onClick={onClose}
      >
        <div
          className={`w-full max-w-4xl ${theme.colors.card} p-8 ${theme.effects.borderRadius} shadow-xl`}
          onClick={(e) => e.stopPropagation()}
        >
          <div className="flex items-center justify-center">
            <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500"></div>
          </div>
        </div>
      </div>
    );
  }

  if (!board) return null;

  return (
    <div
      className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-[90]"
      onClick={onClose}
    >
      <div
        className={`w-full max-w-4xl ${theme.colors.card} ${theme.effects.borderRadius} shadow-xl max-h-[90vh] overflow-hidden flex flex-col`}
        onClick={(e) => e.stopPropagation()}
      >
        {/* Header */}
        <div className="flex items-center justify-between p-6 border-b">
          {isEditing ? (
            <input
              type="text"
              value={editTitle}
              onChange={(e) => setEditTitle(e.target.value)}
              className="flex-1 text-2xl font-bold border-b-2 border-blue-500 focus:outline-none"
              placeholder="보드 제목"
              maxLength={200}
            />
          ) : (
            <h2 className="text-2xl font-bold text-gray-800">{board.title}</h2>
          )}
          <div className="flex items-center gap-2 ml-4">
            {isEditing ? (
              <>
                <button
                  onClick={() => {
                    setIsEditing(false);
                    setEditTitle(board.title);
                    setEditContent(board.content || '');
                    setEditStageId(board.stage?.id || '');
                    setEditRoleIds(board.roles?.map((r) => r.id) || []);
                    setEditImportanceId(board.importance?.id || '');
                    setError(null);
                  }}
                  className="px-4 py-2 text-sm border border-gray-300 rounded-lg hover:bg-gray-50"
                  disabled={isLoading}
                >
                  취소
                </button>
                <button
                  onClick={handleSave}
                  className="px-4 py-2 text-sm bg-blue-500 text-white rounded-lg hover:bg-blue-600"
                  disabled={isLoading}
                >
                  {isLoading ? '저장 중...' : '저장'}
                </button>
              </>
            ) : (
              <>
                <button
                  onClick={() => setIsEditing(true)}
                  className="px-4 py-2 text-sm bg-blue-500 text-white rounded-lg hover:bg-blue-600"
                >
                  수정
                </button>
                <button
                  onClick={handleDelete}
                  className="p-2 text-red-500 hover:bg-red-50 rounded-lg"
                  disabled={isLoading}
                >
                  <Trash2 className="w-5 h-5" />
                </button>
              </>
            )}
            <button
              onClick={onClose}
              className="p-2 rounded-full hover:bg-gray-100 text-gray-500 hover:text-gray-700"
            >
              <X className="w-5 h-5" />
            </button>
          </div>
        </div>

        {/* Error Message */}
        {error && (
          <div className="mx-6 mt-4 p-3 bg-red-50 border border-red-300 rounded-lg text-red-700 text-sm">
            {error}
          </div>
        )}

        {/* Content */}
        <div className="flex-1 overflow-y-auto p-6">
          <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
            {/* Main Content */}
            <div className="lg:col-span-2 space-y-6">
              {/* Description */}
              <div>
                <h3 className="text-sm font-semibold text-gray-700 mb-2">설명</h3>
                {isEditing ? (
                  <textarea
                    value={editContent}
                    onChange={(e) => setEditContent(e.target.value)}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 resize-none"
                    rows={4}
                    placeholder="보드에 대한 설명을 입력하세요"
                    maxLength={5000}
                  />
                ) : (
                  <p className="text-gray-600 whitespace-pre-wrap">
                    {board.content || '설명이 없습니다.'}
                  </p>
                )}
              </div>

              {/* Comments */}
              <div>
                <h3 className="text-sm font-semibold text-gray-700 mb-3 flex items-center gap-2">
                  <MessageSquare className="w-4 h-4" />
                  댓글 ({comments.length})
                </h3>
                <div className="space-y-3">
                  {comments.map((comment) => (
                    <div key={comment.id} className="p-3 bg-gray-50 rounded-lg">
                      <div className="flex items-center gap-2 mb-1">
                        <div className="w-6 h-6 rounded-full bg-blue-500 flex items-center justify-center text-white text-xs font-bold">
                          {comment.userName[0]}
                        </div>
                        <span className="text-sm font-semibold text-gray-800">
                          {comment.userName}
                        </span>
                        <span className="text-xs text-gray-500">
                          {new Date(comment.createdAt).toLocaleString('ko-KR')}
                        </span>
                      </div>
                      <p className="text-sm text-gray-700 ml-8 whitespace-pre-wrap">
                        {comment.content}
                      </p>
                    </div>
                  ))}
                </div>

                {/* Comment Input */}
                <div className="mt-3 flex gap-2">
                  <input
                    type="text"
                    value={newComment}
                    onChange={(e) => setNewComment(e.target.value)}
                    onKeyPress={(e) => {
                      if (e.key === 'Enter' && !e.shiftKey) {
                        e.preventDefault();
                        handleCommentSubmit();
                      }
                    }}
                    placeholder="댓글을 입력하세요..."
                    className="flex-1 px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm"
                    disabled={isSubmittingComment}
                  />
                  <button
                    onClick={handleCommentSubmit}
                    disabled={isSubmittingComment || !newComment.trim()}
                    className="px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    <Send className="w-4 h-4" />
                  </button>
                </div>
              </div>
            </div>

            {/* Sidebar - Custom Fields */}
            <div className="space-y-4">
              {/* Stage */}
              <div>
                <label className="block text-sm font-semibold text-gray-700 mb-2">
                  <Tag className="w-4 h-4 inline mr-1" />
                  진행 단계
                </label>
                {isEditing ? (
                  <div className="space-y-2">
                    {stages.map((stage) => (
                      <button
                        key={stage.id}
                        type="button"
                        onClick={() => setEditStageId(stage.id)}
                        className={`w-full px-3 py-2 rounded-lg border-2 transition text-sm font-medium ${
                          editStageId === stage.id
                            ? 'border-blue-500 bg-blue-50 text-blue-700'
                            : 'border-gray-200 hover:border-gray-300 text-gray-700'
                        }`}
                      >
                        <span
                          className="inline-block w-3 h-3 rounded-full mr-2"
                          style={{ backgroundColor: stage.color || '#6B7280' }}
                        />
                        {stage.name}
                      </button>
                    ))}
                  </div>
                ) : (
                  <div className="flex items-center gap-2 px-3 py-2 bg-gray-50 rounded-lg">
                    <span
                      className="inline-block w-3 h-3 rounded-full"
                      style={{ backgroundColor: board.stage?.color || '#6B7280' }}
                    />
                    <span className="text-sm font-medium">{board.stage?.name}</span>
                  </div>
                )}
              </div>

              {/* Roles */}
              <div>
                <label className="block text-sm font-semibold text-gray-700 mb-2">
                  역할
                </label>
                {isEditing ? (
                  <div className="space-y-2">
                    {roles.map((role) => (
                      <button
                        key={role.id}
                        type="button"
                        onClick={() => toggleRole(role.id)}
                        className={`w-full px-3 py-2 rounded-lg border-2 transition text-sm font-medium ${
                          editRoleIds.includes(role.id)
                            ? 'border-green-500 bg-green-50 text-green-700'
                            : 'border-gray-200 hover:border-gray-300 text-gray-700'
                        }`}
                      >
                        <span
                          className="inline-block w-3 h-3 rounded-full mr-2"
                          style={{ backgroundColor: role.color || '#6B7280' }}
                        />
                        {role.name}
                        {editRoleIds.includes(role.id) && (
                          <span className="ml-2 text-green-600">✓</span>
                        )}
                      </button>
                    ))}
                  </div>
                ) : (
                  <div className="space-y-1">
                    {board.roles?.map((role) => (
                      <div
                        key={role.id}
                        className="flex items-center gap-2 px-3 py-2 bg-gray-50 rounded-lg"
                      >
                        <span
                          className="inline-block w-3 h-3 rounded-full"
                          style={{ backgroundColor: role.color || '#6B7280' }}
                        />
                        <span className="text-sm font-medium">{role.name}</span>
                      </div>
                    ))}
                  </div>
                )}
              </div>

              {/* Importance */}
              {importances.length > 0 && (
                <div>
                  <label className="block text-sm font-semibold text-gray-700 mb-2">
                    중요도
                  </label>
                  {isEditing ? (
                    <div className="space-y-2">
                      {importances.map((importance) => (
                        <button
                          key={importance.id}
                          type="button"
                          onClick={() => setEditImportanceId(importance.id)}
                          className={`w-full px-3 py-2 rounded-lg border-2 transition text-sm font-medium ${
                            editImportanceId === importance.id
                              ? 'border-orange-500 bg-orange-50 text-orange-700'
                              : 'border-gray-200 hover:border-gray-300 text-gray-700'
                          }`}
                        >
                          <span
                            className="inline-block w-3 h-3 rounded-full mr-2"
                            style={{ backgroundColor: importance.color || '#6B7280' }}
                          />
                          {importance.name}
                        </button>
                      ))}
                    </div>
                  ) : board.importance ? (
                    <div className="flex items-center gap-2 px-3 py-2 bg-gray-50 rounded-lg">
                      <span
                        className="inline-block w-3 h-3 rounded-full"
                        style={{ backgroundColor: board.importance?.color || '#6B7280' }}
                      />
                      <span className="text-sm font-medium">{board.importance?.name}</span>
                    </div>
                  ) : (
                    <p className="text-sm text-gray-500">없음</p>
                  )}
                </div>
              )}

              {/* Author & Created At */}
              <div className="pt-4 border-t space-y-2">
                <div className="flex items-center gap-2 text-sm text-gray-600">
                  <User className="w-4 h-4" />
                  <span>작성자: {board.author?.name || 'Unknown'}</span>
                </div>
                <div className="flex items-center gap-2 text-sm text-gray-600">
                  <Calendar className="w-4 h-4" />
                  <span>생성일: {new Date(board.createdAt).toLocaleDateString('ko-KR')}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default BoardDetailModal;
