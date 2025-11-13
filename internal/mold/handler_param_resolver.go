package mold

import (
	"reflect"

	"github.com/gofiber/fiber/v3"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/mold"
)

// TransformerHandlerParamResolver resolves mold.Transformer for handler parameters.
type TransformerHandlerParamResolver struct {
	transformer mold.Transformer
}

// NewTransformerHandlerParamResolver creates a new mold transformer parameter resolver.
func NewTransformerHandlerParamResolver(transformer mold.Transformer) api.HandlerParamResolver {
	return &TransformerHandlerParamResolver{transformer: transformer}
}

func (r *TransformerHandlerParamResolver) Type() reflect.Type {
	return reflect.TypeFor[mold.Transformer]()
}

func (r *TransformerHandlerParamResolver) Resolve(fiber.Ctx) (reflect.Value, error) {
	return reflect.ValueOf(r.transformer), nil
}
