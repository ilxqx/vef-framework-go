package mapx

import (
	"mime/multipart"
	"reflect"

	"github.com/samber/lo"

	"github.com/ilxqx/vef-framework-go/datetime"
	"github.com/ilxqx/vef-framework-go/decimal"
	"github.com/ilxqx/vef-framework-go/null"
)

var (
	// Null.Bool.
	nullBoolType = reflect.TypeFor[null.Bool]()
	boolType     = reflect.TypeFor[bool]()
	boolPtrType  = reflect.TypeFor[*bool]()

	// Null.String.
	nullStringType = reflect.TypeFor[null.String]()
	stringType     = reflect.TypeFor[string]()
	stringPtrType  = reflect.TypeFor[*string]()

	// Null.Int.
	nullIntType = reflect.TypeFor[null.Int]()
	intType     = reflect.TypeFor[int64]()
	intPtrType  = reflect.TypeFor[*int64]()

	// Null.Int16.
	nullInt16Type = reflect.TypeFor[null.Int16]()
	int16Type     = reflect.TypeFor[int16]()
	int16PtrType  = reflect.TypeFor[*int16]()

	// Null.Int32.
	nullInt32Type = reflect.TypeFor[null.Int32]()
	int32Type     = reflect.TypeFor[int32]()
	int32PtrType  = reflect.TypeFor[*int32]()

	// Null.Float.
	nullFloatType = reflect.TypeFor[null.Float]()
	floatType     = reflect.TypeFor[float64]()
	floatPtrType  = reflect.TypeFor[*float64]()

	// Null.Byte.
	nullByteType = reflect.TypeFor[null.Byte]()
	byteType     = reflect.TypeFor[byte]()
	bytePtrType  = reflect.TypeFor[*byte]()

	// Null.DateTime.
	nullDateTimeType = reflect.TypeFor[null.DateTime]()
	dateTimeType     = reflect.TypeFor[datetime.DateTime]()
	dateTimePtrType  = reflect.TypeFor[*datetime.DateTime]()

	// Null.Date.
	nullDateType = reflect.TypeFor[null.Date]()
	dateType     = reflect.TypeFor[datetime.Date]()
	datePtrType  = reflect.TypeFor[*datetime.Date]()

	// Null.Time.
	nullTimeType = reflect.TypeFor[null.Time]()
	timeType     = reflect.TypeFor[datetime.Time]()
	timePtrType  = reflect.TypeFor[*datetime.Time]()

	// Null.Decimal.
	nullDecimalType = reflect.TypeFor[null.Decimal]()
	decimalType     = reflect.TypeFor[decimal.Decimal]()
	decimalPtrType  = reflect.TypeFor[*decimal.Decimal]()

	// Null.Value.
	valueOrZeroMethodIndex int

	// Multipart.FileHeader.
	fileHeaderPtrType      = reflect.TypeFor[*multipart.FileHeader]()
	fileHeaderPtrSliceType = reflect.TypeFor[[]*multipart.FileHeader]()
)

func init() {
	method, _ := reflect.TypeFor[null.Value[any]]().MethodByName("ValueOrZero")
	valueOrZeroMethodIndex = method.Index
}

// convertNullBool handles bidirectional conversion between bool types and null.Bool.
// Supported conversions: bool -> null.Bool, *bool -> null.Bool, null.Bool -> bool, null.Bool -> *bool.
func convertNullBool(from, to reflect.Type, value any) (any, error) {
	if (from == boolType || from == boolPtrType) && to == nullBoolType {
		return lo.TernaryF(
			from == boolType,
			func() null.Bool {
				return null.BoolFrom(value.(bool))
			},
			func() null.Bool {
				return null.BoolFromPtr(value.(*bool))
			},
		), nil
	}

	if from == nullBoolType && (to == boolType || to == boolPtrType) {
		if to == boolType {
			return value.(null.Bool).ValueOrZero(), nil
		}

		return value.(null.Bool).Ptr(), nil
	}

	return value, nil
}

// convertNullString handles bidirectional conversion between string types and null.String.
// Supported conversions: string -> null.String, *string -> null.String, null.String -> string, null.String -> *string.
func convertNullString(from, to reflect.Type, value any) (any, error) {
	if (from == stringType || from == stringPtrType) && to == nullStringType {
		return lo.TernaryF(
			from == stringType,
			func() null.String {
				return null.StringFrom(value.(string))
			},
			func() null.String {
				return null.StringFromPtr(value.(*string))
			},
		), nil
	}

	if from == nullStringType && (to == stringType || to == stringPtrType) {
		if to == stringType {
			return value.(null.String).ValueOrZero(), nil
		}

		return value.(null.String).Ptr(), nil
	}

	return value, nil
}

