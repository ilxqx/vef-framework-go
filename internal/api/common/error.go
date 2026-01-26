package common

import (
	"fmt"

	"github.com/ilxqx/vef-framework-go/api"
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
