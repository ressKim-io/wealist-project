import { useNavigate, useParams } from 'react-router-dom';
import React, { useEffect, useState, useRef } from 'react';
import {
  ChevronDown,
  Plus,
  Home,
  Bell,
  MessageSquare,
  Briefcase,
  File,
  Settings,
  ArrowUp,
  ArrowDown,
} from 'lucide-react';
import { useTheme } from '../contexts/ThemeContext';
import UserProfileModal from '../components/modals/UserProfileModal';
import { UserProfile } from '../types';
import { ProjectModal } from '../components/modals/ProjectModal';
import { CreateBoardModal } from '../components/modals/CreateBoardModal';
import { CustomFieldManageModal } from '../components/modals/CustomFieldManageModal';
import { FilterBar } from '../components/FilterBar';
import {
  getProjects,
  getBoards,
  getProjectStages,
  ProjectResponse,
  BoardResponse,
  CustomStageResponse,
} from '../api/board/boardService';
import { getDefaultColorByIndex } from '../constants/colors';
import { WorkspaceMember, getWorkspaceMembers } from '../api/user/userService';
import { BoardDetailModal } from '../components/modals/BoardDetailModal';
import { LoadingSpinner } from '../components/common/LoadingSpinner';

interface Column {
  stage_id: string;
  title: string;
  color?: string; // hex color from API
  boards: BoardResponse[];
}

// App.tsx에서 onLogout을 받도록 수정됨
interface MainDashboardProps {
  onLogout: () => void;
}

// =============================================================================
// AvatarStack (워크스페이스 회원)
// =============================================================================
interface AvatarStackProps {
  members: WorkspaceMember[];
}

const AvatarStack: React.FC<AvatarStackProps> = ({ members }) => {
  const displayCount = 3;
  const displayMembers = members.slice(0, displayCount);
  const remainingCount = members.length - displayCount;

  const getColorByIndex = (index: number) => {
    const colors = [
      'bg-indigo-500',
      'bg-pink-500',
      'bg-green-500',
      'bg-purple-500',
      'bg-yellow-500',
    ];
    return colors[index % colors.length];
  };

  return (
    <div className="flex -space-x-1.5 p-1 pr-0 overflow-hidden">
      {displayMembers.map((member, index) => (
        <div
          key={member.userId}
          className="w-6 h-6 rounded-full flex items-center justify-center text-xs font-bold ring-1 ring-white overflow-hidden"
          style={{ zIndex: members.length - index }}
          title={`${member.userName} (${member.roleName})`}
        >
          {member.profileImageUrl ? (
            <img
              src={member.profileImageUrl}
              alt={member.userName}
              className="w-full h-full object-cover"
            />
          ) : (
            <div
              className={`w-full h-full flex items-center justify-center text-white ${getColorByIndex(
                index,
              )}`}
            >
              {member.userName[0]}
            </div>
          )}
        </div>
      ))}
      {remainingCount > 0 && (
        <div
          className="w-6 h-6 rounded-full flex items-center justify-center text-xs font-bold ring-1 ring-white bg-gray-400 text-white"
          style={{ zIndex: 0 }}
        >
          +{remainingCount}
        </div>
      )}
    </div>
  );
};

interface AssigneeAvatarStackProps {
  assignees: string | string[];
}

// =============================================================================
// AssigneeAvatarStack (정상)
// =============================================================================
const AssigneeAvatarStack: React.FC<AssigneeAvatarStackProps> = ({ assignees }) => {
  const assigneeList = Array.isArray(assignees)
    ? assignees
    : (assignees as string)
        .split(',')
        .map((name) => name.trim())
        .filter((name) => name.length > 0);

  const initials = assigneeList.map((name) => name[0]).filter((i) => i);
  const displayCount = 3;

  if (initials.length === 0) {
    return (
      <div
        className={`w-8 h-8 rounded-full flex items-center justify-center text-xs font-bold ring-1 ring-gray-200 bg-gray-200 text-gray-700`}
      >
        ?
      </div>
    );
  }

  return (
    <div className="flex -space-x-1 p-1 pr-0 overflow-hidden">
      {initials.slice(0, displayCount).map((initial, index) => (
        <div
          key={index}
          className={`w-6 h-6 rounded-full flex items-center justify-center text-xs font-bold ring-1 ring-white text-white ${
            index === 0 ? 'bg-indigo-500' : index === 1 ? 'bg-pink-500' : 'bg-green-500'
          }`}
          style={{ zIndex: initials.length - index }}
          title={assigneeList[index]}
        >
          {initial}
        </div>
      ))}
      {initials.length > displayCount && (
        <div
          className={`w-6 h-6 rounded-full flex items-center justify-center text-xs font-bold ring-1 ring-white bg-gray-400 text-white`}
          style={{ zIndex: 0 }}
          title={`${initials.length - displayCount}명 외`}
        >
          +{initials.length - displayCount}
        </div>
      )}
    </div>
  );
};

