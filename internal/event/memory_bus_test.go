package event

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ilxqx/vef-framework-go/event"
)

func TestMemoryEventBus_BasicPublishSubscribe(t *testing.T) {
	t.Run("Single subscriber receives event", func(t *testing.T) {
		bus := createTestEventBus(t)

		var (
			receivedEvent event.Event
			wg            sync.WaitGroup
		)

		wg.Add(1)

		unsubscribe := bus.Subscribe("user.created", func(ctx context.Context, evt event.Event) {
			receivedEvent = evt

			wg.Done()
		})
		defer unsubscribe()

		testEvent := event.NewBaseEvent("user.created", event.WithSource("test-service"))
		bus.Publish(testEvent)

		// Wait for event delivery with timeout
		done := make(chan struct{})

		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			require.NotNil(t, receivedEvent)
			assert.Equal(t, "user.created", receivedEvent.Type())
			assert.Equal(t, "test-service", receivedEvent.Source())
			assert.Equal(t, testEvent.Id(), receivedEvent.Id())

		case <-time.After(100 * time.Millisecond):
			t.Fatal("timeout waiting for event delivery")
		}
	})

	t.Run("Multiple subscribers receive same event", func(t *testing.T) {
		bus := createTestEventBus(t)

		var (
			receivedEvents []event.Event
			mu             sync.Mutex
			wg             sync.WaitGroup
		)

		subscriberCount := 3
		wg.Add(subscriberCount)

		// Create multiple subscribers
		var unsubscribers []event.UnsubscribeFunc
		for range subscriberCount {
			unsub := bus.Subscribe("order.placed", func(ctx context.Context, evt event.Event) {
				mu.Lock()

				receivedEvents = append(receivedEvents, evt)

				mu.Unlock()
				wg.Done()
			})
			unsubscribers = append(unsubscribers, unsub)
		}

		defer func() {
			for _, unsub := range unsubscribers {
				unsub()
			}
		}()

		testEvent := event.NewBaseEvent("order.placed",
			event.WithSource("order-service"),
			event.WithMeta("orderId", "12345"),
		)
		bus.Publish(testEvent)

		// Wait for all subscribers to receive the event
		done := make(chan struct{})

		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			mu.Lock()
			assert.Equal(t, subscriberCount, len(receivedEvents))

			for _, evt := range receivedEvents {
				assert.Equal(t, "order.placed", evt.Type())
				assert.Equal(t, "order-service", evt.Source())
				assert.Equal(t, testEvent.Id(), evt.Id())
			}

			mu.Unlock()

		case <-time.After(100 * time.Millisecond):
			t.Fatal("timeout waiting for event delivery to all subscribers")
		}
	})

	t.Run("Subscribers for different event types", func(t *testing.T) {
		bus := createTestEventBus(t)

		var (
			userEvents  []event.Event
			orderEvents []event.Event
			mu          sync.Mutex
			wg          sync.WaitGroup
		)

		wg.Add(2) // Expecting 2 events

		unsubUser := bus.Subscribe("user.registered", func(ctx context.Context, evt event.Event) {
			mu.Lock()

			userEvents = append(userEvents, evt)

			mu.Unlock()
			wg.Done()
		})
		defer unsubUser()

		unsubOrder := bus.Subscribe("order.created", func(ctx context.Context, evt event.Event) {
			mu.Lock()

			orderEvents = append(orderEvents, evt)

			mu.Unlock()
			wg.Done()
		})
		defer unsubOrder()

		// Publish events of different types
		userEvent := event.NewBaseEvent("user.registered")
		orderEvent := event.NewBaseEvent("order.created")

		bus.Publish(userEvent)
		bus.Publish(orderEvent)

		// Wait for both events
		done := make(chan struct{})

		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			mu.Lock()
			assert.Equal(t, 1, len(userEvents))
			assert.Equal(t, 1, len(orderEvents))
			assert.Equal(t, "user.registered", userEvents[0].Type())
			assert.Equal(t, "order.created", orderEvents[0].Type())
			mu.Unlock()

		case <-time.After(100 * time.Millisecond):
			t.Fatal("timeout waiting for events")
		}
	})
}

