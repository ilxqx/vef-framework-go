package copier

import (
	"github.com/ilxqx/vef-framework-go/datetime"
	"github.com/ilxqx/vef-framework-go/decimal"
	"github.com/ilxqx/vef-framework-go/null"
	"github.com/samber/lo"
)

var (
	// Null.String.
	nullStringToStringConverter = TypeConverter{
		SrcType: lo.Empty[null.String](),
		DstType: lo.Empty[string](),
		Fn: func(src any) (any, error) {
			return src.(null.String).ValueOrZero(), nil
		},
	}
	nullStringToStringPtrConverter = TypeConverter{
		SrcType: lo.Empty[null.String](),
		DstType: lo.Empty[*string](),
		Fn: func(src any) (any, error) {
			return src.(null.String).Ptr(), nil
		},
	}
	stringToNullStringConverter = TypeConverter{
		SrcType: lo.Empty[string](),
		DstType: lo.Empty[null.String](),
		Fn: func(src any) (any, error) {
			return null.StringFrom(src.(string)), nil
		},
	}
	stringPtrToNullStringConverter = TypeConverter{
		SrcType: lo.Empty[*string](),
		DstType: lo.Empty[null.String](),
		Fn: func(src any) (any, error) {
			return null.StringFromPtr(src.(*string)), nil
		},
	}

	// Null.Int.
	nullIntToIntConverter = TypeConverter{
		SrcType: lo.Empty[null.Int](),
		DstType: lo.Empty[int64](),
		Fn: func(src any) (any, error) {
			return src.(null.Int).ValueOrZero(), nil
		},
	}
	nullIntToIntPtrConverter = TypeConverter{
		SrcType: lo.Empty[null.Int](),
		DstType: lo.Empty[*int64](),
		Fn: func(src any) (any, error) {
			return src.(null.Int).Ptr(), nil
		},
	}
	intToNullIntConverter = TypeConverter{
		SrcType: lo.Empty[int64](),
		DstType: lo.Empty[null.Int](),
		Fn: func(src any) (any, error) {
			return null.IntFrom(src.(int64)), nil
		},
	}
	intPtrToNullIntConverter = TypeConverter{
		SrcType: lo.Empty[*int64](),
		DstType: lo.Empty[null.Int](),
		Fn: func(src any) (any, error) {
			return null.IntFromPtr(src.(*int64)), nil
		},
	}

	// Null.Int16.
	nullInt16ToInt16Converter = TypeConverter{
		SrcType: lo.Empty[null.Int16](),
		DstType: lo.Empty[int16](),
		Fn: func(src any) (any, error) {
			return src.(null.Int16).ValueOrZero(), nil
		},
	}
	nullInt16ToInt16PtrConverter = TypeConverter{
		SrcType: lo.Empty[null.Int16](),
		DstType: lo.Empty[*int16](),
		Fn: func(src any) (any, error) {
			return src.(null.Int16).Ptr(), nil
		},
	}
	int16ToNullInt16Converter = TypeConverter{
		SrcType: lo.Empty[int16](),
		DstType: lo.Empty[null.Int16](),
		Fn: func(src any) (any, error) {
			return null.Int16From(src.(int16)), nil
		},
	}
	int16PtrToNullInt16Converter = TypeConverter{
		SrcType: lo.Empty[*int16](),
		DstType: lo.Empty[null.Int16](),
		Fn: func(src any) (any, error) {
			return null.Int16FromPtr(src.(*int16)), nil
		},
	}

	// Null.Int32.
	nullInt32ToInt32Converter = TypeConverter{
		SrcType: lo.Empty[null.Int32](),
		DstType: lo.Empty[int32](),
		Fn: func(src any) (any, error) {
			return src.(null.Int32).ValueOrZero(), nil
		},
	}
	nullInt32ToInt32PtrConverter = TypeConverter{
		SrcType: lo.Empty[null.Int32](),
		DstType: lo.Empty[*int32](),
		Fn: func(src any) (any, error) {
			return src.(null.Int32).Ptr(), nil
		},
	}
	int32ToNullInt32Converter = TypeConverter{
		SrcType: lo.Empty[int32](),
		DstType: lo.Empty[null.Int32](),
		Fn: func(src any) (any, error) {
			return null.Int32From(src.(int32)), nil
		},
	}
	int32PtrToNullInt32Converter = TypeConverter{
		SrcType: lo.Empty[*int32](),
		DstType: lo.Empty[null.Int32](),
		Fn: func(src any) (any, error) {
			return null.Int32FromPtr(src.(*int32)), nil
		},
	}

	// Null.Float.
	nullFloatToFloatConverter = TypeConverter{
		SrcType: lo.Empty[null.Float](),
		DstType: lo.Empty[float64](),
		Fn: func(src any) (any, error) {
			return src.(null.Float).ValueOrZero(), nil
		},
	}
	nullFloatToFloatPtrConverter = TypeConverter{
		SrcType: lo.Empty[null.Float](),
		DstType: lo.Empty[*float64](),
		Fn: func(src any) (any, error) {
			return src.(null.Float).Ptr(), nil
		},
	}
	floatToNullFloatConverter = TypeConverter{
		SrcType: lo.Empty[float64](),
		DstType: lo.Empty[null.Float](),
		Fn: func(src any) (any, error) {
			return null.FloatFrom(src.(float64)), nil
		},
	}
	floatPtrToNullFloatConverter = TypeConverter{
		SrcType: lo.Empty[*float64](),
		DstType: lo.Empty[null.Float](),
		Fn: func(src any) (any, error) {
			return null.FloatFromPtr(src.(*float64)), nil
		},
	}

	// Null.Byte.
	nullByteToByteConverter = TypeConverter{
		SrcType: lo.Empty[null.Byte](),
		DstType: lo.Empty[byte](),
		Fn: func(src any) (any, error) {
			return src.(null.Byte).ValueOrZero(), nil
		},
	}
	nullByteToBytePtrConverter = TypeConverter{
		SrcType: lo.Empty[null.Byte](),
		DstType: lo.Empty[*byte](),
		Fn: func(src any) (any, error) {
			return src.(null.Byte).Ptr(), nil
		},
	}
	byteToNullByteConverter = TypeConverter{
		SrcType: lo.Empty[byte](),
		DstType: lo.Empty[null.Byte](),
		Fn: func(src any) (any, error) {
			return null.ByteFrom(src.(byte)), nil
		},
	}
	bytePtrToNullByteConverter = TypeConverter{
		SrcType: lo.Empty[*byte](),
		DstType: lo.Empty[null.Byte](),
		Fn: func(src any) (any, error) {
			return null.ByteFromPtr(src.(*byte)), nil
		},
	}

	// Null.Bool.
	nullBoolToBoolConverter = TypeConverter{
		SrcType: lo.Empty[null.Bool](),
		DstType: lo.Empty[bool](),
		Fn: func(src any) (any, error) {
			return src.(null.Bool).ValueOrZero(), nil
		},
	}
	nullBoolToBoolPtrConverter = TypeConverter{
		SrcType: lo.Empty[null.Bool](),
		DstType: lo.Empty[*bool](),
		Fn: func(src any) (any, error) {
			return src.(null.Bool).Ptr(), nil
		},
	}
	boolToNullBoolConverter = TypeConverter{
		SrcType: lo.Empty[bool](),
		DstType: lo.Empty[null.Bool](),
		Fn: func(src any) (any, error) {
			return null.BoolFrom(src.(bool)), nil
		},
	}
	boolPtrToNullBoolConverter = TypeConverter{
		SrcType: lo.Empty[*bool](),
		DstType: lo.Empty[null.Bool](),
		Fn: func(src any) (any, error) {
			return null.BoolFromPtr(src.(*bool)), nil
		},
	}

	// Null.DateTime.
	nullDateTimeToDateTimeConverter = TypeConverter{
		SrcType: lo.Empty[null.DateTime](),
		DstType: lo.Empty[datetime.DateTime](),
		Fn: func(src any) (any, error) {
			return src.(null.DateTime).ValueOrZero(), nil
		},
	}
	nullDateTimeToDateTimePtrConverter = TypeConverter{
		SrcType: lo.Empty[null.DateTime](),
		DstType: lo.Empty[*datetime.DateTime](),
		Fn: func(src any) (any, error) {
			return src.(null.DateTime).Ptr(), nil
		},
	}
	dateTimeToNullDateTimeConverter = TypeConverter{
		SrcType: lo.Empty[datetime.DateTime](),
		DstType: lo.Empty[null.DateTime](),
		Fn: func(src any) (any, error) {
			return null.DateTimeFrom(src.(datetime.DateTime)), nil
		},
	}
	dateTimePtrToNullDateTimeConverter = TypeConverter{
		SrcType: lo.Empty[*datetime.DateTime](),
		DstType: lo.Empty[null.DateTime](),
		Fn: func(src any) (any, error) {
			return null.DateTimeFromPtr(src.(*datetime.DateTime)), nil
		},
	}

	// Null.Date.
	nullDateToDateConverter = TypeConverter{
		SrcType: lo.Empty[null.Date](),
		DstType: lo.Empty[datetime.Date](),
		Fn: func(src any) (any, error) {
			return src.(null.Date).ValueOrZero(), nil
		},
	}
	nullDateToDatePtrConverter = TypeConverter{
		SrcType: lo.Empty[null.Date](),
		DstType: lo.Empty[*datetime.Date](),
		Fn: func(src any) (any, error) {
			return src.(null.Date).Ptr(), nil
		},
	}
	dateToNullDateConverter = TypeConverter{
		SrcType: lo.Empty[datetime.Date](),
		DstType: lo.Empty[null.Date](),
		Fn: func(src any) (any, error) {
			return null.DateFrom(src.(datetime.Date)), nil
		},
	}
	datePtrToNullDateConverter = TypeConverter{
		SrcType: lo.Empty[*datetime.Date](),
		DstType: lo.Empty[null.Date](),
		Fn: func(src any) (any, error) {
			return null.DateFromPtr(src.(*datetime.Date)), nil
		},
	}

	// Null.Time.
	nullTimeToTimeConverter = TypeConverter{
		SrcType: lo.Empty[null.Time](),
		DstType: lo.Empty[datetime.Time](),
		Fn: func(src any) (any, error) {
			return src.(null.Time).ValueOrZero(), nil
		},
	}
	nullTimeToTimePtrConverter = TypeConverter{
		SrcType: lo.Empty[null.Time](),
		DstType: lo.Empty[*datetime.Time](),
		Fn: func(src any) (any, error) {
			return src.(null.Time).Ptr(), nil
		},
	}
	timeToNullTimeConverter = TypeConverter{
		SrcType: lo.Empty[datetime.Time](),
		DstType: lo.Empty[null.Time](),
		Fn: func(src any) (any, error) {
			return null.TimeFrom(src.(datetime.Time)), nil
		},
	}
	timePtrToNullTimeConverter = TypeConverter{
		SrcType: lo.Empty[*datetime.Time](),
		DstType: lo.Empty[null.Time](),
		Fn: func(src any) (any, error) {
			return null.TimeFromPtr(src.(*datetime.Time)), nil
		},
	}

	// Null.Decimal.
	nullDecimalToDecimalConverter = TypeConverter{
		SrcType: lo.Empty[null.Decimal](),
		DstType: lo.Empty[decimal.Decimal](),
		Fn: func(src any) (any, error) {
			return src.(null.Decimal).ValueOrZero(), nil
		},
	}
	nullDecimalToDecimalPtrConverter = TypeConverter{
		SrcType: lo.Empty[null.Decimal](),
		DstType: lo.Empty[*decimal.Decimal](),
		Fn: func(src any) (any, error) {
			return src.(null.Decimal).Ptr(), nil
		},
	}
	decimalToNullDecimalConverter = TypeConverter{
		SrcType: lo.Empty[decimal.Decimal](),
		DstType: lo.Empty[null.Decimal](),
		Fn: func(src any) (any, error) {
			return null.DecimalFrom(src.(decimal.Decimal)), nil
		},
	}
	decimalPtrToNullDecimalConverter = TypeConverter{
		SrcType: lo.Empty[*decimal.Decimal](),
		DstType: lo.Empty[null.Decimal](),
		Fn: func(src any) (any, error) {
			return null.DecimalFromPtr(src.(*decimal.Decimal)), nil
		},
	}
)
