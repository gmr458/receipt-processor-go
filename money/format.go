package money

import (
	"fmt"
	"strconv"
)

// String formats m as a decimal with exactly two decimals, e.g. "12.34",
// "-0.05" or "0.00". It uses integer math only.
func (m Money) String() string {
	mag := abs64(int64(m))
	sign := ""
	if m < 0 {
		sign = "-"
	}
	return fmt.Sprintf("%s%d.%02d", sign, mag/100, mag%100)
}

// maxParseLen is the longest input Parse accepts. The longest canonical value
// is "-92233720368547758.08" (21 bytes); the rest is headroom for zero padding.
const maxParseLen = 32

// Parse parses a money amount in text form:
//
//	[-]DIGITS.DD
//
// Exactly two decimals are required and leading zeros are allowed ("007.50"
// parses as 7.50). Leading '+', exponents, thousands separators, whitespace,
// a missing integer part (".50") and negative zero ("-0.00") are rejected, as
// is any value outside the int64 cent range, and so is any input longer than
// 32 bytes. No floating point is involved.
func Parse(s string) (Money, error) {
	if len(s) > maxParseLen {
		return 0, fmt.Errorf("%w: input longer than %d bytes", ErrInvalidFormat, maxParseLen)
	}
	orig := s
	bad := func() (Money, error) {
		return 0, fmt.Errorf("%w: %s", ErrInvalidFormat, snippet(orig))
	}

	neg := false
	if len(s) > 0 && s[0] == '-' {
		neg = true
		s = s[1:]
	}
	// At least one integer digit, '.', two decimals.
	if len(s) < 4 || s[len(s)-3] != '.' {
		return bad()
	}
	intPart, frac := s[:len(s)-3], s[len(s)-2:]

	// Magnitude limit: MaxInt64 for positives, 2^63 for negatives.
	limit := uint64(1<<63 - 1)
	if neg {
		limit = 1 << 63
	}

	var mag uint64
	for _, digits := range [2]string{intPart, frac} {
		for i := 0; i < len(digits); i++ {
			c := digits[i]
			if c < '0' || c > '9' {
				return bad()
			}
			d := uint64(c - '0')
			if mag > (limit-d)/10 {
				return 0, fmt.Errorf("%w: %s", ErrOverflow, snippet(orig))
			}
			mag = mag*10 + d
		}
	}

	if neg {
		if mag == 0 { // reject "-0.00"
			return bad()
		}
		return Money(int64(-mag)), nil
	}
	return Money(int64(mag)), nil
}

// maxEchoRunes caps how much of a rejected input is echoed in error messages,
// so a huge malicious payload cannot be reflected back in full.
const maxEchoRunes = 32

// snippet returns s quoted for use in an error message, truncated to at most
// maxEchoRunes runes (with a trailing "..." marker when cut).
func snippet(s string) string {
	n := 0
	for i := range s {
		if n == maxEchoRunes {
			return strconv.Quote(s[:i]) + "..."
		}
		n++
	}
	return strconv.Quote(s)
}
