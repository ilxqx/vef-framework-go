package api

import (
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/timeout"
	"github.com/ilxqx/go-collections"
	"github.com/ilxqx/go-streams"
	"github.com/samber/lo"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/internal/api/common"
	"github.com/ilxqx/vef-framework-go/internal/log"
	"github.com/ilxqx/vef-framework-go/result"
)

var (
	ErrNoHandlerAdapter = errors.New("no handler adapter found")
	logger              = log.Named("api")
)

type EngineOption func(*engine)

// engine implements api.Engine.
type engine struct {
	defaultVersion   string
	defaultTimeout   time.Duration
	defaultAuth      *api.AuthConfig
	defaultRateLimit *api.RateLimitConfig

	operations       collections.ConcurrentMap[api.Identifier, *api.Operation]
	routerOperations collections.Map[api.RouterStrategy, collections.ConcurrentSet[api.Identifier]]
	collectors       []api.OperationsCollector
	resolvers        []api.HandlerResolver
	adapters         []api.HandlerAdapter
	router           fiber.Router
}

func WithDefaultTimeout(timeout time.Duration) EngineOption {
	return func(e *engine) {
		e.defaultTimeout = timeout
	}
}

func WithDefaultVersion(version string) EngineOption {
	return func(e *engine) {
		e.defaultVersion = version
	}
}

func WithDefaultAuth(auth *api.AuthConfig) EngineOption {
	return func(e *engine) {
		e.defaultAuth = auth
	}
}

func WithDefaultRateLimit(rateLimit *api.RateLimitConfig) EngineOption {
	return func(e *engine) {
		e.defaultRateLimit = rateLimit
	}
}

func WithRouters(routers ...api.RouterStrategy) EngineOption {
	return func(e *engine) {
		e.routerOperations = streams.CollectTo(
			streams.FromSlice(routers),
			streams.ToHashMapCollector(
				func(router api.RouterStrategy) api.RouterStrategy { return router },
				func(api.RouterStrategy) collections.ConcurrentSet[api.Identifier] {
					return collections.NewConcurrentHashSet[api.Identifier]()
				},
			),
		)
	}
}

func WithOperationCollectors(collectors ...api.OperationsCollector) EngineOption {
	return func(e *engine) {
		e.collectors = collectors
	}
}

func WithHandlerResolvers(resolvers ...api.HandlerResolver) EngineOption {
	return func(e *engine) {
		e.resolvers = resolvers
	}
}

func WithHandlerAdapters(adapters ...api.HandlerAdapter) EngineOption {
	return func(e *engine) {
		e.adapters = adapters
	}
}

// NewEngine creates a new API engine with the given options.
func NewEngine(opts ...EngineOption) (api.Engine, error) {
	eng := &engine{
		defaultTimeout: 30 * time.Second,
		defaultVersion: api.VersionV1,
		defaultAuth:    api.BearerAuth(),
		defaultRateLimit: &api.RateLimitConfig{
			Max:    100,
			Period: 5 * time.Minute,
		},
		operations: collections.NewConcurrentHashMap[api.Identifier, *api.Operation](),
	}

	for _, opt := range opts {
		opt(eng)
	}

	return eng, nil
}

// Register adds resources to the engine.
func (e *engine) Register(resources ...api.Resource) error {
	for _, res := range resources {
		if err := e.registerResource(res); err != nil {
			return err
		}
	}

	return nil
}

// Mount attaches the engine to a Fiber router.
func (e *engine) Mount(router fiber.Router) error {
	e.router = router
	for rs := range e.routerOperations.SeqKeys() {
		if err := rs.Setup(router); err != nil {
			return fmt.Errorf("failed to setup router %s: %w", rs.Name(), err)
		}
	}

	for rs, ops := range e.routerOperations.Seq() {
		for identifier := range ops.Seq() {
			op, ok := e.operations.Get(identifier)
			if !ok {
				return fmt.Errorf("operation %s not found", identifier)
			}

			handler, err := e.adaptHandler(op)
			if err != nil {
				return fmt.Errorf("failed to adapt handler for %s: %w", identifier, err)
			}

			wh := e.wrapHandlerIfNecessary(handler, op)
			rs.Route(wh, op)
		}
	}

	return nil
}

// Lookup finds an operation by identifier.
func (e *engine) Lookup(identifier api.Identifier) *api.Operation {
	op, _ := e.operations.Get(identifier)

	return op
}

// registerResource registers a single resource.
func (e *engine) registerResource(res api.Resource) error {
	if res == nil {
		return errors.New("resource cannot be nil")
	}

	resourceName := res.Name()
	if resourceName == constants.Empty {
		return errors.New("resource name cannot be empty")
	}

	for _, collector := range e.collectors {
		specs := collector.Collect(res)

		for _, spec := range specs {
			if err := e.registerOperation(res, spec); err != nil {
				return err
			}
		}
	}

	return nil
}

