package api

import (
	"fmt"
	"reflect"
	"time"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/mapx"
	"github.com/ilxqx/vef-framework-go/reflectx"
)

// Identifier uniquely identifies an API operation.
type Identifier struct {
	Resource string `json:"resource" form:"resource" validate:"required,alphanum_us_slash" label_i18n:"api_request_resource"`
	Action   string `json:"action" form:"action" validate:"required" label_i18n:"api_request_action"`
	Version  string `json:"version" form:"version" validate:"required,alphanum" label_i18n:"api_request_version"`
}

// String returns a string representation of the identifier.
func (id Identifier) String() string {
	return id.Resource + constants.Colon + id.Action + constants.Colon + id.Version
}

// Params holds API request parameters.
type Params map[string]any

// Decode decodes params into a struct.
func (p Params) Decode(out any) error {
	return decodeMap(p, out, ErrInvalidParamsType)
}

// Meta holds API request metadata.
type Meta map[string]any

// Decode decodes meta into a struct.
func (m Meta) Decode(out any) error {
	return decodeMap(m, out, ErrInvalidMetaType)
}

// decodeMap decodes a map into a struct with type validation.
func decodeMap(data map[string]any, out any, typeErr error) error {
	if !reflectx.IsPointerToStruct(reflect.TypeOf(out)) {
		return fmt.Errorf("%w, got %T", typeErr, out)
	}

	decoder, err := mapx.NewDecoder(out)
	if err != nil {
		return err
	}

	return decoder.Decode(data)
}

// Request represents a unified API request.
type Request struct {
	Identifier

	Params Params `json:"params"`
	Meta   Meta   `json:"meta"`
}

// GetParam retrieves a value from the request params by key.
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

// OperationSpec defines the specification for an Api endpoint.
type OperationSpec struct {
	// Action is the action name for the Api endpoint
	Action string
	// EnableAudit indicates whether to enable audit logging for this endpoint
	EnableAudit bool
	// Timeout is the request timeout duration
	Timeout time.Duration
	// Public indicates whether this endpoint is publicly accessible
	Public bool
	// PermToken is the permission token required for access
	PermToken string
	// RateLimit represents the rate limit for an Api endpoint
	RateLimit *RateLimitConfig
	// Handler is the business logic handler.
	Handler any
}

// Operation is the runtime operation definition.
// Created by Engine from Resource + OperationSpec.
type Operation struct {
	// Identifier contains the full operation identity.
	Identifier

	// EnableAudit indicates whether audit logging is enabled.
	EnableAudit bool
	// Timeout is the final timeout duration.
	Timeout time.Duration
	// Auth indicates the authentication configuration.
	Auth *AuthConfig
	// RateLimit is the final rate limit configuration.
	RateLimit *RateLimitConfig
	// Handler is the resolved handler (before adaptation).
	Handler any
	// Dynamic indicates whether this operation is registered dynamically.
	Dynamic bool
	// Meta holds additional operation-specific data.
	// For REST: may contain parsed method, path pattern, etc.
	Meta map[string]any
}

// HasRateLimit returns true if rate limiting is configured.
func (o *Operation) HasRateLimit() bool {
	return o.RateLimit != nil && o.RateLimit.Max > 0
}

// RequiresAuth returns true if authentication is required.
func (o *Operation) RequiresAuth() bool {
	return o.Auth.Strategy != AuthStrategyNone
}

// RateLimitConfig defines rate limiting configuration.
type RateLimitConfig struct {
	// Max is the maximum number of requests allowed.
	Max int
	// Period is the time window for rate limiting.
	Period time.Duration
	// Key is a custom rate limit key template.
	// Empty means using the default key.
	Key string
}
