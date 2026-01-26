package api

import (
	"reflect"

	"github.com/gofiber/fiber/v3"

	"github.com/ilxqx/vef-framework-go/security"
)

// Resource defines an API resource that groups related operations.
type Resource interface {
	// Kind returns the resource kind.
	Kind() Kind
	// Name returns the resource name (e.g., "users", "sys/config").
	Name() string
	// Version returns the resource version.
	// Empty string means using Engine's default version.
	Version() string
	// Auth returns the resource authentication configuration.
	Auth() *AuthConfig
	// Operations returns the resource operations.
	Operations() []OperationSpec
}

// RouterStrategy determines how API operations are exposed as HTTP endpoints.
type RouterStrategy interface {
	// Name returns the strategy identifier for logging/debugging.
	Name() string
	// CanHandle returns true if the router can handle the given resource kind.
	CanHandle(kind Kind) bool
	// Setup initializes the router (called once during Mount).
	// Implementations should store the router if needed for Route calls.
	Setup(router fiber.Router) error
	// Route registers an operation with the router.
	Route(handler fiber.Handler, op *Operation)
}

// AuthStrategy handles authentication for a specific auth type.
type AuthStrategy interface {
	// Name returns the strategy name (used in AuthConfig.Strategy).
	Name() string
	// Authenticate validates credentials and returns principal.
	Authenticate(ctx fiber.Ctx, options map[string]any) (*security.Principal, error)
}

// AuthStrategyRegistry manages authentication strategies.
type AuthStrategyRegistry interface {
	// Register adds a strategy to the registry.
	Register(strategy AuthStrategy)
	// Get retrieves a strategy by name.
	Get(name string) (AuthStrategy, bool)
	// Names returns all registered strategy names.
	Names() []string
}

// Middleware represents a processing step in the request pipeline.
type Middleware interface {
	// Name returns the middleware identifier.
	Name() string
	// Order determines execution order.
	// Negative values execute before handler, positive after.
	// Lower values execute first within the same phase.
	Order() int
	// Process handles the request.
	// Call next() to continue the chain.
	Process(ctx fiber.Ctx) error
}

// OperationsProvider provides operation specs.
// Embed types implementing this interface in a resource to contribute operations.
type OperationsProvider interface {
	// Provide returns the operation specs for this provider.
	Provide() []OperationSpec
}

// OperationsCollector collects all operations from a resource.
// This includes operations from embedded providers.
type OperationsCollector interface {
	// Collect gathers all operation specs from a resource.
	// Returns specs from embedded OperationsProviders.
	Collect(resource Resource) []OperationSpec
}

// HandlerResolver resolves a handler from a resource and spec.
type HandlerResolver interface {
	// Resolve finds a handler on the resource and spec.
	// Returns the handler (any type) or an error if not found.
	Resolve(resource Resource, spec OperationSpec) (any, error)
}

// HandlerAdapter converts various handler variants to fiber.Handler.
type HandlerAdapter interface {
	// Adapt converts the handler to a fiber.Handler.
	Adapt(handler any, op *Operation) (fiber.Handler, error)
}

// HandlerParamResolver resolves a handler parameter from the request context.
type HandlerParamResolver interface {
	// Type returns the parameter type this resolver handles.
	Type() reflect.Type
	// Resolve extracts the parameter value from the request.
	Resolve(ctx fiber.Ctx) (reflect.Value, error)
}

// FactoryParamResolver resolves a factory function parameter at startup time.
// Factory functions enable dependency injection at startup while keeping handlers clean.
type FactoryParamResolver interface {
	// Type returns the parameter type this resolver handles.
	Type() reflect.Type
	// Resolve returns the parameter value (called once at startup).
	Resolve() (reflect.Value, error)
}

// Engine is the unified API engine that manages multiple routers.
type Engine interface {
	// Register adds resources to the engine.
	Register(resources ...Resource) error
	// Lookup finds an operation by identifier.
	Lookup(id Identifier) *Operation
	// Mount attaches the engine to a Fiber router.
	Mount(router fiber.Router) error
}
