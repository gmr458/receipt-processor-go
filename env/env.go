package env

import (
	"log"
	"os"
	"strconv"
)

func Getenv[T int | float64 | string | bool](key string) T {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("%s environment variable is empty", key)
	}

	var result T
	switch any(result).(type) {
	case int:
		v, err := strconv.Atoi(value)
		if err != nil {
			log.Fatalf("%s: invalid int value %q", key, value)
		}
		result = any(v).(T)

	case float64:
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			log.Fatalf("%s: invalid float64 value %q", key, value)
		}
		result = any(v).(T)

	case bool:
		v, err := strconv.ParseBool(value)
		if err != nil {
			log.Fatalf("%s: invalid bool value %q", key, value)
		}
		result = any(v).(T)

	case string:
		result = any(value).(T)
	}

	return result
}

func GetenvOrDefault[T int | float64 | string | bool](key string, defaultVal T) T {
	value := os.Getenv(key)
	if value == "" {
		return defaultVal
	}

	switch any(defaultVal).(type) {
	case int:
		i, err := strconv.Atoi(value)
		if err != nil {
			log.Fatalf("%s environment variable must be an integer", key)
		}
		return any(i).(T)

	case float64:
		f, err := strconv.ParseFloat(value, 64)
		if err != nil {
			log.Fatalf("%s environment variable must be a floating-point number", key)
		}
		return any(f).(T)

	case string:
		return any(value).(T)

	case bool:
		b, err := strconv.ParseBool(value)
		if err != nil {
			log.Fatalf("%s environment variable must be a boolean", key)
		}
		return any(b).(T)

	default:
		return defaultVal
	}
}
