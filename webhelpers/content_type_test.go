package webhelpers

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsJSON(t *testing.T) {
	t.Run("applicationJson", func(t *testing.T) {
		app := fiber.New()

		app.Post("/json", func(c fiber.Ctx) error {
			result := IsJSON(c)
			assert.True(t, result, "Should return true for application/json")

			return c.SendStatus(fiber.StatusOK)
		})

		req := httptest.NewRequest("POST", "/json", nil)
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)
	})

	t.Run("applicationJsonWithCharset", func(t *testing.T) {
		app := fiber.New()

		app.Post("/json", func(c fiber.Ctx) error {
			result := IsJSON(c)
			assert.True(t, result, "Should return true for application/json with charset")

			return c.SendStatus(fiber.StatusOK)
		})

		req := httptest.NewRequest("POST", "/json", nil)
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSONCharsetUTF8)

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)
	})

	t.Run("missingContentType", func(t *testing.T) {
		app := fiber.New()

		app.Post("/json", func(c fiber.Ctx) error {
			result := IsJSON(c)
			assert.False(t, result, "Should return false when Content-Type header is missing")

			return c.SendStatus(fiber.StatusOK)
		})

		req := httptest.NewRequest("POST", "/json", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)
	})

	t.Run("nonJsonContentType", func(t *testing.T) {
		app := fiber.New()

		app.Post("/json", func(c fiber.Ctx) error {
			result := IsJSON(c)
			assert.False(t, result, "Should return false for non-JSON Content-Type")

			return c.SendStatus(fiber.StatusOK)
		})

		req := httptest.NewRequest("POST", "/json", nil)
		req.Header.Set(fiber.HeaderContentType, fiber.MIMETextPlain)

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)
	})
}

func TestIsMultipart(t *testing.T) {
	t.Run("multipartFormData", func(t *testing.T) {
		app := fiber.New()

		app.Post("/multipart", func(c fiber.Ctx) error {
			result := IsMultipart(c)
			assert.True(t, result, "Should return true for multipart/form-data")

			return c.SendStatus(fiber.StatusOK)
		})

		req := httptest.NewRequest("POST", "/multipart", nil)
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEMultipartForm+"; boundary=MyBoundary")

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)
	})

	t.Run("nonMultipartContentType", func(t *testing.T) {
		app := fiber.New()

		app.Post("/multipart", func(c fiber.Ctx) error {
			result := IsMultipart(c)
			assert.False(t, result, "Should return false for non-multipart Content-Type")

			return c.SendStatus(fiber.StatusOK)
		})

		req := httptest.NewRequest("POST", "/multipart", nil)
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)
	})

	t.Run("missingContentType", func(t *testing.T) {
		app := fiber.New()

		app.Post("/multipart", func(c fiber.Ctx) error {
			result := IsMultipart(c)
			assert.False(t, result, "Should return false when Content-Type header is missing")

			return c.SendStatus(fiber.StatusOK)
		})

		req := httptest.NewRequest("POST", "/multipart", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)
	})
}
