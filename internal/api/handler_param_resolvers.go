package api

import (
	"reflect"

	"github.com/gofiber/fiber/v3"
	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/contextx"
	"github.com/ilxqx/vef-framework-go/log"
	"github.com/ilxqx/vef-framework-go/orm"
	"github.com/ilxqx/vef-framework-go/security"
)

// presetParamResolvers defines the built-in parameter resolvers available by default.
// Additional resolvers can be supplied via DI and will override these when type overlaps.
var presetParamResolvers = []api.HandlerParamResolver{
	new(ctxParamResolver),
	new(dbParamResolver),
	new(loggerParamResolver),
	new(principalParamResolver),
}

type ctxParamResolver struct{}

func (*ctxParamResolver) Type() reflect.Type {
	return reflect.TypeFor[fiber.Ctx]()
}

func (*ctxParamResolver) Resolve(ctx fiber.Ctx) (reflect.Value, error) {
	return reflect.ValueOf(ctx), nil
}

type dbParamResolver struct{}

func (*dbParamResolver) Type() reflect.Type {
	return reflect.TypeFor[orm.Db]()
}

func (*dbParamResolver) Resolve(ctx fiber.Ctx) (reflect.Value, error) {
	return reflect.ValueOf(contextx.Db(ctx)), nil
}

type loggerParamResolver struct{}

func (*loggerParamResolver) Type() reflect.Type {
	return reflect.TypeFor[log.Logger]()
}

func (*loggerParamResolver) Resolve(ctx fiber.Ctx) (reflect.Value, error) {
	return reflect.ValueOf(contextx.Logger(ctx)), nil
}

type principalParamResolver struct{}

func (*principalParamResolver) Type() reflect.Type {
	return reflect.TypeFor[*security.Principal]()
}

func (*principalParamResolver) Resolve(ctx fiber.Ctx) (reflect.Value, error) {
	return reflect.ValueOf(contextx.Principal(ctx)), nil
}
