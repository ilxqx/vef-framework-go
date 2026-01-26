package api

import (
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/internal/api/common"
)

// mockResource implements api.Resource for testing.
type mockResource struct {
	kind    api.Kind
	name    string
	version string
	auth    *api.AuthConfig
}

func (m *mockResource) Kind() api.Kind                { return m.kind }
func (m *mockResource) Name() string                  { return m.name }
func (m *mockResource) Version() string               { return m.version }
func (m *mockResource) Auth() *api.AuthConfig         { return m.auth }
func (*mockResource) Operations() []api.OperationSpec { return nil }

// mockRouterStrategy implements api.RouterStrategy for testing.
type mockRouterStrategy struct {
	name       string
	setupErr   error
	setupCalls int
	routeCalls int
}

func (m *mockRouterStrategy) Name() string {
	return m.name
}

func (m *mockRouterStrategy) CanHandle(kind api.Kind) bool {
	return kind.String() == m.name
}

func (m *mockRouterStrategy) Setup(fiber.Router) error {
	m.setupCalls++

	return m.setupErr
}

func (m *mockRouterStrategy) Route(fiber.Handler, *api.Operation) {
	m.routeCalls++
}

// mockOperationsCollector implements api.OperationsCollector for testing.
type mockOperationsCollector struct {
	specs []api.OperationSpec
}

func (m *mockOperationsCollector) Collect(api.Resource) []api.OperationSpec {
	return m.specs
}

// mockHandlerResolver implements api.HandlerResolver for testing.
type mockHandlerResolver struct {
	handler   any
	err       error
	returnNil bool
}

func (m *mockHandlerResolver) Resolve(api.Resource, api.OperationSpec) (any, error) {
	if m.err != nil {
		return nil, m.err
	}

	if m.returnNil {
		return nil, nil
	}

	return m.handler, nil
}

// mockHandlerAdapter implements api.HandlerAdapter for testing.
type mockHandlerAdapter struct {
	err       error
	returnNil bool
}

func (m *mockHandlerAdapter) Adapt(handler any, _ *api.Operation) (fiber.Handler, error) {
	if m.err != nil {
		return nil, m.err
	}

	if m.returnNil || handler == nil {
		return nil, nil
	}

	return func(fiber.Ctx) error { return nil }, nil
}

// dummyHandler is a simple handler for testing.
func dummyHandler(fiber.Ctx) error {
	return nil
}

func TestNewEngine(t *testing.T) {
	t.Log("Testing NewEngine constructor")

	t.Run("ReturnsNonNilEngine", func(t *testing.T) {
		eng, err := NewEngine()
		assert.NoError(t, err, "NewEngine should not return error")
		assert.NotNil(t, eng, "NewEngine should return a non-nil engine")
	})

	t.Run("DefaultTimeout", func(t *testing.T) {
		eng, _ := NewEngine()
		e := eng.(*engine)
		assert.Equal(t, 30*time.Second, e.defaultTimeout, "Default timeout should be 30 seconds")
	})

	t.Run("DefaultVersion", func(t *testing.T) {
		eng, _ := NewEngine()
		e := eng.(*engine)
		assert.Equal(t, api.VersionV1, e.defaultVersion, "Default version should be v1")
	})

	t.Run("DefaultAuthIsBearer", func(t *testing.T) {
		eng, _ := NewEngine()
		e := eng.(*engine)
		assert.NotNil(t, e.defaultAuth, "Default auth should not be nil")
		assert.Equal(t, api.AuthStrategyBearer, e.defaultAuth.Strategy, "Default auth strategy should be bearer")
	})

	t.Run("DefaultRateLimit", func(t *testing.T) {
		eng, _ := NewEngine()
		e := eng.(*engine)
		assert.NotNil(t, e.defaultRateLimit, "Default rate limit should not be nil")
		assert.Equal(t, 100, e.defaultRateLimit.Max, "Default rate limit max should be 100")
		assert.Equal(t, 5*time.Minute, e.defaultRateLimit.Period, "Default rate limit period should be 5 minutes")
	})

	t.Run("OperationsMapInitialized", func(t *testing.T) {
		eng, _ := NewEngine()
		e := eng.(*engine)
		assert.NotNil(t, e.operations, "Operations map should be initialized")
	})
}

