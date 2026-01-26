package resolver

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ilxqx/vef-framework-go/api"
)

func TestSelectClosestMatch(t *testing.T) {
	t.Log("Testing selectClosestMatch function")

	t.Run("EmptyCandidates", func(t *testing.T) {
		result := selectClosestMatch("target", []string{})
		assert.Empty(t, result, "Empty candidates should return empty string")
	})

	t.Run("SingleCandidate", func(t *testing.T) {
		result := selectClosestMatch("GetUser", []string{"GetUsers"})
		assert.Equal(t, "GetUsers", result, "Single candidate should be returned")
	})

	t.Run("ExactMatch", func(t *testing.T) {
		result := selectClosestMatch("GetUser", []string{"GetUser", "GetUsers", "GetUserById"})
		assert.Equal(t, "GetUser", result, "Exact match should be returned")
	})

	t.Run("ClosestMatch", func(t *testing.T) {
		result := selectClosestMatch("GetCpu", []string{"GetCPU", "GetMemory", "GetDisk"})
		assert.Equal(t, "GetCPU", result, "Closest match should be returned")
	})

	t.Run("AmbiguousMatch", func(t *testing.T) {
		// Two candidates with same edit distance
		result := selectClosestMatch("ab", []string{"ac", "ad"})
		assert.Empty(t, result, "Ambiguous match should return empty string")
	})

	t.Run("CaseInsensitiveClosest", func(t *testing.T) {
		result := selectClosestMatch("getuser", []string{"GetUser", "GetUsers"})
		assert.Equal(t, "GetUser", result, "Case-insensitive closest match should be returned")
	})
}

func TestValidateHandlerSignature(t *testing.T) {
	t.Log("Testing validateHandlerSignature function")

	t.Run("NoReturnValue", func(t *testing.T) {
		fn := func() {}
		err := validateHandlerSignature(reflect.TypeOf(fn))
		assert.NoError(t, err, "Handler with no return value should be valid")
	})

	t.Run("SingleErrorReturn", func(t *testing.T) {
		fn := func() error { return nil }
		err := validateHandlerSignature(reflect.TypeOf(fn))
		assert.NoError(t, err, "Handler returning error should be valid")
	})

	t.Run("InvalidSingleReturn", func(t *testing.T) {
		fn := func() string { return "" }
		err := validateHandlerSignature(reflect.TypeOf(fn))
		assert.Error(t, err, "Handler returning non-error should be invalid")
		assert.Contains(t, err.Error(), "invalid return type")
	})

	t.Run("TooManyReturns", func(t *testing.T) {
		fn := func() (string, error) { return "", nil }
		err := validateHandlerSignature(reflect.TypeOf(fn))
		assert.Error(t, err, "Handler with too many returns should be invalid")
		assert.Contains(t, err.Error(), "too many return values")
	})

	t.Run("WithParameters", func(t *testing.T) {
		fn := func(a int, b string) error { return nil }
		err := validateHandlerSignature(reflect.TypeOf(fn))
		assert.NoError(t, err, "Handler with parameters should be valid")
	})
}

