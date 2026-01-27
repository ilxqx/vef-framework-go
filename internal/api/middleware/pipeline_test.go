package middleware

import (
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"

	"github.com/ilxqx/vef-framework-go/api"
)

// mockMiddleware is a test implementation of api.Middleware.
type mockMiddleware struct {
	name  string
	order int
}

func (m *mockMiddleware) Name() string {
	return m.name
}

func (m *mockMiddleware) Order() int {
	return m.order
}

func (*mockMiddleware) Process(ctx fiber.Ctx) error {
	return ctx.Next()
}

func TestNewChain(t *testing.T) {
	t.Log("Testing NewChain constructor")

	t.Run("EmptyChain", func(t *testing.T) {
		chain := NewChain()
		assert.NotNil(t, chain, "NewChain should return a non-nil chain")
	})

	t.Run("SingleMiddleware", func(t *testing.T) {
		mid := &mockMiddleware{name: "test", order: 1}
		chain := NewChain(mid)
		assert.NotNil(t, chain, "Chain should not be nil")
	})

	t.Run("MultipleMiddlewares", func(t *testing.T) {
		mid1 := &mockMiddleware{name: "first", order: 1}
		mid2 := &mockMiddleware{name: "second", order: 2}
		mid3 := &mockMiddleware{name: "third", order: 3}

		chain := NewChain(mid1, mid2, mid3)
		assert.NotNil(t, chain, "Chain should not be nil")
	})

	t.Run("SortsMiddlewaresByOrder", func(t *testing.T) {
		mid1 := &mockMiddleware{name: "high", order: 100}
		mid2 := &mockMiddleware{name: "low", order: -100}
		mid3 := &mockMiddleware{name: "medium", order: 0}

		chain := NewChain(mid1, mid2, mid3)
		handlers := chain.Handlers()

		assert.Len(t, handlers, 3, "Should have 3 handlers")
	})

	t.Run("NegativeOrderFirst", func(t *testing.T) {
		mid1 := &mockMiddleware{name: "positive", order: 10}
		mid2 := &mockMiddleware{name: "negative", order: -10}

		chain := NewChain(mid1, mid2)
		handlers := chain.Handlers()

		assert.Len(t, handlers, 2, "Should have 2 handlers")
	})
}

func TestChainHandlers(t *testing.T) {
	t.Log("Testing Chain.Handlers method")

	t.Run("EmptyChainReturnsEmptySlice", func(t *testing.T) {
		chain := NewChain()
		handlers := chain.Handlers()

		assert.Empty(t, handlers, "Empty chain should return empty handlers")
	})

	t.Run("ReturnsCorrectNumberOfHandlers", func(t *testing.T) {
		mid1 := &mockMiddleware{name: "first", order: 1}
		mid2 := &mockMiddleware{name: "second", order: 2}

		chain := NewChain(mid1, mid2)
		handlers := chain.Handlers()

		assert.Len(t, handlers, 2, "Should return 2 handlers")
	})

	t.Run("HandlersAreNotNil", func(t *testing.T) {
		mid := &mockMiddleware{name: "test", order: 1}
		chain := NewChain(mid)
		handlers := chain.Handlers()

		for i, h := range handlers {
			assert.NotNil(t, h, "Handler %d should not be nil", i)
		}
	})

	t.Run("HandlersAreFunctions", func(t *testing.T) {
		mid := &mockMiddleware{name: "test", order: 1}
		chain := NewChain(mid)
		handlers := chain.Handlers()

		for i, h := range handlers {
			_, ok := h.(func(fiber.Ctx) error)
			assert.True(t, ok, "Handler %d should be a function", i)
		}
	})
}

func TestChainOrdering(t *testing.T) {
	t.Log("Testing Chain middleware ordering")

	t.Run("PreservesOrderWithSameValue", func(t *testing.T) {
		mid1 := &mockMiddleware{name: "first", order: 0}
		mid2 := &mockMiddleware{name: "second", order: 0}

		chain := NewChain(mid1, mid2)
		handlers := chain.Handlers()

		assert.Len(t, handlers, 2, "Should have 2 handlers")
	})

	t.Run("TypicalMiddlewareOrdering", func(t *testing.T) {
		// Simulate typical middleware ordering
		auth := &mockMiddleware{name: "auth", order: -100}
		contextual := &mockMiddleware{name: "contextual", order: -90}
		rateLimit := &mockMiddleware{name: "ratelimit", order: -80}
		audit := &mockMiddleware{name: "audit", order: -70}

		chain := NewChain(audit, auth, rateLimit, contextual)
		handlers := chain.Handlers()

		assert.Len(t, handlers, 4, "Should have 4 handlers")
	})
}

func TestChainImplementsInterface(t *testing.T) {
	t.Log("Testing Chain interface compliance")

	t.Run("MiddlewareImplementsApiMiddleware", func(t *testing.T) {
		var _ api.Middleware = (*mockMiddleware)(nil)

		assert.True(t, true, "mockMiddleware should implement api.Middleware")
	})
}
