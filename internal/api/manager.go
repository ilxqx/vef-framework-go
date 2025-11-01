package api

import (
	"github.com/gofiber/fiber/v3/middleware/timeout"
	"github.com/puzpuzpuz/xsync/v4"

	apiPkg "github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/orm"
)

// apiManager implements the Manager interface using xsync.Map for thread-safe operations.
type apiManager struct {
	apis *xsync.Map[apiPkg.Identifier, *apiPkg.Definition]
}

// Register adds a new Api definition to the manager.
// The handler will be wrapped with timeout middleware.
// Returns an error if an Api with the same identifier already exists.
func (m *apiManager) Register(api *apiPkg.Definition) error {
	if existing, loaded := m.apis.LoadOrStore(api.Identifier, wrapHandler(api)); loaded {
		return &DuplicateApiError{
			Identifier: api.Identifier,
			Existing:   existing,
			New:        api,
		}
	}

	return nil
}

// Remove removes an Api definition by its identifier.
func (m *apiManager) Remove(id apiPkg.Identifier) {
	m.apis.Delete(id)
}

// Lookup retrieves an Api definition by its identifier.
// Returns nil if the definition is not found.
func (m *apiManager) Lookup(id apiPkg.Identifier) *apiPkg.Definition {
	if api, ok := m.apis.Load(id); ok {
		return api
	}

	return nil
}

// List returns all registered Api definitions.
func (m *apiManager) List() []*apiPkg.Definition {
	var definitions []*apiPkg.Definition
	m.apis.Range(func(key apiPkg.Identifier, value *apiPkg.Definition) bool {
		definitions = append(definitions, value)

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

// NewManager creates a new Api manager and registers all provided resources.
// Returns an error if any API registration fails, including duplicate definitions.
func NewManager(resources []apiPkg.Resource, db orm.Db, paramResolver *HandlerParamResolverManager) (apiPkg.Manager, error) {
	manager := &apiManager{
		apis: xsync.NewMap[apiPkg.Identifier, *apiPkg.Definition](),
	}

	definition, err := parse(resources, db, paramResolver)
	if err != nil {
		return nil, err
	}

	if err = definition.Register(manager); err != nil {
		return nil, err
	}

	return manager, nil
}
