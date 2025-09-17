package middleware

import (
	"github.com/gofiber/fiber/v3/middleware/requestid"
	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/internal/app"
	"github.com/ilxqx/vef-framework-go/utils"
)

// newRequestIdMiddleware returns a middleware that generates a request ID.
func newRequestIdMiddleware() app.Middleware {
	handler := requestid.New(requestid.Config{
		Generator: utils.GenerateId,
		Header:    constants.HeaderXRequestId,
	})

	return &SimpleMiddleware{
		handler: handler,
		name:    "request_id",
		order:   -650,
	}
}
