package api

import (
	"github.com/gofiber/fiber/v3/middleware/timeout"
	"github.com/puzpuzpuz/xsync/v4"

	"github.com/ilxqx/vef-framework-go/api"
)

// apiManager uses xsync.Map for thread-safe concurrent access.
type apiManager struct {
	apis *xsync.Map[api.Identifier, *api.Definition]
}

func (m *apiManager) Register(apiDef *api.Definition) error {
	if existing, loaded := m.apis.LoadOrStore(apiDef.Identifier, wrapHandler(apiDef)); loaded {
		identifier := apiDef.Identifier

		return &DuplicateError{
			BaseError: BaseError{
				Identifier: &identifier,
			},
			Existing: existing,
		}
	}

	return nil
}

func (m *apiManager) Remove(id api.Identifier) {
	m.apis.Delete(id)
}

func (m *apiManager) Lookup(id api.Identifier) *api.Definition {
	if apiDef, ok := m.apis.Load(id); ok {
		return apiDef
	}

	return nil
}

func (m *apiManager) List() []*api.Definition {
	var definitions []*api.Definition
	m.apis.Range(func(key api.Identifier, value *api.Definition) bool {
		definitions = append(definitions, value)

		return true
	})

	return definitions
}

func wrapHandler(apiDef *api.Definition) *api.Definition {
	originalHandler := apiDef.Handler
	handler := timeout.New(
		originalHandler,
		timeout.Config{
			Timeout: apiDef.GetTimeout(),
		},
	)
	apiDef.Handler = handler

	return apiDef
}

func NewManager(
	resources []api.Resource,
	factoryParamResolver *FactoryParamResolverManager,
	handlerParamResolver *HandlerParamResolverManager,
) (api.Manager, error) {
	manager := &apiManager{
		apis: xsync.NewMap[api.Identifier, *api.Definition](),
	}

	definition, err := parse(resources, factoryParamResolver, handlerParamResolver)
	if err != nil {
		return nil, err
	}

	if err = definition.Register(manager); err != nil {
		return nil, err
	}

	return manager, nil
}
