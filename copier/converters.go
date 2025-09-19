package copier

import (
	"reflect"

	"github.com/ilxqx/vef-framework-go/decimal"
	"github.com/ilxqx/vef-framework-go/mo"
	"github.com/ilxqx/vef-framework-go/null"
)

var (
	// null.String
	nullStringType              = reflect.TypeFor[null.String]()
	stringType                  = reflect.TypeFor[string]()
	stringPtrType               = reflect.TypeFor[*string]()
	nullStringToStringConverter = TypeConverter{
		SrcType: nullStringType,
		DstType: stringType,
		Fn: func(src any) (any, error) {
			return src.(null.String).ValueOrZero(), nil
		},
	}
	nullStringToStringPtrConverter = TypeConverter{
		SrcType: nullStringType,
		DstType: stringPtrType,
		Fn: func(src any) (any, error) {
			return src.(null.String).Ptr(), nil
		},
	}
	stringToNullStringConverter = TypeConverter{
		SrcType: stringType,
		DstType: nullStringType,
		Fn: func(src any) (any, error) {
			return null.StringFrom(src.(string)), nil
		},
	}
	stringPtrToNullStringConverter = TypeConverter{
		SrcType: stringPtrType,
		DstType: nullStringType,
		Fn: func(src any) (any, error) {
			return null.StringFromPtr(src.(*string)), nil
		},
	}

	// null.Int
	nullIntType           = reflect.TypeFor[null.Int]()
	intType               = reflect.TypeFor[int64]()
	intPtrType            = reflect.TypeFor[*int64]()
	nullIntToIntConverter = TypeConverter{
		SrcType: nullIntType,
		DstType: intType,
		Fn: func(src any) (any, error) {
			return src.(null.Int).ValueOrZero(), nil
		},
	}
	nullIntToIntPtrConverter = TypeConverter{
		SrcType: nullIntType,
		DstType: intPtrType,
		Fn: func(src any) (any, error) {
			return src.(null.Int).Ptr(), nil
		},
	}
	intToNullIntConverter = TypeConverter{
		SrcType: intType,
		DstType: nullIntType,
		Fn: func(src any) (any, error) {
			return null.IntFrom(src.(int64)), nil
		},
	}
	intPtrToNullIntConverter = TypeConverter{
		SrcType: intPtrType,
		DstType: nullIntType,
		Fn: func(src any) (any, error) {
			return null.IntFromPtr(src.(*int64)), nil
		},
	}

	// null.Int16
	nullInt16Type             = reflect.TypeFor[null.Int16]()
	int16Type                 = reflect.TypeFor[int16]()
	int16PtrType              = reflect.TypeFor[*int16]()
	nullInt16ToInt16Converter = TypeConverter{
		SrcType: nullInt16Type,
		DstType: int16Type,
		Fn: func(src any) (any, error) {
			return src.(null.Int16).ValueOrZero(), nil
		},
	}
	nullInt16ToInt16PtrConverter = TypeConverter{
		SrcType: nullInt16Type,
		DstType: int16PtrType,
		Fn: func(src any) (any, error) {
			return src.(null.Int16).Ptr(), nil
		},
	}
	int16ToNullInt16Converter = TypeConverter{
		SrcType: int16Type,
		DstType: nullInt16Type,
		Fn: func(src any) (any, error) {
			return null.Int16From(src.(int16)), nil
		},
	}
	int16PtrToNullInt16Converter = TypeConverter{
		SrcType: int16PtrType,
		DstType: nullInt16Type,
		Fn: func(src any) (any, error) {
			return null.Int16FromPtr(src.(*int16)), nil
		},
	}

	// null.Int32
	nullInt32Type             = reflect.TypeFor[null.Int32]()
	int32Type                 = reflect.TypeFor[int32]()
	int32PtrType              = reflect.TypeFor[*int32]()
	nullInt32ToInt32Converter = TypeConverter{
		SrcType: nullInt32Type,
		DstType: int32Type,
		Fn: func(src any) (any, error) {
			return src.(null.Int32).ValueOrZero(), nil
		},
	}
	nullInt32ToInt32PtrConverter = TypeConverter{
		SrcType: nullInt32Type,
		DstType: int32PtrType,
		Fn: func(src any) (any, error) {
			return src.(null.Int32).Ptr(), nil
		},
	}
	int32ToNullInt32Converter = TypeConverter{
		SrcType: int32Type,
		DstType: nullInt32Type,
		Fn: func(src any) (any, error) {
			return null.Int32From(src.(int32)), nil
		},
	}
	int32PtrToNullInt32Converter = TypeConverter{
		SrcType: int32PtrType,
		DstType: nullInt32Type,
		Fn: func(src any) (any, error) {
			return null.Int32FromPtr(src.(*int32)), nil
		},
	}

	// null.Float
	nullFloatType             = reflect.TypeFor[null.Float]()
	floatType                 = reflect.TypeFor[float64]()
	floatPtrType              = reflect.TypeFor[*float64]()
	nullFloatToFloatConverter = TypeConverter{
		SrcType: nullFloatType,
		DstType: floatType,
		Fn: func(src any) (any, error) {
			return src.(null.Float).ValueOrZero(), nil
		},
	}
	nullFloatToFloatPtrConverter = TypeConverter{
		SrcType: nullFloatType,
		DstType: floatPtrType,
		Fn: func(src any) (any, error) {
			return src.(null.Float).Ptr(), nil
		},
	}
	floatToNullFloatConverter = TypeConverter{
		SrcType: floatType,
		DstType: nullFloatType,
		Fn: func(src any) (any, error) {
			return null.FloatFrom(src.(float64)), nil
		},
	}
	floatPtrToNullFloatConverter = TypeConverter{
		SrcType: floatPtrType,
		DstType: nullFloatType,
		Fn: func(src any) (any, error) {
			return null.FloatFromPtr(src.(*float64)), nil
		},
	}

	// null.Byte
	nullByteType            = reflect.TypeFor[null.Byte]()
	byteType                = reflect.TypeFor[byte]()
	bytePtrType             = reflect.TypeFor[*byte]()
	nullByteToByteConverter = TypeConverter{
		SrcType: nullByteType,
		DstType: byteType,
		Fn: func(src any) (any, error) {
			return src.(null.Byte).ValueOrZero(), nil
		},
	}
	nullByteToBytePtrConverter = TypeConverter{
		SrcType: nullByteType,
		DstType: bytePtrType,
		Fn: func(src any) (any, error) {
			return src.(null.Byte).Ptr(), nil
		},
	}
	byteToNullByteConverter = TypeConverter{
		SrcType: byteType,
		DstType: nullByteType,
		Fn: func(src any) (any, error) {
			return null.ByteFrom(src.(byte)), nil
		},
	}
	bytePtrToNullByteConverter = TypeConverter{
		SrcType: bytePtrType,
		DstType: nullByteType,
		Fn: func(src any) (any, error) {
			return null.ByteFromPtr(src.(*byte)), nil
		},
	}

	// null.Bool
	nullBoolType            = reflect.TypeFor[null.Bool]()
	boolType                = reflect.TypeFor[bool]()
	boolPtrType             = reflect.TypeFor[*bool]()
	nullBoolToBoolConverter = TypeConverter{
		SrcType: nullBoolType,
		DstType: boolType,
		Fn: func(src any) (any, error) {
			return src.(null.Bool).ValueOrZero(), nil
		},
	}
	nullBoolToBoolPtrConverter = TypeConverter{
		SrcType: nullBoolType,
		DstType: boolPtrType,
		Fn: func(src any) (any, error) {
			return src.(null.Bool).Ptr(), nil
		},
	}
	boolToNullBoolConverter = TypeConverter{
		SrcType: boolType,
		DstType: nullBoolType,
		Fn: func(src any) (any, error) {
			return null.BoolFrom(src.(bool)), nil
		},
	}
	boolPtrToNullBoolConverter = TypeConverter{
		SrcType: boolPtrType,
		DstType: nullBoolType,
		Fn: func(src any) (any, error) {
			return null.BoolFromPtr(src.(*bool)), nil
		},
	}

	// null.DateTime
	nullDateTimeType                = reflect.TypeFor[null.DateTime]()
	dateTimeType                    = reflect.TypeFor[mo.DateTime]()
	dateTimePtrType                 = reflect.TypeFor[*mo.DateTime]()
	nullDateTimeToDateTimeConverter = TypeConverter{
		SrcType: nullDateTimeType,
		DstType: dateTimeType,
		Fn: func(src any) (any, error) {
			return src.(null.DateTime).ValueOrZero(), nil
		},
	}
	nullDateTimeToDateTimePtrConverter = TypeConverter{
		SrcType: nullDateTimeType,
		DstType: dateTimePtrType,
		Fn: func(src any) (any, error) {
			return src.(null.DateTime).Ptr(), nil
		},
	}
	dateTimeToNullDateTimeConverter = TypeConverter{
		SrcType: dateTimeType,
		DstType: nullDateTimeType,
		Fn: func(src any) (any, error) {
			return null.DateTimeFrom(src.(mo.DateTime)), nil
		},
	}
	dateTimePtrToNullDateTimeConverter = TypeConverter{
		SrcType: dateTimePtrType,
		DstType: nullDateTimeType,
		Fn: func(src any) (any, error) {
			return null.DateTimeFromPtr(src.(*mo.DateTime)), nil
		},
	}

	// null.Date
	nullDateType            = reflect.TypeFor[null.Date]()
	dateType                = reflect.TypeFor[mo.Date]()
	datePtrType             = reflect.TypeFor[*mo.Date]()
	nullDateToDateConverter = TypeConverter{
		SrcType: nullDateType,
		DstType: dateType,
		Fn: func(src any) (any, error) {
			return src.(null.Date).ValueOrZero(), nil
		},
	}
	nullDateToDatePtrConverter = TypeConverter{
		SrcType: nullDateType,
		DstType: datePtrType,
		Fn: func(src any) (any, error) {
			return src.(null.Date).Ptr(), nil
		},
	}
	dateToNullDateConverter = TypeConverter{
		SrcType: dateType,
		DstType: nullDateType,
		Fn: func(src any) (any, error) {
			return null.DateFrom(src.(mo.Date)), nil
		},
	}
	datePtrToNullDateConverter = TypeConverter{
		SrcType: datePtrType,
		DstType: nullDateType,
		Fn: func(src any) (any, error) {
			return null.DateFromPtr(src.(*mo.Date)), nil
		},
	}

	// null.Time
	nullTimeType            = reflect.TypeFor[null.Time]()
	timeType                = reflect.TypeFor[mo.Time]()
	timePtrType             = reflect.TypeFor[*mo.Time]()
	nullTimeToTimeConverter = TypeConverter{
		SrcType: nullTimeType,
		DstType: timeType,
		Fn: func(src any) (any, error) {
			return src.(null.Time).ValueOrZero(), nil
		},
	}
	nullTimeToTimePtrConverter = TypeConverter{
		SrcType: nullTimeType,
		DstType: timePtrType,
		Fn: func(src any) (any, error) {
			return src.(null.Time).Ptr(), nil
		},
	}
	timeToNullTimeConverter = TypeConverter{
		SrcType: timeType,
		DstType: nullTimeType,
		Fn: func(src any) (any, error) {
			return null.TimeFrom(src.(mo.Time)), nil
		},
	}
	timePtrToNullTimeConverter = TypeConverter{
		SrcType: timePtrType,
		DstType: nullTimeType,
		Fn: func(src any) (any, error) {
			return null.TimeFromPtr(src.(*mo.Time)), nil
		},
	}

	// null.Decimal
	nullDecimalType               = reflect.TypeFor[null.Decimal]()
	decimalType                   = reflect.TypeFor[decimal.Decimal]()
	decimalPtrType                = reflect.TypeFor[*decimal.Decimal]()
	nullDecimalToDecimalConverter = TypeConverter{
		SrcType: nullDecimalType,
		DstType: decimalType,
		Fn: func(src any) (any, error) {
			return src.(null.Decimal).ValueOrZero(), nil
		},
	}
	nullDecimalToDecimalPtrConverter = TypeConverter{
		SrcType: nullDecimalType,
		DstType: decimalPtrType,
		Fn: func(src any) (any, error) {
			return src.(null.Decimal).Ptr(), nil
		},
	}
	decimalToNullDecimalConverter = TypeConverter{
		SrcType: decimalType,
		DstType: nullDecimalType,
		Fn: func(src any) (any, error) {
			return null.DecimalFrom(src.(decimal.Decimal)), nil
		},
	}
	decimalPtrToNullDecimalConverter = TypeConverter{
		SrcType: decimalPtrType,
		DstType: nullDecimalType,
		Fn: func(src any) (any, error) {
			return null.DecimalFromPtr(src.(*decimal.Decimal)), nil
		},
	}
)
