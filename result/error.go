package result

import (
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v3"
)

// Error is a custom application error type that implements the error interface.
type Error struct {
	Code    int    // Code is the error code
	Message string // Message is the error message
	Status  int    // Status is the HTTP status code
}

// Error implements the error interface.
func (e Error) Error() string {
	return e.Message
}

// Err creates a new Error.
func Err(message string) Error {
	return Error{
		Code:    ErrCodeDefault,
		Message: message,
		Status:  fiber.StatusOK,
	}
}

// ErrWithCode creates a new Error with a code.
func ErrWithCode(code int, message string) Error {
	return Error{
		Code:    code,
		Message: message,
		Status:  fiber.StatusOK,
	}
}

// ErrWithCodeAndStatus creates a new Error with a code and status.
func ErrWithCodeAndStatus(code int, status int, message string) Error {
	return Error{
		Code:    code,
		Message: message,
		Status:  status,
	}
}

// Errf creates a new Error with a formatted message.
func Errf(messageFormat string, args ...any) Error {
	return Error{
		Code:    ErrCodeDefault,
		Message: fmt.Sprintf(messageFormat, args...),
		Status:  fiber.StatusOK,
	}
}

// ErrWithCodef creates a new Error with a formatted message and code.
func ErrWithCodef(code int, messageFormat string, args ...any) Error {
	return Error{
		Code:    code,
		Message: fmt.Sprintf(messageFormat, args...),
		Status:  fiber.StatusOK,
	}
}

// ErrWithCodeAndStatusf creates a new Error with a formatted message, code and status.
func ErrWithCodeAndStatusf(code int, status int, messageFormat string, args ...any) *Error {
	return &Error{
		Code:    code,
		Message: fmt.Sprintf(messageFormat, args...),
		Status:  status,
	}
}

// AsErr checks if the error is of Error type and returns the error if it is.
func AsErr(err error) (Error, bool) {
	var target Error
	if errors.As(err, &target) {
		return target, true
	}

	return Error{}, false
}

// IsErr checks if the error is of Error type.
func IsErr(err error, target *Error) bool {
	return errors.Is(err, target)
}

// IsErrRecordNotFound checks if the error is a record not found error.
func IsErrRecordNotFound(err error) bool {
	return errors.Is(err, ErrRecordNotFound)
}
