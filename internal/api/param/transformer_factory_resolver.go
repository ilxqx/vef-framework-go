package param

import (
	"reflect"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/mold"
)

type TransformerFactoryResolver struct {
	transformer mold.Transformer
}

func NewTransformerFactoryResolver(transformer mold.Transformer) api.FactoryParamResolver {
	return &TransformerFactoryResolver{transformer: transformer}
}

func (*TransformerFactoryResolver) Type() reflect.Type {
	return reflect.TypeFor[mold.Transformer]()
}

func (r *TransformerFactoryResolver) Resolve() (reflect.Value, error) {
	return reflect.ValueOf(r.transformer), nil
}
