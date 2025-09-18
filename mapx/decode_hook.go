package mapx

import (
	"errors"
	"reflect"

	"github.com/ilxqx/vef-framework-go/null"
)

var (
	nullBoolType           = reflect.TypeFor[null.Bool]()
	boolType               = reflect.TypeFor[bool]()
	valueOrZeroMethodIndex int
)

func init() {
	method, _ := reflect.TypeFor[null.Value[any]]().MethodByName("ValueOrZero")
	valueOrZeroMethodIndex = method.Index
}

// decodeNullBool decodes a null.Bool from a reflect.Type to a reflect.Type.
func decodeNullBool(from reflect.Type, to reflect.Type, value any) (any, error) {
	if from == boolType && to == nullBoolType {
		return null.BoolFrom(value.(bool)), nil
	}
	if from == nullBoolType && to == boolType {
		return value.(null.Bool).ValueOrZero(), nil
	}

	return value, nil
}

// decodeNullValue decodes a null.Value from a reflect.Type to a reflect.Type.
func decodeNullValue(from reflect.Type, to reflect.Type, value any) (any, error) {
	if isNullValue(from) {
		// Use reflection to call ValueOrZero method on the actual value
		method := reflect.ValueOf(value).Method(valueOrZeroMethodIndex)
		if !method.IsValid() {
			return nil, errors.New("ValueOrZero method not found on null.Value type")
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
	// Check both our package and the underlying guregu package
	pkgPath := t.PkgPath()
	if pkgPath != "github.com/ilxqx/vef-framework-go/null" && pkgPath != "github.com/guregu/null/v6" {
		return false
	}

	// For generic types like null.Value[T], the type name includes type parameters
	// We need to check if it starts with "Value"
	name := t.Name()
	return len(name) >= 5 && name[:5] == "Value"
}
