package page

import (
	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/sort"
)

const (
	// DefaultPageNumber is the default page number for pagination (starts from 1).
	DefaultPageNumber int = 1
	// DefaultPageSize is the default page size for pagination.
	DefaultPageSize int = 15
	// MaxPageSize is the maximum allowed page size to prevent excessive data loading.
	MaxPageSize int = 1000
)

// Pageable represents pagination parameters for querying data.
type Pageable struct {
	api.In

	Page int              `json:"page"` // Page is the page number (1-based)
	Size int              `json:"size"` // Size is the number of items per page
	Sort []sort.OrderSpec `json:"sort"` // Sort is the sort field names
}

// Normalize normalizes the pageable parameters.
// It sets default values and enforces limits.
func (p *Pageable) Normalize() {
	if p.Page < 1 {
		p.Page = DefaultPageNumber
	}

	if p.Size < 1 {
		p.Size = DefaultPageSize
	}

	if p.Size > MaxPageSize {
		p.Size = MaxPageSize
	}
}

// Offset returns the zero-based offset for database queries.
func (p Pageable) Offset() int {
	return (p.Page - 1) * p.Size
}

// Page represents a paginated response with metadata and items.
type Page[T any] struct {
	Page  int   `json:"page"`  // Page is the current page number (1-based)
	Size  int   `json:"size"`  // Size is the number of items per page
	Total int64 `json:"total"` // Total is the total number of items across all pages
	Items []T   `json:"items"` // Items contains the data for the current page
}

// TotalPages returns the total number of pages based on the total count.
func (page Page[T]) TotalPages() int {
	if page.Size == 0 {
		return 0
	}

	return int((page.Total + int64(page.Size) - 1) / int64(page.Size))
}

// HasNext returns true if there are more pages after the current one.
func (page Page[T]) HasNext() bool {
	return page.Page < page.TotalPages()
}

// HasPrevious returns true if there are pages before the current one.
func (page Page[T]) HasPrevious() bool {
	return page.Page > 1
}

// New creates a new page from pageable parameters, total count, and items.
// It ensures items is never nil and returns an empty slice if needed.
func New[T any](pageable Pageable, total int64, items []T) Page[T] {
	if items == nil {
		items = []T{}
	}

	return Page[T]{
		Page:  pageable.Page,
		Size:  pageable.Size,
		Total: total,
		Items: items,
	}
}
