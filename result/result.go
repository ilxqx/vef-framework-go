package result

import (
	"github.com/gofiber/fiber/v3"
	"github.com/ilxqx/vef-framework-go/i18n"
)

// The Result is a struct that represents a result of an API call.
type Result struct {
	Code    int    `json:"code"`    // Code is the response code
	Message string `json:"message"` // Message is the response message
	Data    any    `json:"data"`    // Data is the response data
}

// Response returns a JSON response for the result.
func (r Result) Response(ctx fiber.Ctx) error {
	return ctx.Status(fiber.StatusOK).JSON(r)
}

// ResponseWithStatus returns a JSON response for the result with the given status.
func (r Result) ResponseWithStatus(ctx fiber.Ctx, status int) error {
	return ctx.Status(status).JSON(r)
}

// IsOk checks if the result is ok.
func (r Result) IsOk(data any) bool {
	return r.Code == OkCode
}

// Ok creates a new Result with the given data.
func Ok(data ...any) Result {
	var dataToUse any
	if len(data) > 0 {
		dataToUse = data[0]
	}

	return Result{
		Code:    OkCode,
		Message: i18n.T(OkMessage),
		Data:    dataToUse,
	}
}

// OkWithMessage creates a new Result with the given message and data.
func OkWithMessage(message string, data ...any) Result {
	var dataToUse any
	if len(data) > 0 {
		dataToUse = data[0]
	}

	return Result{
		Code:    OkCode,
		Message: message,
		Data:    dataToUse,
	}
}
