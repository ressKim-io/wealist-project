# API Migration Guide - Legacy to New Custom Fields System

**Last Updated:** 2025-11-10
**Branch:** `claude/go-testing-review-011CUz9Kfrhqp4oyvy5WPJzW`

## 📋 Overview

This guide documents the migration from the legacy custom fields system (CustomRole, CustomStage, CustomImportance) to the new unified ProjectField system.

---

## ❌ Removed APIs (Legacy System)

### 1. User Order Management (6 endpoints removed)

```
DELETE /api/projects/:project_id/orders/role-board
DELETE /api/projects/:project_id/orders/stage-board
DELETE /api/projects/:project_id/orders/role-columns
DELETE /api/projects/:project_id/orders/stage-columns
DELETE /api/projects/:project_id/orders/role-boards/:role_id
DELETE /api/projects/:project_id/orders/stage-boards/:stage_id
```

**Replacement:** Use View system with drag-and-drop ordering

```
GET  /api/views/:view_id/boards              # Get boards with filters/grouping
PUT  /api/boards/:board_id/move              # Move board with O(1) fractional indexing
PUT  /api/view-board-orders                  # Update manual board order in view
```

---

### 2. Custom Roles (6 endpoints removed)

```
DELETE POST   /api/custom-fields/roles
DELETE GET    /api/custom-fields/projects/:project_id/roles
DELETE GET    /api/custom-fields/roles/:role_id
DELETE PUT    /api/custom-fields/roles/:role_id
DELETE DELETE /api/custom-fields/roles/:role_id
DELETE PUT    /api/custom-fields/projects/:project_id/roles/order
```

**Replacement:** Use ProjectField API

```http
# Create a "Role" field
POST /api/fields
{
  "project_id": "uuid",
  "name": "Role",
  "field_type": "single_select",
  "options": [
    {"label": "Frontend", "color": "#FF5733"},
    {"label": "Backend", "color": "#33FF57"}
  ]
}

# Get all fields for a project
GET /api/projects/:project_id/fields

# Update field
PATCH /api/fields/:field_id

# Delete field
DELETE /api/fields/:field_id
```

---

### 3. Custom Stages (6 endpoints removed)

```
DELETE POST   /api/custom-fields/stages
DELETE GET    /api/custom-fields/projects/:project_id/stages
DELETE GET    /api/custom-fields/stages/:stage_id
DELETE PUT    /api/custom-fields/stages/:stage_id
DELETE DELETE /api/custom-fields/stages/:stage_id
DELETE PUT    /api/custom-fields/projects/:project_id/stages/order
```

**Replacement:** Use ProjectField API (same as Roles)

```http
# Create a "Stage" field
POST /api/fields
{
  "project_id": "uuid",
  "name": "Stage",
  "field_type": "single_select",
  "options": [
    {"label": "To Do", "color": "#CCCCCC"},
    {"label": "In Progress", "color": "#4169E1"},
    {"label": "Done", "color": "#32CD32"}
  ]
}
```

---

### 4. Custom Importance (6 endpoints removed)

```
DELETE POST   /api/custom-fields/importance
DELETE GET    /api/custom-fields/projects/:project_id/importance
DELETE GET    /api/custom-fields/importance/:importance_id
DELETE PUT    /api/custom-fields/importance/:importance_id
DELETE DELETE /api/custom-fields/importance/:importance_id
DELETE PUT    /api/custom-fields/projects/:project_id/importance/order
```

**Replacement:** Use ProjectField API (same as above)

```http
# Create an "Importance" field
POST /api/fields
{
  "project_id": "uuid",
  "name": "Importance",
  "field_type": "single_select",
  "options": [
    {"label": "Low", "color": "#90EE90"},
    {"label": "Medium", "color": "#FFA500"},
    {"label": "High", "color": "#FF4500"}
  ]
}
```

---

## ✅ New APIs (Unified System)

### 1. Field Management

