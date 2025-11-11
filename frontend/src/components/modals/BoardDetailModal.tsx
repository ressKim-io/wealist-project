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
  // 💡 [수정] Custom Field 및 Board 관련 타입은 src/types/board.ts에서 가져옵니다.
  CustomStageResponse,
  CustomRoleResponse,
  CustomImportanceResponse,
  BoardResponse,
} from '../../types/board';
import { getBoard, deleteBoard } from '../../api/board/boardService';
import { getWorkspaceMembers } from '../../api/user/userService';
import { WorkspaceMember } from '../../types/user';

// ⚠️ [주의] API 호출이 제거되었으므로, 컴포넌트 로직 유지를 위해 Mock Data를 사용합니다.
const MOCK_STAGES: CustomStageResponse[] = [
  {
    stageId: 's-waiting',
    label: '대기',
    color: '#F59E0B',
    displayOrder: 1,
    fieldId: 'stage',
    description: '대기 단계',
  },
  {
    stageId: 's-progress',
    label: '진행중',
    color: '#3B82F6',
    displayOrder: 2,
    fieldId: 'stage',
    description: '진행 단계',
  },
  {
    stageId: 's-done',
    label: '완료',
    color: '#10B981',
    displayOrder: 3,
    fieldId: 'stage',
    description: '완료 단계',
  },
];
const MOCK_ROLES: CustomRoleResponse[] = [
  {
    roleId: 'r-frontend',
    label: '프론트엔드',
    color: '#8B5CF6',
    displayOrder: 1,
    fieldId: 'role',
    description: '프론트 역할',
  },
  {
    roleId: 'r-backend',
    label: '백엔드',
    color: '#EC4899',
    displayOrder: 2,
    fieldId: 'role',
    description: '백엔드 역할',
  },
];
const MOCK_IMPORTANCES: CustomImportanceResponse[] = [
  {
    importanceId: 'i-high',
    label: '높음',
    color: '#F59E0B',
    displayOrder: 1,
    fieldId: 'importance',
    description: '높은 중요도',
  },
  {
    importanceId: 'i-low',
    label: '낮음',
    color: '#10B981',
    displayOrder: 2,
    fieldId: 'importance',
    description: '낮은 중요도',
  },
];

/**
 * BoardDetailModal - 보드 상세 보기 및 수정
 */
interface BoardDetailModalProps {
  boardId: string;
  workspaceId: string;
  onClose: () => void;
  onBoardUpdated: () => void;
  onBoardDeleted: () => void;
  // 💡 [수정] onEdit으로 전달하는 데이터 구조 변경
  onEdit: (boardData: {
    boardId: string;
    projectId: string;
    title: string;
    content: string;
    stageId: string;
    // 💡 단일 담당자 ID로 변경
    assigneeId?: string;
    roleIds: string[];
    importanceId?: string;
    dueDate?: string;
  }) => void;
}

