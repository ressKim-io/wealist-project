// src/pages/Dashboard.tsx (MainDashboard.tsx)

import { useLocation, useParams } from 'react-router-dom';
import React, { useEffect, useState, useRef, useCallback } from 'react';
import { Briefcase } from 'lucide-react';
import { useTheme } from '../contexts/ThemeContext';

// 💡 [분리된 컴포넌트]
import MainLayout from '../components/layout/MainLayout';
import { ProjectHeader } from '../components/layout/ProjectHeader';
import { ProjectContent } from '../components/layout/ProjectContent';

import UserProfileModal from '../components/modals/user/UserProfileModal';
import { LoadingSpinner } from '../components/common/LoadingSpinner';

import { getProjects, getProjectInitSettings } from '../api/board/boardService';
import { getWorkspaceMembers } from '../api/user/userService';

import {
  ProjectResponse,
  FieldWithOptionsResponse,
  FieldOption,
  FieldTypeInfo,
} from '../types/board';
import { WorkspaceMemberResponse } from '../types/user';
import { CustomFieldManageModal } from '../components/modals/board/customFields/CustomFieldManageModal';
import { BoardManageModal } from '../components/modals/board/BoardManageModal';
import { IROLES } from '../types/common';
import { ProjectManageModal } from '../components/modals/board/ProjectManageModal';

interface MainDashboardProps {
  onLogout: () => void;
}

// 💡 [추가] UI/모달 상태를 통합하는 인터페이스
interface UIState {
  showProjectSelector?: boolean;
  showUserProfile?: boolean;
  showCreateProject?: boolean; // 프로젝트 생성 (모드: 'create')
  showManageModal?: boolean; // 커스텀 필드 관리
  showProjectSettings?: boolean; // 프로젝트 설정 (모드: 'edit')
  showCreateBoard?: boolean;
  showProjectDetail?: boolean; // 프로젝트 상세 보기 (모드: 'detail')
}

// 💡 [추가] 필드 옵션 룩업 인터페이스
interface FieldOptionsLookup {
  roles?: FieldOption[];
  importances?: FieldOption[];
  stages?: FieldOption[];
}