func TestIsHandlerFactory(t *testing.T) {
	t.Log("Testing isHandlerFactory function")

	t.Run("NotAFactory", func(t *testing.T) {
		fn := func() error { return nil }
		result := isHandlerFactory(reflect.TypeOf(fn))
		assert.False(t, result, "Regular handler should not be a factory")
	})

	t.Run("FactoryReturningHandler", func(t *testing.T) {
		fn := func() func() error { return nil }
		result := isHandlerFactory(reflect.TypeOf(fn))
		assert.True(t, result, "Function returning handler should be a factory")
	})

	t.Run("FactoryReturningHandlerWithError", func(t *testing.T) {
		fn := func() (func() error, error) { return nil, nil }
		result := isHandlerFactory(reflect.TypeOf(fn))
		assert.True(t, result, "Function returning handler and error should be a factory")
	})

	t.Run("FactoryWithParams", func(t *testing.T) {
		fn := func(a int) func() error { return nil }
		result := isHandlerFactory(reflect.TypeOf(fn))
		assert.True(t, result, "Factory with parameters should be valid")
	})

	t.Run("FactoryReturningVoidHandler", func(t *testing.T) {
		fn := func() func() { return nil }
		result := isHandlerFactory(reflect.TypeOf(fn))
		assert.True(t, result, "Factory returning void handler should be valid")
	})

	t.Run("InvalidFactoryTooManyReturns", func(t *testing.T) {
		fn := func() (func() error, error, int) { return nil, nil, 0 }
		result := isHandlerFactory(reflect.TypeOf(fn))
		assert.False(t, result, "Factory with too many returns should be invalid")
	})

	t.Run("InvalidFactorySecondReturnNotError", func(t *testing.T) {
		fn := func() (func() error, string) { return nil, "" }
		result := isHandlerFactory(reflect.TypeOf(fn))
		assert.False(t, result, "Factory with non-error second return should be invalid")
	})

	t.Run("NotReturningFunc", func(t *testing.T) {
		fn := func() string { return "" }
		result := isHandlerFactory(reflect.TypeOf(fn))
		assert.False(t, result, "Function not returning func should not be a factory")
	})
}

func TestValidateHandler(t *testing.T) {
	t.Log("Testing validateHandler function")

	t.Run("ValidHandler", func(t *testing.T) {
		fn := func() error { return nil }
		err := validateHandler(reflect.ValueOf(fn))
		assert.NoError(t, err, "Valid handler should pass validation")
	})

	t.Run("ValidFactory", func(t *testing.T) {
		fn := func() func() error { return nil }
		err := validateHandler(reflect.ValueOf(fn))
		assert.NoError(t, err, "Valid factory should pass validation")
	})

	t.Run("NotAFunction", func(t *testing.T) {
		err := validateHandler(reflect.ValueOf("not a function"))
		assert.Error(t, err, "Non-function should fail validation")
		assert.Contains(t, err.Error(), "must be a function")
	})

	t.Run("NilFunction", func(t *testing.T) {
		var fn func() error

		err := validateHandler(reflect.ValueOf(fn))
		assert.Error(t, err, "Nil function should fail validation")
		assert.Contains(t, err.Error(), "cannot be nil")
	})
}

// Mock resource for testing.
type mockResource struct {
	api.Resource
}

func (m *mockResource) GetUser() error {
	return nil
}

func (m *mockResource) GetCPU() error {
	return nil
}

func (m *mockResource) CreateUserFactory() func() error {
	return func() error { return nil }
}

func TestFindHandlerMethod(t *testing.T) {
	t.Log("Testing findHandlerMethod function")

	resource := &mockResource{
		Resource: api.NewRPCResource("test"),
	}

	t.Run("ExactMatch", func(t *testing.T) {
		method, err := findHandlerMethod(reflect.ValueOf(resource), "GetUser")
		assert.NoError(t, err, "Exact match should succeed")
		assert.True(t, method.IsValid(), "Method should be valid")
	})

	t.Run("CaseInsensitiveMatch", func(t *testing.T) {
		method, err := findHandlerMethod(reflect.ValueOf(resource), "GetCpu")
		assert.NoError(t, err, "Case-insensitive match should succeed")
		assert.True(t, method.IsValid(), "Method should be valid")
	})

	t.Run("NotFound", func(t *testing.T) {
		_, err := findHandlerMethod(reflect.ValueOf(resource), "NonExistent")
		assert.Error(t, err, "Non-existent method should fail")
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("FactoryMethod", func(t *testing.T) {
		method, err := findHandlerMethod(reflect.ValueOf(resource), "CreateUserFactory")
		assert.NoError(t, err, "Factory method should be found")
		assert.True(t, method.IsValid(), "Method should be valid")
	})
}
