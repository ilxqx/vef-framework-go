package event

import (
	"reflect"

	"github.com/gofiber/fiber/v3"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/event"
)

// PublisherHandlerParamResolver resolves event.Publisher for handler parameters.
type PublisherHandlerParamResolver struct {
	publisher event.Publisher
}

// NewPublisherHandlerParamResolver creates a new event publisher parameter resolver.
func NewPublisherHandlerParamResolver(publisher event.Publisher) api.HandlerParamResolver {
	return &PublisherHandlerParamResolver{publisher: publisher}
}

func (r *PublisherHandlerParamResolver) Type() reflect.Type {
	return reflect.TypeFor[event.Publisher]()
}

func (r *PublisherHandlerParamResolver) Resolve(fiber.Ctx) (reflect.Value, error) {
	return reflect.ValueOf(r.publisher), nil
}
