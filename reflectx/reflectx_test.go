package reflectx

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type BaseStruct struct {
	Value string
}

func (b BaseStruct) BaseMethod() string {
	return "base method"
}

func (b *BaseStruct) BasePointerMethod() string {
	return "base pointer method"
}

type EmbeddedStruct struct {
	BaseStruct

	Name string
}

func (e EmbeddedStruct) EmbeddedMethod() string {
	return "embedded method"
}

type NestedStruct struct {
	EmbeddedStruct

	Age int
}

func (n *NestedStruct) NestedPointerMethod() string {
	return "nested pointer method"
}

type GenericStruct[T any] struct {
	Data T
}

func TestIndirect(t *testing.T) {
	tests := []struct {
		name     string
		input    reflect.Type
		expected reflect.Type
	}{
		{
			name:     "PointerToInt",
			input:    reflect.TypeOf((*int)(nil)),
			expected: reflect.TypeOf(int(0)),
		},
		{
			name:     "PointerToString",
			input:    reflect.TypeOf((*string)(nil)),
			expected: reflect.TypeOf(""),
		},
		{
			name:     "NonPointerInt",
			input:    reflect.TypeOf(int(0)),
			expected: reflect.TypeOf(int(0)),
		},
		{
			name:     "NonPointerStruct",
			input:    reflect.TypeOf(BaseStruct{}),
			expected: reflect.TypeOf(BaseStruct{}),
		},
		{
			name:     "PointerToStruct",
			input:    reflect.TypeOf((*BaseStruct)(nil)),
			expected: reflect.TypeOf(BaseStruct{}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Indirect(tt.input)
			assert.Equal(t, tt.expected, result, "Indirect should return the underlying type")
		})
	}
}

func TestIsSimilarType(t *testing.T) {
	t.Run("IdenticalTypes", func(t *testing.T) {
		t1 := reflect.TypeOf(int(0))
		t2 := reflect.TypeOf(int(0))
		assert.True(t, IsSimilarType(t1, t2), "Identical types should be similar")
	})

	t.Run("DifferentBasicTypes", func(t *testing.T) {
		t1 := reflect.TypeOf(int(0))
		t2 := reflect.TypeOf(string(""))
		assert.False(t, IsSimilarType(t1, t2), "Different basic types should not be similar")
	})

	t.Run("GenericTypesWithSameBase", func(t *testing.T) {
		t1 := reflect.TypeOf(GenericStruct[int]{})
		t2 := reflect.TypeOf(GenericStruct[string]{})
		assert.True(t, IsSimilarType(t1, t2), "Generic types with same base should be similar")
	})

	t.Run("DifferentPackagePath", func(t *testing.T) {
		t1 := reflect.TypeOf(BaseStruct{})
		t2 := reflect.TypeOf(reflect.Value{})
		assert.False(t, IsSimilarType(t1, t2), "Types from different packages should not be similar")
	})

	t.Run("NonGenericTypes", func(t *testing.T) {
		t1 := reflect.TypeOf(BaseStruct{})
		t2 := reflect.TypeOf(EmbeddedStruct{})
		assert.False(t, IsSimilarType(t1, t2), "Different non-generic types should not be similar")
	})
}

func TestApplyIfString(t *testing.T) {
	t.Run("StringValue", func(t *testing.T) {
		result := ApplyIfString("hello", func(s string) int {
			return len(s)
		})
		assert.Equal(t, 5, result, "Should apply function to string value")
	})

	t.Run("ReflectValueString", func(t *testing.T) {
		rv := reflect.ValueOf("world")
		result := ApplyIfString(rv, func(s string) int {
			return len(s)
		})
		assert.Equal(t, 5, result, "Should apply function to reflect.Value string")
	})

	t.Run("PointerToString", func(t *testing.T) {
		str := "test"
		result := ApplyIfString(&str, func(s string) int {
			return len(s)
		})
		assert.Equal(t, 4, result, "Should apply function to pointer to string")
	})

	t.Run("NonStringValueWithDefault", func(t *testing.T) {
		result := ApplyIfString(123, func(s string) int {
			return len(s)
		}, 999)
		assert.Equal(t, 999, result, "Should return default for non-string value")
	})

	t.Run("NonStringValueWithoutDefault", func(t *testing.T) {
		result := ApplyIfString(123, func(s string) int {
			return len(s)
		})
		assert.Equal(t, 0, result, "Should return empty value for non-string without default")
	})

	t.Run("NilPointer", func(t *testing.T) {
		var str *string

		result := ApplyIfString(str, func(s string) int {
			return len(s)
		}, 100)
		assert.Equal(t, 100, result, "Should return default for nil pointer")
	})
}

