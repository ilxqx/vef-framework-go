package param

import (
	"reflect"

	"github.com/gofiber/fiber/v3"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/contextx"
	"github.com/ilxqx/vef-framework-go/log"
)

type LoggerResolver struct{}

func (*LoggerResolver) Type() reflect.Type {
	return reflect.TypeFor[log.Logger]()
}

func (*LoggerResolver) Resolve(ctx fiber.Ctx) (reflect.Value, error) {
	return reflect.ValueOf(contextx.Logger(ctx)), nil
}

func NewLoggerResolver() api.HandlerParamResolver {
	return new(LoggerResolver)
}
