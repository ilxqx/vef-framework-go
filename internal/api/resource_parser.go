package api

import (
	"fmt"
	"reflect"

	"github.com/gofiber/fiber/v3"
	apiPkg "github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/internal/log"
	"github.com/ilxqx/vef-framework-go/orm"
	"github.com/ilxqx/vef-framework-go/reflectx"
	"github.com/samber/lo"
)

var (
	errorType = reflect.TypeFor[error]()
	dbType    = reflect.TypeFor[orm.Db]()
	logger    = log.Named("api")
)

// parseResource processes a single resource and extracts all its API definitions.
// It creates handlers for each API action and builds a complete resource definition.
func parseResource(resource apiPkg.Resource, db orm.Db, paramResolver *handlerParamResolverManager) (ResourceDefinition, error) {
	resourceAPIs := resource.APIs()
	apiDefinitions := make([]*apiPkg.Definition, 0, len(resourceAPIs))
	defaultVersion := resource.Version()
	resourceName := resource.Name()

	for _, api := range resourceAPIs {
		// Parse the handler method for this API action
		handler, err := parseHandler(lo.PascalCase(api.Action), resource, db, paramResolver)
		if err != nil {
			return nil, fmt.Errorf("failed to parse handler for resource '%s' action '%s': %w",
				resourceName, api.Action, err)
		}

		// Determine the API version (API-specific version overrides resource default)
		apiVersion, _ := lo.Coalesce(api.Version, defaultVersion, apiPkg.VersionV1)

		definition := &apiPkg.Definition{
			Identifier: apiPkg.Identifier{
				Version:  apiVersion,
				Resource: resourceName,
				Action:   api.Action,
			},
			EnableAudit:     api.EnableAudit,
			Timeout:         api.Timeout,
			Public:          api.Public,
			PermissionToken: api.PermissionToken,
			Limit:           api.Limit,
			Handler:         handler,
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

// parseHandler creates a Fiber handler from a resource method.
// It supports both regular handler methods and handler factory methods.
// Handler factory methods take orm.Db as input and return a handler function.
func parseHandler(methodName string, resource apiPkg.Resource, db orm.Db, paramResolver *handlerParamResolverManager) (fiber.Handler, error) {
	target := reflect.ValueOf(resource)
	method, err := findHandlerMethod(target, methodName)
	if err != nil {
		return nil, err
	}

	// Check if this is a handler factory method (takes db, returns handler)
	if isHandlerFactory(method.Type()) {
		if db == nil {
			return nil, fmt.Errorf("handler factory method '%s' requires database connection but none provided", methodName)
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

	return method, fmt.Errorf("api action method '%s' not found in resource '%s'", methodName, target.Type().String())
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
		return fmt.Errorf("handler method '%s' has invalid return type '%s', must be 'error'",
			method.String(), method.Out(0).String())
	}

	// Multiple return values are not allowed
	return fmt.Errorf("handler method '%s' has %d return values, must have at most 1 (error) or none",
		method.String(), numOut)
}

// isHandlerFactory checks if a method is a handler factory function.
// A handler factory has one of these signatures:
//   - func(orm.Db) func(...) [error]       // returns handler
//   - func(orm.Db) (func(...) [error], error)  // returns handler and error
//
// The returned function can have any number of parameters (resolved by handlerParamResolverManager)
// but must have either no return value or a single error return value.
func isHandlerFactory(method reflect.Type) bool {
	// Must have exactly 1 input parameter (orm.Db) and 1 or 2 output parameters
	if method.NumIn() != 1 || (method.NumOut() != 1 && method.NumOut() != 2) {
		return false
	}

	// Check if input parameter is orm.Db
	if method.In(0) != dbType {
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
// and dual return values (handler, error) patterns.
func createHandler(method reflect.Value, db orm.Db) (reflect.Value, error) {
	results := method.Call([]reflect.Value{reflect.ValueOf(db)})

	switch len(results) {
	case 1:
		// func(orm.Db) func(...) [error] pattern
		return results[0], nil

	case 2:
		// func(orm.Db) (func(...) [error], error) pattern
		handler := results[0]
		err := results[1]

		// Check if the error is not nil
		if !err.IsNil() {
			return reflect.Value{}, err.Interface().(error)
		}

		return handler, nil

	default:
		return reflect.Value{}, fmt.Errorf("handler factory method should return 1 or 2 values, got %d", len(results))
	}
}

// parse processes a list of resources and returns a composite resource definition.
// It parses each resource and combines them into a single definition.
func parse(resources []apiPkg.Resource, db orm.Db, paramResolver *handlerParamResolverManager) (ResourceDefinition, error) {
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
