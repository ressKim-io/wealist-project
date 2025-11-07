-- ============================================
-- Rollback: Restore Workspace Tables
-- Created: 2025-01-07
-- Description: Restore workspace tables if migration needs to be rolled back
-- ============================================

-- Workspaces Table
CREATE TABLE IF NOT EXISTS workspaces (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    owner_id UUID NOT NULL,
    is_public BOOLEAN DEFAULT false,
    is_deleted BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_workspaces_owner_id ON workspaces(owner_id);
CREATE INDEX idx_workspaces_is_deleted ON workspaces(is_deleted);
CREATE INDEX idx_workspaces_created_at ON workspaces(created_at DESC);

COMMENT ON COLUMN workspaces.owner_id IS 'References users.id (no FK for microservice isolation)';

-- Workspace Members Table
CREATE TABLE IF NOT EXISTS workspace_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL,
    user_id UUID NOT NULL,
    role_id UUID NOT NULL,
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    left_at TIMESTAMP,
    is_default BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_deleted BOOLEAN DEFAULT false,
    CONSTRAINT uni_workspace_members_workspace_id_user_id UNIQUE(workspace_id, user_id)
);
CREATE INDEX idx_workspace_members_workspace_id ON workspace_members(workspace_id);
CREATE INDEX idx_workspace_members_user_id ON workspace_members(user_id);
CREATE INDEX idx_workspace_members_role_id ON workspace_members(role_id);
CREATE INDEX idx_workspace_members_is_deleted ON workspace_members(is_deleted);

COMMENT ON COLUMN workspace_members.workspace_id IS 'References workspaces.id (no FK for sharding)';
COMMENT ON COLUMN workspace_members.user_id IS 'References users.id (no FK for microservice isolation)';
COMMENT ON COLUMN workspace_members.role_id IS 'References roles.id (no FK for sharding)';
COMMENT ON COLUMN workspace_members.left_at IS 'Timestamp when member left the workspace (NULL if still active)';
COMMENT ON COLUMN workspace_members.is_default IS 'Whether this is the user''s default workspace';

-- Workspace Join Requests Table
CREATE TABLE IF NOT EXISTS workspace_join_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL,
    user_id UUID NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    requested_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    processed_at TIMESTAMP,
    processed_by UUID,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_deleted BOOLEAN DEFAULT false,
    CONSTRAINT uni_workspace_join_requests_workspace_id_user_id UNIQUE(workspace_id, user_id)
);
CREATE INDEX idx_workspace_join_requests_workspace_id ON workspace_join_requests(workspace_id);
CREATE INDEX idx_workspace_join_requests_user_id ON workspace_join_requests(user_id);
CREATE INDEX idx_workspace_join_requests_status ON workspace_join_requests(status);
CREATE INDEX idx_workspace_join_requests_is_deleted ON workspace_join_requests(is_deleted);

COMMENT ON COLUMN workspace_join_requests.status IS 'PENDING, APPROVED, or REJECTED';
COMMENT ON COLUMN workspace_join_requests.processed_by IS 'User ID who approved/rejected (no FK)';

-- Remove schema version entry
DELETE FROM schema_versions WHERE version = '20250107000000';
