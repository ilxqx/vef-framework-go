// Package null provides nullable versions of Go types that integrate with SQL databases.
// The Decimal type wraps shopspring/decimal.NullDecimal with additional convenience methods.
package null

import (
	"github.com/samber/lo"

	dec "github.com/shopspring/decimal"

	"github.com/ilxqx/vef-framework-go/decimal"
)

// Decimal represents a decimal.Decimal that may be null.
// It wraps shopspring/decimal.NullDecimal and provides additional convenience methods
// for common operations. The underlying NullDecimal handles SQL database integration
// through the driver.Valuer and sql.Scanner interfaces.
type Decimal struct {
	dec.NullDecimal
}

// NewDecimal creates a new Decimal with the given value and validity.
// If valid is false, the decimal is considered null regardless of the decimal value.
func NewDecimal(d decimal.Decimal, valid bool) Decimal {
	return Decimal{
		dec.NullDecimal{
			Decimal: d,
			Valid:   valid,
		},
	}
}

// DecimalFrom creates a valid (non-null) Decimal from a decimal.Decimal value.
// This is equivalent to NewDecimal(d, true).
func DecimalFrom(d decimal.Decimal) Decimal {
	return NewDecimal(d, true)
}

// DecimalFromPtr creates a Decimal from a pointer to decimal.Decimal.
// If the pointer is nil, returns an invalid (null) Decimal.
// If the pointer is not nil, returns a valid Decimal with the dereferenced value.
func DecimalFromPtr(d *decimal.Decimal) Decimal {
	if d == nil {
		return NewDecimal(lo.Empty[decimal.Decimal](), false)
	}

	return NewDecimal(*d, true)
}

// ValueOrZero returns the decimal value if valid, or decimal.Zero if null.
// This method provides a safe way to get a usable decimal value without
// having to check validity manually.
func (d Decimal) ValueOrZero() decimal.Decimal {
	if !d.Valid {
		return decimal.Zero
	}

	return d.Decimal
}

// ValueOr returns the decimal value if valid, or the provided fallback value if null.
// This allows for custom default values when the decimal is null.
func (d Decimal) ValueOr(v decimal.Decimal) decimal.Decimal {
	if !d.Valid {
		return v
	}

	return d.Decimal
}

// This method modifies the receiver, so it must be called on a pointer.
func (d *Decimal) SetValid(v decimal.Decimal) {
	d.Decimal = v
	d.Valid = true
}

// Ptr returns a pointer to the decimal value if valid, or nil if null.
// This is useful when you need to pass the value to Apis that expect *decimal.Decimal,
// with nil representing the absence of a value.
func (d Decimal) Ptr() *decimal.Decimal {
	if !d.Valid {
		return nil
	}

	return &d.Decimal
}

// IsZero returns true if the decimal is null (invalid).
// Note: This checks for null status, not whether the decimal value itself is zero.
// A valid decimal with value 0 will return false.
func (d Decimal) IsZero() bool {
	return !d.Valid
}

// Equal performs semantic equality comparison between two Decimal values.
// Two decimals are equal if they have the same validity status and,
// if both are valid, their decimal values are mathematically equal
// (using decimal.Equal which handles precision differences correctly).
func (d Decimal) Equal(other Decimal) bool {
	return d.Valid == other.Valid && (!d.Valid || d.Decimal.Equal(other.Decimal))
}

// ExactEqual performs exact equality comparison between two Decimal values.
// Unlike Equal, this method requires both the validity status and the
// underlying decimal representation (including precision and scale) to be identical.
// This is stricter than Equal and should be used when exact representation matters.
func (d Decimal) ExactEqual(other Decimal) bool {
	return d.Valid == other.Valid && (!d.Valid || d.Decimal == other.Decimal)
}
