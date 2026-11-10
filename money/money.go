// Package money provides a Money type that stores USD-style amounts as an
// exact integer number of cents. All arithmetic is integer arithmetic and
// every operation that can overflow returns an error instead of wrapping.
package money

import (
	"errors"
	"math"
	"math/bits"
	"strconv"
)

// Money is an amount of money in cents (1/100 of the major unit).
//
// Money is a plain int64, so it is comparable with ==, <, > etc. and can be
// used as a map key. The zero value is 0.00.
type Money int64

// Errors returned by this package. Use errors.Is to test for them.
var (
	ErrOverflow        = errors.New("money: arithmetic overflow")
	ErrDivideByZero    = errors.New("money: division by zero")
	ErrInvalidFormat   = errors.New("money: invalid format")
	ErrInvalidRounding = errors.New("money: invalid rounding mode")
	ErrInvalidRatio    = errors.New("money: invalid allocation ratios")
)

// Rounding selects how a fractional cent is resolved.
type Rounding int

const (
	// HalfEven rounds to the nearest cent; exact ties go to the even cent
	// (banker's rounding). Unbiased over many operations. This is the
	// default zero value.
	HalfEven Rounding = iota
	// HalfAwayFromZero rounds to the nearest cent; exact ties go to the
	// value with the larger magnitude (2.5 -> 3, -2.5 -> -3).
	HalfAwayFromZero
	// Down truncates toward zero (2.9 -> 2, -2.9 -> -2).
	Down
	// Up rounds away from zero whenever there is any remainder
	// (2.1 -> 3, -2.1 -> -3).
	Up
)

func (r Rounding) String() string {
	switch r {
	case HalfEven:
		return "HalfEven"
	case HalfAwayFromZero:
		return "HalfAwayFromZero"
	case Down:
		return "Down"
	case Up:
		return "Up"
	}
	return "Rounding(" + strconv.Itoa(int(r)) + ")"
}

// FromCents returns a Money holding the given number of cents.
func FromCents(c int64) Money { return Money(c) }

// Cents returns the amount as an integer number of cents.
func (m Money) Cents() int64 { return int64(m) }

// IsZero reports whether m is 0.00.
func (m Money) IsZero() bool { return m == 0 }

// Sign returns -1, 0 or +1 depending on the sign of m.
func (m Money) Sign() int {
	switch {
	case m < 0:
		return -1
	case m > 0:
		return 1
	}
	return 0
}

// Cmp returns -1, 0 or +1 if m is less than, equal to or greater than o.
func (m Money) Cmp(o Money) int {
	switch {
	case m < o:
		return -1
	case m > o:
		return 1
	}
	return 0
}

// Add returns m + o, or ErrOverflow.
func (m Money) Add(o Money) (Money, error) {
	s := m + o
	if (m > 0 && o > 0 && s < 0) || (m < 0 && o < 0 && s >= 0) {
		return 0, ErrOverflow
	}
	return s, nil
}

// Sub returns m - o, or ErrOverflow.
func (m Money) Sub(o Money) (Money, error) {
	d := m - o
	if (m >= 0 && o < 0 && d < 0) || (m < 0 && o > 0 && d >= 0) {
		return 0, ErrOverflow
	}
	return d, nil
}

// Neg returns -m, or ErrOverflow for the minimum representable value.
func (m Money) Neg() (Money, error) {
	if m == math.MinInt64 {
		return 0, ErrOverflow
	}
	return -m, nil
}

// Abs returns |m|, or ErrOverflow for the minimum representable value.
func (m Money) Abs() (Money, error) {
	if m < 0 {
		return m.Neg()
	}
	return m, nil
}

// Sum adds all values, returning ErrOverflow if any step overflows.
func Sum(ms ...Money) (Money, error) {
	var total Money
	for _, m := range ms {
		var err error
		if total, err = total.Add(m); err != nil {
			return 0, err
		}
	}
	return total, nil
}

// MulInt returns m * n exactly (e.g. price * quantity), or ErrOverflow.
func (m Money) MulInt(n int64) (Money, error) {
	return m.MulRatio(n, 1, Down) // denominator 1 => never rounds
}

// Div returns m / n rounded to a whole cent using mode.
// It returns ErrDivideByZero if n is 0.
func (m Money) Div(n int64, mode Rounding) (Money, error) {
	return m.MulRatio(1, n, mode)
}

