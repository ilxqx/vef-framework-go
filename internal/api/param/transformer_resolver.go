package param

import (
	"reflect"

	"github.com/gofiber/fiber/v3"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/mold"
)

type TransformerResolver struct {
	transformer mold.Transformer
}

func NewTransformerResolver(transformer mold.Transformer) api.HandlerParamResolver {
	return &TransformerResolver{transformer: transformer}
}

func (*TransformerResolver) Type() reflect.Type {
	return reflect.TypeFor[mold.Transformer]()
}

func (r *TransformerResolver) Resolve(_ fiber.Ctx) (reflect.Value, error) {
	return reflect.ValueOf(r.transformer), nil
}
