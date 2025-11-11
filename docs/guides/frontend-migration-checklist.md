# Frontend Migration Checklist

## 🎯 Quick Start

**Priority:** HIGH
**Estimated Effort:** 2-3 days
**Breaking Changes:** YES - All custom fields APIs changed

---

## ✅ Step-by-Step Migration

### Phase 1: API Client Updates (Day 1 Morning)

- [ ] **Update API types/interfaces**
  ```typescript
  // ❌ Remove
  interface Board {
    custom_stage_id?: string;
    custom_importance_id?: string;
  }

  // ✅ Add
  interface Board {
    custom_fields_cache: Record<string, any>;
  }
  ```

- [ ] **Update API service methods**
  - [ ] Remove `getCustomStages()`, `getCustomRoles()`, `getCustomImportances()`
  - [ ] Add `getProjectFields(projectId)`
  - [ ] Add `getFieldValues(boardId)`
  - [ ] Add `setFieldValue(boardId, fieldId, value)`
  - [ ] Add `getViews(projectId)`
  - [ ] Add `applyView(viewId)`
  - [ ] Update `moveBoard()` signature

### Phase 2: Component Updates (Day 1 Afternoon)

#### 2.1 Board Form Component
- [ ] Remove hardcoded stage/role/importance dropdowns
- [ ] Add dynamic field renderer based on `field_type`
- [ ] Implement field value saving via `/field-values` API
- [ ] Add validation for required fields

```typescript
// Example dynamic field renderer
const renderField = (field: ProjectField) => {
  switch (field.field_type) {
    case 'single_select':
      return <Select options={field.options} />;
    case 'multi_select':
      return <MultiSelect options={field.options} />;
    case 'text':
      return <Input type="text" />;
    case 'number':
      return <Input type="number" />;
    case 'date':
      return <DatePicker />;
    default:
      return null;
  }
};
```

#### 2.2 Kanban Board Component
- [ ] Replace `custom_stage_id` with `custom_fields_cache[stageFieldId]`
- [ ] Fetch stage field dynamically from project fields
- [ ] Update column headers to use field options
- [ ] Fix board card rendering to use `custom_fields_cache`

#### 2.3 Drag & Drop Handler
- [ ] Update `onDragEnd` to use new `/boards/:id/move` API
- [ ] Remove old batch update logic
- [ ] Add `prev_board_id` and `next_board_id` calculation
- [ ] Test O(1) performance improvement

```typescript
// ✅ New drag handler
const onDragEnd = async (result) => {
  const { draggableId, destination } = result;

  await api.put(`/boards/${draggableId}/move`, {
    view_id: currentViewId,
    target_field_id: stageFieldId,
    new_field_value: destination.droppableId, // New stage
    prev_board_id: getPrevBoardId(destination),
    next_board_id: getNextBoardId(destination)
  });
};
```

### Phase 3: Filter & View Components (Day 2 Morning)

#### 3.1 Filter Component
- [ ] Remove hardcoded stage/role/importance filters
- [ ] Add dynamic filter UI based on project fields
- [ ] Implement filter state management
- [ ] Add "Save as View" functionality

```typescript
// ✅ Dynamic filters
const Filters = ({ projectFields }) => {
  return (
    <>
      {projectFields.map(field => (
        <FilterControl
          key={field.id}
          field={field}
          onChange={(value) => setFilter(field.id, value)}
        />
      ))}
    </>
  );
};
```

#### 3.2 View Management
- [ ] Add "Create View" modal
- [ ] Add "Edit View" functionality
- [ ] Add view selector dropdown
- [ ] Implement view switching

### Phase 4: Settings Pages (Day 2 Afternoon)

#### 4.1 Project Settings - Custom Fields
- [ ] Add "Custom Fields" settings page
- [ ] Implement field creation UI
- [ ] Implement field editing UI
- [ ] Implement field deletion with warning
- [ ] Add field reordering (drag & drop)

#### 4.2 Remove Old Settings
- [ ] Remove "Custom Stages" settings page
- [ ] Remove "Custom Roles" settings page
- [ ] Remove "Custom Importance" settings page

### Phase 5: Data Migration (Day 3 Morning)

- [ ] **Write migration script** (if needed for existing users)
  ```typescript
  async function migrateExistingData() {
    // 1. Fetch all boards with legacy fields
    const boards = await api.get('/boards?include_legacy=true');

    // 2. Create new fields if not exist
    const stageField = await createOrGetField('Stage', 'single_select');

    // 3. Migrate each board
    for (const board of boards) {
      if (board.custom_stage_id) {
        const stageName = await getOldStageName(board.custom_stage_id);
        await api.post('/field-values', {
          board_id: board.id,
          field_id: stageField.id,
          value: stageName
        });
      }
    }
  }
  ```

