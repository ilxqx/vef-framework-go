package excel

// exportOptions contains configuration options for Excel export.
type exportOptions struct {
	// sheetName is the name of the Excel sheet (default: "Sheet1")
	sheetName string
}

// ExportOption is a function that configures ExportOptions.
type ExportOption func(*exportOptions)

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

func WithImportSheetName(name string) ImportOption {
	return func(o *importOptions) {
		o.sheetName = name
	}
}

func WithImportSheetIndex(index int) ImportOption {
	return func(o *importOptions) {
		o.sheetIndex = index
	}
}

func WithSkipRows(rows int) ImportOption {
	return func(o *importOptions) {
		o.skipRows = rows
	}
}