// convertNullInt handles bidirectional conversion between int64 types and null.Int.
// Supported conversions: int64 -> null.Int, *int64 -> null.Int, null.Int -> int64, null.Int -> *int64.
func convertNullInt(from, to reflect.Type, value any) (any, error) {
	if (from == intType || from == intPtrType) && to == nullIntType {
		return lo.TernaryF(
			from == intType,
			func() null.Int {
				return null.IntFrom(value.(int64))
			},
			func() null.Int {
				return null.IntFromPtr(value.(*int64))
			},
		), nil
	}

	if from == nullIntType && (to == intType || to == intPtrType) {
		if to == intType {
			return value.(null.Int).ValueOrZero(), nil
		}

		return value.(null.Int).Ptr(), nil
	}

	return value, nil
}

// convertNullInt16 handles bidirectional conversion between int16 types and null.Int16.
// Supported conversions: int16 -> null.Int16, *int16 -> null.Int16, null.Int16 -> int16, null.Int16 -> *int16.
func convertNullInt16(from, to reflect.Type, value any) (any, error) {
	if (from == int16Type || from == int16PtrType) && to == nullInt16Type {
		return lo.TernaryF(
			from == int16Type,
			func() null.Int16 {
				return null.Int16From(value.(int16))
			},
			func() null.Int16 {
				return null.Int16FromPtr(value.(*int16))
			},
		), nil
	}

	if from == nullInt16Type && (to == int16Type || to == int16PtrType) {
		if to == int16Type {
			return value.(null.Int16).ValueOrZero(), nil
		}

		return value.(null.Int16).Ptr(), nil
	}

	return value, nil
}

// convertNullInt32 handles bidirectional conversion between int32 types and null.Int32.
// Supported conversions: int32 -> null.Int32, *int32 -> null.Int32, null.Int32 -> int32, null.Int32 -> *int32.
func convertNullInt32(from, to reflect.Type, value any) (any, error) {
	if (from == int32Type || from == int32PtrType) && to == nullInt32Type {
		return lo.TernaryF(
			from == int32Type,
			func() null.Int32 {
				return null.Int32From(value.(int32))
			},
			func() null.Int32 {
				return null.Int32FromPtr(value.(*int32))
			},
		), nil
	}

	if from == nullInt32Type && (to == int32Type || to == int32PtrType) {
		if to == int32Type {
			return value.(null.Int32).ValueOrZero(), nil
		}

		return value.(null.Int32).Ptr(), nil
	}

	return value, nil
}

// convertNullFloat handles bidirectional conversion between float64 types and null.Float.
// Supported conversions: float64 -> null.Float, *float64 -> null.Float, null.Float -> float64, null.Float -> *float64.
func convertNullFloat(from, to reflect.Type, value any) (any, error) {
	if (from == floatType || from == floatPtrType) && to == nullFloatType {
		return lo.TernaryF(
			from == floatType,
			func() null.Float {
				return null.FloatFrom(value.(float64))
			},
			func() null.Float {
				return null.FloatFromPtr(value.(*float64))
			},
		), nil
	}

	if from == nullFloatType && (to == floatType || to == floatPtrType) {
		if to == floatType {
			return value.(null.Float).ValueOrZero(), nil
		}

		return value.(null.Float).Ptr(), nil
	}

	return value, nil
}

// convertNullByte handles bidirectional conversion between byte types and null.Byte.
// Supported conversions: byte -> null.Byte, *byte -> null.Byte, null.Byte -> byte, null.Byte -> *byte.
func convertNullByte(from, to reflect.Type, value any) (any, error) {
	if (from == byteType || from == bytePtrType) && to == nullByteType {
		return lo.TernaryF(
			from == byteType,
			func() null.Byte {
				return null.ByteFrom(value.(byte))
			},
			func() null.Byte {
				return null.ByteFromPtr(value.(*byte))
			},
		), nil
	}

	if from == nullByteType && (to == byteType || to == bytePtrType) {
		if to == byteType {
			return value.(null.Byte).ValueOrZero(), nil
		}

		return value.(null.Byte).Ptr(), nil
	}

	return value, nil
}

// convertNullDateTime handles bidirectional conversion between datetime.DateTime types and null.DateTime.
// Supported conversions: datetime.DateTime -> null.DateTime, *datetime.DateTime -> null.DateTime, null.DateTime -> datetime.DateTime, null.DateTime -> *datetime.DateTime.
func convertNullDateTime(from, to reflect.Type, value any) (any, error) {
	if (from == dateTimeType || from == dateTimePtrType) && to == nullDateTimeType {
		return lo.TernaryF(
			from == dateTimeType,
			func() null.DateTime {
				return null.DateTimeFrom(value.(datetime.DateTime))
			},
			func() null.DateTime {
				return null.DateTimeFromPtr(value.(*datetime.DateTime))
			},
		), nil
	}

	if from == nullDateTimeType && (to == dateTimeType || to == dateTimePtrType) {
		if to == dateTimeType {
			return value.(null.DateTime).ValueOrZero(), nil
		}

		return value.(null.DateTime).Ptr(), nil
	}

	return value, nil
}

