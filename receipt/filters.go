package receipt

import (
	"fmt"
	"math"
	"slices"
	"strings"

	"github.com/gmr458/receipt-processor-go/validator"
)

const (
	MAX_PAGE  = 10_000_000
	MAX_LIMIT = 100
)

type Filters struct {
	Page         int
	Limit        int
	Sort         string
	SortSafeList []string
}

func NewFilters(sortSafeList ...string) Filters {
	return Filters{
		SortSafeList: sortSafeList,
	}
}

func (f Filters) IsValid() (bool, map[string]string) {
	v := validator.New()

	v.Check(f.Page > 0, "page", "must be greater than zero")
	v.Check(f.Page <= MAX_PAGE, "page", "must be a maximum of 10 million")
	v.Check(f.Limit > 0, "limit", "must be greater than zero")
	v.Check(f.Limit <= MAX_LIMIT, "limit", "must be a maximum of 100")
	v.Check(slices.Contains(f.SortSafeList, f.Sort), "sort", "invalid sort value")

	return v.Ok(), v.Errors
}

// SortColumn returns the sort value with its "-" direction prefix stripped.
// IsValid must be called first; a value outside the safe list is a
// precondition failure, reported as an error.
func (f Filters) SortColumn() (string, error) {
	if slices.Contains(f.SortSafeList, f.Sort) {
		return strings.TrimPrefix(f.Sort, "-"), nil
	}

	return "", fmt.Errorf("unsafe sort parameter: %s", f.Sort)
}

func (f Filters) SortDirection() string {
	if strings.HasPrefix(f.Sort, "-") {
		return "DESC"
	}

	return "ASC"
}

func (f Filters) Offset() int {
	return (f.Page - 1) * f.Limit
}

type Metadata struct {
	Page      int `json:"page"`
	Limit     int `json:"limit"`
	FirstPage int `json:"firstPage"`
	LastPage  int `json:"lastPage"`
	Total     int `json:"total"`
}

func CalculateMetadata(total, page, limit int) Metadata {
	if total == 0 {
		return Metadata{
			Page:      page,
			Limit:     limit,
			FirstPage: 1,
			LastPage:  0,
			Total:     0,
		}
	}

	return Metadata{
		Page:      page,
		Limit:     limit,
		FirstPage: 1,
		LastPage:  int(math.Ceil(float64(total) / float64(limit))),
		Total:     total,
	}
}
