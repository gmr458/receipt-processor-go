package domain

import (
	"math"
	"slices"
	"strings"

	"github.com/gmr458/receipt-processor/validator"
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
	v.Check(f.Page <= 10_000_000, "page", "must be a maximum if 10 million")
	v.Check(f.Limit > 0, "limit", "must be greater than zero")
	v.Check(f.Limit <= 100, "limit", "must be a maximum if 100")
	v.Check(slices.Contains(f.SortSafeList, f.Sort), "sort", "invalid sort value")

	return v.Ok(), v.Errors
}

func (f Filters) SortColumn() string {
	if slices.Contains(f.SortSafeList, f.Sort) {
		return strings.TrimPrefix(f.Sort, "-")
	}

	panic("unsafe sort parameter: " + f.Sort)
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
	Page      int `json:"page,omitempty"`
	Limit     int `json:"limit,omitempty"`
	FirstPage int `json:"firstPage,omitempty"`
	LastPage  int `json:"lastPage,omitempty"`
	Total     int `json:"total,omitempty"`
}

func CalculateMetadata(total, page, limit int) Metadata {
	if total == 0 {
		return Metadata{}
	}

	return Metadata{
		Page:      page,
		Limit:     limit,
		FirstPage: 1,
		LastPage:  int(math.Ceil(float64(total) / float64(limit))),
		Total:     total,
	}
}
