package api

import (
	"fmt"
	"reflect"

	"github.com/gofiber/fiber/v3"
	"github.com/samber/lo"

	apiPkg "github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/internal/log"
	"github.com/ilxqx/vef-framework-go/orm"
	"github.com/ilxqx/vef-framework-go/reflectx"
)

var (
	errorType    = reflect.TypeFor[error]()
	dbType       = reflect.TypeFor[orm.Db]()
	providerType = reflect.TypeFor[apiPkg.Provider]()
	logger       = log.Named("api")
)

// parseResource processes a single resource and extracts all its API definitions.
// It creates handlers for each API action and builds a complete resource definition.
func parseResource(resource apiPkg.Resource, db orm.Db, paramResolver *HandlerParamResolverManager) (ResourceDefinition, error) {
	resourceAPIs := collectAllAPIs(resource)
	apiDefinitions := make([]*apiPkg.Definition, 0, len(resourceAPIs))
	defaultVersion := resource.Version()
	resourceName := resource.Name()

	for _, api := range resourceAPIs {
		// Parse the handler for this API specification
		handler, err := resolveAPIHandler(api, resource, db, paramResolver)
		if err != nil {
			return nil, fmt.Errorf(
				"failed to resolve handler for resource '%s' action '%s': %w",
				resourceName, api.Action, err,
			)
		}

		// Determine the API version (API-specific version overrides resource default)
		apiVersion, _ := lo.Coalesce(api.Version, defaultVersion, apiPkg.VersionV1)

		definition := &apiPkg.Definition{
			Identifier: apiPkg.Identifier{
				Version:  apiVersion,
				Resource: resourceName,
				Action:   api.Action,
			},
			EnableAudit: api.EnableAudit,
			Timeout:     api.Timeout,
			Public:      api.Public,
			PermToken:   api.PermToken,
			Limit:       api.Limit,
			Handler:     handler,
		}

		logger.Infof(
			"Registered API | Resource: %s, Action: %s, Version: %s, Type: %s",
			resourceName,
			api.Action,
			apiVersion,
			reflect.TypeOf(resource).String(),
		)

		apiDefinitions = append(apiDefinitions, definition)
	}

	return simpleResourceDefinition{
		apis: apiDefinitions,
	}, nil
}

// collectAllAPIs collects all API specs from a resource, including those from embedded anonymous structs
// that implement the api.Provider interface.
func collectAllAPIs(resource apiPkg.Resource) []apiPkg.Spec {
	var allSpecs []apiPkg.Spec

	// First, collect specs from embedded anonymous structs that implement api.Provider
	embeddedSpecs := collectEmbeddedProviderSpecs(resource)
	allSpecs = append(allSpecs, embeddedSpecs...)

	// Then, collect specs from the resource's own APIs() method
	resourceSpecs := resource.APIs()
	allSpecs = append(allSpecs, resourceSpecs...)

	return allSpecs
}

// collectEmbeddedProviderSpecs recursively scans for embedded anonymous structs that implement api.Provider
// and collects their API specifications using the visitor pattern.
func collectEmbeddedProviderSpecs(resource apiPkg.Resource) []apiPkg.Spec {
	var specs []apiPkg.Spec

	visitor := reflectx.Visitor{
		VisitField: func(field reflect.StructField, fieldValue reflect.Value, depth int) reflectx.VisitAction {
			// Only process anonymous (embedded) fields
			if !field.Anonymous {
				return reflectx.Continue
			}

			// Check if the field implements api.Provider interface
			if isProviderImplementation(fieldValue) {
				if provider, ok := fieldValue.Interface().(apiPkg.Provider); ok {
					spec := provider.Provide()
					specs = append(specs, spec)
					logger.Infof(
						"Collected API spec from embedded provider: %s.%s",
						field.Type.String(), spec.Action,
					)
				}
			}

			// Continue recursive traversal into embedded fields
			return reflectx.Continue
		},
	}

	reflectx.VisitOf(resource, visitor)

	return specs
}

// isProviderImplementation checks if a value implements the api.Provider interface.
func isProviderImplementation(value reflect.Value) bool {
	// Get the interface type of the value
	valueType := value.Type()

	// Check if the type implements Provider interface
	return valueType.Implements(providerType)
}

// resolveAPIHandler resolves the appropriate handler for an API specification.
// It prioritizes the Handler field if provided, otherwise falls back to Action-based method lookup.
func resolveAPIHandler(api apiPkg.Spec, resource apiPkg.Resource, db orm.Db, paramResolver *HandlerParamResolverManager) (fiber.Handler, error) {
	if api.Handler != nil {
		// Use provided Handler field
		return parseProvidedHandler(api.Handler, resource, db, paramResolver)
	}

	// Fallback to Action-based method lookup
	return parseHandler(lo.PascalCase(api.Action), resource, db, paramResolver)
}

// parseProvidedHandler creates a Fiber handler from a user-provided handler value.
// The handler value must be a non-nil function that conforms to the framework's handler signature.
// It supports both regular handler functions and handler factory functions.
func parseProvidedHandler(handlerValue any, resource apiPkg.Resource, db orm.Db, paramResolver *HandlerParamResolverManager) (fiber.Handler, error) {
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

	// Check if this is a handler factory function (takes db, returns handler)
	if isHandlerFactory(handlerReflect.Type()) {
		if db == nil {
			return nil, ErrHandlerFactoryRequireDB
		}

		handler, err := createHandler(handlerReflect, db)
		if err != nil {
			return nil, fmt.Errorf("failed to create handler from factory: %w", err)
		}

		return buildHandler(target, handler, paramResolver)
	}

	// Regular handler function validation
	if err := checkHandlerMethod(handlerReflect.Type()); err != nil {
		return nil, err
	}

	return buildHandler(target, handlerReflect, paramResolver)
}

