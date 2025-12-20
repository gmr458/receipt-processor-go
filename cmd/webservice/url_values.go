package main

import (
	"net/url"
	"slices"
	"strconv"
)

// getURLValueStr returns the URL parameter value if it's in safeValues,
// otherwise returns defaultValue. Include "" in safeValues to allow empty
// strings.
func getURLValueStr(
	values url.Values,
	safeValues []string,
	key, fallback string,
) string {
	s := values.Get(key)
	if !slices.Contains(safeValues, s) {
		return fallback
	}
	return s
}

// getURLValuePositiveInt returns the URL parameter as a positive integer.
// Returns fallback if the parameter is missing, not a valid integer, or <= 0.
func getURLValuePositiveInt(
	values url.Values,
	key string,
	fallback int,
) int {
	s := values.Get(key)
	if s == "" {
		return fallback
	}

	value, err := strconv.Atoi(s)
	if err != nil || value <= 0 {
		return fallback
	}

	return value
}