```http
# Create custom field
POST /api/fields
Content-Type: application/json
{
  "project_id": "uuid",
  "name": "Priority",
  "field_type": "single_select",  # single_select, multi_select, text, number, date
  "is_required": false,
  "options": [
    {"label": "P0", "color": "#FF0000"},
    {"label": "P1", "color": "#FFA500"}
  ]
}

# Get all fields for project
GET /api/projects/:project_id/fields

# Get specific field
GET /api/fields/:field_id

# Update field
PATCH /api/fields/:field_id
{
  "name": "Updated Name",
  "options": [...]
}

# Delete field
DELETE /api/fields/:field_id
```

---

### 2. Field Values (Board-Specific)

```http
# Set field value for a board
POST /api/field-values
{
  "board_id": "uuid",
  "field_id": "uuid",
  "value": "P0"  # For single_select
  # OR
  "value": ["P0", "P1"]  # For multi_select
  # OR
  "value": "Some text"  # For text
  # OR
  "value": 42  # For number
}

# Get all field values for a board
GET /api/boards/:board_id/field-values

# Update field value
PUT /api/field-values/:value_id
{
  "value": "P1"
}

# Delete field value
DELETE /api/field-values/:value_id

# Batch update field values
POST /api/field-values/batch
{
  "updates": [
    {"board_id": "uuid1", "field_id": "uuid2", "value": "P0"},
    {"board_id": "uuid3", "field_id": "uuid4", "value": "P1"}
  ]
}
```

---

### 3. Views (Filtering, Sorting, Grouping)

```http
# Create saved view
POST /api/views
{
  "project_id": "uuid",
  "name": "High Priority Items",
  "filters": {
    "field_uuid": {
      "operator": "equals",
      "value": "P0"
    }
  },
  "sort_by": "created_at",
  "sort_direction": "desc",
  "group_by_field_id": "stage_field_uuid"  # Optional: Group by stage
}

# Get all views for project
GET /api/projects/:project_id/views

# Get specific view
GET /api/views/:view_id

# Update view
PATCH /api/views/:view_id

# Delete view
DELETE /api/views/:view_id

# Apply view (get filtered/sorted/grouped boards)
GET /api/views/:view_id/boards?page=1&limit=20
```

---

### 4. Board Movement (Drag & Drop)

```http
# Move board to different column/group (O(1) operation)
PUT /api/boards/:board_id/move
{
  "view_id": "uuid",
  "target_field_id": "stage_field_uuid",
  "new_field_value": "In Progress",
  "prev_board_id": "uuid_or_null",  # Board before this one
  "next_board_id": "uuid_or_null"   # Board after this one
}

# Response
{
  "board_id": "uuid",
  "new_field_value": "In Progress",
  "new_position": "a1V",  # Fractional index
  "message": "보드가 성공적으로 이동되었습니다 (O(1) 연산)"
}
```

---

## 🔄 Data Model Changes

### Board Model (Before)

```typescript
interface Board {
  id: string;
  title: string;
  description: string;
  project_id: string;

  // ❌ REMOVED: Legacy fields
  custom_stage_id?: string;
  custom_importance_id?: string;
  custom_role_ids?: string[];  // Many-to-many via board_roles
}
```

### Board Model (After)

```typescript
interface Board {
  id: string;
  title: string;
  description: string;
  project_id: string;

  // ✅ NEW: All custom fields in JSONB
  custom_fields_cache: {
    [field_id: string]: any;  // Field ID -> Value mapping
  };

  // Example:
  // custom_fields_cache: {
  //   "stage_field_uuid": "In Progress",
  //   "priority_field_uuid": "P0",
  //   "assignee_field_uuid": ["user1", "user2"]
  // }
}
```

---

## 📊 Migration Examples

### Example 1: Creating a Kanban Board

**Before (Legacy):**
```typescript
// 1. Create custom stages
const stages = await Promise.all([
  api.post('/custom-fields/stages', { name: 'To Do', color: '#CCC' }),
  api.post('/custom-fields/stages', { name: 'In Progress', color: '#4169E1' }),
  api.post('/custom-fields/stages', { name: 'Done', color: '#32CD32' })
]);

// 2. Get stage-based board view
const boardView = await api.get(`/projects/${projectId}/orders/stage-board`);

// 3. Drag and drop
await api.put(`/projects/${projectId}/orders/stage-boards/${stageId}`, {
  item_ids: [board1, board2, board3]
});
```

