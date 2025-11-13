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
type HandlerParamResolver interface {
	Type() reflect.Type
	Resolve(ctx fiber.Ctx) (reflect.Value, error)
}

// FactoryParamResolver resolves parameters for handler factory functions.
// Unlike HandlerParamResolver, it executes during resource registration (application startup)
// and does not depend on a specific HTTP request context.
type FactoryParamResolver interface {
	// Type returns the parameter type this resolver handles
	Type() reflect.Type
	// Resolve returns an instance of the type (as reflect.Value)
	// Called during resource registration, resolver should return instances already injected via DI
	Resolve() reflect.Value
}

// Provider defines the interface for providing Api specifications.
// It provides a method to generate or retrieve Api specifications.
type Provider interface {
	// Provide returns an Api specification.
	Provide() Spec
}
