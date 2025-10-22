package api

import (
	"container/list"
	"fmt"
	"reflect"

	"github.com/gofiber/fiber/v3"
	"github.com/samber/lo"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/contextx"
	"github.com/ilxqx/vef-framework-go/log"
	"github.com/ilxqx/vef-framework-go/reflectx"
	"github.com/ilxqx/vef-framework-go/validator"
)

var (
	apiPType         = reflect.TypeFor[api.P]()
	apiMType         = reflect.TypeFor[api.M]()
	loggerType       = reflect.TypeFor[log.Logger]()
	withLoggerMethod = "WithLogger"
)

// ParamResolverFunc resolves a value from the current request context.
type ParamResolverFunc func(ctx fiber.Ctx) (reflect.Value, error)

// HandlerParamResolverManager aggregates all handler parameter resolvers.
// Resolution uses exact type matching. If needed in the future, assignable/interface
// matching (e.g., Implements/AssignableTo) can be added without changing the public Api.
type HandlerParamResolverManager struct {
	// resolvers is a map from concrete type to its resolver function.
	resolvers map[reflect.Type]ParamResolverFunc
}

// NewHandlerParamResolverManager builds a composite resolver by merging preset and user-provided resolvers.
// If the same type is registered multiple times, the last one wins (user-provided overrides preset).
// This constructor is intended to be used by the DI layer to assemble resolvers from groups.
func NewHandlerParamResolverManager(userResolvers []api.HandlerParamResolver) *HandlerParamResolverManager {
	merged := make(map[reflect.Type]ParamResolverFunc, len(userResolvers)+len(presetParamResolvers))
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

	return &HandlerParamResolverManager{
		resolvers: merged,
	}
}

// Resolve looks up a resolver function for the given parameter type.
// It does not perform the actual value conversion; callers should invoke the returned
// function with a request context to obtain the value.
// Returns (resolver, nil) on success, or (nil, error) when no resolver is found.
func (m *HandlerParamResolverManager) Resolve(target reflect.Value, paramType reflect.Type) (ParamResolverFunc, error) {
	if resolver, ok := m.resolvers[paramType]; ok {
		return resolver, nil
		// if value := resolver(ctx); value != nil {
		// 	return reflect.ValueOf(value).Convert(targetType)
		// }
	}

	// Try resolve params from api request if paramType is struct and embeds api.P
	if embedsApiP(paramType) {
		return buildParamsResolver(paramType), nil
	}

	// Try resolve meta from api request if paramType is struct and embeds api.M
	if embedsApiM(paramType) {
		return buildMetaResolver(paramType), nil
	}

	// Try resolve params from target struct fields (including embedded structs)
	if field := findFieldInStruct(target, paramType); field.IsValid() {
		return buildFieldResolver(field, paramType), nil
	}

	return nil, fmt.Errorf("%w: %s", ErrResolveParamType, paramType.String())
}

// findFieldInStruct searches for a field with the specified type in the target struct,
// including embedded anonymous structs and tagged fields using multiple targeted visitor passes.
// Search strategy (in order of priority):
// 1. Direct fields (non-embedded) with matching types
// 2. Fields with api:"in" tag (deep search into tagged structs)
// 3. Embedded anonymous structs (traditional embedding).
func findFieldInStruct(target reflect.Value, paramType reflect.Type) reflect.Value {
	// Priority 1: Search direct fields first (non-recursive, non-embedded only)
	if found := searchDirectFields(target, paramType); found.IsValid() {
		return found
	}

	// Priority 2: Search in fields with api:"in" tag (recursive)
	if found := searchTaggedFields(target, paramType); found.IsValid() {
		return found
	}

	// Priority 3: Search in embedded anonymous structs (recursive)
	if found := searchEmbeddedFields(target, paramType); found.IsValid() {
		return found
	}

	return reflect.Value{}
}

// searchDirectFields searches for direct (non-embedded) fields with matching types.
func searchDirectFields(target reflect.Value, paramType reflect.Type) reflect.Value {
	var foundField reflect.Value

	visitor := reflectx.Visitor{
		VisitField: func(field reflect.StructField, fieldValue reflect.Value, depth int) reflectx.VisitAction {
			// Only check direct (non-embedded) fields
			if !field.Anonymous && reflectx.IsTypeCompatible(fieldValue.Type(), paramType) {
				foundField = fieldValue

				return reflectx.Stop
			}

			return reflectx.Continue
		},
	}

	reflectx.Visit(target, visitor, reflectx.WithDisableRecursive())

	return foundField
}

