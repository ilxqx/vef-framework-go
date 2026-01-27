package shared

import (
	"errors"
	"fmt"

	"github.com/ilxqx/vef-framework-go/api"
)

// Sentinel errors for API engine operations.
var (
	ErrResourceNil              = errors.New("resource cannot be nil")
	ErrResourceNameEmpty        = errors.New("resource name cannot be empty")
	ErrOperationNotFound        = errors.New("operation not found")
	ErrOperationActionEmpty     = errors.New("operation action cannot be empty")
	ErrNoRouterForKind          = errors.New("no router can handle operation type")
	ErrNoRouterFound            = errors.New("no router found")
	ErrNoHandlerResolverFound   = errors.New("no handler resolver found")
	ErrHandlerRequired          = errors.New("handler is required for REST operations")
	ErrMethodNotFound           = errors.New("api action method not found")
	ErrMethodAmbiguous          = errors.New("api action method matches multiple methods")
	ErrHandlerInvalidReturnType = errors.New("handler method has invalid return type, must be 'error'")
	ErrHandlerTooManyReturns    = errors.New("handler method has too many return values")
	ErrHandlerMustBeFunc        = errors.New("provided handler must be a function")
	ErrHandlerNil               = errors.New("provided handler function cannot be nil")
)

type BaseError struct {
	Identifier *api.Identifier
	Err        error
}

func (e *BaseError) Error() string {
	if e.Identifier != nil {
		return fmt.Sprintf(
			"resource=%q action=%q version=%q - %v",
			e.Identifier.Resource,
			e.Identifier.Action,
			e.Identifier.Version,
			e.Err,
		)
	}

	return e.Err.Error()
}

// Unwrap returns the underlying error, allowing errors.As and errors.Is to work correctly.
func (e *BaseError) Unwrap() error {
	return e.Err
}

type DuplicateError struct {
	BaseError

	Existing *api.Operation
}

func (e *DuplicateError) Error() string {
	if e.Identifier != nil {
		return fmt.Sprintf(
			"duplicate api definition: resource=%q, action=%q, version=%q (attempting to override existing api)",
			e.Identifier.Resource,
			e.Identifier.Action,
			e.Identifier.Version,
		)
	}

	return "duplicate api definition"
}

type NotFoundError struct {
	BaseError

	Suggestion *api.Identifier
}

func (e *NotFoundError) Error() string {
	if e.Identifier == nil {
		return "api not found"
	}

	msg := fmt.Sprintf(
		"api not found: resource=%q, action=%q, version=%q",
		e.Identifier.Resource,
		e.Identifier.Action,
		e.Identifier.Version,
	)

	if e.Suggestion != nil {
		msg += fmt.Sprintf(
			" - did you mean: resource=%q, action=%q, version=%q ?",
			e.Suggestion.Resource,
			e.Suggestion.Action,
			e.Suggestion.Version,
		)
	}

	return msg
}