func TestMemoryEventBus_Unsubscribe(t *testing.T) {
	t.Run("Unsubscribe prevents further event delivery", func(t *testing.T) {
		bus := createTestEventBus(t)

		var (
			eventCount int
			mu         sync.Mutex
		)

		unsubscribe := bus.Subscribe("payment.processed", func(ctx context.Context, evt event.Event) {
			mu.Lock()

			eventCount++

			mu.Unlock()
		})

		// First event should be delivered
		event1 := event.NewBaseEvent("payment.processed")
		bus.Publish(event1)

		// Give some time for delivery
		time.Sleep(10 * time.Millisecond)

		// Unsubscribe
		unsubscribe()

		// Second event should not be delivered
		event2 := event.NewBaseEvent("payment.processed")
		bus.Publish(event2)

		// Give some time for potential delivery
		time.Sleep(10 * time.Millisecond)

		mu.Lock()
		assert.Equal(t, 1, eventCount) // Should only receive the first event
		mu.Unlock()
	})

	t.Run("Unsubscribe one of multiple subscribers", func(t *testing.T) {
		bus := createTestEventBus(t)

		var (
			subscriber1Count int
			subscriber2Count int
			mu               sync.Mutex
		)

		// First subscriber
		unsubscribe1 := bus.Subscribe("notification.sent", func(ctx context.Context, evt event.Event) {
			mu.Lock()

			subscriber1Count++

			mu.Unlock()
		})

		// Second subscriber
		unsubscribe2 := bus.Subscribe("notification.sent", func(ctx context.Context, evt event.Event) {
			mu.Lock()

			subscriber2Count++

			mu.Unlock()
		})
		defer unsubscribe2()

		// Publish first event - both should receive
		event1 := event.NewBaseEvent("notification.sent")
		bus.Publish(event1)
		time.Sleep(10 * time.Millisecond)

		// Unsubscribe first subscriber
		unsubscribe1()

		// Publish second event - only second subscriber should receive
		event2 := event.NewBaseEvent("notification.sent")
		bus.Publish(event2)
		time.Sleep(10 * time.Millisecond)

		mu.Lock()
		assert.Equal(t, 1, subscriber1Count) // Should only receive first event
		assert.Equal(t, 2, subscriber2Count) // Should receive both events
		mu.Unlock()
	})

	t.Run("Unsubscribe function is idempotent", func(t *testing.T) {
		bus := createTestEventBus(t)

		var (
			eventCount int
			mu         sync.Mutex
		)

		unsubscribe := bus.Subscribe("test.event", func(ctx context.Context, evt event.Event) {
			mu.Lock()

			eventCount++

			mu.Unlock()
		})

		// Call unsubscribe multiple times - should not panic
		unsubscribe()
		unsubscribe()
		unsubscribe()

		// Event should not be delivered
		testEvent := event.NewBaseEvent("test.event")
		bus.Publish(testEvent)
		time.Sleep(10 * time.Millisecond)

		mu.Lock()
		assert.Equal(t, 0, eventCount)
		mu.Unlock()
	})
}

func TestMemoryEventBus_Lifecycle(t *testing.T) {
	t.Run("Start and shutdown", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		bus := &MemoryBus{
			middlewares: []event.Middleware{},
			subscribers: make(map[string]map[string]*subscription),
			eventCh:     make(chan *eventMessage, 1000),
			ctx:         ctx,
			cancel:      cancel,
		}

		// Start the bus
		err := bus.Start()
		require.NoError(t, err)

		// Verify it's started
		assert.True(t, bus.started)

		// Try to start again - should return error
		err = bus.Start()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already started")

		// Shutdown the bus
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err = bus.Shutdown(shutdownCtx)
		require.NoError(t, err)
	})

	t.Run("Shutdown without start", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		bus := &MemoryBus{
			middlewares: []event.Middleware{},
			subscribers: make(map[string]map[string]*subscription),
			eventCh:     make(chan *eventMessage, 1000),
			ctx:         ctx,
			cancel:      cancel,
		}

		// Shutdown without starting - should not error
		err := bus.Shutdown(context.Background())
		assert.NoError(t, err)
	})

	t.Run("Events are processed after start", func(t *testing.T) {
		bus := createTestEventBus(t)

		var (
			receivedEvent event.Event
			wg            sync.WaitGroup
		)

		wg.Add(1)

		unsubscribe := bus.Subscribe("lifecycle.test", func(ctx context.Context, evt event.Event) {
			receivedEvent = evt

			wg.Done()
		})
		defer unsubscribe()

		testEvent := event.NewBaseEvent("lifecycle.test")
		bus.Publish(testEvent)

		done := make(chan struct{})

		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			assert.Equal(t, testEvent.Id(), receivedEvent.Id())
		case <-time.After(100 * time.Millisecond):
			t.Fatal("timeout waiting for event after start")
		}
	})
}

