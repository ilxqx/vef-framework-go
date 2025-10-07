package event

import (
	"context"
	"time"
)

// Event represents the base interface for all events in the system.
// All custom events should embed this interface to be compatible with the event bus.
type Event interface {
	// Id returns a unique identifier for this specific event instance.
	Id() string
	// Type returns a unique string identifier for the event type.
	// This is used for routing and filtering events.
	Type() string
	// Source returns the source that generated this event.
	Source() string
	// Time returns when the event occurred.
	Time() time.Time
	// Meta returns the metadata for the event.
	Meta() map[string]string
}

// HandlerFunc represents a function that can handle events.
// The handler receives the event and a context for cancellation/timeout control.
type HandlerFunc func(ctx context.Context, event Event)

// Publisher defines the interface for publishing events to the event bus.
type Publisher interface {
	// Publish sends an event to all registered subscribers asynchronously.
	Publish(event Event)
}

// UnsubscribeFunc is a function that can be called to unsubscribe from an event.
type UnsubscribeFunc func()

// Subscriber defines the interface for subscribing to events.
type Subscriber interface {
	// Subscribe registers a handler for events of a specific type.
	// Returns an unsubscribe function that can be called to remove the subscription.
	Subscribe(eventType string, handler HandlerFunc) UnsubscribeFunc
}

// Bus combines Publisher and Subscriber interfaces along with lifecycle management.
type Bus interface {
	Publisher
	Subscriber

	// Start initializes the event bus and begins processing events.
	Start() error
	// Shutdown gracefully shuts down the event bus.
	Shutdown(ctx context.Context) error
}

// Middleware defines an interface for event processing middleware.
// Middleware can intercept and modify events before they reach handlers.
type Middleware interface {
	// Process is called for each event before it's delivered to handlers.
	// It can modify the event, context, or prevent delivery by returning an error.
	Process(ctx context.Context, event Event, next MiddlewareFunc) error
}

// MiddlewareFunc is a function type for middleware processing.
type MiddlewareFunc func(ctx context.Context, event Event) error
