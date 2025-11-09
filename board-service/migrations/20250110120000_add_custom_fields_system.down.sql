-- Rollback: Remove Custom Fields System
-- Date: 2025-01-10

-- Drop tables in reverse order (to handle dependencies)
DROP TABLE IF EXISTS user_board_order CASCADE;
DROP TABLE IF EXISTS saved_views CASCADE;
DROP TABLE IF EXISTS board_field_values CASCADE;
DROP TABLE IF EXISTS field_options CASCADE;
DROP TABLE IF EXISTS project_fields CASCADE;

-- Remove JSONB cache column from boards
ALTER TABLE boards DROP COLUMN IF EXISTS custom_fields_cache;
DROP INDEX IF EXISTS idx_boards_custom_fields;

-- Remove schema version
DELETE FROM schema_versions WHERE version = '20250110120000';
