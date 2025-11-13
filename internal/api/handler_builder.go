package api

import (
	"reflect"

	"github.com/gofiber/fiber/v3"
)

// buildHandler resolves all handler parameters at startup and caches resolvers
// to avoid repeated type lookups during request handling.
func buildHandler(target, handler reflect.Value, paramResolver *HandlerParamResolverManager) (fiber.Handler, error) {
	var (
		handlerType           = handler.Type()
		numIn                 = handlerType.NumIn()
		handlerParamResolvers = make([]ParamResolverFunc, numIn)
	)

	for i := range numIn {
		resolver, err := paramResolver.Resolve(target, handlerType.In(i))
		if err != nil {
			return nil, err
		}

		handlerParamResolvers[i] = resolver
	}

	return func(ctx fiber.Ctx) (err error) {
		handlerParams := make([]reflect.Value, numIn)
		for i, resolverFn := range handlerParamResolvers {
			if handlerParams[i], err = resolverFn(ctx); err != nil {
				return err
			}
		}

		results := handler.Call(handlerParams)
		if len(results) == 0 {
			return nil
		}

		result := results[0]
		if result.IsNil() {
			return nil
		}

		return result.Interface().(error)
	}, nil
}
