package api

import (
	"reflect"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/ilxqx/vef-framework-go/constants"
)

const (
	VersionV1 = "v1" // v1 is the default version
	VersionV2 = "v2"
	VersionV3 = "v3"
	VersionV4 = "v4"
	VersionV5 = "v5"
	VersionV6 = "v6"
	VersionV7 = "v7"
	VersionV8 = "v8"
	VersionV9 = "v9"
)

// Manager defines the interface for managing API definitions.
// It provides methods to register, remove, and lookup API definitions by their identifiers.
type Manager interface {
	// Register adds a new API definition to the manager.
	Register(api *Definition)
	// Remove removes an API definition by its identifier.
	Remove(id Identifier)
	// Lookup retrieves an API definition by its identifier.
	// Returns nil if the definition is not found.
	Lookup(id Identifier) *Definition
	// List returns all registered API definitions.
	List() []*Definition
}

// Resource represents an API resource that contains multiple API configurations.
// It defines the version, name, and list of API configurations for a resource.
type Resource interface {
	// Version returns the version of the resource.
	Version() string
	// Name returns the name of the resource.
	Name() string
	// APIs returns the list of API configurations for this resource.
	APIs() []Config
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

// Identifier uniquely identifies an API endpoint.
// It consists of version, resource name, and action name.
type Identifier struct {
	Version  string `json:"version" validate:"required,alphanum"` // The version of the API endpoint
	Resource string `json:"resource" validate:"required,ascii"`   // The resource name of the API endpoint
	Action   string `json:"action" validate:"required,alphanum"`  // The action name of the API endpoint
}

// String returns a string representation of the identifier.
func (id Identifier) String() string {
	return id.Resource + constants.Colon + id.Action + constants.Colon + id.Version
}

// IsValid checks if the identifier has all required fields.
func (id Identifier) IsValid() bool {
	return id.Resource != constants.Empty && id.Action != constants.Empty && id.Version != constants.Empty
}

// Equals checks if two identifiers are equal.
func (id Identifier) Equals(other Identifier) bool {
	return id.Resource == other.Resource && id.Action == other.Action && id.Version == other.Version
}

// Request represents an API request with identifier, params, and metadata.
type Request struct {
	Identifier
	Params map[string]any `json:"params"` // The params of the request
	Meta   map[string]any `json:"meta"`   // The meta of the request
}

// GetParam retrieves a value from the request param by key.
func (r *Request) GetParam(key string) (any, bool) {
	if r.Params == nil {
		return nil, false
	}

	value, exists := r.Params[key]
	return value, exists
}

// GetMeta retrieves a value from the request metadata by key.
func (r *Request) GetMeta(key string) (any, bool) {
	if r.Meta == nil {
		return nil, false
	}

	value, exists := r.Meta[key]
	return value, exists
}

// Config defines the configuration for an API endpoint.
type Config struct {
	Action          string        // The action name for the API endpoint
	Version         string        // The version of the API endpoint
	EnableAudit     bool          // Whether to enable audit logging for this endpoint
	Timeout         time.Duration // Request timeout duration
	Public          bool          // Whether this endpoint is publicly accessible
	PermissionToken string        // Permission token required for access
	Limit           RateLimit     // Limit represents the rate limit for an API endpoint.
}

// RateLimit represents the rate limit for an API endpoint.
type RateLimit struct {
	Max        int           // Rate limit per time window (0 means no limit)
	Expiration time.Duration // Rate limit expiration time
}

// Definition represents a complete API definition with identifier, configuration, and handler.
type Definition struct {
	Identifier
	EnableAudit     bool          // Whether to enable audit logging for this endpoint
	Timeout         time.Duration // Request timeout duration
	Public          bool          // Whether this endpoint is publicly accessible
	PermissionToken string        // Permission token required for access
	Limit           RateLimit     // Limit represents the rate limit for an API endpoint.
	Handler         fiber.Handler // The actual handler function for this endpoint
}

// IsPublic returns true if the endpoint is publicly accessible.
func (d *Definition) IsPublic() bool {
	return d.Public
}

// RequiresPermission returns true if the endpoint requires a permission token.
func (d *Definition) RequiresPermission() bool {
	return d.PermissionToken != constants.Empty
}

// HasRateLimit returns true if the endpoint has a rate limit configured.
func (d *Definition) HasRateLimit() bool {
	return d.Limit.Max > 0
}

// GetTimeout returns the configured timeout or a default value.
func (d *Definition) GetTimeout() time.Duration {
	if d.Timeout > 0 {
		return d.Timeout
	}

	// default timeout is 30 seconds
	return 30 * time.Second
}

type paramsSentinel struct{}

// In is a struct that can be used to inject parameters into an API handler.
type In struct {
	_ paramsSentinel
}
