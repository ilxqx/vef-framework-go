package result

import (
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v3"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/i18n"
)

// Error represents a business-level error specifically designed for API responses.
//
// Design Philosophy:
//
// This type separates transport-level concerns from business logic concerns:
//   - HTTP Status: Indicates whether the request was successfully processed at the transport layer.
//     Typically remains 200 (fiber.StatusOK) to indicate successful communication.
//   - Code: Represents business-level error codes that indicate the actual result of the operation.
//   - Message: Provides a user-friendly, optionally internationalized error message.
//
// This design is common in many large-scale API systems (e.g., WeChat, Alipay) and offers several advantages:
//   - Unified client handling: All responses use HTTP 200, with business results determined by the Code field
//   - Avoids middleware interference: 4xx/5xx status codes won't trigger special handling by proxies or gateways
//   - Simplified client logic: No need to handle both HTTP errors and business errors separately
//
// Note: Error is NOT intended for general-purpose error handling or wrapping internal errors.
// It is specifically designed for constructing API response payloads and works in conjunction
// with the result.Result type to provide a consistent API response format.
//
// Example usage:
//
//	// Simple error with default message
//	return result.Err()
//
//	// Error with custom message
//	return result.Err("user not found")
//
//	// Error with custom message and options
//	return result.Err("unauthorized access", WithCode(401), WithStatus(fiber.StatusUnauthorized))
//
//	// Formatted error with arguments
//	return result.Errf("user %s not found", username, WithCode(404))
type Error struct {
	// Code is the business error code
	Code int
	// Message is the user-friendly error message
	Message string
	// Status is the HTTP status code (defaults to 200 for business errors)
	Status int
}

// Error implements the error interface.
func (e Error) Error() string {
	return e.Message
}

// Err creates a new Error.
func Err(messageOrOptions ...any) Error {
	var (
		message string
		options []errOption
	)

	for i, v := range messageOrOptions {
		switch v := v.(type) {
		case string:
			if i != 0 {
				panic("result.Err: message string must be the first argument if provided. Correct usage: Err(\"message\", options...) or Err(options...)")
			}

			message = v

		case errOption:
			options = append(options, v)
		default:
			panic(fmt.Sprintf("result.Err: invalid argument type %T at position %d. Only string message and errOption functions are allowed", v, i))
		}
	}

	if message == constants.Empty {
		message = i18n.T(ErrMessage)
	}

	err := Error{
		Code:    ErrCodeDefault,
		Message: message,
		Status:  fiber.StatusOK,
	}

	for _, opt := range options {
		opt(&err)
	}

	return err
}

// Errf creates a new Error with a formatted message.
func Errf(messageFormat string, args ...any) Error {
	var (
		messageArgs      []any
		options          []errOption
		firstOptionIndex = -1
	)

	if len(args) == 0 {
		panic("result.Errf: at least one format argument is required. Use Err() for static messages without formatting")
	}

	for i, v := range args {
		switch v := v.(type) {
		case errOption:
			if firstOptionIndex == -1 {
				firstOptionIndex = i
			}

			options = append(options, v)

		default:
			if firstOptionIndex != -1 && i > firstOptionIndex {
				panic("result.Errf: all message format arguments must come before option functions. Correct usage: Errf(\"format %s %d\", arg1, arg2, WithCode(...))")
			}

			messageArgs = append(messageArgs, v)
		}
	}

	err := Error{
		Code:    ErrCodeDefault,
		Message: fmt.Sprintf(messageFormat, messageArgs...),
		Status:  fiber.StatusOK,
	}

	for _, opt := range options {
		opt(&err)
	}

	return err
}

// AsErr checks if the error is of Error type and returns the error if it is.
func AsErr(err error) (Error, bool) {
	var target Error
	if errors.As(err, &target) {
		return target, true
	}

	return Error{}, false
}

// IsRecordNotFound checks if the error is a record not found error.
func IsRecordNotFound(err error) bool {
	return errors.Is(err, ErrRecordNotFound)
}
