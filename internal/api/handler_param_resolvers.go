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

var presetParamResolvers = []api.HandlerParamResolver{
	new(CtxParamResolver),
	new(DbParamResolver),
	new(LoggerParamResolver),
	new(PrincipalParamResolver),
}

type CtxParamResolver struct{}

func (*CtxParamResolver) Type() reflect.Type {
	return reflect.TypeFor[fiber.Ctx]()
}

func (*CtxParamResolver) Resolve(ctx fiber.Ctx) (reflect.Value, error) {
	return reflect.ValueOf(ctx), nil
}

type DbParamResolver struct{}

func (*DbParamResolver) Type() reflect.Type {
	return reflect.TypeFor[orm.Db]()
}

func (*DbParamResolver) Resolve(ctx fiber.Ctx) (reflect.Value, error) {
	return reflect.ValueOf(contextx.Db(ctx)), nil
}

type LoggerParamResolver struct{}

func (*LoggerParamResolver) Type() reflect.Type {
	return reflect.TypeFor[log.Logger]()
}

func (*LoggerParamResolver) Resolve(ctx fiber.Ctx) (reflect.Value, error) {
	return reflect.ValueOf(contextx.Logger(ctx)), nil
}

type PrincipalParamResolver struct{}

func (*PrincipalParamResolver) Type() reflect.Type {
	return reflect.TypeFor[*security.Principal]()
}

func (*PrincipalParamResolver) Resolve(ctx fiber.Ctx) (reflect.Value, error) {
	return reflect.ValueOf(contextx.Principal(ctx)), nil
}
