-- Migration: Add Custom Fields System (Jira-style)
-- Date: 2025-01-10
-- Description:
--   - Replaces fixed Role/Stage/Importance with flexible custom field system
--   - Supports: Text, Number, Select, Date, User, Checkbox, URL types
--   - Multi-select fields enable multi-dimensional classification
--   - Saved views for filtering and grouping

-- ==================== 1. Project Fields (Field Definitions) ====================

CREATE TABLE IF NOT EXISTS project_fields (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL, -- References projects.id (no FK for sharding)

    -- Basic information
    name VARCHAR(255) NOT NULL,
    field_type VARCHAR(50) NOT NULL, -- 'text', 'number', 'single_select', 'multi_select', 'date', 'datetime', 'single_user', 'multi_user', 'checkbox', 'url'
    description TEXT,

    -- Display settings
    display_order INT NOT NULL DEFAULT 0,
    is_required BOOLEAN NOT NULL DEFAULT false,
    is_system_default BOOLEAN NOT NULL DEFAULT false, -- System fields (e.g., migrated from Stage)

    -- Type-specific configuration (JSON)
    config TEXT NOT NULL DEFAULT '{}', -- Stored as JSON string
    /*
    Configuration examples by type:
    {
      "text": {"max_length": 500, "is_long": false},
      "number": {"min": 0, "max": 100, "decimal_places": 2},
      "single_select": {},
      "multi_select": {"max_selections": 5},
      "date": {"include_time": false},
      "datetime": {"include_time": true},
      "single_user": {"project_members_only": true},
      "multi_user": {"max_users": 3, "project_members_only": true},
      "checkbox": {"default_value": false},
      "url": {"enable_preview": true}
    }
    */

    -- Permissions (simplified)
    can_edit_roles TEXT, -- Comma-separated roles: 'ADMIN,OWNER' (NULL = all can edit)

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
COMMENT ON COLUMN project_fields.project_id IS 'References projects.id (no FK for sharding)';
COMMENT ON COLUMN project_fields.field_type IS 'Field data type: text, number, single_select, multi_select, date, datetime, single_user, multi_user, checkbox, url';
COMMENT ON COLUMN project_fields.config IS 'JSON configuration for type-specific settings';

-- ==================== 2. Field Options (for Select types) ====================

CREATE TABLE IF NOT EXISTS field_options (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    field_id UUID NOT NULL, -- References project_fields.id (no FK)

    label VARCHAR(255) NOT NULL,
    color VARCHAR(7), -- HEX color: #RRGGBB
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
COMMENT ON COLUMN field_options.field_id IS 'References project_fields.id (no FK)';
COMMENT ON COLUMN field_options.color IS 'HEX color code for UI display';

-- ==================== 3. Board Field Values (EAV pattern) ====================

CREATE TABLE IF NOT EXISTS board_field_values (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    board_id UUID NOT NULL, -- References boards.id (no FK)
    field_id UUID NOT NULL, -- References project_fields.id (no FK)

    -- Value columns (only one should be NOT NULL based on field type)
    value_text TEXT,
    value_number NUMERIC(15, 4),
    value_date TIMESTAMP,
    value_boolean BOOLEAN,
    value_option_id UUID, -- References field_options.id (no FK)
    value_user_id UUID,   -- References users.id (no FK)

    -- Display order (for multi_select, multi_user)
    display_order INT DEFAULT 0,

    -- Metadata
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_deleted BOOLEAN NOT NULL DEFAULT false,

    -- Unique constraint: prevents duplicate values for multi-select/multi-user
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
COMMENT ON COLUMN board_field_values.board_id IS 'References boards.id (no FK)';
COMMENT ON COLUMN board_field_values.field_id IS 'References project_fields.id (no FK)';
COMMENT ON COLUMN board_field_values.display_order IS 'Order for multi-select and multi-user values';

-- ==================== 4. Boards: Add JSONB cache column ====================

-- Add JSONB cache column to boards table for fast querying
ALTER TABLE boards ADD COLUMN IF NOT EXISTS custom_fields_cache TEXT DEFAULT '{}';

-- GIN index for JSONB queries (PostgreSQL optimizes JSON queries)
-- Note: Using TEXT column to store JSON (GORM handles serialization)
CREATE INDEX IF NOT EXISTS idx_boards_custom_fields ON boards(custom_fields_cache) WHERE custom_fields_cache != '{}';

COMMENT ON COLUMN boards.custom_fields_cache IS 'JSON cache of all field values for fast filtering (updated on field value changes)';

-- ==================== 5. Saved Views (Filters + Grouping) ====================

CREATE TABLE IF NOT EXISTS saved_views (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL, -- References projects.id (no FK)
    created_by UUID NOT NULL, -- References users.id (no FK)

    name VARCHAR(255) NOT NULL,
    description TEXT,
    is_default BOOLEAN DEFAULT false,
    is_shared BOOLEAN DEFAULT false, -- Shared with all project members

    -- Filter configuration (JSON)
    filters TEXT DEFAULT '{}',
    /*
    Filter examples:
    {
      "field_uuid_1": {"operator": "contains", "value": "search term"},
      "field_uuid_2": {"operator": "in", "value": ["option_id_1", "option_id_2"]},
      "field_uuid_3": {"operator": ">=", "value": 10},
      "title": {"operator": "contains", "value": "bug"}
    }
    */

    -- Sort configuration
    sort_by VARCHAR(255), -- field_id or special: 'created_at', 'updated_at', 'title'
    sort_direction VARCHAR(4) DEFAULT 'asc', -- 'asc' or 'desc'

    -- Group by (only multi_select fields allowed)
    group_by_field_id UUID, -- References project_fields.id (no FK)

    -- Metadata
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_deleted BOOLEAN NOT NULL DEFAULT false
);

CREATE INDEX idx_saved_views_project ON saved_views(project_id) WHERE is_deleted = false;
CREATE INDEX idx_saved_views_creator ON saved_views(created_by) WHERE is_deleted = false;
CREATE INDEX idx_saved_views_default ON saved_views(project_id, is_default) WHERE is_deleted = false AND is_default = true;

COMMENT ON TABLE saved_views IS 'User-defined views with filters, sorting, and grouping';
COMMENT ON COLUMN saved_views.project_id IS 'References projects.id (no FK)';
COMMENT ON COLUMN saved_views.created_by IS 'References users.id (no FK)';
COMMENT ON COLUMN saved_views.filters IS 'JSON filter conditions';
COMMENT ON COLUMN saved_views.group_by_field_id IS 'Multi-select field to group boards by';

-- ==================== 6. User Board Order (Manual Sorting) ====================

CREATE TABLE IF NOT EXISTS user_board_order (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    view_id UUID NOT NULL, -- References saved_views.id (no FK)
    user_id UUID NOT NULL, -- References users.id (no FK)
    board_id UUID NOT NULL, -- References boards.id (no FK)

    display_order INT NOT NULL,

    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT uq_user_board_order UNIQUE(view_id, user_id, board_id)
);

CREATE INDEX idx_user_board_order_view_user ON user_board_order(view_id, user_id, display_order);
CREATE INDEX idx_user_board_order_board ON user_board_order(board_id);

COMMENT ON TABLE user_board_order IS 'User-specific manual board ordering within views';
COMMENT ON COLUMN user_board_order.view_id IS 'References saved_views.id (no FK)';
COMMENT ON COLUMN user_board_order.user_id IS 'References users.id (no FK)';
COMMENT ON COLUMN user_board_order.board_id IS 'References boards.id (no FK)';

-- ==================== 7. Update Schema Version ====================

INSERT INTO schema_versions (version, description)
VALUES ('20250110120000', 'Add custom fields system (Jira-style: Text, Number, Select, Date, User, Checkbox, URL)');
