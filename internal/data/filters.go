package data

import (
	"math"
	"strings"

	"github.com/kirwadee/appletree/internal/validator"
)

type Filters struct {
	Page     int
	PageSize int
	Sort     string
	SortList []string
}

func ValidateFilters(v *validator.Validator, f Filters) {
	//check page and page_size parameters
	v.Check(f.Page > 0, "page", "must be greater than 0")
	v.Check(f.Page <= 1000, "page", "must be a maximum of 1000")
	v.Check(f.PageSize > 0, "page_size", "must be greater than 0")
	v.Check(f.Page <= 100, "page_size", "must be a maximum of 100")
	//check that the sort parameter matches a value in the acceptable sort list
	v.Check(validator.In(f.Sort, f.SortList...), "sort", "invalid sort value")
}

// The sortColumn() method safely extract the sort field query parameter
func (f Filters) sortColumn() string {
	for _, safeValue := range f.SortList {
		if f.Sort == safeValue {
			return strings.TrimPrefix(f.Sort, "-")
		}
	}
	panic("unsafe sort parameter " + f.Sort)
}

// sortOrder determones whether we should sort by DES or ASC
func (f Filters) sortOrder() string {
	if strings.HasPrefix(f.Sort, "-") {
		return "DESC"
	}
	return "ASC"
}

// limit() method determines the LIMIT
func (f Filters) limit() int {
	return f.PageSize
}

// The offset() method calculates the OFFSET
func (f Filters) offset() int {
	return (f.Page - 1) * f.PageSize
}

// Metadata type contains metadata to help with pagination
type Metadata struct {
	CurrentPage  int `json:"current_page,omitempty"`
	PageSize     int `json:"page_size,omitempty"`
	FirstPage    int `json:"first_page,omitempty"`
	LastPage     int `json:"last_page,omitempty"`
	TotalRecords int `json:"total_records,omitempty"`
}

// The calculateMetadata() computes the values for the Metadata fields
func calculateMetadata(totalRecords int, page int, pageSize int) Metadata {
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
