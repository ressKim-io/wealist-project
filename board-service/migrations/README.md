# Database Migrations

This directory contains SQL migration files for the Board Service database schema.

## 📋 File Naming Convention

**Format**: `{YYYYMMDDHHMMSS}_{description}.{up|down}.sql`

**Examples**:
- `20250106120000_baseline_v1.0.0.up.sql`
- `20250106120000_baseline_v1.0.0.down.sql`
- `20250110153000_add_attachments_table.up.sql`
- `20250110153000_add_attachments_table.down.sql`

**Execution Order**: Alphabetical order (timestamp-based)

## 🚀 Execution Methods

### Development Environment (AutoMigrate)

For rapid prototyping, use GORM AutoMigrate:

```bash
# Set environment variable
export USE_AUTO_MIGRATE=true

# Run server (AutoMigrate executes automatically)
go run cmd/api/main.go
```

**Pros**:
- Fast iteration
- No manual SQL writing
- Automatic schema updates

**Cons**:
- Less control over migrations
- Cannot rollback easily
- Not suitable for production

### Production Environment (Manual Migrations)

For production, use manual SQL migrations:

```bash
# Apply all pending migrations
./scripts/db/apply_migrations.sh prod

# Rollback specific migration
./scripts/db/rollback.sh prod 20250106120000

# Dump current schema
./scripts/db/dump_schema.sh prod
```

**Pros**:
- Full control over schema changes
- Rollback capability
- Audit trail via `schema_versions` table
- Production-safe

**Cons**:
- Requires manual SQL writing
- More time-consuming

## 📝 Creating New Migrations

### Step 1: Modify Domain Models

Update models in `internal/domain/`:

```go
// internal/domain/attachment.go
type Attachment struct {
    BaseModel
    BoardID   uuid.UUID `gorm:"type:uuid;not null;index" json:"board_id"`
    FileName  string    `gorm:"type:varchar(255);not null" json:"file_name"`
    FileURL   string    `gorm:"type:text;not null" json:"file_url"`
    FileSize  int64     `gorm:"not null" json:"file_size"`
    IsDeleted bool      `gorm:"default:false" json:"is_deleted"`
}
```

### Step 2: Test with AutoMigrate

```bash
# Run locally with AutoMigrate
ENV=dev USE_AUTO_MIGRATE=true go run cmd/api/main.go
```

### Step 3: Dump Schema

```bash
# Dump current schema to docs/schema/
./scripts/db/dump_schema.sh dev
```

### Step 4: Create Migration Files

```bash
# Create up migration
cat > migrations/20250110153000_add_attachments_table.up.sql << 'EOF'
CREATE TABLE IF NOT EXISTS attachments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    board_id UUID NOT NULL,
    file_name VARCHAR(255) NOT NULL,
    file_url TEXT NOT NULL,
    file_size BIGINT NOT NULL,
    is_deleted BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_attachments_board_id ON attachments(board_id);

COMMENT ON COLUMN attachments.board_id IS 'References boards.id (no FK for sharding compatibility)';

INSERT INTO schema_versions (version, description)
VALUES ('20250110153000', 'Add attachments table');
EOF

# Create down migration
cat > migrations/20250110153000_add_attachments_table.down.sql << 'EOF'
DROP TABLE IF EXISTS attachments CASCADE;
DELETE FROM schema_versions WHERE version = '20250110153000';
EOF
```

### Step 5: Test Migration

```bash
# Apply migration
./scripts/db/apply_migrations.sh dev

# Test rollback
./scripts/db/rollback.sh dev 20250110153000

# Reapply to verify idempotency
./scripts/db/apply_migrations.sh dev
```

### Step 6: Update AutoMigrate List

Add new model to `internal/database/postgres.go`:

```go
func autoMigrateAll(db *gorm.DB, logger *zap.Logger) error {
    models := []interface{}{
        &domain.SchemaVersion{},
        // ... existing models ...
        &domain.Attachment{}, // Add new model
    }
    return db.AutoMigrate(models...)
}
```

### Step 7: Create Pull Request

- Include both `.up.sql` and `.down.sql` files
- Update CLAUDE.md if needed
- Test on staging environment before production

## 🗄️ Schema Version Tracking

All applied migrations are tracked in the `schema_versions` table:

```sql
SELECT version, description, applied_at
FROM schema_versions
ORDER BY applied_at DESC;
```

## 📚 Migration History

| Version | Description | Applied Date |
|---------|-------------|--------------|
| 20250106120000 | Baseline v1.0.0 - Initial schema consolidation | 2025-01-06 |

## ⚠️ Important Rules

1. **Never modify committed migrations** - Create new ones instead
2. **Always write both up and down migrations** - Rollback must be possible
3. **Test migrations locally first** - Apply → Rollback → Reapply
4. **No Foreign Keys** - Use comments instead: `-- References table.column (no FK)`
5. **UUID Primary Keys** - Always use `UUID DEFAULT gen_random_uuid()`
6. **Idempotent SQL** - Use `IF NOT EXISTS` / `IF EXISTS` clauses
7. **Comment your migrations** - Explain complex schema changes

## 🔍 Troubleshooting

### Migration fails with "relation already exists"

The migration is not idempotent. Add `IF NOT EXISTS`:

```sql
CREATE TABLE IF NOT EXISTS my_table (...);
```

### AutoMigrate vs Manual Migration mismatch

Dump both schemas and compare:

```bash
# AutoMigrate schema
ENV=dev USE_AUTO_MIGRATE=true go run cmd/api/main.go
./scripts/db/dump_schema.sh dev

# Manual migration schema
./scripts/db/apply_migrations.sh dev
./scripts/db/dump_schema.sh dev

# Compare
diff docs/schema/backups/schema_*.sql
```

### Rollback fails

Check if the down migration correctly reverses the up migration. Test locally first.

## 📖 Additional Resources

- [GORM Documentation](https://gorm.io/docs/)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [Database Migration Best Practices](https://www.prisma.io/dataguide/types/relational/what-are-database-migrations)
