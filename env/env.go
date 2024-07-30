package env

import (
	"log"
	"os"
	"strconv"
)

func Getenv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("%s environment variable is empty", key)
	}

	return value
}

func GetenvInt(key string) int {
	value := Getenv(key)

	i, err := strconv.Atoi(value)
	if err != nil {
		log.Fatalf("%s environment variable must be an integer", key)
	}

	return i
}

func GetenvFloat(key string) float64 {
	value := Getenv(key)

	f, err := strconv.ParseFloat(value, 64)
	if err != nil {
		log.Fatalf("%s environment variable must be a floating-point number", key)
	}

	return f
}

func GetenvBool(key string) bool {
	value := Getenv(key)

	b, err := strconv.ParseBool(value)
	if err != nil {
		log.Fatalf("%s environment variable must be a boolean (1, t, T, TRUE, true, True, 0, f, F, FALSE, false, False)", key)
	}

	return b
}