func TestEngineOptions(t *testing.T) {
	t.Log("Testing Engine options")

	t.Run("WithDefaultTimeout", func(t *testing.T) {
		customTimeout := 60 * time.Second
		eng, _ := NewEngine(WithDefaultTimeout(customTimeout))
		e := eng.(*engine)
		assert.Equal(t, customTimeout, e.defaultTimeout, "Custom timeout should be applied")
	})

	t.Run("WithDefaultVersion", func(t *testing.T) {
		customVersion := "v2"
		eng, _ := NewEngine(WithDefaultVersion(customVersion))
		e := eng.(*engine)
		assert.Equal(t, customVersion, e.defaultVersion, "Custom version should be applied")
	})

	t.Run("WithDefaultAuth", func(t *testing.T) {
		customAuth := api.SignatureAuth()
		eng, _ := NewEngine(WithDefaultAuth(customAuth))
		e := eng.(*engine)
		assert.Equal(t, api.AuthStrategySignature, e.defaultAuth.Strategy, "Custom auth strategy should be applied")
	})

	t.Run("WithDefaultRateLimit", func(t *testing.T) {
		customRateLimit := &api.RateLimitConfig{
			Max:    50,
			Period: 1 * time.Minute,
		}
		eng, _ := NewEngine(WithDefaultRateLimit(customRateLimit))
		e := eng.(*engine)
		assert.Equal(t, 50, e.defaultRateLimit.Max, "Custom rate limit max should be applied")
		assert.Equal(t, 1*time.Minute, e.defaultRateLimit.Period, "Custom rate limit period should be applied")
	})

	t.Run("WithRouters", func(t *testing.T) {
		router1 := &mockRouterStrategy{name: "rpc"}
		router2 := &mockRouterStrategy{name: "rest"}
		eng, _ := NewEngine(WithRouters(router1, router2))
		e := eng.(*engine)
		assert.NotNil(t, e.routerOperations, "Router operations map should be initialized")
		assert.Equal(t, 2, e.routerOperations.Size(), "Should have 2 routers")
	})

	t.Run("WithOperationCollectors", func(t *testing.T) {
		collector := &mockOperationsCollector{}
		eng, _ := NewEngine(WithOperationCollectors(collector))
		e := eng.(*engine)
		assert.Len(t, e.collectors, 1, "Should have 1 collector")
	})

	t.Run("WithHandlerResolvers", func(t *testing.T) {
		resolver := &mockHandlerResolver{handler: dummyHandler}
		eng, _ := NewEngine(WithHandlerResolvers(resolver))
		e := eng.(*engine)
		assert.Len(t, e.resolvers, 1, "Should have 1 resolver")
	})

	t.Run("WithHandlerAdapters", func(t *testing.T) {
		adapter := &mockHandlerAdapter{}
		eng, _ := NewEngine(WithHandlerAdapters(adapter))
		e := eng.(*engine)
		assert.Len(t, e.adapters, 1, "Should have 1 adapter")
	})

	t.Run("MultipleOptions", func(t *testing.T) {
		eng, _ := NewEngine(
			WithDefaultTimeout(45*time.Second),
			WithDefaultVersion("v3"),
			WithDefaultAuth(api.Public()),
		)
		e := eng.(*engine)
		assert.Equal(t, 45*time.Second, e.defaultTimeout, "Custom timeout should be applied")
		assert.Equal(t, "v3", e.defaultVersion, "Custom version should be applied")
		assert.Equal(t, api.AuthStrategyNone, e.defaultAuth.Strategy, "Custom auth should be applied")
	})
}

