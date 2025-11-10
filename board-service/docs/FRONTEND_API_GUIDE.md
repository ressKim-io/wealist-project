# Frontend API Guide - Board Ordering with Fractional Indexing

This guide explains how to use the Board Service API for managing board ordering using Fractional Indexing.

## Table of Contents
- [Overview](#overview)
- [What is Fractional Indexing?](#what-is-fractional-indexing)
- [API Endpoints](#api-endpoints)
- [Frontend Implementation Guide](#frontend-implementation-guide)
- [Common Scenarios](#common-scenarios)
- [Migration from Old API](#migration-from-old-api)
- [Error Handling](#error-handling)

---

## Overview

The Board Service uses **Fractional Indexing** for efficient board ordering. This means:
- ✅ **O(1) operations**: Moving a board only updates 1 row, not N rows
- ✅ **No cascading updates**: Other boards remain untouched
- ✅ **Infinite precision**: Can always insert between any two positions
- ✅ **Simple API**: Only need `before_position` and `after_position`

---

## What is Fractional Indexing?

Fractional Indexing uses **lexicographically sortable strings** instead of integers for positioning.

### Examples:

```
Position Strings: "a0", "a0V", "a1", "a2", "a3"
                   ↑     ↑     ↑     ↑     ↑
                 first  between between ...
                        a0-a1   a1-a2
```

### Why This Works:

- **Lexicographic sorting**: Strings are sorted alphabetically by the database
- **Between any two positions**: Can always generate a position between two existing ones
- **No reordering needed**: Insert board-c between board-a and board-b without touching them

### Used By:

- Linear
- Jira
- Notion
- Figma

---

## API Endpoints

### 1. Move Board (Order Change)

**Endpoint**: `POST /api/v1/boards/{board_id}/move`

**Purpose**: Move a board within or between columns, updating its position and custom field value.

#### Request Body:

```json
{
  "view_id": "123e4567-e89b-12d3-a456-426614174000",
  "group_by_field_id": "223e4567-e89b-12d3-a456-426614174000",
  "new_field_value": "323e4567-e89b-12d3-a456-426614174000",
  "before_position": "a1",     // Optional: position of board BEFORE target
  "after_position": "a2"        // Optional: position of board AFTER target
}
```

#### Request Fields:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `view_id` | string (UUID) | Yes | The saved view ID (for user-specific ordering) |
| `group_by_field_id` | string (UUID) | Yes | The custom field used for grouping (e.g., "Status" field) |
| `new_field_value` | string (UUID) | Yes | The new value for the custom field (e.g., "In Progress" option ID) |
| `before_position` | string | No | Position string of the board that should come BEFORE the moved board |
| `after_position` | string | No | Position string of the board that should come AFTER the moved board |

#### Response:

```json
{
  "board_id": "423e4567-e89b-12d3-a456-426614174000",
  "new_field_value": "323e4567-e89b-12d3-a456-426614174000",
  "new_position": "a1V",
  "message": "Board moved successfully"
}
```

---

## Frontend Implementation Guide

### Data Structure

Each board should have:

```typescript
interface Board {
  id: string;              // Board UUID
  title: string;
  position: string;        // Fractional index position (e.g., "a0", "a1V")
  customFields: {
    [fieldId: string]: string;  // fieldId → optionId
  };
}

interface Column {
  fieldOptionId: string;   // The option ID for this column (e.g., "Todo", "In Progress")
  boards: Board[];         // Boards in this column, sorted by position
}
```

### Step 1: Display Boards

Sort boards by their `position` string (lexicographic order):

```typescript
// Fetch boards from API
const boards = await fetchBoards(viewId);

// Group by custom field value
const columns = groupByFieldValue(boards, groupByFieldId);

// Sort each column by position
columns.forEach(column => {
  column.boards.sort((a, b) => a.position.localeCompare(b.position));
});
```

### Step 2: Handle Drag-and-Drop

When user drags a board to a new position:

```typescript
function onBoardDrop(
  boardId: string,
  targetColumnId: string,
  targetIndex: number,
  targetColumn: Board[]
) {
  // Find before and after boards
  const beforeBoard = targetIndex > 0 ? targetColumn[targetIndex - 1] : null;
  const afterBoard = targetIndex < targetColumn.length ? targetColumn[targetIndex] : null;

  // Prepare API request
  const request = {
    view_id: currentViewId,
    group_by_field_id: groupByFieldId,
    new_field_value: targetColumnId,  // New column's option ID
    before_position: beforeBoard?.position,  // Can be undefined
    after_position: afterBoard?.position     // Can be undefined
  };

  // Call API
  const response = await moveBoard(boardId, request);

  // Update local state with new position
  updateBoardPosition(boardId, response.new_position);
}
```

---

## Common Scenarios

### Scenario 1: Move Within Same Column

**UI Action**: User drags "Board-2" from position 1 to position 2 (after "Board-3")

**Current Order**:
```
Column: "Todo"
├─ Board-1 (position: "a0")
├─ Board-2 (position: "a1")  ← Moving this
└─ Board-3 (position: "a2")
```

**API Request**:
```json
{
  "view_id": "view-123",
  "group_by_field_id": "status-field-id",
  "new_field_value": "todo-option-id",  // Same as current
  "before_position": "a2",               // Board-3's position
  "after_position": null                 // No board after
}
```

**Result**:
- Board-2 gets new position: `"a3"` (after "a2")
- Only Board-2 is updated in DB
- Board-1 and Board-3 remain unchanged

**New Order**:
```
Column: "Todo"
├─ Board-1 (position: "a0")  ← Unchanged
├─ Board-3 (position: "a2")  ← Unchanged
└─ Board-2 (position: "a3")  ← Updated!
```

---

### Scenario 2: Move to Different Column

**UI Action**: User drags "Board-2" from "Todo" to "In Progress" (first position)

**Current State**:
```
Column: "Todo"
├─ Board-1 (position: "a0")
├─ Board-2 (position: "a1")  ← Moving this
└─ Board-3 (position: "a2")

Column: "In Progress"
├─ Board-4 (position: "b0")
└─ Board-5 (position: "b1")
```

**API Request**:
```json
{
  "view_id": "view-123",
  "group_by_field_id": "status-field-id",
  "new_field_value": "in-progress-option-id",  // Different column
  "before_position": null,                      // No board before
  "after_position": "b0"                        // Board-4's position
}
```

**Result**:
- Board-2 custom field updated: `status-field-id → in-progress-option-id`
- Board-2 gets new position: `"aV"` (before "b0")
- Only Board-2 is updated in DB

**New State**:
```
Column: "Todo"
├─ Board-1 (position: "a0")  ← Unchanged
└─ Board-3 (position: "a2")  ← Unchanged

Column: "In Progress"
├─ Board-2 (position: "aV")  ← Updated! (custom field + position)
├─ Board-4 (position: "b0")  ← Unchanged
└─ Board-5 (position: "b1")  ← Unchanged
```

---

### Scenario 3: Insert Between Two Boards

**UI Action**: User drags "Board-2" between "Board-4" and "Board-5"

**Current State**:
```
Column: "In Progress"
├─ Board-4 (position: "a0")
├─ Board-5 (position: "a1")
```

**API Request**:
```json
{
  "view_id": "view-123",
  "group_by_field_id": "status-field-id",
  "new_field_value": "in-progress-option-id",
  "before_position": "a0",   // Board-4's position
  "after_position": "a1"      // Board-5's position
}
```

**Result**:
- Backend calculates midpoint position: `"a0V"`
- Board-2 gets position between "a0" and "a1"
- Only Board-2 is updated

**New State**:
```
Column: "In Progress"
├─ Board-4 (position: "a0")   ← Unchanged
├─ Board-2 (position: "a0V")  ← Inserted! (a0 < a0V < a1)
└─ Board-5 (position: "a1")   ← Unchanged
```

---

### Scenario 4: Move to Empty Column

**UI Action**: User drags "Board-2" to empty "Done" column

**API Request**:
```json
{
  "view_id": "view-123",
  "group_by_field_id": "status-field-id",
  "new_field_value": "done-option-id",
  "before_position": null,   // No boards in column
  "after_position": null
}
```

**Result**:
- Backend generates initial position: `"a0"`
- Board-2 is the first board in "Done" column

---

## Migration from Old API

### Old API (Integer-based):

```json
{
  "view_id": "view-123",
  "group_by_field_id": "status-field-id",
  "new_field_value": "in-progress-option-id",
  "new_display_order": 2  // ❌ Old way
}
```

**Problems**:
- Required calculating exact integer position
- Moving to position 2 meant updating all boards from position 2 onwards
- High DB load for large lists

### New API (Fractional Indexing):

```json
{
  "view_id": "view-123",
  "group_by_field_id": "status-field-id",
  "new_field_value": "in-progress-option-id",
  "before_position": "a1",  // ✅ New way
  "after_position": "a2"
}
```

**Benefits**:
- No need to calculate exact position
- Backend handles position generation
- Only 1 row updated regardless of list size

---

## Error Handling

### Common Errors:

#### 1. Invalid Position Order

```json
{
  "error": "INVALID_POSITION_ORDER",
  "message": "before_position must be < after_position lexicographically"
}
```

**Cause**: Frontend sent `before_position` that is >= `after_position`

**Fix**: Ensure proper ordering before sending request

#### 2. Board Not Found

```json
{
  "error": "BOARD_NOT_FOUND",
  "message": "Board not found or you don't have permission"
}
```

**Cause**: Invalid board ID or user doesn't have access

#### 3. View Not Found

```json
{
  "error": "VIEW_NOT_FOUND",
  "message": "Saved view not found"
}
```

**Cause**: Invalid view ID

#### 4. Field Not Found

```json
{
  "error": "FIELD_NOT_FOUND",
  "message": "Custom field not found"
}
```

**Cause**: Invalid `group_by_field_id`

#### 5. Option Not Found

```json
{
  "error": "OPTION_NOT_FOUND",
  "message": "Field option not found"
}
```

**Cause**: Invalid `new_field_value` (option ID doesn't exist)

---

## TypeScript Example

Complete implementation example:

```typescript
import axios from 'axios';

interface MoveBoardRequest {
  view_id: string;
  group_by_field_id: string;
  new_field_value: string;
  before_position?: string;
  after_position?: string;
}

interface MoveBoardResponse {
  board_id: string;
  new_field_value: string;
  new_position: string;
  message: string;
}

class BoardAPI {
  private baseURL: string;

  constructor(baseURL: string) {
    this.baseURL = baseURL;
  }

  async moveBoard(
    boardId: string,
    request: MoveBoardRequest
  ): Promise<MoveBoardResponse> {
    const response = await axios.post<MoveBoardResponse>(
      `${this.baseURL}/api/v1/boards/${boardId}/move`,
      request,
      {
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${this.getAuthToken()}`
        }
      }
    );

    return response.data;
  }

  private getAuthToken(): string {
    // Implement your auth token retrieval
    return localStorage.getItem('auth_token') || '';
  }
}

// Usage in React component
function BoardColumn({
  boards,
  columnId,
  viewId,
  groupByFieldId
}: BoardColumnProps) {
  const api = new BoardAPI(process.env.REACT_APP_API_URL);

  const onDrop = async (
    draggedBoardId: string,
    targetIndex: number
  ) => {
    // Find before and after boards
    const beforeBoard = targetIndex > 0 ? boards[targetIndex - 1] : null;
    const afterBoard = targetIndex < boards.length ? boards[targetIndex] : null;

    try {
      const response = await api.moveBoard(draggedBoardId, {
        view_id: viewId,
        group_by_field_id: groupByFieldId,
        new_field_value: columnId,
        before_position: beforeBoard?.position,
        after_position: afterBoard?.position
      });

      // Update local state
      updateBoardInState(draggedBoardId, {
        position: response.new_position,
        customFields: {
          [groupByFieldId]: response.new_field_value
        }
      });

      console.log(`Board moved to position: ${response.new_position}`);
    } catch (error) {
      console.error('Failed to move board:', error);
      // Revert UI state
      revertBoardPosition(draggedBoardId);
    }
  };

  return (
    <div className="board-column">
      {boards.map((board, index) => (
        <BoardCard
          key={board.id}
          board={board}
          onDrop={(draggedId) => onDrop(draggedId, index)}
        />
      ))}
    </div>
  );
}
```

---

## Performance Considerations

### 1. Optimistic Updates

Update UI immediately, then sync with server:

```typescript
// 1. Update UI optimistically
updateUIImmediately(boardId, newPosition);

// 2. Send API request
try {
  const response = await api.moveBoard(boardId, request);
  // 3. Confirm with server response
  confirmUIUpdate(boardId, response.new_position);
} catch (error) {
  // 4. Revert on error
  revertUIUpdate(boardId);
}
```

### 2. Debounce Rapid Moves

If user drags multiple times quickly:

```typescript
const debouncedMove = debounce(async (boardId, request) => {
  await api.moveBoard(boardId, request);
}, 300);
```

### 3. Batch Fetch Boards

Fetch all boards for a view in one request:

```typescript
// ✅ Good: Single request
const boards = await api.getBoardsByView(viewId);

// ❌ Bad: Multiple requests
for (const column of columns) {
  await api.getBoardsByColumn(column.id);  // N+1 problem
}
```

---

## Testing Tips

### Test Cases:

1. **Move to first position**: `before_position: null, after_position: "a0"`
2. **Move to last position**: `before_position: "z9", after_position: null`
3. **Move between two boards**: Both positions provided
4. **Move to empty column**: Both positions `null`
5. **Cross-column move**: Change `new_field_value`
6. **Same position (no-op)**: Should succeed without changes

### Example Test:

```typescript
describe('Board Movement', () => {
  it('should move board to first position', async () => {
    const response = await api.moveBoard('board-123', {
      view_id: 'view-456',
      group_by_field_id: 'field-789',
      new_field_value: 'option-abc',
      before_position: null,
      after_position: 'a0'
    });

    expect(response.new_position).toBeLessThan('a0');
  });

  it('should move board between two positions', async () => {
    const response = await api.moveBoard('board-123', {
      view_id: 'view-456',
      group_by_field_id: 'field-789',
      new_field_value: 'option-abc',
      before_position: 'a0',
      after_position: 'a1'
    });

    expect(response.new_position).toBeGreaterThan('a0');
    expect(response.new_position).toBeLessThan('a1');
  });
});
```

---

## FAQ

### Q: What if I don't know the before/after positions?

**A**: If you're inserting at the beginning or end of a list, you can omit one of the positions:
- First position: `before_position: null, after_position: <first-board-position>`
- Last position: `before_position: <last-board-position>, after_position: null`
- Empty column: Both `null`

### Q: Do I need to track positions on the frontend?

**A**: Yes, each board should store its `position` string. When rendering, sort boards by position using `localeCompare()`.

### Q: What happens if two users move boards simultaneously?

**A**: Each move generates a unique position. The database handles concurrent updates. Use WebSocket or polling to sync changes across users.

### Q: Can positions run out of space?

**A**: No. Fractional indexing has infinite precision. You can always insert between any two positions.

### Q: How do I sort boards?

**A**: Use lexicographic (alphabetical) sorting:

```typescript
boards.sort((a, b) => a.position.localeCompare(b.position));
```

### Q: What if a position becomes very long?

**A**: Positions can grow (e.g., "a0VVVVVV..."). If needed, implement position rebalancing by periodically regenerating positions ("a0", "a1", "a2", etc.). This is rare and only needed after thousands of inserts.

---

## Summary

**Key Points**:
1. ✅ Use `before_position` and `after_position` in API requests
2. ✅ Sort boards lexicographically by position
3. ✅ Only 1 board is updated per move (O(1) operation)
4. ✅ Backend generates optimal positions automatically
5. ✅ Handle errors gracefully with optimistic updates

**Performance**:
- Old system: Moving to position 50 in a 100-board list = 50 DB updates
- New system: Moving anywhere = 1 DB update

**Developer Experience**:
- Old: Calculate exact integer position, handle reordering logic
- New: Just provide before/after positions, backend handles the rest

For more details, see:
- [Fractional Indexing Blog Post](https://www.figma.com/blog/realtime-editing-of-ordered-sequences/)
- [Custom Fields API Documentation](../CUSTOM_FIELDS_API.md)
