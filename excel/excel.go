package excel

import (
	"reflect"

	"github.com/ilxqx/vef-framework-go/tabular"
)

func NewImporter(typ reflect.Type, opts ...ImportOption) tabular.Importer {
	return newImporter(typ, opts...)
}

func NewImporterFor[T any](opts ...ImportOption) tabular.Importer {
	return newImporter(reflect.TypeFor[T](), opts...)
}

func NewExporter(typ reflect.Type, opts ...ExportOption) tabular.Exporter {
	return newExporter(typ, opts...)
}

func NewExporterFor[T any](opts ...ExportOption) tabular.Exporter {
	return newExporter(reflect.TypeFor[T](), opts...)
}
