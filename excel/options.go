package excel

// exportOptions contains configuration options for Excel export.
type exportOptions struct {
	// sheetName is the name of the Excel sheet (default: "Sheet1")
	sheetName string
}

// ExportOption is a function that configures ExportOptions.
type ExportOption func(*exportOptions)

// WithSheetName sets the sheet name for export.
func WithSheetName(name string) ExportOption {
	return func(o *exportOptions) {
		o.sheetName = name
	}
}

// importOptions contains configuration options for Excel import.
type importOptions struct {
	// sheetName is the name of the Excel sheet to import (default: first sheet)
	sheetName string
	// sheetIndex is the index of the Excel sheet to import (0-based, default: 0)
	sheetIndex int
	// skipRows is the number of rows to skip before the header row (default: 0)
	skipRows int
}

// ImportOption is a function that configures ImportOptions.
type ImportOption func(*importOptions)

// WithImportSheetName sets the sheet name for import.
func WithImportSheetName(name string) ImportOption {
	return func(o *importOptions) {
		o.sheetName = name
	}
}

// WithImportSheetIndex sets the sheet index for import.
func WithImportSheetIndex(index int) ImportOption {
	return func(o *importOptions) {
		o.sheetIndex = index
	}
}

// WithSkipRows sets the number of rows to skip before the header row.
func WithSkipRows(rows int) ImportOption {
	return func(o *importOptions) {
		o.skipRows = rows
	}
}
