package money

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"strconv"
)

var (
	_ driver.Valuer = Money(0)
	_ sql.Scanner   = (*Money)(nil)
)

// Value implements driver.Valuer. Money is stored as integer cents, so use
// a BIGINT column.
func (m Money) Value() (driver.Value, error) { return int64(m), nil }

// Scan implements sql.Scanner. It reads integer cents (BIGINT). Some drivers
// return integers as []byte or string; those are parsed as integer cents too,
// NOT as "12.34". NULL is an error: use sql.Null[money.Money] for nullable
// columns.
func (m *Money) Scan(src any) error {
	switch v := src.(type) {
	case int64:
		*m = Money(v)
		return nil
	case []byte:
		return m.scanString(string(v))
	case string:
		return m.scanString(v)
	case nil:
		return fmt.Errorf("money: cannot scan NULL into Money (use sql.Null[Money])")
	}
	return fmt.Errorf("money: cannot scan %T into Money", src)
}

func (m *Money) scanString(s string) error {
	c, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return fmt.Errorf("money: cannot scan %q as integer cents: %w", s, err)
	}
	*m = Money(c)
	return nil
}
