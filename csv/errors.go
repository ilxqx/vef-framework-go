package csv

import "errors"

var (
	// ErrDataMustBeSlice indicates exported data must be a slice.
	ErrDataMustBeSlice = errors.New("data must be a slice")
	// ErrNoDataRowsFound indicates there are no data rows after skips/header.
	ErrNoDataRowsFound = errors.New("no data rows found")
	// ErrDuplicateColumnName indicates duplicate header column name.
	ErrDuplicateColumnName = errors.New("duplicate column name")
	// ErrFieldNotSettable indicates struct field is not settable during import.
	ErrFieldNotSettable = errors.New("field is not settable")
)
