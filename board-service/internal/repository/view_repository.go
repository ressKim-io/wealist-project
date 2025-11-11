package repository

import (
	"board-service/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ViewRepository는 SavedView 엔티티만 관리합니다
type ViewRepository interface {
	Create(view *domain.SavedView) error
	FindByID(id uuid.UUID) (*domain.SavedView, error)
	FindByProject(projectID uuid.UUID) ([]domain.SavedView, error)
	FindDefault(projectID uuid.UUID) (*domain.SavedView, error)
	Update(view *domain.SavedView) error
	Delete(id uuid.UUID) error
}

type viewRepository struct {
	db *gorm.DB
}

// NewViewRepository는 새로운 ViewRepository를 생성합니다
func NewViewRepository(db *gorm.DB) ViewRepository {
	return &viewRepository{db: db}
}

func (r *viewRepository) Create(view *domain.SavedView) error {
	return r.db.Create(view).Error
}

func (r *viewRepository) FindByID(id uuid.UUID) (*domain.SavedView, error) {
	var view domain.SavedView
	if err := r.db.Where("id = ? AND is_deleted = ?", id, false).First(&view).Error; err != nil {
		return nil, err
	}
	return &view, nil
}

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

func (r *viewRepository) Update(view *domain.SavedView) error {
	return r.db.Save(view).Error
}

func (r *viewRepository) Delete(id uuid.UUID) error {
	return r.db.Model(&domain.SavedView{}).
		Where("id = ?", id).
		Update("is_deleted", true).Error
}
