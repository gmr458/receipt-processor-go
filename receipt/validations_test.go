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
