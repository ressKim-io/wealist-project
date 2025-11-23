package pagination

import "math"

// PageRequest represents pagination parameters
type PageRequest struct {
	Page  int
	Limit int
}

// PageResponse represents pagination metadata
type PageResponse struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
}

const (
	// DefaultPage is the default page number
	DefaultPage = 1

	// DefaultLimit is the default page size
	DefaultLimit = 20

	// MaxLimit is the maximum allowed page size
	MaxLimit = 100
)

// NewPageRequest creates a PageRequest with defaults applied
func NewPageRequest(page, limit int) PageRequest {
	if page < 1 {
		page = DefaultPage
	}
	if limit < 1 {
		limit = DefaultLimit
	}
	if limit > MaxLimit {
		limit = MaxLimit
	}

	return PageRequest{
		Page:  page,
		Limit: limit,
	}
}

// Offset calculates the database offset from page and limit
func (p PageRequest) Offset() int {
	return (p.Page - 1) * p.Limit
}

// CalculateOffset is a standalone function for calculating offset
func CalculateOffset(page, limit int) int {
	if page < 1 {
		page = 1
	}
	return (page - 1) * limit
}

// NewPageResponse creates a PageResponse from pagination parameters and total count
func NewPageResponse(page, limit int, total int64) PageResponse {
	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	if totalPages == 0 {
		totalPages = 1
	}

	return PageResponse{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}
}

// PaginatedResult is a generic wrapper for paginated data
type PaginatedResult[T any] struct {
	Data       []T          `json:"data"`
	Pagination PageResponse `json:"pagination"`
}

// NewPaginatedResult creates a PaginatedResult with data and pagination metadata
func NewPaginatedResult[T any](data []T, page, limit int, total int64) PaginatedResult[T] {
	return PaginatedResult[T]{
		Data:       data,
		Pagination: NewPageResponse(page, limit, total),
	}
}

// ValidatePaginationParams validates and normalizes pagination parameters
// Returns normalized page and limit values
func ValidatePaginationParams(page, limit int) (int, int) {
	if page < 1 {
		page = DefaultPage
	}
	if limit < 1 {
		limit = DefaultLimit
	}
	if limit > MaxLimit {
		limit = MaxLimit
	}
	return page, limit
}
