package api

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"net"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/extractors"
	"github.com/gofiber/fiber/v3/middleware/keyauth"
	"github.com/samber/lo"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/contextx"
	isecurity "github.com/ilxqx/vef-framework-go/internal/security"
	"github.com/ilxqx/vef-framework-go/result"
	"github.com/ilxqx/vef-framework-go/security"
	"github.com/ilxqx/vef-framework-go/webhelpers"
)

func buildAuthenticationMiddleware(manager api.Manager, auth security.AuthManager) fiber.Handler {
	return keyauth.New(keyauth.Config{
		Extractor: extractors.Chain(
			extractors.FromAuthHeader(constants.AuthSchemeBearer),
			extractors.FromQuery(constants.QueryKeyAccessToken),
		),
		Next: func(ctx fiber.Ctx) bool {
			request := contextx.ApiRequest(ctx)
			definition := manager.Lookup(request.Identifier)

			return definition.IsPublic()
		},
		ErrorHandler: func(ctx fiber.Ctx, err error) error {
			identifier := contextx.ApiRequest(ctx).Identifier

			return &BaseError{
				Identifier: &identifier,
				Err:        lo.Ternary[error](errors.Is(err, keyauth.ErrMissingOrMalformedAPIKey), fiber.ErrUnauthorized, err),
			}
		},
		Validator: func(ctx fiber.Ctx, accessToken string) (bool, error) {
			principal, err := auth.Authenticate(ctx.Context(), security.Authentication{
				Type:      isecurity.AuthTypeToken,
				Principal: accessToken,
			})
			if err != nil {
				return false, err
			}

			contextx.SetPrincipal(ctx, principal)
			ctx.SetContext(
				contextx.SetPrincipal(ctx.Context(), principal),
			)

			return true, nil
		},
	})
}

func buildOpenApiAuthenticationMiddleware(manager api.Manager, auth security.AuthManager) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		request := contextx.ApiRequest(ctx)
		definition := manager.Lookup(request.Identifier)

		if definition.IsPublic() {
			return ctx.Next()
		}

		appId := ctx.Get(constants.HeaderXAppId)
		timestamp := ctx.Get(constants.HeaderXTimestamp)
		signatureHex := ctx.Get(constants.HeaderXSignature)

		body := ctx.Body()
		sum := sha256.Sum256(body)
		bodySha256Base64 := base64.StdEncoding.EncodeToString(sum[:])

		credentials := signatureHex + constants.At + timestamp + constants.At + bodySha256Base64

		principal, err := auth.Authenticate(ctx.Context(), security.Authentication{
			Type:        isecurity.AuthTypeOpenApi,
			Principal:   appId,
			Credentials: credentials,
		})
		if err != nil {
			return &BaseError{
				Identifier: &request.Identifier,
				Err:        err,
			}
		}

		if principal != nil && principal.Details != nil {
			switch cfg := principal.Details.(type) {
			case security.ExternalAppConfig:
				if !cfg.Enabled {
					return &BaseError{
						Identifier: &request.Identifier,
						Err:        result.ErrExternalAppDisabled,
					}
				}

				if strings.TrimSpace(cfg.IpWhitelist) != constants.Empty {
					if !ipAllowed(webhelpers.GetIp(ctx), cfg.IpWhitelist) {
						return &BaseError{
							Identifier: &request.Identifier,
							Err:        result.ErrIpNotAllowed,
						}
					}
				}

			case *security.ExternalAppConfig:
				if cfg != nil {
					if !cfg.Enabled {
						return &BaseError{
							Identifier: &request.Identifier,
							Err:        result.ErrExternalAppDisabled,
						}
					}

					if strings.TrimSpace(cfg.IpWhitelist) != constants.Empty {
						if !ipAllowed(webhelpers.GetIp(ctx), cfg.IpWhitelist) {
							return &BaseError{
								Identifier: &request.Identifier,
								Err:        result.ErrIpNotAllowed,
							}
						}
					}
				}
			}
		}

		contextx.SetPrincipal(ctx, principal)

		return ctx.Next()
	}
}

func ipAllowed(clientIP, whitelist string) bool {
	if strings.TrimSpace(whitelist) == constants.Empty {
		return true
	}

	ip := net.ParseIP(clientIP)
	if ip == nil {
		return false
	}

	for entry := range strings.SplitSeq(whitelist, constants.Comma) {
		entry = strings.TrimSpace(entry)
		if entry == constants.Empty {
			continue
		}

		if strings.Contains(entry, constants.Slash) {
			_, ipNet, err := net.ParseCIDR(entry)
			if err != nil {
				continue
			}

			if ipNet.Contains(ip) {
				return true
			}
		} else if entry == clientIP {
			return true
		}
	}

	return false
}
