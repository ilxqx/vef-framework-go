package param

import (
	"reflect"

	"github.com/gofiber/fiber/v3"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/internal/api/common"
)

type MetaResolver struct{}

func (*MetaResolver) Type() reflect.Type {
	return reflect.TypeFor[api.Meta]()
}

func (*MetaResolver) Resolve(ctx fiber.Ctx) (reflect.Value, error) {
	if req := common.Request(ctx); req != nil && req.Meta != nil {
		return reflect.ValueOf(req.Meta), nil
	}

	return reflect.ValueOf(api.Meta{}), nil
}

func NewMetaResolver() api.HandlerParamResolver {
	return new(MetaResolver)
}
