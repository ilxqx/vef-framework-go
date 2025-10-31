package webhelpers

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetIp(t *testing.T) {
	t.Run("xForwardedForHeader", func(t *testing.T) {
		app := fiber.New()
		forwardedIP := "192.168.1.100"

		app.Get("/test", func(c fiber.Ctx) error {
			ip := GetIp(c)
			assert.Equal(t, forwardedIP, ip, "Should use X-Forwarded-For header")

			return c.SendString(ip)
		})

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Forwarded-For", forwardedIP)
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)
	})

	t.Run("fallbackToDirectIP", func(t *testing.T) {
		app := fiber.New()

		app.Get("/test", func(c fiber.Ctx) error {
			ip := GetIp(c)
			assert.NotEmpty(t, ip, "Should return direct IP when X-Forwarded-For is not present")

			return c.SendString(ip)
		})

		req := httptest.NewRequest("GET", "/test", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)
	})

	t.Run("xForwardedForOverridesDirectIP", func(t *testing.T) {
		app := fiber.New()
		forwardedIP := "10.0.0.1"

		app.Get("/test", func(c fiber.Ctx) error {
			ip := GetIp(c)
			assert.Equal(t, forwardedIP, ip, "Should use X-Forwarded-For over direct IP")

			return c.SendString(ip)
		})

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Forwarded-For", forwardedIP)
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)
	})

	t.Run("emptyXForwardedForHeader", func(t *testing.T) {
		app := fiber.New()

		app.Get("/test", func(c fiber.Ctx) error {
			ip := GetIp(c)
			assert.NotEmpty(t, ip, "Should fall back to direct IP when header is empty")

			return c.SendString(ip)
		})

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Forwarded-For", "")
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)
	})

	t.Run("multipleIPsInXForwardedFor", func(t *testing.T) {
		app := fiber.New()
		forwardedIPs := "203.0.113.195, 70.41.3.18, 150.172.238.178"

		app.Get("/test", func(c fiber.Ctx) error {
			ip := GetIp(c)
			assert.Equal(t, forwardedIPs, ip, "Should return the full X-Forwarded-For value")

			return c.SendString(ip)
		})

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Forwarded-For", forwardedIPs)
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)
	})
}
