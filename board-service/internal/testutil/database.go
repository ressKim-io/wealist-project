package testutil

import (
	"board-service/internal/domain"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// TestDB wraps a GORM database connection for testing
type TestDB struct {
	DB     *gorm.DB
	DBPath string // For SQLite cleanup
}

// SetupTestDB creates an in-memory SQLite database for testing
func SetupTestDB() *TestDB {
	// Use in-memory SQLite for fast tests
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // Silence logs in tests
	})
	if err != nil {
		log.Fatalf("Failed to connect to test database: %v", err)
	}

	// Auto-migrate all domain models
	if err := migrateTestModels(db); err != nil {
		log.Fatalf("Failed to migrate test database: %v", err)
	}

	return &TestDB{
		DB:     db,
		DBPath: ":memory:",
	}
}

// SetupTestDBFile creates a file-based SQLite database for testing
func SetupTestDBFile() *TestDB {
	// Create temp directory
	tmpDir := os.TempDir()
	dbPath := filepath.Join(tmpDir, fmt.Sprintf("test_%d.db", os.Getpid()))

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatalf("Failed to connect to test database: %v", err)
	}

	if err := migrateTestModels(db); err != nil {
		log.Fatalf("Failed to migrate test database: %v", err)
	}

	return &TestDB{
		DB:     db,
		DBPath: dbPath,
	}
}

// SetupPostgresTestDB creates a PostgreSQL test database (for integration tests)
// Requires TEST_DATABASE_URL environment variable
func SetupPostgresTestDB() *TestDB {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		log.Println("TEST_DATABASE_URL not set, skipping Postgres test")
		return nil
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatalf("Failed to connect to Postgres test database: %v", err)
	}

	if err := migrateTestModels(db); err != nil {
		log.Fatalf("Failed to migrate test database: %v", err)
	}

	return &TestDB{
		DB:     db,
		DBPath: dsn,
	}
}

// migrateTestModels runs auto-migration for all domain models
func migrateTestModels(db *gorm.DB) error {
	return db.AutoMigrate(
		&domain.Project{},
		&domain.ProjectMember{},
		&domain.ProjectJoinRequest{},
		&domain.Role{},
		&domain.Board{},
		&domain.ProjectField{},
		&domain.FieldOption{},
		&domain.BoardFieldValue{},
		&domain.SavedView{},
		&domain.UserBoardOrder{},
		&domain.Comment{},
	)
}

// Teardown closes the database connection and removes temporary files
func (t *TestDB) Teardown() {
	sqlDB, err := t.DB.DB()
	if err == nil {
		sqlDB.Close()
	}

	// Remove file if not in-memory
	if t.DBPath != ":memory:" && t.DBPath != "" {
		os.Remove(t.DBPath)
	}
}

// Clean removes all data from tables (keeps schema)
func (t *TestDB) Clean() {
	// Order matters due to foreign keys
	tables := []interface{}{
		&domain.Comment{},
		&domain.UserBoardOrder{},
		&domain.SavedView{},
		&domain.BoardFieldValue{},
		&domain.FieldOption{},
		&domain.ProjectField{},
		&domain.Board{},
		&domain.ProjectJoinRequest{},
		&domain.ProjectMember{},
		&domain.Project{},
		&domain.Role{},
	}

	for _, table := range tables {
		t.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(table)
	}
}

// SeedRoles creates default roles (OWNER, ADMIN, MEMBER)
func (t *TestDB) SeedRoles() (*domain.Role, *domain.Role, *domain.Role) {
	owner := NewOwnerRole()
	admin := NewAdminRole()
	member := NewMemberRole()

	t.DB.Create(owner)
	t.DB.Create(admin)
	t.DB.Create(member)

	return owner, admin, member
}

// SeedProject creates a test project with owner member
func (t *TestDB) SeedProject(ownerRole *domain.Role) (*domain.Project, *domain.ProjectMember, uuid.UUID) {
	userID := uuid.New()
	project := NewTestProjectWithID(uuid.New(), userID)

	t.DB.Create(project)

	member := NewTestProjectOwner(project.ID, userID, ownerRole)
	t.DB.Create(member)

	return project, member, userID
}

// SeedBasicSetup creates roles + project + owner (most common test setup)
func (t *TestDB) SeedBasicSetup() (*domain.Project, *domain.Role, uuid.UUID) {
	ownerRole, _, _ := t.SeedRoles()
	project, _, userID := t.SeedProject(ownerRole)
	return project, ownerRole, userID
}

// BeginTransaction starts a transaction for testing (can be rolled back)
func (t *TestDB) BeginTransaction() *gorm.DB {
	return t.DB.Begin()
}

// RollbackTransaction rolls back a transaction
func RollbackTransaction(tx *gorm.DB) {
	tx.Rollback()
}
