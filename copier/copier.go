package copier

import "github.com/jinzhu/copier"

// converters is the default converters for copier
var converters = []TypeConverter{
	nullStringConverter,
	stringConverter,
	nullIntConverter,
	intConverter,
	nullInt16Converter,
	int16Converter,
	nullInt32Converter,
	int32Converter,
	nullFloatConverter,
	floatConverter,
	nullByteConverter,
	byteConverter,
	nullBoolConverter,
	boolConverter,
	nullDateTimeConverter,
	dateTimeConverter,
	nullDateConverter,
	dateConverter,
	nullTimeConverter,
	timeConverter,
	nullDecimalConverter,
	decimalConverter,
}

type (
	copyOption    func(option *copier.Option)
	TypeConverter = copier.TypeConverter
)

// WithIgnoreEmpty ignore empty fields
func WithIgnoreEmpty() copyOption {
	return func(option *copier.Option) {
		option.IgnoreEmpty = true
	}
}

// WithDeepCopy deep copy fields
func WithDeepCopy() copyOption {
	return func(option *copier.Option) {
		option.DeepCopy = true
	}
}

// WithCaseInsensitive case-insensitive
func WithCaseInsensitive() copyOption {
	return func(option *copier.Option) {
		option.CaseSensitive = false
	}
}

// WithFieldNameMapping field name mapping
func WithFieldNameMapping(fieldMapping ...copier.FieldNameMapping) copyOption {
	return func(option *copier.Option) {
		option.FieldNameMapping = append(option.FieldNameMapping, fieldMapping...)
	}
}

// WithTypeConverters sets the type converters
func WithTypeConverters(converters ...TypeConverter) copyOption {
	return func(option *copier.Option) {
		option.Converters = append(option.Converters, converters...)
	}
}

// Copy src to dst
func Copy(src any, dst any, options ...copyOption) error {
	option := copier.Option{
		CaseSensitive: true,
		Converters:    converters,
	}
	for _, opt := range options {
		opt(&option)
	}

	return copier.CopyWithOption(dst, src, option)
}