func TestFindMethod(t *testing.T) {
	t.Run("DirectMethodOnValue", func(t *testing.T) {
		base := BaseStruct{Value: "test"}
		rv := reflect.ValueOf(base)

		method := FindMethod(rv, "BaseMethod")
		require.True(t, method.IsValid(), "Should find direct method on value")

		result := method.Call(nil)
		assert.Equal(t, "base method", result[0].String(), "Method should return expected value")
	})

	t.Run("PointerReceiverMethod", func(t *testing.T) {
		base := BaseStruct{Value: "test"}
		rv := reflect.ValueOf(base)

		method := FindMethod(rv, "BasePointerMethod")
		require.True(t, method.IsValid(), "Should find pointer receiver method")

		result := method.Call(nil)
		assert.Equal(t, "base pointer method", result[0].String(), "Method should return expected value")
	})

	t.Run("PromotedMethodFromEmbeddedStruct", func(t *testing.T) {
		embedded := EmbeddedStruct{
			BaseStruct: BaseStruct{Value: "test"},
			Name:       "embedded",
		}
		rv := reflect.ValueOf(embedded)

		method := FindMethod(rv, "BaseMethod")
		require.True(t, method.IsValid(), "Should find promoted method from embedded struct")

		result := method.Call(nil)
		assert.Equal(t, "base method", result[0].String(), "Promoted method should return expected value")
	})

	t.Run("PromotedPointerMethodFromEmbeddedStruct", func(t *testing.T) {
		embedded := EmbeddedStruct{
			BaseStruct: BaseStruct{Value: "test"},
			Name:       "embedded",
		}
		rv := reflect.ValueOf(embedded)

		method := FindMethod(rv, "BasePointerMethod")
		require.True(t, method.IsValid(), "Should find promoted pointer method from embedded struct")

		result := method.Call(nil)
		assert.Equal(t, "base pointer method", result[0].String(), "Promoted pointer method should return expected value")
	})

	t.Run("MethodOnEmbeddedStruct", func(t *testing.T) {
		embedded := EmbeddedStruct{
			BaseStruct: BaseStruct{Value: "test"},
			Name:       "embedded",
		}
		rv := reflect.ValueOf(embedded)

		method := FindMethod(rv, "EmbeddedMethod")
		require.True(t, method.IsValid(), "Should find method on embedded struct")

		result := method.Call(nil)
		assert.Equal(t, "embedded method", result[0].String(), "Embedded method should return expected value")
	})

	t.Run("NestedEmbeddedStructMethods", func(t *testing.T) {
		nested := NestedStruct{
			EmbeddedStruct: EmbeddedStruct{
				BaseStruct: BaseStruct{Value: "test"},
				Name:       "nested",
			},
			Age: 25,
		}
		rv := reflect.ValueOf(nested)

		method := FindMethod(rv, "BaseMethod")
		require.True(t, method.IsValid(), "Should find method from deeply nested BaseStruct")

		result := method.Call(nil)
		assert.Equal(t, "base method", result[0].String(), "Nested method should return expected value")
	})

	t.Run("PointerReceiverMethodOnNestedStruct", func(t *testing.T) {
		nested := NestedStruct{
			EmbeddedStruct: EmbeddedStruct{
				BaseStruct: BaseStruct{Value: "test"},
				Name:       "nested",
			},
			Age: 25,
		}
		rv := reflect.ValueOf(nested)

		method := FindMethod(rv, "NestedPointerMethod")
		require.True(t, method.IsValid(), "Should find pointer receiver method on nested struct")

		result := method.Call(nil)
		assert.Equal(t, "nested pointer method", result[0].String(), "Nested pointer method should return expected value")
	})

	t.Run("MethodWithPointerValue", func(t *testing.T) {
		base := &BaseStruct{Value: "test"}
		rv := reflect.ValueOf(base)

		method := FindMethod(rv, "BaseMethod")
		require.True(t, method.IsValid(), "Should find method with pointer value")

		result := method.Call(nil)
		assert.Equal(t, "base method", result[0].String(), "Method should return expected value")
	})

	t.Run("NonExistentMethod", func(t *testing.T) {
		base := BaseStruct{Value: "test"}
		rv := reflect.ValueOf(base)

		method := FindMethod(rv, "NonExistentMethod")
		assert.False(t, method.IsValid(), "Should not find non-existent method")
	})

	t.Run("MethodOnNonStructType", func(t *testing.T) {
		rv := reflect.ValueOf(42)

		method := FindMethod(rv, "SomeMethod")
		assert.False(t, method.IsValid(), "Should not find method on non-struct type")
	})

	t.Run("MethodOnNonAddressableValue", func(t *testing.T) {
		getValue := func() BaseStruct {
			return BaseStruct{Value: "test"}
		}

		rv := reflect.ValueOf(getValue())

		method := FindMethod(rv, "BasePointerMethod")
		require.True(t, method.IsValid(), "Should find pointer receiver method on non-addressable value")

		result := method.Call(nil)
		assert.Equal(t, "base pointer method", result[0].String(), "Method should return expected value")
	})
}