func TestEngineRegister(t *testing.T) {
	t.Log("Testing Engine.Register method")

	t.Run("NilResource", func(t *testing.T) {
		eng, _ := NewEngine()
		err := eng.Register(nil)
		assert.Error(t, err, "Register should return error for nil resource")
		assert.Contains(t, err.Error(), "nil", "Error message should mention nil")
	})

	t.Run("EmptyResourceName", func(t *testing.T) {
		eng, _ := NewEngine(
			WithOperationCollectors(&mockOperationsCollector{
				specs: []api.OperationSpec{{Action: "create"}},
			}),
		)
		res := &mockResource{kind: api.KindRPC, name: ""}
		err := eng.Register(res)
		assert.Error(t, err, "Register should return error for empty resource name")
		assert.Contains(t, err.Error(), "empty", "Error message should mention empty")
	})

	t.Run("EmptyActionName", func(t *testing.T) {
		router := &mockRouterStrategy{name: "rpc"}
		collector := &mockOperationsCollector{
			specs: []api.OperationSpec{{Action: ""}},
		}
		eng, _ := NewEngine(
			WithRouters(router),
			WithOperationCollectors(collector),
		)
		res := &mockResource{kind: api.KindRPC, name: "test/resource"}
		err := eng.Register(res)
		assert.Error(t, err, "Register should return error for empty action name")
		assert.Contains(t, err.Error(), "empty", "Error message should mention empty")
	})

	t.Run("NoRouterForKind", func(t *testing.T) {
		router := &mockRouterStrategy{name: "rest"} // Only REST router, but resource is RPC
		collector := &mockOperationsCollector{
			specs: []api.OperationSpec{{Action: "create", Handler: dummyHandler}},
		}
		eng, _ := NewEngine(
			WithRouters(router),
			WithOperationCollectors(collector),
		)
		res := &mockResource{kind: api.KindRPC, name: "test/resource"}
		err := eng.Register(res)
		assert.Error(t, err, "Register should return error when no router handles kind")
		assert.Contains(t, err.Error(), "no router", "Error message should mention no router")
	})

	t.Run("HandlerResolverError", func(t *testing.T) {
		router := &mockRouterStrategy{name: "rpc"}
		collector := &mockOperationsCollector{
			specs: []api.OperationSpec{{Action: "create"}},
		}
		resolver := &mockHandlerResolver{err: errors.New("resolve failed")}
		eng, _ := NewEngine(
			WithRouters(router),
			WithOperationCollectors(collector),
			WithHandlerResolvers(resolver),
		)
		res := &mockResource{kind: api.KindRPC, name: "test/resource"}
		err := eng.Register(res)
		assert.Error(t, err, "Register should return error when handler resolver fails")
		assert.Contains(t, err.Error(), "resolve", "Error message should mention resolve")
	})

	t.Run("NoHandlerResolver", func(t *testing.T) {
		router := &mockRouterStrategy{name: "rpc"}
		collector := &mockOperationsCollector{
			specs: []api.OperationSpec{{Action: "create"}},
		}
		eng, _ := NewEngine(
			WithRouters(router),
			WithOperationCollectors(collector),
		)
		res := &mockResource{kind: api.KindRPC, name: "test/resource"}
		err := eng.Register(res)
		assert.Error(t, err, "Register should return error when no handler resolver found")
	})

	t.Run("SuccessfulRegistration", func(t *testing.T) {
		router := &mockRouterStrategy{name: "rpc"}
		collector := &mockOperationsCollector{
			specs: []api.OperationSpec{{Action: "create"}},
		}
		resolver := &mockHandlerResolver{handler: dummyHandler}
		eng, _ := NewEngine(
			WithRouters(router),
			WithOperationCollectors(collector),
			WithHandlerResolvers(resolver),
		)
		res := &mockResource{kind: api.KindRPC, name: "test/resource"}
		err := eng.Register(res)
		assert.NoError(t, err, "Register should succeed")
	})

	t.Run("DuplicateOperation", func(t *testing.T) {
		router := &mockRouterStrategy{name: "rpc"}
		collector := &mockOperationsCollector{
			specs: []api.OperationSpec{{Action: "create"}},
		}
		resolver := &mockHandlerResolver{handler: dummyHandler}
		eng, _ := NewEngine(
			WithRouters(router),
			WithOperationCollectors(collector),
			WithHandlerResolvers(resolver),
		)
		res := &mockResource{kind: api.KindRPC, name: "test/resource"}

		err := eng.Register(res)
		assert.NoError(t, err, "First registration should succeed")

		// Second registration of the same resource should fail
		// because Identifier value is used as map key (not pointer)
		err = eng.Register(res)
		assert.Error(t, err, "Duplicate registration should fail")
		assert.Contains(t, err.Error(), "duplicate", "Error should indicate duplicate")
	})

	t.Run("MultipleResources", func(t *testing.T) {
		router := &mockRouterStrategy{name: "rpc"}
		collector := &mockOperationsCollector{
			specs: []api.OperationSpec{{Action: "create"}},
		}
		resolver := &mockHandlerResolver{handler: dummyHandler}
		eng, _ := NewEngine(
			WithRouters(router),
			WithOperationCollectors(collector),
			WithHandlerResolvers(resolver),
		)
		res1 := &mockResource{kind: api.KindRPC, name: "test/resource1"}
		res2 := &mockResource{kind: api.KindRPC, name: "test/resource2"}

		err := eng.Register(res1, res2)
		assert.NoError(t, err, "Registering multiple resources should succeed")
	})

	t.Run("ResourceWithCustomVersion", func(t *testing.T) {
		router := &mockRouterStrategy{name: "rpc"}
		collector := &mockOperationsCollector{
			specs: []api.OperationSpec{{Action: "create"}},
		}
		resolver := &mockHandlerResolver{handler: dummyHandler}
		eng, _ := NewEngine(
			WithRouters(router),
			WithOperationCollectors(collector),
			WithHandlerResolvers(resolver),
		)
		res := &mockResource{kind: api.KindRPC, name: "test/resource", version: "v2"}

		err := eng.Register(res)
		assert.NoError(t, err, "Registration with custom version should succeed")

		op := eng.Lookup(api.Identifier{Resource: "test/resource", Action: "create", Version: "v2"})
		assert.NotNil(t, op, "Operation with custom version should be found")
	})

	t.Run("ResourceWithCustomAuth", func(t *testing.T) {
		router := &mockRouterStrategy{name: "rpc"}
		collector := &mockOperationsCollector{
			specs: []api.OperationSpec{{Action: "create"}},
		}
		resolver := &mockHandlerResolver{handler: dummyHandler}
		eng, _ := NewEngine(
			WithRouters(router),
			WithOperationCollectors(collector),
			WithHandlerResolvers(resolver),
		)
		customAuth := api.SignatureAuth()
		res := &mockResource{kind: api.KindRPC, name: "test/resource", auth: customAuth}

		err := eng.Register(res)
		assert.NoError(t, err, "Registration with custom auth should succeed")

		op := eng.Lookup(api.Identifier{Resource: "test/resource", Action: "create", Version: api.VersionV1})
		assert.NotNil(t, op, "Operation should be found")
		assert.Equal(t, api.AuthStrategySignature, op.Auth.Strategy, "Operation should use resource's custom auth")
	})

	t.Run("OperationWithPermToken", func(t *testing.T) {
		router := &mockRouterStrategy{name: "rpc"}
		collector := &mockOperationsCollector{
			specs: []api.OperationSpec{{Action: "delete", PermToken: "sys:user:delete"}},
		}
		resolver := &mockHandlerResolver{handler: dummyHandler}
		eng, _ := NewEngine(
			WithRouters(router),
			WithOperationCollectors(collector),
			WithHandlerResolvers(resolver),
		)
		res := &mockResource{kind: api.KindRPC, name: "sys/user"}

		err := eng.Register(res)
		assert.NoError(t, err, "Registration with perm token should succeed")

		op := eng.Lookup(api.Identifier{Resource: "sys/user", Action: "delete", Version: api.VersionV1})
		assert.NotNil(t, op, "Operation should be found")
		assert.NotNil(t, op.Auth.Options, "Auth options should not be nil")
		assert.Equal(t, "sys:user:delete", op.Auth.Options[common.AuthOptionPermToken], "PermToken should be stored correctly")
	})

	t.Run("OperationWithCustomTimeout", func(t *testing.T) {
		router := &mockRouterStrategy{name: "rpc"}
		collector := &mockOperationsCollector{
			specs: []api.OperationSpec{{Action: "export", Timeout: 2 * time.Minute}},
		}
		resolver := &mockHandlerResolver{handler: dummyHandler}
		eng, _ := NewEngine(
			WithRouters(router),
			WithOperationCollectors(collector),
			WithHandlerResolvers(resolver),
		)
		res := &mockResource{kind: api.KindRPC, name: "test/resource"}

		err := eng.Register(res)
		assert.NoError(t, err, "Registration with custom timeout should succeed")

		op := eng.Lookup(api.Identifier{Resource: "test/resource", Action: "export", Version: api.VersionV1})
		assert.NotNil(t, op, "Operation should be found")
		assert.Equal(t, 2*time.Minute, op.Timeout, "Operation should have custom timeout")
	})

	t.Run("OperationWithCustomRateLimit", func(t *testing.T) {
		router := &mockRouterStrategy{name: "rpc"}
		customRateLimit := &api.RateLimitConfig{Max: 10, Period: 1 * time.Minute}
		collector := &mockOperationsCollector{
			specs: []api.OperationSpec{{Action: "login", RateLimit: customRateLimit}},
		}
		resolver := &mockHandlerResolver{handler: dummyHandler}
		eng, _ := NewEngine(
			WithRouters(router),
			WithOperationCollectors(collector),
			WithHandlerResolvers(resolver),
		)
		res := &mockResource{kind: api.KindRPC, name: "auth"}

		err := eng.Register(res)
		assert.NoError(t, err, "Registration with custom rate limit should succeed")

		op := eng.Lookup(api.Identifier{Resource: "auth", Action: "login", Version: api.VersionV1})
		assert.NotNil(t, op, "Operation should be found")
		assert.Equal(t, 10, op.RateLimit.Max, "Operation should have custom rate limit max")
	})

	t.Run("OperationWithEnableAudit", func(t *testing.T) {
		router := &mockRouterStrategy{name: "rpc"}
		collector := &mockOperationsCollector{
			specs: []api.OperationSpec{{Action: "update", EnableAudit: true}},
		}
		resolver := &mockHandlerResolver{handler: dummyHandler}
		eng, _ := NewEngine(
			WithRouters(router),
			WithOperationCollectors(collector),
			WithHandlerResolvers(resolver),
		)
		res := &mockResource{kind: api.KindRPC, name: "test/resource"}

		err := eng.Register(res)
		assert.NoError(t, err, "Registration with audit enabled should succeed")

		op := eng.Lookup(api.Identifier{Resource: "test/resource", Action: "update", Version: api.VersionV1})
		assert.NotNil(t, op, "Operation should be found")
		assert.True(t, op.EnableAudit, "Operation should have audit enabled")
	})
}

