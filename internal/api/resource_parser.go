package api

import (
	"fmt"
	"reflect"

	"github.com/gofiber/fiber/v3"
	"github.com/samber/lo"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/internal/log"
	"github.com/ilxqx/vef-framework-go/reflectx"
)

var (
	errorType    = reflect.TypeFor[error]()
	providerType = reflect.TypeFor[api.Provider]()
	logger       = log.Named("api")
)

func parseResource(
	resource api.Resource,
	factoryParamResolver *FactoryParamResolverManager,
	handlerParamResolver *HandlerParamResolverManager,
) (ResourceDefinition, error) {
	resourceApis := collectAllApis(resource)
	apiDefinitions := make([]*api.Definition, 0, len(resourceApis))
	defaultVersion := resource.Version()
	resourceName := resource.Name()

	for _, apiSpec := range resourceApis {
		handler, err := resolveApiHandler(apiSpec, resource, factoryParamResolver, handlerParamResolver)
		if err != nil {
			return nil, fmt.Errorf(
				"failed to resolve handler for resource %q action %q: %w",
				resourceName, apiSpec.Action, err,
			)
		}

		apiVersion := lo.CoalesceOrEmpty(apiSpec.Version, defaultVersion, api.VersionV1)
		definition := &api.Definition{
			Identifier: api.Identifier{
				Version:  apiVersion,
				Resource: resourceName,
				Action:   apiSpec.Action,
			},
			EnableAudit: apiSpec.EnableAudit,
			Timeout:     apiSpec.Timeout,
			Public:      apiSpec.Public,
			PermToken:   apiSpec.PermToken,
			Limit:       apiSpec.Limit,
			Handler:     handler,
		}

		logger.Infof(
			"Registered Api | Resource: %s, Action: %s, Version: %s, Type: %s",
			resourceName,
			apiSpec.Action,
			apiVersion,
			reflect.TypeOf(resource).String(),
		)

		apiDefinitions = append(apiDefinitions, definition)
	}

	return simpleResourceDefinition{
		apis: apiDefinitions,
	}, nil
}

// collectAllApis includes specs from embedded anonymous structs implementing api.Provider.
func collectAllApis(resource api.Resource) []api.Spec {
	var allSpecs []api.Spec

	embeddedSpecs := collectEmbeddedProviderSpecs(resource)
	allSpecs = append(allSpecs, embeddedSpecs...)

	resourceSpecs := resource.Apis()
	allSpecs = append(allSpecs, resourceSpecs...)

	return allSpecs
}

// collectEmbeddedProviderSpecs uses the visitor pattern to avoid manual recursion complexity.
func collectEmbeddedProviderSpecs(resource api.Resource) []api.Spec {
	var specs []api.Spec

	visitor := reflectx.Visitor{
		VisitField: func(field reflect.StructField, fieldValue reflect.Value, depth int) reflectx.VisitAction {
			if !field.Anonymous {
				return reflectx.Continue
			}

			if isProviderImplementation(fieldValue) {
				if provider, ok := fieldValue.Interface().(api.Provider); ok {
					spec := provider.Provide()
					specs = append(specs, spec)
					logger.Infof(
						"Collected Api spec from embedded provider: %s.%s",
						field.Type.String(), spec.Action,
					)
				}
			}

			return reflectx.Continue
		},
	}

	reflectx.VisitOf(resource, visitor)

	return specs
}

func isProviderImplementation(value reflect.Value) bool {
	return value.Type().Implements(providerType)
}

// resolveApiHandler prioritizes explicit Handler over Action-based method lookup.
func resolveApiHandler(
	apiSpec api.Spec,
	resource api.Resource,
	factoryParamResolver *FactoryParamResolverManager,
	handlerParamResolver *HandlerParamResolverManager,
) (fiber.Handler, error) {
	if apiSpec.Handler != nil {
		return parseProvidedHandler(apiSpec.Handler, resource, factoryParamResolver, handlerParamResolver)
	}

	return parseHandler(lo.PascalCase(apiSpec.Action), resource, factoryParamResolver, handlerParamResolver)
}

func parseProvidedHandler(
	handlerValue any,
	resource api.Resource,
	factoryParamResolver *FactoryParamResolverManager,
	handlerParamResolver *HandlerParamResolverManager,
) (fiber.Handler, error) {
	if handlerValue == nil {
		return nil, ErrProvidedHandlerNil
	}

	handlerReflect := reflect.ValueOf(handlerValue)
	if handlerReflect.Kind() != reflect.Func {
		return nil, fmt.Errorf("%w, got %s", ErrProvidedHandlerMustFunc, handlerReflect.Kind())
	}

	if handlerReflect.IsNil() {
		return nil, ErrProvidedHandlerFuncNil
	}

	target := reflect.ValueOf(resource)

	// Factory functions support dependency injection at startup time
	if isHandlerFactory(handlerReflect.Type()) {
		handler, err := createHandler(handlerReflect, target, factoryParamResolver)
		if err != nil {
			return nil, fmt.Errorf("failed to create handler from factory: %w", err)
		}

		return buildHandler(target, handler, handlerParamResolver)
	}

	if err := checkHandlerMethod(handlerReflect.Type()); err != nil {
		return nil, err
	}

	return buildHandler(target, handlerReflect, handlerParamResolver)
}

