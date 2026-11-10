package money

import (
	"database/sql"
	"encoding/json"
	"errors"
	"math"
	"math/big"
	"math/rand"
	"testing"
)

func TestStringParseRoundTrip(t *testing.T) {
	tests := []struct {
		cents Money
		str   string
	}{
		{0, "0.00"}, {1, "0.01"}, {5, "0.05"}, {-5, "-0.05"}, {99, "0.99"},
		{100, "1.00"}, {1234, "12.34"}, {-1234, "-12.34"},
		{math.MaxInt64, "92233720368547758.07"},
		{math.MinInt64, "-92233720368547758.08"},
	}
	for _, tc := range tests {
		if got := tc.cents.String(); got != tc.str {
			t.Errorf("String(%d) = %q, want %q", int64(tc.cents), got, tc.str)
		}
		got, err := Parse(tc.str)
		if err != nil || got != tc.cents {
			t.Errorf("Parse(%q) = %d, %v; want %d", tc.str, int64(got), err, int64(tc.cents))
		}
	}
}

func TestParseRejects(t *testing.T) {
	invalid := []string{
		"", "-", ".", "12", "12.", "12.3", "12.345", ".50", "-.50", "+1.00",
		"1e2", "1.0e1", "1,000.00", " 1.00", "1.00 ",
		"-0.00", "1.0a", "a.00", "--1.00", "1..00", "1.00.00", "١.٠٠",
	}
	for _, s := range invalid {
		if v, err := Parse(s); !errors.Is(err, ErrInvalidFormat) {
			t.Errorf("Parse(%q) = %d, %v; want ErrInvalidFormat", s, int64(v), err)
		}
	}
	for _, s := range []string{"92233720368547758.08", "-92233720368547758.09", "999999999999999999999.00"} {
		if _, err := Parse(s); !errors.Is(err, ErrOverflow) {
			t.Errorf("Parse(%q) err = %v; want ErrOverflow", s, err)
		}
	}
}

func TestParseLeadingZeros(t *testing.T) {
	tests := map[string]Money{"01.00": 100, "00.00": 0, "007.50": 750, "-007.50": -750, "0.05": 5, "000000000000000000000001.00": 100}
	for in, want := range tests {
		got, err := Parse(in)
		if err != nil || got != want {
			t.Errorf("Parse(%q) = %d, %v; want %d", in, int64(got), err, int64(want))
		}
	}
	// Output stays canonical.
	if got, _ := Parse("007.50"); got.String() != "7.50" {
		t.Errorf("String() = %q", got.String())
	}
}

// zeros returns an n-byte input of n-4 leading zeros followed by "1.00".
func zeros(n int) string {
	b := make([]byte, n-4)
	for i := range b {
		b[i] = '0'
	}
	return string(b) + "1.00"
}

func TestParseMaxLength(t *testing.T) {
	if got, err := Parse(zeros(32)); err != nil || got != 100 {
		t.Errorf("32 bytes: %d, %v", int64(got), err)
	}
	if _, err := Parse(zeros(33)); !errors.Is(err, ErrInvalidFormat) {
		t.Errorf("33 bytes: err = %v", err)
	}
	// The longest legitimate values are far below the limit.
	if len("-92233720368547758.08") != 21 {
		t.Error("longest canonical value is expected to be 21 bytes")
	}
	// Too-long input does not echo the input.
	if _, err := Parse(zeros(40)); err == nil || len(err.Error()) > 80 {
		t.Errorf("error = %v", err)
	}
}

