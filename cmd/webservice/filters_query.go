package main

import (
	"errors"
	"net/url"
	"slices"
	"strconv"

	"github.com/gmr458/receipt-processor-go/errs"
	"github.com/gmr458/receipt-processor-go/receipt"
	"github.com/gmr458/receipt-processor-go/validator"
)

const (
	defaultPage  = 1
	defaultLimit = 10
)

var errInvalidQueryParam = errors.New("invalid query parameter value")

// NewFiltersFromURL builds receipt.Filters from the request's query values.
// defaultSort and sortSafeList are entity policy supplied by the caller
// (a leading "-" requests descending order). Missing parameters fall back
// to the defaults; invalid values produce one EINVALID error whose Details
// name each bad field.
func NewFiltersFromURL(values url.Values, defaultSort string, sortSafeList ...string) (receipt.Filters, error) {
	f := receipt.NewFilters(sortSafeList...)
	v := validator.New()

	page, err := queryPositiveInt(values, "page", defaultPage)
	if err == nil {
		f.Page = page
	}
	v.Check(err == nil, "page", "must be an integer greater than zero")

	limit, err := queryPositiveInt(values, "limit", defaultLimit)
	if err == nil {
		f.Limit = limit
	}
	v.Check(err == nil, "limit", "must be an integer greater than zero")

	sort, err := querySafeValue(values, sortSafeList, "sort", defaultSort)
	if err == nil {
		f.Sort = sort
	}
	v.Check(err == nil, "sort", "invalid sort value")

	if !v.Ok() {
		return f, &errs.Error{
			Code:    errs.EINVALID,
			Message: "Invalid query parameter(s)",
			Details: v.Errors,
		}
	}
	return f, nil
}

// queryPositiveInt reads a positive integer query parameter. An absent
// parameter returns the fallback; a present-but-unparseable or non-positive
// value is an error.
func queryPositiveInt(values url.Values, key string, fallback int) (int, error) {
	s, ok := values[key]
	if !ok {
		return fallback, nil
	}
	if len(s) != 1 {
		return 0, errInvalidQueryParam
	}

	value, err := strconv.Atoi(s[0])
	if err != nil || value <= 0 {
		return 0, errInvalidQueryParam
	}

	return value, nil
}

// querySafeValue reads a query parameter that must be one of the safe
// values. An absent parameter returns the fallback; a present-but-unsafe
// value is an error.
func querySafeValue(values url.Values, safe []string, key, fallback string) (string, error) {
	s, ok := values[key]
	if !ok {
		return fallback, nil
	}
	if len(s) == 1 && slices.Contains(safe, s[0]) {
		return s[0], nil
	}

	return "", errInvalidQueryParam
}
