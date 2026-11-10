package money

import (
	"bytes"
	"fmt"
)

// Compile-time interface checks.
var (
	_ interface{ MarshalJSON() ([]byte, error) } = Money(0)
	_ interface{ UnmarshalJSON([]byte) error }   = (*Money)(nil)
	_ interface{ MarshalText() ([]byte, error) } = Money(0)
	_ interface{ UnmarshalText([]byte) error }   = (*Money)(nil)
	_ fmt.Stringer                               = Money(0)
)

// MarshalJSON encodes m as a JSON string holding a decimal, e.g. "12.34".
func (m Money) MarshalJSON() ([]byte, error) {
	s := m.String()
	b := make([]byte, 0, len(s)+2)
	b = append(b, '"')
	b = append(b, s...)
	b = append(b, '"')
	return b, nil
}

// UnmarshalJSON decodes a JSON string such as "12.34". JSON numbers, other
// types and malformed strings are rejected. JSON null is a no-op, following
// the convention of encoding/json. On error *m is left unchanged.
func (m *Money) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte("null")) {
		return nil
	}
	if len(data) < 2 || data[0] != '"' || data[len(data)-1] != '"' {
		// Only a bounded prefix is echoed (4 bytes per rune at most, plus one
		// byte so snippet can tell the input was longer).
		head := data
		if len(head) > 4*maxEchoRunes+1 {
			head = head[:4*maxEchoRunes+1]
		}
		return fmt.Errorf("%w: expected a JSON string, got %s", ErrInvalidFormat, snippet(string(head)))
	}
	// The inner text must be pure ASCII digits/'-'/'.', so any escape
	// sequence (backslash) simply fails Parse.
	v, err := Parse(string(data[1 : len(data)-1]))
	if err != nil {
		return err
	}
	*m = v
	return nil
}

// MarshalText implements encoding.TextMarshaler ("12.34"). It is used by
// encoding/json for map keys, and by YAML/TOML/query-param libraries.
func (m Money) MarshalText() ([]byte, error) { return []byte(m.String()), nil }

// UnmarshalText implements encoding.TextUnmarshaler using the same strict
// rules as Parse. On error *m is left unchanged.
func (m *Money) UnmarshalText(text []byte) error {
	v, err := Parse(string(text))
	if err != nil {
		return err
	}
	*m = v
	return nil
}