// parseHandler creates a Fiber handler from a resource method.
// It supports both regular handler methods and handler factory methods.
// Handler factory methods take orm.Db as input and return a handler function.
func parseHandler(methodName string, resource apiPkg.Resource, db orm.Db, paramResolver *HandlerParamResolverManager) (fiber.Handler, error) {
	target := reflect.ValueOf(resource)

	method, err := findHandlerMethod(target, methodName)
	if err != nil {
		return nil, err
	}

	// Check if this is a handler factory method (takes db, returns handler)
	if isHandlerFactory(method.Type()) {
		if db == nil {
			return nil, fmt.Errorf("%w: %s", ErrHandlerFactoryMethodRequireDB, methodName)
		}

		handler, err := createHandler(method, db)
		if err != nil {
			return nil, fmt.Errorf("failed to create handler for method '%s': %w", methodName, err)
		}

		return buildHandler(target, handler, paramResolver)
	}

	// Regular handler method validation
	if err = checkHandlerMethod(method.Type()); err != nil {
		return nil, err
	}

	return buildHandler(target, method, paramResolver)
}

// findHandlerMethod locates a method by name on the given resource target.
// Returns the method if found, or an error with details about the resource type.
func findHandlerMethod(target reflect.Value, methodName string) (reflect.Value, error) {
	method := reflectx.FindMethod(target, methodName)
	if method.IsValid() {
		return method, nil
	}

	return method, fmt.Errorf("%w '%s' in resource '%s'", ErrAPIMethodNotFound, methodName, target.Type().String())
}

// checkHandlerMethod validates that a method conforms to the framework's handler signature.
// Handler methods can have any number of parameters (resolved via handlerParamResolverManager)
// but must return either nothing or a single error value.
func checkHandlerMethod(method reflect.Type) error {
	numOut := method.NumOut()

	// No return value is valid
	if numOut == 0 {
		return nil
	}

	// Single return value must be error type
	if numOut == 1 {
		if method.Out(0) == errorType {
			return nil
		}

		return fmt.Errorf("%w: '%s' -> '%s'",
			ErrHandlerMethodInvalidReturn, method.String(), method.Out(0).String())
	}

	// Multiple return values are not allowed
	return fmt.Errorf("%w: '%s' has %d returns",
		ErrHandlerMethodTooManyReturns, method.String(), numOut)
}

// isHandlerFactory checks if a method is a handler factory function.
// A handler factory has one of these signatures:
//   - func() func(...) [error]                 // returns handler (no parameters)
//   - func() (func(...) [error], error)        // returns handler and error (no parameters)
//   - func(orm.Db) func(...) [error]           // returns handler (with db parameter)
//   - func(orm.Db) (func(...) [error], error)  // returns handler and error (with db parameter)
//
// The returned function can have any number of parameters (resolved by handlerParamResolverManager)
// but must have either no return value or a single error return value.
func isHandlerFactory(method reflect.Type) bool {
	// Must have 0 or 1 input parameter and 1 or 2 output parameters
	numIn := method.NumIn()
	if numIn > 1 || (method.NumOut() != 1 && method.NumOut() != 2) {
		return false
	}

	// If there's an input parameter, it must be orm.Db
	if numIn == 1 && method.In(0) != dbType {
		return false
	}

	// First return value must be a valid handler function
	handlerType := method.Out(0)
	if handlerType.Kind() != reflect.Func {
		return false
	}

	if checkHandlerMethod(handlerType) != nil {
		return false
	}

	// If there's a second return value, it must be error
	if method.NumOut() == 2 {
		return method.Out(1) == errorType
	}

	return true
}

// createHandler invokes a handler factory function with the provided database connection
// and returns the created handler function. Supports both single return value (handler)
// and dual return values (handler, error) patterns. Also supports factories with no parameters.
func createHandler(method reflect.Value, db orm.Db) (reflect.Value, error) {
	// Determine if factory needs db parameter
	var results []reflect.Value
	if method.Type().NumIn() == 0 {
		// func() func(...) [error] or func() (func(...) [error], error)
		results = method.Call([]reflect.Value{})
	} else {
		// func(orm.Db) func(...) [error] or func(orm.Db) (func(...) [error], error)
		results = method.Call([]reflect.Value{reflect.ValueOf(db)})
	}

	switch len(results) {
	case 1:
		// func([orm.Db]) func(...) [error] pattern
		return results[0], nil

	case 2:
		// func([orm.Db]) (func(...) [error], error) pattern
		handler := results[0]
		err := results[1]

		// Check if the error is not nil
		if !err.IsNil() {
			return reflect.Value{}, err.Interface().(error)
		}

		return handler, nil

	default:
		return reflect.Value{}, fmt.Errorf("%w, got %d", ErrHandlerFactoryInvalidReturn, len(results))
	}
}

// parse processes a list of resources and returns a composite resource definition.
// It parses each resource and combines them into a single definition.
func parse(resources []apiPkg.Resource, db orm.Db, paramResolver *HandlerParamResolverManager) (ResourceDefinition, error) {
	definitions := make([]ResourceDefinition, 0, len(resources))
	for _, resource := range resources {
		definition, err := parseResource(resource, db, paramResolver)
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
