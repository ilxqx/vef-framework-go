package contextx

import (
	"context"

	"github.com/gofiber/fiber/v3"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/log"
	"github.com/ilxqx/vef-framework-go/mold"
	"github.com/ilxqx/vef-framework-go/orm"
	"github.com/ilxqx/vef-framework-go/security"
)

type contextKey int

const (
	KeyRequest contextKey = iota
	KeyRequestId
	KeyPrincipal
	KeyLogger
	KeyDb
	KeyTransformer
)

// APIRequest returns the api.APIRequest from fiber context.
func APIRequest(ctx context.Context) *api.Request {
	req, _ := ctx.Value(KeyRequest).(*api.Request)

	return req
}

// SetAPIRequest stores the api.Request into fiber context.
func SetAPIRequest(ctx context.Context, request *api.Request) context.Context {
	switch c := ctx.(type) {
	case fiber.Ctx:
		c.Locals(KeyRequest, request)

		return c
	default:
		return context.WithValue(ctx, KeyRequest, request)
	}
}

// RequestId returns the request id from fiber context.
func RequestId(ctx context.Context) string {
	return ctx.Value(KeyRequestId).(string)
}

// SetRequestId stores the request id into fiber context.
func SetRequestId(ctx context.Context, requestId string) context.Context {
	switch c := ctx.(type) {
	case fiber.Ctx:
		c.Locals(KeyRequestId, requestId)

		return c
	default:
		return context.WithValue(ctx, KeyRequestId, requestId)
	}
}

// Principal returns the security.Principal from fiber context.
func Principal(ctx context.Context) *security.Principal {
	principal, _ := ctx.Value(KeyPrincipal).(*security.Principal)

	return principal
}

// SetPrincipal stores the security.Principal into fiber context.
func SetPrincipal(ctx context.Context, principal *security.Principal) context.Context {
	switch c := ctx.(type) {
	case fiber.Ctx:
		c.Locals(KeyPrincipal, principal)

		return c
	default:
		return context.WithValue(ctx, KeyPrincipal, principal)
	}
}

// Logger returns the log.Logger from fiber context.
func Logger(ctx context.Context) log.Logger {
	logger, _ := ctx.Value(KeyLogger).(log.Logger)

	return logger
}

// SetLogger stores the log.Logger into fiber context.
func SetLogger(ctx context.Context, logger log.Logger) context.Context {
	switch c := ctx.(type) {
	case fiber.Ctx:
		c.Locals(KeyLogger, logger)

		return c
	default:
		return context.WithValue(ctx, KeyLogger, logger)
	}
}

// Db returns the orm.Db from fiber context.
func Db(ctx context.Context) orm.Db {
	db, _ := ctx.Value(KeyDb).(orm.Db)

	return db
}

// SetDb stores the orm.Db into fiber context.
func SetDb(ctx context.Context, db orm.Db) context.Context {
	switch c := ctx.(type) {
	case fiber.Ctx:
		c.Locals(KeyDb, db)

		return c
	default:
		return context.WithValue(ctx, KeyDb, db)
	}
}

// Transformer returns the mold.Transformer from fiber context.
func Transformer(ctx context.Context) mold.Transformer {
	transformer, _ := ctx.Value(KeyTransformer).(mold.Transformer)

	return transformer
}

// SetTransformer stores the mold.Transformer into fiber context.
func SetTransformer(ctx context.Context, transformer mold.Transformer) context.Context {
	switch c := ctx.(type) {
	case fiber.Ctx:
		c.Locals(KeyTransformer, transformer)

		return c
	default:
		return context.WithValue(ctx, KeyTransformer, transformer)
	}
}
