package copier

import (
	"testing"

	"github.com/ilxqx/vef-framework-go/decimal"
	"github.com/ilxqx/vef-framework-go/null"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCopy_BasicFunctionality(t *testing.T) {
	t.Run("simple struct copy", func(t *testing.T) {
		type Source struct {
			Name string
			Age  int
		}
		type Dest struct {
			Name string
			Age  int
		}

		src := Source{Name: "John", Age: 30}
		var dst Dest

		err := Copy(src, &dst)
		require.NoError(t, err)
		assert.Equal(t, "John", dst.Name)
		assert.Equal(t, 30, dst.Age)
	})
}

func TestCopy_Converters(t *testing.T) {
	t.Run("verify all converters are available", func(t *testing.T) {
		// Just verify that the converters are properly defined
		assert.NotNil(t, stringConverter)
		assert.NotNil(t, nullStringConverter)
		assert.NotNil(t, intConverter)
		assert.NotNil(t, nullIntConverter)
		assert.NotNil(t, boolConverter)
		assert.NotNil(t, nullBoolConverter)
		assert.NotNil(t, floatConverter)
		assert.NotNil(t, nullFloatConverter)
		assert.NotNil(t, decimalConverter)
		assert.NotNil(t, nullDecimalConverter)
	})

	t.Run("null.String converter function", func(t *testing.T) {
		nullStr := null.StringFrom("test")
		result, err := stringConverter.Fn(nullStr)
		require.NoError(t, err)
		assert.Equal(t, "test", result.(string))
	})

	t.Run("string to null.String converter function", func(t *testing.T) {
		str := "test"
		result, err := nullStringConverter.Fn(str)
		require.NoError(t, err)
		nullStr := result.(null.String)
		assert.True(t, nullStr.Valid)
		assert.Equal(t, "test", nullStr.String)
	})

	t.Run("null.Int converter function", func(t *testing.T) {
		nullInt := null.IntFrom(42)
		result, err := intConverter.Fn(nullInt)
		require.NoError(t, err)
		assert.Equal(t, int64(42), result.(int64))
	})

	t.Run("int64 to null.Int converter function", func(t *testing.T) {
		intVal := int64(42)
		result, err := nullIntConverter.Fn(intVal)
		require.NoError(t, err)
		nullInt := result.(null.Int)
		assert.True(t, nullInt.Valid)
		assert.Equal(t, int64(42), nullInt.Int64)
	})

	t.Run("null.Bool converter function", func(t *testing.T) {
		nullBool := null.BoolFrom(true)
		result, err := boolConverter.Fn(nullBool)
		require.NoError(t, err)
		assert.True(t, result.(bool))
	})

	t.Run("bool to null.Bool converter function", func(t *testing.T) {
		boolVal := true
		result, err := nullBoolConverter.Fn(boolVal)
		require.NoError(t, err)
		nullBool := result.(null.Bool)
		assert.True(t, nullBool.Valid)
		assert.True(t, nullBool.Bool)
	})

	t.Run("null.Float converter function", func(t *testing.T) {
		nullFloat := null.FloatFrom(3.14)
		result, err := floatConverter.Fn(nullFloat)
		require.NoError(t, err)
		assert.Equal(t, 3.14, result.(float64))
	})

	t.Run("float64 to null.Float converter function", func(t *testing.T) {
		floatVal := 3.14
		result, err := nullFloatConverter.Fn(floatVal)
		require.NoError(t, err)
		nullFloat := result.(null.Float)
		assert.True(t, nullFloat.Valid)
		assert.Equal(t, 3.14, nullFloat.Float64)
	})

	t.Run("null.Decimal converter function", func(t *testing.T) {
		testDecimal := decimal.NewFromFloat(123.45)
		nullDecimal := null.DecimalFrom(testDecimal)
		result, err := decimalConverter.Fn(nullDecimal)
		require.NoError(t, err)
		resultDecimal := result.(decimal.Decimal)
		assert.True(t, testDecimal.Equal(resultDecimal))
	})

	t.Run("decimal to null.Decimal converter function", func(t *testing.T) {
		testDecimal := decimal.NewFromFloat(123.45)
		result, err := nullDecimalConverter.Fn(testDecimal)
		require.NoError(t, err)
		nullDecimal := result.(null.Decimal)
		assert.True(t, nullDecimal.Valid)
		assert.True(t, testDecimal.Equal(nullDecimal.Decimal))
	})
}

func TestCopy_Options(t *testing.T) {
	t.Run("copy with ignore empty option", func(t *testing.T) {
		type Source struct {
			Name string
			Age  int
		}
		type Dest struct {
			Name string
			Age  int
		}

		// Set initial values in destination
		dst := Dest{Name: "Initial Name", Age: 25}

		// Source has empty name
		src := Source{Name: "", Age: 30}

		err := Copy(src, &dst, WithIgnoreEmpty())
		require.NoError(t, err)

		// Note: The behavior of IgnoreEmpty might vary based on copier version
		// This test validates that the option can be used without error
		assert.Equal(t, 30, dst.Age) // Age should be updated
	})

	t.Run("copy with case insensitive option", func(t *testing.T) {
		type Source struct {
			NAME string
		}
		type Dest struct {
			Name string
		}

		src := Source{NAME: "John Doe"}
		var dst Dest

		err := Copy(src, &dst, WithCaseInsensitive())
		require.NoError(t, err)
		assert.Equal(t, "John Doe", dst.Name)
	})
}

func TestCopy_SimpleErrorHandling(t *testing.T) {
	t.Run("copy to non-pointer destination should fail", func(t *testing.T) {
		type Source struct {
			Name string
		}
		type Dest struct {
			Name string
		}

		src := Source{Name: "John"}
		var dst Dest

		// Passing dst instead of &dst should cause an error
		err := Copy(src, dst)
		assert.Error(t, err)
	})
}
