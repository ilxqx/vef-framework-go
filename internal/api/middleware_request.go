package api

import (
	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v3"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/contextx"
	"github.com/ilxqx/vef-framework-go/webhelpers"
)

// requestMiddleware parses the request body and validates the Api definition exists.
// It stores the parsed request in the context for use by subsequent middlewares.
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
			if err := ctx.Bind().Form(&request); err != nil {
				return err
			}

			if params := ctx.FormValue("params"); params != constants.Empty {
				if err := json.Unmarshal([]byte(params), &request.Params); err != nil {
					return err
				}
			}

			if meta := ctx.FormValue("meta"); meta != constants.Empty {
				if err := json.Unmarshal([]byte(meta), &request.Meta); err != nil {
					return err
				}
			}

			if form, err := ctx.MultipartForm(); err == nil && form != nil {
				for key, files := range form.File {
					if len(files) > 0 {
						request.Params[key] = files
					}
				}
			}
		}

		definition := manager.Lookup(request.Identifier)
		if definition == nil {
			return fiber.ErrNotFound
		}

		contextx.SetApiRequest(ctx, &request)

		return ctx.Next()
	}
}
