package api

import (
	"errors"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/utils/v2"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/contextx"
	"github.com/ilxqx/vef-framework-go/encoding"
	"github.com/ilxqx/vef-framework-go/event"
	"github.com/ilxqx/vef-framework-go/result"
	"github.com/ilxqx/vef-framework-go/webhelpers"
)

func buildAuditMiddleware(manager api.Manager, publisher event.Publisher) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		request := contextx.ApiRequest(ctx)
		definition := manager.Lookup(request.Identifier)

		if !definition.EnableAudit {
			return ctx.Next()
		}

		startTime := time.Now()
		handlerErr := ctx.Next()
		elapsed := int(time.Since(startTime).Milliseconds())

		auditEvent, err := buildAuditEvent(ctx, request, elapsed, handlerErr)
		if err != nil {
			contextx.Logger(ctx).Errorf("failed to build audit event: %v", err)

			return handlerErr
		}

		publisher.Publish(auditEvent)

		return handlerErr
	}
}

func buildAuditEvent(ctx fiber.Ctx, request *api.Request, elapsed int, handlerErr error) (*api.AuditEvent, error) {
	principal := contextx.Principal(ctx)

	var userId string
	if principal != nil {
		userId = principal.Id
	}

	requestId := contextx.RequestId(ctx)
	requestIP := webhelpers.GetIp(ctx)
	userAgent := utils.CopyString(ctx.Get(fiber.HeaderUserAgent))

	var (
		resultCode    int
		resultMessage string
		resultData    any
	)

	if handlerErr == nil {
		res, err := encoding.FromJson[result.Result](string(utils.CopyBytes(ctx.Response().Body())))
		if err != nil {
			return nil, fmt.Errorf("failed to decode response body for audit event: %w", err)
		}

		resultCode = res.Code
		resultMessage = res.Message
		resultData = res.Data
	} else {
		if err, ok := result.AsErr(handlerErr); ok {
			resultCode = err.Code
			resultMessage = err.Message
		} else {
			var err *fiber.Error
			if errors.As(handlerErr, &err) {
				resultCode = err.Code
				resultMessage = err.Message
			} else {
				resultCode = result.ErrCodeUnknown
				resultMessage = handlerErr.Error()
			}
		}
	}

	return api.NewAuditEvent(
		request.Resource,
		request.Action,
		request.Version,
		userId,
		userAgent,
		requestId,
		requestIP,
		request.Params,
		request.Meta,
		resultCode,
		resultMessage,
		resultData,
		elapsed,
	), nil
}
