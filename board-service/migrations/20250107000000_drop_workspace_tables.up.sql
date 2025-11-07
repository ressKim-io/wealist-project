-- ============================================
-- Drop Workspace Tables Migration
-- Created: 2025-01-07
-- Description: Remove workspace management from Board Service
--              Workspace functionality moved to User Service
-- ============================================

-- Drop workspace-related tables (preserve roles table)
DROP TABLE IF EXISTS workspace_join_requests CASCADE;
DROP TABLE IF EXISTS workspace_members CASCADE;
DROP TABLE IF EXISTS workspaces CASCADE;

-- Update schema version
INSERT INTO schema_versions (version, description)
VALUES ('20250107000000', 'Drop workspace tables - moved to User Service');
