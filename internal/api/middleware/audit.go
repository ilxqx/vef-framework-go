package middleware

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
	"github.com/ilxqx/vef-framework-go/i18n"
	"github.com/ilxqx/vef-framework-go/internal/api/common"
	"github.com/ilxqx/vef-framework-go/internal/app"
	"github.com/ilxqx/vef-framework-go/result"
	"github.com/ilxqx/vef-framework-go/security"
	"github.com/ilxqx/vef-framework-go/webhelpers"
)

// Audit handles audit logging.
type Audit struct {
	publisher event.Publisher
}

// NewAudit creates a new audit middleware.
func NewAudit(publisher event.Publisher) api.Middleware {
	return &Audit{
		publisher: publisher,
	}
}

// Name returns the middleware name.
func (m *Audit) Name() string {
	return "audit"
}

// Order returns the middleware order.
func (m *Audit) Order() int {
	return -60
}

// Process handles the audit logging.
func (m *Audit) Process(ctx fiber.Ctx) error {
	op := common.Operation(ctx)
	if op == nil {
		contextx.Logger(ctx).Warnf("Audit skipped: %v", ErrOperationNotFound)

		return ctx.Next()
	}

	return m.audit(ctx, op)
}

func (m *Audit) audit(ctx fiber.Ctx, op *api.Operation) error {
	if !op.EnableAudit || m.publisher == nil {
		return ctx.Next()
	}

	var (
		start      = time.Now()
		handlerErr = ctx.Next()
		elapsed    = time.Since(start).Milliseconds()
	)

	evt, buildErr := buildAuditEvent(ctx, elapsed, handlerErr)
	if buildErr != nil {
		contextx.Logger(ctx).Errorf("%v: %v", ErrAuditEventBuildFailed, buildErr)

		return handlerErr
	}

	m.publisher.Publish(evt)

	return handlerErr
}

func buildAuditEvent(ctx fiber.Ctx, elapsed int64, err error) (*api.AuditEvent, error) {
	req := common.Request(ctx)
	if req == nil {
		return nil, ErrRequestNotFound
	}

	principal := contextx.Principal(ctx)
	if principal == nil {
		principal = security.PrincipalAnonymous
	}

	var (
		userID     = principal.ID
		requestID  = contextx.RequestID(ctx)
		requestIP  = webhelpers.GetIP(ctx)
		userAgent  = utils.CopyString(ctx.Get(fiber.HeaderUserAgent))
		resultCode int
		resultMsg  string
		resultData any
	)

	if err != nil {
		resultCode, resultMsg = extractErrorInfo(err)
	} else {
		res, decodeErr := encoding.FromJSON[result.Result](string(utils.CopyBytes(ctx.Response().Body())))
		if decodeErr != nil {
			return nil, fmt.Errorf("%w: %w", ErrResponseDecodeFailed, decodeErr)
		}

		resultCode = res.Code
		resultMsg = res.Message
		resultData = res.Data
	}

	return api.NewAuditEvent(
		req.Resource, req.Action, req.Version,
		userID, userAgent, requestID, requestIP,
		req.Params, req.Meta,
		resultCode, resultMsg, resultData, elapsed,
	), nil
}

func extractErrorInfo(err error) (code int, message string) {
	if resultErr, ok := result.AsErr(err); ok {
		return resultErr.Code, resultErr.Message
	}

	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) {
		mappedCode, messageKey, _ := app.MapFiberError(fiberErr.Code)

		return mappedCode, i18n.T(messageKey)
	}

	return result.ErrCodeUnknown, err.Error()
}