func TestErrorEchoIsBounded(t *testing.T) {
	huge := make([]byte, 1<<20)
	for i := range huge {
		huge[i] = 'x'
	}
	_, err := Parse(string(huge))
	if err == nil || len(err.Error()) > 200 {
		t.Errorf("Parse error too long or nil: %d", len(err.Error()))
	}
	var m Money
	err = m.UnmarshalJSON(append(append([]byte{'['}, huge...), ']'))
	if err == nil || len(err.Error()) > 200 {
		t.Errorf("UnmarshalJSON error too long or nil")
	}
	err = m.UnmarshalText(huge)
	if err == nil || len(err.Error()) > 200 {
		t.Errorf("UnmarshalText error too long or nil")
	}
	// Short inputs are echoed in full, without a marker.
	if _, err := Parse("12.3"); err == nil || err.Error() != `money: invalid format: "12.3"` {
		t.Errorf("short echo = %v", err)
	}
	// Multi-byte runes are cut on rune boundaries.
	if _, err := Parse(string(make([]rune, 0)) + "ééééééééééééééééééééééééééééééééééééé"); err == nil || len(err.Error()) > 200 {
		t.Errorf("multibyte echo = %v", err)
	}
}

func TestAddSubNegOverflow(t *testing.T) {
	if _, err := Money(math.MaxInt64).Add(1); !errors.Is(err, ErrOverflow) {
		t.Error("Add overflow not detected")
	}
	if _, err := Money(math.MinInt64).Add(-1); !errors.Is(err, ErrOverflow) {
		t.Error("Add underflow not detected")
	}
	if _, err := Money(math.MinInt64).Sub(1); !errors.Is(err, ErrOverflow) {
		t.Error("Sub underflow not detected")
	}
	if _, err := Money(math.MaxInt64).Sub(-1); !errors.Is(err, ErrOverflow) {
		t.Error("Sub overflow not detected")
	}
	if _, err := Money(0).Sub(math.MinInt64); !errors.Is(err, ErrOverflow) {
		t.Error("0 - MinInt64 overflow not detected")
	}
	if _, err := Money(math.MinInt64).Neg(); !errors.Is(err, ErrOverflow) {
		t.Error("Neg(Min) overflow not detected")
	}
	if got, err := Money(-5).Abs(); err != nil || got != 5 {
		t.Errorf("Abs(-5) = %d, %v", got, err)
	}
	if got, err := Money(math.MaxInt64).Sub(math.MaxInt64); err != nil || got != 0 {
		t.Errorf("Max-Max = %d, %v", got, err)
	}
	if got, err := Sum(100, 250, -50); err != nil || got != 300 {
		t.Errorf("Sum = %d, %v", got, err)
	}
	if _, err := Sum(math.MaxInt64, 1); !errors.Is(err, ErrOverflow) {
		t.Error("Sum overflow not detected")
	}
}

func TestRoundingTable(t *testing.T) {
	// Dividing by 2 turns odd cents into exact ties: x/2 = n + 0.5.
	tests := []struct {
		cents Money
		mode  Rounding
		want  Money
	}{
		{5, HalfEven, 2}, {7, HalfEven, 4}, {-5, HalfEven, -2}, {-7, HalfEven, -4},
		{5, HalfAwayFromZero, 3}, {7, HalfAwayFromZero, 4}, {-5, HalfAwayFromZero, -3}, {-7, HalfAwayFromZero, -4},
		{5, Down, 2}, {-5, Down, -2},
		{5, Up, 3}, {-5, Up, -3},
		{4, Up, 2}, {-4, Down, -2}, // exact: no rounding
	}
	for _, tc := range tests {
		got, err := tc.cents.Div(2, tc.mode)
		if err != nil || got != tc.want {
			t.Errorf("%d / 2 (%v) = %d, %v; want %d", int64(tc.cents), tc.mode, int64(got), err, int64(tc.want))
		}
	}
}

func TestNonTieRounding(t *testing.T) {
	// 10 cents / 3 = 3.33..., 20 / 3 = 6.66...
	cases := []struct {
		c    Money
		mode Rounding
		want Money
	}{
		{10, HalfEven, 3}, {20, HalfEven, 7}, {10, Up, 4}, {10, Down, 3},
		{-10, HalfEven, -3}, {-20, HalfAwayFromZero, -7}, {-10, Up, -4}, {-10, Down, -3},
	}
	for _, tc := range cases {
		got, err := tc.c.Div(3, tc.mode)
		if err != nil || got != tc.want {
			t.Errorf("%d / 3 (%v) = %d, %v; want %d", int64(tc.c), tc.mode, int64(got), err, int64(tc.want))
		}
	}
}