func TestEngineLookup(t *testing.T) {
	t.Log("Testing Engine.Lookup method")

	t.Run("NotFound", func(t *testing.T) {
		eng, _ := NewEngine()
		op := eng.Lookup(api.Identifier{Resource: "nonexistent", Action: "create", Version: api.VersionV1})
		assert.Nil(t, op, "Lookup should return nil for non-existent operation")
	})

	t.Run("Found", func(t *testing.T) {
		router := &mockRouterStrategy{name: "rpc"}
		collector := &mockOperationsCollector{
			specs: []api.OperationSpec{{Action: "create"}},
		}
		resolver := &mockHandlerResolver{handler: dummyHandler}
		eng, _ := NewEngine(
			WithRouters(router),
			WithOperationCollectors(collector),
			WithHandlerResolvers(resolver),
		)
		res := &mockResource{kind: api.KindRPC, name: "test/resource"}
		_ = eng.Register(res)

		op := eng.Lookup(api.Identifier{Resource: "test/resource", Action: "create", Version: api.VersionV1})
		assert.NotNil(t, op, "Lookup should return operation")
		assert.Equal(t, "test/resource", op.Resource, "Operation resource should match")
		assert.Equal(t, "create", op.Action, "Operation action should match")
		assert.Equal(t, api.VersionV1, op.Version, "Operation version should match")
	})

	t.Run("VersionMismatch", func(t *testing.T) {
		router := &mockRouterStrategy{name: "rpc"}
		collector := &mockOperationsCollector{
			specs: []api.OperationSpec{{Action: "create"}},
		}
		resolver := &mockHandlerResolver{handler: dummyHandler}
		eng, _ := NewEngine(
			WithRouters(router),
			WithOperationCollectors(collector),
			WithHandlerResolvers(resolver),
		)
		res := &mockResource{kind: api.KindRPC, name: "test/resource"}
		_ = eng.Register(res)

		op := eng.Lookup(api.Identifier{Resource: "test/resource", Action: "create", Version: "v2"})
		assert.Nil(t, op, "Lookup should return nil for version mismatch")
	})

	t.Run("ActionMismatch", func(t *testing.T) {
		router := &mockRouterStrategy{name: "rpc"}
		collector := &mockOperationsCollector{
			specs: []api.OperationSpec{{Action: "create"}},
		}
		resolver := &mockHandlerResolver{handler: dummyHandler}
		eng, _ := NewEngine(
			WithRouters(router),
			WithOperationCollectors(collector),
			WithHandlerResolvers(resolver),
		)
		res := &mockResource{kind: api.KindRPC, name: "test/resource"}
		_ = eng.Register(res)

		op := eng.Lookup(api.Identifier{Resource: "test/resource", Action: "update", Version: api.VersionV1})
		assert.Nil(t, op, "Lookup should return nil for action mismatch")
	})

	t.Run("MultipleOperations", func(t *testing.T) {
		router := &mockRouterStrategy{name: "rpc"}
		collector := &mockOperationsCollector{
			specs: []api.OperationSpec{
				{Action: "create"},
				{Action: "update"},
				{Action: "delete"},
			},
		}
		resolver := &mockHandlerResolver{handler: dummyHandler}
		eng, _ := NewEngine(
			WithRouters(router),
			WithOperationCollectors(collector),
			WithHandlerResolvers(resolver),
		)
		res := &mockResource{kind: api.KindRPC, name: "test/resource"}
		_ = eng.Register(res)

		for _, action := range []string{"create", "update", "delete"} {
			op := eng.Lookup(api.Identifier{Resource: "test/resource", Action: action, Version: api.VersionV1})
			assert.NotNil(t, op, "Lookup should return operation for action %s", action)
		}
	})
}