func BenchmarkFindMethod(b *testing.B) {
	nested := NestedStruct{
		EmbeddedStruct: EmbeddedStruct{
			BaseStruct: BaseStruct{Value: "test"},
			Name:       "nested",
		},
		Age: 25,
	}
	rv := reflect.ValueOf(nested)

	b.ResetTimer()

	for b.Loop() {
		FindMethod(rv, "BaseMethod")
	}
}

func BenchmarkIndirect(b *testing.B) {
	ptrType := reflect.TypeOf((*BaseStruct)(nil))

	b.ResetTimer()

	for b.Loop() {
		Indirect(ptrType)
	}
}

type TestInterface interface {
	TestMethod() string
}

type TestStruct struct {
	Value string
}

func (t TestStruct) TestMethod() string {
	return t.Value
}

type AnotherStruct struct {
	Data int
}

func TestIsTypeCompatible(t *testing.T) {
	t.Run("ExactTypeMatch", func(t *testing.T) {
		stringType := reflect.TypeOf("")
		assert.True(t, IsTypeCompatible(stringType, stringType), "Exact type match should be compatible")
	})

	t.Run("AssignableTypes", func(t *testing.T) {
		intType := reflect.TypeOf(int(0))
		int32Type := reflect.TypeOf(int32(0))

		assert.False(t, IsTypeCompatible(intType, int32Type), "int is not assignable to int32")
		assert.True(t, IsTypeCompatible(intType, intType), "Same types should be assignable")
	})

	t.Run("InterfaceImplementation", func(t *testing.T) {
		testStructType := reflect.TypeOf(TestStruct{})
		testInterfaceType := reflect.TypeOf((*TestInterface)(nil)).Elem()

		assert.True(t, IsTypeCompatible(testStructType, testInterfaceType), "TestStruct implements TestInterface")

		anotherStructType := reflect.TypeOf(AnotherStruct{})
		assert.False(t, IsTypeCompatible(anotherStructType, testInterfaceType), "AnotherStruct does not implement TestInterface")
	})

	t.Run("PointerToPointerCompatibility", func(t *testing.T) {
		stringPtrType := reflect.TypeOf((*string)(nil))
		stringPtrType2 := reflect.TypeOf((*string)(nil))
		intPtrType := reflect.TypeOf((*int)(nil))

		assert.True(t, IsTypeCompatible(stringPtrType, stringPtrType2), "Same pointer types should be compatible")
		assert.False(t, IsTypeCompatible(stringPtrType, intPtrType), "Different pointer types should not be compatible")
	})

	t.Run("ValueToPointerCompatibility", func(t *testing.T) {
		stringType := reflect.TypeOf("")
		stringPtrType := reflect.TypeOf((*string)(nil))
		intType := reflect.TypeOf(int(0))

		assert.True(t, IsTypeCompatible(stringType, stringPtrType), "string -> *string should be compatible")
		assert.False(t, IsTypeCompatible(intType, stringPtrType), "int -> *string should not be compatible")
	})

	t.Run("PointerToValueCompatibility", func(t *testing.T) {
		stringType := reflect.TypeOf("")
		stringPtrType := reflect.TypeOf((*string)(nil))
		intPtrType := reflect.TypeOf((*int)(nil))

		assert.True(t, IsTypeCompatible(stringPtrType, stringType), "*string -> string should be compatible")
		assert.False(t, IsTypeCompatible(intPtrType, stringType), "*int -> string should not be compatible")
	})

	t.Run("InterfacePointerCompatibility", func(t *testing.T) {
		testStructPtrType := reflect.TypeOf((*TestStruct)(nil))
		testInterfaceType := reflect.TypeOf((*TestInterface)(nil)).Elem()

		assert.True(t, IsTypeCompatible(testStructPtrType, testInterfaceType), "*TestStruct implements TestInterface")
	})
}

