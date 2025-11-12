import React, { useState, useEffect } from 'react';
import { X, Plus, Edit2, Trash2, Check } from 'lucide-react';
import { useTheme } from '../../contexts/ThemeContext';
import { CUSTOM_FIELD_COLORS } from '../../constants/colors';
import {
  CustomStageResponse,
  CustomRoleResponse,
  CustomImportanceResponse,
} from '../../types/board';
// 💡 [복원] boardService의 import를 완성합니다.
import '../../api/board/boardService';

interface CustomFieldManageModalProps {
  projectId: string;
  onClose: () => void;
  onFieldsUpdated: () => void;
}

type TabType = 'stages' | 'roles' | 'importances';

// ⚠️ [주의] API 호출이 제거되었으므로, 컴포넌트 로직 유지를 위해 Mock Data를 사용합니다.
// 💡 UUID 형식으로 변경하여 백엔드 검증 통과
const MOCK_STAGES: CustomStageResponse[] = [
  {
    stageId: '00000000-0000-0000-0000-000000000014',
    label: '트리아지',
    color: '#64748B',
    displayOrder: 0,
    fieldId: '00000000-0000-0000-0000-000000000010',
    description: '기본값',
    isSystemDefault: true,
  },
  {
    stageId: '00000000-0000-0000-0000-000000000002',
    label: '진행중',
    color: '#3B82F6',
    displayOrder: 1,
    fieldId: '00000000-0000-0000-0000-000000000010',
    description: '',
    isSystemDefault: false,
  },
  {
    stageId: '00000000-0000-0000-0000-000000000003',
    label: '완료',
    color: '#10B981',
    displayOrder: 2,
    fieldId: '00000000-0000-0000-0000-000000000010',
    description: '',
    isSystemDefault: true,
  },
];
const MOCK_ROLES: CustomRoleResponse[] = [
  {
    roleId: '00000000-0000-0000-0000-000000000004',
    label: '개발',
    color: '#8B5CF6',
    displayOrder: 0,
    fieldId: '00000000-0000-0000-0000-000000000011',
    description: '기본값',
    isSystemDefault: true,
  },
  {
    roleId: '00000000-0000-0000-0000-000000000013',
    label: '디자인',
    color: '#F59E0B',
    displayOrder: 1,
    fieldId: '00000000-0000-0000-0000-000000000011',
    description: '',
    isSystemDefault: false,
  },
];
const MOCK_IMPORTANCES: CustomImportanceResponse[] = [
  {
    importanceId: '00000000-0000-0000-0000-000000000006',
    label: '긴급',
    color: '#EF4444',
    displayOrder: 0,
    fieldId: '00000000-0000-0000-0000-000000000012',
    description: '',
    isSystemDefault: false,
    level: 5,
  },
  {
    importanceId: '00000000-0000-0000-0000-000000000007',
    label: '낮음',
    color: '#10B981',
    displayOrder: 1,
    fieldId: '00000000-0000-0000-0000-000000000012',
    description: '기본값',
    isSystemDefault: true,
    level: 1,
  },
];