func TestEngineMount(t *testing.T) {
	t.Log("Testing Engine.Mount method")

	t.Run("SetupError", func(t *testing.T) {
		router := &mockRouterStrategy{name: "rpc", setupErr: errors.New("setup failed")}
		eng, _ := NewEngine(WithRouters(router))

		app := fiber.New()
		defer app.Shutdown()

		err := eng.Mount(app)
		assert.Error(t, err, "Mount should return error when setup fails")
		assert.Contains(t, err.Error(), "setup", "Error message should mention setup")
	})

	t.Run("SuccessfulMount", func(t *testing.T) {
		router := &mockRouterStrategy{name: "rpc"}
		collector := &mockOperationsCollector{
			specs: []api.OperationSpec{{Action: "create"}},
		}
		resolver := &mockHandlerResolver{handler: dummyHandler}
		adapter := &mockHandlerAdapter{}
		eng, _ := NewEngine(
			WithRouters(router),
			WithOperationCollectors(collector),
			WithHandlerResolvers(resolver),
			WithHandlerAdapters(adapter),
		)
		res := &mockResource{kind: api.KindRPC, name: "test/resource"}
		_ = eng.Register(res)

		app := fiber.New()
		defer app.Shutdown()

		err := eng.Mount(app)
		assert.NoError(t, err, "Mount should succeed")
		assert.Equal(t, 1, router.setupCalls, "Setup should be called once")
		assert.Equal(t, 1, router.routeCalls, "Route should be called once")
	})

	t.Run("MultipleRouters", func(t *testing.T) {
		rpcRouter := &mockRouterStrategy{name: "rpc"}
		restRouter := &mockRouterStrategy{name: "rest"}
		collector := &mockOperationsCollector{
			specs: []api.OperationSpec{{Action: "create"}},
		}
		resolver := &mockHandlerResolver{handler: dummyHandler}
		adapter := &mockHandlerAdapter{}
		eng, _ := NewEngine(
			WithRouters(rpcRouter, restRouter),
			WithOperationCollectors(collector),
			WithHandlerResolvers(resolver),
			WithHandlerAdapters(adapter),
		)
		res1 := &mockResource{kind: api.KindRPC, name: "test/rpc"}
		res2 := &mockResource{kind: api.KindREST, name: "test/rest"}
		_ = eng.Register(res1)
		_ = eng.Register(res2)

		app := fiber.New()
		defer app.Shutdown()

		err := eng.Mount(app)
		assert.NoError(t, err, "Mount should succeed with multiple routers")
		assert.Equal(t, 1, rpcRouter.setupCalls, "RPC router setup should be called once")
		assert.Equal(t, 1, restRouter.setupCalls, "REST router setup should be called once")
	})

	t.Run("AdapterError", func(t *testing.T) {
		router := &mockRouterStrategy{name: "rpc"}
		collector := &mockOperationsCollector{
			specs: []api.OperationSpec{{Action: "create"}},
		}
		resolver := &mockHandlerResolver{handler: dummyHandler}
		adapter := &mockHandlerAdapter{err: errors.New("adapt failed")}
		eng, _ := NewEngine(
			WithRouters(router),
			WithOperationCollectors(collector),
			WithHandlerResolvers(resolver),
			WithHandlerAdapters(adapter),
		)
		res := &mockResource{kind: api.KindRPC, name: "test/resource"}
		_ = eng.Register(res)

		app := fiber.New()
		defer app.Shutdown()

		err := eng.Mount(app)
		assert.Error(t, err, "Mount should return error when adapter fails")
		assert.Contains(t, err.Error(), "adapt", "Error message should mention adapt")
	})

	t.Run("OperationNotFoundInOperationsMap", func(t *testing.T) {
		router := &mockRouterStrategy{name: "rpc"}
		eng, _ := NewEngine(WithRouters(router))
		e := eng.(*engine)

		// Manually add an identifier to routerOperations without adding to operations
		identifier := api.Identifier{Resource: "ghost/resource", Action: "create", Version: api.VersionV1}
		if ops, ok := e.routerOperations.Get(router); ok {
			ops.Add(identifier)
		}

		app := fiber.New()
		defer app.Shutdown()

		err := eng.Mount(app)
		assert.Error(t, err, "Mount should return error when operation not found")
		assert.Contains(t, err.Error(), "not found", "Error message should mention not found")
	})
}

func TestResolveTimeout(t *testing.T) {
	t.Log("Testing resolveTimeout method")

	t.Run("ZeroUsesDefault", func(t *testing.T) {
		eng, _ := NewEngine(WithDefaultTimeout(30 * time.Second))
		e := eng.(*engine)
		result := e.resolveTimeout(0)
		assert.Equal(t, 30*time.Second, result, "Zero timeout should use default")
	})

	t.Run("CustomTimeoutUsed", func(t *testing.T) {
		eng, _ := NewEngine(WithDefaultTimeout(30 * time.Second))
		e := eng.(*engine)
		result := e.resolveTimeout(60 * time.Second)
		assert.Equal(t, 60*time.Second, result, "Custom timeout should be used")
	})

	t.Run("NegativeUsesDefault", func(t *testing.T) {
		eng, _ := NewEngine(WithDefaultTimeout(30 * time.Second))
		e := eng.(*engine)
		result := e.resolveTimeout(-1 * time.Second)
		assert.Equal(t, 30*time.Second, result, "Negative timeout should use default")
	})
}