**After (New System):**
```typescript
// 1. Create "Stage" field
const stageField = await api.post('/fields', {
  project_id: projectId,
  name: 'Stage',
  field_type: 'single_select',
  options: [
    { label: 'To Do', color: '#CCC' },
    { label: 'In Progress', color: '#4169E1' },
    { label: 'Done', color: '#32CD32' }
  ]
});

// 2. Create a view with stage grouping
const view = await api.post('/views', {
  project_id: projectId,
  name: 'Kanban View',
  group_by_field_id: stageField.id,
  sort_by: 'created_at'
});

// 3. Get grouped boards
const boards = await api.get(`/views/${view.id}/boards`);

// 4. Drag and drop (move board to "In Progress")
await api.put(`/boards/${boardId}/move`, {
  view_id: view.id,
  target_field_id: stageField.id,
  new_field_value: 'In Progress',
  prev_board_id: null,  // First in column
  next_board_id: existingBoardId
});
```

---

### Example 2: Filtering Boards by Priority

**Before (Legacy):**
```typescript
// Limited filtering - only by stage/role/importance IDs
const boards = await api.get(`/boards?project_id=${projectId}&stage_id=${stageId}`);
```

**After (New System):**
```typescript
// Create view with flexible filters
const view = await api.post('/views', {
  project_id: projectId,
  name: 'High Priority',
  filters: {
    [priorityFieldId]: {
      operator: 'equals',
      value: 'P0'
    },
    [statusFieldId]: {
      operator: 'not_equals',
      value: 'Done'
    }
  },
  sort_by: 'created_at',
  sort_direction: 'desc'
});

// Apply view
const boards = await api.get(`/views/${view.id}/boards?page=1&limit=20`);
```

---

### Example 3: Multi-Select Roles

**Before (Legacy):**
```typescript
// Many-to-many via board_roles table
await api.post('/boards', {
  title: 'New Board',
  custom_role_ids: [role1, role2, role3]
});
```

**After (New System):**
```typescript
// 1. Create board
const board = await api.post('/boards', {
  title: 'New Board',
  project_id: projectId
});

// 2. Create Role field (if not exists)
const roleField = await api.post('/fields', {
  project_id: projectId,
  name: 'Roles',
  field_type: 'multi_select',  // Multi-select for multiple roles
  options: [
    { label: 'Frontend', color: '#FF5733' },
    { label: 'Backend', color: '#33FF57' },
    { label: 'Design', color: '#3357FF' }
  ]
});

// 3. Set role values
await api.post('/field-values', {
  board_id: board.id,
  field_id: roleField.id,
  value: ['Frontend', 'Backend']  // Array for multi_select
});
```

---

## 🎨 Frontend Components to Update

### 1. Kanban Board Component

**Changes needed:**
- Replace `custom_stage_id` with `custom_fields_cache[stage_field_id]`
- Use View API for grouping instead of hardcoded stage columns
- Update drag-and-drop to use `/boards/:id/move` endpoint

```typescript
// Old
const stageId = board.custom_stage_id;

// New
const stageFieldId = '...'; // Get from project fields
const stageValue = board.custom_fields_cache[stageFieldId];
```

---

### 2. Board Form Component

**Changes needed:**
- Fetch all ProjectFields for the project
- Dynamically render field inputs based on `field_type`
- Save values using `/field-values` API instead of setting board properties directly

```typescript
// Old
<select name="stage">
  {stages.map(s => <option value={s.id}>{s.name}</option>)}
</select>

// New
{projectFields.map(field => (
  field.field_type === 'single_select' ? (
    <select onChange={e => setFieldValue(field.id, e.target.value)}>
      {field.options.map(opt => (
        <option value={opt.label}>{opt.label}</option>
      ))}
    </select>
  ) : field.field_type === 'multi_select' ? (
    <MultiSelect
      options={field.options}
      onChange={values => setFieldValue(field.id, values)}
    />
  ) : null
))}
```

---

### 3. Filter Component

**Changes needed:**
- Use View API for saving filters
- Support dynamic field-based filtering instead of hardcoded stage/role/importance

