package param

import (
	"reflect"

	"github.com/gofiber/fiber/v3"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/internal/api/shared"
)

type ParamsResolver struct{}

func (*ParamsResolver) Type() reflect.Type {
	return reflect.TypeFor[api.Params]()
}

func (*ParamsResolver) Resolve(ctx fiber.Ctx) (reflect.Value, error) {
	if req := shared.Request(ctx); req != nil && req.Params != nil {
		return reflect.ValueOf(req.Params), nil
	}

	return reflect.ValueOf(api.Params{}), nil
}

func NewParamsResolver() api.HandlerParamResolver {
	return new(ParamsResolver)
}