func TestResolveRateLimit(t *testing.T) {
	t.Log("Testing resolveRateLimit method")

	t.Run("NilUsesDefault", func(t *testing.T) {
		defaultLimit := &api.RateLimitConfig{Max: 100, Period: 5 * time.Minute}
		eng, _ := NewEngine(WithDefaultRateLimit(defaultLimit))
		e := eng.(*engine)
		result := e.resolveRateLimit(nil)
		assert.Equal(t, defaultLimit, result, "Nil rate limit should use default")
	})

	t.Run("CustomRateLimitUsed", func(t *testing.T) {
		defaultLimit := &api.RateLimitConfig{Max: 100, Period: 5 * time.Minute}
		customLimit := &api.RateLimitConfig{Max: 10, Period: 1 * time.Minute}
		eng, _ := NewEngine(WithDefaultRateLimit(defaultLimit))
		e := eng.(*engine)
		result := e.resolveRateLimit(customLimit)
		assert.Equal(t, customLimit, result, "Custom rate limit should be used")
	})
}

func TestFindRouterStrategy(t *testing.T) {
	t.Log("Testing findRouterStrategy method")

	t.Run("FindRPCRouter", func(t *testing.T) {
		rpcRouter := &mockRouterStrategy{name: "rpc"}
		restRouter := &mockRouterStrategy{name: "rest"}
		eng, _ := NewEngine(WithRouters(rpcRouter, restRouter))
		e := eng.(*engine)
		result := e.findRouterStrategy(api.KindRPC)
		assert.NotNil(t, result, "Should find RPC router")
		assert.Equal(t, "rpc", result.Name(), "Router name should be rpc")
	})

	t.Run("FindRESTRouter", func(t *testing.T) {
		rpcRouter := &mockRouterStrategy{name: "rpc"}
		restRouter := &mockRouterStrategy{name: "rest"}
		eng, _ := NewEngine(WithRouters(rpcRouter, restRouter))
		e := eng.(*engine)
		result := e.findRouterStrategy(api.KindREST)
		assert.NotNil(t, result, "Should find REST router")
		assert.Equal(t, "rest", result.Name(), "Router name should be rest")
	})

	t.Run("NotFound", func(t *testing.T) {
		rpcRouter := &mockRouterStrategy{name: "rpc"}
		eng, _ := NewEngine(WithRouters(rpcRouter))
		e := eng.(*engine)
		result := e.findRouterStrategy(api.KindREST)
		assert.Nil(t, result, "Should return nil for unknown kind")
	})
}

func TestAdaptHandler(t *testing.T) {
	t.Log("Testing adaptHandler method")

	t.Run("FirstAdapterSucceeds", func(t *testing.T) {
		adapter1 := &mockHandlerAdapter{}
		adapter2 := &mockHandlerAdapter{}
		eng, _ := NewEngine(WithHandlerAdapters(adapter1, adapter2))
		e := eng.(*engine)
		op := &api.Operation{Handler: dummyHandler}
		result, err := e.adaptHandler(op)
		assert.NoError(t, err, "adaptHandler should not return error")
		assert.NotNil(t, result, "adaptHandler should return handler")
	})

	t.Run("FirstAdapterReturnsNilSecondSucceeds", func(t *testing.T) {
		adapter1 := &mockHandlerAdapter{returnNil: true} // Returns nil to trigger chain
		adapter2 := &mockHandlerAdapter{}
		eng, _ := NewEngine(WithHandlerAdapters(adapter1, adapter2))
		e := eng.(*engine)
		op := &api.Operation{Handler: dummyHandler}
		result, err := e.adaptHandler(op)
		assert.NoError(t, err, "adaptHandler should not return error")
		assert.NotNil(t, result, "adaptHandler should return handler from second adapter")
	})

	t.Run("NoAdaptersConfigured", func(t *testing.T) {
		eng, _ := NewEngine() // No adapters
		e := eng.(*engine)
		op := &api.Operation{Handler: dummyHandler}
		result, err := e.adaptHandler(op)
		assert.Error(t, err, "adaptHandler should return error when no adapter handles")
		assert.Nil(t, result, "adaptHandler should return nil handler")
		assert.ErrorIs(t, err, ErrNoHandlerAdapter, "Error should be ErrNoHandlerAdapter")
	})

	t.Run("AdapterError", func(t *testing.T) {
		adapter := &mockHandlerAdapter{err: errors.New("adapt failed")}
		eng, _ := NewEngine(WithHandlerAdapters(adapter))
		e := eng.(*engine)
		op := &api.Operation{Handler: dummyHandler}
		result, err := e.adaptHandler(op)
		assert.Error(t, err, "adaptHandler should return error")
		assert.Nil(t, result, "adaptHandler should return nil handler")
	})
}

func TestWrapHandlerIfNecessary(t *testing.T) {
	t.Log("Testing wrapHandlerIfNecessary method")

	t.Run("NoWrapWhenZeroTimeout", func(t *testing.T) {
		eng, _ := NewEngine()
		e := eng.(*engine)
		op := &api.Operation{Timeout: 0}
		handler := func(fiber.Ctx) error { return nil }
		result := e.wrapHandlerIfNecessary(handler, op)
		// Cannot directly compare functions, but we can verify it doesn't panic
		assert.NotNil(t, result, "Handler should not be nil")
	})

	t.Run("WrapsWhenPositiveTimeout", func(t *testing.T) {
		eng, _ := NewEngine()
		e := eng.(*engine)
		op := &api.Operation{Timeout: 30 * time.Second}
		handler := func(fiber.Ctx) error { return nil }
		result := e.wrapHandlerIfNecessary(handler, op)
		assert.NotNil(t, result, "Handler should not be nil")
	})

	t.Run("NoWrapWhenNegativeTimeout", func(t *testing.T) {
		eng, _ := NewEngine()
		e := eng.(*engine)
		op := &api.Operation{Timeout: -1 * time.Second}
		handler := func(fiber.Ctx) error { return nil }
		result := e.wrapHandlerIfNecessary(handler, op)
		assert.NotNil(t, result, "Handler should not be nil")
	})

	t.Run("TimeoutReturnsErrRequestTimeout", func(t *testing.T) {
		eng, _ := NewEngine()
		e := eng.(*engine)
		op := &api.Operation{Timeout: 10 * time.Millisecond}
		slowHandler := func(fiber.Ctx) error {
			time.Sleep(100 * time.Millisecond)

			return nil
		}
		wrappedHandler := e.wrapHandlerIfNecessary(slowHandler, op)

		app := fiber.New()
		defer app.Shutdown()

		app.Get("/test", wrappedHandler)

		req := httptest.NewRequest("GET", "/test", nil)
		resp, err := app.Test(req, fiber.TestConfig{Timeout: 5 * time.Second})
		assert.NoError(t, err, "Test request should not error")
		// Fiber timeout middleware returns 500 by default when OnTimeout returns an error
		// The actual status code depends on error handler configuration
		assert.True(t, resp.StatusCode >= 400, "Should return error status code on timeout")
	})
}

