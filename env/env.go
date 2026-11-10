// Package env reads typed values from environment variables. Functions
// return errors so the caller decides what to do with invalid values.
package env

import (
	"fmt"
	"os"
	"strconv"
)

// Getenv reads key from the environment and parses it as T. An empty value
// is an error.
func Getenv[T int | float64 | string | bool](key string) (T, error) {
	value := os.Getenv(key)
	if value == "" {
		var zero T
		return zero, fmt.Errorf("%s environment variable is empty", key)
	}

	return parseEnv[T](key, value)
}

// GetenvOrDefault reads key from the environment and parses it as T,
// returning defaultVal when key is unset.
func GetenvOrDefault[T int | float64 | string | bool](key string, defaultVal T) (T, error) {
	value := os.Getenv(key)
	if value == "" {
		return defaultVal, nil
	}

	return parseEnv[T](key, value)
}

func parseEnv[T int | float64 | string | bool](key, value string) (T, error) {
	var zero T

	switch any(zero).(type) {
	case int:
		v, err := strconv.Atoi(value)
		if err != nil {
			return zero, fmt.Errorf("%s: invalid int value %q", key, value)
		}
		return any(v).(T), nil

	case float64:
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return zero, fmt.Errorf("%s: invalid float64 value %q", key, value)
		}
		return any(v).(T), nil

	case bool:
		v, err := strconv.ParseBool(value)
		if err != nil {
			return zero, fmt.Errorf("%s: invalid bool value %q", key, value)
		}
		return any(v).(T), nil

	case string:
		return any(value).(T), nil
	}

	// Unreachable: T is constrained to the four types above.
	return zero, fmt.Errorf("%s: unsupported environment variable type", key)
}