// convertNullDate handles bidirectional conversion between datetime.Date types and null.Date.
// Supported conversions: datetime.Date -> null.Date, *datetime.Date -> null.Date, null.Date -> datetime.Date, null.Date -> *datetime.Date.
func convertNullDate(from, to reflect.Type, value any) (any, error) {
	if (from == dateType || from == datePtrType) && to == nullDateType {
		return lo.TernaryF(
			from == dateType,
			func() null.Date {
				return null.DateFrom(value.(datetime.Date))
			},
			func() null.Date {
				return null.DateFromPtr(value.(*datetime.Date))
			},
		), nil
	}

	if from == nullDateType && (to == dateType || to == datePtrType) {
		if to == dateType {
			return value.(null.Date).ValueOrZero(), nil
		}

		return value.(null.Date).Ptr(), nil
	}

	return value, nil
}

// convertNullTime handles bidirectional conversion between datetime.Time types and null.Time.
// Supported conversions: datetime.Time -> null.Time, *datetime.Time -> null.Time, null.Time -> datetime.Time, null.Time -> *datetime.Time.
func convertNullTime(from, to reflect.Type, value any) (any, error) {
	if (from == timeType || from == timePtrType) && to == nullTimeType {
		return lo.TernaryF(
			from == timeType,
			func() null.Time {
				return null.TimeFrom(value.(datetime.Time))
			},
			func() null.Time {
				return null.TimeFromPtr(value.(*datetime.Time))
			},
		), nil
	}

	if from == nullTimeType && (to == timeType || to == timePtrType) {
		if to == timeType {
			return value.(null.Time).ValueOrZero(), nil
		}

		return value.(null.Time).Ptr(), nil
	}

	return value, nil
}

// convertNullDecimal handles bidirectional conversion between decimal.Decimal types and null.Decimal.
// Supported conversions: decimal.Decimal -> null.Decimal, *decimal.Decimal -> null.Decimal, null.Decimal -> decimal.Decimal, null.Decimal -> *decimal.Decimal.
func convertNullDecimal(from, to reflect.Type, value any) (any, error) {
	if (from == decimalType || from == decimalPtrType) && to == nullDecimalType {
		return lo.TernaryF(
			from == decimalType,
			func() null.Decimal {
				return null.DecimalFrom(value.(decimal.Decimal))
			},
			func() null.Decimal {
				return null.DecimalFromPtr(value.(*decimal.Decimal))
			},
		), nil
	}

	if from == nullDecimalType && (to == decimalType || to == decimalPtrType) {
		if to == decimalType {
			return value.(null.Decimal).ValueOrZero(), nil
		}

		return value.(null.Decimal).Ptr(), nil
	}

	return value, nil
}

// convertNullValue handles bidirectional conversion between any type and null.Value[T].
// Supports converting from null.Value[T] to T using ValueOrZero(), and from T to null.Value[T] using ValueFrom().
func convertNullValue(from, to reflect.Type, value any) (any, error) {
	if isNullValue(from) {
		// Use reflection to call ValueOrZero method on the actual value
		method := reflect.ValueOf(value).Method(valueOrZeroMethodIndex)
		if !method.IsValid() {
			return nil, ErrValueOrZeroMethodNotFound
		}

		result := method.Call(nil)

		return result[0].Interface(), nil
	}

	if isNullValue(to) {
		// For target null.Value types, use null.ValueFrom to create the appropriate type
		return null.ValueFrom(value), nil
	}

	return value, nil
}

// isNullValue checks if a reflect.Type is a null.Value.
func isNullValue(t reflect.Type) bool {
	// Check both our package and the underlying null package
	pkgPath := t.PkgPath()
	if pkgPath != "github.com/ilxqx/vef-framework-go/null" && pkgPath != "github.com/guregu/null/v6" {
		return false
	}

	// For generic types like null.Value[T], the type name includes type parameters
	// We need to check if it starts with "Value"
	name := t.Name()

	return len(name) >= 5 && name[:5] == "Value"
}

// convertFileHeader handles conversion from []*multipart.FileHeader to *multipart.FileHeader.
// Extracts the first file from a slice when converting to a single file pointer.
func convertFileHeader(from, to reflect.Type, value any) (any, error) {
	if from == fileHeaderPtrSliceType && to == fileHeaderPtrType {
		files := value.([]*multipart.FileHeader)
		if len(files) == 1 {
			return files[0], nil
		}
	}

	return value, nil
}
