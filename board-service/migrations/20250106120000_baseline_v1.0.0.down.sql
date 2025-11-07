-- Rollback baseline migration v2.0.0 (without workspace tables)

DROP TABLE IF EXISTS comments CASCADE;
DROP TABLE IF EXISTS user_board_order_in_stage CASCADE;
DROP TABLE IF EXISTS user_board_order_in_role CASCADE;
DROP TABLE IF EXISTS user_stage_column_order CASCADE;
DROP TABLE IF EXISTS user_role_column_order CASCADE;
DROP TABLE IF EXISTS board_roles CASCADE;
DROP TABLE IF EXISTS boards CASCADE;
DROP TABLE IF EXISTS custom_importance CASCADE;
DROP TABLE IF EXISTS custom_stages CASCADE;
DROP TABLE IF EXISTS custom_roles CASCADE;
DROP TABLE IF EXISTS project_join_requests CASCADE;
DROP TABLE IF EXISTS project_members CASCADE;
DROP TABLE IF EXISTS projects CASCADE;
-- Workspace tables removed - managed by User Service
DROP TABLE IF EXISTS roles CASCADE;
DROP TABLE IF EXISTS schema_versions CASCADE;
