-- ============================================
-- Board Service Baseline Schema v3.0.0 - Rollback
-- ============================================

-- Drop all tables in reverse order
DROP TABLE IF EXISTS comments CASCADE;
DROP TABLE IF EXISTS user_board_order CASCADE;
DROP TABLE IF EXISTS saved_views CASCADE;
DROP TABLE IF EXISTS boards CASCADE;
DROP TABLE IF EXISTS board_field_values CASCADE;
DROP TABLE IF EXISTS field_options CASCADE;
DROP TABLE IF EXISTS project_fields CASCADE;
DROP TABLE IF EXISTS project_join_requests CASCADE;
DROP TABLE IF EXISTS project_members CASCADE;
DROP TABLE IF EXISTS projects CASCADE;
DROP TABLE IF EXISTS roles CASCADE;
DROP TABLE IF EXISTS schema_versions CASCADE;
