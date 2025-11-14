package api

import (
	"encoding/json"

	"github.com/gofiber/fiber/v3"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/contextx"
	"github.com/ilxqx/vef-framework-go/i18n"
	"github.com/ilxqx/vef-framework-go/result"
	"github.com/ilxqx/vef-framework-go/webhelpers"
)

// requestMiddleware stores the parsed request in the context for use by subsequent middlewares.
func requestMiddleware(manager api.Manager) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		request := api.Request{
			Params: api.Params{},
			Meta:   api.Meta{},
		}

		if webhelpers.IsJson(ctx) {
			if err := ctx.Bind().Body(&request); err != nil {
				return err
			}
		} else {
			if err := parseFormRequest(ctx, &request); err != nil {
				return err
			}
		}

		definition := manager.Lookup(request.Identifier)
		if definition == nil {
			return &Error{
				Identifier: request.Identifier,
				Err:        fiber.ErrNotFound,
			}
		}

		contextx.SetApiRequest(ctx, &request)

		return ctx.Next()
	}
}

// parseFormRequest parses form or multipart/form-data requests into api.Request.
func parseFormRequest(ctx fiber.Ctx, request *api.Request) error {
	if err := ctx.Bind().Form(request); err != nil {
		return err
	}

	if params := ctx.FormValue("params"); params != constants.Empty {
		if err := json.Unmarshal([]byte(params), &request.Params); err != nil {
			contextx.Logger(ctx).Warnf("Failed to parse params json: %v", err)

			return result.Err(
				i18n.T(result.ErrMessageApiRequestParamsInvalidJson),
				result.WithCode(result.ErrCodeBadRequest),
			)
		}
	}

	if meta := ctx.FormValue("meta"); meta != constants.Empty {
		if err := json.Unmarshal([]byte(meta), &request.Meta); err != nil {
			contextx.Logger(ctx).Warnf("Failed to parse meta json: %v", err)

			return result.Err(
				i18n.T(result.ErrMessageApiRequestMetaInvalidJson),
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