func TestConvertValue(t *testing.T) {
	t.Run("SameTypesNoConversionNeeded", func(t *testing.T) {
		original := reflect.ValueOf("hello")
		result, err := ConvertValue(original, reflect.TypeOf(""))

		require.NoError(t, err, "Should not error for same types")
		assert.Equal(t, "hello", result.String(), "Should return original value")
	})

	t.Run("PointerToValueConversion", func(t *testing.T) {
		str := "test"
		ptrValue := reflect.ValueOf(&str)
		stringType := reflect.TypeOf("")

		result, err := ConvertValue(ptrValue, stringType)

		require.NoError(t, err, "Should convert pointer to value")
		assert.Equal(t, "test", result.String(), "Should dereference pointer value")
	})

	t.Run("NilPointerToValueConversion", func(t *testing.T) {
		var str *string

		ptrValue := reflect.ValueOf(str)
		stringType := reflect.TypeOf("")

		result, err := ConvertValue(ptrValue, stringType)

		require.NoError(t, err, "Should convert nil pointer to zero value")
		assert.Equal(t, "", result.String(), "Should return zero value for nil pointer")
	})

	t.Run("ValueToPointerConversion", func(t *testing.T) {
		original := reflect.ValueOf("hello")
		stringPtrType := reflect.TypeOf((*string)(nil))

		result, err := ConvertValue(original, stringPtrType)

		require.NoError(t, err, "Should convert value to pointer")
		assert.True(t, result.Kind() == reflect.Pointer, "Result should be pointer")
		assert.Equal(t, "hello", result.Elem().String(), "Pointer should point to original value")
	})

	t.Run("PointerToPointerConversion", func(t *testing.T) {
		str := "test"
		ptrValue := reflect.ValueOf(&str)
		stringPtrType := reflect.TypeOf((*string)(nil))

		result, err := ConvertValue(ptrValue, stringPtrType)

		require.NoError(t, err, "Should convert pointer to pointer")
		assert.True(t, result.Kind() == reflect.Pointer, "Result should be pointer")
		assert.Equal(t, "test", result.Elem().String(), "Pointer should point to correct value")
	})

	t.Run("NilPointerToPointerConversion", func(t *testing.T) {
		var str *string

		ptrValue := reflect.ValueOf(str)
		stringPtrType := reflect.TypeOf((*string)(nil))

		result, err := ConvertValue(ptrValue, stringPtrType)

		require.NoError(t, err, "Should convert nil pointer to nil pointer")
		assert.True(t, result.IsZero(), "Result should be nil pointer")
	})

	t.Run("InterfaceImplementationConversion", func(t *testing.T) {
		testStruct := TestStruct{Value: "interface test"}
		original := reflect.ValueOf(testStruct)
		interfaceType := reflect.TypeOf((*TestInterface)(nil)).Elem()

		result, err := ConvertValue(original, interfaceType)

		require.NoError(t, err, "Should convert to interface type")

		testInterface := result.Interface().(TestInterface)
		assert.Equal(t, "interface test", testInterface.TestMethod(), "Interface method should work correctly")
	})

	t.Run("IncompatibleTypeConversion", func(t *testing.T) {
		intValue := reflect.ValueOf(42)
		stringType := reflect.TypeOf("")

		_, err := ConvertValue(intValue, stringType)

		assert.Error(t, err, "Should error for incompatible types")
		assert.Contains(t, err.Error(), "cannot convert source type", "Error message should indicate incompatibility")
	})

	t.Run("StructConversion", func(t *testing.T) {
		testStruct := TestStruct{Value: "test"}
		original := reflect.ValueOf(testStruct)
		testStructType := reflect.TypeOf(TestStruct{})

		result, err := ConvertValue(original, testStructType)

		require.NoError(t, err, "Should convert struct to same struct type")

		convertedStruct := result.Interface().(TestStruct)
		assert.Equal(t, "test", convertedStruct.Value, "Struct value should be preserved")
	})
}
