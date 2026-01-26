package param

import (
	"reflect"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/event"
)

type PublisherFactoryResolver struct {
	publisher event.Publisher
}

func NewPublisherFactoryResolver(publisher event.Publisher) api.FactoryParamResolver {
	return &PublisherFactoryResolver{publisher: publisher}
}

func (r *PublisherFactoryResolver) Type() reflect.Type {
	return reflect.TypeFor[event.Publisher]()
}

func (r *PublisherFactoryResolver) Resolve() (reflect.Value, error) {
	return reflect.ValueOf(r.publisher), nil
}
