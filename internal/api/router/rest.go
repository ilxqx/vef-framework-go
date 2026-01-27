package router

import (
	"encoding/json"
	"maps"
	"slices"
	"strings"

	"github.com/gofiber/fiber/v3"

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
	DefaultRESTBasePath = "/api"
)

// REST implements api.RouterStrategy for RESTful routing.
type REST struct {
	basePath string
	chain    *middleware.Chain
	group    fiber.Router
}

// NewREST creates a new RESTful router.
func NewREST(basePath string, chain *middleware.Chain) api.RouterStrategy {
	if basePath == constants.Empty {
		basePath = DefaultRESTBasePath
	}

	return &REST{
		basePath: basePath,
		chain:    chain,
	}
}

func (*REST) Name() string {
	return api.KindREST.String()
}

func (*REST) CanHandle(kind api.Kind) bool {
	return kind == api.KindREST
}

func (r *REST) Setup(router fiber.Router) error {
	r.group = router.Group(r.basePath)

	return nil
}

// Route registers an operation with the router.
func (r *REST) Route(handler fiber.Handler, op *api.Operation) {
	method, subPath := r.parseAction(op.Action)
	fullPath := r.buildPath(op.Resource, subPath)

	resolver := r.createResolver(op)
	handlers := append(slices.Clone(r.chain.Handlers()), handler)

	r.group.Add([]string{method}, fullPath, resolver, handlers...)

	op.Meta[shared.MetaKeyRESTHttpMethod] = method
	op.Meta[shared.MetaKeyRESTHttpPath] = r.basePath + fullPath
}

// createResolver creates a middleware that parses request and sets operation in context.
func (r *REST) createResolver(op *api.Operation) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		req, err := r.parseRequest(ctx, op)
		if err != nil {
			return err
		}

		shared.SetOperation(ctx, op)
		shared.SetRequest(ctx, req)

		return ctx.Next()
	}
}

// parseAction extracts HTTP method and sub-path from action string.
// Format: "METHOD [/path]" (e.g., "GET", "POST /items", "DELETE /:id").
func (*REST) parseAction(action string) (method, subPath string) {
	parts := strings.SplitN(action, constants.Space, 2)
	method = strings.ToUpper(parts[0])

	if len(parts) < 2 {
		return method, subPath
	}

	subPath = strings.TrimSpace(parts[1])
	if !strings.HasPrefix(subPath, constants.Slash) {
		subPath = constants.Slash + subPath
	}

	return method, subPath
}

// buildPath constructs the full URL path for the operation.
func (*REST) buildPath(resource, subPath string) string {
	return constants.Slash + resource + subPath
}

// parseRequest extracts Request from HTTP request context.
func (r *REST) parseRequest(ctx fiber.Ctx, op *api.Operation) (*api.Request, error) {
	req := &api.Request{
		Identifier: op.Identifier,
		Params:     make(api.Params),
		Meta:       make(api.Meta),
	}

	r.extractMeta(ctx, req)
	r.extractPathParams(ctx, req)
	r.extractQueryParams(ctx, req)

	if err := r.parseBodyIfNeeded(ctx, req); err != nil {
		return nil, err
	}

	return req, nil
}

// extractMeta extracts meta values from HTTP headers with prefix X-Meta-.
func (*REST) extractMeta(ctx fiber.Ctx, req *api.Request) {
	for key, values := range ctx.GetReqHeaders() {
		if metaKey, found := strings.CutPrefix(key, constants.HeaderXMetaPrefix); found && len(values) > 0 {
			req.Meta[metaKey] = values[0]
		}
	}
}

// extractPathParams extracts path parameters from the route.
func (*REST) extractPathParams(ctx fiber.Ctx, req *api.Request) {
	for _, key := range ctx.Route().Params {
		req.Params[key] = ctx.Params(key)
	}
}

// extractQueryParams extracts query parameters from the URL.
func (*REST) extractQueryParams(ctx fiber.Ctx, req *api.Request) {
	for key, value := range ctx.Queries() {
		req.Params[key] = value
	}
}

// parseBodyIfNeeded parses request body for POST/PUT/PATCH methods.
func (r *REST) parseBodyIfNeeded(ctx fiber.Ctx, req *api.Request) error {
	method := ctx.Method()
	if method != fiber.MethodPost && method != fiber.MethodPut && method != fiber.MethodPatch {
		return nil
	}

	return r.parseBody(ctx, req)
}

// parseBody parses request body based on content type.
func (r *REST) parseBody(ctx fiber.Ctx, req *api.Request) error {
	if webhelpers.IsJSON(ctx) {
		return r.parseJSONBody(ctx, req)
	}

	if webhelpers.IsMultipart(ctx) {
		return r.parseMultipartForm(ctx, req)
	}

	return nil
}

// parseJSONBody parses JSON request body into params.
func (*REST) parseJSONBody(ctx fiber.Ctx, req *api.Request) error {
	body := ctx.Body()
	if len(body) == 0 {
		return nil
	}

	var data map[string]any
	if err := json.Unmarshal(body, &data); err != nil {
		contextx.Logger(ctx).Warnf("Failed to parse JSON body: %v", err)

		return result.Err(
			i18n.T(result.ErrMessageApiRequestParamsInvalidJSON),
			result.WithCode(result.ErrCodeBadRequest),
		)
	}

	maps.Copy(req.Params, data)

	return nil
}

// parseMultipartForm handles multipart form data.
func (*REST) parseMultipartForm(ctx fiber.Ctx, req *api.Request) error {
	form, err := ctx.MultipartForm()
	if err != nil {
		return err
	}

	if form == nil {
		return nil
	}

	for key, values := range form.Value {
		switch len(values) {
		case 0:
			continue
		case 1:
			req.Params[key] = values[0]
		default:
			req.Params[key] = values
		}
	}

	for key, files := range form.File {
		if len(files) > 0 {
			req.Params[key] = files
		}
	}

	return nil
}
