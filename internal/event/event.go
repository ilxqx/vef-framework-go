package event

import (
	"context"

	"github.com/ilxqx/vef-framework-go/event"
)

// NewMemoryBus creates an in-memory event bus bound to the given context.
func NewMemoryBus(c context.Context, middlewares []event.Middleware) event.Bus {
	ctx, cancel := context.WithCancel(c)

	bus := &MemoryBus{
		middlewares: middlewares,
		subscribers: make(map[string]map[string]*subscription),
		eventCh:     make(chan *eventMessage, 1000), // buffered to avoid blocking publishers
		ctx:         ctx,
		cancel:      cancel,
	}

	return bus
}
