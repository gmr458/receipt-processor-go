package receipt

import (
	"testing"
)

func TestIsAlphanumeric(t *testing.T) {
	tests := []struct {
		input rune
		want  bool
	}{
		{'a', true},
		{'Z', true},
		{'0', true},
		{'9', true},
		{' ', false},
		{'!', false},
		{'@', false},
		{'-', false},
		{'\n', false},
	}

	for _, tt := range tests {
		got := isAlphanumeric(tt.input)
		if got != tt.want {
			t.Errorf("isAlphanumeric(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestHasZeroDecimal(t *testing.T) {
	tests := []struct {
		input float64
		want  bool
	}{
		{0.0, true},
		{1.0, true},
		{100.0, true},
		{0.5, false},
		{1.99, false},
		{3.14159, false},
	}

	for _, tt := range tests {
		got := hasZeroDecimal(tt.input)
		if got != tt.want {
			t.Errorf("hasZeroDecimal(%f) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestXIsMultipleOfY(t *testing.T) {
	tests := []struct {
		x    float64
		y    float64
		want bool
	}{
		{10.0, 5.0, true},
		{9.0, 3.0, true},
		{0.25, 0.25, true},
		{1.0, 0.25, true},
		{1.0, 0.25, true},
		{7.0, 3.0, false},
		{1.0, 0.30, false},
		{0.0, 1.0, true},
	}

	for _, tt := range tests {
		got := xIsMultipleOfy(tt.x, tt.y)
		if got != tt.want {
			t.Errorf("xIsMultipleOfy(%f, %f) = %v, want %v", tt.x, tt.y, got, tt.want)
		}
	}
}

func TestIsOdd(t *testing.T) {
	tests := []struct {
		input int
		want  bool
	}{
		{1, true},
		{3, true},
		{101, true},
		{0, false},
		{2, false},
		{100, false},
		{-1, true},
		{-2, false},
	}

	for _, tt := range tests {
		got := isOdd(tt.input)
		if got != tt.want {
			t.Errorf("isOdd(%d) = %v, want %v", tt.input, got, tt.want)
		}
	}
}