```typescript
// Old - Hardcoded filters
<select name="stage_filter">
  <option value="">All Stages</option>
  {stages.map(s => <option value={s.id}>{s.name}</option>)}
</select>

// New - Dynamic filters from ProjectFields
{projectFields.filter(f => f.field_type === 'single_select').map(field => (
  <FilterDropdown
    key={field.id}
    field={field}
    onChange={value => updateFilter(field.id, value)}
  />
))}
```

---

## 🚀 Performance Improvements

### 1. Drag & Drop

**Before:** O(n) - Update all items in column
```typescript
// Update all board orders (slow)
await api.put('/orders/stage-boards/stage1', {
  item_ids: [board1, board2, board3, ..., board100]  // All 100 items
});
```

**After:** O(1) - Fractional indexing
```typescript
// Update only 1 board (fast)
await api.put(`/boards/${boardId}/move`, {
  view_id: viewId,
  prev_board_id: prevBoard,
  next_board_id: nextBoard
  // Calculates position automatically: "a1" -> "a1V" -> "a2"
});
```

---

### 2. Custom Fields Storage

**Before:** Multiple tables + joins
```sql
-- 3+ queries with joins
SELECT * FROM boards WHERE id = ?;
SELECT * FROM custom_stages WHERE id = ?;
SELECT * FROM board_roles WHERE board_id = ?;
SELECT * FROM custom_roles WHERE id IN (?);
```

**After:** Single JSONB column with GIN index
```sql
-- 1 query with JSONB index
SELECT * FROM boards
WHERE custom_fields_cache @> '{"priority": "P0"}'::jsonb;
```

---

## 📝 Testing Checklist

- [ ] Update E2E tests for board creation
- [ ] Update E2E tests for drag-and-drop
- [ ] Update E2E tests for filtering
- [ ] Update E2E tests for views
- [ ] Test field value batch updates
- [ ] Test custom field deletion (cascading)
- [ ] Test view grouping by different fields
- [ ] Performance test: Create 1000 boards with custom fields
- [ ] Performance test: Move board 100 times (should be O(1))

---

## 🔗 API Reference

**Base URL:** `http://localhost:8000/api`

### Quick Reference

| Feature | Legacy Endpoint | New Endpoint |
|---------|----------------|--------------|
| Create Stage | `POST /custom-fields/stages` | `POST /fields` (field_type: single_select) |
| Get Stages | `GET /projects/:id/stages` | `GET /projects/:id/fields` |
| Create Role | `POST /custom-fields/roles` | `POST /fields` (field_type: multi_select) |
| Set Board Stage | `PUT /boards/:id` + `custom_stage_id` | `POST /field-values` |
| Kanban View | `GET /projects/:id/orders/stage-board` | `GET /views/:id/boards` (group_by) |
| Drag & Drop | `PUT /orders/stage-boards/:id` | `PUT /boards/:id/move` |
| Filter Boards | `GET /boards?stage_id=x` | `GET /views/:id/boards` (with filters) |

---

## ❓ FAQs

### Q: What happens to existing data?

A: You need to run a data migration:
1. Read all boards with `custom_stage_id`, `custom_importance_id`, `board_roles`
2. Create corresponding ProjectFields
3. Create FieldValues for each board
4. Update `custom_fields_cache` JSONB column

### Q: Can I still use the old endpoints temporarily?

A: No, they have been completely removed in this version. You must migrate to the new system.

### Q: How do I handle default values?

A: Set `default_value` when creating a ProjectField:
```json
{
  "name": "Stage",
  "field_type": "single_select",
  "default_value": "To Do",
  "options": [...]
}
```

### Q: Can I create custom field types?

A: Currently supported: `single_select`, `multi_select`, `text`, `number`, `date`, `checkbox`, `user`

### Q: How do I migrate existing board orders?

A: Use the new View system:
1. Create a View with grouping
2. Call `/view-board-orders` to set manual order
3. Fractional indexing will maintain order automatically on moves

---

## 📞 Support

If you have questions about this migration:
- Check the [Board Service README](./board-service/README.md)
- Review test files in `board-service/internal/service/*_test.go`
- Contact the backend team

---

**Generated:** 2025-11-10
**Author:** Claude (Automated)
