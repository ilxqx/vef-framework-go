package reflectx

import (
	"reflect"
	"strings"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/samber/lo"
)

// Indirect returns the underlying type of pointer type.
func Indirect(t reflect.Type) reflect.Type {
	if t.Kind() == reflect.Pointer {
		return t.Elem()
	}

	return t
}

// IsSimilarType checks if two types are similar.
// Two types are considered similar if:
//  1. They are identical (always returns true for identical types)
//  2. They are generic types with the same base type but different type parameters
//     (e.g., List[int] and List[string] are similar because they share the same base type List)
//
// This is useful for comparing generic types where the type parameters may differ,
// but the underlying structure is the same.
func IsSimilarType(t1, t2 reflect.Type) bool {
	if t1 == t2 {
		return true
	}

	if t1.PkgPath() != t2.PkgPath() {
		return false
	}

	name1, name2 := t1.Name(), t2.Name()
	index1, index2 := strings.IndexByte(name1, constants.ByteLeftBracket), strings.IndexByte(name2, constants.ByteLeftBracket)
	if index1 > -1 && index2 > -1 {
		if index1 != index2 {
			return false
		}

		return name1[:index1] == name2[:index2]
	}

	return false
}

// ApplyIfString applies a function to a string value.
// If the value is not a string, it will be converted to a string.
// If the value is a pointer, it will be delivered and the value will be converted to a string.
// If the value is not a string or pointer, it will return the default value or empty value of the type.
func ApplyIfString[T any](value any, fn func(string) T, defaultValue ...T) T {
	var rv reflect.Value
	if v, ok := value.(reflect.Value); ok {
		rv = reflect.Indirect(v)
	} else {
		rv = reflect.Indirect(reflect.ValueOf(value))
	}

	kind := rv.Kind()
	if kind == reflect.String {
		return fn(rv.String())
	}

	if len(defaultValue) > 0 {
		return defaultValue[0]
	}

	return lo.Empty[T]()
}

// FindMethod finds a method on a target value.
// It supports method on the target value, pointer receiver, and promoted methods from embedded fields.
// Go's MethodByName automatically handles promoted methods from embedded structs.
func FindMethod(target reflect.Value, name string) reflect.Value {
	// Direct method lookup (includes promoted methods from embedded fields)
	method := target.MethodByName(name)
	if method.IsValid() {
		return method
	}

	// Pointer receiver lookup
	if target.Kind() != reflect.Pointer {
		var pointerValue reflect.Value
		if target.CanAddr() {
			pointerValue = target.Addr()
		} else {
			pointerValue = reflect.New(target.Type())
			pointerValue.Elem().Set(target)
		}

		method = pointerValue.MethodByName(name)
		if method.IsValid() {
			return method
		}
	}

	return reflect.Value{}
}
