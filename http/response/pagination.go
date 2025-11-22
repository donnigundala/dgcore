package response

import (
	"math"
	"net/http"
)

// PaginatedResponse represents a paginated API response.
type PaginatedResponse struct {
	Success bool           `json:"success"`
	Data    interface{}    `json:"data"`
	Meta    PaginationMeta `json:"meta"`
}

// PaginationMeta contains pagination metadata.
type PaginationMeta struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// Paginated writes a paginated JSON response.
func Paginated(w http.ResponseWriter, data interface{}, meta PaginationMeta) {
	// Calculate total pages if not provided
	if meta.TotalPages == 0 && meta.PerPage > 0 {
		meta.TotalPages = int(math.Ceil(float64(meta.Total) / float64(meta.PerPage)))
	}

	JSON(w, http.StatusOK, PaginatedResponse{
		Success: true,
		Data:    data,
		Meta:    meta,
	})
}

// NewPaginationMeta creates pagination metadata from request parameters.
func NewPaginationMeta(page, perPage, total int) PaginationMeta {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20 // Default per page
	}

	totalPages := int(math.Ceil(float64(total) / float64(perPage)))

	return PaginationMeta{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
	}
}

// CalculateOffset calculates the database offset from page and perPage.
func CalculateOffset(page, perPage int) int {
	if page < 1 {
		page = 1
	}
	return (page - 1) * perPage
}
