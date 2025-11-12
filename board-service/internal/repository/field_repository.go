package repository

import (
	"board-service/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// FieldRepository는 기존 인터페이스 (5개 엔티티 통합 관리)
// 호환성을 위해 유지하지만, 내부적으로는 분리된 repository들을 사용합니다
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
	UpdateBoardFieldCache(boardID uuid.UUID) (string, error)

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

// fieldRepository는 기존 인터페이스 호환성을 위한 어댑터입니다
// 내부적으로는 분리된 5개의 repository를 사용합니다 (어댑터 패턴)
type fieldRepository struct {
	projectField ProjectFieldRepository
	option       FieldOptionRepository
	value        FieldValueRepository
	view         ViewRepository
	boardOrder   BoardOrderRepository
}

// NewFieldRepository는 새로운 분리된 repository들을 사용하는 어댑터를 생성합니다
func NewFieldRepository(db *gorm.DB) FieldRepository {
	return &fieldRepository{
		projectField: NewProjectFieldRepository(db),
		option:       NewFieldOptionRepository(db),
		value:        NewFieldValueRepository(db),
		view:         NewViewRepository(db),
		boardOrder:   NewBoardOrderRepository(db),
	}
}

// ==================== Project Field Implementation ====================
// 내부적으로 ProjectFieldRepository 위임

func (r *fieldRepository) CreateField(field *domain.ProjectField) error {
	return r.projectField.Create(field)
}

func (r *fieldRepository) FindFieldByID(id uuid.UUID) (*domain.ProjectField, error) {
	return r.projectField.FindByID(id)
}

func (r *fieldRepository) FindFieldsByProject(projectID uuid.UUID) ([]domain.ProjectField, error) {
	return r.projectField.FindByProject(projectID)
}

func (r *fieldRepository) FindFieldsByIDs(ids []uuid.UUID) ([]domain.ProjectField, error) {
	return r.projectField.FindByIDs(ids)
}

func (r *fieldRepository) UpdateField(field *domain.ProjectField) error {
	return r.projectField.Update(field)
}

func (r *fieldRepository) DeleteField(id uuid.UUID) error {
	return r.projectField.Delete(id)
}

func (r *fieldRepository) UpdateFieldOrder(fieldID uuid.UUID, newOrder int) error {
	return r.projectField.UpdateOrder(fieldID, newOrder)
}

func (r *fieldRepository) BatchUpdateFieldOrders(orders map[uuid.UUID]int) error {
	return r.projectField.BatchUpdateOrders(orders)
}

// ==================== Field Option Implementation ====================
// 내부적으로 FieldOptionRepository 위임

func (r *fieldRepository) CreateOption(option *domain.FieldOption) error {
	return r.option.Create(option)
}

func (r *fieldRepository) FindOptionByID(id uuid.UUID) (*domain.FieldOption, error) {
	return r.option.FindByID(id)
}

func (r *fieldRepository) FindOptionsByField(fieldID uuid.UUID) ([]domain.FieldOption, error) {
	return r.option.FindByField(fieldID)
}

func (r *fieldRepository) FindOptionsByIDs(ids []uuid.UUID) ([]domain.FieldOption, error) {
	return r.option.FindByIDs(ids)
}

func (r *fieldRepository) UpdateOption(option *domain.FieldOption) error {
	return r.option.Update(option)
}

func (r *fieldRepository) DeleteOption(id uuid.UUID) error {
	return r.option.Delete(id)
}

func (r *fieldRepository) UpdateOptionOrder(optionID uuid.UUID, newOrder int) error {
	return r.option.UpdateOrder(optionID, newOrder)
}

func (r *fieldRepository) BatchUpdateOptionOrders(orders map[uuid.UUID]int) error {
	return r.option.BatchUpdateOrders(orders)
}

// ==================== Board Field Value Implementation ====================
// 내부적으로 FieldValueRepository 위임

func (r *fieldRepository) SetFieldValue(value *domain.BoardFieldValue) error {
	return r.value.Set(value)
}

func (r *fieldRepository) FindFieldValuesByBoard(boardID uuid.UUID) ([]domain.BoardFieldValue, error) {
	return r.value.FindByBoard(boardID)
}

func (r *fieldRepository) FindFieldValuesByBoardAndField(boardID, fieldID uuid.UUID) ([]domain.BoardFieldValue, error) {
	return r.value.FindByBoardAndField(boardID, fieldID)
}

func (r *fieldRepository) FindFieldValuesByBoards(boardIDs []uuid.UUID) (map[uuid.UUID][]domain.BoardFieldValue, error) {
	return r.value.FindByBoards(boardIDs)
}

func (r *fieldRepository) DeleteFieldValue(boardID, fieldID uuid.UUID) error {
	return r.value.Delete(boardID, fieldID)
}

func (r *fieldRepository) DeleteFieldValueByID(id uuid.UUID) error {
	return r.value.DeleteByID(id)
}

func (r *fieldRepository) BatchSetFieldValues(values []domain.BoardFieldValue) error {
	return r.value.BatchSet(values)
}

func (r *fieldRepository) BatchDeleteFieldValues(boardID, fieldID uuid.UUID) error {
	return r.value.BatchDelete(boardID, fieldID)
}

func (r *fieldRepository) UpdateBoardFieldCache(boardID uuid.UUID) (string, error) {
	return r.value.UpdateBoardCache(boardID)
}

// ==================== Saved View Implementation ====================
// 내부적으로 ViewRepository 위임

func (r *fieldRepository) CreateView(view *domain.SavedView) error {
	return r.view.Create(view)
}

func (r *fieldRepository) FindViewByID(id uuid.UUID) (*domain.SavedView, error) {
	return r.view.FindByID(id)
}

func (r *fieldRepository) FindViewsByProject(projectID uuid.UUID) ([]domain.SavedView, error) {
	return r.view.FindByProject(projectID)
}

func (r *fieldRepository) FindDefaultView(projectID uuid.UUID) (*domain.SavedView, error) {
	return r.view.FindDefault(projectID)
}

func (r *fieldRepository) UpdateView(view *domain.SavedView) error {
	return r.view.Update(view)
}

func (r *fieldRepository) DeleteView(id uuid.UUID) error {
	return r.view.Delete(id)
}

// ==================== User Board Order Implementation ====================
// 내부적으로 BoardOrderRepository 위임

func (r *fieldRepository) SetBoardOrder(order *domain.UserBoardOrder) error {
	return r.boardOrder.Set(order)
}

func (r *fieldRepository) FindBoardOrdersByView(viewID, userID uuid.UUID) ([]domain.UserBoardOrder, error) {
	return r.boardOrder.FindByView(viewID, userID)
}

func (r *fieldRepository) BatchUpdateBoardOrders(orders []domain.UserBoardOrder) error {
	return r.boardOrder.BatchUpdate(orders)
}

func (r *fieldRepository) DeleteBoardOrder(viewID, userID, boardID uuid.UUID) error {
	return r.boardOrder.Delete(viewID, userID, boardID)
}
