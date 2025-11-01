package api

import (
	"reflect"

	"github.com/gofiber/fiber/v3"
)

// Manager defines the interface for managing Api definitions.
// It provides methods to register, remove, and lookup Api definitions by their identifiers.
type Manager interface {
	// Register adds a new Api definition to the manager.
	// Returns an error if an Api with the same identifier already exists.
	Register(api *Definition) error
	// Remove removes an Api definition by its identifier.
	Remove(id Identifier)
	// Lookup retrieves an Api definition by its identifier.
	// Returns nil if the definition is not found.
	Lookup(id Identifier) *Definition
	// List returns all registered Api definitions.
	List() []*Definition
}

// Resource represents an Api resource that contains multiple Api specifications.
// It defines the version, name, and list of Api specifications for a resource.
type Resource interface {
	// Version returns the version of the resource.
	Version() string
	// Name returns the name of the resource.
	Name() string
	// Apis returns the list of Api specifications for this resource.
	Apis() []Spec
}

// HandlerParamResolver declares a pluggable strategy to resolve a single handler
// parameter from the current request context. Implementations should be pure and
// fast, as they are invoked for every handler call.
//
// Contract:
//   - Type() must return the exact parameter type this resolver handles.
//   - Resolve(ctx) returns the concrete value for that type (or nil if unavailable).
//     Returning nil signals "not resolvable" for the current request.
//
// Extensibility:
//   - Multiple resolvers can be registered. When types overlap, user-provided
//     resolvers override built-in ones at composition time.
//   - Prefer inexpensive lookups (e.g., values cached in context) to avoid
//     per-request overhead.
type HandlerParamResolver interface {
	Type() reflect.Type
	Resolve(ctx fiber.Ctx) (reflect.Value, error)
}

// Provider defines the interface for providing Api specifications.
// It provides a method to generate or retrieve Api specifications.
type Provider interface {
	// Provide returns an Api specification.
	Provide() Spec
}