### Phase 6: Testing (Day 3 Afternoon)

#### 6.1 Unit Tests
- [ ] Test dynamic field renderer
- [ ] Test field value state management
- [ ] Test drag & drop logic
- [ ] Test filter logic
- [ ] Test view management

#### 6.2 Integration Tests
- [ ] Test board creation with custom fields
- [ ] Test board editing with custom fields
- [ ] Test kanban drag & drop
- [ ] Test filtering by custom fields
- [ ] Test view creation and switching
- [ ] Test field CRUD operations

#### 6.3 E2E Tests
- [ ] Create project → Add fields → Create board
- [ ] Filter boards by multiple criteria
- [ ] Create view → Apply view → Edit view
- [ ] Drag board between columns
- [ ] Delete field (should handle gracefully)

---

## 🚨 Breaking Changes to Watch

### 1. Board Response Structure
```typescript
// ❌ Old
{
  "id": "uuid",
  "custom_stage_id": "stage-uuid",
  "custom_importance_id": "importance-uuid"
}

// ✅ New
{
  "id": "uuid",
  "custom_fields_cache": {
    "stage-field-uuid": "In Progress",
    "importance-field-uuid": "High"
  }
}
```

### 2. Create Board Request
```typescript
// ❌ Old
await api.post('/boards', {
  title: "New Board",
  custom_stage_id: "stage-uuid",
  custom_role_ids: ["role1", "role2"]
});

// ✅ New (2 API calls)
const board = await api.post('/boards', {
  title: "New Board",
  project_id: "project-uuid"
});

await api.post('/field-values/batch', {
  updates: [
    { board_id: board.id, field_id: stageFieldId, value: "To Do" },
    { board_id: board.id, field_id: roleFieldId, value: ["Frontend", "Backend"] }
  ]
});
```

### 3. Kanban View Data
```typescript
// ❌ Old
const view = await api.get(`/projects/${projectId}/orders/stage-board`);
// Returns: { columns: [{ stage_id, stage_name, boards: [...] }] }

// ✅ New
const view = await api.get(`/views/${viewId}/boards`);
// Returns: { boards: [...], grouped_by: "field_id", groups: [...] }
```

---

## 📦 New Dependencies (if needed)

```json
{
  "dependencies": {
    "@dnd-kit/core": "^6.0.0",           // Better drag & drop
    "react-select": "^5.7.0",            // Multi-select fields
    "react-datepicker": "^4.16.0",       // Date fields
    "immer": "^10.0.0"                   // Immutable state for complex filters
  }
}
```

---

## 🎨 UI/UX Improvements to Consider

### 1. Dynamic Field Management
- Add "+ Add Field" button in project settings
- Show field type icons (📋 single-select, ☑️ multi-select, 📅 date, etc.)
- Allow inline field editing in board cards

### 2. Advanced Filtering
- Add "AND/OR" logic between filters
- Add "Save filter as View" quick action
- Show active filters as chips

### 3. View Switcher
- Add view tabs at top of kanban board
- Add "Duplicate View" option
- Show view sharing options

---

## 🔍 Debugging Tips

### Check if fields are loaded correctly
```typescript
console.log('Project Fields:', projectFields);
// Expected: Array of ProjectField objects with options
```

### Check if field values are saved
```typescript
console.log('Board Custom Fields:', board.custom_fields_cache);
// Expected: { "field-uuid": "value", ... }
```

### Check view grouping
```typescript
const view = await api.get(`/views/${viewId}/boards`);
console.log('Grouped By:', view.grouped_by);
console.log('Groups:', view.groups);
```

---

## 📞 Get Help

- **API Issues:** Check `API_MIGRATION_GUIDE.md`
- **Backend Team:** @backend-team on Slack
- **Frontend Lead:** @frontend-lead on Slack

---

## ✨ Performance Wins

After migration, you should see:

- ✅ **50% faster board creation** (fewer API calls)
- ✅ **90% faster drag & drop** (O(1) vs O(n))
- ✅ **Unlimited custom fields** (not just stage/role/importance)
- ✅ **Better filtering** (any field, any operator)
- ✅ **Saved views** (no need to recreate filters)

---

**Last Updated:** 2025-11-10
**Status:** 🚧 Migration in progress
