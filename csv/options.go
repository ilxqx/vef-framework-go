package csv

// importOptions contains configuration options for CSV import.
type importOptions struct {
	// delimiter is the field delimiter (default: ',')
	delimiter rune
	// hasHeader indicates whether the CSV has a header row (default: true)
	hasHeader bool
	// skipRows is the number of rows to skip before the header row (default: 0)
	skipRows int
	// trimSpace indicates whether to trim leading/trailing spaces (default: true)
	trimSpace bool
	// comment is the comment character (default: 0, disabled)
	comment rune
}

// ImportOption is a function that configures importOptions.
type ImportOption func(*importOptions)

func WithImportDelimiter(delimiter rune) ImportOption {
	return func(o *importOptions) {
		o.delimiter = delimiter
	}
}

// WithoutHeader indicates the CSV file does not have a header row.
// By default, hasHeader is true, so calling this function sets it to false.
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

// WithoutTrimSpace disables trimming of leading/trailing spaces.
// By default, trimSpace is true, so calling this function sets it to false.
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

// exportOptions contains configuration options for CSV export.
type exportOptions struct {
	// delimiter is the field delimiter (default: ',')
	delimiter rune
	// writeHeader indicates whether to write the header row (default: true)
	writeHeader bool
	// useCrlf indicates whether to use \r\n as line terminator (default: false)
	useCrlf bool
}

// ExportOption is a function that configures exportOptions.
type ExportOption func(*exportOptions)

func WithExportDelimiter(delimiter rune) ExportOption {
	return func(o *exportOptions) {
		o.delimiter = delimiter
	}
}

// WithoutWriteHeader disables writing the header row.
// By default, writeHeader is true, so calling this function sets it to false.
func WithoutWriteHeader() ExportOption {
	return func(o *exportOptions) {
		o.writeHeader = false
	}
}

// WithCrlf enables using \r\n as line terminator instead of \n.
// By default, useCRLF is false, so calling this function sets it to true.
func WithCrlf() ExportOption {
	return func(o *exportOptions) {
		o.useCrlf = true
	}
}
