package data

import (
	"math"
	"strings"

	"greenlight.samedarslan28.net/internal/validator"
)

// Filters represents pagination and sorting options for listing resources.
// Used in query parameters to control page size, page number, and sorting order.
type Filters struct {
	// Page number of the results to fetch (starting from 1)
	Page int `json:"page" example:"1"`

	// Number of results per page
	PageSize int `json:"page_size" example:"20"`

	// Field to sort results by (e.g., "name" or "-created_at" for descending)
	Sort string `json:"sort" example:"-created_at"`

	// List of allowed sort fields (internal use; not passed by client)
	SortSafelist []string `json:"-" swaggerignore:"true"`
}

// Metadata provides information about the paginated response.
// Returned alongside paginated results to help the client navigate pages.
type Metadata struct {
	// Current page number
	CurrentPage int `json:"current_page,omitempty" example:"1"`

	// Number of results per page
	PageSize int `json:"page_size,omitempty" example:"20"`

	// First page number in the result set
	FirstPage int `json:"first_page,omitempty" example:"1"`

	// Last page number in the result set
	LastPage int `json:"last_page,omitempty" example:"5"`

	// Total number of records across all pages
	TotalRecords int `json:"total_records,omitempty" example:"100"`
}

func ValidateFilters(v *validator.Validator, f Filters) {
	v.Check(f.Page > 0, "page", "must be greater than zero")
	v.Check(f.Page <= 10_000_000, "page", "must be a maximum of 10 million")

	v.Check(f.PageSize > 0, "page_size", "must be greater than zero")
	v.Check(f.PageSize <= 100, "page_size", "must be a maximum of 100")

	v.Check(validator.In(f.Sort, f.SortSafelist...), "sort", "invalid sort value")
}

func (f Filters) sortColumn() string {
	for _, safeValue := range f.SortSafelist {
		if f.Sort == safeValue {
			return strings.TrimPrefix(f.Sort, "-")
		}
	}
	panic("unsafe sort parameter: " + f.Sort)
}

// Return the sort direction ("ASC" or "DESC") depending on the prefix character of the Sort field.
func (f Filters) sortDirection() string {
	if strings.HasPrefix(f.Sort, "-") {
		return "DESC"
	}
	return "ASC"
}
func (f Filters) limit() int {
	return f.PageSize
}
func (f Filters) offset() int {
	return (f.Page - 1) * f.PageSize
}

func calculateMetadata(totalRecords, page, pageSize int) Metadata {
	if totalRecords == 0 {
		return Metadata{}
	}
	return Metadata{
		CurrentPage:  page,
		PageSize:     pageSize,
		FirstPage:    1,
		LastPage:     int(math.Ceil(float64(totalRecords) / float64(pageSize))),
		TotalRecords: totalRecords,
	}
}