func TestResolveAuthConfig(t *testing.T) {
	t.Log("Testing resolveAuthConfig method")

	t.Run("PublicSpecReturnsPublicAuth", func(t *testing.T) {
		eng, _ := NewEngine(WithDefaultAuth(api.BearerAuth()))
		e := eng.(*engine)
		res := &mockResource{kind: api.KindRPC, name: "test/resource"}
		spec := api.OperationSpec{Action: "ping", Public: true}

		result := e.resolveAuthConfig(res, spec)
		assert.NotNil(t, result, "Auth config should not be nil")
		assert.Equal(t, api.AuthStrategyNone, result.Strategy, "Public spec should return public auth (none strategy)")
	})

	t.Run("ResourceAuthOverridesDefault", func(t *testing.T) {
		defaultAuth := api.BearerAuth()
		resourceAuth := api.SignatureAuth()
		eng, _ := NewEngine(WithDefaultAuth(defaultAuth))
		e := eng.(*engine)
		res := &mockResource{kind: api.KindRPC, name: "test/resource", auth: resourceAuth}
		spec := api.OperationSpec{Action: "create"}

		result := e.resolveAuthConfig(res, spec)
		assert.NotNil(t, result, "Auth config should not be nil")
		assert.Equal(t, api.AuthStrategySignature, result.Strategy, "Resource auth should override default")
	})

	t.Run("DefaultAuthUsedWhenNoResourceAuth", func(t *testing.T) {
		defaultAuth := api.BearerAuth()
		eng, _ := NewEngine(WithDefaultAuth(defaultAuth))
		e := eng.(*engine)
		res := &mockResource{kind: api.KindRPC, name: "test/resource"}
		spec := api.OperationSpec{Action: "create"}

		result := e.resolveAuthConfig(res, spec)
		assert.NotNil(t, result, "Auth config should not be nil")
		assert.Equal(t, api.AuthStrategyBearer, result.Strategy, "Default auth should be used")
	})

	t.Run("AuthConfigIsCloned", func(t *testing.T) {
		defaultAuth := api.BearerAuth()
		eng, _ := NewEngine(WithDefaultAuth(defaultAuth))
		e := eng.(*engine)
		res := &mockResource{kind: api.KindRPC, name: "test/resource"}
		spec := api.OperationSpec{Action: "create"}

		result := e.resolveAuthConfig(res, spec)
		result.Strategy = "modified"

		assert.Equal(t, api.AuthStrategyBearer, defaultAuth.Strategy, "Original auth should not be modified")
	})

	t.Run("ResourceAuthConfigIsCloned", func(t *testing.T) {
		resourceAuth := api.SignatureAuth()
		eng, _ := NewEngine(WithDefaultAuth(api.BearerAuth()))
		e := eng.(*engine)
		res := &mockResource{kind: api.KindRPC, name: "test/resource", auth: resourceAuth}
		spec := api.OperationSpec{Action: "create"}

		result := e.resolveAuthConfig(res, spec)
		result.Strategy = "modified"

		assert.Equal(t, api.AuthStrategySignature, resourceAuth.Strategy, "Original resource auth should not be modified")
	})
}

func TestResolveHandler(t *testing.T) {
	t.Log("Testing resolveHandler method")

	t.Run("FirstResolverSucceeds", func(t *testing.T) {
		resolver1 := &mockHandlerResolver{handler: dummyHandler}
		resolver2 := &mockHandlerResolver{handler: dummyHandler}
		eng, _ := NewEngine(WithHandlerResolvers(resolver1, resolver2))
		e := eng.(*engine)
		res := &mockResource{kind: api.KindRPC, name: "test/resource"}
		spec := api.OperationSpec{Action: "create"}

		result, err := e.resolveHandler(spec, res)
		assert.NoError(t, err, "resolveHandler should not return error")
		assert.NotNil(t, result, "resolveHandler should return handler")
	})

	t.Run("FirstResolverReturnsNilSecondSucceeds", func(t *testing.T) {
		resolver1 := &mockHandlerResolver{returnNil: true}
		resolver2 := &mockHandlerResolver{handler: dummyHandler}
		eng, _ := NewEngine(WithHandlerResolvers(resolver1, resolver2))
		e := eng.(*engine)
		res := &mockResource{kind: api.KindRPC, name: "test/resource"}
		spec := api.OperationSpec{Action: "create"}

		result, err := e.resolveHandler(spec, res)
		assert.NoError(t, err, "resolveHandler should not return error")
		assert.NotNil(t, result, "resolveHandler should return handler from second resolver")
	})

	t.Run("AllResolversReturnNil", func(t *testing.T) {
		resolver1 := &mockHandlerResolver{returnNil: true}
		resolver2 := &mockHandlerResolver{returnNil: true}
		eng, _ := NewEngine(WithHandlerResolvers(resolver1, resolver2))
		e := eng.(*engine)
		res := &mockResource{kind: api.KindRPC, name: "test/resource"}
		spec := api.OperationSpec{Action: "create"}

		result, err := e.resolveHandler(spec, res)
		assert.Error(t, err, "resolveHandler should return error when all resolvers return nil")
		assert.Nil(t, result, "resolveHandler should return nil handler")
		assert.Contains(t, err.Error(), "no handler resolver found", "Error should mention no resolver found")
	})

	t.Run("ResolverError", func(t *testing.T) {
		resolver := &mockHandlerResolver{err: errors.New("resolve failed")}
		eng, _ := NewEngine(WithHandlerResolvers(resolver))
		e := eng.(*engine)
		res := &mockResource{kind: api.KindRPC, name: "test/resource"}
		spec := api.OperationSpec{Action: "create"}

		result, err := e.resolveHandler(spec, res)
		assert.Error(t, err, "resolveHandler should return error")
		assert.Nil(t, result, "resolveHandler should return nil handler")
	})
}

