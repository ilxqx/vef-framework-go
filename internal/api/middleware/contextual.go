package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v3"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/contextx"
	"github.com/ilxqx/vef-framework-go/internal/api/shared"
	"github.com/ilxqx/vef-framework-go/orm"
	"github.com/ilxqx/vef-framework-go/security"
)

// Contextual injects DB and Logger into the request context.
// It sets up a contextual database with the operator ID and a scoped logger
// with request identification information.
type Contextual struct {
	db orm.DB
}

// NewContextual creates a new context middleware.
func NewContextual(db orm.DB) api.Middleware {
	return &Contextual{
		db: db,
	}
}

// Name returns the middleware name.
func (*Contextual) Name() string {
	return "contextual"
}

// Order returns the middleware order.
// Runs after authentication (-100) but before authorization (-90).
func (*Contextual) Order() int {
	return -90
}

// Process sets up the request context with DB and Logger.
func (m *Contextual) Process(ctx fiber.Ctx) error {
	principal := contextx.Principal(ctx)
	if principal == nil {
		principal = security.PrincipalAnonymous
	}

	db := m.db.WithNamedArg(constants.PlaceholderKeyOperator, principal.ID)
	contextx.SetDB(ctx, db)
	ctx.SetContext(contextx.SetDB(ctx.Context(), db))

	req := shared.Request(ctx)

	lgr := contextx.Logger(ctx)
	if req != nil && lgr != nil {
		scopedLogger := lgr.
			Named(buildRequestLoggerName(req.Resource, req.Action, req.Version)).
			Named(buildPrincipalLoggerName(principal))
		contextx.SetLogger(ctx, scopedLogger)
		ctx.SetContext(contextx.SetLogger(ctx.Context(), scopedLogger))
	}

	return ctx.Next()
}

// buildRequestLoggerName creates a logger name from request info.
// Format: resource:action@version.
func buildRequestLoggerName(resource, action, version string) string {
	var sb strings.Builder

	sb.WriteString(resource)
	sb.WriteByte(constants.ByteColon)
	sb.WriteString(action)
	sb.WriteByte(constants.ByteAt)
	sb.WriteString(version)

	return sb.String()
}

// buildPrincipalLoggerName creates a logger name from principal info.
// Format: type:id@name.
func buildPrincipalLoggerName(principal *security.Principal) string {
	var sb strings.Builder

	sb.WriteString(string(principal.Type))
	sb.WriteByte(constants.ByteColon)
	sb.WriteString(principal.ID)
	sb.WriteByte(constants.ByteAt)
	sb.WriteString(principal.Name)

	return sb.String()
}
