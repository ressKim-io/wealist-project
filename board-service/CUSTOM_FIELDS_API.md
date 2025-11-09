# Custom Fields System API Documentation

> Jira-style custom fields system for flexible board management

## Overview

The Custom Fields System provides a flexible, extensible field management system similar to Jira. It supports 10 different field types with type-safe validation, Redis caching, and advanced filtering/sorting/grouping capabilities.

## Features

### ✨ Core Features
- **10 Field Types**: text, number, single_select, multi_select, date, datetime, single_user, multi_user, checkbox, url
- **Type-Safe Validation**: Automatic validation based on field type and configuration
- **EAV Pattern**: Entity-Attribute-Value pattern for dynamic field storage
- **JSONB Caching**: High-performance custom_fields_cache on boards table with GIN index
- **Redis Caching**: Multi-layer caching for fields, options, values, and views
- **Saved Views**: Filter, sort, and group boards with custom views
- **Multi-Dimensional Classification**: Multi-select fields support complex categorization
- **Permission Control**: ADMIN+ required for field/option management

### 📊 Supported Field Types

| Type | Description | Value Column | Example Config |
|------|-------------|--------------|----------------|
| `text` | Short or long text | `value_text` | `{"max_length": 500, "is_long": false}` |
| `number` | Numeric values | `value_number` | `{"min": 0, "max": 100, "decimal_places": 2}` |
| `single_select` | Single choice from options | `value_option_id` | `{}` |
| `multi_select` | Multiple choices from options | `value_option_id` (multiple rows) | `{"max_selections": 5}` |
| `date` | Date only | `value_date` | `{}` |
| `datetime` | Date and time | `value_date` | `{"include_time": true}` |
| `single_user` | Single user assignment | `value_user_id` | `{"project_members_only": true}` |
| `multi_user` | Multiple user assignments | `value_user_id` (multiple rows) | `{"max_users": 3}` |
| `checkbox` | Boolean value | `value_boolean` | `{"default_value": false}` |
| `url` | Web link | `value_text` | `{"enable_preview": true}` |

---

## API Endpoints

### 1. Field Management

#### Create Field
```http
POST /api/fields
Authorization: Bearer {token}
Content-Type: application/json

{
    "project_id": "uuid",
    "name": "Priority",
    "field_type": "single_select",
    "description": "Task priority level",
    "is_required": true,
    "config": {
        "max_selections": 1
    }
}
```

**Response** (200 OK):
```json
{
    "field_id": "uuid",
    "project_id": "uuid",
    "name": "Priority",
    "field_type": "single_select",
    "description": "Task priority level",
    "display_order": 0,
    "is_required": true,
    "is_system_default": false,
    "config": {},
    "created_at": "2025-01-10T10:00:00Z",
    "updated_at": "2025-01-10T10:00:00Z"
}
```

#### Get All Fields (Project)
```http
GET /api/projects/{project_id}/fields
Authorization: Bearer {token}
```

**Response** (200 OK):
```json
[
    {
        "field_id": "uuid",
        "name": "Priority",
        "field_type": "single_select",
        "display_order": 0,
        ...
    }
]
```

#### Get Field by ID
```http
GET /api/fields/{field_id}
Authorization: Bearer {token}
```

#### Update Field
```http
PATCH /api/fields/{field_id}
Authorization: Bearer {token}
Content-Type: application/json

{
    "name": "Updated Name",
    "description": "Updated description",
    "is_required": false
}
```

#### Delete Field
```http
DELETE /api/fields/{field_id}
Authorization: Bearer {token}
```

**Note**: System default fields cannot be deleted.

#### Update Field Order
```http
PUT /api/projects/{project_id}/field-orders
Authorization: Bearer {token}
Content-Type: application/json

{
    "field_orders": [
        {"field_id": "uuid-1", "display_order": 0},
        {"field_id": "uuid-2", "display_order": 1}
    ]
}
```

---

### 2. Field Options Management

