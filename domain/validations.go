package domain

import (
	"math"
	"unicode"
)

func isAlphanumeric(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r)
}

func hasZeroDecimal(f float64) bool {
	return math.Mod(f, 1.0) == 0
}

func xIsMultipleOfy(x, y float64) bool {
	return math.Mod(x, y) == 0.0
}

func isOdd(n int) bool {
	return n%2 != 0
}
