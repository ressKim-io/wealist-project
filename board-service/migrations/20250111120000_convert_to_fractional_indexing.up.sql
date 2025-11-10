-- Migration: Convert display_order to fractional indexing
-- Date: 2025-01-11
-- Description:
--   - Replace integer display_order with string position (fractional indexing)
--   - Provides O(1) insert/move operations without reordering other rows
--   - Eliminates DB bottleneck when inserting boards in the middle of a list

-- ==================== 1. Add position column to user_board_order ====================

ALTER TABLE user_board_order ADD COLUMN position VARCHAR(255);

-- Generate initial positions based on existing display_order
-- Convert integer order to fractional positions: 0 → "a0", 1 → "a1", 2 → "a2", etc.
UPDATE user_board_order
SET position = 'a' || display_order::TEXT
WHERE position IS NULL;

-- Make position NOT NULL after data migration
ALTER TABLE user_board_order ALTER COLUMN position SET NOT NULL;

-- Create index on position for efficient sorting
CREATE INDEX idx_user_board_order_position ON user_board_order(view_id, user_id, position);

-- Drop old display_order column
ALTER TABLE user_board_order DROP COLUMN display_order;

-- Update unique constraint to use position instead of display_order
-- (No unique constraint on position needed - positions can be arbitrary between boards)

-- ==================== 2. Update schema version ====================

INSERT INTO schema_versions (version, description)
VALUES ('20250111120000', 'Convert to fractional indexing for board ordering (eliminates reordering bottleneck)');

-- ==================== 3. Comments ====================

COMMENT ON COLUMN user_board_order.position IS 'Fractional index position string (lexicographically sortable, e.g., "a0", "a0V", "a1")';

-- ==================== 4. Performance Notes ====================

-- Fractional Indexing Benefits:
-- 1. O(1) insert/move: Only 1 row updated (the moved board)
-- 2. No cascading updates: Other boards remain untouched
-- 3. Lexicographic sorting: Fast ORDER BY position
-- 4. Infinite precision: Can always insert between any two positions
--
-- Example:
--   Before: board-a (pos: "a0"), board-b (pos: "a1")
--   Insert board-c between them: board-c (pos: "a0V") - only 1 INSERT!
--   No need to update board-b or any other boards
