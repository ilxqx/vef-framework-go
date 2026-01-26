package copier

import (
	"github.com/samber/lo"

	"github.com/ilxqx/vef-framework-go/datetime"
	"github.com/ilxqx/vef-framework-go/decimal"
	"github.com/ilxqx/vef-framework-go/null"
)

// Nullable defines the interface for null wrapper types.
type Nullable[T any] interface {
	ValueOrZero() T
	Ptr() *T
}

// makeNullToValueConverter creates a converter from null type to value type.
func makeNullToValueConverter[N Nullable[T], T any]() TypeConverter {
	return TypeConverter{
		SrcType: lo.Empty[N](),
		DstType: lo.Empty[T](),
		Fn: func(src any) (any, error) {
			return src.(N).ValueOrZero(), nil
		},
	}
}

// makeNullToPtrConverter creates a converter from null type to pointer type.
func makeNullToPtrConverter[N Nullable[T], T any]() TypeConverter {
	return TypeConverter{
		SrcType: lo.Empty[N](),
		DstType: lo.Empty[*T](),
		Fn: func(src any) (any, error) {
			return src.(N).Ptr(), nil
		},
	}
}

// makeValueToNullConverter creates a converter from value type to null type.
func makeValueToNullConverter[T, N any](fromFn func(T) N) TypeConverter {
	return TypeConverter{
		SrcType: lo.Empty[T](),
		DstType: lo.Empty[N](),
		Fn: func(src any) (any, error) {
			return fromFn(src.(T)), nil
		},
	}
}

// makePtrToNullConverter creates a converter from pointer type to null type.
func makePtrToNullConverter[T, N any](fromPtrFn func(*T) N) TypeConverter {
	return TypeConverter{
		SrcType: lo.Empty[*T](),
		DstType: lo.Empty[N](),
		Fn: func(src any) (any, error) {
			return fromPtrFn(src.(*T)), nil
		},
	}
}

var (
	// null.String converters
	nullStringToStringConverter    = makeNullToValueConverter[null.String, string]()
	nullStringToStringPtrConverter = makeNullToPtrConverter[null.String, string]()
	stringToNullStringConverter    = makeValueToNullConverter(null.StringFrom)
	stringPtrToNullStringConverter = makePtrToNullConverter(null.StringFromPtr)

	// null.Int converters
	nullIntToIntConverter    = makeNullToValueConverter[null.Int, int64]()
	nullIntToIntPtrConverter = makeNullToPtrConverter[null.Int, int64]()
	intToNullIntConverter    = makeValueToNullConverter(null.IntFrom)
	intPtrToNullIntConverter = makePtrToNullConverter(null.IntFromPtr)

	// null.Int16 converters
	nullInt16ToInt16Converter    = makeNullToValueConverter[null.Int16, int16]()
	nullInt16ToInt16PtrConverter = makeNullToPtrConverter[null.Int16, int16]()
	int16ToNullInt16Converter    = makeValueToNullConverter(null.Int16From)
	int16PtrToNullInt16Converter = makePtrToNullConverter(null.Int16FromPtr)

	// null.Int32 converters
	nullInt32ToInt32Converter    = makeNullToValueConverter[null.Int32, int32]()
	nullInt32ToInt32PtrConverter = makeNullToPtrConverter[null.Int32, int32]()
	int32ToNullInt32Converter    = makeValueToNullConverter(null.Int32From)
	int32PtrToNullInt32Converter = makePtrToNullConverter(null.Int32FromPtr)

	// null.Float converters
	nullFloatToFloatConverter    = makeNullToValueConverter[null.Float, float64]()
	nullFloatToFloatPtrConverter = makeNullToPtrConverter[null.Float, float64]()
	floatToNullFloatConverter    = makeValueToNullConverter(null.FloatFrom)
	floatPtrToNullFloatConverter = makePtrToNullConverter(null.FloatFromPtr)

	// null.Byte converters
	nullByteToByteConverter    = makeNullToValueConverter[null.Byte, byte]()
	nullByteToBytePtrConverter = makeNullToPtrConverter[null.Byte, byte]()
	byteToNullByteConverter    = makeValueToNullConverter(null.ByteFrom)
	bytePtrToNullByteConverter = makePtrToNullConverter(null.ByteFromPtr)

	// null.Bool converters
	nullBoolToBoolConverter    = makeNullToValueConverter[null.Bool, bool]()
	nullBoolToBoolPtrConverter = makeNullToPtrConverter[null.Bool, bool]()
	boolToNullBoolConverter    = makeValueToNullConverter(null.BoolFrom)
	boolPtrToNullBoolConverter = makePtrToNullConverter(null.BoolFromPtr)

	// null.DateTime converters
	nullDateTimeToDateTimeConverter    = makeNullToValueConverter[null.DateTime, datetime.DateTime]()
	nullDateTimeToDateTimePtrConverter = makeNullToPtrConverter[null.DateTime, datetime.DateTime]()
	dateTimeToNullDateTimeConverter    = makeValueToNullConverter(null.DateTimeFrom)
	dateTimePtrToNullDateTimeConverter = makePtrToNullConverter(null.DateTimeFromPtr)

	// null.Date converters
	nullDateToDateConverter    = makeNullToValueConverter[null.Date, datetime.Date]()
	nullDateToDatePtrConverter = makeNullToPtrConverter[null.Date, datetime.Date]()
	dateToNullDateConverter    = makeValueToNullConverter(null.DateFrom)
	datePtrToNullDateConverter = makePtrToNullConverter(null.DateFromPtr)

	// null.Time converters
	nullTimeToTimeConverter    = makeNullToValueConverter[null.Time, datetime.Time]()
	nullTimeToTimePtrConverter = makeNullToPtrConverter[null.Time, datetime.Time]()
	timeToNullTimeConverter    = makeValueToNullConverter(null.TimeFrom)
	timePtrToNullTimeConverter = makePtrToNullConverter(null.TimeFromPtr)

	// null.Decimal converters
	nullDecimalToDecimalConverter    = makeNullToValueConverter[null.Decimal, decimal.Decimal]()
	nullDecimalToDecimalPtrConverter = makeNullToPtrConverter[null.Decimal, decimal.Decimal]()
	decimalToNullDecimalConverter    = makeValueToNullConverter(null.DecimalFrom)
	decimalPtrToNullDecimalConverter = makePtrToNullConverter(null.DecimalFromPtr)
)
