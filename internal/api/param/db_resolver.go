package param

import (
	"reflect"

	"github.com/gofiber/fiber/v3"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/contextx"
	"github.com/ilxqx/vef-framework-go/orm"
)

type DBResolver struct{}

func (*DBResolver) Type() reflect.Type {
	return reflect.TypeFor[orm.DB]()
}

func (*DBResolver) Resolve(ctx fiber.Ctx) (reflect.Value, error) {
	return reflect.ValueOf(contextx.DB(ctx)), nil
}

func NewDBResolver() api.HandlerParamResolver {
	return new(DBResolver)
}