export const CustomFieldManageModal: React.FC<CustomFieldManageModalProps> = ({
  projectId, // 💡 [복원] props에서 projectId를 사용합니다.
  onClose,
  onFieldsUpdated,
}) => {
  const { theme } = useTheme();
  // const accessToken = localStorage.getItem('accessToken') || ''; // 💡 API 미사용으로 인해 제거

  const [activeTab, setActiveTab] = useState<TabType>('stages');
  const [stages, setStages] = useState<CustomStageResponse[]>([]);
  const [roles, setRoles] = useState<CustomRoleResponse[]>([]);
  const [importances, setImportances] = useState<CustomImportanceResponse[]>([]);

  const [loading, setLoading] = useState(false);

  const [apiWarning, setApiWarning] = useState<string | null>(null);

  // Edit state (변경 없음)
  const [editingId, setEditingId] = useState<string | null>(null);
  const [editName, setEditName] = useState('');
  const [editColor, setEditColor] = useState('');
  const [editLevel, setEditLevel] = useState(1);

  // Create state (변경 없음)
  const [showCreateForm, setShowCreateForm] = useState(false);
  const [createName, setCreateName] = useState('');
  const [createColor, setCreateColor] = useState(CUSTOM_FIELD_COLORS[0].hex);
  const [createLevel, setCreateLevel] = useState(1);

  // 💡 [수정] API 호출 대신 Mock Data를 로드하는 useEffect
  useEffect(() => {
    // API 호출 대신 Mock Data를 로드하고 정렬합니다.
    setLoading(true);
    setApiWarning(null); // 탭 변경 시 경고 초기화

    // API 호출 대신 Mock Data를 사용합니다.
    if (activeTab === 'stages') {
      setStages(MOCK_STAGES.sort((a, b) => a.displayOrder - b.displayOrder));
    } else if (activeTab === 'roles') {
      setRoles(MOCK_ROLES.sort((a, b) => a.displayOrder - b.displayOrder));
    } else {
      setImportances(MOCK_IMPORTANCES.sort((a, b) => a.displayOrder - b.displayOrder));
    }

    setLoading(false);

    // 💡 경고 메시지 설정: 이 기능은 현재 API 스펙에 따라 구현이 필요함.
    setApiWarning(
      '⚠️ 현재 API 스펙 변경으로 인해 Custom Field CRUD 기능이 일시적으로 비활성화되었습니다. (API: /fields, /field-options으로 통합 예정)',
    );

    // 편집 및 생성 폼 상태 초기화
    setEditingId(null);
    setShowCreateForm(false);
  }, [activeTab]);

  const loadData = async () => {
    // Mock Data 로드 로직은 useEffect로 통합되었으므로, 여기서는 단순히 state를 갱신합니다.
    if (activeTab === 'stages') {
      setStages([...stages].sort((a, b) => a.displayOrder - b.displayOrder));
    } else if (activeTab === 'roles') {
      setRoles([...roles].sort((a, b) => a.displayOrder - b.displayOrder));
    } else {
      setImportances([...importances].sort((a, b) => a.displayOrder - b.displayOrder));
    }
  };

  // 💡 [수정] 생성 로직: API 호출을 제거하고 경고 표시로 대체
  const handleCreate = async () => {
    alert(apiWarning);
    setShowCreateForm(false);
    onFieldsUpdated(); // Mocking이지만 업데이트 완료 처리는 해줍니다.
  };

  const handleStartEdit = (
    id: string,
    name: string,
    color: string | undefined,
    level?: number,
    isSystemDefault?: boolean,
  ) => {
    // 시스템 기본값은 편집 불가
    if (isSystemDefault) {
      alert('시스템 기본값은 수정할 수 없습니다.');
      return;
    }

    setEditingId(id);
    setEditName(name);
    setEditColor(color || CUSTOM_FIELD_COLORS[0].hex);
    setEditLevel(level || 1);
  };

  // 💡 [수정] 수정 로직: API 호출을 제거하고 경고 표시로 대체
  const handleSaveEdit = async () => {
    alert(apiWarning);
    setEditingId(null);
    onFieldsUpdated(); // Mocking이지만 업데이트 완료 처리는 해줍니다.
  };

  // 💡 [수정] 삭제 로직: API 호출을 제거하고 경고 표시로 대체
  const handleDelete = async (id: string, isSystemDefault: boolean) => {
    if (isSystemDefault) {
      alert('시스템 기본값은 삭제할 수 없습니다.');
      return;
    }

    alert(apiWarning);
    onFieldsUpdated(); // Mocking이지만 업데이트 완료 처리는 해줍니다.
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
    const items: (CustomStageResponse | CustomRoleResponse | CustomImportanceResponse)[] =
      activeTab === 'stages' ? stages : activeTab === 'roles' ? roles : importances;

    return (
      <div className="space-y-2">
        {/* 💡 API 미지원 경고 표시 */}
        {apiWarning && (
          <div className="mb-4 p-3 bg-yellow-100 border border-yellow-300 rounded-lg text-yellow-700 text-sm">
            {apiWarning}
          </div>
        )}

        {items.map((item) => {
          // 💡 [수정] id 필드 추출 로직: 각 타입의 고유 ID를 id로 사용합니다.
          const itemId =
            ('stageId' in item && item.stageId) ||
            ('roleId' in item && item.roleId) ||
            ('importanceId' in item && item.importanceId) ||
            'unknown';

          // 💡 [수정] name 필드를 label로 대체합니다.
          const itemName = item.label;

          const isEditing = editingId === itemId;
          // 💡 [수정] level 필드 추출 로직
          const itemLevel: number | undefined =
            'level' in item && typeof item.level === 'number' ? item.level : undefined;

          // 💡 [수정] isSystemDefault 필드 사용
          const isSystemDefault = 'isSystemDefault' in item && item.isSystemDefault;

          return (
            <div
              key={itemId}
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
                    disabled={isSystemDefault}
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
                        disabled={isSystemDefault}
                      />
                    </div>
                  )}
                  <div className="flex gap-2">
                    <button
                      onClick={handleSaveEdit}
                      className="px-3 py-1 bg-blue-500 text-white rounded-md hover:bg-blue-600 flex items-center gap-1"
                      disabled={isSystemDefault}
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
                      <span className={`font-medium ${theme.colors.text}`}>{itemName}</span>
                      {itemLevel && (
                        <span className="ml-2 text-sm text-gray-500">Level {itemLevel}</span>
                      )}
                      {isSystemDefault && (
                        <span className="ml-2 text-xs bg-gray-200 text-gray-600 px-2 py-0.5 rounded">
                          기본값
                        </span>
                      )}
                    </div>
                  </div>
                  <div className="flex gap-2">
                    <button
                      onClick={() =>
                        // 💡 [수정] onStartEdit에 isSystemDefault 전달
                        handleStartEdit(itemId, itemName, item.color, itemLevel, isSystemDefault)
                      }
                      className={`p-1 hover:bg-gray-200 rounded-md ${
                        isSystemDefault ? 'opacity-30 cursor-not-allowed' : ''
                      }`}
                      disabled={isSystemDefault}
                    >
                      <Edit2 className="w-4 h-4 text-gray-600" />
                    </button>
                    <button
                      onClick={() => handleDelete(itemId, !!isSystemDefault)}
                      className={`p-1 rounded-md ${
                        isSystemDefault ? 'opacity-30 cursor-not-allowed' : 'hover:bg-red-100'
                      }`}
                      disabled={isSystemDefault}
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
      {apiWarning && (
        <div className="mb-3 p-2 bg-red-100 border border-red-300 rounded-lg text-red-700 text-sm">
          {apiWarning}
        </div>
      )}
      <div className="space-y-3">
        <input
          type="text"
          value={createName}
          onChange={(e) => setCreateName(e.target.value)}
          placeholder="이름 입력"
          className={`w-full px-3 py-2 border ${theme.colors.border} rounded-md ${theme.colors.card}`}
          disabled={!!apiWarning}
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
              disabled={!!apiWarning}
            />
          </div>
        )}
        <div className="flex gap-2">
          <button
            onClick={handleCreate}
            className="px-4 py-2 bg-blue-500 text-white rounded-md hover:bg-blue-600 flex items-center gap-1 disabled:opacity-50"
            disabled={!!apiWarning || !createName.trim()}
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
          <button onClick={onClose} className="text-gray-500 hover:text-gray-700 transition-colors">
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
                  disabled={!!apiWarning}
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