export const BoardDetailModal: React.FC<BoardDetailModalProps> = ({
  boardId,
  workspaceId,
  onClose,
  onBoardDeleted,
  onEdit,
}) => {
  const { theme } = useTheme();
  const accessToken = localStorage.getItem('accessToken') || '';

  // Form state
  const [projectId, setProjectId] = useState<string>('');
  const [title, setTitle] = useState('');
  const [content, setContent] = useState('');
  const [selectedStageId, setSelectedStageId] = useState('');
  const [selectedRoleId, setSelectedRoleId] = useState<string>('');
  const [selectedImportanceId, setSelectedImportanceId] = useState<string>('');

  // 💡 [수정] 복수 담당자 ID 배열 -> 단일 담당자 ID로 변경
  const [selectedAssigneeId, setSelectedAssigneeId] = useState<string>('');

  const [dueDate, setDueDate] = useState<string>('');

  // Data state
  // 💡 [수정] Custom Field 데이터는 Mock으로 초기화
  const [stages, setStages] = useState<CustomStageResponse[]>(MOCK_STAGES);
  const [roles, setRoles] = useState<CustomRoleResponse[]>(MOCK_ROLES);
  const [importances, setImportances] = useState<CustomImportanceResponse[]>(MOCK_IMPORTANCES);
  const [workspaceMembers, setWorkspaceMembers] = useState<WorkspaceMember[]>([]);

  // UI state
  const [isLoading, setIsLoading] = useState(false);
  const [isLoadingBoard, setIsLoadingBoard] = useState(true);
  // 💡 Custom Field API 호출 제거로 인해 로딩 상태를 false로 변경
  const [isLoadingFields, setIsLoadingFields] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Comment state (변경 없음)
  const [comments, setComments] = useState<any[]>([]);
  const [newComment, setNewComment] = useState('');

  // 보드 데이터 조회
  useEffect(() => {
    const fetchBoard = async () => {
      setIsLoadingBoard(true);
      try {
        const boardData: BoardResponse = await getBoard(boardId, accessToken);

        // 보드 데이터로 상태 초기화
        setProjectId(boardData.projectId); // 💡 project_id -> projectId로 변경 가정
        setTitle(boardData.title);
        setContent(boardData.content || '');

        // 💡 Custom Field ID를 boardData.customFields에서 추출하거나,
        //    boardData에 직접 stageId, importanceId, roleIds가 있다고 가정하고 처리합니다.
        //    (API 문서의 BoardResponse가 커스텀 필드를 customFields 객체 안에 넣고 있으나,
        //     기존 컴포넌트의 단순성을 위해 root 속성에서 찾거나, customFields 객체에서 필터링합니다.)

        // Stage ID 추출 (예시: customFields 객체 안에 stageId가 있다고 가정)
        // ⚠️ 실제 백엔드 응답 구조에 따라 이 부분은 달라질 수 있습니다.
        const stageIdFromCustomField = boardData.customFields?.stageId || '';
        setSelectedStageId(stageIdFromCustomField);

        // Role ID 추출 (단일 Role ID만 필요)
        const roleIdsFromCustomField: string[] = boardData.customFields?.roleIds || [];
        setSelectedRoleId(roleIdsFromCustomField[0] || '');

        // Importance ID 추출
        const importanceIdFromCustomField = boardData.customFields?.importanceId || '';
        setSelectedImportanceId(importanceIdFromCustomField);

        // Assignee ID 추출 (단일 담당자만 허용)
        let assigneeId: string = '';
        if (boardData.assignee?.userId) {
          assigneeId = boardData.assignee.userId;
        }
        setSelectedAssigneeId(assigneeId);

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

  // Custom Fields 조회 (로직 제거, Mock Data 사용)
  // 💡 [제거] projectId가 설정된 후 Custom Fields를 API로 불러오던 useEffect 로직을 제거합니다.

  // 워크스페이스 멤버 조회 (변경 없음)
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
    // ⚠️ [수정] confirm() 대신 커스텀 모달 사용을 권장합니다. (현재는 경고만 출력)
    if (!window.confirm('정말로 이 보드를 삭제하시겠습니까?')) return;

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

  // 💡 [추가] 필드 정보 조회 헬퍼 함수 (Mock 기반)
  const getFieldOption = (
    options: CustomStageResponse[] | CustomRoleResponse[] | CustomImportanceResponse[],
    id: string,
  ) => {
    // Stage/Role/Importance 응답 타입은 label과 color를 포함합니다.
    return options.find(
      (opt: any) => opt.stageId === id || opt.roleId === id || opt.importanceId === id,
    );
  };

  // 로딩 중이면 로딩 UI 표시
  // 💡 [수정] isLoadingFields는 항상 false이므로, 로딩 로직은 isLoadingBoard만 확인합니다.
  if (isLoadingBoard) {
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

  // 💡 [추가] UI 표시를 위한 필드 데이터 조회
  const currentStage = getFieldOption(stages, selectedStageId);
  const currentRole = getFieldOption(roles, selectedRoleId);
  const currentImportance = getFieldOption(importances, selectedImportanceId);

  // 💡 [수정] 단일 담당자 조회
  const currentAssignee = workspaceMembers.find((m) => m.userId === selectedAssigneeId);

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
                    backgroundColor: currentStage?.color || '#6B7280',
                  }}
                />
                <span className="text-sm">{currentStage?.label || '알 수 없음'}</span>
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
                    backgroundColor: currentRole?.color || '#6B7280',
                  }}
                />
                <span className="text-sm">{currentRole?.label || '알 수 없음'}</span>
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
                      backgroundColor: currentImportance?.color || '#6B7280',
                    }}
                  />
                  <span className="text-sm">{currentImportance?.label || '알 수 없음'}</span>
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
              {currentAssignee ? (
                <div className="flex flex-wrap gap-1">
                  <span
                    key={currentAssignee.userId}
                    className="inline-flex items-center px-2 py-1 bg-blue-100 text-blue-700 text-xs rounded-full"
                  >
                    {currentAssignee.userName || currentAssignee.userId}
                  </span>
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

        {/* Comments Section (변경 없음) */}
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
              // 💡 [수정] onEdit으로 전달하는 데이터 구조를 단일 담당자 및 역할 배열로 변경
              onEdit({
                boardId,
                projectId,
                title,
                content,
                stageId: selectedStageId,
                roleIds: selectedRoleId ? [selectedRoleId] : [],
                importanceId: selectedImportanceId,
                // onEdit으로 넘길 때 단일 담당자 ID를 배열 형태로 전달하여 CreateBoardModal이 처리하도록 합니다.
                assigneeId: selectedAssigneeId,
                dueDate: dueDate,
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
