package mo

import (
	"database/sql/driver"
	"fmt"

	"github.com/goccy/go-json"
	"github.com/spf13/cast"
)

// JSON is a generic wrapper for JSON data that provides database and JSON marshaling support.
// It can hold any type T and automatically handles serialization/deserialization.
type JSON[T any] struct {
	value T // value is the underlying data of type T
}

// Scan implements the sql.Scanner interface for database scanning.
// It accepts various types and converts them to JSON bytes using cast library for type conversion.
func (j *JSON[T]) Scan(src any) error {
	if src == nil {
		return nil
	}

	var bs []byte
	switch value := src.(type) {
	case []byte:
		bs = value
	case *[]byte:
		if value == nil {
			return nil
		}
		bs = *value
	case string:
		bs = []byte(value)
	case *string:
		if value == nil {
			return nil
		}
		bs = []byte(*value)
	default:
		// Use cast library to convert to string, then to bytes
		str, err := cast.ToStringE(src)
		if err != nil {
			return fmt.Errorf("failed to convert JSON value to string: %w", err)
		}
		bs = []byte(str)
	}

	var value T
	if err := json.Unmarshal(bs, &value); err != nil {
		return fmt.Errorf("failed to unmarshal JSON value: %w", err)
	}

	*j = JSON[T]{value: value}
	return nil
}

// Value implements the driver.Valuer interface for database storage.
// It marshals the wrapped value to JSON string for database storage.
func (j JSON[T]) Value() (driver.Value, error) {
	bs, err := json.Marshal(j.value)
	if err != nil {
		return nil, err
	}

	return string(bs), nil
}

// MarshalJSON implements the json.Marshaler interface.
// It marshals the wrapped value to JSON bytes.
func (j JSON[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(j.value)
}

// UnmarshalJSON implements the json.Unmarshaler interface.
// It unmarshals JSON bytes into the wrapped value.
func (j *JSON[T]) UnmarshalJSON(bs []byte) error {
	return json.Unmarshal(bs, &j.value)
}

// Unwrap returns the underlying value without the JSON wrapper.
func (j JSON[T]) Unwrap() T {
	return j.value
}

// NewJSON creates a new JSON wrapper with the given value.
func NewJSON[T any](value T) JSON[T] {
	return JSON[T]{
		value: value,
	}
}
