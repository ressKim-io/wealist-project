package uow

import (
	"board-service/internal/repository"
	"gorm.io/gorm"
)

// UnitOfWork는 여러 repository 작업을 하나의 트랜잭션으로 묶습니다
// Repository pattern과 함께 사용하여 트랜잭션 경계를 명확히 합니다
type UnitOfWork interface {
	// Do executes a function within a transaction
	// If the function returns an error, the transaction is rolled back
	// Otherwise, the transaction is committed
	Do(fn func(repos *Repositories) error) error

	// GetDB returns the underlying database connection
	GetDB() *gorm.DB
}

// Repositories는 트랜잭션 내에서 사용할 수 있는 모든 repository를 포함합니다
type Repositories struct {
	Board   repository.BoardRepository
	Project repository.ProjectRepository
	Comment repository.CommentRepository
	Field   repository.FieldRepository
	Role    repository.RoleRepository
}

type unitOfWork struct {
	db *gorm.DB
}

// NewUnitOfWork creates a new Unit of Work
func NewUnitOfWork(db *gorm.DB) UnitOfWork {
	return &unitOfWork{db: db}
}

// Do executes a function within a transaction
func (uow *unitOfWork) Do(fn func(repos *Repositories) error) error {
	return uow.db.Transaction(func(tx *gorm.DB) error {
		// Create repositories with the transaction database
		repos := &Repositories{
			Board:   repository.NewBoardRepository(tx),
			Project: repository.NewProjectRepository(tx),
			Comment: repository.NewCommentRepository(tx),
			Field:   repository.NewFieldRepository(tx),
			Role:    repository.NewRoleRepository(tx),
		}

		// Execute the business logic
		return fn(repos)
	})
}

// GetDB returns the underlying database connection
func (uow *unitOfWork) GetDB() *gorm.DB {
	return uow.db
}
