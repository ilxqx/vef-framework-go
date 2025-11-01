package result

import (
	"github.com/gofiber/fiber/v3"

	"github.com/ilxqx/vef-framework-go/i18n"
)

// The Result is a struct that represents a result of an Api call.
type Result struct {
	// Code is the response code
	Code int `json:"code"`
	// Message is the response message
	Message string `json:"message"`
	// Data is the response data
	Data any `json:"data"`
}

// Response returns a JSON response for the result.
// Optionally accepts a custom HTTP status code; defaults to 200 (StatusOK) if not provided.
//
// Usage examples:
//
//	// Default status 200
//	return result.Ok(data).Response(ctx)
//
//	// Custom status code
//	return result.Ok(data).Response(ctx, fiber.StatusCreated)
func (r Result) Response(ctx fiber.Ctx, status ...int) error {
	statusCode := fiber.StatusOK
	if len(status) > 0 {
		statusCode = status[0]
	}

	return ctx.Status(statusCode).JSON(r)
}

// IsOk checks if the result is ok.
func (r Result) IsOk() bool {
	return r.Code == OkCode
}

// Ok creates a new Result with optional data and options.
//
// Usage examples:
//
//	// Simple success without data
//	return result.Ok()
//
//	// Success with data
//	return result.Ok(userData)
//
//	// Success with custom message
//	return result.Ok(WithMessage("operation completed"))
//
//	// Success with data and custom message
//	return result.Ok(userData, WithMessage("user created successfully"))
//
// Parameter order rules:
//   - If provided, data must come before any option functions
//   - Only one data argument is allowed
//   - Multiple option functions can be provided
func Ok(dataOrOptions ...any) Result {
	var (
		data             any
		options          []okOption
		firstOptionIndex = -1
		dataCount        = 0
	)

	for i, v := range dataOrOptions {
		switch v := v.(type) {
		case okOption:
			if firstOptionIndex == -1 {
				firstOptionIndex = i
			}

			options = append(options, v)

		default:
			if firstOptionIndex != -1 && i > firstOptionIndex {
				panic("result.Ok: data must come before option functions. Correct usage: Ok(data, WithMessage(...))")
			}

			dataCount++
			if dataCount > 1 {
				panic("result.Ok: only one data argument is allowed. Correct usage: Ok() or Ok(data) or Ok(data, options...)")
			}

			data = v
		}
	}

	result := Result{
		Code:    OkCode,
		Message: i18n.T(OkMessage),
		Data:    data,
	}

	for _, opt := range options {
		opt(&result)
	}

	return result
}
