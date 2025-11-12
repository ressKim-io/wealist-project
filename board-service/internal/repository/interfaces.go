package repository

// ==================== Repository Interfaces ====================
// 이 파일은 모든 Repository 인터페이스의 중앙 문서 역할을 합니다.
// Interface Segregation Principle (ISP)를 따라 각 Repository는
// 명확한 단일 책임을 가지며, 작고 집중된 인터페이스를 제공합니다.
//
// Best Practices:
// 1. 각 Repository는 하나의 Aggregate만 관리 (DDD 원칙)
// 2. 서비스는 인터페이스에 의존, 구현체에 의존하지 않음
// 3. 테스트 시 Mock 구현체로 쉽게 교체 가능
// 4. Generic BaseRepository를 통한 공통 CRUD 중복 제거
//
// Repository 목록:
// - BoardRepository       : Board 엔티티 관리
// - ProjectRepository     : Project 엔티티 관리
// - CommentRepository     : Comment 엔티티 관리
// - RoleRepository        : Role 엔티티 관리
// - FieldRepository       : Custom Fields 통합 관리 (Adapter 패턴)
// - ProjectFieldRepository: ProjectField 엔티티 관리
// - FieldOptionRepository : FieldOption 엔티티 관리
// - FieldValueRepository  : BoardFieldValue 엔티티 관리
// - ViewRepository        : SavedView 엔티티 관리
// - BoardOrderRepository  : UserBoardOrder 엔티티 관리
//
// 각 인터페이스의 상세 정의는 해당 파일을 참조하세요:
// - board_repository.go
// - project_repository.go
// - comment_repository.go
// - role_repository.go
// - field_repository.go (Adapter - 호환성 유지)
// - project_field_repository.go
// - field_option_repository.go
// - field_value_repository.go
// - view_repository.go
// - board_order_repository.go
//
// ==================== 사용 예시 ====================
//
// Service에서 Repository 주입:
//
//	type boardService struct {
//	    boardRepo   repository.BoardRepository
//	    projectRepo repository.ProjectRepository
//	    // ...
//	}
//
//	func NewBoardService(
//	    boardRepo repository.BoardRepository,
//	    projectRepo repository.ProjectRepository,
//	) BoardService {
//	    return &boardService{
//	        boardRepo:   boardRepo,
//	        projectRepo: projectRepo,
//	    }
//	}
//
// 테스트에서 Mock 사용:
//
//	type MockBoardRepository struct {
//	    mock.Mock
//	}
//
//	func (m *MockBoardRepository) FindByID(id uuid.UUID) (*domain.Board, error) {
//	    args := m.Called(id)
//	    return args.Get(0).(*domain.Board), args.Error(1)
//	}
//
// ==================== Repository 패턴 아키텍처 ====================
//
// 1. Domain Layer (internal/domain/)
//    - 비즈니스 엔티티 정의
//    - Rich Domain Model (비즈니스 로직 캡슐화)
//
// 2. Repository Layer (internal/repository/)
//    - 데이터 접근 추상화
//    - Generic Base Repository (공통 CRUD)
//    - 도메인별 독립 Repository
//
// 3. Service Layer (internal/service/)
//    - 비즈니스 로직 조율
//    - Repository 인터페이스에 의존
//    - UnitOfWork를 통한 트랜잭션 관리
//
// 4. Handler Layer (internal/handler/)
//    - HTTP 요청/응답 처리
//    - DTO ↔ Domain 변환
//
// ==================== 의존성 흐름 ====================
//
// Handler → Service (interface) → Repository (interface) → Domain
//                ↑                         ↑
//          (dependency)              (dependency)
//
// 의존성 역전 원칙 (DIP):
// - 상위 레벨 모듈(Service)은 하위 레벨 모듈(Repository)에 의존하지 않음
// - 둘 다 추상화(interface)에 의존
// - 구현체는 런타임에 주입 (Dependency Injection)
