import React, { useState, useEffect } from 'react';
import { X, Tag, CheckSquare, AlertCircle, Calendar, User, Plus, Settings } from 'lucide-react';
import { useTheme } from '../../contexts/ThemeContext';
import { CUSTOM_FIELD_COLORS } from '../../constants/colors';
import {
  CustomStageResponse,
  CustomRoleResponse,
  CustomImportanceResponse,
  getProjectStages,
  getProjectRoles,
  getProjectImportances,
  createBoard,
  updateBoard,
  createStage,
  createRole,
  createImportance,
} from '../../api/board/boardService';
import { WorkspaceMember, getWorkspaceMembers } from '../../api/user/userService';

interface CreateBoardModalProps {
  projectId: string;
  stageId?: string; // 컬럼에서 열었을 때 미리 선택된 stageId
  editData?: {
    boardId: string;
    projectId: string;
    title: string;
    content: string;
    stageId: string;
    roleId: string;
    importanceId: string;
    assigneeIds: string[];
    dueDate: string;
  } | null;
  workspaceId: string;
  onClose: () => void;
  onBoardCreated: () => void;
}

export const CreateBoardModal: React.FC<CreateBoardModalProps> = ({
  projectId,
  stageId: initialStageId,
  editData,
  workspaceId,
  onClose,
  onBoardCreated,
}) => {
  const { theme } = useTheme();
  const accessToken = localStorage.getItem('access_token') || '';

  // Form state
  const [title, setTitle] = useState(editData?.title || '');
  const [content, setContent] = useState(editData?.content || '');
  const [selectedStageId, setSelectedStageId] = useState(editData?.stageId || initialStageId || '');
  const [selectedRoleId, setSelectedRoleId] = useState<string>(editData?.roleId || '');
  const [selectedImportanceId, setSelectedImportanceId] = useState<string>(editData?.importanceId || '');
  const [selectedAssigneeIds, setSelectedAssigneeIds] = useState<string[]>(editData?.assigneeIds || []);
  const [dueDate, setDueDate] = useState<string>(editData?.dueDate || '');

  // Assignee search state
  const [assigneeSearch, setAssigneeSearch] = useState('');

  // Data state
  const [stages, setStages] = useState<CustomStageResponse[]>([]);
  const [roles, setRoles] = useState<CustomRoleResponse[]>([]);
  const [importances, setImportances] = useState<CustomImportanceResponse[]>([]);
  const [workspaceMembers, setWorkspaceMembers] = useState<WorkspaceMember[]>([]);

  // UI state
  const [isLoading, setIsLoading] = useState(false);
  const [isLoadingFields, setIsLoadingFields] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Inline creation state
  const [showCreateStage, setShowCreateStage] = useState(false);
  const [showCreateRole, setShowCreateRole] = useState(false);
  const [showCreateImportance, setShowCreateImportance] = useState(false);
  const [newFieldName, setNewFieldName] = useState('');
  const [newFieldColor, setNewFieldColor] = useState(CUSTOM_FIELD_COLORS[0].hex);
  const [newImportanceLevel, setNewImportanceLevel] = useState(1);

  // Dropdown states
  const [showRoleDropdown, setShowRoleDropdown] = useState(false);
  const [showStageDropdown, setShowStageDropdown] = useState(false);
  const [showImportanceDropdown, setShowImportanceDropdown] = useState(false);

  // 1. Custom Fields 조회
  useEffect(() => {
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

        // 기본값 설정
        if (!selectedStageId && stagesData.length > 0) {
          setSelectedStageId(stagesData[0].id);
        }
        if (rolesData.length > 0) {
          setSelectedRoleId(rolesData[0].id);
        }
        // Importance는 선택 사항이므로 기본값 없음

        console.log('✅ Custom Fields 로드:', {
          stages: stagesData.length,
          roles: rolesData.length,
          importances: importancesData.length,
        });
      } catch (err) {
        console.error('❌ Custom Fields 로드 실패:', err);
        setError('커스텀 필드를 불러오는데 실패했습니다.');
      } finally {
        setIsLoadingFields(false);
      }
    };

    fetchCustomFields();
  }, [projectId, accessToken]);

  // 1.2 워크스페이스 멤버 조회
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

  // 1.3 드롭다운 외부 클릭 감지
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      const target = event.target as HTMLElement;
      // 드롭다운 버튼이나 메뉴 내부 클릭이 아닌 경우 드롭다운 닫기
      if (!target.closest('.role-dropdown-container')) {
        setShowRoleDropdown(false);
      }
      if (!target.closest('.stage-dropdown-container')) {
        setShowStageDropdown(false);
      }
      if (!target.closest('.importance-dropdown-container')) {
        setShowImportanceDropdown(false);
      }
      if (!target.closest('.assignee-dropdown-container')) {
        setAssigneeSearch(''); // 검색어 비우기
      }
    };

    if (showRoleDropdown || showStageDropdown || showImportanceDropdown || assigneeSearch.trim()) {
      document.addEventListener('mousedown', handleClickOutside);
    }

    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, [showRoleDropdown, showStageDropdown, showImportanceDropdown, assigneeSearch]);

  // 2. Inline custom field creation handlers
  const handleCreateCustomField = async (type: 'stage' | 'role' | 'importance') => {
    if (!newFieldName.trim()) {
      setError('이름을 입력해주세요.');
      return;
    }

    setIsLoading(true);
    try {
      let newField:
        | CustomStageResponse
        | CustomRoleResponse
        | CustomImportanceResponse
        | undefined;

      if (type === 'stage') {
        newField = await createStage(
          { projectId, name: newFieldName.trim(), color: newFieldColor },
          accessToken,
        );
        setStages([...stages, newField as CustomStageResponse]);
        setSelectedStageId(newField.id);
        setShowCreateStage(false);
      } else if (type === 'role') {
        newField = await createRole(
          { projectId, name: newFieldName.trim(), color: newFieldColor },
          accessToken,
        );
        setRoles([...roles, newField as CustomRoleResponse]);
        setSelectedRoleId(newField.id);
        setShowCreateRole(false);
      } else if (type === 'importance') {
        newField = await createImportance(
          { projectId, name: newFieldName.trim(), color: newFieldColor, level: newImportanceLevel },
          accessToken,
        );
        setImportances([...importances, newField as CustomImportanceResponse]);
        setSelectedImportanceId(newField.id);
        setShowCreateImportance(false);
      }

      // Reset form
      setNewFieldName('');
      setNewFieldColor(CUSTOM_FIELD_COLORS[0].hex);
      setNewImportanceLevel(1);
      setError(null);

      console.log(`✅ ${type} 생성 성공:`, newField);
    } catch (err) {
      const error = err as Error;
      console.error(`❌ ${type} 생성 실패:`, error);
      setError(error.message || '커스텀 필드 생성에 실패했습니다.');
    } finally {
      setIsLoading(false);
    }
  };

  const cancelInlineCreation = () => {
    setShowCreateStage(false);
    setShowCreateRole(false);
    setShowCreateImportance(false);
    setNewFieldName('');
    setNewFieldColor(CUSTOM_FIELD_COLORS[0].hex);
    setNewImportanceLevel(1);
    setError(null);
  };

  // 3. 제출 핸들러
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    // Validation
    if (!title.trim()) {
      setError('보드 제목은 필수입니다.');
      return;
    }
    if (!selectedStageId) {
      setError('진행 단계를 선택해주세요.');
      return;
    }
    if (!selectedRoleId) {
      setError('역할을 선택해주세요.');
      return;
    }

    setIsLoading(true);
    setError(null);

    try {
      const boardData = {
        projectId,
        title: title.trim(),
        content: content.trim() || undefined,
        stageId: selectedStageId,
        roleIds: [selectedRoleId],
        importanceId: selectedImportanceId || undefined,
        assigneeIds: selectedAssigneeIds.length > 0 ? selectedAssigneeIds : undefined,
        dueDate: dueDate || undefined,
      };

      if (editData) {
        // 수정 모드
        await updateBoard(editData.boardId, boardData, accessToken);
        console.log('✅ 보드 수정 성공:', title);
      } else {
        // 생성 모드
        await createBoard(boardData, accessToken);
        console.log('✅ 보드 생성 성공:', title);
      }

      onBoardCreated();
      onClose();
    } catch (err) {
      const error = err as Error;
      console.error(`❌ 보드 ${editData ? '수정' : '생성'} 실패:`, error);
      setError(error.message || `보드 ${editData ? '수정' : '생성'}에 실패했습니다.`);
    } finally {
      setIsLoading(false);
    }
  };

  // Helper: Color Picker Component
  const renderColorPicker = (selectedColor: string, onColorChange: (color: string) => void) => (
    <div className="grid grid-cols-6 gap-2 mt-2">
      {CUSTOM_FIELD_COLORS.map((color) => (
        <button
          key={color.hex}
          type="button"
          className={`w-8 h-8 rounded-md border-2 transition-all ${
            selectedColor === color.hex
              ? 'border-gray-800 ring-2 ring-blue-500 scale-110'
              : 'border-gray-300 hover:scale-105'
          }`}
          style={{ backgroundColor: color.hex }}
          onClick={() => onColorChange(color.hex)}
          title={color.name}
          disabled={isLoading}
        />
      ))}
    </div>
  );

  // Helper: Creation Modal (작은 모달로 표시)
  const renderCreationModal = (type: 'stage' | 'role' | 'importance', title: string) => (
    <div
      className="fixed inset-0 bg-black bg-opacity-30 flex items-center justify-center z-[100]"
      onClick={cancelInlineCreation}
    >
      <div
        className="bg-white rounded-lg shadow-xl p-6 w-full max-w-md"
        onClick={(e) => e.stopPropagation()}
      >
        <h3 className="text-lg font-bold text-gray-800 mb-4">새 {title} 추가</h3>
        <div className="space-y-4">
          <div>
            <label className="text-sm font-semibold text-gray-700 block mb-2">
              {title} 이름 <span className="text-red-500">*</span>
            </label>
            <input
              type="text"
              value={newFieldName}
              onChange={(e) => setNewFieldName(e.target.value)}
              placeholder={`예: ${type === 'stage' ? '진행중' : type === 'role' ? '디자이너' : '매우 높음'}`}
              className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm"
              disabled={isLoading}
              autoFocus
            />
          </div>
          <div>
            <label className="text-sm font-semibold text-gray-700 block mb-2">색상 선택</label>
            {renderColorPicker(newFieldColor, setNewFieldColor)}
          </div>
          {type === 'importance' && (
            <div>
              <label className="text-sm font-semibold text-gray-700 block mb-2">
                중요도 레벨 (1-5)
              </label>
              <input
                type="number"
                min="1"
                max="5"
                value={newImportanceLevel}
                onChange={(e) => setNewImportanceLevel(parseInt(e.target.value))}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm"
                disabled={isLoading}
              />
            </div>
          )}
          <div className="flex gap-3 pt-2">
            <button
              type="button"
              onClick={cancelInlineCreation}
              className="flex-1 px-4 py-2 border border-gray-300 text-gray-700 font-semibold rounded-lg hover:bg-gray-50 transition"
              disabled={isLoading}
            >
              취소
            </button>
            <button
              type="button"
              onClick={() => handleCreateCustomField(type)}
              className="flex-1 px-4 py-2 bg-blue-500 text-white font-semibold rounded-lg hover:bg-blue-600 transition"
              disabled={isLoading || !newFieldName.trim()}
            >
              {isLoading ? '추가 중...' : '추가'}
            </button>
          </div>
        </div>
      </div>
    </div>
  );

  return (
    <div
      className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-[90]"
      onClick={onClose}
    >
      <div
        className={`relative w-full max-w-2xl ${theme.colors.card} ${theme.effects.borderRadius} shadow-xl max-h-[90vh] flex flex-col overflow-hidden`}
        onClick={(e) => e.stopPropagation()}
      >
        {/* Header */}
        <div className="flex items-center justify-between px-6 pt-6 pb-4 border-b flex-shrink-0">
          <h2 className="text-xl font-bold text-gray-800">{editData ? '보드 수정' : '새 보드 만들기'}</h2>
          <button
            onClick={onClose}
            className="p-2 rounded-full hover:bg-gray-100 text-gray-500 hover:text-gray-700 transition"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Scrollable Content Area */}
        <div className="flex-1 overflow-y-auto px-6">
          {/* Error Message */}
          {error && (
            <div className="mt-4 mb-4 p-3 bg-red-50 border border-red-300 rounded-lg text-red-700 text-sm">
              {error}
            </div>
          )}

          {/* Loading State */}
          {isLoadingFields ? (
            <div className="flex items-center justify-center py-12">
              <div className="text-center">
                <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500 mx-auto mb-4"></div>
                <p className="text-gray-600">커스텀 필드를 불러오는 중...</p>
              </div>
            </div>
          ) : (
            <form onSubmit={handleSubmit} className="space-y-4 pb-4">
            {/* Title */}
            <div>
              <label className="block text-sm font-semibold text-gray-700 mb-2">
                보드 제목 <span className="text-red-500">*</span>
              </label>
              <input
                type="text"
                value={title}
                onChange={(e) => setTitle(e.target.value)}
                placeholder="예: 사용자 인증 API 구현"
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm"
                disabled={isLoading}
                maxLength={200}
              />
            </div>

            {/* Content */}
            <div>
              <label className="block text-sm font-semibold text-gray-700 mb-2">
                설명 (선택)
              </label>
              <textarea
                value={content}
                onChange={(e) => setContent(e.target.value)}
                placeholder="보드에 대한 자세한 설명을 입력하세요"
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm resize-none"
                rows={3}
                disabled={isLoading}
                maxLength={5000}
              />
            </div>

            {/* Stage and Role Selection */}
            <div className="grid grid-cols-2 gap-4">
              {/* Stage Selection */}
              <div className="relative stage-dropdown-container">
                <label className="block text-sm font-semibold text-gray-700 mb-2">
                  <CheckSquare className="w-4 h-4 inline mr-1" />
                  진행 단계 <span className="text-red-500">*</span>
                </label>
                <button
                  type="button"
                  onClick={() => setShowStageDropdown(!showStageDropdown)}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg bg-white hover:bg-gray-50 transition text-sm text-left flex items-center justify-between"
                  disabled={isLoading}
                >
                  <span className="flex items-center gap-2">
                    {selectedStageId && stages.find((s) => s.id === selectedStageId) && (
                      <>
                        <span
                          className="w-3 h-3 rounded-full"
                          style={{
                            backgroundColor:
                              stages.find((s) => s.id === selectedStageId)?.color || '#6B7280',
                          }}
                        />
                        {stages.find((s) => s.id === selectedStageId)?.name}
                      </>
                    )}
                  </span>
                  <CheckSquare className="w-4 h-4 text-gray-400" />
                </button>
                {/* 드롭다운 메뉴 */}
                {showStageDropdown && (
                  <div className="absolute z-10 w-full mt-1 bg-white border border-gray-300 rounded-lg shadow-lg max-h-48 overflow-y-auto">
                    {stages.map((stage) => (
                      <button
                        key={stage.id}
                        type="button"
                        onClick={() => {
                          setSelectedStageId(stage.id);
                          setShowStageDropdown(false);
                        }}
                        className={`w-full px-3 py-2 text-left hover:bg-gray-100 transition text-sm flex items-center gap-2 ${
                          selectedStageId === stage.id ? 'bg-blue-50' : ''
                        }`}
                      >
                        <span
                          className="w-3 h-3 rounded-full"
                          style={{ backgroundColor: stage.color || '#6B7280' }}
                        />
                        {stage.name}
                      </button>
                    ))}
                    <button
                      type="button"
                      onClick={() => {
                        setShowStageDropdown(false);
                        setShowCreateStage(true);
                      }}
                      className="w-full px-3 py-2 text-left hover:bg-blue-50 transition text-sm text-blue-600 font-medium border-t border-gray-200 flex items-center gap-2"
                    >
                      <Plus className="w-4 h-4" />+ 새 진행 단계 추가
                    </button>
                  </div>
                )}
              </div>

              {/* Role Selection */}
              <div className="relative role-dropdown-container">
                <label className="block text-sm font-semibold text-gray-700 mb-2">
                  <Tag className="w-4 h-4 inline mr-1" />
                  역할 <span className="text-red-500">*</span>
                </label>
                <button
                  type="button"
                  onClick={() => setShowRoleDropdown(!showRoleDropdown)}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg bg-white hover:bg-gray-50 transition text-sm text-left flex items-center justify-between"
                  disabled={isLoading}
                >
                  <span className="flex items-center gap-2">
                    {selectedRoleId && roles.find((r) => r.id === selectedRoleId) && (
                      <>
                        <span
                          className="w-3 h-3 rounded-full"
                          style={{
                            backgroundColor:
                              roles.find((r) => r.id === selectedRoleId)?.color || '#6B7280',
                          }}
                        />
                        {roles.find((r) => r.id === selectedRoleId)?.name}
                      </>
                    )}
                  </span>
                  <Tag className="w-4 h-4 text-gray-400" />
                </button>
                {/* 드롭다운 메뉴 */}
                {showRoleDropdown && (
                  <div className="absolute z-10 w-full mt-1 bg-white border border-gray-300 rounded-lg shadow-lg max-h-48 overflow-y-auto">
                    {roles.map((role) => (
                      <button
                        key={role.id}
                        type="button"
                        onClick={() => {
                          setSelectedRoleId(role.id);
                          setShowRoleDropdown(false);
                        }}
                        className={`w-full px-3 py-2 text-left hover:bg-gray-100 transition text-sm flex items-center gap-2 ${
                          selectedRoleId === role.id ? 'bg-blue-50' : ''
                        }`}
                      >
                        <span
                          className="w-3 h-3 rounded-full"
                          style={{ backgroundColor: role.color || '#6B7280' }}
                        />
                        {role.name}
                      </button>
                    ))}
                    <button
                      type="button"
                      onClick={() => {
                        setShowRoleDropdown(false);
                        setShowCreateRole(true);
                      }}
                      className="w-full px-3 py-2 text-left hover:bg-blue-50 transition text-sm text-blue-600 font-medium border-t border-gray-200 flex items-center gap-2"
                    >
                      <Plus className="w-4 h-4" />+ 새 역할 추가
                    </button>
                  </div>
                )}
              </div>
            </div>

            {/* Importance and Field Management */}
            <div className="grid grid-cols-2 gap-4">
              {/* Importance Selection */}
              <div className="relative importance-dropdown-container">
                <label className="block text-sm font-semibold text-gray-700 mb-2">
                  <AlertCircle className="w-4 h-4 inline mr-1" />
                  중요도 (선택)
                </label>
                <button
                  type="button"
                  onClick={() => setShowImportanceDropdown(!showImportanceDropdown)}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg bg-white hover:bg-gray-50 transition text-sm text-left flex items-center justify-between"
                  disabled={isLoading}
                >
                  <span className="flex items-center gap-2">
                    {selectedImportanceId ? (
                      importances.find((i) => i.id === selectedImportanceId) && (
                        <>
                          <span
                            className="w-3 h-3 rounded-full"
                            style={{
                              backgroundColor:
                                importances.find((i) => i.id === selectedImportanceId)?.color ||
                                '#6B7280',
                            }}
                          />
                          {importances.find((i) => i.id === selectedImportanceId)?.name}
                          {'level' in (importances.find((i) => i.id === selectedImportanceId) || {})
                            ? ` (Lv.${(importances.find((i) => i.id === selectedImportanceId) as any).level})`
                            : ''}
                        </>
                      )
                    ) : (
                      <span className="text-gray-500">없음</span>
                    )}
                  </span>
                  <AlertCircle className="w-4 h-4 text-gray-400" />
                </button>
                {/* 드롭다운 메뉴 */}
                {showImportanceDropdown && (
                  <div className="absolute z-10 w-full mt-1 bg-white border border-gray-300 rounded-lg shadow-lg max-h-48 overflow-y-auto">
                    <button
                      type="button"
                      onClick={() => {
                        setSelectedImportanceId('');
                        setShowImportanceDropdown(false);
                      }}
                      className={`w-full px-3 py-2 text-left hover:bg-gray-100 transition text-sm flex items-center gap-2 ${
                        selectedImportanceId === '' ? 'bg-blue-50' : ''
                      }`}
                    >
                      <span className="w-3 h-3 rounded-full bg-gray-300" />
                      없음
                    </button>
                    {importances.map((importance) => (
                      <button
                        key={importance.id}
                        type="button"
                        onClick={() => {
                          setSelectedImportanceId(importance.id);
                          setShowImportanceDropdown(false);
                        }}
                        className={`w-full px-3 py-2 text-left hover:bg-gray-100 transition text-sm flex items-center gap-2 ${
                          selectedImportanceId === importance.id ? 'bg-blue-50' : ''
                        }`}
                      >
                        <span
                          className="w-3 h-3 rounded-full"
                          style={{ backgroundColor: importance.color || '#6B7280' }}
                        />
                        {importance.name}
                        {'level' in importance && (
                          <span className="text-xs text-gray-500">Lv.{importance.level}</span>
                        )}
                      </button>
                    ))}
                    <button
                      type="button"
                      onClick={() => {
                        setShowImportanceDropdown(false);
                        setShowCreateImportance(true);
                      }}
                      className="w-full px-3 py-2 text-left hover:bg-blue-50 transition text-sm text-blue-600 font-medium border-t border-gray-200 flex items-center gap-2"
                    >
                      <Plus className="w-4 h-4" />+ 새 중요도 추가
                    </button>
                  </div>
                )}
              </div>

              {/* Field Management */}
              <div>
                <label className="block text-sm font-semibold text-gray-700 mb-2">
                  <Settings className="w-4 h-4 inline mr-1" />
                  필드 관리
                </label>
                <button
                  type="button"
                  onClick={() => {
                    // Open CustomFieldManageModal
                    // This would need to be implemented in Dashboard or parent component
                    console.log('Open field management modal');
                  }}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg bg-white hover:bg-gray-50 transition text-sm text-left flex items-center justify-between"
                >
                  <span className="text-gray-600">커스텀 필드 관리</span>
                  <Settings className="w-4 h-4 text-gray-400" />
                </button>
              </div>
            </div>

            {/* Assignee and Due Date */}
            <div className="grid grid-cols-2 gap-4">
              {/* Assignee - Multi Select */}
              <div className="relative assignee-dropdown-container">
                <label className="block text-sm font-semibold text-gray-700 mb-2">
                  <User className="w-4 h-4 inline mr-1" />
                  담당자 (선택)
                </label>

                {/* Input with Tags Inside */}
                <div className="w-full min-h-[42px] px-2 py-1 border border-gray-300 rounded-lg focus-within:ring-2 focus-within:ring-blue-500 bg-white flex flex-wrap items-center gap-1">
                  {/* Selected Assignees Tags Inside Input */}
                  {selectedAssigneeIds.map((userId) => {
                    const member = workspaceMembers.find((m) => m.userId === userId);
                    return (
                      <span
                        key={userId}
                        className="inline-flex items-center gap-1 px-2 py-1 bg-blue-100 text-blue-700 text-xs rounded-full"
                      >
                        {member?.name || userId}
                        <button
                          type="button"
                          onClick={() => {
                            setSelectedAssigneeIds(selectedAssigneeIds.filter((id) => id !== userId));
                          }}
                          className="hover:text-blue-900"
                        >
                          <X className="w-3 h-3" />
                        </button>
                      </span>
                    );
                  })}

                  {/* Search Input */}
                  <input
                    type="text"
                    value={assigneeSearch}
                    onChange={(e) => {
                      setAssigneeSearch(e.target.value);
                    }}
                    placeholder={selectedAssigneeIds.length === 0 ? "담당자 검색..." : ""}
                    className="flex-1 min-w-[120px] px-1 py-1 text-sm focus:outline-none"
                    disabled={isLoading}
                  />
                </div>

                {/* Dropdown - z-index higher than modal, only show when searching */}
                {assigneeSearch.trim() && (
                  <div className="absolute z-[110] w-full mt-1 bg-white border border-gray-300 rounded-lg shadow-lg max-h-48 overflow-y-auto">
                    {workspaceMembers
                      .filter((member) =>
                        member.name.toLowerCase().includes(assigneeSearch.toLowerCase()) ||
                        member.email.toLowerCase().includes(assigneeSearch.toLowerCase())
                      )
                      .map((member) => {
                        const isSelected = selectedAssigneeIds.includes(member.userId);
                        return (
                          <button
                            key={member.userId}
                            type="button"
                            onClick={() => {
                              if (isSelected) {
                                setSelectedAssigneeIds(selectedAssigneeIds.filter((id) => id !== member.userId));
                              } else {
                                setSelectedAssigneeIds([...selectedAssigneeIds, member.userId]);
                              }
                              setAssigneeSearch('');
                            }}
                            className={`w-full px-3 py-2 text-left text-sm hover:bg-gray-100 flex items-center justify-between ${
                              isSelected ? 'bg-blue-50' : ''
                            }`}
                          >
                            <div>
                              <div className="font-medium">{member.name}</div>
                              <div className="text-xs text-gray-500">{member.email}</div>
                            </div>
                            {isSelected && (
                              <CheckSquare className="w-4 h-4 text-blue-600" />
                            )}
                          </button>
                        );
                      })}
                    {workspaceMembers.filter((member) =>
                      member.name.toLowerCase().includes(assigneeSearch.toLowerCase()) ||
                      member.email.toLowerCase().includes(assigneeSearch.toLowerCase())
                    ).length === 0 && (
                      <div className="px-3 py-2 text-sm text-gray-500 text-center">
                        검색 결과가 없습니다
                      </div>
                    )}
                  </div>
                )}
              </div>

              {/* Due Date */}
              <div>
                <label className="block text-sm font-semibold text-gray-700 mb-2">
                  <Calendar className="w-4 h-4 inline mr-1" />
                  마감일 (선택)
                </label>
                <input
                  type="date"
                  value={dueDate}
                  onChange={(e) => setDueDate(e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm"
                  disabled={isLoading}
                />
              </div>
            </div>

            {/* Actions */}
            <div className="flex gap-3 pt-4 border-t sticky bottom-0 bg-white">
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
                {isLoading ? (editData ? '수정 중...' : '생성 중...') : (editData ? '보드 수정' : '보드 만들기')}
              </button>
            </div>
          </form>
          )}
        </div>
      </div>

      {/* Creation Modals */}
      {showCreateStage && renderCreationModal('stage', '진행 단계')}
      {showCreateRole && renderCreationModal('role', '역할')}
      {showCreateImportance && renderCreationModal('importance', '중요도')}
    </div>
  );
};
