package param

import (
	"reflect"

	"github.com/gofiber/fiber/v3"

	"github.com/ilxqx/vef-framework-go/api"
)

type CtxResolver struct{}

func (*CtxResolver) Type() reflect.Type {
	return reflect.TypeFor[fiber.Ctx]()
}

func (*CtxResolver) Resolve(ctx fiber.Ctx) (reflect.Value, error) {
	return reflect.ValueOf(ctx), nil
}

func NewCtxResolver() api.HandlerParamResolver {
	return new(CtxResolver)
}