func TestMulBasisPointsExample(t *testing.T) {
	// 5% of $0.50 = 2.5 cents (the tie example)
	if got, _ := Money(50).MulBasisPoints(500, HalfEven); got != 2 {
		t.Errorf("half-even = %d, want 2", got)
	}
	if got, _ := Money(50).MulBasisPoints(500, HalfAwayFromZero); got != 3 {
		t.Errorf("half-away = %d, want 3", got)
	}
	// 7.25% of $19.99 = 144.9275 cents -> 145
	if got, _ := Money(1999).MulBasisPoints(725, HalfEven); got != 145 {
		t.Errorf("tax = %d, want 145", got)
	}
}

func TestMulIntDivErrors(t *testing.T) {
	if got, err := Money(1999).MulInt(3); err != nil || got != 5997 {
		t.Errorf("MulInt = %d, %v", got, err)
	}
	if got, err := Money(1999).MulInt(-3); err != nil || got != -5997 {
		t.Errorf("MulInt neg = %d, %v", got, err)
	}
	if _, err := Money(math.MaxInt64).MulInt(2); !errors.Is(err, ErrOverflow) {
		t.Error("MulInt overflow not detected")
	}
	if _, err := Money(math.MinInt64).MulInt(-1); !errors.Is(err, ErrOverflow) {
		t.Error("Min * -1 overflow not detected")
	}
	if got, err := Money(math.MinInt64).MulInt(1); err != nil || got != math.MinInt64 {
		t.Errorf("Min * 1 = %d, %v", got, err)
	}
	if got, err := Money(math.MinInt64).Div(-1, HalfEven); !errors.Is(err, ErrOverflow) {
		t.Errorf("Min / -1 = %d, %v; want overflow", got, err)
	}
	if _, err := Money(1).Div(0, HalfEven); !errors.Is(err, ErrDivideByZero) {
		t.Error("divide by zero not detected")
	}
	if _, err := Money(1).MulRatio(1, 1, Rounding(99)); !errors.Is(err, ErrInvalidRounding) {
		t.Error("invalid rounding not detected")
	}
	// Large intermediate product that still fits after dividing.
	if got, err := Money(math.MaxInt64).MulRatio(math.MaxInt64, math.MaxInt64, Down); err != nil || got != math.MaxInt64 {
		t.Errorf("Max*Max/Max = %d, %v", got, err)
	}
}

// refMulRatio is an independent math/big reference implementation.
func refMulRatio(m, num, den int64, mode Rounding) (int64, bool) {
	n := new(big.Int).Mul(big.NewInt(m), big.NewInt(num))
	d := big.NewInt(den)
	q, r := new(big.Int).QuoRem(n, d, new(big.Int)) // truncated toward zero
	if r.Sign() != 0 {
		neg := n.Sign()*d.Sign() < 0
		twice := new(big.Int).Abs(new(big.Int).Lsh(r, 1))
		cmp := twice.Cmp(new(big.Int).Abs(d))
		away := false
		switch mode {
		case Up:
			away = true
		case HalfAwayFromZero:
			away = cmp >= 0
		case HalfEven:
			away = cmp > 0 || (cmp == 0 && q.Bit(0) == 1)
		}
		if away {
			if neg {
				q.Sub(q, big.NewInt(1))
			} else {
				q.Add(q, big.NewInt(1))
			}
		}
	}
	return q.Int64(), q.IsInt64()
}

func TestMulRatioAgainstBigInt(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	edge := []int64{0, 1, -1, 2, -2, 3, 100, -100, 10000, math.MaxInt64, math.MinInt64, math.MaxInt64 - 1, math.MinInt64 + 1}
	pick := func() int64 {
		switch rng.Intn(3) {
		case 0:
			return edge[rng.Intn(len(edge))]
		case 1:
			return rng.Int63n(2000) - 1000
		}
		return int64(rng.Uint64())
	}
	for range 300000 {
		m, num, den := pick(), pick(), pick()
		if den == 0 {
			continue
		}
		mode := Rounding(rng.Intn(4))
		want, ok := refMulRatio(m, num, den, mode)
		got, err := Money(m).MulRatio(num, den, mode)
		if !ok {
			if !errors.Is(err, ErrOverflow) {
				t.Fatalf("MulRatio(%d,%d,%d,%v) = %d, %v; want overflow", m, num, den, mode, int64(got), err)
			}
			continue
		}
		if err != nil || int64(got) != want {
			t.Fatalf("MulRatio(%d,%d,%d,%v) = %d, %v; want %d", m, num, den, mode, int64(got), err, want)
		}
	}
}

