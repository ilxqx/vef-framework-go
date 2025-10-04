package excel

import (
	"fmt"
	"reflect"
	"time"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/datetime"
	"github.com/ilxqx/vef-framework-go/decimal"
	"github.com/ilxqx/vef-framework-go/null"
	"github.com/spf13/cast"
)

var (
	// Cached reflect types for performance
	typeNullString   = reflect.TypeFor[null.String]()
	typeNullInt      = reflect.TypeFor[null.Int]()
	typeNullInt16    = reflect.TypeFor[null.Int16]()
	typeNullInt32    = reflect.TypeFor[null.Int32]()
	typeNullFloat    = reflect.TypeFor[null.Float]()
	typeNullBool     = reflect.TypeFor[null.Bool]()
	typeNullByte     = reflect.TypeFor[null.Byte]()
	typeNullDateTime = reflect.TypeFor[null.DateTime]()
	typeNullDate     = reflect.TypeFor[null.Date]()
	typeNullTime     = reflect.TypeFor[null.Time]()
	typeNullDecimal  = reflect.TypeFor[null.Decimal]()
	typeTime         = reflect.TypeFor[time.Time]()
	typeDecimal      = reflect.TypeFor[decimal.Decimal]()
)

// defaultParser is the built-in parser that handles common Go types.
type defaultParser struct {
	format string
}

// Parse implements the ValueParser interface for common Go types.
func (p *defaultParser) Parse(cellValue string, targetType reflect.Type) (any, error) {
	// Handle empty cell
	if cellValue == constants.Empty {
		return reflect.Zero(targetType).Interface(), nil
	}

	// Handle pointer types
	if targetType.Kind() == reflect.Pointer {
		elemType := targetType.Elem()
		value, err := p.parseValue(cellValue, elemType)
		if err != nil {
			return nil, err
		}
		// Create pointer to the value
		ptr := reflect.New(elemType)
		ptr.Elem().Set(reflect.ValueOf(value))
		return ptr.Interface(), nil
	}

	return p.parseValue(cellValue, targetType)
}

// parseValue parses the cell value to the target type.
func (p *defaultParser) parseValue(cellValue string, targetType reflect.Type) (any, error) {
	// Handle null types
	switch targetType {
	case typeNullString:
		return null.StringFrom(cellValue), nil
	case typeNullInt:
		v, err := cast.ToInt64E(cellValue)
		if err != nil {
			return null.Int{}, fmt.Errorf("parse int: %w", err)
		}
		return null.IntFrom(v), nil
	case typeNullInt16:
		v, err := cast.ToInt16E(cellValue)
		if err != nil {
			return null.Int16{}, fmt.Errorf("parse int16: %w", err)
		}
		return null.Int16From(v), nil
	case typeNullInt32:
		v, err := cast.ToInt32E(cellValue)
		if err != nil {
			return null.Int32{}, fmt.Errorf("parse int32: %w", err)
		}
		return null.Int32From(v), nil
	case typeNullFloat:
		v, err := cast.ToFloat64E(cellValue)
		if err != nil {
			return null.Float{}, fmt.Errorf("parse float: %w", err)
		}
		return null.FloatFrom(v), nil
	case typeNullBool:
		v, err := cast.ToBoolE(cellValue)
		if err != nil {
			return null.Bool{}, fmt.Errorf("parse bool: %w", err)
		}
		return null.BoolFrom(v), nil
	case typeNullByte:
		v, err := cast.ToUint8E(cellValue)
		if err != nil {
			return null.Byte{}, fmt.Errorf("parse byte: %w", err)
		}
		return null.ByteFrom(v), nil
	case typeNullDateTime:
		format := p.format
		if format == constants.Empty {
			format = time.DateTime
		}
		v, err := time.ParseInLocation(format, cellValue, time.Local)
		if err != nil {
			return null.DateTime{}, fmt.Errorf("parse datetime: %w", err)
		}
		return null.DateTimeFrom(datetime.DateTime(v)), nil
	case typeNullDate:
		format := p.format
		if format == constants.Empty {
			format = time.DateOnly
		}
		v, err := time.ParseInLocation(format, cellValue, time.Local)
		if err != nil {
			return null.Date{}, fmt.Errorf("parse date: %w", err)
		}
		return null.DateFrom(datetime.Date(v)), nil
	case typeNullTime:
		format := p.format
		if format == constants.Empty {
			format = time.TimeOnly
		}
		v, err := time.ParseInLocation(format, cellValue, time.Local)
		if err != nil {
			return null.Time{}, fmt.Errorf("parse time: %w", err)
		}
		return null.TimeFrom(datetime.Time(v)), nil
	case typeNullDecimal:
		v, err := decimal.NewFromString(cellValue)
		if err != nil {
			return null.Decimal{}, fmt.Errorf("parse decimal: %w", err)
		}
		return null.DecimalFrom(v), nil
	}

	// Handle basic types
	switch targetType.Kind() {
	case reflect.String:
		return cellValue, nil
	case reflect.Int:
		return cast.ToIntE(cellValue)
	case reflect.Int8:
		return cast.ToInt8E(cellValue)
	case reflect.Int16:
		return cast.ToInt16E(cellValue)
	case reflect.Int32:
		return cast.ToInt32E(cellValue)
	case reflect.Int64:
		return cast.ToInt64E(cellValue)
	case reflect.Uint:
		return cast.ToUintE(cellValue)
	case reflect.Uint8:
		return cast.ToUint8E(cellValue)
	case reflect.Uint16:
		return cast.ToUint16E(cellValue)
	case reflect.Uint32:
		return cast.ToUint32E(cellValue)
	case reflect.Uint64:
		return cast.ToUint64E(cellValue)
	case reflect.Float32:
		return cast.ToFloat32E(cellValue)
	case reflect.Float64:
		return cast.ToFloat64E(cellValue)
	case reflect.Bool:
		return cast.ToBoolE(cellValue)
	case reflect.Struct:
		// Handle time.Time
		if targetType == typeTime {
			format := p.format
			if format == constants.Empty {
				format = time.DateTime
			}
			return time.ParseInLocation(format, cellValue, time.Local)
		}
		// Handle decimal.Decimal
		if targetType == typeDecimal {
			return decimal.NewFromString(cellValue)
		}
	}

	return nil, fmt.Errorf("unsupported type: %v", targetType)
}

// newDefaultParser creates a default parser with optional format template.
func newDefaultParser(format string) ValueParser {
	return &defaultParser{format: format}
}
