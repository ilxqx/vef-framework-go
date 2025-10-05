package csv

import (
	"reflect"

	"github.com/ilxqx/vef-framework-go/tabular"
)

// NewImporter creates a new Importer with the specified type.
func NewImporter(typ reflect.Type, opts ...ImportOption) tabular.Importer {
	return newImporter(typ, opts...)
}

// NewImporterFor creates a new Importer with the specified type T.
func NewImporterFor[T any](opts ...ImportOption) tabular.Importer {
	return newImporter(reflect.TypeFor[T](), opts...)
}

// NewExporter creates a new Exporter with the specified type.
func NewExporter(typ reflect.Type, opts ...ExportOption) tabular.Exporter {
	return newExporter(typ, opts...)
}

// NewExporterFor creates a new Exporter with the specified type T.
func NewExporterFor[T any](opts ...ExportOption) tabular.Exporter {
	return newExporter(reflect.TypeFor[T](), opts...)
}
