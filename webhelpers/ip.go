package webhelpers

import "github.com/gofiber/fiber/v3"

// GetIp retrieves X-Forwarded-For header or falls back to direct IP.
func GetIp(ctx fiber.Ctx) string {
	return ctx.Get(fiber.HeaderXForwardedFor, ctx.IP())
}
