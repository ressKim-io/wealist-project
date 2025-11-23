// src/components/layout/ProjectHeader.tsx

import React, { useRef, useEffect } from 'react';
import { ChevronDown, Plus, ClipboardCheck } from 'lucide-react';
import { useTheme } from '../../contexts/ThemeContext';
import { ProjectResponse } from '../../types/board';
import { WorkspaceMemberResponse } from '../../types/user';
import { AvatarStack } from '../common/AvartarStack';
import { IROLES } from '../../types/common';

interface ProjectHeaderProps {
  // Data
  projects: ProjectResponse[];
  selectedProject: ProjectResponse | null;
  workspaceMembers: WorkspaceMemberResponse[];
  userRole: IROLES; // 💡 [추가] 사용자 역할 정보 (권한 제어용)

  // State Handlers
  setSelectedProject: (project: ProjectResponse | null) => void;
  setShowCreateProject: (show: boolean) => void;
  setShowProjectDetail: (show: boolean) => void; // 💡 [추가] 프로젝트 상세 모달 핸들러
  // UI State
  showProjectSelector: boolean;
  setShowProjectSelector: (show: boolean) => void;
}

const sidebarWidth = 'w-16 sm:w-20'; // MainLayout과 동일한 값 사용

export const ProjectHeader: React.FC<ProjectHeaderProps> = ({
  projects,
  selectedProject,
  workspaceMembers,
  userRole, // 💡 [추가] userRole
  setSelectedProject,
  setShowCreateProject,
  setShowProjectDetail, // 💡 [추가] Project Detail 핸들러
  showProjectSelector,
  setShowProjectSelector,
}) => {
  const { theme } = useTheme();
  const projectSelectorRef = useRef<HTMLDivElement>(null);

  // 💡 [수정] 외부 클릭 감지 로직
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      // 💡 [핵심 수정] event.target이 Element인지 확인하고 타입 캐스팅
      const target = event.target as Element;

      // 드롭다운이 열려 있고, 클릭 위치가 드롭다운 컨테이너 바깥일 때 닫기
      if (
        showProjectSelector &&
        projectSelectorRef.current &&
        !projectSelectorRef.current.contains(target)
      ) {
        // 버튼 영역(.project-selector-trigger)을 제외하고 닫도록 처리합니다.
        // Node 타입에는 closest가 없으므로, Element로 접근 가능한지 확인합니다.
        if (target.closest) {
          // 버튼 자체를 클릭했을 때는 닫지 않습니다 (onClick 핸들러가 토글 역할을 수행하므로)
          if (!target.closest('.project-selector-trigger')) {
            setShowProjectSelector(false);
          }
        } else {
          // closest 메서드를 지원하지 않는 경우 (예외 처리)
          setShowProjectSelector(false);
        }
      }
    };

    // 드롭다운이 열려 있을 때만 리스너 활성화
    if (showProjectSelector) {
      document.addEventListener('mousedown', handleClickOutside);
    }

    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, [showProjectSelector, setShowProjectSelector]);

  return (
    <header
      className={`fixed top-0 left-0 h-16 flex items-center justify-between pl-20 pr-6 sm:pl-28 sm:pr-4 py-2 sm:py-3 ${theme.colors.card} shadow-md z-20 w-full`}
      style={{
        width: `calc(100% - ${sidebarWidth})`,
        left: sidebarWidth,
      }}
    >
      <div className="flex items-center gap-1 relative">
        <button
          onClick={() => {
            setShowProjectSelector(!showProjectSelector);
          }}
          className={`flex items-center gap-2 font-bold text-xl ${theme.colors.text} hover:opacity-80 transition project-selector-trigger`}
        >
          {selectedProject?.name || '프로젝트를 생성하세요'}
        </button>
        {/* 💡 [수정/통합] 프로젝트 상세 정보 버튼 (모든 사용자 접근 가능) */}
        {selectedProject && (
          <button
            onClick={() => setShowProjectDetail(true)}
            className={`p-2 rounded-lg transition ${theme.colors.text} hover:bg-gray-100`}
            title="프로젝트 상세 정보 보기"
          >
            <ClipboardCheck className="w-5 h-5 text-gray-600" />
          </button>
        )}
        <ChevronDown
          onClick={() => setShowProjectSelector(!showProjectSelector)}
          className={`w-5 h-5 text-gray-500 transition-transform ${
            showProjectSelector ? 'rotate-180' : 'rotate-0'
          } project-selector-trigger`}
          style={{ strokeWidth: 2.5 }}
        />
        {showProjectSelector && (
          <div
            ref={projectSelectorRef}
            className={`absolute top-full -left-6 top-8 mt-1 w-80 ${theme.colors.card} ${theme.effects.cardBorderWidth} ${theme.colors.border} z-50 ${theme.effects.borderRadius}`}
          >
            <div className="p-3 max-h-80 overflow-y-auto">
              <h3 className="text-xs text-gray-400 mb-2 px-1 font-semibold">
                프로젝트 ({projects?.length})
              </h3>
              {projects?.length === 0 ? (
                <p className="text-sm text-gray-500 p-2">프로젝트가 없습니다.</p>
              ) : (
                projects?.map((project) => (
                  <button
                    key={project.projectId}
                    onClick={() => {
                      setSelectedProject(project);
                      setShowProjectSelector(false);
                    }}
                    className={`w-full px-3 py-2 text-left text-sm rounded transition truncate ${
                      selectedProject?.projectId === project.projectId
                        ? 'bg-blue-100 text-blue-700 font-semibold'
                        : 'hover:bg-gray-100 text-gray-800'
                    }`}
                  >
                    # {project.name}
                  </button>
                ))
              )}
            </div>
            <div className="pt-2 pb-2 border-t">
              <button
                onClick={() => {
                  setShowCreateProject(true);
                  setShowProjectSelector(false);
                }}
                className="w-full px-3 py-2 text-left text-sm flex items-center gap-2 text-blue-500 hover:bg-gray-100 rounded-b-lg transition"
              >
                <Plus className="w-4 h-4" /> 새 프로젝트
              </button>
            </div>
          </div>
        )}
      </div>
      {selectedProject && (
        <button
          className={`flex items-center gap-2 p-1 rounded-lg transition hover:bg-gray-100`}
          title="조직원"
        >
          <AvatarStack members={workspaceMembers} />
        </button>
      )}
    </header>
  );
};
