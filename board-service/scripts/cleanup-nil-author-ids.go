package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Board represents the board table structure
type Board struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	ProjectID uuid.UUID `gorm:"type:uuid;not null"`
	AuthorID  uuid.UUID `gorm:"type:uuid;not null"`
	Title     string    `gorm:"type:varchar(255);not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (Board) TableName() string {
	return "boards"
}

// Project represents the project table structure
type Project struct {
	ID      uuid.UUID `gorm:"type:uuid;primaryKey"`
	OwnerID uuid.UUID `gorm:"type:uuid;not null"`
}

func (Project) TableName() string {
	return "projects"
}

var nilUUID = uuid.UUID{}

func main() {
	// Check for dry-run mode
	if len(os.Args) > 1 && (os.Args[1] == "--dry-run" || os.Args[1] == "-n") {
		fmt.Println("Running in DRY-RUN mode (no database connection required)")
		fmt.Println("This mode demonstrates the script functionality without connecting to a database.")
		runDryRunDemo()
		return
	}

	// Get database connection string from environment
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "")
	dbName := getEnv("DB_NAME", "project_board")

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName,
	)

	fmt.Printf("Connecting to database: %s@%s:%s/%s\n\n", dbUser, dbHost, dbPort, dbName)

	// Connect to database
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get database instance: %v", err)
	}
	defer sqlDB.Close()

	ctx := context.Background()

	// Find boards with nil UUID author_id
	fmt.Println("Searching for boards with nil UUID (00000000-0000-0000-0000-000000000000) author_id...")
	
	var boards []Board
	result := db.WithContext(ctx).Where("author_id = ?", nilUUID).Find(&boards)
	if result.Error != nil {
		log.Fatalf("Failed to query boards: %v", result.Error)
	}

	if len(boards) == 0 {
		fmt.Println("✓ No boards found with nil UUID author_id. Database is clean!")
		return
	}

	fmt.Printf("\n⚠️  Found %d board(s) with nil UUID author_id:\n\n", len(boards))
	
	// Display affected boards
	for i, board := range boards {
		fmt.Printf("%d. Board ID: %s\n", i+1, board.ID)
		fmt.Printf("   Project ID: %s\n", board.ProjectID)
		fmt.Printf("   Title: %s\n", board.Title)
		fmt.Printf("   Created At: %s\n", board.CreatedAt.Format(time.RFC3339))
		
		// Try to get project owner
		var project Project
		if err := db.WithContext(ctx).Where("id = ?", board.ProjectID).First(&project).Error; err == nil {
			fmt.Printf("   Project Owner ID: %s\n", project.OwnerID)
		}
		fmt.Println()
	}

	// Ask user what to do
	fmt.Println("Options:")
	fmt.Println("1. Update author_id to project owner_id (recommended)")
	fmt.Println("2. Delete these boards")
	fmt.Println("3. Exit without changes")
	fmt.Print("\nEnter your choice (1-3): ")

	var choice int
	_, err = fmt.Scanf("%d", &choice)
	if err != nil {
		log.Fatalf("Invalid input: %v", err)
	}

	switch choice {
	case 1:
		// Update author_id to project owner_id
		fmt.Println("\nUpdating boards with project owner_id...")
		
		updated := 0
		failed := 0
		
		for _, board := range boards {
			var project Project
			if err := db.WithContext(ctx).Where("id = ?", board.ProjectID).First(&project).Error; err != nil {
				fmt.Printf("✗ Failed to get project for board %s: %v\n", board.ID, err)
				failed++
				continue
			}

			// Update the board's author_id
			result := db.WithContext(ctx).Model(&Board{}).
				Where("id = ?", board.ID).
				Update("author_id", project.OwnerID)
			
			if result.Error != nil {
				fmt.Printf("✗ Failed to update board %s: %v\n", board.ID, result.Error)
				failed++
				continue
			}

			fmt.Printf("✓ Updated board %s: author_id set to %s\n", board.ID, project.OwnerID)
			updated++
		}

		fmt.Printf("\n✓ Successfully updated %d board(s)\n", updated)
		if failed > 0 {
			fmt.Printf("✗ Failed to update %d board(s)\n", failed)
		}

	case 2:
		// Delete boards
		fmt.Println("\nDeleting boards...")
		
		deleted := 0
		failed := 0
		
		for _, board := range boards {
			result := db.WithContext(ctx).Delete(&Board{}, board.ID)
			if result.Error != nil {
				fmt.Printf("✗ Failed to delete board %s: %v\n", board.ID, result.Error)
				failed++
				continue
			}

			fmt.Printf("✓ Deleted board %s\n", board.ID)
			deleted++
		}

		fmt.Printf("\n✓ Successfully deleted %d board(s)\n", deleted)
		if failed > 0 {
			fmt.Printf("✗ Failed to delete %d board(s)\n", failed)
		}

	case 3:
		fmt.Println("\nExiting without changes.")
		return

	default:
		fmt.Println("\nInvalid choice. Exiting without changes.")
		return
	}

	// Verify the fix
	fmt.Println("\nVerifying fix...")
	var remainingBoards []Board
	result = db.WithContext(ctx).Where("author_id = ?", nilUUID).Find(&remainingBoards)
	if result.Error != nil {
		log.Fatalf("Failed to verify: %v", result.Error)
	}

	if len(remainingBoards) == 0 {
		fmt.Println("✓ Verification successful: No boards with nil UUID author_id remain!")
	} else {
		fmt.Printf("⚠️  Warning: %d board(s) still have nil UUID author_id\n", len(remainingBoards))
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// runDryRunDemo demonstrates the script functionality without a database connection
func runDryRunDemo() {
	fmt.Println("=== 데모: nil UUID author_id를 가진 보드 검색 ===")
	
	// Simulate finding boards with nil UUID
	demoBoards := []struct {
		ID        string
		ProjectID string
		Title     string
		CreatedAt string
		OwnerID   string
	}{
		{
			ID:        "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
			ProjectID: "11111111-2222-3333-4444-555555555555",
			Title:     "테스트 보드 1",
			CreatedAt: "2024-01-15T10:30:00Z",
			OwnerID:   "99999999-8888-7777-6666-555555555555",
		},
		{
			ID:        "b2c3d4e5-f6a7-8901-bcde-f12345678901",
			ProjectID: "22222222-3333-4444-5555-666666666666",
			Title:     "테스트 보드 2",
			CreatedAt: "2024-01-16T14:20:00Z",
			OwnerID:   "88888888-7777-6666-5555-444444444444",
		},
	}

	fmt.Printf("⚠️  nil UUID author_id를 가진 보드 %d개 발견:\n\n", len(demoBoards))
	
	for i, board := range demoBoards {
		fmt.Printf("%d. 보드 ID: %s\n", i+1, board.ID)
		fmt.Printf("   프로젝트 ID: %s\n", board.ProjectID)
		fmt.Printf("   제목: %s\n", board.Title)
		fmt.Printf("   생성일: %s\n", board.CreatedAt)
		fmt.Printf("   프로젝트 소유자 ID: %s\n", board.OwnerID)
		fmt.Println()
	}

	fmt.Println("옵션:")
	fmt.Println("1. author_id를 프로젝트 owner_id로 업데이트 (권장)")
	fmt.Println("2. 이 보드들 삭제")
	fmt.Println("3. 변경 없이 종료")
	fmt.Println("\n[DRY-RUN 모드] 실제 데이터베이스에 연결하려면 이 플래그 없이 스크립트를 실행하세요.")
	fmt.Println("\n사용법:")
	fmt.Println("  ./scripts/cleanup-nil-author-ids.sh")
	fmt.Println("  또는")
	fmt.Println("  go run scripts/cleanup-nil-author-ids.go")
}