// =============================================================================
// MainDashboard (컨테이너 역할)
// =============================================================================
const MainDashboard: React.FC<MainDashboardProps> = ({ onLogout }) => {
  const { workspaceId } = useParams<{ workspaceId: string }>();
  const currentWorkspaceId = workspaceId || '';
  const location = useLocation(); // 💡 useLocation 훅 추가
  const { theme } = useTheme(); // 💡 [추가] location.state에서 userRole 추출 (기본값 설정 필요)
  // 타입 가정이 필요하거나, location.state를 명시적으로 타입 캐스팅해야 할 수 있습니다.
  const passedRole = ((location.state as any)?.userRole as IROLES) || 'GUEST'; // GUEST 등 기본값 설정

  // 💡 currentRole을 useRef 대신 state로 관리하거나, Props로 전달해야 함.
  // 여기서는 currentRole.current를 passedRole로 대체할 수 있습니다.
  const currentRole = useRef<IROLES>(passedRole); // 초기 로드 시점의 역할 설정

  // [핵심 상태]
  const [projects, setProjects] = useState<ProjectResponse[]>([]);
  const [selectedProject, setSelectedProject] = useState<ProjectResponse | null>(null);
  const [workspaceMembers, setWorkspaceMembers] = useState<WorkspaceMemberResponse[]>([]);
  const [isLoadingProjects, setIsLoadingProjects] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const [uiState, setUiState] = useState<UIState>({});
  const [editBoardData, setEditBoardData] = useState<any>(null);
  const [editFieldData, setEditFieldData] = useState<any>(null);

  // 💡 [추가] 초기 옵션 데이터를 저장할 상태 (ProjectContent로 전달)
  const [fieldOptionsLookup, setFieldOptionsLookup] = useState<FieldOptionsLookup>({
    roles: [],
    importances: [],
    stages: [],
  });

  const [fieldTypesLookup, setFieldTypesLookup] = useState<FieldTypeInfo[]>([]);

  const toggleUiState = useCallback((key: keyof UIState, show?: boolean) => {
    setUiState((prev) => ({
      ...prev,
      [key]: show !== undefined ? show : !prev?.[key],
    }));
  }, []);

  // 💡 [수정] Helper: FieldWithOptionsResponse -> FieldOption 변환
  const mapFieldOptions = (fields: FieldWithOptionsResponse[]): FieldOptionsLookup => {
    const roles: FieldOption[] = [];
    const importances: FieldOption[] = [];
    const stages: FieldOption[] = [];

    fields?.forEach((field) => {
      // fieldType 확인
      if (
        field.fieldType === 'select' ||
        field.fieldType === 'single_select' ||
        field.fieldType === 'multi_select'
      ) {
        field.options.forEach((opt) => {
          const mappedOption: FieldOption = {
            optionId: opt.optionId,
            optionValue: opt.optionValue,
            optionLabel: opt.optionLabel,
            color: opt.color,
          };

          // fieldName으로 분류
          const fieldName = field?.fieldName;
          if (fieldName === 'Role' || fieldName === 'role') {
            roles.push(mappedOption);
          } else if (fieldName === 'Importance' || fieldName === 'importance') {
            importances.push(mappedOption);
          } else if (fieldName === 'Stage' || fieldName === 'stage') {
            stages.push(mappedOption);
          }
        });
      }
    });

    return { roles, importances, stages };
  };

  // 1. 프로젝트 목록 조회 함수 (Header Dropdown용)
  const fetchProjects = useCallback(async () => {
    if (!currentWorkspaceId) return;

    setIsLoadingProjects(true);
    setError(null);
    try {
      const fetchedProjects = await getProjects(currentWorkspaceId);
      setProjects(fetchedProjects);
      // 💡 [수정] 현재 선택된 프로젝트가 없거나 (초기 로드),
      //    목록에 프로젝트가 있고, 현재 선택된 프로젝트가 목록에 없는 경우 (생성 후)
      //    가장 첫 번째 프로젝트(최신 프로젝트)를 선택하도록 변경합니다.
      const shouldSelectNewProject =
        !selectedProject ||
        (fetchedProjects.length > 0 &&
          !fetchedProjects.some((p) => p.projectId === selectedProject.projectId));

      if (fetchedProjects.length > 0 && shouldSelectNewProject) {
        // API가 최신순으로 정렬해서 반환한다고 가정하고 첫 번째 요소를 선택합니다.
        setSelectedProject(fetchedProjects[0]);
      }
      // 💡 [참고] 만약 선택된 프로젝트가 목록에 여전히 있다면, 변경하지 않습니다 (예: 수정 시).
    } catch (err: any) {
      const error = err as Error;
      setError(`프로젝트 목록 로드 실패: ${error.message}`);
    } finally {
      setIsLoadingProjects(false);
    }
  }, [currentWorkspaceId, selectedProject]);
  // 💡 [추가] 프로젝트 생성 후 호출될 핸들러
  const handleProjectCreated = useCallback((newProject: ProjectResponse) => {
    // 1. 프로젝트 목록에 새로 생성된 프로젝트 추가 (가장 앞에 추가)
    setProjects((prev) => [newProject, ...prev]);

    // 2. 새로 생성된 프로젝트를 즉시 선택 상태로 설정
    setSelectedProject(newProject);

    // 3. (선택 사항) UI 상태 업데이트 (필요하다면 InitSettings도 다시 로드됨)
    //    selectedProject가 변경되면 useEffect가 InitSettings를 트리거합니다.
  }, []);

  // 2. 워크스페이스 회원 조회 함수
  const fetchWorkspaceMembers = useCallback(async () => {
    if (!currentWorkspaceId) return;
    try {
      const members = await getWorkspaceMembers(currentWorkspaceId);
      setWorkspaceMembers(members);
    } catch (err) {
      setWorkspaceMembers([]);
    }
  }, [currentWorkspaceId]);

  // 💡 [핵심 구현] 프로젝트 선택 시 모든 데이터 로드 (InitSettings)
  const fetchProjectContentInitSettings = useCallback(async () => {
    if (!selectedProject) return;

    setError(null);
    try {
      // 💡 [API 호출] GET /api/projects/{projectId}/init-settings
      const initData = await getProjectInitSettings(selectedProject.projectId);
      // 2. 필드 옵션 룩업 테이블 생성
      const fieldLookup = mapFieldOptions(initData.fields);
      setFieldTypesLookup(initData.fieldTypes);
      setFieldOptionsLookup(fieldLookup);

      console.log('✅ Project Init Data (Fields/Boards) Loaded.');
    } catch (err: any) {
      setError(`초기 컨텐츠 로드 실패: ${err.message}`);
    }
  }, [selectedProject]);

  // 3. 초기 로드 및 트리거
  useEffect(() => {
    fetchProjects();
    fetchWorkspaceMembers();
  }, []);

  // 💡 [핵심] selectedProject 변경 시 InitSettings 로드 트리거
  useEffect(() => {
    if (selectedProject) {
      fetchProjectContentInitSettings();
    }
  }, [selectedProject, fetchProjectContentInitSettings]);

  // 💡 ProjectContent에서 보드/필드 업데이트 시 호출될 함수
  const handleBoardContentUpdate = useCallback(() => {
    fetchProjectContentInitSettings();
  }, [fetchProjectContentInitSettings]);

  // 💡 필드가 생성된 후 호출될 핸들러
  const afterFieldCreated = useCallback(() => {
    toggleUiState('showManageModal', false);
    setEditFieldData(null);
    handleBoardContentUpdate(); // 💡 데이터 변경 알림 -> InitSettings 재실행
  }, [handleBoardContentUpdate, toggleUiState]);

  const handleCustomField = useCallback(
    (editFieldData: any) => {
      toggleUiState('showManageModal', true);
      setEditFieldData(editFieldData);
    },
    [toggleUiState],
  );

  return (
    <MainLayout
      onLogout={onLogout}
      workspaceId={currentWorkspaceId}
      onProfileModalOpen={() => toggleUiState('showUserProfile', true)}
    >
      {/* 1. 헤더 영역 */}
      <ProjectHeader
        projects={projects}
        userRole={currentRole?.current} // 💡 [수정] userRole prop 추가
        selectedProject={selectedProject}
        workspaceMembers={workspaceMembers}
        setSelectedProject={setSelectedProject}
        setShowCreateProject={() => toggleUiState('showCreateProject', true)}
        showProjectSelector={uiState?.showProjectSelector || false}
        setShowProjectSelector={(show) => toggleUiState('showProjectSelector', show)}
        setShowProjectDetail={() => toggleUiState('showProjectDetail', true)}
      />

      {/* 2. 메인 콘텐츠 영역 */}
      <div className="flex-grow flex flex-col p-3 sm:p-6 overflow-auto mt-16 ml-20">
        {error && (
          <div className="mb-4 p-4 bg-red-50 border border-red-300 rounded-lg text-red-700">
            {error}
          </div>
        )}

        {isLoadingProjects ? (
          <LoadingSpinner message="프로젝트 목록 로드 중..." />
        ) : selectedProject ? (
          <ProjectContent
            selectedProject={selectedProject}
            workspaceId={currentWorkspaceId}
            onProjectContentUpdate={handleBoardContentUpdate}
            onManageModalOpen={() => toggleUiState('showManageModal', true)}
            onEditBoard={setEditBoardData}
            showCreateBoard={uiState?.showCreateBoard || false}
            setShowCreateBoard={(show) => toggleUiState('showCreateBoard', show)}
            fieldOptionsLookup={fieldOptionsLookup}
          />
        ) : (
          <div className="flex flex-col items-center justify-center h-full text-center p-8">
            <Briefcase className="w-16 h-16 mb-4 text-gray-400" />
            <h2 className={`${theme.font.size.xl} ${theme.colors.text} mb-2`}>
              프로젝트를 선택하세요
            </h2>
            <p className={`${theme.colors.subText}`}>프로젝트 목록을 불러오고 선택하세요.</p>
          </div>
        )}
      </div>

      {/* 3. 모달 영역 */}
      {/* UserProfile Modal */}
      {uiState?.showUserProfile && (
        <UserProfileModal onClose={() => toggleUiState('showUserProfile', false)} />
      )}

      {/* 💡 [통합] Project Manage Modal (Create, Settings/Edit, Detail) */}
      {/* Create Project Modal */}
      {uiState?.showCreateProject && (
        <ProjectManageModal
          workspaceId={currentWorkspaceId}
          onClose={() => toggleUiState('showCreateProject', false)}
          onProjectSaved={fetchProjects}
          onProjectCreated={handleProjectCreated}
          userRole={currentRole.current} // 💡 역할 전달
          initialMode="create" // 💡 모드 지정
        />
      )}

      {/* Project Settings Modal (Edit/Settings) */}
      {uiState?.showProjectSettings && selectedProject && (
        <ProjectManageModal
          workspaceId={currentWorkspaceId}
          project={selectedProject}
          onClose={() => toggleUiState('showProjectSettings', false)}
          onProjectSaved={fetchProjects}
          userRole={currentRole.current} // 💡 역할 전달
          initialMode="edit" // 💡 모드 지정 (설정/수정으로 바로 진입)
        />
      )}

      {/* Project Detail Modal (Detail) */}
      {uiState?.showProjectDetail && selectedProject && (
        <ProjectManageModal
          workspaceId={currentWorkspaceId}
          project={selectedProject}
          onClose={() => toggleUiState('showProjectDetail', false)}
          onProjectSaved={fetchProjects}
          userRole={currentRole.current} // 💡 역할 전달
          initialMode="detail" // 💡 모드 지정 (상세 보기로 바로 진입)
        />
      )}

      {/* 💡 Custom Field Add Modal (필드 추가/정의) */}
      {uiState?.showManageModal && selectedProject && (
        <CustomFieldManageModal
          editFieldData={editFieldData}
          filedTypesLookup={fieldTypesLookup}
          projectId={selectedProject.projectId}
          onClose={() => toggleUiState('showManageModal', false)}
          afterFieldCreated={afterFieldCreated}
        />
      )}

      {/* Create/Edit Board Modal */}
      {(editBoardData || uiState?.showCreateBoard) && selectedProject && (
        <BoardManageModal
          projectId={selectedProject?.projectId}
          editData={editBoardData}
          workspaceId={currentWorkspaceId}
          onClose={() => {
            setEditBoardData(null);
            toggleUiState('showCreateBoard', false);
          }}
          handleCustomField={handleCustomField}
          onBoardCreated={handleBoardContentUpdate}
          fieldOptionsLookup={fieldOptionsLookup}
        />
      )}
    </MainLayout>
  );
};

export default MainDashboard;
