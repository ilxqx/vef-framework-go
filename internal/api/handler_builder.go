package api

import (
	"reflect"

	"github.com/gofiber/fiber/v3"
)

// buildHandler creates a Fiber handler from a reflected method.
// It supports parameter injection via handlerParamResolverManager and a single optional error return.
func buildHandler(target, handler reflect.Value, paramResolver *HandlerParamResolverManager) (fiber.Handler, error) {
	t := handler.Type()
	numIn := t.NumIn()

	handlerParamResolvers := make([]ParamResolverFunc, numIn)
	for i := range numIn {
		paramType := t.In(i)
		// Resolve parameter value via resolver
		resolver, err := paramResolver.Resolve(target, paramType)
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
