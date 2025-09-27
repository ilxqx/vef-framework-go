package sql

import (
	"database/sql/driver"

	"github.com/samber/lo"
)

// Bool is a simple wrapper around the built-in bool type that implements the database driver.Valuer interface.
// It converts Go boolean values to database-compatible integer values (1 for true, 0 for false).
// This is useful when you need to store boolean values in databases that don't have native boolean support.
type Bool bool

// Value implements the driver.Valuer interface for database storage.
// It converts the boolean value to an int16: true becomes 1, false becomes 0.
// This ensures consistent boolean representation across different database systems.
func (b Bool) Value() (driver.Value, error) {
	return lo.Ternary[int16](bool(b), 1, 0), nil
}
