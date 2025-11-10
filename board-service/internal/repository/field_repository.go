package repository

import (
	"board-service/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type FieldRepository interface {
	// ==================== Project Field Methods ====================
	CreateField(field *domain.ProjectField) error
	FindFieldByID(id uuid.UUID) (*domain.ProjectField, error)
	FindFieldsByProject(projectID uuid.UUID) ([]domain.ProjectField, error)
	FindFieldsByIDs(ids []uuid.UUID) ([]domain.ProjectField, error)
	UpdateField(field *domain.ProjectField) error
	DeleteField(id uuid.UUID) error
	UpdateFieldOrder(fieldID uuid.UUID, newOrder int) error
	BatchUpdateFieldOrders(orders map[uuid.UUID]int) error

	// ==================== Field Option Methods ====================
	CreateOption(option *domain.FieldOption) error
	FindOptionByID(id uuid.UUID) (*domain.FieldOption, error)
	FindOptionsByField(fieldID uuid.UUID) ([]domain.FieldOption, error)
	FindOptionsByIDs(ids []uuid.UUID) ([]domain.FieldOption, error)
	UpdateOption(option *domain.FieldOption) error
	DeleteOption(id uuid.UUID) error
	UpdateOptionOrder(optionID uuid.UUID, newOrder int) error
	BatchUpdateOptionOrders(orders map[uuid.UUID]int) error

	// ==================== Board Field Value Methods ====================
	SetFieldValue(value *domain.BoardFieldValue) error
	FindFieldValuesByBoard(boardID uuid.UUID) ([]domain.BoardFieldValue, error)
	FindFieldValuesByBoardAndField(boardID, fieldID uuid.UUID) ([]domain.BoardFieldValue, error)
	FindFieldValuesByBoards(boardIDs []uuid.UUID) (map[uuid.UUID][]domain.BoardFieldValue, error)
	DeleteFieldValue(boardID, fieldID uuid.UUID) error
	DeleteFieldValueByID(id uuid.UUID) error
	BatchSetFieldValues(values []domain.BoardFieldValue) error
	BatchDeleteFieldValues(boardID, fieldID uuid.UUID) error

	// Cache update
	UpdateBoardFieldCache(boardID uuid.UUID) (string, error) // Returns JSON string

	// ==================== Saved View Methods ====================
	CreateView(view *domain.SavedView) error
	FindViewByID(id uuid.UUID) (*domain.SavedView, error)
	FindViewsByProject(projectID uuid.UUID) ([]domain.SavedView, error)
	FindDefaultView(projectID uuid.UUID) (*domain.SavedView, error)
	UpdateView(view *domain.SavedView) error
	DeleteView(id uuid.UUID) error

	// ==================== User Board Order Methods ====================
	SetBoardOrder(order *domain.UserBoardOrder) error
	FindBoardOrdersByView(viewID, userID uuid.UUID) ([]domain.UserBoardOrder, error)
	BatchUpdateBoardOrders(orders []domain.UserBoardOrder) error
	DeleteBoardOrder(viewID, userID, boardID uuid.UUID) error
}

type fieldRepository struct {
	db *gorm.DB
}

func NewFieldRepository(db *gorm.DB) FieldRepository {
	return &fieldRepository{db: db}
}

// ==================== Project Field Implementation ====================

func (r *fieldRepository) CreateField(field *domain.ProjectField) error {
	return r.db.Create(field).Error
}

func (r *fieldRepository) FindFieldByID(id uuid.UUID) (*domain.ProjectField, error) {
	var field domain.ProjectField
	if err := r.db.Where("id = ? AND is_deleted = ?", id, false).First(&field).Error; err != nil {
		return nil, err
	}
	return &field, nil
}

func (r *fieldRepository) FindFieldsByProject(projectID uuid.UUID) ([]domain.ProjectField, error) {
	var fields []domain.ProjectField
	if err := r.db.Where("project_id = ? AND is_deleted = ?", projectID, false).
		Order("display_order ASC, created_at ASC").
		Find(&fields).Error; err != nil {
		return nil, err
	}
	return fields, nil
}

