package webhelpers

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsJSON(t *testing.T) {
	t.Run("Returns true when Content-Type is application/json", func(t *testing.T) {
		app := fiber.New()

		app.Post("/json", func(c fiber.Ctx) error {
			result := IsJson(c)
			assert.True(t, result)

			return c.SendStatus(fiber.StatusOK)
		})

		req := httptest.NewRequest("POST", "/json", nil)
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)
	})

	t.Run("Returns true for Content-Type with charset", func(t *testing.T) {
		app := fiber.New()

		app.Post("/json", func(c fiber.Ctx) error {
			result := IsJson(c)
			assert.True(t, result)

			return c.SendStatus(fiber.StatusOK)
		})

		req := httptest.NewRequest("POST", "/json", nil)
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSONCharsetUTF8)

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)
	})

	t.Run("Returns false when Content-Type header is missing", func(t *testing.T) {
		app := fiber.New()

		app.Post("/json", func(c fiber.Ctx) error {
			result := IsJson(c)
			assert.False(t, result)

			return c.SendStatus(fiber.StatusOK)
		})

		req := httptest.NewRequest("POST", "/json", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)
	})

	t.Run("Returns false for non JSON Content-Type", func(t *testing.T) {
		app := fiber.New()

		app.Post("/json", func(c fiber.Ctx) error {
			result := IsJson(c)
			assert.False(t, result)

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
	t.Run("Returns true for multipart form data", func(t *testing.T) {
		app := fiber.New()

		app.Post("/multipart", func(c fiber.Ctx) error {
			result := IsMultipart(c)
			assert.True(t, result)

			return c.SendStatus(fiber.StatusOK)
		})

		req := httptest.NewRequest("POST", "/multipart", nil)
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEMultipartForm+"; boundary=MyBoundary")

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)
	})

	t.Run("Returns false for non multipart Content-Type", func(t *testing.T) {
		app := fiber.New()

		app.Post("/multipart", func(c fiber.Ctx) error {
			result := IsMultipart(c)
			assert.False(t, result)

			return c.SendStatus(fiber.StatusOK)
		})

		req := httptest.NewRequest("POST", "/multipart", nil)
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)
	})

	t.Run("Returns false when Content-Type header is missing", func(t *testing.T) {
		app := fiber.New()

		app.Post("/multipart", func(c fiber.Ctx) error {
			result := IsMultipart(c)
			assert.False(t, result)

			return c.SendStatus(fiber.StatusOK)
		})

		req := httptest.NewRequest("POST", "/multipart", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)
	})
}
