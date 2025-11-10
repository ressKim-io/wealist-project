-- ============================================
-- Board Service Baseline Schema v3.0.0
-- Created: 2025-01-13
-- Description: Clean baseline with Custom Fields System (no legacy fields)
-- ============================================

-- Schema Versions Table
CREATE TABLE IF NOT EXISTS schema_versions (
    version VARCHAR(50) PRIMARY KEY,
    description TEXT,
    applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ==================== System Tables ====================

-- Roles Table (System-wide default roles)
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

-- ==================== Project Management ====================

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

-- ==================== Custom Fields System (Jira-style) ====================

-- Project Fields (Field Definitions)
CREATE TABLE IF NOT EXISTS project_fields (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL,

    -- Basic information
    name VARCHAR(255) NOT NULL,
    field_type VARCHAR(50) NOT NULL,
    description TEXT,

    -- Display settings
    display_order INT NOT NULL DEFAULT 0,
    is_required BOOLEAN NOT NULL DEFAULT false,
    is_system_default BOOLEAN NOT NULL DEFAULT false,

    -- Type-specific configuration (JSON)
    config TEXT NOT NULL DEFAULT '{}',

    -- Permissions
    can_edit_roles TEXT,

    -- Metadata
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_deleted BOOLEAN NOT NULL DEFAULT false,

    CONSTRAINT uq_project_field_name UNIQUE(project_id, name, is_deleted)
);

CREATE INDEX idx_project_fields_project ON project_fields(project_id, display_order) WHERE is_deleted = false;
CREATE INDEX idx_project_fields_type ON project_fields(field_type) WHERE is_deleted = false;
CREATE INDEX idx_project_fields_system ON project_fields(project_id, is_system_default) WHERE is_deleted = false;

COMMENT ON TABLE project_fields IS 'Custom field definitions per project (Jira-style)';
COMMENT ON COLUMN project_fields.field_type IS 'text, number, single_select, multi_select, date, datetime, single_user, multi_user, checkbox, url';
COMMENT ON COLUMN project_fields.config IS 'JSON configuration for type-specific settings';

-- Field Options (for Select types)
CREATE TABLE IF NOT EXISTS field_options (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    field_id UUID NOT NULL,

    label VARCHAR(255) NOT NULL,
    color VARCHAR(7),
    description TEXT,
    display_order INT NOT NULL DEFAULT 0,

    -- Metadata
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_deleted BOOLEAN NOT NULL DEFAULT false,

    CONSTRAINT uq_field_option_label UNIQUE(field_id, label, is_deleted)
);

CREATE INDEX idx_field_options_field ON field_options(field_id, display_order) WHERE is_deleted = false;
CREATE INDEX idx_field_options_deleted ON field_options(is_deleted);

COMMENT ON TABLE field_options IS 'Options for single_select and multi_select field types';
COMMENT ON COLUMN field_options.color IS 'HEX color code for UI display';

-- Board Field Values (EAV pattern)
CREATE TABLE IF NOT EXISTS board_field_values (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    board_id UUID NOT NULL,
    field_id UUID NOT NULL,

    -- Value columns (only one should be NOT NULL based on field type)
    value_text TEXT,
    value_number NUMERIC(15, 4),
    value_date TIMESTAMP,
    value_boolean BOOLEAN,
    value_option_id UUID,
    value_user_id UUID,

    -- Display order (for multi_select, multi_user)
    display_order INT DEFAULT 0,

    -- Metadata
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_deleted BOOLEAN NOT NULL DEFAULT false,

    CONSTRAINT uq_board_field_value UNIQUE(board_id, field_id, value_option_id, value_user_id, is_deleted)
);

-- Core indexes
CREATE INDEX idx_bfv_board ON board_field_values(board_id) WHERE is_deleted = false;
CREATE INDEX idx_bfv_field ON board_field_values(field_id) WHERE is_deleted = false;
CREATE INDEX idx_bfv_board_field ON board_field_values(board_id, field_id) WHERE is_deleted = false;

-- Type-specific indexes for filtering
CREATE INDEX idx_bfv_text ON board_field_values(field_id, value_text) WHERE value_text IS NOT NULL AND is_deleted = false;
CREATE INDEX idx_bfv_number ON board_field_values(field_id, value_number) WHERE value_number IS NOT NULL AND is_deleted = false;
CREATE INDEX idx_bfv_date ON board_field_values(field_id, value_date) WHERE value_date IS NOT NULL AND is_deleted = false;
CREATE INDEX idx_bfv_option ON board_field_values(field_id, value_option_id) WHERE value_option_id IS NOT NULL AND is_deleted = false;
CREATE INDEX idx_bfv_user ON board_field_values(field_id, value_user_id) WHERE value_user_id IS NOT NULL AND is_deleted = false;

COMMENT ON TABLE board_field_values IS 'EAV pattern for storing dynamic field values per board';

-- ==================== Boards (Kanban Cards) ====================

-- Boards Table
CREATE TABLE IF NOT EXISTS boards (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    assignee_id UUID,
    created_by UUID NOT NULL,
    due_date TIMESTAMP,

    -- JSONB cache for fast querying (denormalized from board_field_values)
    custom_fields_cache JSONB DEFAULT '{}'::jsonb,

    -- Fractional indexing for ordering
    position VARCHAR(255),

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_deleted BOOLEAN DEFAULT false
);

CREATE INDEX idx_boards_project_id ON boards(project_id);
CREATE INDEX idx_boards_assignee_id ON boards(assignee_id);
CREATE INDEX idx_boards_created_by ON boards(created_by);
CREATE INDEX idx_boards_due_date ON boards(due_date);
CREATE INDEX idx_boards_is_deleted ON boards(is_deleted);
CREATE INDEX idx_boards_position ON boards(project_id, position) WHERE is_deleted = false;

-- GIN index for JSONB queries
CREATE INDEX idx_boards_custom_fields_gin ON boards USING GIN(custom_fields_cache)
    WHERE custom_fields_cache IS NOT NULL AND custom_fields_cache != '{}'::jsonb;

COMMENT ON TABLE boards IS 'Kanban cards/boards within projects';
COMMENT ON COLUMN boards.custom_fields_cache IS 'JSONB cache of all field values for fast filtering (updated on field value changes)';
COMMENT ON COLUMN boards.position IS 'Fractional index for O(1) reordering (lexicographic sort)';

-- ==================== Saved Views ====================

-- Saved Views (Filters + Grouping)
CREATE TABLE IF NOT EXISTS saved_views (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL,
    created_by UUID NOT NULL,

    name VARCHAR(255) NOT NULL,
    description TEXT,
    is_default BOOLEAN DEFAULT false,
    is_shared BOOLEAN DEFAULT false,

    -- Filter configuration (JSON)
    filters TEXT DEFAULT '{}',

    -- Sort configuration
    sort_by VARCHAR(255),
    sort_direction VARCHAR(4) DEFAULT 'asc',

    -- Group by
    group_by_field_id UUID,

    -- Metadata
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_deleted BOOLEAN NOT NULL DEFAULT false
);

CREATE INDEX idx_saved_views_project ON saved_views(project_id) WHERE is_deleted = false;
CREATE INDEX idx_saved_views_creator ON saved_views(created_by) WHERE is_deleted = false;
CREATE INDEX idx_saved_views_default ON saved_views(project_id, is_default) WHERE is_deleted = false AND is_default = true;

COMMENT ON TABLE saved_views IS 'User-defined views with filters, sorting, and grouping';
COMMENT ON COLUMN saved_views.filters IS 'JSON filter conditions';

-- User Board Order (Fractional Indexing)
CREATE TABLE IF NOT EXISTS user_board_order (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    view_id UUID NOT NULL,
    user_id UUID NOT NULL,
    board_id UUID NOT NULL,

    position VARCHAR(255) NOT NULL,

    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT uq_user_board_order UNIQUE(view_id, user_id, board_id)
);

CREATE INDEX idx_user_board_order_view_user ON user_board_order(view_id, user_id, position);
CREATE INDEX idx_user_board_order_board ON user_board_order(board_id);

COMMENT ON TABLE user_board_order IS 'User-specific manual board ordering within views using fractional indexing';
COMMENT ON COLUMN user_board_order.position IS 'Fractional index for O(1) reordering';

-- ==================== Comments ====================

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
COMMENT ON COLUMN comments.deleted_at IS 'Soft delete timestamp (GORM DeletedAt)';

-- ==================== Default Data ====================

-- Default Roles
INSERT INTO roles (name, level, description) VALUES
    ('OWNER', 100, 'Workspace owner with full permissions'),
    ('ADMIN', 50, 'Administrator with management permissions'),
    ('MEMBER', 10, 'Regular member with basic permissions')
ON CONFLICT (name) DO NOTHING;

-- ==================== Schema Version ====================

INSERT INTO schema_versions (version, description)
VALUES ('20250113120000', 'Baseline v3.0.0 - Clean baseline with Custom Fields System (no legacy fields)');