#### Create Option
```http
POST /api/options
Authorization: Bearer {token}
Content-Type: application/json

{
    "field_id": "uuid",
    "label": "High",
    "color": "#FF0000",
    "description": "High priority"
}
```

**Response** (200 OK):
```json
{
    "option_id": "uuid",
    "field_id": "uuid",
    "label": "High",
    "color": "#FF0000",
    "description": "High priority",
    "display_order": 0,
    "created_at": "2025-01-10T10:00:00Z"
}
```

#### Get Options (by Field)
```http
GET /api/fields/{field_id}/options
Authorization: Bearer {token}
```

#### Get Option by ID
```http
GET /api/options/{option_id}
Authorization: Bearer {token}
```

#### Update Option
```http
PATCH /api/options/{option_id}
Authorization: Bearer {token}
Content-Type: application/json

{
    "label": "Very High",
    "color": "#FF0000"
}
```

#### Delete Option
```http
DELETE /api/options/{option_id}
Authorization: Bearer {token}
```

#### Update Option Order
```http
PUT /api/fields/{field_id}/option-orders
Authorization: Bearer {token}
Content-Type: application/json

{
    "option_orders": [
        {"option_id": "uuid-1", "display_order": 0},
        {"option_id": "uuid-2", "display_order": 1}
    ]
}
```

---

### 3. Field Values Management

#### Set Field Value
```http
POST /api/field-values
Authorization: Bearer {token}
Content-Type: application/json

{
    "board_id": "uuid",
    "field_id": "uuid",
    "value": "High" | 42 | "2025-01-10T00:00:00Z" | true | "option-uuid"
}
```

**Supported for**: text, number, date, datetime, single_select, single_user, checkbox, url

#### Set Multi-Select Value
```http
POST /api/field-values/multi-select
Authorization: Bearer {token}
Content-Type: application/json

{
    "board_id": "uuid",
    "field_id": "uuid",
    "values": [
        {"value": "option-uuid-1", "display_order": 0},
        {"value": "option-uuid-2", "display_order": 1}
    ]
}
```

**Supported for**: multi_select, multi_user

#### Get Board Field Values
```http
GET /api/boards/{board_id}/field-values
Authorization: Bearer {token}
```

**Response** (200 OK):
```json
{
    "board_id": "uuid",
    "field_values": [
        {
            "field_id": "uuid",
            "field_name": "Priority",
            "field_type": "single_select",
            "value": "option-uuid",
            "display_value": "High"
        },
        {
            "field_id": "uuid",
            "field_name": "Tags",
            "field_type": "multi_select",
            "values": [
                {"value": "option-uuid-1", "label": "Bug", "display_order": 0},
                {"value": "option-uuid-2", "label": "Frontend", "display_order": 1}
            ]
        }
    ]
}
```

#### Delete Field Value
```http
DELETE /api/boards/{board_id}/fields/{field_id}/values
Authorization: Bearer {token}
```

---

### 4. Saved Views (Filters/Sorting/Grouping)

#### Create View
```http
POST /api/views
Authorization: Bearer {token}
Content-Type: application/json

{
    "project_id": "uuid",
    "name": "High Priority Bugs",
    "description": "All high priority bugs",
    "is_shared": true,
    "filters": {
        "priority-field-uuid": {
            "operator": "eq",
            "value": "high-option-uuid"
        },
        "tags-field-uuid": {
            "operator": "in",
            "value": ["bug-option-uuid"]
        }
    },
    "sort_by": "created_at",
    "sort_direction": "desc",
    "group_by_field_id": "tags-field-uuid"
}
```

**Supported Operators**:
- `eq`: Equals
- `ne`: Not equals
- `contains`: Contains substring
- `in`: Value in array
- `gt`: Greater than
- `lt`: Less than
- `gte`: Greater than or equal
- `lte`: Less than or equal

