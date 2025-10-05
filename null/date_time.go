package null

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/datetime"
)

// DateTime is a nullable datetime.DateTime. It supports SQL and JSON serialization.
// It will marshal to null if null.
type DateTime struct {
	sql.Null[datetime.DateTime]
}

// NewDateTime creates a new DateTime.
func NewDateTime(dt datetime.DateTime, valid bool) DateTime {
	return DateTime{
		Null: sql.Null[datetime.DateTime]{
			V:     dt,
			Valid: valid,
		},
	}
}

// DateTimeFrom creates a new DateTime that will always be valid.
func DateTimeFrom(dt datetime.DateTime) DateTime {
	return NewDateTime(dt, true)
}

// DateTimeFromPtr creates a new DateTime that will be null if dt is nil.
func DateTimeFromPtr(dt *datetime.DateTime) DateTime {
	if dt == nil {
		return NewDateTime(datetime.DateTime{}, false)
	}

	return NewDateTime(*dt, true)
}

// ValueOrZero returns the inner value if valid, otherwise zero.
func (dt DateTime) ValueOrZero() datetime.DateTime {
	if !dt.Valid {
		return datetime.DateTime{}
	}

	return dt.V
}

// ValueOr returns the inner value if valid, otherwise v.
func (dt DateTime) ValueOr(v datetime.DateTime) datetime.DateTime {
	if !dt.Valid {
		return v
	}

	return dt.V
}

// MarshalJSON implements json.Marshaler.
// It will encode null if this datetime is null.
func (dt DateTime) MarshalJSON() ([]byte, error) {
	if !dt.Valid {
		return constants.JSONNullBytes, nil
	}

	return dt.V.MarshalJSON()
}

// UnmarshalJSON implements json.Unmarshaler.
// It supports string and null input.
func (dt *DateTime) UnmarshalJSON(data []byte) error {
	if len(data) > 0 && data[0] == 'n' {
		dt.Valid = false

		return nil
	}

	if err := json.Unmarshal(data, &dt.V); err != nil {
		return fmt.Errorf("null: couldn't unmarshal JSON: %w", err)
	}

	dt.Valid = true

	return nil
}

// MarshalText implements encoding.TextMarshaler.
// It returns an empty string if invalid, otherwise datetime.DateTime's MarshalText.
func (dt DateTime) MarshalText() ([]byte, error) {
	if !dt.Valid {
		return []byte{}, nil
	}

	return dt.V.MarshalText()
}

// UnmarshalText implements encoding.TextUnmarshaler.
// It has backwards compatibility with v3 in that the string "null" is considered equivalent to an empty string
// and unmarshaling will succeed. This may be removed in a future version.
func (dt *DateTime) UnmarshalText(text []byte) error {
	str := string(text)
	// allowing "null" is for backwards compatibility with v3
	if str == constants.Empty || str == constants.JSONNull {
		dt.Valid = false

		return nil
	}

	if err := dt.V.UnmarshalText(text); err != nil {
		return fmt.Errorf("null: couldn't unmarshal text: %w", err)
	}

	dt.Valid = true

	return nil
}

// SetValid changes this DateTime's value and sets it to be non-null.
func (dt *DateTime) SetValid(v datetime.DateTime) {
	dt.V = v
	dt.Valid = true
}

// Ptr returns a pointer to this DateTime's value, or a nil pointer if this DateTime is null.
func (dt DateTime) Ptr() *datetime.DateTime {
	if !dt.Valid {
		return nil
	}

	return &dt.V
}

// IsZero returns true for invalid DateTimes, hopefully for future omitempty support.
// A non-null DateTime with a zero value will not be considered zero.
func (dt DateTime) IsZero() bool {
	return !dt.Valid
}

// Equal returns true if both DateTime objects encode the same datetime or are both null.
// Two datetimes can be equal even if they are in different locations.
// For example, 2023-01-01 12:00:00 +0200 CEST and 2023-01-01 10:00:00 UTC are Equal.
func (dt DateTime) Equal(other DateTime) bool {
	return dt.Valid == other.Valid && (!dt.Valid || dt.V.Equal(other.V))
}

// ExactEqual returns true if both DateTime objects are equal or both null.
// ExactEqual returns false for datetimes that are in different locations or
// have a different monotonic clock reading.
func (dt DateTime) ExactEqual(other DateTime) bool {
	return dt.Valid == other.Valid && (!dt.Valid || dt.V == other.V)
}
