import React, { useState, useEffect } from 'react';
import { X, Tag, CheckSquare, AlertCircle, Calendar, User, Plus } from 'lucide-react';
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
  createStage,
  createRole,
  createImportance,
} from '../../api/board/boardService';

interface CreateBoardModalProps {
  projectId: string;
  stageId?: string; // 컬럼에서 열었을 때 미리 선택된 stageId
  onClose: () => void;
  onBoardCreated: () => void;
}

export const CreateBoardModal: React.FC<CreateBoardModalProps> = ({
  projectId,
  stageId: initialStageId,
  onClose,
  onBoardCreated,
}) => {
  const { theme } = useTheme();
  const accessToken = localStorage.getItem('access_token') || '';

  // Form state
  const [title, setTitle] = useState('');
  const [content, setContent] = useState('');
  const [selectedStageId, setSelectedStageId] = useState(initialStageId || '');
  const [selectedRoleIds, setSelectedRoleIds] = useState<string[]>([]);
  const [selectedImportanceId, setSelectedImportanceId] = useState<string>('');
  const [assigneeId, setAssigneeId] = useState<string>('');
  const [dueDate, setDueDate] = useState<string>('');

  // Data state
  const [stages, setStages] = useState<CustomStageResponse[]>([]);
  const [roles, setRoles] = useState<CustomRoleResponse[]>([]);
  const [importances, setImportances] = useState<CustomImportanceResponse[]>([]);

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

  // Role dropdown state
  const [showRoleDropdown, setShowRoleDropdown] = useState(false);

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
          setSelectedRoleIds([rolesData[0].id]);
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
  }, [projectId, accessToken, selectedStageId]);

  // 1.1 드롭다운 외부 클릭 감지
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      const target = event.target as HTMLElement;
      // 드롭다운 버튼이나 메뉴 내부 클릭이 아닌 경우 드롭다운 닫기
      if (!target.closest('.role-dropdown-container')) {
        setShowRoleDropdown(false);
      }
    };

    if (showRoleDropdown) {
      document.addEventListener('mousedown', handleClickOutside);
    }

    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, [showRoleDropdown]);

  // 2. Role 토글 핸들러
  const toggleRole = (roleId: string) => {
    setSelectedRoleIds((prev) => {
      if (prev.includes(roleId)) {
        // 최소 1개는 선택되어야 함
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

  // 2.1 Inline custom field creation handlers
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
        setSelectedRoleIds([...selectedRoleIds, newField.id]);
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
    if (selectedRoleIds.length === 0) {
      setError('최소 1개의 역할을 선택해야 합니다.');
      return;
    }

    setIsLoading(true);
    setError(null);

    try {
      await createBoard(
        {
          projectId,
          title: title.trim(),
          content: content.trim() || undefined,
          stageId: selectedStageId,
          roleIds: selectedRoleIds,
          importanceId: selectedImportanceId || undefined,
          assigneeId: assigneeId || undefined,
          dueDate: dueDate || undefined,
        },
        accessToken,
      );

      console.log('✅ 보드 생성 성공:', title);
      onBoardCreated();
      onClose();
    } catch (err) {
      const error = err as Error;
      console.error('❌ 보드 생성 실패:', error);
      setError(error.message || '보드 생성에 실패했습니다.');
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
        className={`relative w-full max-w-2xl ${theme.colors.card} p-6 ${theme.effects.borderRadius} shadow-xl max-h-[90vh] overflow-y-auto`}
        onClick={(e) => e.stopPropagation()}
      >
        {/* Header */}
        <div className="flex items-center justify-between mb-4 sticky top-0 bg-white pb-2 border-b">
          <h2 className="text-xl font-bold text-gray-800">새 보드 만들기</h2>
          <button
            onClick={onClose}
            className="p-2 rounded-full hover:bg-gray-100 text-gray-500 hover:text-gray-700 transition"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Error Message */}
        {error && (
          <div className="mb-4 p-3 bg-red-50 border border-red-300 rounded-lg text-red-700 text-sm">
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
          <form onSubmit={handleSubmit} className="space-y-4">
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

            {/* Stage Selection */}
            <div>
              <label className="block text-sm font-semibold text-gray-700 mb-2">
                <CheckSquare className="w-4 h-4 inline mr-1" />
                진행 단계 <span className="text-red-500">*</span>
              </label>
              <select
                value={selectedStageId}
                onChange={(e) => {
                  if (e.target.value === '__create_new__') {
                    setShowCreateStage(true);
                  } else {
                    setSelectedStageId(e.target.value);
                  }
                }}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm bg-white"
                disabled={isLoading}
              >
                {stages.map((stage) => (
                  <option key={stage.id} value={stage.id}>
                    {stage.name}
                  </option>
                ))}
                <option value="__create_new__">+ 새 진행 단계 추가</option>
              </select>
            </div>

            {/* Role Selection */}
            <div className="relative">
              <label className="block text-sm font-semibold text-gray-700 mb-2">
                <Tag className="w-4 h-4 inline mr-1" />
                역할 선택 <span className="text-red-500">*</span>
                <span className="text-xs text-gray-500 ml-2">(최소 1개)</span>
              </label>
              {/* 선택된 역할 태그들 */}
              <div className="flex flex-wrap gap-2 mb-2">
                {selectedRoleIds.map((roleId) => {
                  const role = roles.find((r) => r.id === roleId);
                  if (!role) return null;
                  return (
                    <div
                      key={roleId}
                      className="inline-flex items-center gap-2 px-3 py-1.5 bg-green-100 text-green-800 rounded-lg text-sm font-medium"
                    >
                      <span
                        className="w-3 h-3 rounded-full"
                        style={{ backgroundColor: role.color || '#6B7280' }}
                      />
                      {role.name}
                      <button
                        type="button"
                        onClick={() => toggleRole(roleId)}
                        className="text-green-600 hover:text-green-800"
                        disabled={isLoading}
                      >
                        <X className="w-3.5 h-3.5" />
                      </button>
                    </div>
                  );
                })}
              </div>
              {/* 역할 추가 버튼 */}
              <div className="relative role-dropdown-container">
                <button
                  type="button"
                  onClick={() => setShowRoleDropdown(!showRoleDropdown)}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg bg-white hover:bg-gray-50 transition text-sm text-left flex items-center justify-between"
                  disabled={isLoading}
                >
                  <span className="text-gray-600">+ 역할 추가</span>
                  <Tag className="w-4 h-4 text-gray-400" />
                </button>
                {/* 드롭다운 메뉴 */}
                {showRoleDropdown && (
                  <div className="absolute z-10 w-full mt-1 bg-white border border-gray-300 rounded-lg shadow-lg max-h-48 overflow-y-auto">
                    {roles
                      .filter((role) => !selectedRoleIds.includes(role.id))
                      .map((role) => (
                        <button
                          key={role.id}
                          type="button"
                          onClick={() => {
                            toggleRole(role.id);
                            setShowRoleDropdown(false);
                          }}
                          className="w-full px-3 py-2 text-left hover:bg-gray-100 transition text-sm flex items-center gap-2"
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

            {/* Importance Selection */}
            <div>
              <label className="block text-sm font-semibold text-gray-700 mb-2">
                <AlertCircle className="w-4 h-4 inline mr-1" />
                중요도 (선택)
              </label>
              <select
                value={selectedImportanceId}
                onChange={(e) => {
                  if (e.target.value === '__create_new__') {
                    setShowCreateImportance(true);
                  } else {
                    setSelectedImportanceId(e.target.value);
                  }
                }}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm bg-white"
                disabled={isLoading}
              >
                <option value="">없음</option>
                {importances.map((importance) => (
                  <option key={importance.id} value={importance.id}>
                    {importance.name}
                    {'level' in importance ? ` (Lv.${importance.level})` : ''}
                  </option>
                ))}
                <option value="__create_new__">+ 새 중요도 추가</option>
              </select>
            </div>

            {/* Assignee and Due Date */}
            <div className="grid grid-cols-2 gap-4">
              {/* Assignee */}
              <div>
                <label className="block text-sm font-semibold text-gray-700 mb-2">
                  <User className="w-4 h-4 inline mr-1" />
                  담당자 (선택)
                </label>
                <input
                  type="text"
                  value={assigneeId}
                  onChange={(e) => setAssigneeId(e.target.value)}
                  placeholder="담당자 ID"
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm"
                  disabled={isLoading}
                />
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
                {isLoading ? '생성 중...' : '보드 만들기'}
              </button>
            </div>
          </form>
        )}
      </div>

      {/* Creation Modals */}
      {showCreateStage && renderCreationModal('stage', '진행 단계')}
      {showCreateRole && renderCreationModal('role', '역할')}
      {showCreateImportance && renderCreationModal('importance', '중요도')}
    </div>
  );
};
