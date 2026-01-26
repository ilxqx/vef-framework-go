package param

import (
	"reflect"

	"github.com/gofiber/fiber/v3"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/event"
)

type PublisherResolver struct {
	publisher event.Publisher
}

func NewPublisherResolver(publisher event.Publisher) api.HandlerParamResolver {
	return &PublisherResolver{publisher: publisher}
}

func (r *PublisherResolver) Type() reflect.Type {
	return reflect.TypeFor[event.Publisher]()
}

func (r *PublisherResolver) Resolve(fiber.Ctx) (reflect.Value, error) {
	return reflect.ValueOf(r.publisher), nil
}
