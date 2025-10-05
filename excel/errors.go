package excel

import "errors"

var (
	// ErrDataMustBeSlice indicates exported data must be a slice.
	ErrDataMustBeSlice = errors.New("data must be a slice")
	// ErrSheetIndexOutOfRange indicates sheet index out of range.
	ErrSheetIndexOutOfRange = errors.New("sheet index out of range")
	// ErrNoDataRowsFound indicates there are no data rows after skips.
	ErrNoDataRowsFound = errors.New("no data rows found")
	// ErrDuplicateColumnName indicates duplicate header column name.
	ErrDuplicateColumnName = errors.New("duplicate column name")
	// ErrFieldNotSettable indicates struct field is not settable during import.
	ErrFieldNotSettable = errors.New("field is not settable")
)