// MulBasisPoints returns m * bp / 10000 rounded using mode.
// 1 basis point = 0.01%, so 725 bp = 7.25%.
func (m Money) MulBasisPoints(bp int64, mode Rounding) (Money, error) {
	return m.MulRatio(bp, 10_000, mode)
}

// MulRatio returns m * num / den rounded to a whole cent using mode.
//
// The intermediate product is computed in 128 bits and there is exactly one
// rounding step at the end, so the result is the correctly rounded value of
// the exact rational number. It returns ErrDivideByZero if den is 0 and
// ErrOverflow if the result does not fit in a Money.
func (m Money) MulRatio(num, den int64, mode Rounding) (Money, error) {
	if den == 0 {
		return 0, ErrDivideByZero
	}
	if mode < HalfEven || mode > Up {
		return 0, ErrInvalidRounding
	}

	neg := (m < 0) != (num < 0) != (den < 0)
	if m == 0 || num == 0 {
		neg = false
	}

	hi, lo := bits.Mul64(abs64(int64(m)), abs64(num))
	d := abs64(den)
	if hi >= d { // quotient would not fit in 64 bits (and Div64 would panic)
		return 0, ErrOverflow
	}
	q, r := bits.Div64(hi, lo, d)

	if r != 0 && roundAway(mode, q, r, d) {
		q++
		if q == 0 {
			return 0, ErrOverflow
		}
	}

	if neg {
		if q > 1<<63 {
			return 0, ErrOverflow
		}
		return Money(int64(-q)), nil
	}
	if q > math.MaxInt64 {
		return 0, ErrOverflow
	}
	return Money(int64(q)), nil
}

// roundAway reports whether the magnitude q (with non-zero remainder r out
// of divisor d) must be incremented by one.
func roundAway(mode Rounding, q, r, d uint64) bool {
	switch mode {
	case Down:
		return false
	case Up:
		return true
	}
	// r < d <= 2^63, so 2*r cannot overflow uint64.
	twice := r * 2
	switch {
	case twice > d:
		return true
	case twice < d:
		return false
	}
	// Exact tie.
	if mode == HalfAwayFromZero {
		return true
	}
	return q&1 == 1 // HalfEven: round to even
}

// abs64 returns |x| as a uint64; correct for math.MinInt64.
func abs64(x int64) uint64 {
	if x < 0 {
		return uint64(-x)
	}
	return uint64(x)
}

// Split divides m into n parts that add up exactly to m. Leftover cents are
// handed out one at a time to the first parts, so 1.00 split 3 ways yields
// 0.34, 0.33, 0.33.
func (m Money) Split(n int) ([]Money, error) {
	if n <= 0 {
		return nil, ErrInvalidRatio
	}
	ratios := make([]int64, n)
	for i := range ratios {
		ratios[i] = 1
	}
	return m.Allocate(ratios)
}

// Allocate divides m proportionally to ratios (e.g. {70, 20, 10}). The parts
// always add up exactly to m: leftover cents go one at a time to the first
// parts with a non-zero ratio. Ratios must be non-negative and not all zero.
func (m Money) Allocate(ratios []int64) ([]Money, error) {
	if len(ratios) == 0 {
		return nil, ErrInvalidRatio
	}
	var sum int64
	for _, r := range ratios {
		if r < 0 {
			return nil, ErrInvalidRatio
		}
		if sum > math.MaxInt64-r {
			return nil, ErrOverflow
		}
		sum += r
	}
	if sum == 0 {
		return nil, ErrInvalidRatio
	}

	parts := make([]Money, len(ratios))
	var allocated Money
	for i, r := range ratios {
		share, err := m.MulRatio(r, sum, Down) // truncates toward zero
		if err != nil {
			return nil, err
		}
		parts[i] = share
		allocated += share // |allocated| <= |m|, cannot overflow
	}

	// remainder has the same sign as m (or is zero) and |remainder| < len(ratios).
	remainder := m - allocated
	step := Money(1)
	if remainder < 0 {
		step = -1
	}
	for i := 0; remainder != 0; i++ {
		if ratios[i%len(ratios)] == 0 {
			continue
		}
		parts[i%len(ratios)] += step
		remainder -= step
	}
	return parts, nil
}