// searchTaggedFields searches recursively in fields with api:"in" tag.
func searchTaggedFields(target reflect.Value, paramType reflect.Type) reflect.Value {
	var foundField reflect.Value

	visitor := reflectx.Visitor{
		VisitField: func(field reflect.StructField, fieldValue reflect.Value, depth int) reflectx.VisitAction {
			// Skip embedded fields (handled in third pass)
			if field.Anonymous {
				return reflectx.SkipChildren
			}

			// Check for matching type in any depth
			if reflectx.IsTypeCompatible(fieldValue.Type(), paramType) {
				foundField = fieldValue

				return reflectx.Stop
			}

			return reflectx.Continue
		},
	}

	// Use dive tag to recurse into api:"in" tagged fields
	reflectx.Visit(target, visitor, reflectx.WithDiveTag("api", "in"))

	return foundField
}

// searchEmbeddedFields searches recursively in embedded anonymous structs.
func searchEmbeddedFields(target reflect.Value, paramType reflect.Type) reflect.Value {
	var foundField reflect.Value

	visitor := reflectx.Visitor{
		VisitField: func(field reflect.StructField, fieldValue reflect.Value, depth int) reflectx.VisitAction {
			// Only process embedded fields
			if !field.Anonymous {
				return reflectx.SkipChildren
			}

			// Check for matching type
			if reflectx.IsTypeCompatible(fieldValue.Type(), paramType) {
				foundField = fieldValue

				return reflectx.Stop
			}

			return reflectx.Continue
		},
	}

	// Default recursive behavior will handle anonymous embedded fields
	reflectx.Visit(target, visitor)

	return foundField
}

// embedsApiP checks if the given struct type embeds api.P.
func embedsApiP(t reflect.Type) bool {
	return embedsSentinelType(t, apiPType)
}

// embedsApiM checks if the given struct type embeds api.M.
func embedsApiM(t reflect.Type) bool {
	return embedsSentinelType(t, apiMType)
}

// embedsSentinelType checks if the given struct type embeds a sentinel type (api.P or api.M).
func embedsSentinelType(t, sentinelType reflect.Type) bool {
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

		if t == sentinelType {
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

// buildParamsResolver constructs a parameter resolver for parameter structs that embed api.P.
func buildParamsResolver(paramType reflect.Type) ParamResolverFunc {
	t := reflectx.Indirect(paramType)

	return func(ctx fiber.Ctx) (reflect.Value, error) {
		request := contextx.ApiRequest(ctx)
		// Create a new instance of the param type
		paramValue := reflect.New(t)
		if err := request.Params.Decode(paramValue.Interface()); err != nil {
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

// buildMetaResolver constructs a meta resolver for meta structs that embed api.M.
func buildMetaResolver(metaType reflect.Type) ParamResolverFunc {
	t := reflectx.Indirect(metaType)

	return func(ctx fiber.Ctx) (reflect.Value, error) {
		request := contextx.ApiRequest(ctx)
		// Create a new instance of the meta type
		metaValue := reflect.New(t)
		if err := request.Meta.Decode(metaValue.Interface()); err != nil {
			return lo.Empty[reflect.Value](), err
		}

		if err := validator.Validate(metaValue.Interface()); err != nil {
			return lo.Empty[reflect.Value](), err
		}

		if metaType.Kind() == reflect.Pointer {
			return metaValue, nil
		}

		return metaValue.Elem(), nil
	}
}

// buildFieldResolver builds a resolver function for a struct field with type conversion support.
// It handles pointer compatibility and WithLogger method calls if applicable.
func buildFieldResolver(field reflect.Value, targetType reflect.Type) ParamResolverFunc {
	fieldType := field.Type()
	requiresConfigureLogger := hasWithLoggerMethod(fieldType)

	return func(ctx fiber.Ctx) (reflect.Value, error) {
		var resolvedValue reflect.Value

		if requiresConfigureLogger {
			logger := contextx.Logger(ctx)
			resolvedValue = callWithLogger(field, logger)
		} else {
			resolvedValue = field
		}

		// Perform type conversion if needed
		return reflectx.ConvertValue(resolvedValue, targetType)
	}
}

// hasWithLoggerMethod checks if the given type has a WithLogger method.
func hasWithLoggerMethod(t reflect.Type) bool {
	// First try to find method on the type itself
	method, found := t.MethodByName(withLoggerMethod)
	if !found {
		// If not found, and it's not already a pointer, try pointer type
		// because WithLogger might be defined on pointer receiver
		if t.Kind() != reflect.Pointer {
			ptrType := reflect.PointerTo(t)
			method, found = ptrType.MethodByName(withLoggerMethod)
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

// callWithLogger calls the WithLogger method on the value with the given logger.
func callWithLogger(field reflect.Value, logger log.Logger) reflect.Value {
	method := reflectx.FindMethod(field, withLoggerMethod)
	if method.IsValid() {
		// Call WithLogger method with the logger
		results := method.Call([]reflect.Value{reflect.ValueOf(logger)})
		if len(results) > 0 {
			return results[0]
		}
	}

	return field
}