func TestSplitAllocate(t *testing.T) {
	got, err := Money(100).Split(3)
	if err != nil || len(got) != 3 || got[0] != 34 || got[1] != 33 || got[2] != 33 {
		t.Errorf("Split(100,3) = %v, %v", got, err)
	}
	got, _ = Money(-100).Split(3)
	if got[0] != -34 || got[1] != -33 || got[2] != -33 {
		t.Errorf("Split(-100,3) = %v", got)
	}
	got, _ = Money(5).Allocate([]int64{70, 30})
	if got[0] != 4 || got[1] != 1 { // 3.5 + 1.5 -> floors 3,1 + remainder 1 to first
		t.Errorf("Allocate = %v", got)
	}
	got, _ = Money(1).Allocate([]int64{0, 1, 1})
	if got[0] != 0 || got[1] != 1 || got[2] != 0 {
		t.Errorf("zero ratio must not receive cents: %v", got)
	}
	for _, bad := range [][]int64{nil, {}, {0, 0}, {-1, 2}} {
		if _, err := Money(10).Allocate(bad); !errors.Is(err, ErrInvalidRatio) {
			t.Errorf("Allocate(%v) err = %v", bad, err)
		}
	}
	if _, err := Money(10).Split(0); !errors.Is(err, ErrInvalidRatio) {
		t.Error("Split(0) should fail")
	}
	if _, err := Money(10).Allocate([]int64{math.MaxInt64, 1}); !errors.Is(err, ErrOverflow) {
		t.Error("ratio sum overflow not detected")
	}
}

func TestAllocateAlwaysSumsExactly(t *testing.T) {
	rng := rand.New(rand.NewSource(2))
	for range 50000 {
		m := Money(rng.Int63n(2_000_000) - 1_000_000)
		ratios := make([]int64, 1+rng.Intn(8))
		var any bool
		for j := range ratios {
			ratios[j] = rng.Int63n(5)
			any = any || ratios[j] > 0
		}
		if !any {
			continue
		}
		parts, err := m.Allocate(ratios)
		if err != nil {
			t.Fatal(err)
		}
		sum, _ := Sum(parts...)
		if sum != m {
			t.Fatalf("Allocate(%d,%v) = %v sums to %d", m, ratios, parts, sum)
		}
		for j, p := range parts {
			if ratios[j] == 0 && p != 0 {
				t.Fatalf("zero ratio got %d: %v %v", p, ratios, parts)
			}
		}
	}
}

type invoice struct {
	Total Money            `json:"total"`
	Tip   *Money           `json:"tip,omitempty"`
	ByKey map[Money]string `json:"by_key,omitempty"`
}

func TestJSON(t *testing.T) {
	tip := Money(-5)
	b, err := json.Marshal(invoice{Total: 1234, Tip: &tip})
	if err != nil || string(b) != `{"total":"12.34","tip":"-0.05"}` {
		t.Errorf("Marshal = %s, %v", b, err)
	}
	var inv invoice
	if err := json.Unmarshal([]byte(`{"total":"12.34","tip":"-0.05"}`), &inv); err != nil || inv.Total != 1234 || *inv.Tip != -5 {
		t.Errorf("Unmarshal = %+v, %v", inv, err)
	}
	// map keys use TextMarshaler
	b, err = json.Marshal(invoice{ByKey: map[Money]string{150: "a"}})
	if err != nil || string(b) != `{"total":"0.00","by_key":{"1.50":"a"}}` {
		t.Errorf("map Marshal = %s, %v", b, err)
	}
	var back invoice
	if err := json.Unmarshal(b, &back); err != nil || back.ByKey[150] != "a" {
		t.Errorf("map Unmarshal = %+v, %v", back, err)
	}
}

