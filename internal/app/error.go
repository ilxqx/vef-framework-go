package app

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/ilxqx/vef-framework-go/internal/i18n"
	"github.com/ilxqx/vef-framework-go/result"
	"github.com/samber/lo"
)

// handleError handles the error and returns the response
func handleError(ctx fiber.Ctx, err error) error {
	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) {
		r := &result.Result{
			Code: lo.If(fiberErr.Code == fiber.StatusNotFound, result.ErrCodeNotFound).
				ElseIf(fiberErr.Code == fiber.StatusUnauthorized, result.ErrCodeUnauthenticated).
				ElseIf(fiberErr.Code == fiber.StatusForbidden, result.ErrCodeAccessDenied).
				ElseIf(fiberErr.Code == fiber.StatusUnsupportedMediaType, result.ErrCodeUnsupportedMediaType).
				ElseIf(fiberErr.Code == fiber.StatusRequestTimeout, result.ErrCodeRequestTimeout).
				Else(result.ErrCodeDefault),
			Message: lo.If(fiberErr.Code == fiber.StatusNotFound, i18n.T(result.ErrMessageNotFound)).
				ElseIf(fiberErr.Code == fiber.StatusUnauthorized, i18n.T(result.ErrMessageUnauthenticated)).
				ElseIf(fiberErr.Code == fiber.StatusForbidden, i18n.T(result.ErrMessageAccessDenied)).
				ElseIf(fiberErr.Code == fiber.StatusUnsupportedMediaType, i18n.T(result.ErrMessageUnsupportedMediaType)).
				ElseIf(fiberErr.Code == fiber.StatusRequestTimeout, i18n.T(result.ErrMessageRequestTimeout)).
				Else(fiberErr.Error()),
		}

		return r.ResponseWithStatus(ctx, fiberErr.Code)
	}

	if resultErr, ok := result.AsErr(err); ok {
		return responseError(resultErr, ctx)
	}

	return responseError(result.ErrUnknown, ctx)
}

// responseError returns the response with the error
func responseError(e result.Error, ctx fiber.Ctx) error {
	r := result.Result{
		Code:    e.Code,
		Message: e.Message,
	}

	return r.ResponseWithStatus(ctx, e.Status)
}
