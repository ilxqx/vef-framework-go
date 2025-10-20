package api

import (
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

// Identifier uniquely identifies an Api endpoint.
// It consists of version, resource name, and action name.
type Identifier struct {
	// The version of the Api endpoint
	Version string `json:"version" form:"version" validate:"required,alphanum" label_i18n:"api_request_version"`
	// The resource name of the Api endpoint
	Resource string `json:"resource" form:"resource" validate:"required,alphanum_us_slash" label_i18n:"api_request_resource"`
	// The action name of the Api endpoint
	Action string `json:"action" form:"action" validate:"required,alphanum_us" label_i18n:"api_request_action"`
}

// String returns a string representation of the identifier.
func (id Identifier) String() string {
	return id.Resource + constants.Colon + id.Action + constants.Colon + id.Version
}

// IsValid checks if the identifier has all required fields.
func (id Identifier) IsValid() bool {
	return id.Resource != constants.Empty &&
		id.Action != constants.Empty &&
		id.Version != constants.Empty
}

// Equals checks if two identifiers are equal.
func (id Identifier) Equals(other Identifier) bool {
	return id.Resource == other.Resource &&
		id.Action == other.Action &&
		id.Version == other.Version
}

type Params map[string]any

type Meta map[string]any

// Request represents an Api request with identifier, params, and metadata.
type Request struct {
	Identifier

	// The params of the request
	Params Params `json:"params"`
	// The meta of the request
	Meta Meta `json:"meta"`
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

// Spec defines the specification for an Api endpoint.
type Spec struct {
	// Action is the action name for the Api endpoint
	Action string
	// Version is the version of the Api endpoint
	Version string
	// EnableAudit indicates whether to enable audit logging for this endpoint
	EnableAudit bool
	// Timeout is the request timeout duration
	Timeout time.Duration
	// Public indicates whether this endpoint is publicly accessible
	Public bool
	// PermToken is the permission token required for access
	PermToken string
	// Limit represents the rate limit for an Api endpoint
	Limit RateLimit
	// Handler is optional. If not provided, the system will automatically search for a method
	// in the struct using the Action name converted to PascalCase format.
	// The search supports both direct methods and methods from embedded anonymous structs.
	// For example, if Action is "create_user", the system will look for "CreateUser" method.
	// The handler function should be compatible with the Api framework's handler signature.
	Handler any
}

// RateLimit represents the rate limit for an Api endpoint.
type RateLimit struct {
	// Max is the rate limit per time window (0 means no limit)
	Max int
	// Expiration is the rate limit expiration time
	Expiration time.Duration
}

// Definition represents a complete Api definition with identifier, configuration, and handler.
type Definition struct {
	Identifier

	// EnableAudit indicates whether to enable audit logging for this endpoint
	EnableAudit bool
	// Timeout is the request timeout duration
	Timeout time.Duration
	// Public indicates whether this endpoint is publicly accessible
	Public bool
	// PermToken is the permission token required for access
	PermToken string
	// Limit represents the rate limit for an Api endpoint
	Limit RateLimit
	// Handler is the actual handler function for this endpoint
	Handler fiber.Handler
}

// IsPublic returns true if the endpoint is publicly accessible.
func (d *Definition) IsPublic() bool {
	return d.Public
}

// RequiresPermission returns true if the endpoint requires a permission token.
func (d *Definition) RequiresPermission() bool {
	return d.PermToken != constants.Empty
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

// In is a struct that can be used to inject parameters into an Api handler.
type In struct {
	_ paramsSentinel `bun:"-"`
}
