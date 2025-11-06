package result

import "fmt"

type errOption func(*Error)

// WithCode sets the error code.
func WithCode(code int) errOption {
	return func(e *Error) {
		e.Code = code
	}
}

// WithStatus sets the HTTP status code.
func WithStatus(status int) errOption {
	return func(e *Error) {
		e.Status = status
	}
}

type okOption func(*Result)

// WithMessage sets a custom message for the result.
func WithMessage(message string) okOption {
	return func(r *Result) {
		r.Message = message
	}
}

// WithMessagef sets a formatted message for the result.
func WithMessagef(messageFormat string, args ...any) okOption {
	return func(r *Result) {
		r.Message = fmt.Sprintf(messageFormat, args...)
	}
}
