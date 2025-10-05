package tabular

import (
	"errors"
	"fmt"

	"github.com/ilxqx/vef-framework-go/constants"
)

// ErrUnsupportedType indicates the target type is not supported by the parser.
var ErrUnsupportedType = errors.New("unsupported type")

// ImportError represents an error that occurred during data import.
type ImportError struct {
	// Row is the row number where the error occurred (1-based, includes header row)
	Row int
	// Column is the column name where the error occurred (optional)
	Column string
	// Field is the struct field name where the error occurred (optional)
	Field string
	// Err is the underlying error
	Err error
}

// Error implements the error interface.
func (e ImportError) Error() string {
	if e.Column != constants.Empty && e.Field != constants.Empty {
		return fmt.Sprintf("row %d, column %s (field %s): %v", e.Row, e.Column, e.Field, e.Err)
	}

	if e.Column != constants.Empty {
		return fmt.Sprintf("row %d, column %s: %v", e.Row, e.Column, e.Err)
	}

	if e.Field != constants.Empty {
		return fmt.Sprintf("row %d, field %s: %v", e.Row, e.Field, e.Err)
	}

	return fmt.Sprintf("row %d: %v", e.Row, e.Err)
}

// Unwrap returns the underlying error.
func (e ImportError) Unwrap() error {
	return e.Err
}

// ExportError represents an error that occurred during data export.
type ExportError struct {
	// Row is the data row index where the error occurred (0-based)
	Row int
	// Column is the column name where the error occurred (optional)
	Column string
	// Field is the struct field name where the error occurred (optional)
	Field string
	// Err is the underlying error
	Err error
}

// Error implements the error interface.
func (e ExportError) Error() string {
	if e.Column != constants.Empty && e.Field != constants.Empty {
		return fmt.Sprintf("row %d, column %s (field %s): %v", e.Row, e.Column, e.Field, e.Err)
	}

	if e.Column != constants.Empty {
		return fmt.Sprintf("row %d, column %s: %v", e.Row, e.Column, e.Err)
	}

	if e.Field != constants.Empty {
		return fmt.Sprintf("row %d, field %s: %v", e.Row, e.Field, e.Err)
	}

	return fmt.Sprintf("row %d: %v", e.Row, e.Err)
}

// Unwrap returns the underlying error.
func (e ExportError) Unwrap() error {
	return e.Err
}
