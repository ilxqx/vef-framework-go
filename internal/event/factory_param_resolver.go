package event

import (
	"reflect"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/event"
)

// PublisherFactoryParamResolver provides event.Publisher for handler factory functions.
type PublisherFactoryParamResolver struct {
	publisher event.Publisher
}

// NewPublisherFactoryParamResolver creates a new PublisherFactoryParamResolver.
func NewPublisherFactoryParamResolver(publisher event.Publisher) api.FactoryParamResolver {
	return &PublisherFactoryParamResolver{publisher: publisher}
}

// Type returns the type this resolver handles.
func (r *PublisherFactoryParamResolver) Type() reflect.Type {
	return reflect.TypeFor[event.Publisher]()
}

// Resolve returns the publisher instance.
func (r *PublisherFactoryParamResolver) Resolve() reflect.Value {
	return reflect.ValueOf(r.publisher)
}
