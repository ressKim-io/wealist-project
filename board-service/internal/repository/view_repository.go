package repository

import (
	"board-service/internal/domain"
	"board-service/internal/repository/base"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ViewRepository는 SavedView 엔티티만 관리합니다
type ViewRepository interface {
	// 공통 CRUD 메서드 (base repository에서 제공)
	Create(view *domain.SavedView) error
	FindByID(id uuid.UUID) (*domain.SavedView, error)
	Update(view *domain.SavedView) error
	Delete(id uuid.UUID) error

	// SavedView 전용 메서드
	FindByProject(projectID uuid.UUID) ([]domain.SavedView, error)
	FindDefault(projectID uuid.UUID) (*domain.SavedView, error)
}

type viewRepository struct {
	base.BaseRepository[*domain.SavedView]
	db *gorm.DB
}

// NewViewRepository는 새로운 ViewRepository를 생성합니다
func NewViewRepository(db *gorm.DB) ViewRepository {
	return &viewRepository{
		BaseRepository: base.NewBaseRepository[*domain.SavedView](db),
		db:             db,
	}
}

// ==================== 공통 CRUD는 base repository에 위임 ====================
// Create, FindByID, Update, Delete는 BaseRepository의 구현을 사용합니다

// ==================== SavedView 전용 메서드 ====================

func (r *viewRepository) FindByProject(projectID uuid.UUID) ([]domain.SavedView, error) {
	var views []domain.SavedView
	if err := r.db.Where("project_id = ? AND is_deleted = ?", projectID, false).
		Order("is_default DESC, created_at ASC").
		Find(&views).Error; err != nil {
		return nil, err
	}
	return views, nil
}

func (r *viewRepository) FindDefault(projectID uuid.UUID) (*domain.SavedView, error) {
	var view domain.SavedView
	if err := r.db.Where("project_id = ? AND is_default = ? AND is_deleted = ?", projectID, true, false).
		First(&view).Error; err != nil {
		return nil, err
	}
	return &view, nil
}
