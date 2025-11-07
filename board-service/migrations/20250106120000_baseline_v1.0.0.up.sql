-- ============================================
-- Board Service Baseline Schema v2.0.0
-- Created: 2025-01-06
-- Updated: 2025-01-07 (Removed workspace tables - moved to User Service)
-- Description: Board schema without workspace management
-- ============================================

-- Schema Versions Table
CREATE TABLE IF NOT EXISTS schema_versions (
    version VARCHAR(50) PRIMARY KEY,
    description TEXT,
    applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Roles Table
CREATE TABLE IF NOT EXISTS roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) NOT NULL,
    level INT NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_deleted BOOLEAN DEFAULT false,
    CONSTRAINT uni_roles_name UNIQUE(name)
);
CREATE INDEX idx_roles_name ON roles(name);
CREATE INDEX idx_roles_level ON roles(level);
CREATE INDEX idx_roles_is_deleted ON roles(is_deleted);

COMMENT ON TABLE roles IS 'System-wide default roles (OWNER, ADMIN, MEMBER)';
COMMENT ON COLUMN roles.level IS 'Permission level: higher = more permissions';

-- NOTE: Workspace tables removed - managed by User Service
-- Workspace validation is done via User Service API calls

-- Projects Table
CREATE TABLE IF NOT EXISTS projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    owner_id UUID NOT NULL,
    is_public BOOLEAN DEFAULT false,
    is_deleted BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_projects_workspace_id ON projects(workspace_id);
CREATE INDEX idx_projects_owner_id ON projects(owner_id);
CREATE INDEX idx_projects_is_deleted ON projects(is_deleted);

COMMENT ON COLUMN projects.workspace_id IS 'References User Service workspace ID (validated via API)';
COMMENT ON COLUMN projects.owner_id IS 'References users.id (no FK for microservice isolation)';

-- Project Members Table
CREATE TABLE IF NOT EXISTS project_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL,
    user_id UUID NOT NULL,
    role_id UUID NOT NULL,
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_deleted BOOLEAN DEFAULT false,
    CONSTRAINT uni_project_members_project_id_user_id UNIQUE(project_id, user_id)
);
CREATE INDEX idx_project_members_project_id ON project_members(project_id);
CREATE INDEX idx_project_members_user_id ON project_members(user_id);
CREATE INDEX idx_project_members_role_id ON project_members(role_id);
CREATE INDEX idx_project_members_is_deleted ON project_members(is_deleted);

COMMENT ON COLUMN project_members.project_id IS 'References projects.id (no FK for sharding)';
COMMENT ON COLUMN project_members.user_id IS 'References users.id (no FK for microservice isolation)';
COMMENT ON COLUMN project_members.role_id IS 'References roles.id (no FK for sharding)';

-- Project Join Requests Table
CREATE TABLE IF NOT EXISTS project_join_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL,
    user_id UUID NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    requested_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_deleted BOOLEAN DEFAULT false,
    CONSTRAINT uni_project_join_requests_project_id_user_id UNIQUE(project_id, user_id)
);
CREATE INDEX idx_project_join_requests_project_id ON project_join_requests(project_id);
CREATE INDEX idx_project_join_requests_user_id ON project_join_requests(user_id);
CREATE INDEX idx_project_join_requests_status ON project_join_requests(status);
CREATE INDEX idx_project_join_requests_is_deleted ON project_join_requests(is_deleted);

COMMENT ON COLUMN project_join_requests.status IS 'PENDING, APPROVED, or REJECTED';

-- Custom Roles Table
CREATE TABLE IF NOT EXISTS custom_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL,
    name VARCHAR(50) NOT NULL,
    color VARCHAR(7),
    is_system_default BOOLEAN NOT NULL DEFAULT false,
    display_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_deleted BOOLEAN DEFAULT false,
    CONSTRAINT uni_custom_roles_project_id_name UNIQUE(project_id, name)
);
CREATE INDEX idx_custom_roles_project_id ON custom_roles(project_id);
CREATE INDEX idx_custom_roles_is_system_default ON custom_roles(is_system_default);
CREATE INDEX idx_custom_roles_is_deleted ON custom_roles(is_deleted);

COMMENT ON TABLE custom_roles IS 'Project-specific custom roles for board categorization';
COMMENT ON COLUMN custom_roles.project_id IS 'References projects.id (no FK for sharding)';
COMMENT ON COLUMN custom_roles.is_system_default IS 'Whether this is a system-created default role';

-- Custom Stages Table
CREATE TABLE IF NOT EXISTS custom_stages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL,
    name VARCHAR(50) NOT NULL,
    color VARCHAR(7),
    is_system_default BOOLEAN NOT NULL DEFAULT false,
    display_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_deleted BOOLEAN DEFAULT false,
    CONSTRAINT uni_custom_stages_project_id_name UNIQUE(project_id, name)
);
CREATE INDEX idx_custom_stages_project_id ON custom_stages(project_id);
CREATE INDEX idx_custom_stages_is_system_default ON custom_stages(is_system_default);
CREATE INDEX idx_custom_stages_is_deleted ON custom_stages(is_deleted);

