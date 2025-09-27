package event

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ilxqx/vef-framework-go/event"
	"github.com/ilxqx/vef-framework-go/id"
	"go.uber.org/fx"
)

// memoryEventBus is a simple, thread-safe in-memory event bus implementation
type memoryEventBus struct {
	middlewares []event.Middleware

	// Event subscription management
	subscribers map[string]map[string]*subscription

	// Event processing
	eventCh chan *eventMessage

	// Lifecycle management
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
	mu      sync.RWMutex
	started bool
}

// subscription represents an event subscription
type subscription struct {
	id        string
	eventType string
	handler   event.HandlerFunc
	created   time.Time
}

// eventMessage wraps an event for processing
type eventMessage struct {
	event event.Event
}

// newMemoryEventBus creates a new memory-based event bus
func newMemoryEventBus(middlewares []event.Middleware, lc fx.Lifecycle, c context.Context) *memoryEventBus {
	ctx, cancel := context.WithCancel(c)

	bus := &memoryEventBus{
		middlewares: middlewares,
		subscribers: make(map[string]map[string]*subscription),
		eventCh:     make(chan *eventMessage, 1000),
		ctx:         ctx,
		cancel:      cancel,
	}

	lc.Append(
		fx.StartStopHook(
			func() error {
				if err := bus.Start(); err != nil {
					return fmt.Errorf("failed to start event bus: %v", err)
				}

				logger.Info("event bus started")
				return nil
			},
			func() error {
				if err := bus.Shutdown(c); err != nil {
					return fmt.Errorf("failed to stop event bus: %v", err)
				}

				logger.Info("event bus stopped")
				return nil
			},
		),
	)

	return bus
}

// Start initializes and starts the event bus
func (b *memoryEventBus) Start() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.started {
		return fmt.Errorf("event bus already started")
	}

	// Start event processor goroutine
	b.wg.Go(b.processEvents)
	b.started = true
	return nil
}

// Shutdown gracefully shuts down the event bus
func (b *memoryEventBus) Shutdown(ctx context.Context) error {
	b.mu.Lock()
	if !b.started {
		b.mu.Unlock()
		return nil
	}
	b.mu.Unlock()

	// Signal shutdown
	b.cancel()
	close(b.eventCh)

	// Wait for shutdown with timeout
	done := make(chan struct{})
	go func() {
		b.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-time.After(10 * time.Second):
		return fmt.Errorf("shutdown timeout exceeded")
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Publish publishes an event asynchronously and returns a completion channel
func (b *memoryEventBus) Publish(event event.Event) {
	message := &eventMessage{
		event: event,
	}

	b.eventCh <- message
}

// Subscribe registers a handler for specific event types
func (b *memoryEventBus) Subscribe(eventType string, handler event.HandlerFunc) event.UnsubscribeFunc {
	id := id.GenerateUuid()
	sub := &subscription{
		id:        id,
		eventType: eventType,
		handler:   handler,
		created:   time.Now(),
	}

	b.mu.Lock()
	if b.subscribers[eventType] == nil {
		b.subscribers[eventType] = make(map[string]*subscription)
	}
	b.subscribers[eventType][id] = sub
	b.mu.Unlock()

	// Return unsubscribe function
	unsubscribe := func() {
		b.mu.Lock()
		defer b.mu.Unlock()

		if subs, exists := b.subscribers[eventType]; exists {
			if _, exists := subs[id]; exists {
				delete(subs, id)
				if len(subs) == 0 {
					delete(b.subscribers, eventType)
				}
			}
		}
	}

	return unsubscribe
}

// processEvents is the main event processing goroutine
func (b *memoryEventBus) processEvents() {
	for {
		select {
		case message, ok := <-b.eventCh:
			if !ok {
				return
			}
			go b.handleEvent(message)
		case <-b.ctx.Done():
			return
		}
	}
}

// handleEvent processes a single event message
func (b *memoryEventBus) handleEvent(message *eventMessage) {
	if err := b.deliverEvent(message.event); err != nil {
		logger.Errorf("error delivering event: %v", err)
	}
}

// deliverEvent delivers an event to all matching subscribers
func (b *memoryEventBus) deliverEvent(evt event.Event) error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	eventType := evt.Type()
	// Process middleware chain first
	processedEvent := evt
	for _, middleware := range b.middlewares {
		if err := middleware.Process(b.ctx, processedEvent, func(ctx context.Context, e event.Event) error {
			processedEvent = e
			return nil
		}); err != nil {
			return err
		}
	}

	// Deliver to specific subscribers
	if subs, exists := b.subscribers[eventType]; exists {
		for _, sub := range subs {
			sub.handler(b.ctx, processedEvent)
		}
	}

	return nil
}
