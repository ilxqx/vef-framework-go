package resolver

import (
	"fmt"

	"github.com/ilxqx/vef-framework-go/api"
)

type REST struct{}

func NewRest() api.HandlerResolver {
	return new(REST)
}

func (r *REST) Resolve(resource api.Resource, spec api.OperationSpec) (any, error) {
	if resource.Kind() != api.KindREST {
		return nil, nil
	}

	if spec.Handler == nil {
		return nil, fmt.Errorf("handler is required for REST operations (resource: %s, action: %s)",
			resource.Name(), spec.Action)
	}

	return resolveHandlerFromSpec(spec, resource)
}
