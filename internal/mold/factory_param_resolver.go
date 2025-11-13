package mold

import (
	"reflect"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/mold"
)

// TransformerFactoryParamResolver provides mold.Transformer for handler factory functions.
type TransformerFactoryParamResolver struct {
	transformer mold.Transformer
}

// NewTransformerFactoryParamResolver creates a new TransformerFactoryParamResolver.
func NewTransformerFactoryParamResolver(transformer mold.Transformer) api.FactoryParamResolver {
	return &TransformerFactoryParamResolver{transformer: transformer}
}

// Type returns the type this resolver handles.
func (r *TransformerFactoryParamResolver) Type() reflect.Type {
	return reflect.TypeFor[mold.Transformer]()
}

// Resolve returns the transformer instance.
func (r *TransformerFactoryParamResolver) Resolve() reflect.Value {
	return reflect.ValueOf(r.transformer)
}
