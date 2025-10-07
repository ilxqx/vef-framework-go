package webhelpers

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetIP(t *testing.T) {
	t.Run("Uses X-Forwarded-For header when present", func(t *testing.T) {
		app := fiber.New()
		forwardedIP := "192.168.1.100"

		app.Get("/test", func(c fiber.Ctx) error {
			ip := GetIP(c)
			assert.Equal(t, forwardedIP, ip)

			return c.SendString(ip)
		})

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Forwarded-For", forwardedIP)
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)
	})

	t.Run("Falls back to direct IP when X-Forwarded-For is not present", func(t *testing.T) {
		app := fiber.New()

		app.Get("/test", func(c fiber.Ctx) error {
			ip := GetIP(c)
			// The IP should be the test IP (usually "0.0.0.0")
			assert.NotEmpty(t, ip)

			return c.SendString(ip)
		})

		req := httptest.NewRequest("GET", "/test", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)
	})

	t.Run("Uses X-Forwarded-For over direct IP", func(t *testing.T) {
		app := fiber.New()
		forwardedIP := "10.0.0.1"

		app.Get("/test", func(c fiber.Ctx) error {
			ip := GetIP(c)
			assert.Equal(t, forwardedIP, ip)

			return c.SendString(ip)
		})

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Forwarded-For", forwardedIP)
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)
	})

	t.Run("Handles empty X-Forwarded-For header", func(t *testing.T) {
		app := fiber.New()

		app.Get("/test", func(c fiber.Ctx) error {
			ip := GetIP(c)
			// Should fall back to direct IP when header is empty
			assert.NotEmpty(t, ip)

			return c.SendString(ip)
		})

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Forwarded-For", "")
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)
	})

	t.Run("Handles multiple IPs in X-Forwarded-For", func(t *testing.T) {
		app := fiber.New()
		// X-Forwarded-For can contain multiple IPs, typically the first one is the original client
		forwardedIPs := "203.0.113.195, 70.41.3.18, 150.172.238.178"

		app.Get("/test", func(c fiber.Ctx) error {
			ip := GetIP(c)
			assert.Equal(t, forwardedIPs, ip)

			return c.SendString(ip)
		})

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Forwarded-For", forwardedIPs)
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)
	})
}
