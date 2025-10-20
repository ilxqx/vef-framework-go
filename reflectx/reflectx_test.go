package reflectx

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test structs for method testing.
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

// Generic types for testing IsSimilarType.
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
			name:     "pointer to int",
			input:    reflect.TypeOf((*int)(nil)),
			expected: reflect.TypeOf(int(0)),
		},
		{
			name:     "pointer to string",
			input:    reflect.TypeOf((*string)(nil)),
			expected: reflect.TypeOf(""),
		},
		{
			name:     "non-pointer int",
			input:    reflect.TypeOf(int(0)),
			expected: reflect.TypeOf(int(0)),
		},
		{
			name:     "non-pointer struct",
			input:    reflect.TypeOf(BaseStruct{}),
			expected: reflect.TypeOf(BaseStruct{}),
		},
		{
			name:     "pointer to struct",
			input:    reflect.TypeOf((*BaseStruct)(nil)),
			expected: reflect.TypeOf(BaseStruct{}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Indirect(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsSimilarType(t *testing.T) {
	t.Run("Identical types", func(t *testing.T) {
		t1 := reflect.TypeOf(int(0))
		t2 := reflect.TypeOf(int(0))
		assert.True(t, IsSimilarType(t1, t2))
	})

	t.Run("Different basic types", func(t *testing.T) {
		t1 := reflect.TypeOf(int(0))
		t2 := reflect.TypeOf(string(""))
		assert.False(t, IsSimilarType(t1, t2))
	})

	t.Run("Generic types with same base", func(t *testing.T) {
		t1 := reflect.TypeOf(GenericStruct[int]{})
		t2 := reflect.TypeOf(GenericStruct[string]{})
		assert.True(t, IsSimilarType(t1, t2))
	})

	t.Run("Different package path", func(t *testing.T) {
		t1 := reflect.TypeOf(BaseStruct{})
		t2 := reflect.TypeOf(reflect.Value{})
		assert.False(t, IsSimilarType(t1, t2))
	})

	t.Run("Non-generic types", func(t *testing.T) {
		t1 := reflect.TypeOf(BaseStruct{})
		t2 := reflect.TypeOf(EmbeddedStruct{})
		assert.False(t, IsSimilarType(t1, t2))
	})
}

func TestApplyIfString(t *testing.T) {
	t.Run("String value", func(t *testing.T) {
		result := ApplyIfString("hello", func(s string) int {
			return len(s)
		})
		assert.Equal(t, 5, result)
	})

	t.Run("Reflect.Value string", func(t *testing.T) {
		rv := reflect.ValueOf("world")
		result := ApplyIfString(rv, func(s string) int {
			return len(s)
		})
		assert.Equal(t, 5, result)
	})

	t.Run("Pointer to string", func(t *testing.T) {
		str := "test"
		result := ApplyIfString(&str, func(s string) int {
			return len(s)
		})
		assert.Equal(t, 4, result)
	})

	t.Run("Non-string value with default", func(t *testing.T) {
		result := ApplyIfString(123, func(s string) int {
			return len(s)
		}, 999)
		assert.Equal(t, 999, result)
	})

	t.Run("Non-string value without default", func(t *testing.T) {
		result := ApplyIfString(123, func(s string) int {
			return len(s)
		})
		assert.Equal(t, 0, result) // empty value for int
	})

	t.Run("Nil pointer", func(t *testing.T) {
		var str *string

		result := ApplyIfString(str, func(s string) int {
			return len(s)
		}, 100)
		assert.Equal(t, 100, result)
	})
}

func TestFindMethod(t *testing.T) {
	t.Run("Direct method on value", func(t *testing.T) {
		base := BaseStruct{Value: "test"}
		rv := reflect.ValueOf(base)

		method := FindMethod(rv, "BaseMethod")
		require.True(t, method.IsValid())

		result := method.Call(nil)
		assert.Equal(t, "base method", result[0].String())
	})

	t.Run("Pointer receiver method", func(t *testing.T) {
		base := BaseStruct{Value: "test"}
		rv := reflect.ValueOf(base)

		method := FindMethod(rv, "BasePointerMethod")
		require.True(t, method.IsValid())

		result := method.Call(nil)
		assert.Equal(t, "base pointer method", result[0].String())
	})

	t.Run("Promoted method from embedded struct", func(t *testing.T) {
		embedded := EmbeddedStruct{
			BaseStruct: BaseStruct{Value: "test"},
			Name:       "embedded",
		}
		rv := reflect.ValueOf(embedded)

		// Test promoted method from BaseStruct
		method := FindMethod(rv, "BaseMethod")
		require.True(t, method.IsValid())

		result := method.Call(nil)
		assert.Equal(t, "base method", result[0].String())
	})

	t.Run("Promoted pointer method from embedded struct", func(t *testing.T) {
		embedded := EmbeddedStruct{
			BaseStruct: BaseStruct{Value: "test"},
			Name:       "embedded",
		}
		rv := reflect.ValueOf(embedded)

		// Test promoted pointer method from BaseStruct
		method := FindMethod(rv, "BasePointerMethod")
		require.True(t, method.IsValid())

		result := method.Call(nil)
		assert.Equal(t, "base pointer method", result[0].String())
	})

	t.Run("Method on embedded struct", func(t *testing.T) {
		embedded := EmbeddedStruct{
			BaseStruct: BaseStruct{Value: "test"},
			Name:       "embedded",
		}
		rv := reflect.ValueOf(embedded)

		method := FindMethod(rv, "EmbeddedMethod")
		require.True(t, method.IsValid())

		result := method.Call(nil)
		assert.Equal(t, "embedded method", result[0].String())
	})

	t.Run("Nested embedded struct methods", func(t *testing.T) {
		nested := NestedStruct{
			EmbeddedStruct: EmbeddedStruct{
				BaseStruct: BaseStruct{Value: "test"},
				Name:       "nested",
			},
			Age: 25,
		}
		rv := reflect.ValueOf(nested)

		// Test method from deeply nested BaseStruct
		method := FindMethod(rv, "BaseMethod")
		require.True(t, method.IsValid())

		result := method.Call(nil)
		assert.Equal(t, "base method", result[0].String())
	})

	t.Run("Pointer receiver method on nested struct", func(t *testing.T) {
		nested := NestedStruct{
			EmbeddedStruct: EmbeddedStruct{
				BaseStruct: BaseStruct{Value: "test"},
				Name:       "nested",
			},
			Age: 25,
		}
		rv := reflect.ValueOf(nested)

		method := FindMethod(rv, "NestedPointerMethod")
		require.True(t, method.IsValid())

		result := method.Call(nil)
		assert.Equal(t, "nested pointer method", result[0].String())
	})

	t.Run("Method with pointer value", func(t *testing.T) {
		base := &BaseStruct{Value: "test"}
		rv := reflect.ValueOf(base)

		method := FindMethod(rv, "BaseMethod")
		require.True(t, method.IsValid())

		result := method.Call(nil)
		assert.Equal(t, "base method", result[0].String())
	})

	t.Run("Non-existent method", func(t *testing.T) {
		base := BaseStruct{Value: "test"}
		rv := reflect.ValueOf(base)

		method := FindMethod(rv, "NonExistentMethod")
		assert.False(t, method.IsValid())
	})

	t.Run("Method on non-struct type", func(t *testing.T) {
		rv := reflect.ValueOf(42)

		method := FindMethod(rv, "SomeMethod")
		assert.False(t, method.IsValid())
	})

	t.Run("Method on non-addressable value", func(t *testing.T) {
		getValue := func() BaseStruct {
			return BaseStruct{Value: "test"}
		}

		rv := reflect.ValueOf(getValue())

		// Should still work for pointer receiver methods through the fallback logic
		method := FindMethod(rv, "BasePointerMethod")
		require.True(t, method.IsValid())

		result := method.Call(nil)
		assert.Equal(t, "base pointer method", result[0].String())
	})
}

// Benchmark tests.
func BenchmarkFindMethod(b *testing.B) {
	// Expensive initialization - creating complex nested struct and reflect.Value
	nested := NestedStruct{
		EmbeddedStruct: EmbeddedStruct{
			BaseStruct: BaseStruct{Value: "test"},
			Name:       "nested",
		},
		Age: 25,
	}
	rv := reflect.ValueOf(nested)

	// Reset timer after initialization to exclude setup cost
	b.ResetTimer()

	for b.Loop() {
		FindMethod(rv, "BaseMethod")
	}
}

func BenchmarkIndirect(b *testing.B) {
	// Simple initialization - getting reflect.Type
	ptrType := reflect.TypeOf((*BaseStruct)(nil))

	// Reset timer after initialization
	b.ResetTimer()

	for b.Loop() {
		Indirect(ptrType)
	}
}

// Test types and interfaces for compatibility tests.
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
	t.Run("Exact type match", func(t *testing.T) {
		stringType := reflect.TypeOf("")
		assert.True(t, IsTypeCompatible(stringType, stringType))
	})

	t.Run("Assignable types", func(t *testing.T) {
		intType := reflect.TypeOf(int(0))
		int32Type := reflect.TypeOf(int32(0))

		// int is not assignable to int32
		assert.False(t, IsTypeCompatible(intType, int32Type))

		// Same types should be assignable
		assert.True(t, IsTypeCompatible(intType, intType))
	})

	t.Run("Interface implementation", func(t *testing.T) {
		testStructType := reflect.TypeOf(TestStruct{})
		testInterfaceType := reflect.TypeOf((*TestInterface)(nil)).Elem()

		// TestStruct implements TestInterface
		assert.True(t, IsTypeCompatible(testStructType, testInterfaceType))

		// AnotherStruct does not implement TestInterface
		anotherStructType := reflect.TypeOf(AnotherStruct{})
		assert.False(t, IsTypeCompatible(anotherStructType, testInterfaceType))
	})

	t.Run("Pointer to pointer compatibility", func(t *testing.T) {
		stringPtrType := reflect.TypeOf((*string)(nil))
		stringPtrType2 := reflect.TypeOf((*string)(nil))
		intPtrType := reflect.TypeOf((*int)(nil))

		// Same pointer types
		assert.True(t, IsTypeCompatible(stringPtrType, stringPtrType2))

		// Different pointer types
		assert.False(t, IsTypeCompatible(stringPtrType, intPtrType))
	})

	t.Run("Value to pointer compatibility", func(t *testing.T) {
		stringType := reflect.TypeOf("")
		stringPtrType := reflect.TypeOf((*string)(nil))
		intType := reflect.TypeOf(int(0))

		// string -> *string
		assert.True(t, IsTypeCompatible(stringType, stringPtrType))

		// int -> *string (not compatible)
		assert.False(t, IsTypeCompatible(intType, stringPtrType))
	})

	t.Run("Pointer to value compatibility", func(t *testing.T) {
		stringType := reflect.TypeOf("")
		stringPtrType := reflect.TypeOf((*string)(nil))
		intPtrType := reflect.TypeOf((*int)(nil))

		// *string -> string
		assert.True(t, IsTypeCompatible(stringPtrType, stringType))

		// *int -> string (not compatible)
		assert.False(t, IsTypeCompatible(intPtrType, stringType))
	})

	t.Run("Interface pointer compatibility", func(t *testing.T) {
		testStructPtrType := reflect.TypeOf((*TestStruct)(nil))
		testInterfaceType := reflect.TypeOf((*TestInterface)(nil)).Elem()

		// *TestStruct implements TestInterface
		assert.True(t, IsTypeCompatible(testStructPtrType, testInterfaceType))
	})
}

func TestConvertValue(t *testing.T) {
	t.Run("Same types - no conversion needed", func(t *testing.T) {
		original := reflect.ValueOf("hello")
		result, err := ConvertValue(original, reflect.TypeOf(""))

		require.NoError(t, err)
		assert.Equal(t, "hello", result.String())
	})

	t.Run("Pointer to value conversion", func(t *testing.T) {
		str := "test"
		ptrValue := reflect.ValueOf(&str)
		stringType := reflect.TypeOf("")

		result, err := ConvertValue(ptrValue, stringType)

		require.NoError(t, err)
		assert.Equal(t, "test", result.String())
	})

	t.Run("Nil pointer to value conversion", func(t *testing.T) {
		var str *string

		ptrValue := reflect.ValueOf(str)
		stringType := reflect.TypeOf("")

		result, err := ConvertValue(ptrValue, stringType)

		require.NoError(t, err)
		assert.Equal(t, "", result.String()) // zero value
	})

	t.Run("Value to pointer conversion", func(t *testing.T) {
		original := reflect.ValueOf("hello")
		stringPtrType := reflect.TypeOf((*string)(nil))

		result, err := ConvertValue(original, stringPtrType)

		require.NoError(t, err)
		assert.True(t, result.Kind() == reflect.Pointer)
		assert.Equal(t, "hello", result.Elem().String())
	})

	t.Run("Pointer to pointer conversion", func(t *testing.T) {
		str := "test"
		ptrValue := reflect.ValueOf(&str)
		stringPtrType := reflect.TypeOf((*string)(nil))

		result, err := ConvertValue(ptrValue, stringPtrType)

		require.NoError(t, err)
		assert.True(t, result.Kind() == reflect.Pointer)
		assert.Equal(t, "test", result.Elem().String())
	})

	t.Run("Nil pointer to pointer conversion", func(t *testing.T) {
		var str *string

		ptrValue := reflect.ValueOf(str)
		stringPtrType := reflect.TypeOf((*string)(nil))

		result, err := ConvertValue(ptrValue, stringPtrType)

		require.NoError(t, err)
		assert.True(t, result.IsZero()) // nil pointer
	})

	t.Run("Interface implementation conversion", func(t *testing.T) {
		testStruct := TestStruct{Value: "interface test"}
		original := reflect.ValueOf(testStruct)
		interfaceType := reflect.TypeOf((*TestInterface)(nil)).Elem()

		result, err := ConvertValue(original, interfaceType)

		require.NoError(t, err)
		// Call the interface method to verify conversion worked
		testInterface := result.Interface().(TestInterface)
		assert.Equal(t, "interface test", testInterface.TestMethod())
	})

	t.Run("Incompatible type conversion", func(t *testing.T) {
		intValue := reflect.ValueOf(42)
		stringType := reflect.TypeOf("")

		_, err := ConvertValue(intValue, stringType)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot convert source type")
	})

	t.Run("Struct conversion", func(t *testing.T) {
		testStruct := TestStruct{Value: "test"}
		original := reflect.ValueOf(testStruct)
		testStructType := reflect.TypeOf(TestStruct{})

		result, err := ConvertValue(original, testStructType)

		require.NoError(t, err)

		convertedStruct := result.Interface().(TestStruct)
		assert.Equal(t, "test", convertedStruct.Value)
	})
}
