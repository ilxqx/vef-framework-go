package contextx

import (
	"context"

	"github.com/gofiber/fiber/v3"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/log"
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
	KeyDataPermApplier
)

func ApiRequest(ctx context.Context) *api.Request {
	req, _ := ctx.Value(KeyRequest).(*api.Request)

	return req
}

func SetApiRequest(ctx context.Context, request *api.Request) context.Context {
	switch c := ctx.(type) {
	case fiber.Ctx:
		c.Locals(KeyRequest, request)

		return c
	default:
		return context.WithValue(ctx, KeyRequest, request)
	}
}

func RequestId(ctx context.Context) string {
	id, _ := ctx.Value(KeyRequestId).(string)

	return id
}

func SetRequestId(ctx context.Context, requestId string) context.Context {
	switch c := ctx.(type) {
	case fiber.Ctx:
		c.Locals(KeyRequestId, requestId)

		return c
	default:
		return context.WithValue(ctx, KeyRequestId, requestId)
	}
}

func Principal(ctx context.Context) *security.Principal {
	principal, _ := ctx.Value(KeyPrincipal).(*security.Principal)

	return principal
}

func SetPrincipal(ctx context.Context, principal *security.Principal) context.Context {
	switch c := ctx.(type) {
	case fiber.Ctx:
		c.Locals(KeyPrincipal, principal)

		return c
	default:
		return context.WithValue(ctx, KeyPrincipal, principal)
	}
}

func Logger(ctx context.Context) log.Logger {
	logger, _ := ctx.Value(KeyLogger).(log.Logger)

	return logger
}

func SetLogger(ctx context.Context, logger log.Logger) context.Context {
	switch c := ctx.(type) {
	case fiber.Ctx:
		c.Locals(KeyLogger, logger)

		return c
	default:
		return context.WithValue(ctx, KeyLogger, logger)
	}
}

func Db(ctx context.Context) orm.Db {
	db, _ := ctx.Value(KeyDb).(orm.Db)

	return db
}

func SetDb(ctx context.Context, db orm.Db) context.Context {
	switch c := ctx.(type) {
	case fiber.Ctx:
		c.Locals(KeyDb, db)

		return c
	default:
		return context.WithValue(ctx, KeyDb, db)
	}
}

func DataPermApplier(ctx context.Context) security.DataPermissionApplier {
	applier, _ := ctx.Value(KeyDataPermApplier).(security.DataPermissionApplier)

	return applier
}

func SetDataPermApplier(ctx context.Context, applier security.DataPermissionApplier) context.Context {
	switch c := ctx.(type) {
	case fiber.Ctx:
		c.Locals(KeyDataPermApplier, applier)

		return c
	default:
		return context.WithValue(ctx, KeyDataPermApplier, applier)
	}
}
