package excel

type exportOptions struct {
	sheetName string
}

type ExportOption func(*exportOptions)

func WithSheetName(name string) ExportOption {
	return func(o *exportOptions) {
		o.sheetName = name
	}
}

type importOptions struct {
	sheetName  string
	sheetIndex int
	skipRows   int
}

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
