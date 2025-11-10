-- Rollback: Convert back to integer display_order

-- ==================== 1. Add display_order column ====================

ALTER TABLE user_board_order ADD COLUMN display_order INT;

-- Convert positions back to integers using ROW_NUMBER
WITH ordered_boards AS (
    SELECT
        id,
        ROW_NUMBER() OVER (PARTITION BY view_id, user_id ORDER BY position) - 1 AS new_order
    FROM user_board_order
)
UPDATE user_board_order
SET display_order = ordered_boards.new_order
FROM ordered_boards
WHERE user_board_order.id = ordered_boards.id;

-- Make display_order NOT NULL
ALTER TABLE user_board_order ALTER COLUMN display_order SET NOT NULL;

-- Recreate old index
CREATE INDEX idx_user_board_order_view_user ON user_board_order(view_id, user_id, display_order);

-- Drop position column and its index
DROP INDEX IF EXISTS idx_user_board_order_position;
ALTER TABLE user_board_order DROP COLUMN position;

-- ==================== 2. Remove schema version ====================

DELETE FROM schema_versions WHERE version = '20250111120000';