**Response** (200 OK):
```json
{
    "view_id": "uuid",
    "project_id": "uuid",
    "created_by": "uuid",
    "name": "High Priority Bugs",
    "description": "All high priority bugs",
    "is_default": false,
    "is_shared": true,
    "filters": {...},
    "sort_by": "created_at",
    "sort_direction": "desc",
    "group_by_field_id": "uuid",
    "created_at": "2025-01-10T10:00:00Z"
}
```

#### Get View by ID
```http
GET /api/views/{view_id}
Authorization: Bearer {token}
```

#### Get Views by Project
```http
GET /api/projects/{project_id}/views
Authorization: Bearer {token}
```

#### Update View
```http
PATCH /api/views/{view_id}
Authorization: Bearer {token}
Content-Type: application/json

{
    "name": "Updated View Name",
    "is_shared": false
}
```

#### Delete View
```http
DELETE /api/views/{view_id}
Authorization: Bearer {token}
```

**Note**: Only the creator can delete a view.

#### Apply View (Get Filtered Boards)
```http
GET /api/views/{view_id}/boards?page=1&limit=20
Authorization: Bearer {token}
```

**Response** (200 OK - No Grouping):
```json
{
    "boards": [
        {
            "id": "uuid",
            "title": "Task 1",
            "created_at": "2025-01-10T10:00:00Z",
            ...
        }
    ],
    "total": 42,
    "page": 1,
    "limit": 20
}
```

**Response** (200 OK - With Grouping):
```json
{
    "groups": [
        {
            "group_id": "bug-option-uuid",
            "group_name": "Bug",
            "boards": [
                {"id": "uuid", "title": "Task 1", ...}
            ],
            "count": 10
        },
        {
            "group_id": "feature-option-uuid",
            "group_name": "Feature",
            "boards": [
                {"id": "uuid", "title": "Task 2", ...}
            ],
            "count": 5
        }
    ],
    "total": 15
}
```

#### Update Board Order in View
```http
PUT /api/view-board-orders
Authorization: Bearer {token}
Content-Type: application/json

{
    "view_id": "uuid",
    "board_orders": [
        {"board_id": "uuid-1", "display_order": 0},
        {"board_id": "uuid-2", "display_order": 1}
    ]
}
```

---

## Performance Optimizations

### 1. JSONB Caching
- All field values cached in `custom_fields_cache` column (JSONB type)
- GIN index on `custom_fields_cache` for fast filtering
- Automatic cache update on field value changes

### 2. Redis Caching
- **Project Fields**: TTL 5 minutes
- **Field Options**: TTL 5 minutes
- **Board Field Values**: Invalidated on update
- **View Results**: Cached with filter hash, invalidated on view deletion

### 3. Cache Invalidation Strategy
- **Write-through**: DB update → Cache invalidation
- **Set-based tracking**: No blocking SCAN operations
- **Atomic operations**: Redis Pipeline for consistency

---

## Testing

### Run Integration Tests
```bash
./test-custom-fields.sh
```

This script tests all 22 endpoints with:
- Field CRUD operations
- Option management
- Field value setting
- Saved views
- Cache performance

### Import Postman Collection
```bash
# Import Postman_Custom_Fields_Collection.json
```

### Run Unit Tests
```bash
cd internal/service
go test -v
```

---

## Migration Guide

### From Fixed Fields to Custom Fields

**Before** (Fixed structure):
- Role (fixed list)
- Stage (fixed list)
- Importance (fixed list)

**After** (Flexible system):
1. Create custom fields to replace fixed fields:
   ```http
   POST /api/fields
   {
       "name": "Status",
       "field_type": "single_select",
       ...
   }
   ```

2. Create options for select fields:
   ```http
   POST /api/options
   {
       "field_id": "status-field-uuid",
       "label": "In Progress",
       ...
   }
   ```

3. Set values on boards:
   ```http
   POST /api/field-values
   {
       "board_id": "board-uuid",
       "field_id": "status-field-uuid",
       "value": "in-progress-option-uuid"
   }
   ```

