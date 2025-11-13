package csv

type importOptions struct {
	delimiter rune
	hasHeader bool
	skipRows  int
	trimSpace bool
	comment   rune
}

type ImportOption func(*importOptions)

func WithImportDelimiter(delimiter rune) ImportOption {
	return func(o *importOptions) {
		o.delimiter = delimiter
	}
}

func WithoutHeader() ImportOption {
	return func(o *importOptions) {
		o.hasHeader = false
	}
}

func WithSkipRows(rows int) ImportOption {
	return func(o *importOptions) {
		o.skipRows = rows
	}
}

func WithoutTrimSpace() ImportOption {
	return func(o *importOptions) {
		o.trimSpace = false
	}
}

func WithComment(comment rune) ImportOption {
	return func(o *importOptions) {
		o.comment = comment
	}
}

type exportOptions struct {
	delimiter   rune
	writeHeader bool
	useCrlf     bool
}

type ExportOption func(*exportOptions)

func WithExportDelimiter(delimiter rune) ExportOption {
	return func(o *exportOptions) {
		o.delimiter = delimiter
	}
}

func WithoutWriteHeader() ExportOption {
	return func(o *exportOptions) {
		o.writeHeader = false
	}
}

// WithCrlf enables Windows-style line endings for compatibility with legacy systems.
func WithCrlf() ExportOption {
	return func(o *exportOptions) {
		o.useCrlf = true
	}
}
