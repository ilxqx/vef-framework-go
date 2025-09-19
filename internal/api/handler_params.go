package api

import (
	"container/list"
	"fmt"
	"reflect"

	"github.com/gofiber/fiber/v3"
	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/contextx"
	"github.com/ilxqx/vef-framework-go/log"
	"github.com/ilxqx/vef-framework-go/mapx"
	"github.com/ilxqx/vef-framework-go/reflectx"
	"github.com/ilxqx/vef-framework-go/validator"
	"github.com/samber/lo"
)

var (
	paramsType        = reflect.TypeFor[api.In]()
	loggerType        = reflect.TypeFor[log.Logger]()
	withLoaggerMethod = "WithLogger"
)

// paramResolverFn resolves a value from the current request context.
type paramResolverFn func(ctx fiber.Ctx) (reflect.Value, error)

// handlerParamResolverManager aggregates all handler parameter resolvers.
// Resolution uses exact type matching. If needed in the future, assignable/interface
// matching (e.g., Implements/AssignableTo) can be added without changing the public API.
type handlerParamResolverManager struct {
	// resolvers is a map from concrete type to its resolver function.
	resolvers map[reflect.Type]paramResolverFn
}

// newParamResolver builds a composite resolver by merging preset and user-provided resolvers.
// If the same type is registered multiple times, the last one wins (user-provided overrides preset).
// This constructor is intended to be used by the DI layer to assemble resolvers from groups.
func newParamResolver(userResolvers []api.HandlerParamResolver) *handlerParamResolverManager {
	merged := make(map[reflect.Type]paramResolverFn, len(userResolvers)+len(presetParamResolvers))
	// preset first
	for _, resolver := range presetParamResolvers {
		t := resolver.Type()
		merged[t] = resolver.Resolve
	}
	// user-provided override
	for _, resolver := range userResolvers {
		t := resolver.Type()
		merged[t] = resolver.Resolve
	}

	return &handlerParamResolverManager{
		resolvers: merged,
	}
}

// Resolve looks up a resolver function for the given parameter type.
// It does not perform the actual value conversion; callers should invoke the returned
// function with a request context to obtain the value.
// Returns (resolver, nil) on success, or (nil, error) when no resolver is found.
func (m *handlerParamResolverManager) Resolve(target reflect.Value, paramType reflect.Type) (paramResolverFn, error) {
	if resolver, ok := m.resolvers[paramType]; ok {
		return resolver, nil
		// if value := resolver(ctx); value != nil {
		// 	return reflect.ValueOf(value).Convert(targetType)
		// }
	}

	// Try resolve params from api request if paramType is struct and is embeds api.Params
	if hasParamsEmbedded(paramType) {
		return buildParamsResolver(paramType), nil
	}

	// Try resolve params from target struct fields
	if reflect.Indirect(target).Kind() == reflect.Struct {
		structVal := reflect.Indirect(target)
		for i := 0; i < structVal.NumField(); i++ {
			field := structVal.Field(i)

			if field.Type().AssignableTo(paramType) && field.IsValid() {
				return buildFieldResolver(field), nil
			}
		}
	}

	return nil, fmt.Errorf("failed to resolve api handler parameter type: %s", paramType.String())
}

// hasParamsEmbedded checks if the given struct type embeds api.Params
func hasParamsEmbedded(t reflect.Type) bool {
	t = reflectx.Indirect(t)
	if t.Kind() != reflect.Struct {
		return false
	}

	// We are going to do a breadth-first search of all embedded fields.
	types := list.New()
	types.PushBack(t)
	for types.Len() > 0 {
		t := types.Remove(types.Front()).(reflect.Type)

		if t.Kind() != reflect.Struct {
			continue
		}

		if t == paramsType {
			return true
		}

		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			if f.Anonymous {
				types.PushBack(f.Type)
			}
		}
	}

	return false
}

// buildParamsResolver builds a resolver function for api.Params type
func buildParamsResolver(paramType reflect.Type) paramResolverFn {
	t := reflectx.Indirect(paramType)
	return func(ctx fiber.Ctx) (reflect.Value, error) {
		request := contextx.APIRequest(ctx)
		// Create a new instance of the param type
		paramValue := reflect.New(t)
		if err := unmarshalParams(request.Params, paramValue.Interface()); err != nil {
			return lo.Empty[reflect.Value](), err
		}
		if err := validator.Validate(paramValue.Interface()); err != nil {
			return lo.Empty[reflect.Value](), err
		}

		if paramType.Kind() == reflect.Pointer {
			return paramValue, nil
		}

		return paramValue.Elem(), nil
	}
}

// buildFieldResolver builds a resolver function for a struct field
func buildFieldResolver(field reflect.Value) paramResolverFn {
	// Check if the type implements LoggerConfigurable interface
	// For generic interfaces, we need to check the method exists
	requiresConfigureLogger := hasWithLoggerMethod(field.Type())

	return func(ctx fiber.Ctx) (reflect.Value, error) {
		if requiresConfigureLogger {
			logger := contextx.Logger(ctx)
			return callWithLogger(field, logger), nil
		}

		return field, nil
	}
}

// hasWithLoggerMethod checks if the given type has a WithLogger method
func hasWithLoggerMethod(t reflect.Type) bool {
	// First try to find method on the type itself
	method, found := t.MethodByName(withLoaggerMethod)
	if !found {
		// If not found, and it's not already a pointer, try pointer type
		// because WithLogger might be defined on pointer receiver
		if t.Kind() != reflect.Pointer {
			ptrType := reflect.PointerTo(t)
			method, found = ptrType.MethodByName(withLoaggerMethod)
		}
	}

	if !found {
		return false
	}

	// Check method has exactly 2 inputs: receiver + one parameter
	if method.Type.NumIn() != 2 {
		return false
	}

	// Check the parameter (index 1, receiver is index 0) is log.Logger type
	paramType := method.Type.In(1)

	return loggerType.AssignableTo(paramType)
}

// callWithLogger calls the WithLogger method on the value with the given logger
func callWithLogger(field reflect.Value, logger log.Logger) reflect.Value {
	method := reflectx.FindMethod(field, withLoaggerMethod)
	if method.IsValid() {
		// Call WithLogger method with the logger
		results := method.Call([]reflect.Value{reflect.ValueOf(logger)})
		if len(results) > 0 {
			return results[0]
		}
	}

	return field
}

// unmarshalParams unmarshals the request params into the given struct.
func unmarshalParams(params map[string]any, out any) error {
	t := reflect.TypeOf(out)
	if t.Kind() != reflect.Pointer || t.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("the parameter of UnmarshalParams function must be a pointer to a struct, but got %s", t.Kind().String())
	}

	decoder, err := mapx.NewDecoder(out)
	if err != nil {
		return err
	}

	if err = decoder.Decode(params); err != nil {
		return err
	}

	return nil
}
