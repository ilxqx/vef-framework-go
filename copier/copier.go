package copier

import "github.com/jinzhu/copier"

// converters is the default converters for copier
var converters = []TypeConverter{
	// null.String converters
	stringToNullStringConverter,
	stringPtrToNullStringConverter,
	nullStringToStringConverter,
	nullStringToStringPtrConverter,

	// null.Int converters
	intToNullIntConverter,
	intPtrToNullIntConverter,
	nullIntToIntConverter,
	nullIntToIntPtrConverter,

	// null.Int16 converters
	int16ToNullInt16Converter,
	int16PtrToNullInt16Converter,
	nullInt16ToInt16Converter,
	nullInt16ToInt16PtrConverter,

	// null.Int32 converters
	int32ToNullInt32Converter,
	int32PtrToNullInt32Converter,
	nullInt32ToInt32Converter,
	nullInt32ToInt32PtrConverter,

	// null.Float converters
	floatToNullFloatConverter,
	floatPtrToNullFloatConverter,
	nullFloatToFloatConverter,
	nullFloatToFloatPtrConverter,

	// null.Byte converters
	byteToNullByteConverter,
	bytePtrToNullByteConverter,
	nullByteToByteConverter,
	nullByteToBytePtrConverter,

	// null.Bool converters
	boolToNullBoolConverter,
	boolPtrToNullBoolConverter,
	nullBoolToBoolConverter,
	nullBoolToBoolPtrConverter,

	// null.DateTime converters
	dateTimeToNullDateTimeConverter,
	dateTimePtrToNullDateTimeConverter,
	nullDateTimeToDateTimeConverter,
	nullDateTimeToDateTimePtrConverter,

	// null.Date converters
	dateToNullDateConverter,
	datePtrToNullDateConverter,
	nullDateToDateConverter,
	nullDateToDatePtrConverter,

	// null.Time converters
	timeToNullTimeConverter,
	timePtrToNullTimeConverter,
	nullTimeToTimeConverter,
	nullTimeToTimePtrConverter,

	// null.Decimal converters
	decimalToNullDecimalConverter,
	decimalPtrToNullDecimalConverter,
	nullDecimalToDecimalConverter,
	nullDecimalToDecimalPtrConverter,
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
