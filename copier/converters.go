package copier

import (
	"reflect"

	"github.com/ilxqx/vef-framework-go/decimal"
	"github.com/ilxqx/vef-framework-go/mo"
	"github.com/ilxqx/vef-framework-go/null"
)

var (
	// null.String
	nullStringType  = reflect.TypeFor[null.String]()
	stringType      = reflect.TypeFor[string]()
	stringConverter = TypeConverter{
		SrcType: nullStringType,
		DstType: stringType,
		Fn: func(src any) (any, error) {
			return src.(null.String).ValueOrZero(), nil
		},
	}
	nullStringConverter = TypeConverter{
		SrcType: stringType,
		DstType: nullStringType,
		Fn: func(src any) (any, error) {
			return null.StringFrom(src.(string)), nil
		},
	}

	// null.Int
	nullIntType  = reflect.TypeFor[null.Int]()
	intType      = reflect.TypeFor[int64]()
	intConverter = TypeConverter{
		SrcType: nullIntType,
		DstType: intType,
		Fn: func(src any) (any, error) {
			return src.(null.Int).ValueOrZero(), nil
		},
	}
	nullIntConverter = TypeConverter{
		SrcType: intType,
		DstType: nullIntType,
		Fn: func(src any) (any, error) {
			return null.IntFrom(src.(int64)), nil
		},
	}

	// null.Int16
	nullInt16Type  = reflect.TypeFor[null.Int16]()
	int16Type      = reflect.TypeFor[int16]()
	int16Converter = TypeConverter{
		SrcType: nullInt16Type,
		DstType: int16Type,
		Fn: func(src any) (any, error) {
			return src.(null.Int16).ValueOrZero(), nil
		},
	}
	nullInt16Converter = TypeConverter{
		SrcType: int16Type,
		DstType: nullInt16Type,
		Fn: func(src any) (any, error) {
			return null.Int16From(src.(int16)), nil
		},
	}

	// null.Int32
	nullInt32Type  = reflect.TypeFor[null.Int32]()
	int32Type      = reflect.TypeFor[int32]()
	int32Converter = TypeConverter{
		SrcType: nullInt32Type,
		DstType: int32Type,
		Fn: func(src any) (any, error) {
			return src.(null.Int32).ValueOrZero(), nil
		},
	}
	nullInt32Converter = TypeConverter{
		SrcType: int32Type,
		DstType: nullInt32Type,
		Fn: func(src any) (any, error) {
			return null.Int32From(src.(int32)), nil
		},
	}

	// null.Float
	nullFloatType  = reflect.TypeFor[null.Float]()
	floatType      = reflect.TypeFor[float64]()
	floatConverter = TypeConverter{
		SrcType: nullFloatType,
		DstType: floatType,
		Fn: func(src any) (any, error) {
			return src.(null.Float).ValueOrZero(), nil
		},
	}
	nullFloatConverter = TypeConverter{
		SrcType: floatType,
		DstType: nullFloatType,
		Fn: func(src any) (any, error) {
			return null.FloatFrom(src.(float64)), nil
		},
	}

	// null.Byte
	nullByteType  = reflect.TypeFor[null.Byte]()
	byteType      = reflect.TypeFor[byte]()
	byteConverter = TypeConverter{
		SrcType: nullByteType,
		DstType: byteType,
		Fn: func(src any) (any, error) {
			return src.(null.Byte).ValueOrZero(), nil
		},
	}
	nullByteConverter = TypeConverter{
		SrcType: byteType,
		DstType: nullByteType,
		Fn: func(src any) (any, error) {
			return null.ByteFrom(src.(byte)), nil
		},
	}

	// null.Bool
	nullBoolType  = reflect.TypeFor[null.Bool]()
	boolType      = reflect.TypeFor[bool]()
	boolConverter = TypeConverter{
		SrcType: nullBoolType,
		DstType: boolType,
		Fn: func(src any) (any, error) {
			return src.(null.Bool).ValueOrZero(), nil
		},
	}
	nullBoolConverter = TypeConverter{
		SrcType: boolType,
		DstType: nullBoolType,
		Fn: func(src any) (any, error) {
			return null.BoolFrom(src.(bool)), nil
		},
	}

	// null.DateTime
	nullDateTimeType  = reflect.TypeFor[null.DateTime]()
	dateTimeType      = reflect.TypeFor[mo.DateTime]()
	dateTimeConverter = TypeConverter{
		SrcType: nullDateTimeType,
		DstType: dateTimeType,
		Fn: func(src any) (any, error) {
			return src.(null.DateTime).ValueOrZero(), nil
		},
	}
	nullDateTimeConverter = TypeConverter{
		SrcType: dateTimeType,
		DstType: nullDateTimeType,
		Fn: func(src any) (any, error) {
			return null.DateTimeFrom(src.(mo.DateTime)), nil
		},
	}

	// null.Date
	nullDateType  = reflect.TypeFor[null.Date]()
	dateType      = reflect.TypeFor[mo.Date]()
	dateConverter = TypeConverter{
		SrcType: nullDateType,
		DstType: dateType,
		Fn: func(src any) (any, error) {
			return src.(null.Date).ValueOrZero(), nil
		},
	}
	nullDateConverter = TypeConverter{
		SrcType: dateType,
		DstType: nullDateType,
		Fn: func(src any) (any, error) {
			return null.DateFrom(src.(mo.Date)), nil
		},
	}

	// null.Time
	nullTimeType  = reflect.TypeFor[null.Time]()
	timeType      = reflect.TypeFor[mo.Time]()
	timeConverter = TypeConverter{
		SrcType: nullTimeType,
		DstType: timeType,
		Fn: func(src any) (any, error) {
			return src.(null.Time).ValueOrZero(), nil
		},
	}
	nullTimeConverter = TypeConverter{
		SrcType: timeType,
		DstType: nullTimeType,
		Fn: func(src any) (any, error) {
			return null.TimeFrom(src.(mo.Time)), nil
		},
	}

	// null.Decimal
	nullDecimalType  = reflect.TypeFor[null.Decimal]()
	decimalType      = reflect.TypeFor[decimal.Decimal]()
	decimalConverter = TypeConverter{
		SrcType: nullDecimalType,
		DstType: decimalType,
		Fn: func(src any) (any, error) {
			return src.(null.Decimal).ValueOrZero(), nil
		},
	}
	nullDecimalConverter = TypeConverter{
		SrcType: decimalType,
		DstType: nullDecimalType,
		Fn: func(src any) (any, error) {
			return null.DecimalFrom(src.(decimal.Decimal)), nil
		},
	}
)
