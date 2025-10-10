package api

import (
	"crypto/sha256"
	"encoding/base64"
	"net"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/extractors"
	"github.com/gofiber/fiber/v3/middleware/keyauth"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/contextx"
	"github.com/ilxqx/vef-framework-go/internal/security"
	"github.com/ilxqx/vef-framework-go/result"
	securityPkg "github.com/ilxqx/vef-framework-go/security"
	"github.com/ilxqx/vef-framework-go/webhelpers"
)

// buildAuthenticationMiddleware creates a keyauth middleware for API authentication.
// It extracts tokens from Authorization header or query parameter and validates them.
func buildAuthenticationMiddleware(manager api.Manager, auth securityPkg.AuthManager) fiber.Handler {
	return keyauth.New(keyauth.Config{
		Extractor: extractors.Chain(
			extractors.FromAuthHeader(constants.AuthSchemeBearer),
			extractors.FromQuery(constants.QueryKeyAccessToken),
		),
		Next: func(ctx fiber.Ctx) bool {
			request := contextx.APIRequest(ctx)
			definition := manager.Lookup(request.Identifier)

			return definition.IsPublic()
		},
		ErrorHandler: func(ctx fiber.Ctx, _ error) error {
			return fiber.ErrUnauthorized
		},
		Validator: func(ctx fiber.Ctx, accessToken string) (bool, error) {
			principal, err := auth.Authenticate(ctx.Context(), securityPkg.Authentication{
				Type:      security.AuthTypeToken,
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

// buildOpenAPIAuthenticationMiddleware creates middleware for OpenAPI authentication.
// It allows public endpoints to pass through and validates OpenAPI tokens for protected endpoints.
func buildOpenAPIAuthenticationMiddleware(manager api.Manager, auth securityPkg.AuthManager) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		request := contextx.APIRequest(ctx)
		definition := manager.Lookup(request.Identifier)

		if definition.IsPublic() {
			return ctx.Next()
		}

		// Extract headers
		appId := ctx.Get(constants.HeaderXAppId)
		timestamp := ctx.Get(constants.HeaderXTimestamp)
		signatureHex := ctx.Get(constants.HeaderXSignature)

		// Compute bodySha256Base64 from raw body bytes
		body := ctx.Body()
		sum := sha256.Sum256(body)
		bodySha256Base64 := base64.StdEncoding.EncodeToString(sum[:])

		// Build credentials: "<signatureHex>@<timestamp>@<bodySha256Base64>"
		credentials := signatureHex + constants.At + timestamp + constants.At + bodySha256Base64

		principal, err := auth.Authenticate(ctx.Context(), securityPkg.Authentication{
			Type:        security.AuthTypeOpenAPI,
			Principal:   appId,
			Credentials: credentials,
		})
		if err != nil {
			return err
		}

		// Optional external app config enforcement
		if principal != nil && principal.Details != nil {
			switch cfg := principal.Details.(type) {
			case securityPkg.ExternalAppConfig:
				if !cfg.Enabled {
					return result.ErrExternalAppDisabled
				}

				if strings.TrimSpace(cfg.IpWhitelist) != constants.Empty {
					if !ipAllowed(webhelpers.GetIP(ctx), cfg.IpWhitelist) {
						return result.ErrIpNotAllowed
					}
				}

			case *securityPkg.ExternalAppConfig:
				if cfg != nil {
					if !cfg.Enabled {
						return result.ErrExternalAppDisabled
					}

					if strings.TrimSpace(cfg.IpWhitelist) != constants.Empty {
						if !ipAllowed(webhelpers.GetIP(ctx), cfg.IpWhitelist) {
							return result.ErrIpNotAllowed
						}
					}
				}
			}
		}

		contextx.SetPrincipal(ctx, principal)

		return ctx.Next()
	}
}

// ipAllowed checks if client IP is in whitelist (comma-separated IP or CIDR list).
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
