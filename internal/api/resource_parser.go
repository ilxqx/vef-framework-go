package api

import (
	"fmt"
	"reflect"

	"github.com/gofiber/fiber/v3"
	apiPkg "github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/internal/log"
	"github.com/ilxqx/vef-framework-go/reflectx"
	"github.com/samber/lo"
)

var (
	errorType = reflect.TypeFor[error]()
	logger    = log.Named("api")
)

func parseResource(resource apiPkg.Resource, paramResolver *handlerParamResolverManager) (ResourceDefinition, error) {
	apis := make([]*apiPkg.Definition, 0, len(resource.APIs()))
	version := resource.Version()
	name := resource.Name()

	for _, api := range resource.APIs() {
		handler, err := parseHandler(lo.PascalCase(api.Action), resource, paramResolver)
		if err != nil {
			return nil, err
		}

		version, _ := lo.Coalesce(api.Version, version, apiPkg.VersionV1)
		definition := &apiPkg.Definition{
			Identifier: apiPkg.Identifier{
				Version:  version,
				Resource: name,
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
			"Found API in '%s' | Resource: %s, Action: %s, Version: %s",
			reflect.TypeOf(resource).String(),
			name,
			api.Action,
			version,
		)

		apis = append(apis, definition)
	}

	return simpleResourceDefinition{
		apis: apis,
	}, nil
}

func parseHandler(methodName string, resource apiPkg.Resource, paramResolver *handlerParamResolverManager) (fiber.Handler, error) {
	target := reflect.ValueOf(resource)
	method, err := findHandlerMethod(target, methodName)
	if err != nil {
		return nil, err
	}

	if err = checkHandlerMethod(method); err != nil {
		return nil, err
	}

	return buildHandler(target, method, paramResolver)
}

func findHandlerMethod(target reflect.Value, methodName string) (reflect.Value, error) {
	method := reflectx.FindMethod(target, methodName)
	if method.IsValid() {
		return method, nil
	}

	return method, fmt.Errorf("api action method '%s' not found in resource '%s'", methodName, target.Type().String())
}

func checkHandlerMethod(method reflect.Value) error {
	t := method.Type()
	// Allow any number of parameters; they will be injected.
	// Still enforce at most one return value and that it is error.
	if t.NumOut() > 1 || (t.NumOut() == 1 && t.Out(0) != errorType) {
		return fmt.Errorf("method '%s' must have a single return value of type '%s' or no return value", t.String(), errorType.Name())
	}
	return nil
}

// parse processes a list of resources and returns a composite resource definition.
// It parses each resource and combines them into a single definition.
func parse(resources []apiPkg.Resource, paramResolver *handlerParamResolverManager) (ResourceDefinition, error) {
	definitions := make([]ResourceDefinition, 0, len(resources))
	for _, resource := range resources {
		definition, err := parseResource(resource, paramResolver)
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