func TestJSONRejectsAndNull(t *testing.T) {
	bad := []string{`12.34`, `1234`, `true`, `[]`, `{}`, `""`, `"12"`, `"12.3"`, `"12.345"`,
		`"1e2"`, `"+1.00"`, `"1,000.00"`, `" 1.00"`, `"\u0031.00"`, `"1.00\n"`, `"abc"`}
	for _, s := range bad {
		m := Money(777)
		err := json.Unmarshal([]byte(s), &m)
		if err == nil {
			t.Errorf("Unmarshal(%s) accepted, got %d", s, int64(m))
		}
		if m != 777 {
			t.Errorf("Unmarshal(%s) modified value on error: %d", s, int64(m))
		}
	}
	m := Money(777)
	if err := json.Unmarshal([]byte(`null`), &m); err != nil || m != 777 {
		t.Errorf("null should be a no-op: %d, %v", int64(m), err)
	}
	if err := m.UnmarshalJSON([]byte(`"1.00"`)); err != nil || m != 100 {
		t.Errorf("direct call = %d, %v", int64(m), err)
	}
}

func TestSQL(t *testing.T) {
	v, err := Money(1234).Value()
	if err != nil || v != int64(1234) {
		t.Errorf("Value = %v, %v", v, err)
	}
	var m Money
	for _, src := range []any{int64(1234), []byte("1234"), "1234"} {
		m = 0
		if err := m.Scan(src); err != nil || m != 1234 {
			t.Errorf("Scan(%T) = %d, %v", src, int64(m), err)
		}
	}
	for _, src := range []any{nil, 12.34, "12.34", []byte("abc"), true} {
		if err := m.Scan(src); err == nil {
			t.Errorf("Scan(%v) should fail", src)
		}
	}
	var n sql.Null[Money]
	if err := n.Scan(nil); err != nil || n.Valid {
		t.Errorf("sql.Null[Money] with NULL = %+v, %v", n, err)
	}
	if err := n.Scan(int64(5)); err != nil || !n.Valid || n.V != 5 {
		t.Errorf("sql.Null[Money] = %+v, %v", n, err)
	}
}

func TestCmpSign(t *testing.T) {
	if Money(1).Cmp(2) != -1 || Money(2).Cmp(2) != 0 || Money(3).Cmp(2) != 1 {
		t.Error("Cmp")
	}
	if Money(-3).Sign() != -1 || Money(0).Sign() != 0 || Money(3).Sign() != 1 || !Money(0).IsZero() {
		t.Error("Sign/IsZero")
	}
}

func FuzzParseFormat(f *testing.F) {
	for _, s := range []string{"0.00", "12.34", "-0.05", "92233720368547758.07",
		"05.00", "01.00", "-007.50", // leading-zero normalization
		"00000" + "92233720368547758.07", // leading zeros within the length cap
		zeros(32), zeros(33), "x", "1.0", ""} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		m, err := Parse(s)
		if err != nil {
			return
		}
		// Anything accepted must survive a format/parse round trip.
		back, err := Parse(m.String())
		if err != nil || back != m {
			t.Fatalf("Parse(%q) = %d, reparse of %q gave %d, %v", s, int64(m), m.String(), int64(back), err)
		}
	})
}

func FuzzFormatParse(f *testing.F) {
	f.Add(int64(0))
	f.Add(int64(math.MinInt64))
	f.Add(int64(math.MaxInt64))
	f.Fuzz(func(t *testing.T, c int64) {
		m := Money(c)
		back, err := Parse(m.String())
		if err != nil || back != m {
			t.Fatalf("round trip %d -> %q -> %d, %v", c, m.String(), int64(back), err)
		}
		b, _ := json.Marshal(m)
		var j Money
		if err := json.Unmarshal(b, &j); err != nil || j != m {
			t.Fatalf("json round trip %d: %v", c, err)
		}
	})
}
