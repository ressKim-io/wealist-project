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

  // Helper: Inline Creation Form
  const renderInlineCreationForm = (
    type: 'stage' | 'role' | 'importance',
    title: string,
  ) => (
    <div className="mt-3 p-4 border-2 border-dashed border-blue-300 rounded-lg bg-blue-50">
      <h4 className="text-sm font-semibold text-gray-700 mb-3">새 {title} 추가</h4>
      <div className="space-y-3">
        <input
          type="text"
          value={newFieldName}
          onChange={(e) => setNewFieldName(e.target.value)}
          placeholder={`${title} 이름`}
          className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm"
          disabled={isLoading}
          autoFocus
        />
        <div>
          <label className="text-xs text-gray-600 block mb-1">색상 선택</label>
          {renderColorPicker(newFieldColor, setNewFieldColor)}
        </div>
        {type === 'importance' && (
          <div>
            <label className="text-xs text-gray-600 block mb-1">중요도 레벨 (1-5)</label>
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
        <div className="flex gap-2">
          <button
            type="button"
            onClick={() => handleCreateCustomField(type)}
            className="flex-1 px-3 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 transition text-sm font-medium"
            disabled={isLoading || !newFieldName.trim()}
          >
            추가
          </button>
          <button
            type="button"
            onClick={cancelInlineCreation}
            className="flex-1 px-3 py-2 bg-gray-300 text-gray-700 rounded-lg hover:bg-gray-400 transition text-sm font-medium"
            disabled={isLoading}
          >
            취소
          </button>
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
              <div className="grid grid-cols-2 gap-2">
                {stages.map((stage) => (
                  <button
                    key={stage.id}
                    type="button"
                    onClick={() => setSelectedStageId(stage.id)}
                    className={`px-3 py-2 rounded-lg border-2 transition text-sm font-medium ${
                      selectedStageId === stage.id
                        ? 'border-blue-500 bg-blue-50 text-blue-700'
                        : 'border-gray-200 hover:border-gray-300 text-gray-700'
                    }`}
                    disabled={isLoading}
                  >
                    <span
                      className="inline-block w-3 h-3 rounded-full mr-2"
                      style={{ backgroundColor: stage.color || '#6B7280' }}
                    />
                    {stage.name}
                  </button>
                ))}
              </div>
              {!showCreateStage && !showCreateRole && !showCreateImportance && (
                <button
                  type="button"
                  onClick={() => setShowCreateStage(true)}
                  className="mt-2 w-full py-2 border-2 border-dashed border-gray-300 rounded-lg text-gray-600 hover:border-blue-400 hover:text-blue-600 transition text-sm flex items-center justify-center gap-1"
                  disabled={isLoading}
                >
                  <Plus className="w-4 h-4" />
                  새 진행 단계 추가
                </button>
              )}
              {showCreateStage && renderInlineCreationForm('stage', '진행 단계')}
            </div>

            {/* Role Selection */}
            <div>
              <label className="block text-sm font-semibold text-gray-700 mb-2">
                <Tag className="w-4 h-4 inline mr-1" />
                역할 선택 <span className="text-red-500">*</span>
                <span className="text-xs text-gray-500 ml-2">(최소 1개)</span>
              </label>
              <div className="grid grid-cols-2 gap-2">
                {roles.map((role) => (
                  <button
                    key={role.id}
                    type="button"
                    onClick={() => toggleRole(role.id)}
                    className={`px-3 py-2 rounded-lg border-2 transition text-sm font-medium ${
                      selectedRoleIds.includes(role.id)
                        ? 'border-green-500 bg-green-50 text-green-700'
                        : 'border-gray-200 hover:border-gray-300 text-gray-700'
                    }`}
                    disabled={isLoading}
                  >
                    <span
                      className="inline-block w-3 h-3 rounded-full mr-2"
                      style={{ backgroundColor: role.color || '#6B7280' }}
                    />
                    {role.name}
                    {selectedRoleIds.includes(role.id) && (
                      <span className="ml-2 text-green-600">✓</span>
                    )}
                  </button>
                ))}
              </div>
              {!showCreateStage && !showCreateRole && !showCreateImportance && (
                <button
                  type="button"
                  onClick={() => setShowCreateRole(true)}
                  className="mt-2 w-full py-2 border-2 border-dashed border-gray-300 rounded-lg text-gray-600 hover:border-blue-400 hover:text-blue-600 transition text-sm flex items-center justify-center gap-1"
                  disabled={isLoading}
                >
                  <Plus className="w-4 h-4" />
                  새 역할 추가
                </button>
              )}
              {showCreateRole && renderInlineCreationForm('role', '역할')}
            </div>

            {/* Importance Selection */}
            <div>
              <label className="block text-sm font-semibold text-gray-700 mb-2">
                <AlertCircle className="w-4 h-4 inline mr-1" />
                중요도 (선택)
              </label>
              <div className="grid grid-cols-2 gap-2">
                {/* "없음" 버튼 */}
                <button
                  type="button"
                  onClick={() => setSelectedImportanceId('')}
                  className={`px-3 py-2 rounded-lg border-2 transition text-sm font-medium ${
                    selectedImportanceId === ''
                      ? 'border-purple-500 bg-purple-50 text-purple-700'
                      : 'border-gray-200 hover:border-gray-300 text-gray-700'
                  }`}
                  disabled={isLoading}
                >
                  <span className="inline-block w-3 h-3 rounded-full mr-2 bg-gray-300" />
                  없음
                </button>
                {importances.map((importance) => (
                  <button
                    key={importance.id}
                    type="button"
                    onClick={() => setSelectedImportanceId(importance.id)}
                    className={`px-3 py-2 rounded-lg border-2 transition text-sm font-medium ${
                      selectedImportanceId === importance.id
                        ? 'border-purple-500 bg-purple-50 text-purple-700'
                        : 'border-gray-200 hover:border-gray-300 text-gray-700'
                    }`}
                    disabled={isLoading}
                  >
                    <span
                      className="inline-block w-3 h-3 rounded-full mr-2"
                      style={{ backgroundColor: importance.color || '#6B7280' }}
                    />
                    {importance.name}
                    {'level' in importance && (
                      <span className="ml-1 text-xs">Lv.{importance.level}</span>
                    )}
                  </button>
                ))}
              </div>
              {!showCreateStage && !showCreateRole && !showCreateImportance && (
                <button
                  type="button"
                  onClick={() => setShowCreateImportance(true)}
                  className="mt-2 w-full py-2 border-2 border-dashed border-gray-300 rounded-lg text-gray-600 hover:border-blue-400 hover:text-blue-600 transition text-sm flex items-center justify-center gap-1"
                  disabled={isLoading}
                >
                  <Plus className="w-4 h-4" />
                  새 중요도 추가
                </button>
              )}
              {showCreateImportance && renderInlineCreationForm('importance', '중요도')}
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
    </div>
  );
};