**Benefits**:
- ✅ Unlimited custom fields
- ✅ 10 different field types
- ✅ Dynamic filtering/grouping
- ✅ Better performance (JSONB + Redis)

---

## Database Schema

### Tables
- `project_fields`: Field definitions (10 types)
- `field_options`: Options for select-type fields
- `board_field_values`: EAV pattern for field values
- `saved_views`: Filter/sort/group configurations
- `user_board_order`: Manual board ordering within views
- `boards.custom_fields_cache`: JSONB cache column with GIN index

### Key Features
- **No Foreign Keys**: Sharding-ready design
- **Soft Deletes**: `is_deleted` flag on all tables
- **UUID Primary Keys**: Distributed-safe identifiers
- **JSONB Optimization**: Native PostgreSQL JSON support

---

## Error Handling

### Common Error Codes

| Status | Code | Message |
|--------|------|---------|
| 400 | `BAD_REQUEST` | Invalid field type / Invalid config |
| 403 | `FORBIDDEN` | Permission denied (ADMIN+ required) |
| 404 | `NOT_FOUND` | Field not found / Option not found |
| 500 | `INTERNAL_SERVER_ERROR` | Database error / Cache error |

### Error Response Format
```json
{
    "code": "FORBIDDEN",
    "message": "필드 생성 권한이 없습니다 (ADMIN 이상)"
}
```

---

## Best Practices

### 1. Field Design
- Use `single_select` for mutually exclusive choices
- Use `multi_select` for multi-dimensional classification
- Set `is_required` carefully to avoid blocking board creation
- Use meaningful `display_order` for consistent UI

### 2. Performance
- Leverage Redis cache for frequently accessed fields
- Use JSONB filters instead of JOIN-heavy queries
- Paginate view results for large datasets
- Cache view results with filter hash

### 3. Security
- Only ADMIN+ can create/modify fields
- Only view creator can delete views
- Validate all field values before setting
- Sanitize user input for text/url fields

---

## Example Workflows

### Workflow 1: Create Priority Field
```bash
# 1. Create field
curl -X POST http://localhost:8081/api/fields \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "project_id": "xxx",
    "name": "Priority",
    "field_type": "single_select"
  }'

# 2. Create options
for priority in "High" "Medium" "Low"; do
  curl -X POST http://localhost:8081/api/options \
    -H "Authorization: Bearer $TOKEN" \
    -d '{
      "field_id": "field-uuid",
      "label": "'$priority'",
      "color": "#FF0000"
    }'
done

# 3. Set value on board
curl -X POST http://localhost:8081/api/field-values \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "board_id": "board-uuid",
    "field_id": "field-uuid",
    "value": "high-option-uuid"
  }'
```

### Workflow 2: Create Filtered View
```bash
# Create view for high priority bugs
curl -X POST http://localhost:8081/api/views \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "project_id": "xxx",
    "name": "High Priority Bugs",
    "is_shared": true,
    "filters": {
      "priority-field-uuid": {"operator": "eq", "value": "high-option-uuid"},
      "tags-field-uuid": {"operator": "in", "value": ["bug-option-uuid"]}
    },
    "sort_by": "created_at",
    "sort_direction": "desc"
  }'

# Apply view
curl http://localhost:8081/api/views/{view-uuid}/boards?page=1&limit=20 \
  -H "Authorization: Bearer $TOKEN"
```

---

## Support

For issues or questions:
- GitHub Issues: [wealist-project/issues](https://github.com/ressKim-io/wealist-project/issues)
- Documentation: This file and README.md
- Test Scripts: `test-custom-fields.sh`
- Postman Collection: `Postman_Custom_Fields_Collection.json`

---

**Version**: 1.0.0
**Last Updated**: 2025-01-10
**Commits**:
- `388e918`: Phase 1 (Core Infrastructure)
- `b1f9f91`: Phase 2 (Services + Handlers)
- `56e7fb3`: Phase 3 (SavedView + Redis)
- `3ab2483`: JSONB Optimization
- `90fe22e`: Redis Cache Integration
