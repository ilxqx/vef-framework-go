package tabular

import (
	"reflect"
	"sort"

	"github.com/ilxqx/vef-framework-go/internal/log"
)

var logger = log.Named("tabular")

// Schema contains the pre-parsed tabular metadata for a struct type.
type Schema struct {
	columns []*Column
}

// Column represents metadata for a single column in tabular data.
type Column struct {
	Index     []int   // Field index path in the struct
	Name      string  // Column name (header)
	Width     float64 // Column width hint (for display/export)
	Order     int     // Column order (for sorting)
	Default   string  // Default value used during import when cell is empty
	Format    string  // Format template (e.g., date format, number format)
	Formatter string  // Custom formatter name for export
	Parser    string  // Custom parser name for import
}

// NewSchema creates a Schema instance by parsing struct fields with tabular tags.
func NewSchema(typ reflect.Type) *Schema {
	columns := parseStruct(typ)

	sort.SliceStable(columns, func(i, j int) bool {
		return columns[i].Order < columns[j].Order
	})

	return &Schema{columns: columns}
}

// NewSchemaFor creates a Schema instance from type T.
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
	names := make([]string, len(s.columns))
	for i, col := range s.columns {
		names[i] = col.Name
	}
	return names
}
