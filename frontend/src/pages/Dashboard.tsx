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
} from 'lucide-react';
import { useTheme } from '../contexts/ThemeContext';
import UserProfileModal from '../components/modals/UserProfileModal';
import { UserProfile } from '../types';
import { Board, BoardWithCustomFields } from '../types/board';
import BoardDetailModal from '../components/modals/BoardDetailModal';
import { CreateProjectModal } from '../components/modals/CreateProjectModal';
import { CreateBoardModal } from '../components/modals/CreateBoardModal';
import { getProjects, getBoards, ProjectResponse, BoardResponse } from '../api/board/boardService';

interface Column {
  id: string;
  title: string;
  boards: BoardWithCustomFields[];
}

// App.tsx에서 onLogout을 받도록 수정됨
interface MainDashboardProps {
  onLogout: () => void;
}

// =============================================================================
// AvatarStack (정상)
// =============================================================================
const AvatarStack: React.FC = () => {
  const mockHeaderAvatars = ['김', '박', '이', '최'];
  return (
    <div className="flex -space-x-1.5 p-1 pr-0 overflow-hidden">
      {mockHeaderAvatars.slice(0, 3).map((initial, index) => (
        <div
          key={index}
          className={`w-6 h-6 rounded-full flex items-center justify-center text-xs font-bold ring-1 ring-white text-white ${
            index === 0 ? 'bg-indigo-500' : index === 1 ? 'bg-pink-500' : 'bg-green-500'
          }`}
          style={{ zIndex: mockHeaderAvatars.length - index }}
        >
          {initial}
        </div>
      ))}
      {mockHeaderAvatars.length > 3 && (
        <div
          className={`w-6 h-6 rounded-full flex items-center justify-center text-xs font-bold ring-1 ring-white bg-gray-400 text-white`}
          style={{ zIndex: 0 }}
        >
          +{mockHeaderAvatars.length - 3}
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
  const accessToken = localStorage.getItem('access_token') || '';

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

  const [userProfile, _setUserProfile] = useState<UserProfile>({
    name: 'User',
    email: 'user@example.com',
    avatar: 'U',
  });

  // UI 상태
  const [showUserMenu, setShowUserMenu] = useState<boolean>(false);
  const [showProjectSelector, setShowProjectSelector] = useState<boolean>(false);
  const [showUserProfile, setShowUserProfile] = useState<boolean>(false);
  const [showCreateProject, setShowCreateProject] = useState<boolean>(false);
  const [showCreateBoard, setShowCreateBoard] = useState<boolean>(false);
  const [createBoardStageId, setCreateBoardStageId] = useState<string>('');
  const [selectedBoard, setSelectedBoard] = useState<BoardWithCustomFields | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

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

  // 2. 초기 로드
  useEffect(() => {
    fetchProjects();
  }, [fetchProjects]);

  // 3. 보드 목록 조회 함수 (재사용 가능)
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
      const boardsResponse = await getBoards(selectedProject.id, accessToken);
      console.log('✅ Boards loaded:', boardsResponse);

      // Stage별로 보드를 그룹화
      const stageMap = new Map<string, BoardResponse[]>();

      boardsResponse.boards.forEach((board) => {
        const stageName = board.stage?.name || 'To Do';
        if (!stageMap.has(stageName)) {
          stageMap.set(stageName, []);
        }
        stageMap.get(stageName)!.push(board);
      });

      // Column 형식으로 변환
      const mockColumns: Column[] = Array.from(stageMap).map(([stageName, boards]) => ({
        id: stageName,
        title: stageName,
        boards: boards.map((b) => ({
          id: b.id,
          title: b.title,
          assignee_id: b.assignee?.userId || '',
          status: stageName,
          assignee: b.assignee?.name || 'Unassigned',
          customFieldValues: {
            'cf-stage': b.stage?.name || stageName,
            'cf-importance': b.importance?.name || 'Normal',
          },
        })),
      }));

      setColumns(mockColumns);
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
  const [draggedBoard, setDraggedBoard] = useState<BoardWithCustomFields | null>(null);
  const [draggedFromColumn, setDraggedFromColumn] = useState<string | null>(null);

  const handleDragStart = (board: Board, columnId: string): void => {
    setDraggedBoard(board as BoardWithCustomFields);
    setDraggedFromColumn(columnId);
  };

  const handleDragOver = (e: React.DragEvent<HTMLDivElement>): void => {
    e.preventDefault();
  };

  const handleDrop = (targetColumnId: string): void => {
    if (!draggedBoard || !draggedFromColumn || draggedFromColumn === targetColumnId) return;

    const updatedBoard: BoardWithCustomFields = {
      ...draggedBoard,
      status: targetColumnId,
    };

    const newColumns = columns.map((col) => {
      if (col.id === draggedFromColumn) {
        return { ...col, boards: col.boards.filter((t) => t.id !== draggedBoard.id) };
      }
      if (col.id === targetColumnId) {
        return { ...col, boards: [...col.boards, updatedBoard] };
      }
      return col;
    });

    setColumns(newColumns);
    setDraggedBoard(null);
    setDraggedFromColumn(null);

    console.log(`[Mock] Board ${draggedBoard.id} 상태를 ${targetColumnId}(으)로 변경`);
  };

  const columnColors = ['bg-blue-500', 'bg-yellow-500', 'bg-purple-500'];

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
              className={`w-10 h-10 rounded-full bg-gray-300 flex items-center justify-center text-sm font-bold ring-2 ring-white/50 text-gray-700`}
            >
              {userProfile.avatar}
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
          <div className="flex items-center gap-2 relative">
            <button
              onClick={() => setShowProjectSelector(!showProjectSelector)}
              className={`flex items-center gap-2 font-bold text-xl ${theme.colors.text} hover:opacity-80 transition`}
            >
              {selectedProject?.name || '프로젝트 선택'}
              <ChevronDown
                className={`w-5 h-5 text-gray-500 transition-transform ${
                  showProjectSelector ? 'rotate-180' : 'rotate-0'
                }`}
                style={{ strokeWidth: 2.5 }}
              />
            </button>

            {showProjectSelector && (
              <div
                ref={projectSelectorRef}
                className={`absolute top-full -left-4 mt-1 w-80 ${theme.colors.card} ${theme.effects.cardBorderWidth} ${theme.colors.border} z-50 ${theme.effects.borderRadius}`}
              >
                <div className="p-3 max-h-80 overflow-y-auto">
                  <h3 className="text-xs text-gray-400 mb-2 px-1 font-semibold">
                    프로젝트 ({projects.length})
                  </h3>
                  {projects.length === 0 ? (
                    <p className="text-sm text-gray-500 p-2">프로젝트가 없습니다.</p>
                  ) : (
                    projects.map((project) => (
                      <button
                        key={project.id}
                        onClick={() => {
                          setSelectedProject(project);
                          setShowProjectSelector(false);
                        }}
                        className={`w-full px-3 py-2 text-left text-sm rounded transition truncate ${
                          selectedProject?.id === project.id
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
          {canAccessSettings && (
            <button
              // onClick={() => setIsSettingsModalOpen(true)}
              className={`flex items-center gap-1 p-2 rounded-lg transition ${theme.colors.secondary} ${theme.colors.text} hover:bg-gray-100 font-semibold text-sm`}
              title="조직 설정 및 멤버 관리"
            >
              <Settings className="w-4 h-4" />
              설정
            </button>
          )}
          {selectedProject && (
            <button
              className={`flex items-center gap-2 p-1 rounded-lg transition ${
                canAccessSettings ? 'hover:bg-blue-100' : 'hover:bg-gray-100'
              }`}
              title="조직원"
            >
              <AvatarStack />
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
            <div className="flex items-center justify-center h-full">
              <div className="text-center">
                <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500 mx-auto mb-4"></div>
                <p className={`${theme.colors.text}`}>프로젝트를 로드 중...</p>
              </div>
            </div>
          ) : selectedProject ? (
            <div className="flex flex-col lg:flex-row gap-3 sm:gap-4 min-w-max pb-4">
              {columns.map((column, idx) => (
                <div
                  key={column.id}
                  onDragOver={handleDragOver}
                  onDrop={() => handleDrop(column.id)}
                  className="w-full lg:w-80 lg:flex-shrink-0 relative"
                >
                  <div
                    className={`relative ${theme.effects.cardBorderWidth} ${theme.colors.border} p-3 sm:p-4 ${theme.colors.card} ${theme.effects.borderRadius}`}
                  >
                    <div className={`flex items-center justify-between pb-2`}>
                      <h3
                        className={`font-bold ${theme.colors.text} flex items-center gap-2 ${theme.font.size.xs}`}
                      >
                        <span
                          className={`w-3 h-3 sm:w-4 sm:h-4 ${
                            columnColors[idx % columnColors.length]
                          } ${theme.effects.cardBorderWidth} ${theme.colors.border}`}
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
                        <div key={board.id} className="relative">
                          <div
                            draggable
                            onDragStart={() => handleDragStart(board, column.id)}
                            onClick={() => setSelectedBoard(board)}
                            className={`relative ${theme.colors.card} p-3 sm:p-4 ${theme.effects.cardBorderWidth} ${theme.colors.border} hover:border-blue-500 transition cursor-pointer ${theme.effects.borderRadius}`}
                          >
                            <h3
                              className={`font-bold ${theme.colors.text} mb-2 sm:mb-3 ${theme.font.size.xs} break-words`}
                            >
                              {board.title}
                            </h3>
                            <div className="flex items-center justify-between">
                              <AssigneeAvatarStack assignees={board.assignee} />
                            </div>
                          </div>
                        </div>
                      ))}
                      <button
                        className={`relative w-full py-3 sm:py-4 ${theme.effects.cardBorderWidth} border-dashed ${theme.colors.border} ${theme.colors.card} hover:bg-gray-100 transition flex items-center justify-center gap-2 ${theme.font.size.xs} ${theme.effects.borderRadius}`}
                        onClick={() => {
                          setCreateBoardStageId('');
                          setShowCreateBoard(true);
                        }}
                      >
                        <Plus className="w-3 h-3 sm:w-4 sm:h-4" style={{ strokeWidth: 3 }} />
                        보드 추가
                      </button>
                    </div>
                  </div>
                </div>
              ))}
            </div>
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
                className={`w-10 h-10 ${theme.colors.primary} flex items-center justify-center text-white text-base font-bold rounded-md`}
              >
                {userProfile.avatar}
              </div>
              <div>
                <h3 className="font-bold text-lg text-gray-900">{userProfile.name}</h3>
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
        <CreateProjectModal
          workspaceId={currentWorkspaceId}
          onClose={() => setShowCreateProject(false)}
          onProjectCreated={fetchProjects}
        />
      )}

      {showCreateBoard && selectedProject && (
        <CreateBoardModal
          projectId={selectedProject.id}
          stageId={createBoardStageId}
          onClose={() => setShowCreateBoard(false)}
          onBoardCreated={fetchBoards}
        />
      )}

      {selectedBoard && (
        <BoardDetailModal board={selectedBoard} onClose={() => setSelectedBoard(null)} />
      )}
    </div>
  );
};

export default MainDashboard;
