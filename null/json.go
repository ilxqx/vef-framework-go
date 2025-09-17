package null

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"github.com/ilxqx/vef-framework-go/mo"
)

// JSON is a nullable mo.JSON[T]. It supports SQL and JSON serialization.
// It will marshal to null if null.
type JSON[T any] struct {
	sql.Null[mo.JSON[T]]
}

// Value implements the driver Valuer interface.
func (j JSON[T]) Value() (driver.Value, error) {
	if !j.Valid {
		return nil, nil
	}

	return j.V.Value()
}

// Scan implements the sql.Scanner interface.
func (j *JSON[T]) Scan(src any) error {
	if src == nil {
		j.Valid = false
		return nil
	}

	var value mo.JSON[T]
	if err := value.Scan(src); err != nil {
		return err
	}

	j.V = value
	j.Valid = true
	return nil
}

// NewJSON creates a new JSON.
func NewJSON[T any](j mo.JSON[T], valid bool) JSON[T] {
	return JSON[T]{
		Null: sql.Null[mo.JSON[T]]{
			V:     j,
			Valid: valid,
		},
	}
}

// JSONFrom creates a new JSON that will always be valid.
func JSONFrom[T any](j mo.JSON[T]) JSON[T] {
	return NewJSON(j, true)
}

// JSONFromValue creates a new JSON from a value that will always be valid.
func JSONFromValue[T any](value T) JSON[T] {
	return JSONFrom(mo.NewJSON(value))
}

// JSONFromPtr creates a new JSON that will be null if j is nil.
func JSONFromPtr[T any](j *mo.JSON[T]) JSON[T] {
	if j == nil {
		return NewJSON(mo.JSON[T]{}, false)
	}

	return NewJSON(*j, true)
}

// ValueOrZero returns the inner value if valid, otherwise zero.
func (j JSON[T]) ValueOrZero() mo.JSON[T] {
	if !j.Valid {
		return mo.JSON[T]{}
	}

	return j.V
}

// ValueOr returns the inner value if valid, otherwise v.
func (j JSON[T]) ValueOr(v mo.JSON[T]) mo.JSON[T] {
	if !j.Valid {
		return v
	}

	return j.V
}

// UnwrapOrZero returns the unwrapped inner value if valid, otherwise zero value of T.
func (j JSON[T]) UnwrapOrZero() T {
	if !j.Valid {
		var zero T
		return zero
	}

	return j.V.Unwrap()
}

// UnwrapOr returns the unwrapped inner value if valid, otherwise v.
func (j JSON[T]) UnwrapOr(v T) T {
	if !j.Valid {
		return v
	}

	return j.V.Unwrap()
}

// MarshalJSON implements json.Marshaler.
// It will encode null if this JSON is null.
func (j JSON[T]) MarshalJSON() ([]byte, error) {
	if !j.Valid {
		return mo.JSONNullBytes, nil
	}

	return j.V.MarshalJSON()
}

// UnmarshalJSON implements json.Unmarshaler.
// It supports string and null input.
func (j *JSON[T]) UnmarshalJSON(data []byte) error {
	if len(data) > 0 && data[0] == 'n' {
		j.Valid = false
		return nil
	}

	if err := json.Unmarshal(data, &j.V); err != nil {
		return fmt.Errorf("null: couldn't unmarshal JSON: %w", err)
	}

	j.Valid = true
	return nil
}

// SetValid changes this JSON's value and sets it to be non-null.
func (j *JSON[T]) SetValid(v mo.JSON[T]) {
	j.V = v
	j.Valid = true
}

// SetValidValue changes this JSON's value from unwrapped value and sets it to be non-null.
func (j *JSON[T]) SetValidValue(v T) {
	j.V = mo.NewJSON(v)
	j.Valid = true
}

// Ptr returns a pointer to this JSON's value, or a nil pointer if this JSON is null.
func (j JSON[T]) Ptr() *mo.JSON[T] {
	if !j.Valid {
		return nil
	}
	return &j.V
}

// IsZero returns true for invalid JSONs, hopefully for future omitempty support.
// A non-null JSON with a zero value will not be considered zero.
func (j JSON[T]) IsZero() bool {
	return !j.Valid
}
