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

type ParamResolverFunc func(ctx fiber.Ctx) (reflect.Value, error)

// HandlerParamResolverManager uses exact type matching for resolution.
// If needed in the future, assignable/interface matching (e.g., Implements/AssignableTo)
// can be added without changing the public API.
type HandlerParamResolverManager struct {
	resolvers map[reflect.Type]ParamResolverFunc
}

// NewHandlerParamResolverManager merges preset and user-provided resolvers.
// User-provided resolvers override preset ones for the same type.
// Intended to be used by the DI layer to assemble resolvers from groups.
func NewHandlerParamResolverManager(userResolvers []api.HandlerParamResolver) *HandlerParamResolverManager {
	merged := make(map[reflect.Type]ParamResolverFunc, len(userResolvers)+len(presetParamResolvers))
	// Preset first
	for _, resolver := range presetParamResolvers {
		t := resolver.Type()
		merged[t] = resolver.Resolve
	}
	// User-provided override
	for _, resolver := range userResolvers {
		t := resolver.Type()
		merged[t] = resolver.Resolve
	}

	return &HandlerParamResolverManager{
		resolvers: merged,
	}
}

func (m *HandlerParamResolverManager) Resolve(target reflect.Value, paramType reflect.Type) (ParamResolverFunc, error) {
	if resolver, ok := m.resolvers[paramType]; ok {
		return resolver, nil
	}

	if embedsApiP(paramType) {
		return buildParamsResolver(paramType), nil
	}

	if embedsApiM(paramType) {
		return buildMetaResolver(paramType), nil
	}

	if field := findFieldInStruct(target, paramType); field.IsValid() {
		return buildFieldResolver(field, paramType), nil
	}

	return nil, fmt.Errorf("%w: %s", ErrResolveHandlerParamType, paramType.String())
}

// findFieldInStruct uses a multi-pass search strategy to balance explicitness with flexibility:
// 1. Direct fields first (most explicit, avoids ambiguity)
// 2. Fields with api:"in" tag (explicit opt-in for deep nesting)
// 3. Embedded anonymous structs last (traditional Go embedding, most implicit).
func findFieldInStruct(target reflect.Value, paramType reflect.Type) reflect.Value {
	if found := searchDirectFields(target, paramType); found.IsValid() {
		return found
	}

	if found := searchTaggedFields(target, paramType); found.IsValid() {
		return found
	}

	if found := searchEmbeddedFields(target, paramType); found.IsValid() {
		return found
	}

	return reflect.Value{}
}

func searchDirectFields(target reflect.Value, paramType reflect.Type) reflect.Value {
	var foundField reflect.Value

	visitor := reflectx.Visitor{
		VisitField: func(field reflect.StructField, fieldValue reflect.Value, depth int) reflectx.VisitAction {
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

func searchTaggedFields(target reflect.Value, paramType reflect.Type) reflect.Value {
	var foundField reflect.Value

	visitor := reflectx.Visitor{
		VisitField: func(field reflect.StructField, fieldValue reflect.Value, depth int) reflectx.VisitAction {
			if field.Anonymous {
				return reflectx.SkipChildren
			}

			if reflectx.IsTypeCompatible(fieldValue.Type(), paramType) {
				foundField = fieldValue

				return reflectx.Stop
			}

			return reflectx.Continue
		},
	}

	// Recurse only into fields explicitly tagged with api:"in"
	reflectx.Visit(target, visitor, reflectx.WithDiveTag("api", "in"))

	return foundField
}

func searchEmbeddedFields(target reflect.Value, paramType reflect.Type) reflect.Value {
	var foundField reflect.Value

	visitor := reflectx.Visitor{
		VisitField: func(field reflect.StructField, fieldValue reflect.Value, depth int) reflectx.VisitAction {
			if !field.Anonymous {
				return reflectx.SkipChildren
			}

			if reflectx.IsTypeCompatible(fieldValue.Type(), paramType) {
				foundField = fieldValue

				return reflectx.Stop
			}

			return reflectx.Continue
		},
	}

	reflectx.Visit(target, visitor)

	return foundField
}

func embedsApiP(t reflect.Type) bool {
	return embedsSentinelType(t, apiPType)
}

func embedsApiM(t reflect.Type) bool {
	return embedsSentinelType(t, apiMType)
}

// embedsSentinelType uses breadth-first search to handle deeply nested embeddings correctly.
func embedsSentinelType(t, sentinelType reflect.Type) bool {
	t = reflectx.Indirect(t)
	if t.Kind() != reflect.Struct {
		return false
	}

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

func buildParamsResolver(paramType reflect.Type) ParamResolverFunc {
	t := reflectx.Indirect(paramType)

	return func(ctx fiber.Ctx) (reflect.Value, error) {
		request := contextx.ApiRequest(ctx)

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

func buildMetaResolver(metaType reflect.Type) ParamResolverFunc {
	t := reflectx.Indirect(metaType)

	return func(ctx fiber.Ctx) (reflect.Value, error) {
		request := contextx.ApiRequest(ctx)

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

// buildFieldResolver handles pointer compatibility and WithLogger method calls
// to support fields that need request-scoped configuration.
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

		return reflectx.ConvertValue(resolvedValue, targetType)
	}
}

func hasWithLoggerMethod(t reflect.Type) bool {
	method, found := t.MethodByName(withLoggerMethod)
	if !found {
		// Check pointer receiver if not already a pointer
		// because WithLogger is often defined on pointer receivers
		if t.Kind() != reflect.Pointer {
			ptrType := reflect.PointerTo(t)
			method, found = ptrType.MethodByName(withLoggerMethod)
		}
	}

	if !found {
		return false
	}

	if method.Type.NumIn() != 2 {
		return false
	}

	paramType := method.Type.In(1)

	return loggerType.AssignableTo(paramType)
}

func callWithLogger(field reflect.Value, logger log.Logger) reflect.Value {
	method := reflectx.FindMethod(field, withLoggerMethod)
	if method.IsValid() {
		results := method.Call([]reflect.Value{reflect.ValueOf(logger)})
		if len(results) > 0 {
			return results[0]
		}
	}

	return field
}
