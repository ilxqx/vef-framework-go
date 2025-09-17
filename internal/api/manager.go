package api

import (
	"sync"

	"github.com/gofiber/fiber/v3/middleware/timeout"
	apiPkg "github.com/ilxqx/vef-framework-go/api"
)

// apiManager implements the Manager interface using sync.Map for thread-safe operations.
type apiManager struct {
	apis sync.Map // map[Identifier]*Definition
}

// Register adds a new API definition to the manager.
// The handler will be wrapped with timeout middleware.
func (m *apiManager) Register(api *apiPkg.Definition) {
	m.apis.Store(api.Identifier, wrapHandler(api))
}

// Remove removes an API definition by its identifier.
func (m *apiManager) Remove(id apiPkg.Identifier) {
	m.apis.Delete(id)
}

// Lookup retrieves an API definition by its identifier.
// Returns nil if the definition is not found.
func (m *apiManager) Lookup(id apiPkg.Identifier) *apiPkg.Definition {
	if api, ok := m.apis.Load(id); ok {
		return api.(*apiPkg.Definition)
	}

	return nil
}

// List returns all registered API definitions.
func (m *apiManager) List() []*apiPkg.Definition {
	var definitions []*apiPkg.Definition
	m.apis.Range(func(key, value any) bool {
		definitions = append(definitions, value.(*apiPkg.Definition))
		return true
	})
	return definitions
}

// wrapHandler wraps the original handler with timeout middleware.
// If no timeout is specified, it defaults to 30 seconds.
func wrapHandler(api *apiPkg.Definition) *apiPkg.Definition {
	originalHandler := api.Handler
	handler := timeout.New(
		originalHandler,
		timeout.Config{
			Timeout: api.GetTimeout(),
		},
	)
	api.Handler = handler

	return api
}

// newManager creates a new API manager and registers all provided resources.
func newManager(resources []apiPkg.Resource, paramResolver *handlerParamResolverManager) (apiPkg.Manager, error) {
	manager := new(apiManager)
	definition, err := parse(resources, paramResolver)
	if err != nil {
		return nil, err
	}

	definition.Register(manager)
	return manager, nil
}
