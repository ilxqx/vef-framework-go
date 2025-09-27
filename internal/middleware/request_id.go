package middleware

import (
	"github.com/gofiber/fiber/v3/middleware/requestid"
	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/id"
	"github.com/ilxqx/vef-framework-go/internal/app"
)

// NewRequestIdMiddleware returns a middleware that generates a request ID.
func NewRequestIdMiddleware() app.Middleware {
	handler := requestid.New(requestid.Config{
		Generator: id.GenerateUuid,
		Header:    constants.HeaderXRequestId,
	})

	return &SimpleMiddleware{
		handler: handler,
		name:    "request_id",
		order:   -650,
	}
}
