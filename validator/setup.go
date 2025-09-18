package validator

import (
	"reflect"

	"github.com/ilxqx/vef-framework-go/null"
	"github.com/shopspring/decimal"
)

// setup initializes preset validation rules and type functions.
func setup() {
	if err := RegisterValidationRules(presetValidationRules...); err != nil {
		panic(err)
	}

	RegisterTypeFunc(
		func(field reflect.Value) any {
			switch v := field.Interface().(type) {
			case null.String:
				if v.Valid {
					return v.String
				}
			case null.Int:
				if v.Valid {
					return v.Int64
				}
			case null.Int16:
				if v.Valid {
					return v.Int16
				}
			case null.Int32:
				if v.Valid {
					return v.Int32
				}
			case null.Float:
				if v.Valid {
					return v.Float64
				}
			case null.Bool:
				if v.Valid {
					return v.Bool
				}
			case null.Byte:
				if v.Valid {
					return v.Byte
				}
			case null.DateTime:
				if v.Valid {
					return v.V
				}
			case null.Date:
				if v.Valid {
					return v.V
				}
			case null.Time:
				if v.Valid {
					return v.V
				}
			case null.Decimal:
				if v.Valid {
					return v.Decimal
				}
			default:
				logger.Warnf("Unsupported null type: %T", field.Interface())
			}

			return nil
		},
		null.String{},
		null.Int{},
		null.Int16{},
		null.Int32{},
		null.Float{},
		null.Bool{},
		null.Byte{},
		null.DateTime{},
		null.Date{},
		null.Time{},
		null.Decimal{},
	)

	// Register commonly used JSON types
	RegisterNullJSONTypeFunc[map[string]any]()
	RegisterNullJSONTypeFunc[map[string]string]()
	RegisterNullJSONTypeFunc[[]string]()
	RegisterNullJSONTypeFunc[[]int]()
	RegisterNullJSONTypeFunc[[]float64]()
	RegisterNullJSONTypeFunc[[]decimal.Decimal]()
	RegisterNullJSONTypeFunc[any]()
}
