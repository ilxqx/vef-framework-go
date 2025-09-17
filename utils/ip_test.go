package utils

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetIP(t *testing.T) {
	t.Run("uses X-Forwarded-For header when present", func(t *testing.T) {
		app := fiber.New()
		forwardedIP := "192.168.1.100"

		app.Get("/test", func(c fiber.Ctx) error {
			ip := GetIP(c)
			assert.Equal(t, forwardedIP, ip)
			return c.SendString(ip)
		})

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set(fiber.HeaderXForwardedFor, forwardedIP)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	})

	t.Run("falls back to direct IP when X-Forwarded-For is empty", func(t *testing.T) {
		app := fiber.New()

		app.Get("/test", func(c fiber.Ctx) error {
			ip := GetIP(c)
			// Should return the direct IP (which is the test client IP)
			assert.NotEmpty(t, ip)
			return c.SendString(ip)
		})

		req := httptest.NewRequest("GET", "/test", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	})

	t.Run("handles multiple IPs in X-Forwarded-For", func(t *testing.T) {
		app := fiber.New()
		forwardedIPs := "192.168.1.100, 10.0.0.50, 172.16.0.1"

		app.Get("/test", func(c fiber.Ctx) error {
			ip := GetIP(c)
			assert.Equal(t, forwardedIPs, ip)
			return c.SendString(ip)
		})

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set(fiber.HeaderXForwardedFor, forwardedIPs)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	})

	t.Run("handles IPv6 addresses", func(t *testing.T) {
		app := fiber.New()
		ipv6Address := "2001:0db8:85a3:0000:0000:8a2e:0370:7334"

		app.Get("/test", func(c fiber.Ctx) error {
			ip := GetIP(c)
			assert.Equal(t, ipv6Address, ip)
			return c.SendString(ip)
		})

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set(fiber.HeaderXForwardedFor, ipv6Address)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	})

	t.Run("handles localhost addresses", func(t *testing.T) {
		app := fiber.New()
		localhostIP := "127.0.0.1"

		app.Get("/test", func(c fiber.Ctx) error {
			ip := GetIP(c)
			assert.Equal(t, localhostIP, ip)
			return c.SendString(ip)
		})

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set(fiber.HeaderXForwardedFor, localhostIP)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	})

	t.Run("empty X-Forwarded-For header falls back to IP", func(t *testing.T) {
		app := fiber.New()

		app.Get("/test", func(c fiber.Ctx) error {
			ip := GetIP(c)
			// When X-Forwarded-For is empty, should fall back to c.IP()
			assert.NotEmpty(t, ip)
			return c.SendString(ip)
		})

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set(fiber.HeaderXForwardedFor, "")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	})
}