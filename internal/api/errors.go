package api

import (
	"errors"
	"fmt"

	"github.com/ilxqx/vef-framework-go/api"
)

var (
	// ErrResolveParamType indicates failing to resolve handler parameter type.
	ErrResolveParamType = errors.New("failed to resolve api handler parameter type")
	// ErrProvidedHandlerNil indicates provided handler is nil.
	ErrProvidedHandlerNil = errors.New("provided handler cannot be nil")
	// ErrProvidedHandlerMustFunc indicates provided handler must be a function.
	ErrProvidedHandlerMustFunc = errors.New("provided handler must be a function")
	// ErrProvidedHandlerFuncNil indicates provided handler function is nil.
	ErrProvidedHandlerFuncNil = errors.New("provided handler function cannot be nil")
	// ErrHandlerFactoryRequireDB indicates handler factory requires db.
	ErrHandlerFactoryRequireDB = errors.New("handler factory function requires database connection but none provided")
	// ErrHandlerFactoryMethodRequireDB indicates handler factory method requires db.
	ErrHandlerFactoryMethodRequireDB = errors.New("handler factory method requires database connection but none provided")
	// ErrApiMethodNotFound indicates api action method not found.
	ErrApiMethodNotFound = errors.New("api action method not found in resource")
	// ErrHandlerMethodInvalidReturn indicates handler method invalid return type.
	ErrHandlerMethodInvalidReturn = errors.New("handler method has invalid return type, must be 'error'")
	// ErrHandlerMethodTooManyReturns indicates handler method has too many returns.
	ErrHandlerMethodTooManyReturns = errors.New("handler method has too many return values, must have at most 1 (error) or none")
	// ErrHandlerFactoryInvalidReturn indicates handler factory invalid return count.
	ErrHandlerFactoryInvalidReturn = errors.New("handler factory method should return 1 or 2 values")
)

// DuplicateApiError represents an error when attempting to register a duplicate Api definition.
// It contains information about both the existing and new Api definitions.
type DuplicateApiError struct {
	Identifier api.Identifier
	Existing   *api.Definition
	New        *api.Definition
}

// Error returns a formatted error message with details about the duplicate Api.
func (e *DuplicateApiError) Error() string {
	return fmt.Sprintf(
		"duplicate api definition: resource=%q, action=%q, version=%q (attempting to override existing api)",
		e.Identifier.Resource,
		e.Identifier.Action,
		e.Identifier.Version,
	)
}
