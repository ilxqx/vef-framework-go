package excel

import "reflect"

// NewImporter creates a new Importer with the specified type.
func NewImporter(typ reflect.Type, opts ...ImportOption) Importer {
	return newDefaultImporter(typ, opts...)
}

// NewImporterFor creates a new Importer with the specified type T.
func NewImporterFor[T any](opts ...ImportOption) Importer {
	return newDefaultImporter(reflect.TypeFor[T](), opts...)
}

// NewExporter creates a new Exporter with the specified type.
func NewExporter(typ reflect.Type, opts ...ExportOption) Exporter {
	return newDefaultExporter(typ, opts...)
}

// NewExporterFor creates a new Exporter with the specified type T.
func NewExporterFor[T any](opts ...ExportOption) Exporter {
	return newDefaultExporter(reflect.TypeFor[T](), opts...)
}
