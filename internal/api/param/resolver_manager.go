package param

import (
	"fmt"
	"reflect"

	"github.com/gofiber/fiber/v3"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/contextx"
	"github.com/ilxqx/vef-framework-go/internal/api/common"
	"github.com/ilxqx/vef-framework-go/log"
	"github.com/ilxqx/vef-framework-go/reflectx"
	"github.com/ilxqx/vef-framework-go/validator"
)

var (
	loggerType       = reflect.TypeFor[log.Logger]()
	withLoggerMethod = "WithLogger"
)

type HandlerParamResolverFunc func(ctx fiber.Ctx) (reflect.Value, error)

type HandlerParamResolverManager struct {
	resolvers map[reflect.Type]HandlerParamResolverFunc
}

func NewHandlerParamResolverManager(resolvers []api.HandlerParamResolver) *HandlerParamResolverManager {
	resolverMap := make(map[reflect.Type]HandlerParamResolverFunc, len(resolvers))

	for _, resolver := range resolvers {
		t := resolver.Type()
		resolverMap[t] = resolver.Resolve
	}

	return &HandlerParamResolverManager{
		resolvers: resolverMap,
	}
}

func (m *HandlerParamResolverManager) Resolve(target reflect.Value, paramType reflect.Type) (HandlerParamResolverFunc, error) {
	if resolver, ok := m.resolvers[paramType]; ok {
		return resolver, nil
	}

	if embedsApiParams(paramType) || isBuiltinParamsType(paramType) {
		return buildParamsResolver(paramType), nil
	}

	if embedsApiMeta(paramType) || isBuiltinMetaType(paramType) {
		return buildMetaResolver(paramType), nil
	}

	if field := findFieldInStruct(target, paramType); field.IsValid() {
		return buildFieldResolver(field, paramType), nil
	}

	return nil, fmt.Errorf("%w: %s", ErrResolveHandlerParamType, paramType.String())
}

func buildParamsResolver(paramType reflect.Type) HandlerParamResolverFunc {
	t := reflectx.Indirect(paramType)

	return func(ctx fiber.Ctx) (reflect.Value, error) {
		req := common.Request(ctx)
		if req == nil {
			return reflect.Value{}, fmt.Errorf("%w: %w", ErrResolveHandlerParamType, ErrRequestNotFound)
		}

		paramValue := reflect.New(t)
		if err := req.Params.Decode(paramValue.Interface()); err != nil {
			return reflect.Value{}, err
		}

		if err := validator.Validate(paramValue.Interface()); err != nil {
			return reflect.Value{}, err
		}

		if paramType.Kind() == reflect.Pointer {
			return paramValue, nil
		}

		return paramValue.Elem(), nil
	}
}

func buildMetaResolver(metaType reflect.Type) HandlerParamResolverFunc {
	t := reflectx.Indirect(metaType)

	return func(ctx fiber.Ctx) (reflect.Value, error) {
		req := common.Request(ctx)
		if req == nil {
			return reflect.Value{}, fmt.Errorf("%w: %w", ErrResolveHandlerParamType, ErrRequestNotFound)
		}

		metaValue := reflect.New(t)
		if err := req.Meta.Decode(metaValue.Interface()); err != nil {
			return reflect.Value{}, err
		}

		if err := validator.Validate(metaValue.Interface()); err != nil {
			return reflect.Value{}, err
		}

		if metaType.Kind() == reflect.Pointer {
			return metaValue, nil
		}

		return metaValue.Elem(), nil
	}
}

func buildFieldResolver(field reflect.Value, targetType reflect.Type) HandlerParamResolverFunc {
	fieldType := field.Type()
	needsLogger := hasWithLoggerMethod(fieldType)

	return func(ctx fiber.Ctx) (reflect.Value, error) {
		var resolvedValue reflect.Value

		if needsLogger {
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
	if !found && t.Kind() != reflect.Pointer {
		ptrType := reflect.PointerTo(t)
		method, found = ptrType.MethodByName(withLoggerMethod)
	}

	if !found || method.Type.NumIn() != 2 {
		return false
	}

	return loggerType.AssignableTo(method.Type.In(1))
}

func callWithLogger(field reflect.Value, logger log.Logger) reflect.Value {
	if method := reflectx.FindMethod(field, withLoggerMethod); method.IsValid() {
		if results := method.Call([]reflect.Value{reflect.ValueOf(logger)}); len(results) > 0 {
			return results[0]
		}
	}

	return field
}