func (e *engine) registerOperation(res api.Resource, spec api.OperationSpec) error {
	if spec.Action == constants.Empty {
		return fmt.Errorf("operation action cannot be empty for resource %s", res.Name())
	}

	resName := res.Name()
	resVersion := lo.CoalesceOrEmpty(res.Version(), e.defaultVersion, api.VersionV1)

	ac := e.resolveAuthConfig(res, spec)
	if spec.PermToken != constants.Empty {
		if ac.Options == nil {
			ac.Options = make(map[string]any)
		}

		ac.Options[common.AuthOptionPermToken] = spec.PermToken
	}

	rs := e.findRouterStrategy(res.Kind())
	if rs == nil {
		return fmt.Errorf("no router can handle operation type %s", res.Kind())
	}

	h, err := e.resolveHandler(spec, res)
	if err != nil {
		return fmt.Errorf("failed to resolve handler for %s:%s: %w", resName, spec.Action, err)
	}

	op := &api.Operation{
		Identifier: api.Identifier{
			Resource: resName,
			Action:   spec.Action,
			Version:  resVersion,
		},
		Auth:        ac,
		Timeout:     e.resolveTimeout(spec.Timeout),
		RateLimit:   e.resolveRateLimit(spec.RateLimit),
		EnableAudit: spec.EnableAudit,
		Meta: map[string]any{
			common.MetaKeyResource: res,
		},
		Handler: h,
	}

	if existing, inserted := e.operations.PutIfAbsent(op.Identifier, op); !inserted {
		return &common.DuplicateError{
			BaseError: common.BaseError{
				Identifier: &op.Identifier,
			},
			Existing: existing,
		}
	}

	operations, ok := e.routerOperations.Get(rs)
	if !ok {
		return fmt.Errorf("no router found for %s", rs.Name())
	}

	operations.AddIfAbsent(op.Identifier)

	if e.router != nil {
		handler, err := e.adaptHandler(op)
		if err != nil {
			return fmt.Errorf("failed to adapt handler for %s: %w", op.Identifier, err)
		}

		rs.Route(e.wrapHandlerIfNecessary(handler, op), op)
	}

	logger.Infof("Registered %s operation: resource=%s, action=%s, version=%s, type=%s, auth=%s, audit=%v",
		rs.Name(),
		op.Resource, op.Action, op.Version,
		reflect.TypeOf(res).String(), op.Auth.Strategy, op.EnableAudit)

	return nil
}

func (e *engine) resolveAuthConfig(res api.Resource, spec api.OperationSpec) *api.AuthConfig {
	if spec.Public {
		return api.Public()
	}

	if res.Auth() != nil {
		return res.Auth().Clone()
	}

	return e.defaultAuth.Clone()
}

// findRouterStrategy finds a router that can handle the given operation spec.
func (e *engine) findRouterStrategy(kind api.Kind) api.RouterStrategy {
	for router := range e.routerOperations.SeqKeys() {
		if router.Name() == kind.String() {
			return router
		}
	}

	return nil
}

// resolveHandler resolves the handler from spec or resource.
func (e *engine) resolveHandler(spec api.OperationSpec, res api.Resource) (any, error) {
	for _, resolver := range e.resolvers {
		handler, err := resolver.Resolve(res, spec)
		if err != nil {
			return nil, err
		}

		if handler != nil {
			return handler, nil
		}
	}

	return nil, fmt.Errorf("no handler resolver found for %s:%s", spec.Action, res.Name())
}

// resolveTimeout returns operation timeout or default.
func (e *engine) resolveTimeout(t time.Duration) time.Duration {
	if t > 0 {
		return t
	}

	return e.defaultTimeout
}

// resolveRateLimit returns operation rate limit or default.
func (e *engine) resolveRateLimit(limit *api.RateLimitConfig) *api.RateLimitConfig {
	if limit != nil {
		return limit
	}

	return e.defaultRateLimit
}

// adaptHandler uses the adapter chain to convert the handler.
func (e *engine) adaptHandler(op *api.Operation) (fiber.Handler, error) {
	for _, adapter := range e.adapters {
		handler, err := adapter.Adapt(op.Handler, op)
		if err != nil {
			return nil, err
		}

		if handler != nil {
			return handler, nil
		}
	}

	return nil, fmt.Errorf("%w: %T", ErrNoHandlerAdapter, op.Handler)
}

func (e *engine) wrapHandlerIfNecessary(handler fiber.Handler, op *api.Operation) fiber.Handler {
	if op.Timeout <= 0 {
		return handler
	}

	return timeout.New(handler, timeout.Config{
		Timeout: op.Timeout,
		OnTimeout: func(fiber.Ctx) error {
			return result.ErrRequestTimeout
		},
	})
}