func TestRegisterResource(t *testing.T) {
	t.Log("Testing registerResource method")

	t.Run("NoCollectorsReturnsSuccess", func(t *testing.T) {
		router := &mockRouterStrategy{name: "rpc"}
		eng, _ := NewEngine(WithRouters(router))
		e := eng.(*engine)
		res := &mockResource{kind: api.KindRPC, name: "test/resource"}

		err := e.registerResource(res)
		assert.NoError(t, err, "registerResource with no collectors should succeed")
	})

	t.Run("CollectorReturnsEmptySpecs", func(t *testing.T) {
		router := &mockRouterStrategy{name: "rpc"}
		collector := &mockOperationsCollector{specs: []api.OperationSpec{}}
		eng, _ := NewEngine(WithRouters(router), WithOperationCollectors(collector))
		e := eng.(*engine)
		res := &mockResource{kind: api.KindRPC, name: "test/resource"}

		err := e.registerResource(res)
		assert.NoError(t, err, "registerResource with empty specs should succeed")
	})
}

func TestRegisterAfterMount(t *testing.T) {
	t.Log("Testing Register after Mount")

	t.Run("RegisterAfterMountRoutesImmediately", func(t *testing.T) {
		router := &mockRouterStrategy{name: "rpc"}
		collector := &mockOperationsCollector{
			specs: []api.OperationSpec{{Action: "create"}},
		}
		resolver := &mockHandlerResolver{handler: dummyHandler}
		adapter := &mockHandlerAdapter{}
		eng, _ := NewEngine(
			WithRouters(router),
			WithOperationCollectors(collector),
			WithHandlerResolvers(resolver),
			WithHandlerAdapters(adapter),
		)

		app := fiber.New()
		defer app.Shutdown()

		err := eng.Mount(app)
		assert.NoError(t, err, "Mount should succeed")
		assert.Equal(t, 0, router.routeCalls, "No routes before registration")

		res := &mockResource{kind: api.KindRPC, name: "test/resource"}
		err = eng.Register(res)
		assert.NoError(t, err, "Register after Mount should succeed")
		assert.Equal(t, 1, router.routeCalls, "Route should be called immediately after registration")
	})

	t.Run("RegisterAfterMountAdapterError", func(t *testing.T) {
		router := &mockRouterStrategy{name: "rpc"}
		collector := &mockOperationsCollector{
			specs: []api.OperationSpec{{Action: "create"}},
		}
		resolver := &mockHandlerResolver{handler: dummyHandler}
		adapter := &mockHandlerAdapter{err: errors.New("adapt failed")}
		eng, _ := NewEngine(
			WithRouters(router),
			WithOperationCollectors(collector),
			WithHandlerResolvers(resolver),
			WithHandlerAdapters(adapter),
		)

		app := fiber.New()
		defer app.Shutdown()

		err := eng.Mount(app)
		assert.NoError(t, err, "Mount should succeed with no operations")

		res := &mockResource{kind: api.KindRPC, name: "test/resource"}
		err = eng.Register(res)
		assert.Error(t, err, "Register after Mount should fail when adapter fails")
		assert.Contains(t, err.Error(), "adapt", "Error should mention adapt")
	})
}

func TestAdaptHandlerAllReturnNil(t *testing.T) {
	t.Log("Testing adaptHandler when all adapters return nil")

	t.Run("SingleAdapterReturnsNil", func(t *testing.T) {
		adapter := &mockHandlerAdapter{returnNil: true}
		eng, _ := NewEngine(WithHandlerAdapters(adapter))
		e := eng.(*engine)
		op := &api.Operation{Handler: dummyHandler}
		result, err := e.adaptHandler(op)
		assert.Error(t, err, "adaptHandler should return error when adapter returns nil")
		assert.Nil(t, result, "adaptHandler should return nil handler")
		assert.ErrorIs(t, err, ErrNoHandlerAdapter, "Error should be ErrNoHandlerAdapter")
	})

	t.Run("MultipleAdaptersAllReturnNil", func(t *testing.T) {
		adapter1 := &mockHandlerAdapter{returnNil: true}
		adapter2 := &mockHandlerAdapter{returnNil: true}
		eng, _ := NewEngine(WithHandlerAdapters(adapter1, adapter2))
		e := eng.(*engine)
		op := &api.Operation{Handler: dummyHandler}
		result, err := e.adaptHandler(op)
		assert.Error(t, err, "adaptHandler should return error when all adapters return nil")
		assert.Nil(t, result, "adaptHandler should return nil handler")
		assert.ErrorIs(t, err, ErrNoHandlerAdapter, "Error should be ErrNoHandlerAdapter")
	})
}