COMMENT ON TABLE custom_stages IS 'Project-specific workflow stages (e.g., To Do, In Progress, Done)';
COMMENT ON COLUMN custom_stages.project_id IS 'References projects.id (no FK for sharding)';

-- Custom Importance Table
CREATE TABLE IF NOT EXISTS custom_importance (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL,
    name VARCHAR(50) NOT NULL,
    color VARCHAR(7),
    is_system_default BOOLEAN NOT NULL DEFAULT false,
    display_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_deleted BOOLEAN DEFAULT false,
    CONSTRAINT uni_custom_importance_project_id_name UNIQUE(project_id, name)
);
CREATE INDEX idx_custom_importance_project_id ON custom_importance(project_id);
CREATE INDEX idx_custom_importance_is_system_default ON custom_importance(is_system_default);
CREATE INDEX idx_custom_importance_is_deleted ON custom_importance(is_deleted);

COMMENT ON TABLE custom_importance IS 'Project-specific importance levels (e.g., Low, Medium, High, Critical)';
COMMENT ON COLUMN custom_importance.project_id IS 'References projects.id (no FK for sharding)';

-- Boards Table (Kanban Cards)
CREATE TABLE IF NOT EXISTS boards (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    custom_stage_id UUID NOT NULL,
    custom_importance_id UUID,
    assignee_id UUID,
    created_by UUID NOT NULL,
    due_date TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_deleted BOOLEAN DEFAULT false
);
CREATE INDEX idx_boards_project_id ON boards(project_id);
CREATE INDEX idx_boards_custom_stage_id ON boards(custom_stage_id);
CREATE INDEX idx_boards_custom_importance_id ON boards(custom_importance_id);
CREATE INDEX idx_boards_assignee_id ON boards(assignee_id);
CREATE INDEX idx_boards_created_by ON boards(created_by);
CREATE INDEX idx_boards_due_date ON boards(due_date);
CREATE INDEX idx_boards_is_deleted ON boards(is_deleted);

COMMENT ON TABLE boards IS 'Kanban cards/boards within projects';
COMMENT ON COLUMN boards.project_id IS 'References projects.id (no FK for sharding)';
COMMENT ON COLUMN boards.custom_stage_id IS 'References custom_stages.id (no FK for sharding)';
COMMENT ON COLUMN boards.custom_importance_id IS 'References custom_importance.id (no FK for sharding)';
COMMENT ON COLUMN boards.assignee_id IS 'References users.id (no FK for microservice isolation)';
COMMENT ON COLUMN boards.created_by IS 'References users.id (no FK for microservice isolation)';

-- Board Roles Table (Many-to-Many)
CREATE TABLE IF NOT EXISTS board_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    board_id UUID NOT NULL,
    custom_role_id UUID NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_deleted BOOLEAN DEFAULT false,
    CONSTRAINT uni_board_roles_board_id_custom_role_id UNIQUE(board_id, custom_role_id)
);
CREATE INDEX idx_board_roles_board_id ON board_roles(board_id);
CREATE INDEX idx_board_roles_custom_role_id ON board_roles(custom_role_id);
CREATE INDEX idx_board_roles_is_deleted ON board_roles(is_deleted);

COMMENT ON TABLE board_roles IS 'Many-to-many relationship between boards and custom roles';
COMMENT ON COLUMN board_roles.board_id IS 'References boards.id (no FK for sharding)';
COMMENT ON COLUMN board_roles.custom_role_id IS 'References custom_roles.id (no FK for sharding)';

-- User Role Column Order (Drag-and-Drop State)
CREATE TABLE IF NOT EXISTS user_role_column_order (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    project_id UUID NOT NULL,
    custom_role_id UUID NOT NULL,
    display_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_deleted BOOLEAN DEFAULT false,
    CONSTRAINT idx_user_role_unique UNIQUE(user_id, project_id, custom_role_id)
);
CREATE INDEX idx_user_role_order ON user_role_column_order(user_id, project_id);
CREATE INDEX idx_user_role_column_order_is_deleted ON user_role_column_order(is_deleted);

COMMENT ON TABLE user_role_column_order IS 'User-specific display order for role-based board columns';
COMMENT ON COLUMN user_role_column_order.user_id IS 'References users.id (no FK for microservice isolation)';
COMMENT ON COLUMN user_role_column_order.project_id IS 'References projects.id (no FK for sharding)';
COMMENT ON COLUMN user_role_column_order.custom_role_id IS 'References custom_roles.id (no FK for sharding)';

