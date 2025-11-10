-- Rollback: Make legacy fields optional
-- Date: 2025-01-12

-- This rollback requires all boards to have a custom_stage_id before running
-- If any boards exist without custom_stage_id, this will fail

-- Remove deprecation comments
COMMENT ON COLUMN boards.custom_stage_id IS 'References custom_stages.id (no FK for sharding)';
COMMENT ON COLUMN boards.custom_importance_id IS 'References custom_importance.id (no FK for sharding)';

-- Make custom_stage_id required again
ALTER TABLE boards ALTER COLUMN custom_stage_id SET NOT NULL;

-- Remove schema version
DELETE FROM schema_versions WHERE version = '20250112120000';
