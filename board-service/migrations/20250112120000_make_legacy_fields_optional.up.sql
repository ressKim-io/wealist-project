-- Migration: Make legacy fields optional (Stage, Role, Importance)
-- Date: 2025-01-12
-- Description:
--   - Make custom_stage_id nullable to support new custom fields system
--   - Boards can now be created without legacy fields
--   - Use custom fields (board_field_values) instead of fixed fields

-- ==================== 1. Make custom_stage_id nullable ====================

ALTER TABLE boards ALTER COLUMN custom_stage_id DROP NOT NULL;

-- ==================== 2. Add comments ====================

COMMENT ON COLUMN boards.custom_stage_id IS 'DEPRECATED: Legacy stage field (nullable). Use custom fields instead via board_field_values';
COMMENT ON COLUMN boards.custom_importance_id IS 'DEPRECATED: Legacy importance field (nullable). Use custom fields instead via board_field_values';

-- ==================== 3. Update schema version ====================

INSERT INTO schema_versions (version, description)
VALUES ('20250112120000', 'Make legacy fields optional (custom_stage_id nullable for new custom fields system)');

-- ==================== 4. Migration Notes ====================

-- This migration enables gradual transition from legacy fixed fields to flexible custom fields:
--
-- OLD SYSTEM (Legacy):
--   - Fixed fields: Stage (required), Role (many-to-many), Importance (optional)
--   - Stored in: boards.custom_stage_id, board_roles table, boards.custom_importance_id
--   - API: CreateBoard requires stage_id and role_ids
--
-- NEW SYSTEM (Jira-style):
--   - Flexible custom fields: Any number of fields with various types
--   - Stored in: board_field_values (EAV pattern) + boards.custom_fields_cache (JSONB)
--   - API: CreateBoard then SetFieldValue, or use custom fields directly
--
-- Migration Path:
--   1. This migration makes stage optional
--   2. Update application code to handle nullable stage
--   3. Create boards using new custom fields system
--   4. Eventually migrate existing boards and drop legacy columns