-- User Stage Column Order (Drag-and-Drop State)
CREATE TABLE IF NOT EXISTS user_stage_column_order (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    project_id UUID NOT NULL,
    custom_stage_id UUID NOT NULL,
    display_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_deleted BOOLEAN DEFAULT false,
    CONSTRAINT idx_user_stage_unique UNIQUE(user_id, project_id, custom_stage_id)
);
CREATE INDEX idx_user_stage_order ON user_stage_column_order(user_id, project_id);
CREATE INDEX idx_user_stage_column_order_is_deleted ON user_stage_column_order(is_deleted);

COMMENT ON TABLE user_stage_column_order IS 'User-specific display order for stage-based board columns';
COMMENT ON COLUMN user_stage_column_order.user_id IS 'References users.id (no FK for microservice isolation)';
COMMENT ON COLUMN user_stage_column_order.project_id IS 'References projects.id (no FK for sharding)';
COMMENT ON COLUMN user_stage_column_order.custom_stage_id IS 'References custom_stages.id (no FK for sharding)';

-- User Board Order In Role (Drag-and-Drop State)
CREATE TABLE IF NOT EXISTS user_board_order_in_role (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    project_id UUID NOT NULL,
    custom_role_id UUID NOT NULL,
    board_id UUID NOT NULL,
    display_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_deleted BOOLEAN DEFAULT false,
    CONSTRAINT idx_user_board_role_unique UNIQUE(user_id, project_id, custom_role_id, board_id)
);
CREATE INDEX idx_user_board_role_order ON user_board_order_in_role(user_id, project_id, custom_role_id);
CREATE INDEX idx_user_board_order_in_role_is_deleted ON user_board_order_in_role(is_deleted);

COMMENT ON TABLE user_board_order_in_role IS 'User-specific display order for boards within role columns';
COMMENT ON COLUMN user_board_order_in_role.user_id IS 'References users.id (no FK for microservice isolation)';
COMMENT ON COLUMN user_board_order_in_role.project_id IS 'References projects.id (no FK for sharding)';
COMMENT ON COLUMN user_board_order_in_role.custom_role_id IS 'References custom_roles.id (no FK for sharding)';
COMMENT ON COLUMN user_board_order_in_role.board_id IS 'References boards.id (no FK for sharding)';

-- User Board Order In Stage (Drag-and-Drop State)
CREATE TABLE IF NOT EXISTS user_board_order_in_stage (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    project_id UUID NOT NULL,
    custom_stage_id UUID NOT NULL,
    board_id UUID NOT NULL,
    display_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_deleted BOOLEAN DEFAULT false,
    CONSTRAINT idx_user_board_stage_unique UNIQUE(user_id, project_id, custom_stage_id, board_id)
);
CREATE INDEX idx_user_board_stage_order ON user_board_order_in_stage(user_id, project_id, custom_stage_id);
CREATE INDEX idx_user_board_order_in_stage_is_deleted ON user_board_order_in_stage(is_deleted);

COMMENT ON TABLE user_board_order_in_stage IS 'User-specific display order for boards within stage columns';
COMMENT ON COLUMN user_board_order_in_stage.user_id IS 'References users.id (no FK for microservice isolation)';
COMMENT ON COLUMN user_board_order_in_stage.project_id IS 'References projects.id (no FK for sharding)';
COMMENT ON COLUMN user_board_order_in_stage.custom_stage_id IS 'References custom_stages.id (no FK for sharding)';
COMMENT ON COLUMN user_board_order_in_stage.board_id IS 'References boards.id (no FK for sharding)';

-- Comments Table
CREATE TABLE IF NOT EXISTS comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    content TEXT NOT NULL,
    user_id UUID NOT NULL,
    board_id UUID NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);
CREATE INDEX idx_comments_board_id ON comments(board_id);
CREATE INDEX idx_comments_user_id ON comments(user_id);
CREATE INDEX idx_comments_deleted_at ON comments(deleted_at);

COMMENT ON TABLE comments IS 'Comments on board cards (uses GORM soft delete via deleted_at)';
COMMENT ON COLUMN comments.user_id IS 'References users.id (no FK for microservice isolation)';
COMMENT ON COLUMN comments.board_id IS 'References boards.id (no FK for sharding)';
COMMENT ON COLUMN comments.deleted_at IS 'Soft delete timestamp (GORM DeletedAt)';

-- Default Roles
INSERT INTO roles (name, level, description) VALUES
    ('OWNER', 100, 'Workspace owner with full permissions'),
    ('ADMIN', 50, 'Administrator with management permissions'),
    ('MEMBER', 10, 'Regular member with basic permissions')
ON CONFLICT (name) DO NOTHING;

-- Schema Version
INSERT INTO schema_versions (version, description)
VALUES ('20250106120000', 'Baseline v2.0.0 - Board schema without workspace tables (managed by User Service)');
