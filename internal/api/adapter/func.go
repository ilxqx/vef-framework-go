package adapter

import (
	"fmt"
	"reflect"

	"github.com/gofiber/fiber/v3"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/internal/api/common"
	"github.com/ilxqx/vef-framework-go/internal/api/handler"
	"github.com/ilxqx/vef-framework-go/internal/api/param"
)

type FuncHandler struct {
	paramResolver   *param.HandlerParamResolverManager
	factoryResolver *param.FactoryParamResolverManager
}

func NewFuncHandler(paramResolver *param.HandlerParamResolverManager, factoryResolver *param.FactoryParamResolverManager) api.HandlerAdapter {
	return &FuncHandler{
		paramResolver:   paramResolver,
		factoryResolver: factoryResolver,
	}
}

func (a *FuncHandler) Adapt(h any, op *api.Operation) (fiber.Handler, error) {
	if funcH, ok := h.(handler.Func); ok {
		return a.adaptHandler(funcH, op)
	}

	return nil, nil
}

func (a *FuncHandler) adaptHandler(funcH handler.Func, op *api.Operation) (fiber.Handler, error) {
	resource := reflect.ValueOf(op.Meta[common.MetaKeyResource].(api.Resource))
	h := funcH.H()

	if funcH.IsFactory() {
		var err error
		if h, err = a.createHandler(h, resource); err != nil {
			return nil, err
		}
	}

	return a.buildHandler(h, resource)
}

func (a *FuncHandler) createHandler(factory, target reflect.Value) (reflect.Value, error) {
	fType := factory.Type()
	numIn := fType.NumIn()
	params := make([]reflect.Value, numIn)

	for i := range numIn {
		paramType := fType.In(i)

		resolverFn, err := a.factoryResolver.Resolve(target, paramType)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("failed to resolve factory parameter %d (type %s): %w", i, paramType, err)
		}

		value, err := resolverFn()
		if err != nil {
			return reflect.Value{}, fmt.Errorf("failed to resolve factory parameter %d: %w", i, err)
		}

		params[i] = value
	}

	results := factory.Call(params)

	switch len(results) {
	case 1:
		return results[0], nil
	case 2:
		if !results[1].IsNil() {
			return reflect.Value{}, results[1].Interface().(error)
		}

		return results[0], nil

	default:
		return reflect.Value{}, fmt.Errorf("%w, got %d", ErrHandlerFactoryInvalidReturn, len(results))
	}
}

func (a *FuncHandler) buildHandler(handler, target reflect.Value) (fiber.Handler, error) {
	hType := handler.Type()
	numIn := hType.NumIn()
	resolvers := make([]param.HandlerParamResolverFunc, numIn)

	for i := range numIn {
		resolver, err := a.paramResolver.Resolve(target, hType.In(i))
		if err != nil {
			return nil, err
		}

		resolvers[i] = resolver
	}

	return func(ctx fiber.Ctx) error {
		params := make([]reflect.Value, numIn)
		for i, resolverFn := range resolvers {
			var err error
			if params[i], err = resolverFn(ctx); err != nil {
				return err
			}
		}

		results := handler.Call(params)
		if len(results) == 0 || results[0].IsNil() {
			return nil
		}

		return results[0].Interface().(error)
	}, nil
}