// =============================================================================
// MainDashboard
// =============================================================================
const MainDashboard: React.FC<MainDashboardProps> = ({ onLogout }) => {
  const navigate = useNavigate();

  // 1. URL에서 :workspaceId 값을 가져옵니다.
  const { workspaceId } = useParams<{ workspaceId: string }>();
  // 2. localStorage에서 토큰을 가져옵니다.
  const accessToken = localStorage.getItem('accessToken') || '';

  // 3. prop 대신 URL 파라미터를 사용합니다.
  const currentWorkspaceId = workspaceId || '';

  // 4. 워크스페이스 로고 클릭 핸들러
  const handleBackToSelect = () => {
    navigate('/workspaces');
  };

  const { theme } = useTheme();
  const currentRole = useRef<'OWNER' | 'ORGANIZER' | 'MEMBER'>('ORGANIZER');
  const canAccessSettings = currentRole.current === 'OWNER' || currentRole.current === 'ORGANIZER';
  // 상태 관리
  const [projects, setProjects] = useState<ProjectResponse[]>([]);
  const [columns, setColumns] = useState<Column[]>([]);
  const [selectedProject, setSelectedProject] = useState<ProjectResponse | null>(null);
  const [workspaceMembers, setWorkspaceMembers] = useState<WorkspaceMember[]>([]);

  const [userProfile, _setUserProfile] = useState<UserProfile>({
    profileId: '',
    userId: '',
    nickName: 'User',
    email: 'user@example.com',
    profileImageUrl: null,
    createdAt: '',
    updatedAt: '',
  });

  // UI 상태
  const [showUserMenu, setShowUserMenu] = useState<boolean>(false);
  const [showProjectSelector, setShowProjectSelector] = useState<boolean>(false);
  const [showUserProfile, setShowUserProfile] = useState<boolean>(false);
  const [showCreateProject, setShowCreateProject] = useState<boolean>(false);
  const [showCreateBoard, setShowCreateBoard] = useState<boolean>(false);
  const [createBoardStageId, setCreateBoardStageId] = useState<string>('');
  const [selectedBoardId, setSelectedBoardId] = useState<string | null>(null);
  const [editBoardData, setEditBoardData] = useState<any>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showManageModal, setShowManageModal] = useState<boolean>(false);
  const [showProjectSettings, setShowProjectSettings] = useState<boolean>(false);

  // Filter/View 상태
  const [currentView, setCurrentView] = useState<'stage' | 'role'>('stage');
  const [searchQuery, setSearchQuery] = useState<string>('');
  const [filterOption, setFilterOption] = useState<string>('all');
  const [currentLayout, setCurrentLayout] = useState<'table' | 'board'>('board');
  const [showCompleted, setShowCompleted] = useState<boolean>(false);

  // Table sorting state
  const [sortColumn, setSortColumn] = useState<
    'title' | 'stage' | 'role' | 'importance' | 'assignee' | 'dueDate' | null
  >(null);
  const [sortDirection, setSortDirection] = useState<'asc' | 'desc'>('asc');

  // TODO: Implement search and filter logic
  console.log('Current filters:', {
    currentView,
    searchQuery,
    filterOption,
    currentLayout,
    showCompleted,
  });
  useEffect(() => {
    console.log(`Selected Board ID changed: ${selectedBoardId}`);
  }, [selectedBoardId]);

  // Ref
  const userMenuRef = useRef<HTMLDivElement>(null);
  const projectSelectorRef = useRef<HTMLDivElement>(null);

  // 1. 프로젝트 목록 조회 함수 (재사용 가능)
  const fetchProjects = React.useCallback(async () => {
    if (!currentWorkspaceId || !accessToken) return;

    setIsLoading(true);
    setError(null);
    console.log(currentWorkspaceId);
    try {
      console.log(`[Dashboard] 프로젝트 로드 시작 (Workspace: ${currentWorkspaceId})`);
      const fetchedProjects = await getProjects(currentWorkspaceId, accessToken);
      console.log('✅ Projects loaded:', fetchedProjects);

      setProjects(fetchedProjects);

      if (fetchedProjects.length > 0) {
        setSelectedProject(fetchedProjects[0]);
      } else {
        setSelectedProject(null);
        setColumns([]);
      }
    } catch (err) {
      const error = err as Error;
      console.error('❌ 프로젝트 로드 실패:', error);
      setError(`프로젝트 로드 실패: ${error.message}`);
      setProjects([]);
      setColumns([]);
    } finally {
      setIsLoading(false);
    }
  }, [currentWorkspaceId, accessToken]);

  // 2. 워크스페이스 회원 조회 함수
  const fetchWorkspaceMembers = React.useCallback(async () => {
    if (!currentWorkspaceId || !accessToken) return;

    try {
      console.log(`[Dashboard] 워크스페이스 회원 로드 시작 (Workspace: ${currentWorkspaceId})`);
      const members = await getWorkspaceMembers(currentWorkspaceId, accessToken);
      console.log('✅ Workspace members loaded:', members);
      setWorkspaceMembers(members);
    } catch (err) {
      const error = err as Error;
      console.error('❌ 워크스페이스 회원 로드 실패:', error);
      setWorkspaceMembers([]);
    }
  }, [currentWorkspaceId, accessToken]);

  // 3. 초기 로드
  useEffect(() => {
    fetchProjects();
    fetchWorkspaceMembers();
    handleDragEnd();
  }, [fetchProjects, fetchWorkspaceMembers]);

  // 4. 보드 목록 조회 함수 (재사용 가능)
  const fetchBoards = React.useCallback(async () => {
    if (!selectedProject || !accessToken) {
      setColumns([]);
      return;
    }

    setIsLoading(true);
    setError(null);
    console.log(selectedProject);
    try {
      console.log(`[Dashboard] 보드 로드 시작 (Project: ${selectedProject.name})`);

      // 1. 프로젝트의 모든 Stages 조회
      const stages = await getProjectStages(selectedProject.project_id, accessToken);
      console.log('✅ Stages loaded:', stages);

      // 2. 보드 조회
      const boardsResponse = await getBoards(selectedProject.project_id, accessToken);
      console.log('✅ Boards loaded:', boardsResponse);

      // 3. Stage별로 빈 컬럼 먼저 생성
      const stageMap = new Map<string, { stage: CustomStageResponse; boards: BoardResponse[] }>();
      stages.forEach((stage) => {
        stageMap.set(stage.stage_id, { stage, boards: [] });
      });

      // 4. 보드를 해당 Stage 컬럼에 추가
      boardsResponse.boards.forEach((board) => {
        const stage_id = board.stage?.stage_id;
        if (stage_id && stageMap.has(stage_id)) {
          stageMap.get(stage_id)!.boards.push(board);
        }
      });

      // 5. Column 형식으로 변환 (displayOrder 순서대로)
      const sortedStages = Array.from(stageMap.values()).sort(
        (a, b) => a.stage.displayOrder - b.stage.displayOrder,
      );

      const columns: Column[] = sortedStages.map(({ stage, boards }) => ({
        stage_id: stage.stage_id,
        title: stage.name,
        color: stage.color, // Store the color from API
        boards: boards,
      }));

      setColumns(columns);
    } catch (err) {
      const error = err as Error;
      console.error('❌ 보드 로드 실패:', error);
      setError(`보드 로드 실패: ${error.message}`);
      setColumns([]);
    } finally {
      setIsLoading(false);
    }
  }, [selectedProject, accessToken]);

  // 4. 프로젝트 선택 시 보드 로드
  useEffect(() => {
    fetchBoards();
  }, [fetchBoards]);

  // 2. 드래그 앤 드롭 (용어 변경)
  const [draggedBoard, setDraggedBoard] = useState<BoardResponse | null>(null);
  const [draggedFromColumn, setDraggedFromColumn] = useState<string | null>(null);
  const [draggedColumn, setDraggedColumn] = useState<Column | null>(null);
  const [dragOverBoardId, setDragOverBoardId] = useState<string | null>(null);
  const [dragOverColumn, setDragOverColumn] = useState<string | null>(null);

  const handleDragStart = (board: BoardResponse, columnId: string): void => {
    setDraggedBoard(board);
    setDraggedFromColumn(columnId);
  };

  const handleDragOver = (e: React.DragEvent<HTMLDivElement>): void => {
    e.preventDefault();
  };

  const handleDrop = async (targetColumnId: string): Promise<void> => {
    if (!draggedBoard || !draggedFromColumn) return;

    // Reset drag over state
    setDragOverColumn(null);

    // Same column: reorder boards within column
    if (draggedFromColumn === targetColumnId) {
      if (!dragOverBoardId || dragOverBoardId === draggedBoard.board_id) {
        setDraggedBoard(null);
        setDraggedFromColumn(null);
        setDragOverBoardId(null);
        return;
      }

      const targetColumn = columns.find((col) => col.stage_id === targetColumnId);
      if (!targetColumn || !selectedProject) {
        setDraggedBoard(null);
        setDraggedFromColumn(null);
        setDragOverBoardId(null);
        return;
      }

      // Reorder boards
      const draggedIndex = targetColumn.boards.findIndex(
        (b) => b.board_id === draggedBoard.board_id,
      );
      const targetIndex = targetColumn.boards.findIndex((b) => b.board_id === dragOverBoardId);

      if (draggedIndex === -1 || targetIndex === -1) {
        setDraggedBoard(null);
        setDraggedFromColumn(null);
        setDragOverBoardId(null);
        return;
      }

      const newBoards = [...targetColumn.boards];
      const [removed] = newBoards.splice(draggedIndex, 1);
      newBoards.splice(targetIndex, 0, removed);

      const newColumns = columns.map((col) => {
        if (col.stage_id === targetColumnId) {
          return { ...col, boards: newBoards };
        }
        return col;
      });

      setColumns(newColumns);
      setDraggedBoard(null);
      setDraggedFromColumn(null);
      setDragOverBoardId(null);

      console.log(`✅ Stage 내 Board 순서 변경 (로컬)`);
      return;
    }

    const updatedBoard: BoardResponse = {
      ...draggedBoard,
      stage: { ...draggedBoard.stage!, id: targetColumnId },
    };

    // Optimistic UI update
    const newColumns = columns.map((col) => {
      if (col.stage_id === draggedFromColumn) {
        return { ...col, boards: col.boards.filter((t) => t.board_id !== draggedBoard.board_id) };
      }
      if (col.stage_id === targetColumnId) {
        // Insert at the position indicated by dragOverBoardId
        if (dragOverBoardId) {
          const targetIndex = col.boards.findIndex((b) => b.board_id === dragOverBoardId);
          if (targetIndex !== -1) {
            const newBoards = [...col.boards];
            newBoards.splice(targetIndex, 0, updatedBoard);
            return { ...col, boards: newBoards };
          }
        }
        // If no dragOverBoardId, add to the end
        return { ...col, boards: [...col.boards, updatedBoard] };
      }
      return col;
    });

    setColumns(newColumns);
    setDraggedBoard(null);
    setDraggedFromColumn(null);
    setDragOverBoardId(null);

    console.log(`✅ Board ${draggedBoard.board_id} Stage 변경 (로컬): ${targetColumnId}`);
  };

  // Column drag handlers
  const handleColumnDragStart = (column: Column): void => {
    setDraggedColumn(column);
  };

  const handleColumnDragOver = (e: React.DragEvent<HTMLDivElement>): void => {
    e.preventDefault();
  };

  const handleColumnDrop = async (targetColumn: Column): Promise<void> => {
    if (!draggedColumn || draggedColumn.stage_id === targetColumn.stage_id) {
      setDraggedColumn(null);
      return;
    }

    const draggedIndex = columns.findIndex((col) => col.stage_id === draggedColumn.stage_id);
    const targetIndex = columns.findIndex((col) => col.stage_id === targetColumn.stage_id);

    if (draggedIndex === -1 || targetIndex === -1) {
      setDraggedColumn(null);
      return;
    }

    // Reorder columns
    const newColumns = [...columns];
    const [removed] = newColumns.splice(draggedIndex, 1);
    newColumns.splice(targetIndex, 0, removed);

    setColumns(newColumns);
    setDraggedColumn(null);

    console.log(`✅ Stage 컬럼 순서 변경 (로컬)`);
  };

  // 💡 [추가] 드래그 종료 시 상태를 초기화하는 핸들러
  const handleDragEnd = (): void => {
    // 마우스를 놓았을 때, 드래그 상태와 컬럼 드래그 상태를 모두 초기화
    setDraggedBoard(null);
    setDraggedFromColumn(null);
    setDraggedColumn(null);
    setDragOverBoardId(null);
    setDragOverColumn(null);
    console.log('✅ Drag End: 드래그 상태 초기화');
  };

  // Table sorting handler
  const handleSort = (
    column: 'title' | 'stage' | 'role' | 'importance' | 'assignee' | 'dueDate',
  ) => {
    if (sortColumn === column) {
      // Toggle direction if same column
      setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc');
    } else {
      // New column, default to ascending
      setSortColumn(column);
      setSortDirection('asc');
    }
  };

  // 외부 클릭 감지 (동일)
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (userMenuRef.current && !userMenuRef.current.contains(event.target as Node)) {
        setShowUserMenu(false);
      }
      if (
        showProjectSelector &&
        projectSelectorRef.current &&
        !projectSelectorRef.current.contains(event.target as Node)
      ) {
        setShowProjectSelector(false);
      }
    };
    document.addEventListener('mousedown', handleClickOutside);
    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, [showProjectSelector]);

  const sidebarWidth = 'w-16 sm:w-20';

  return (
    <div className={`min-h-screen flex ${theme.colors.background} relative`}>
      {/* 백그라운드 패턴 (동일) */}
      <div
        className="fixed inset-0 opacity-5"
        style={{
          backgroundImage:
            'linear-gradient(#000 1px, transparent 1px), linear-gradient(90deg, #000 1px, transparent 1px)',
          backgroundSize: '20px 20px',
        }}
      ></div>

      {/* 사이드바 */}
      <aside
        className={`${sidebarWidth} fixed top-0 left-0 h-full flex flex-col justify-between ${theme.colors.primary} text-white shadow-xl z-50 flex-shrink-0`}
      >
        <div className="flex flex-col flex-grow items-center">
          {/* 3. 워크스페이스 로고 클릭 기능 추가 (스타일 복구) */}
          <div className={`py-3 flex justify-center w-full relative`}>
            <button
              onClick={handleBackToSelect}
              title="워크스페이스 목록으로"
              // ✅ UI 깨짐 문제 해결: className 복구
              className={`w-12 h-12 rounded-lg mx-auto flex items-center justify-center text-xl font-bold transition 
                    bg-white text-blue-800 ring-2 ring-white/50 hover:bg-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-300`}
            >
              {currentWorkspaceId.slice(0, 1).toUpperCase()}
            </button>
          </div>

          {/* 사이드바 메뉴 (동일) */}
          <div className="flex flex-col gap-2 mt-4 flex-grow px-2 w-full pt-4">
            <button
              className={`w-12 h-12 rounded-lg mx-auto flex items-center justify-center transition bg-blue-600 text-white ring-2 ring-white/50`}
              title="홈"
            >
              <Home className="w-6 h-6" />
            </button>
            <button
              className={`w-12 h-12 rounded-lg mx-auto flex items-center justify-center bg-gray-700 hover:bg-gray-600 text-white opacity-50 transition`}
              title="DM"
            >
              <MessageSquare className="w-6 h-6" />
            </button>
            <button
              className={`w-12 h-12 rounded-lg mx-auto flex items-center justify-center bg-gray-700 hover:bg-gray-600 text-white opacity-50 transition`}
              title="알림"
            >
              <Bell className="w-6 h-6" />
            </button>
            <button
              className={`w-12 h-12 rounded-lg mx-auto flex items-center justify-center bg-gray-700 hover:bg-gray-600 text-white opacity-50 transition`}
              title="파일"
            >
              <File className="w-6 h-6" />
            </button>
          </div>
        </div>

        {/* 하단 유저 메뉴 (동일) */}
        <div className={`py-3 px-2 border-t border-gray-700`}>
          <button
            onClick={() => setShowUserMenu(!showUserMenu)}
            className={`w-full flex items-center justify-center py-2 text-sm rounded-lg hover:bg-blue-600 transition relative`}
            title="계정 메뉴"
          >
            <div
              className={`w-10 h-10 rounded-full bg-gray-300 flex items-center justify-center text-sm font-bold ring-2 ring-white/50 text-gray-700 overflow-hidden`}
            >
              {userProfile.profileImageUrl ? (
                <img
                  src={userProfile.profileImageUrl}
                  alt={userProfile.nickName}
                  className="w-full h-full object-cover"
                />
              ) : (
                userProfile.nickName[0]?.toUpperCase() || 'U'
              )}
            </div>
          </button>
        </div>
      </aside>

      {/* 메인 콘텐츠 (동일) */}
      <div
        className="flex-grow flex flex-col relative z-10"
        style={{ marginLeft: sidebarWidth, minHeight: '100vh' }}
      >
        {/* 헤더 (동일) */}
        <header
          className={`fixed top-0 left-0 h-16 flex items-center justify-between pl-20 pr-6 sm:pl-28 sm:pr-4 py-2 sm:py-3 ${theme.colors.card} shadow-md z-20 w-full`}
          style={{
            width: `calc(100% - ${sidebarWidth})`,
            left: sidebarWidth,
          }}
        >
          <div className="flex items-center gap-1 relative">
            <button
              onClick={() => setShowProjectSelector(!showProjectSelector)}
              className={`flex items-center gap-2 font-bold text-xl ${theme.colors.text} hover:opacity-80 transition`}
            >
              {selectedProject?.name || '프로젝트 선택'}
            </button>
            {canAccessSettings && selectedProject && (
              <button
                onClick={() => setShowProjectSettings(true)}
                className={`p-2 rounded-lg transition ${theme.colors.text} hover:bg-gray-100`}
                title="프로젝트 설정"
              >
                <Settings className="w-5 h-5" />
              </button>
            )}
            <ChevronDown
              onClick={() => setShowProjectSelector(!showProjectSelector)}
              className={`w-5 h-5 text-gray-500 transition-transform ${
                showProjectSelector ? 'rotate-180' : 'rotate-0'
              }`}
              style={{ strokeWidth: 2.5 }}
            />
            {showProjectSelector && (
              <div
                ref={projectSelectorRef}
                className={`absolute top-full -left-6 top-8 mt-1 w-80 ${theme.colors.card} ${theme.effects.cardBorderWidth} ${theme.colors.border} z-50 ${theme.effects.borderRadius}`}
              >
                <div className="p-3 max-h-80 overflow-y-auto">
                  <h3 className="text-xs text-gray-400 mb-2 px-1 font-semibold">
                    프로젝트 ({projects.length})
                  </h3>
                  {projects.length === 0 ? (
                    <p className="text-sm text-gray-500 p-2">프로젝트가 없습니다.</p>
                  ) : (
                    projects?.map((project) => (
                      <button
                        key={project.project_id}
                        onClick={() => {
                          setSelectedProject(project);
                          setShowProjectSelector(false);
                        }}
                        className={`w-full px-3 py-2 text-left text-sm rounded transition truncate ${
                          selectedProject?.project_id === project.project_id
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
              className={`flex items-center gap-2 p-1 rounded-lg transition ${
                canAccessSettings ? 'hover:bg-blue-100' : 'hover:bg-gray-100'
              }`}
              title="조직원"
            >
              <AvatarStack members={workspaceMembers} />
            </button>
          )}
        </header>

        {/* 보드 영역 (동일) */}
        <div className="flex-grow flex flex-col p-3 sm:p-6 overflow-auto mt-16 ml-20">
          {error && (
            <div className="mb-4 p-4 bg-red-50 border border-red-300 rounded-lg text-red-700">
              {error}
            </div>
          )}

          {isLoading && projects.length === 0 ? (
            <LoadingSpinner message="프로젝트를 로드 중..." />
          ) : selectedProject ? (
            <>
              {/* FilterBar */}
              <FilterBar
                onSearchChange={setSearchQuery}
                onViewChange={setCurrentView}
                onFilterChange={setFilterOption}
                onManageClick={() => setShowManageModal(true)}
                currentView={currentView}
                onLayoutChange={setCurrentLayout}
                onShowCompletedChange={setShowCompleted}
                currentLayout={currentLayout}
                showCompleted={showCompleted}
              />

              {/* Boards or Table */}
              {currentLayout === 'table' ? (
                // Table Layout
                <div className="mt-4 overflow-x-auto">
                  <table
                    className={`w-full ${theme.colors.card} ${theme.effects.borderRadius} overflow-hidden shadow-lg`}
                  >
                    <thead className="bg-gray-100 border-b border-gray-200">
                      <tr>
                        {/* Title Column */}
                        <th className="px-4 py-3 text-left">
                          <button
                            onClick={() => handleSort('title')}
                            className="flex items-center gap-2 font-semibold text-sm text-gray-700 hover:text-blue-600 transition"
                          >
                            제목
                            {sortColumn === 'title' &&
                              (sortDirection === 'asc' ? (
                                <ArrowUp className="w-4 h-4" />
                              ) : (
                                <ArrowDown className="w-4 h-4" />
                              ))}
                          </button>
                        </th>
                        {/* Stage Column */}
                        <th className="px-4 py-3 text-left">
                          <button
                            onClick={() => handleSort('stage')}
                            className="flex items-center gap-2 font-semibold text-sm text-gray-700 hover:text-blue-600 transition"
                          >
                            진행 단계
                            {sortColumn === 'stage' &&
                              (sortDirection === 'asc' ? (
                                <ArrowUp className="w-4 h-4" />
                              ) : (
                                <ArrowDown className="w-4 h-4" />
                              ))}
                          </button>
                        </th>
                        {/* Role Column */}
                        <th className="px-4 py-3 text-left">
                          <button
                            onClick={() => handleSort('role')}
                            className="flex items-center gap-2 font-semibold text-sm text-gray-700 hover:text-blue-600 transition"
                          >
                            역할
                            {sortColumn === 'role' &&
                              (sortDirection === 'asc' ? (
                                <ArrowUp className="w-4 h-4" />
                              ) : (
                                <ArrowDown className="w-4 h-4" />
                              ))}
                          </button>
                        </th>
                        {/* Importance Column */}
                        <th className="px-4 py-3 text-left">
                          <button
                            onClick={() => handleSort('importance')}
                            className="flex items-center gap-2 font-semibold text-sm text-gray-700 hover:text-blue-600 transition"
                          >
                            중요도
                            {sortColumn === 'importance' &&
                              (sortDirection === 'asc' ? (
                                <ArrowUp className="w-4 h-4" />
                              ) : (
                                <ArrowDown className="w-4 h-4" />
                              ))}
                          </button>
                        </th>
                        {/* Assignee Column */}
                        <th className="px-4 py-3 text-left">
                          <button
                            onClick={() => handleSort('assignee')}
                            className="flex items-center gap-2 font-semibold text-sm text-gray-700 hover:text-blue-600 transition"
                          >
                            담당자
                            {sortColumn === 'assignee' &&
                              (sortDirection === 'asc' ? (
                                <ArrowUp className="w-4 h-4" />
                              ) : (
                                <ArrowDown className="w-4 h-4" />
                              ))}
                          </button>
                        </th>
                        {/* Due Date Column */}
                        <th className="px-4 py-3 text-left">
                          <button
                            onClick={() => handleSort('dueDate')}
                            className="flex items-center gap-2 font-semibold text-sm text-gray-700 hover:text-blue-600 transition"
                          >
                            마감일
                            {sortColumn === 'dueDate' &&
                              (sortDirection === 'asc' ? (
                                <ArrowUp className="w-4 h-4" />
                              ) : (
                                <ArrowDown className="w-4 h-4" />
                              ))}
                          </button>
                        </th>
                      </tr>
                    </thead>
                    <tbody>
                      {(() => {
                        // Flatten all boards from all columns
                        const allBoards = columns.flatMap((column) =>
                          column.boards.map((board) => ({
                            ...board,
                            stageName: column.title,
                            stageColor: column.color,
                          })),
                        );

                        // Filter boards based on search query
                        const filteredBoards = searchQuery.trim()
                          ? allBoards.filter((board) => {
                              const query = searchQuery.toLowerCase();
                              const titleMatch = board.title.toLowerCase().includes(query);
                              const contentMatch = board.content?.toLowerCase().includes(query);
                              return titleMatch || contentMatch;
                            })
                          : allBoards;

                        // Sort boards
                        const sortedBoards = [...filteredBoards].sort((a, b) => {
                          if (!sortColumn) return 0;

                          let aValue: any;
                          let bValue: any;

                          switch (sortColumn) {
                            case 'title':
                              aValue = a.title.toLowerCase();
                              bValue = b.title.toLowerCase();
                              break;
                            case 'stage':
                              aValue = a.stageName.toLowerCase();
                              bValue = b.stageName.toLowerCase();
                              break;
                            case 'role':
                              aValue = a.roles?.[0]?.name?.toLowerCase() || '';
                              bValue = b.roles?.[0]?.name?.toLowerCase() || '';
                              break;
                            case 'importance':
                              aValue = a.importance?.level || 0;
                              bValue = b.importance?.level || 0;
                              break;
                            case 'assignee':
                              aValue = a.assignee?.name?.toLowerCase() || '';
                              bValue = b.assignee?.name?.toLowerCase() || '';
                              break;
                            case 'dueDate':
                              aValue = a.dueDate ? new Date(a.dueDate).getTime() : 0;
                              bValue = b.dueDate ? new Date(b.dueDate).getTime() : 0;
                              break;
                            default:
                              return 0;
                          }

                          if (aValue < bValue) return sortDirection === 'asc' ? -1 : 1;
                          if (aValue > bValue) return sortDirection === 'asc' ? 1 : -1;
                          return 0;
                        });

                        return sortedBoards.map((board) => (
                          <tr
                            key={board.board_id}
                            onClick={() => setSelectedBoardId(board.board_id)}
                            className="border-b border-gray-200 hover:bg-gray-50 cursor-pointer transition"
                          >
                            {/* Title */}
                            <td className="px-4 py-3 font-semibold text-gray-800">{board.title}</td>
                            {/* Stage */}
                            <td className="px-4 py-3">
                              <div className="flex items-center gap-2">
                                <span
                                  className="w-3 h-3 rounded-full"
                                  style={{ backgroundColor: board.stageColor || '#6B7280' }}
                                />
                                <span className="text-sm">{board.stageName}</span>
                              </div>
                            </td>
                            {/* Role */}
                            <td className="px-4 py-3">
                              {board.roles && board.roles.length > 0 ? (
                                <div className="flex items-center gap-2">
                                  <span
                                    className="w-3 h-3 rounded-full"
                                    style={{ backgroundColor: board.roles[0].color || '#6B7280' }}
                                  />
                                  <span className="text-sm">{board.roles[0].name}</span>
                                </div>
                              ) : (
                                <span className="text-sm text-gray-500">없음</span>
                              )}
                            </td>
                            {/* Importance */}
                            <td className="px-4 py-3">
                              {board.importance ? (
                                <div className="flex items-center gap-2">
                                  <span
                                    className="w-3 h-3 rounded-full"
                                    style={{ backgroundColor: board.importance.color || '#6B7280' }}
                                  />
                                  <span className="text-sm">{board.importance.name}</span>
                                </div>
                              ) : (
                                <span className="text-sm text-gray-500">없음</span>
                              )}
                            </td>
                            {/* Assignee */}
                            <td className="px-4 py-3">
                              <AssigneeAvatarStack
                                assignees={board.assignee?.name || 'Unassigned'}
                              />
                            </td>
                            {/* Due Date */}
                            <td className="px-4 py-3 text-sm text-gray-600">
                              {board.dueDate
                                ? new Date(board.dueDate).toLocaleDateString('ko-KR')
                                : '없음'}
                            </td>
                          </tr>
                        ));
                      })()}

                      {/* Add Board Row */}
                      <tr
                        onClick={() => {
                          setCreateBoardStageId('');
                          setShowCreateBoard(true);
                        }}
                        className="border-t-2 border-gray-300 hover:bg-blue-50 cursor-pointer transition"
                      >
                        <td colSpan={6} className="px-4 py-4">
                          <div className="flex items-center justify-center gap-2 text-blue-600 font-semibold">
                            <Plus className="w-5 h-5" />
                            <span>보드 추가</span>
                          </div>
                        </td>
                      </tr>
                    </tbody>
                  </table>
                  {columns?.flatMap((col) => col.boards).length === 0 && (
                    <div className="text-center py-12 text-gray-500">
                      보드가 없습니다. 보드를 추가해보세요.
                    </div>
                  )}
                </div>
              ) : (
                // Board Layout
                <div className="flex flex-col lg:flex-row gap-3 sm:gap-4 min-w-max pb-4 mt-4">
                  {(() => {
                    // Filter columns based on search query
                    const filteredColumns = searchQuery.trim()
                      ? columns?.map((column) => ({
                          ...column,
                          boards: column.boards.filter((board) => {
                            const query = searchQuery.toLowerCase();
                            const titleMatch = board.title.toLowerCase().includes(query);
                            const contentMatch = board.content?.toLowerCase().includes(query);
                            return titleMatch || contentMatch;
                          }),
                        }))
                      : columns;

                    return filteredColumns?.map((column, idx) => (
                      <div
                        key={column.stage_id}
                        draggable
                        onDragStart={() => handleColumnDragStart(column)}
                        onDragEnd={handleDragEnd}
                        onDragOver={(e) => {
                          handleDragOver(e);
                          handleColumnDragOver(e);
                          if (draggedBoard && !draggedColumn) {
                            setDragOverColumn(column.stage_id);
                          }
                        }}
                        onDragLeave={() => {
                          if (draggedBoard && !draggedColumn) {
                            setDragOverColumn(null);
                          }
                        }}
                        onDrop={() => {
                          if (draggedColumn) {
                            handleColumnDrop(column);
                          } else {
                            handleDrop(column.stage_id);
                          }
                        }}
                        className={`w-full lg:w-80 lg:flex-shrink-0 relative transition-all cursor-move ${
                          draggedColumn?.stage_id === column.stage_id
                            ? 'opacity-50 scale-95 shadow-2xl rotate-2'
                            : 'opacity-100'
                        }`}
                      >
                        <div
                          className={`relative ${theme.effects.cardBorderWidth} ${
                            dragOverColumn === column.stage_id &&
                            draggedFromColumn !== column.stage_id
                              ? 'border-blue-500 border-2 bg-blue-50 dark:bg-blue-900/20 shadow-lg'
                              : theme.colors.border
                          } p-3 sm:p-4 ${theme.colors.card} ${
                            theme.effects.borderRadius
                          } transition-all duration-200`}
                        >
                          <div className={`flex items-center justify-between pb-2`}>
                            <h3
                              className={`font-bold ${theme.colors.text} flex items-center gap-2 ${theme.font.size.xs}`}
                            >
                              <span
                                className={`w-3 h-3 sm:w-4 sm:h-4 ${theme.effects.cardBorderWidth} ${theme.colors.border}`}
                                style={{
                                  backgroundColor: column.color || getDefaultColorByIndex(idx).hex,
                                }}
                              ></span>
                              {column.title}
                              <span
                                className={`bg-black text-white px-1 sm:px-2 py-1 ${theme.effects.cardBorderWidth} ${theme.colors.border} text-[8px] sm:text-xs`}
                              >
                                {column.boards.length}
                              </span>
                            </h3>
                          </div>

                          <div className="space-y-2 sm:space-y-3">
                            {column.boards.map((board) => (
                              <div
                                onDragEnd={handleDragEnd}
                                key={board.board_id + column.stage_id} // 보드 감싸는 최상위 div
                                className="relative"
                                onDragOver={(e) => {
                                  e.preventDefault();
                                  e.stopPropagation();
                                  if (draggedBoard && draggedBoard.board_id !== board.board_id) {
                                    setDragOverBoardId(board.board_id);
                                  }
                                }}
                                onDragLeave={(e) => {
                                  e.stopPropagation();
                                  setDragOverBoardId(null);
                                }}
                              >
                                {/* Drop indicator line - shows where the dragged board will be inserted */}
                                {dragOverBoardId === board.board_id &&
                                  draggedBoard &&
                                  draggedBoard.board_id !== board.board_id && (
                                    <div className="absolute -top-2 left-0 right-0 h-1 bg-blue-500 rounded-full shadow-lg shadow-blue-500/50 z-10"></div>
                                  )}
                                <div
                                  draggable
                                  onDragStart={(e) => {
                                    e.stopPropagation();
                                    handleDragStart(board, column.stage_id);
                                  }}
                                  // 💡 [추가] 보드 카드 드래그 종료 시 상태 초기화
                                  onDragEnd={handleDragEnd}
                                  onClick={() => setSelectedBoardId(board.board_id)}
                                  className={`relative ${theme.colors.card} p-3 sm:p-4 ${
                                    theme.effects.cardBorderWidth
                                  } ${
                                    theme.colors.border
                                  } hover:border-blue-500 transition-all cursor-pointer ${
                                    theme.effects.borderRadius
                                  } 
                                  ${
                                    draggedBoard?.board_id === board.board_id
                                      ? 'opacity-50 scale-95 shadow-2xl rotate-1'
                                      : 'opacity-100'
                                  }
                                  `}
                                >
                                  <h3
                                    className={`font-bold ${theme.colors.text} mb-2 sm:mb-3 ${theme.font.size.xs} break-words`}
                                  >
                                    {board.title}
                                  </h3>
                                  <div className="flex items-center justify-between">
                                    <AssigneeAvatarStack
                                      assignees={board.assignee?.name || 'Unassigned'}
                                    />
                                  </div>
                                </div>
                              </div>
                            ))}

                            {/* Drop indicator for empty column or below all boards */}
                            {dragOverColumn === column.stage_id &&
                              draggedBoard &&
                              !draggedColumn &&
                              !dragOverBoardId && (
                                <div className="relative py-2">
                                  <div className="h-1 bg-blue-500 rounded-full shadow-lg shadow-blue-500/50"></div>
                                </div>
                              )}

                            <button
                              className={`relative w-full py-3 sm:py-4 ${theme.effects.cardBorderWidth} border-dashed ${theme.colors.border} ${theme.colors.card} hover:bg-gray-100 transition flex items-center justify-center gap-2 ${theme.font.size.xs} ${theme.effects.borderRadius}`}
                              onClick={() => {
                                setCreateBoardStageId(column.stage_id);
                                setShowCreateBoard(true);
                              }}
                              onDragOver={(e) => {
                                e.preventDefault();
                                e.stopPropagation();
                                if (draggedBoard && !draggedColumn) {
                                  setDragOverColumn(column.stage_id);
                                  setDragOverBoardId(null);
                                }
                              }}
                            >
                              <Plus className="w-3 h-3 sm:w-4 sm:h-4" style={{ strokeWidth: 3 }} />
                              보드 추가
                            </button>
                          </div>
                        </div>
                      </div>
                    ));
                  })()}
                </div>
              )}
            </>
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
      </div>

      {/* 모달 (하단) (동일) */}
      {showUserMenu && (
        <div
          ref={userMenuRef}
          className={`absolute bottom-16 left-12 sm:left-16 w-64 ${theme.colors.card} ${theme.effects.cardBorderWidth} ${theme.colors.border} z-50 ${theme.effects.borderRadius} shadow-2xl`}
        >
          <div className="p-3 pb-3 mb-2 border-b border-gray-200">
            <div className="flex items-center gap-3">
              <div
                className={`w-10 h-10 ${theme.colors.primary} flex items-center justify-center text-white text-base font-bold rounded-md overflow-hidden`}
              >
                {userProfile.profileImageUrl ? (
                  <img
                    src={userProfile.profileImageUrl}
                    alt={userProfile.nickName}
                    className="w-full h-full object-cover"
                  />
                ) : (
                  userProfile.nickName[0]?.toUpperCase() || 'U'
                )}
              </div>
              <div>
                <h3 className="font-bold text-lg text-gray-900">{userProfile.nickName}</h3>
                <div className="flex items-center text-green-600 text-xs mt-1">
                  <span className="w-2 h-2 bg-green-500 rounded-full mr-1"></span>
                  대화 가능
                </div>
              </div>
            </div>
          </div>

          <div className="space-y-1 p-2 pt-0">
            <button
              onClick={() => {
                setShowUserProfile(true);
                setShowUserMenu(false);
              }}
              className="w-full text-left px-2 py-1.5 text-sm text-gray-800 hover:bg-blue-50 hover:text-blue-700 rounded transition"
            >
              프로필
            </button>
          </div>

          <div className="pt-2 pb-2 border-t border-gray-200 mx-2">
            <button
              onClick={onLogout}
              className="w-full text-left px-2 py-1.5 text-sm text-gray-800 hover:bg-red-50 hover:text-red-700 rounded transition"
            >
              로그아웃
            </button>
          </div>
        </div>
      )}

      {showUserProfile && userProfile && (
        <UserProfileModal user={userProfile} onClose={() => setShowUserProfile(false)} />
      )}

      {showCreateProject && (
        <ProjectModal
          workspaceId={currentWorkspaceId}
          onClose={() => setShowCreateProject(false)}
          onProjectSaved={fetchProjects}
        />
      )}

      {showCreateBoard && selectedProject && (
        <CreateBoardModal
          projectId={selectedProject.project_id}
          stage_id={createBoardStageId}
          editData={editBoardData}
          workspaceId={currentWorkspaceId}
          onClose={() => {
            setShowCreateBoard(false);
            setEditBoardData(null);
          }}
          onBoardCreated={fetchBoards}
        />
      )}

      {selectedBoardId && (
        <BoardDetailModal
          boardId={selectedBoardId}
          workspaceId={currentWorkspaceId}
          onClose={() => setSelectedBoardId(null)}
          onBoardUpdated={fetchBoards}
          onBoardDeleted={fetchBoards}
          onEdit={(boardData) => {
            setEditBoardData(boardData);
            setSelectedBoardId(null);
            setShowCreateBoard(true);
          }}
        />
      )}

      {/* Custom Field Manage Modal */}
      {showManageModal && selectedProject && (
        <CustomFieldManageModal
          projectId={selectedProject.project_id}
          onClose={() => setShowManageModal(false)}
          onFieldsUpdated={fetchBoards}
        />
      )}

      {/* Project Settings Modal */}
      {showProjectSettings && selectedProject && (
        <ProjectModal
          workspaceId={currentWorkspaceId}
          project={selectedProject}
          onClose={() => setShowProjectSettings(false)}
          onProjectSaved={fetchProjects}
        />
      )}
    </div>
  );
};

export default MainDashboard;
