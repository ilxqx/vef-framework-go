package router

import (
	"encoding/json"
	"fmt"

	"github.com/gofiber/fiber/v3"
	"github.com/hbollon/go-edlib"
	"github.com/ilxqx/go-collections"
	"github.com/ilxqx/go-streams"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/contextx"
	"github.com/ilxqx/vef-framework-go/i18n"
	"github.com/ilxqx/vef-framework-go/internal/api/middleware"
	"github.com/ilxqx/vef-framework-go/internal/api/shared"
	"github.com/ilxqx/vef-framework-go/result"
	"github.com/ilxqx/vef-framework-go/webhelpers"
)

const (
	DefaultRPCEndpoint = "/api"
	FormKeyParams      = "params"
	FormKeyMeta        = "meta"
)

// RPC implements api.RouterStrategy for RPC-style single endpoint routing.
type RPC struct {
	path       string
	chain      *middleware.Chain
	operations collections.ConcurrentMap[api.Identifier, *routeEntry]
}

type routeEntry struct {
	op      *api.Operation
	handler fiber.Handler
}

// NewRPC creates a new RPC-style router.
// If path is empty, defaults to "/api".
func NewRPC(path string, chain *middleware.Chain) api.RouterStrategy {
	if path == constants.Empty {
		path = DefaultRPCEndpoint
	}

	rs := &RPC{
		path:       path,
		chain:      chain,
		operations: collections.NewConcurrentHashMap[api.Identifier, *routeEntry](),
	}

	return rs
}

func (*RPC) Name() string {
	return api.KindRPC.String()
}

func (*RPC) CanHandle(kind api.Kind) bool {
	return kind == api.KindRPC
}

func (r *RPC) Setup(router fiber.Router) error {
	handlers := streams.Concat(
		streams.FromSlice(r.chain.Handlers()),
		streams.FromSlice([]any{r.dispatch}),
	).Collect()

	group := router.Group(r.path)
	group.Post(constants.Empty, r.resolve, handlers...)

	return nil
}

// resolve is a middleware that parses the request and sets the operation in context.
func (r *RPC) resolve(ctx fiber.Ctx) error {
	req, err := r.parseRequest(ctx)
	if err != nil {
		return err
	}

	entry, ok := r.operations.Get(req.Identifier)
	if !ok {
		return &shared.NotFoundError{
			BaseError: shared.BaseError{
				Identifier: &req.Identifier,
				Err:        fiber.ErrNotFound,
			},
			Suggestion: r.findClosestApi(req.Identifier),
		}
	}

	shared.SetRequest(ctx, req)
	shared.SetOperation(ctx, entry.op)

	return ctx.Next()
}

func (r *RPC) Route(handler fiber.Handler, op *api.Operation) {
	r.operations.Put(op.Identifier, &routeEntry{
		op:      op,
		handler: handler,
	})
}

func (r *RPC) dispatch(ctx fiber.Ctx) error {
	req := shared.Request(ctx)

	entry, ok := r.operations.Get(req.Identifier)
	if !ok {
		return &shared.NotFoundError{
			BaseError: shared.BaseError{
				Identifier: &req.Identifier,
				Err:        fiber.ErrNotFound,
			},
			Suggestion: r.findClosestApi(req.Identifier),
		}
	}

	return entry.handler(ctx)
}

func (*RPC) parseRequest(ctx fiber.Ctx) (*api.Request, error) {
	req := &api.Request{
		Params: api.Params{},
		Meta:   api.Meta{},
	}

	if webhelpers.IsJSON(ctx) {
		if err := ctx.Bind().Body(req); err != nil {
			return nil, err
		}
	} else {
		if err := parseFormRequest(ctx, req); err != nil {
			return nil, err
		}
	}

	return req, nil
}

func (r *RPC) findClosestApi(requested api.Identifier) *api.Identifier {
	if r.operations.IsEmpty() {
		return nil
	}

	requestedStr := identifierToString(requested)

	var (
		closest     *api.Identifier
		minDistance = -1
	)

	for id := range r.operations.SeqKeys() {
		distance := edlib.LevenshteinDistance(requestedStr, identifierToString(id))
		if minDistance < 0 || distance < minDistance {
			minDistance = distance
			closest = &id
		}
	}

	// Only suggest if the distance is less than half the requested string length
	if closest != nil && minDistance < len(requestedStr)/2 {
		return closest
	}

	return nil
}

func parseFormRequest(ctx fiber.Ctx, request *api.Request) error {
	if err := ctx.Bind().Form(request); err != nil {
		return err
	}

	if params := ctx.FormValue(FormKeyParams); params != constants.Empty {
		if err := json.Unmarshal([]byte(params), &request.Params); err != nil {
			contextx.Logger(ctx).Warnf("Failed to parse params json: %v", err)

			return result.Err(
				i18n.T(result.ErrMessageApiRequestParamsInvalidJSON),
				result.WithCode(result.ErrCodeBadRequest),
			)
		}
	}

	if meta := ctx.FormValue(FormKeyMeta); meta != constants.Empty {
		if err := json.Unmarshal([]byte(meta), &request.Meta); err != nil {
			contextx.Logger(ctx).Warnf("Failed to parse meta json: %v", err)

			return result.Err(
				i18n.T(result.ErrMessageApiRequestMetaInvalidJSON),
				result.WithCode(result.ErrCodeBadRequest),
			)
		}
	}

	if webhelpers.IsMultipart(ctx) {
		if form, err := ctx.MultipartForm(); err == nil && form != nil {
			for key, files := range form.File {
				if len(files) > 0 {
					request.Params[key] = files
				}
			}
		}
	}

	return nil
}

func identifierToString(identifier api.Identifier) string {
	return fmt.Sprintf("%s/%s@%s", identifier.Resource, identifier.Action, identifier.Version)
}
