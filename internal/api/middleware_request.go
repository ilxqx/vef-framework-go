package api

import (
	"encoding/json"
	"fmt"

	"github.com/gofiber/fiber/v3"
	"github.com/hbollon/go-edlib"

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
			return &NotFoundError{
				BaseError: BaseError{
					Identifier: &request.Identifier,
					Err:        fiber.ErrNotFound,
				},
				Suggestion: findClosestApi(manager, request.Identifier),
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

// identifierToString converts an api.Identifier to a comparable string for similarity matching.
func identifierToString(id api.Identifier) string {
	return fmt.Sprintf("%s/%s@%s", id.Resource, id.Action, id.Version)
}

// findClosestApi finds the most similar registered API to the requested identifier.
// Returns nil if no similar API is found or if the similarity is too low.
func findClosestApi(manager api.Manager, requested api.Identifier) *api.Identifier {
	allDefinitions := manager.List()
	if len(allDefinitions) == 0 {
		return nil
	}

	requestedStr := identifierToString(requested)

	var (
		closestIdentifier *api.Identifier
		minDistance       = -1
	)

	for _, def := range allDefinitions {
		defStr := identifierToString(def.Identifier)

		// Use Levenshtein distance for similarity matching
		distance := edlib.LevenshteinDistance(requestedStr, defStr)

		if minDistance == -1 || distance < minDistance {
			minDistance = distance
			identifier := def.Identifier
			closestIdentifier = &identifier
		}
	}

	// Only return suggestion if similarity is reasonable (distance < 50% of requested string length)
	if closestIdentifier != nil && minDistance < len(requestedStr)/2 {
		return closestIdentifier
	}

	return nil
}