// parseHandler supports both regular handlers and factory methods with arbitrary parameters.
func parseHandler(
	methodName string,
	resource api.Resource,
	factoryParamResolver *FactoryParamResolverManager,
	handlerParamResolver *HandlerParamResolverManager,
) (fiber.Handler, error) {
	target := reflect.ValueOf(resource)

	method, err := findHandlerMethod(target, methodName)
	if err != nil {
		return nil, err
	}

	// Factory methods allow dependencies to be injected at resource registration
	if isHandlerFactory(method.Type()) {
		handler, err := createHandler(method, target, factoryParamResolver)
		if err != nil {
			return nil, fmt.Errorf("failed to create handler for method %q: %w", methodName, err)
		}

		return buildHandler(target, handler, handlerParamResolver)
	}

	if err = checkHandlerMethod(method.Type()); err != nil {
		return nil, err
	}

	return buildHandler(target, method, handlerParamResolver)
}

func findHandlerMethod(target reflect.Value, methodName string) (reflect.Value, error) {
	method := reflectx.FindMethod(target, methodName)
	if method.IsValid() {
		return method, nil
	}

	return method, fmt.Errorf("%w %q in resource %q", ErrApiMethodNotFound, methodName, target.Type().String())
}

// checkHandlerMethod allows flexible parameters (resolved at runtime)
// but constrains return values to maintain predictable error handling.
func checkHandlerMethod(method reflect.Type) error {
	numOut := method.NumOut()

	if numOut == 0 {
		return nil
	}

	if numOut == 1 {
		if method.Out(0) == errorType {
			return nil
		}

		return fmt.Errorf("%w: %q -> %q",
			ErrHandlerMethodInvalidReturn, method.String(), method.Out(0).String())
	}

	return fmt.Errorf("%w: %q has %d returns",
		ErrHandlerMethodTooManyReturns, method.String(), numOut)
}

// isHandlerFactory checks for factory signatures that return handler closures.
// Factory functions enable dependency injection at startup while keeping handlers clean.
//
// Supported signatures:
//   - func(...any) func(...) [error]
//   - func(...any) (func(...) [error], error)
//
// Parameters resolved via FactoryParamResolver; returned handler validated via checkHandlerMethod.
func isHandlerFactory(method reflect.Type) bool {
	numOut := method.NumOut()
	if numOut < 1 || numOut > 2 {
		return false
	}

	handlerType := method.Out(0)
	if handlerType.Kind() != reflect.Func {
		return false
	}

	if checkHandlerMethod(handlerType) != nil {
		return false
	}

	if numOut == 2 && method.Out(1) != errorType {
		return false
	}

	return true
}

// createHandler executes factory functions at startup to fail fast on misconfiguration.
func createHandler(
	method reflect.Value,
	target reflect.Value,
	factoryParamResolver *FactoryParamResolverManager,
) (reflect.Value, error) {
	var (
		methodType    = method.Type()
		numIn         = methodType.NumIn()
		factoryParams = make([]reflect.Value, numIn)
	)

	for i := range numIn {
		paramType := methodType.In(i)

		resolverFn, err := factoryParamResolver.Resolve(target, paramType)
		if err != nil {
			return reflect.Value{}, fmt.Errorf(
				"failed to resolve factory parameter %d (type %s): %w",
				i, paramType, err,
			)
		}

		factoryParams[i] = resolverFn()
	}

	results := method.Call(factoryParams)

	switch len(results) {
	case 1:
		return results[0], nil

	case 2:
		handler := results[0]
		err := results[1]

		if !err.IsNil() {
			return reflect.Value{}, err.Interface().(error)
		}

		return handler, nil

	default:
		return reflect.Value{}, fmt.Errorf(
			"%w, got %d",
			ErrHandlerFactoryInvalidReturn,
			len(results),
		)
	}
}

func parse(
	resources []api.Resource,
	factoryParamResolver *FactoryParamResolverManager,
	handlerParamResolver *HandlerParamResolverManager,
) (ResourceDefinition, error) {
	definitions := make([]ResourceDefinition, 0, len(resources))
	for _, resource := range resources {
		definition, err := parseResource(resource, factoryParamResolver, handlerParamResolver)
		if err != nil {
			return nil, err
		}

		if definition != nil {
			definitions = append(definitions, definition)
		}
	}

	return compositeResourceDefinition{
		definitions: definitions,
	}, nil
}
