import React, { useState, useEffect } from 'react';
import { X, Plus, Edit2, Trash2, Check } from 'lucide-react';
import { useTheme } from '../../contexts/ThemeContext';
import { CUSTOM_FIELD_COLORS } from '../../constants/colors';
import {
  CustomStageResponse,
  CustomRoleResponse,
  CustomImportanceResponse,
  getProjectStages,
  getProjectRoles,
  getProjectImportances,
  createStage,
  updateStage,
  deleteStage,
  createRole,
  updateRole,
  deleteRole,
  createImportance,
  updateImportance,
  deleteImportance,
} from '../../api/board/boardService';

interface CustomFieldManageModalProps {
  projectId: string;
  onClose: () => void;
  onFieldsUpdated: () => void;
}

type TabType = 'stages' | 'roles' | 'importances';

export const CustomFieldManageModal: React.FC<CustomFieldManageModalProps> = ({
  projectId,
  onClose,
  onFieldsUpdated,
}) => {
  const { theme } = useTheme();
  const accessToken = localStorage.getItem('access_token') || '';

  const [activeTab, setActiveTab] = useState<TabType>('stages');
  const [stages, setStages] = useState<CustomStageResponse[]>([]);
  const [roles, setRoles] = useState<CustomRoleResponse[]>([]);
  const [importances, setImportances] = useState<CustomImportanceResponse[]>([]);
  const [loading, setLoading] = useState(false);

  // Edit state
  const [editingId, setEditingId] = useState<string | null>(null);
  const [editName, setEditName] = useState('');
  const [editColor, setEditColor] = useState('');
  const [editLevel, setEditLevel] = useState(1);

  // Create state
  const [showCreateForm, setShowCreateForm] = useState(false);
  const [createName, setCreateName] = useState('');
  const [createColor, setCreateColor] = useState(CUSTOM_FIELD_COLORS[0].hex);
  const [createLevel, setCreateLevel] = useState(1);

  useEffect(() => {
    loadData();
  }, [projectId, activeTab]);

  const loadData = async () => {
    setLoading(true);
    try {
      if (activeTab === 'stages') {
        const data = await getProjectStages(projectId, accessToken);
        setStages(data.sort((a, b) => a.displayOrder - b.displayOrder));
      } else if (activeTab === 'roles') {
        const data = await getProjectRoles(projectId, accessToken);
        setRoles(data.sort((a, b) => a.displayOrder - b.displayOrder));
      } else {
        const data = await getProjectImportances(projectId, accessToken);
        setImportances(data.sort((a, b) => a.displayOrder - b.displayOrder));
      }
    } catch (error) {
      console.error('Failed to load custom fields:', error);
      alert('데이터를 불러오는데 실패했습니다.');
    } finally {
      setLoading(false);
    }
  };

  const handleCreate = async () => {
    if (!createName.trim()) {
      alert('이름을 입력해주세요.');
      return;
    }

    try {
      if (activeTab === 'stages') {
        await createStage({ projectId, name: createName, color: createColor }, accessToken);
      } else if (activeTab === 'roles') {
        await createRole({ projectId, name: createName, color: createColor }, accessToken);
      } else {
        await createImportance(
          { projectId, name: createName, color: createColor, level: createLevel },
          accessToken,
        );
      }

      setCreateName('');
      setCreateColor(CUSTOM_FIELD_COLORS[0].hex);
      setCreateLevel(1);
      setShowCreateForm(false);
      await loadData();
      onFieldsUpdated();
    } catch (error) {
      console.error('Failed to create custom field:', error);
      alert('생성에 실패했습니다.');
    }
  };

  const handleStartEdit = (
    id: string,
    name: string,
    color: string | undefined,
    level?: number,
  ) => {
    setEditingId(id);
    setEditName(name);
    setEditColor(color || CUSTOM_FIELD_COLORS[0].hex);
    setEditLevel(level || 1);
  };

  const handleSaveEdit = async () => {
    if (!editName.trim() || !editingId) return;

    try {
      if (activeTab === 'stages') {
        await updateStage(editingId, { name: editName, color: editColor }, accessToken);
      } else if (activeTab === 'roles') {
        await updateRole(editingId, { name: editName, color: editColor }, accessToken);
      } else {
        await updateImportance(
          editingId,
          { name: editName, color: editColor, level: editLevel },
          accessToken,
        );
      }

      setEditingId(null);
      await loadData();
      onFieldsUpdated();
    } catch (error) {
      console.error('Failed to update custom field:', error);
      alert('수정에 실패했습니다.');
    }
  };

  const handleDelete = async (id: string, isSystemDefault: boolean) => {
    if (isSystemDefault) {
      alert('시스템 기본값은 삭제할 수 없습니다.');
      return;
    }

    if (!confirm('정말 삭제하시겠습니까?')) return;

    try {
      if (activeTab === 'stages') {
        await deleteStage(id, accessToken);
      } else if (activeTab === 'roles') {
        await deleteRole(id, accessToken);
      } else {
        await deleteImportance(id, accessToken);
      }

      await loadData();
      onFieldsUpdated();
    } catch (error) {
      console.error('Failed to delete custom field:', error);
      alert('삭제에 실패했습니다.');
    }
  };

  const renderColorPicker = (selectedColor: string, onColorChange: (color: string) => void) => (
    <div className="grid grid-cols-6 gap-2 mt-2">
      {CUSTOM_FIELD_COLORS.map((color) => (
        <button
          key={color.hex}
          type="button"
          className={`w-8 h-8 rounded-md border-2 ${
            selectedColor === color.hex ? 'border-gray-800 ring-2 ring-blue-500' : 'border-gray-300'
          } transition-all`}
          style={{ backgroundColor: color.hex }}
          onClick={() => onColorChange(color.hex)}
          title={color.name}
        />
      ))}
    </div>
  );

  const renderItemsList = () => {
    const items =
      activeTab === 'stages' ? stages : activeTab === 'roles' ? roles : importances;

    return (
      <div className="space-y-2">
        {items.map((item) => {
          const isEditing = editingId === item.id;
          const itemLevel: number | undefined = 'level' in item ? (item.level as number) : undefined;

          return (
            <div
              key={item.id}
              className={`p-3 ${theme.colors.card} border ${theme.colors.border} rounded-md`}
            >
              {isEditing ? (
                <div className="space-y-3">
                  <input
                    type="text"
                    value={editName}
                    onChange={(e) => setEditName(e.target.value)}
                    className={`w-full px-3 py-2 border ${theme.colors.border} rounded-md ${theme.colors.card}`}
                    placeholder="이름"
                  />
                  {renderColorPicker(editColor, setEditColor)}
                  {activeTab === 'importances' && (
                    <div>
                      <label className="text-sm text-gray-600">Level (1-5)</label>
                      <input
                        type="number"
                        min="1"
                        max="5"
                        value={editLevel}
                        onChange={(e) => setEditLevel(parseInt(e.target.value))}
                        className={`w-full px-3 py-2 border ${theme.colors.border} rounded-md ${theme.colors.card}`}
                      />
                    </div>
                  )}
                  <div className="flex gap-2">
                    <button
                      onClick={handleSaveEdit}
                      className="px-3 py-1 bg-blue-500 text-white rounded-md hover:bg-blue-600 flex items-center gap-1"
                    >
                      <Check className="w-4 h-4" />
                      저장
                    </button>
                    <button
                      onClick={() => setEditingId(null)}
                      className="px-3 py-1 bg-gray-300 text-gray-700 rounded-md hover:bg-gray-400"
                    >
                      취소
                    </button>
                  </div>
                </div>
              ) : (
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <div
                      className="w-6 h-6 rounded-md border"
                      style={{ backgroundColor: item.color || CUSTOM_FIELD_COLORS[0].hex }}
                    />
                    <div>
                      <span className={`font-medium ${theme.colors.text}`}>{item.name}</span>
                      {itemLevel && (
                        <span className="ml-2 text-sm text-gray-500">Level {itemLevel}</span>
                      )}
                      {item.isSystemDefault && (
                        <span className="ml-2 text-xs bg-gray-200 text-gray-600 px-2 py-0.5 rounded">
                          기본값
                        </span>
                      )}
                    </div>
                  </div>
                  <div className="flex gap-2">
                    <button
                      onClick={() =>
                        handleStartEdit(item.id, item.name, item.color, itemLevel)
                      }
                      className="p-1 hover:bg-gray-200 rounded-md"
                    >
                      <Edit2 className="w-4 h-4 text-gray-600" />
                    </button>
                    <button
                      onClick={() => handleDelete(item.id, item.isSystemDefault)}
                      className={`p-1 rounded-md ${
                        item.isSystemDefault
                          ? 'opacity-30 cursor-not-allowed'
                          : 'hover:bg-red-100'
                      }`}
                      disabled={item.isSystemDefault}
                    >
                      <Trash2 className="w-4 h-4 text-red-600" />
                    </button>
                  </div>
                </div>
              )}
            </div>
          );
        })}
      </div>
    );
  };

  const renderCreateForm = () => (
    <div className={`mt-4 p-4 border ${theme.colors.border} rounded-md ${theme.colors.card}`}>
      <h4 className="font-semibold mb-3">새로운 항목 추가</h4>
      <div className="space-y-3">
        <input
          type="text"
          value={createName}
          onChange={(e) => setCreateName(e.target.value)}
          placeholder="이름 입력"
          className={`w-full px-3 py-2 border ${theme.colors.border} rounded-md ${theme.colors.card}`}
        />
        {renderColorPicker(createColor, setCreateColor)}
        {activeTab === 'importances' && (
          <div>
            <label className="text-sm text-gray-600">Level (1-5)</label>
            <input
              type="number"
              min="1"
              max="5"
              value={createLevel}
              onChange={(e) => setCreateLevel(parseInt(e.target.value))}
              className={`w-full px-3 py-2 border ${theme.colors.border} rounded-md ${theme.colors.card}`}
            />
          </div>
        )}
        <div className="flex gap-2">
          <button
            onClick={handleCreate}
            className="px-4 py-2 bg-blue-500 text-white rounded-md hover:bg-blue-600 flex items-center gap-1"
          >
            <Plus className="w-4 h-4" />
            추가
          </button>
          <button
            onClick={() => setShowCreateForm(false)}
            className="px-4 py-2 bg-gray-300 text-gray-700 rounded-md hover:bg-gray-400"
          >
            취소
          </button>
        </div>
      </div>
    </div>
  );

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
      <div
        className={`${theme.colors.card} rounded-lg shadow-xl w-full max-w-2xl max-h-[90vh] overflow-hidden flex flex-col`}
      >
        {/* Header */}
        <div className={`p-6 border-b ${theme.colors.border} flex justify-between items-center`}>
          <h2 className={`text-2xl font-bold ${theme.colors.text}`}>커스텀 필드 관리</h2>
          <button
            onClick={onClose}
            className="text-gray-500 hover:text-gray-700 transition-colors"
          >
            <X className="w-6 h-6" />
          </button>
        </div>

        {/* Tabs */}
        <div className={`flex border-b ${theme.colors.border} px-6`}>
          <button
            onClick={() => setActiveTab('stages')}
            className={`px-4 py-3 border-b-2 transition-colors ${
              activeTab === 'stages'
                ? 'border-blue-500 text-blue-600 font-semibold'
                : 'border-transparent text-gray-600 hover:text-gray-800'
            }`}
          >
            진행 단계 (Stage)
          </button>
          <button
            onClick={() => setActiveTab('roles')}
            className={`px-4 py-3 border-b-2 transition-colors ${
              activeTab === 'roles'
                ? 'border-blue-500 text-blue-600 font-semibold'
                : 'border-transparent text-gray-600 hover:text-gray-800'
            }`}
          >
            역할 (Role)
          </button>
          <button
            onClick={() => setActiveTab('importances')}
            className={`px-4 py-3 border-b-2 transition-colors ${
              activeTab === 'importances'
                ? 'border-blue-500 text-blue-600 font-semibold'
                : 'border-transparent text-gray-600 hover:text-gray-800'
            }`}
          >
            중요도 (Importance)
          </button>
        </div>

        {/* Content */}
        <div className="flex-1 overflow-y-auto p-6">
          {loading ? (
            <div className="text-center py-8">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500 mx-auto"></div>
            </div>
          ) : (
            <>
              {renderItemsList()}

              {!showCreateForm && (
                <button
                  onClick={() => setShowCreateForm(true)}
                  className="mt-4 w-full py-3 border-2 border-dashed border-gray-300 rounded-md text-gray-600 hover:border-blue-500 hover:text-blue-500 transition-colors flex items-center justify-center gap-2"
                >
                  <Plus className="w-5 h-5" />
                  새로운 항목 추가
                </button>
              )}

              {showCreateForm && renderCreateForm()}
            </>
          )}
        </div>
      </div>
    </div>
  );
};