func (r *fieldRepository) FindFieldsByIDs(ids []uuid.UUID) ([]domain.ProjectField, error) {
	var fields []domain.ProjectField
	if err := r.db.Where("id IN ? AND is_deleted = ?", ids, false).
		Order("display_order ASC").
		Find(&fields).Error; err != nil {
		return nil, err
	}
	return fields, nil
}

func (r *fieldRepository) UpdateField(field *domain.ProjectField) error {
	return r.db.Save(field).Error
}

func (r *fieldRepository) DeleteField(id uuid.UUID) error {
	return r.db.Model(&domain.ProjectField{}).
		Where("id = ?", id).
		Update("is_deleted", true).Error
}

func (r *fieldRepository) UpdateFieldOrder(fieldID uuid.UUID, newOrder int) error {
	return r.db.Model(&domain.ProjectField{}).
		Where("id = ?", fieldID).
		Update("display_order", newOrder).Error
}

func (r *fieldRepository) BatchUpdateFieldOrders(orders map[uuid.UUID]int) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for fieldID, order := range orders {
			if err := tx.Model(&domain.ProjectField{}).
				Where("id = ?", fieldID).
				Update("display_order", order).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// ==================== Field Option Implementation ====================

func (r *fieldRepository) CreateOption(option *domain.FieldOption) error {
	return r.db.Create(option).Error
}

func (r *fieldRepository) FindOptionByID(id uuid.UUID) (*domain.FieldOption, error) {
	var option domain.FieldOption
	if err := r.db.Where("id = ? AND is_deleted = ?", id, false).First(&option).Error; err != nil {
		return nil, err
	}
	return &option, nil
}

func (r *fieldRepository) FindOptionsByField(fieldID uuid.UUID) ([]domain.FieldOption, error) {
	var options []domain.FieldOption
	if err := r.db.Where("field_id = ? AND is_deleted = ?", fieldID, false).
		Order("display_order ASC, created_at ASC").
		Find(&options).Error; err != nil {
		return nil, err
	}
	return options, nil
}

func (r *fieldRepository) FindOptionsByIDs(ids []uuid.UUID) ([]domain.FieldOption, error) {
	var options []domain.FieldOption
	if err := r.db.Where("id IN ? AND is_deleted = ?", ids, false).
		Order("display_order ASC").
		Find(&options).Error; err != nil {
		return nil, err
	}
	return options, nil
}

func (r *fieldRepository) UpdateOption(option *domain.FieldOption) error {
	return r.db.Save(option).Error
}

func (r *fieldRepository) DeleteOption(id uuid.UUID) error {
	return r.db.Model(&domain.FieldOption{}).
		Where("id = ?", id).
		Update("is_deleted", true).Error
}

func (r *fieldRepository) UpdateOptionOrder(optionID uuid.UUID, newOrder int) error {
	return r.db.Model(&domain.FieldOption{}).
		Where("id = ?", optionID).
		Update("display_order", newOrder).Error
}

func (r *fieldRepository) BatchUpdateOptionOrders(orders map[uuid.UUID]int) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for optionID, order := range orders {
			if err := tx.Model(&domain.FieldOption{}).
				Where("id = ?", optionID).
				Update("display_order", order).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// ==================== Board Field Value Implementation ====================

func (r *fieldRepository) SetFieldValue(value *domain.BoardFieldValue) error {
	return r.db.Save(value).Error
}

func (r *fieldRepository) FindFieldValuesByBoard(boardID uuid.UUID) ([]domain.BoardFieldValue, error) {
	var values []domain.BoardFieldValue
	if err := r.db.Where("board_id = ? AND is_deleted = ?", boardID, false).
		Order("field_id, display_order ASC").
		Find(&values).Error; err != nil {
		return nil, err
	}
	return values, nil
}

func (r *fieldRepository) FindFieldValuesByBoardAndField(boardID, fieldID uuid.UUID) ([]domain.BoardFieldValue, error) {
	var values []domain.BoardFieldValue
	if err := r.db.Where("board_id = ? AND field_id = ? AND is_deleted = ?", boardID, fieldID, false).
		Order("display_order ASC").
		Find(&values).Error; err != nil {
		return nil, err
	}
	return values, nil
}

func (r *fieldRepository) FindFieldValuesByBoards(boardIDs []uuid.UUID) (map[uuid.UUID][]domain.BoardFieldValue, error) {
	var values []domain.BoardFieldValue
	if err := r.db.Where("board_id IN ? AND is_deleted = ?", boardIDs, false).
		Order("board_id, field_id, display_order ASC").
		Find(&values).Error; err != nil {
		return nil, err
	}

	// Group by board_id
	result := make(map[uuid.UUID][]domain.BoardFieldValue)
	for _, value := range values {
		result[value.BoardID] = append(result[value.BoardID], value)
	}
	return result, nil
}

func (r *fieldRepository) DeleteFieldValue(boardID, fieldID uuid.UUID) error {
	return r.db.Model(&domain.BoardFieldValue{}).
		Where("board_id = ? AND field_id = ?", boardID, fieldID).
		Update("is_deleted", true).Error
}

func (r *fieldRepository) DeleteFieldValueByID(id uuid.UUID) error {
	return r.db.Model(&domain.BoardFieldValue{}).
		Where("id = ?", id).
		Update("is_deleted", true).Error
}

func (r *fieldRepository) BatchSetFieldValues(values []domain.BoardFieldValue) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, value := range values {
			if err := tx.Save(&value).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *fieldRepository) BatchDeleteFieldValues(boardID, fieldID uuid.UUID) error {
	return r.db.Model(&domain.BoardFieldValue{}).
		Where("board_id = ? AND field_id = ?", boardID, fieldID).
		Update("is_deleted", true).Error
}

// UpdateBoardFieldCache generates and updates the custom_fields_cache JSON for a board
func (r *fieldRepository) UpdateBoardFieldCache(boardID uuid.UUID) (string, error) {
	// This will be implemented with proper JSON marshaling in service layer
	// For now, return empty object
	return "{}", nil
}

// ==================== Saved View Implementation ====================

func (r *fieldRepository) CreateView(view *domain.SavedView) error {
	return r.db.Create(view).Error
}

func (r *fieldRepository) FindViewByID(id uuid.UUID) (*domain.SavedView, error) {
	var view domain.SavedView
	if err := r.db.Where("id = ? AND is_deleted = ?", id, false).First(&view).Error; err != nil {
		return nil, err
	}
	return &view, nil
}

func (r *fieldRepository) FindViewsByProject(projectID uuid.UUID) ([]domain.SavedView, error) {
	var views []domain.SavedView
	if err := r.db.Where("project_id = ? AND is_deleted = ?", projectID, false).
		Order("is_default DESC, created_at ASC").
		Find(&views).Error; err != nil {
		return nil, err
	}
	return views, nil
}

func (r *fieldRepository) FindDefaultView(projectID uuid.UUID) (*domain.SavedView, error) {
	var view domain.SavedView
	if err := r.db.Where("project_id = ? AND is_default = ? AND is_deleted = ?", projectID, true, false).
		First(&view).Error; err != nil {
		return nil, err
	}
	return &view, nil
}

func (r *fieldRepository) UpdateView(view *domain.SavedView) error {
	return r.db.Save(view).Error
}

func (r *fieldRepository) DeleteView(id uuid.UUID) error {
	return r.db.Model(&domain.SavedView{}).
		Where("id = ?", id).
		Update("is_deleted", true).Error
}

// ==================== User Board Order Implementation ====================

func (r *fieldRepository) SetBoardOrder(order *domain.UserBoardOrder) error {
	// Use UPSERT to handle conflicts (PostgreSQL ON CONFLICT DO UPDATE)
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "view_id"}, {Name: "user_id"}, {Name: "board_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"position", "updated_at"}),
	}).Create(order).Error
}

func (r *fieldRepository) FindBoardOrdersByView(viewID, userID uuid.UUID) ([]domain.UserBoardOrder, error) {
	var orders []domain.UserBoardOrder
	if err := r.db.Where("view_id = ? AND user_id = ?", viewID, userID).
		Order("position ASC"). // Fractional indexing: lexicographic sort
		Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}

func (r *fieldRepository) BatchUpdateBoardOrders(orders []domain.UserBoardOrder) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, order := range orders {
			if err := tx.Save(&order).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *fieldRepository) DeleteBoardOrder(viewID, userID, boardID uuid.UUID) error {
	return r.db.Where("view_id = ? AND user_id = ? AND board_id = ?", viewID, userID, boardID).
		Delete(&domain.UserBoardOrder{}).Error
}
