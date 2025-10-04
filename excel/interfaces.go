package excel

import (
	"bytes"
	"io"
	"reflect"
)

// ValueParser defines the interface for custom value parsers.
// Parsers convert Excel cell strings to Go values during import.
type ValueParser interface {
	// Parse converts an Excel cell string to a Go value
	Parse(cellValue string, targetType reflect.Type) (any, error)
}

// Formatter defines the interface for custom value formatters.
// Formatters convert Go values to Excel cell strings during export.
type Formatter interface {
	// Format converts a Go value to an Excel cell string
	Format(value any) (string, error)
}

// Importer defines the interface for importing data from Excel files.
type Importer interface {
	// RegisterParser registers a custom parser with the given name.
	RegisterParser(name string, parser ValueParser)
	// ImportFromFile imports data from an Excel file.
	ImportFromFile(filename string) (any, []ImportError, error)
	// Import imports data from an io.Reader.
	Import(reader io.Reader) (any, []ImportError, error)
}

// Exporter defines the interface for exporting data to Excel files.
type Exporter interface {
	// RegisterFormatter registers a custom formatter with the given name.
	RegisterFormatter(name string, formatter Formatter)
	// ExportToFile exports data to an Excel file.
	ExportToFile(data any, filename string) error
	// Export exports data to a bytes.Buffer.
	Export(data any) (*bytes.Buffer, error)
}