func TestMemoryEventBus_Middleware(t *testing.T) {
	t.Run("Middleware processes events", func(t *testing.T) {
		var (
			processedEvents []event.Event
			mu              sync.Mutex
		)

		middleware := &testMiddleware{
			processFunc: func(ctx context.Context, evt event.Event, next event.MiddlewareFunc) error {
				mu.Lock()

				processedEvents = append(processedEvents, evt)

				mu.Unlock()

				return next(ctx, evt)
			},
		}

		bus := createTestEventBusWithMiddleware(t, []event.Middleware{middleware})

		var (
			receivedEvent event.Event
			wg            sync.WaitGroup
		)

		wg.Add(1)

		unsubscribe := bus.Subscribe("middleware.test", func(ctx context.Context, evt event.Event) {
			receivedEvent = evt

			wg.Done()
		})
		defer unsubscribe()

		testEvent := event.NewBaseEvent("middleware.test")
		bus.Publish(testEvent)

		done := make(chan struct{})

		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			mu.Lock()
			assert.Equal(t, 1, len(processedEvents))
			assert.Equal(t, testEvent.Id(), processedEvents[0].Id())
			mu.Unlock()
			assert.Equal(t, testEvent.Id(), receivedEvent.Id())

		case <-time.After(100 * time.Millisecond):
			t.Fatal("timeout waiting for middleware processing")
		}
	})

	t.Run("Middleware chain processes in order", func(t *testing.T) {
		var (
			processingOrder []string
			mu              sync.Mutex
		)

		middleware1 := &testMiddleware{
			processFunc: func(ctx context.Context, evt event.Event, next event.MiddlewareFunc) error {
				mu.Lock()

				processingOrder = append(processingOrder, "middleware1")

				mu.Unlock()

				return next(ctx, evt)
			},
		}

		middleware2 := &testMiddleware{
			processFunc: func(ctx context.Context, evt event.Event, next event.MiddlewareFunc) error {
				mu.Lock()

				processingOrder = append(processingOrder, "middleware2")

				mu.Unlock()

				return next(ctx, evt)
			},
		}

		bus := createTestEventBusWithMiddleware(t, []event.Middleware{middleware1, middleware2})

		var wg sync.WaitGroup
		wg.Add(1)

		unsubscribe := bus.Subscribe("chain.test", func(ctx context.Context, evt event.Event) {
			wg.Done()
		})
		defer unsubscribe()

		testEvent := event.NewBaseEvent("chain.test")
		bus.Publish(testEvent)

		done := make(chan struct{})

		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			mu.Lock()
			assert.Equal(t, []string{"middleware1", "middleware2"}, processingOrder)
			mu.Unlock()
		case <-time.After(100 * time.Millisecond):
			t.Fatal("timeout waiting for middleware chain processing")
		}
	})
}

func TestMemoryEventBus_Concurrency(t *testing.T) {
	t.Run("Concurrent publish and subscribe", func(t *testing.T) {
		bus := createTestEventBus(t)

		const (
			numPublishers      = 10
			numSubscribers     = 5
			eventsPerPublisher = 20
		)

		var (
			totalReceived int
			mu            sync.Mutex
			wg            sync.WaitGroup
		)

		// Create subscribers
		var unsubscribers []event.UnsubscribeFunc

		for range numSubscribers {
			wg.Add(eventsPerPublisher * numPublishers) // Each subscriber should receive all events

			unsub := bus.Subscribe("concurrent.test", func(ctx context.Context, evt event.Event) {
				mu.Lock()

				totalReceived++

				mu.Unlock()
				wg.Done()
			})
			unsubscribers = append(unsubscribers, unsub)
		}

		defer func() {
			for _, unsub := range unsubscribers {
				unsub()
			}
		}()

		// Create publishers
		for i := range numPublishers {
			go func(publisherID int) {
				for j := range eventsPerPublisher {
					event := event.NewBaseEvent("concurrent.test",
						event.WithMeta("publisherID", string(rune(publisherID+'0'))),
						event.WithMeta("eventNum", string(rune(j+'0'))),
					)
					bus.Publish(event)
				}
			}(i)
		}

		// Wait for all events to be processed
		done := make(chan struct{})

		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			expectedTotal := numPublishers * eventsPerPublisher * numSubscribers

			mu.Lock()
			assert.Equal(t, expectedTotal, totalReceived)
			mu.Unlock()

		case <-time.After(5 * time.Second):
			t.Fatal("timeout waiting for concurrent event processing")
		}
	})

	t.Run("Concurrent subscribe and unsubscribe", func(t *testing.T) {
		bus := createTestEventBus(t)

		const numRoutines = 50

		var wg sync.WaitGroup
		wg.Add(numRoutines)

		// Concurrently subscribe and unsubscribe
		for range numRoutines {
			go func() {
				defer wg.Done()

				unsubscribe := bus.Subscribe("concurrent.unsub.test", func(ctx context.Context, evt event.Event) {
					// Do nothing
				})

				// Immediately unsubscribe
				unsubscribe()
			}()
		}

		// Wait for all routines to complete
		done := make(chan struct{})

		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			// Test passes if no deadlock or panic occurs
		case <-time.After(5 * time.Second):
			t.Fatal("timeout during concurrent subscribe/unsubscribe")
		}
	})
}

// Helper functions and test utilities

// testMiddleware implements the Middleware interface for testing.
type testMiddleware struct {
	processFunc func(ctx context.Context, event event.Event, next event.MiddlewareFunc) error
}

func (m *testMiddleware) Process(ctx context.Context, evt event.Event, next event.MiddlewareFunc) error {
	return m.processFunc(ctx, evt, next)
}

// createTestEventBus creates a memory event bus for testing.
func createTestEventBus(t *testing.T) event.Bus {
	return createTestEventBusWithMiddleware(t, []event.Middleware{})
}

// createTestEventBusWithMiddleware creates a memory event bus with custom middleware for testing.
func createTestEventBusWithMiddleware(t *testing.T, middlewares []event.Middleware) event.Bus {
	bus := NewMemoryBus(middlewares)

	err := bus.Start()
	require.NoError(t, err)

	t.Cleanup(func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		_ = bus.Shutdown(shutdownCtx)
	})

	return bus
}
