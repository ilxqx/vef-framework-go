package tabular

import (
	"reflect"
	"sort"

	"github.com/ilxqx/go-streams"

	"github.com/ilxqx/vef-framework-go/internal/log"
)

var logger = log.Named("tabular")

// Schema contains the pre-parsed tabular metadata for a struct type.
type Schema struct {
	columns []*Column
}

// Column represents metadata for a single column in tabular data.
type Column struct {
	// Index is the field index path in the struct
	Index []int
	// Name is the column name (header)
	Name string
	// Width is the column width hint (for display/export)
	Width float64
	// Order is the column order (for sorting)
	Order int
	// Default is the default value used during import when cell is empty
	Default string
	// Format is the format template (e.g., date format, number format)
	Format string
	// Formatter is the custom formatter name for export
	Formatter string
	// Parser is the custom parser name for import
	Parser string
}

// NewSchema creates a Schema instance by parsing struct fields with tabular tags from the given reflect.Type.
// Returns an empty Schema if the type is not a struct.
func NewSchema(typ reflect.Type) *Schema {
	columns := parseStruct(typ)

	// Sort columns by order
	sort.SliceStable(columns, func(i, j int) bool {
		return columns[i].Order < columns[j].Order
	})

	return &Schema{columns: columns}
}

// NewSchemaFor creates a Schema instance by parsing struct fields with tabular tags from type T.
// This is a generic convenience function that calls NewSchema with reflect.TypeFor[T]().
func NewSchemaFor[T any]() *Schema {
	return NewSchema(reflect.TypeFor[T]())
}

// Columns returns all columns in the schema.
func (s *Schema) Columns() []*Column {
	return s.columns
}

// ColumnCount returns the number of columns.
func (s *Schema) ColumnCount() int {
	return len(s.columns)
}

// ColumnNames returns all column names.
func (s *Schema) ColumnNames() []string {
	// Use streams.MapTo for declarative column name extraction
	return streams.MapTo(
		streams.FromSlice(s.columns),
		func(col *Column) string { return col.Name },
	).Collect()
}
