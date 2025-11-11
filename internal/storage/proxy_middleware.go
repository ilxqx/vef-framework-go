package storage

import (
	"errors"
	"mime"
	"net/url"
	"path/filepath"

	"github.com/gofiber/fiber/v3"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/i18n"
	"github.com/ilxqx/vef-framework-go/internal/app"
	"github.com/ilxqx/vef-framework-go/result"
	"github.com/ilxqx/vef-framework-go/storage"
)

// ProxyMiddleware implements a middleware that proxies file requests to the storage service.
type ProxyMiddleware struct {
	service storage.Service
}

func (p *ProxyMiddleware) Name() string {
	return "storage-proxy"
}

func (p *ProxyMiddleware) Order() int {
	// Apply after main middlewares but before SPA (order < 1000)
	return 900
}

func (p *ProxyMiddleware) Apply(router fiber.Router) {
	logger.Info("Registering storage proxy: GET /files/*")
	router.Get("/files/+", p.handleFileProxy)
}

// handleFileProxy handles file proxy requests.
// URL format: GET /files/{key}
// Example: GET /files/temp/2025/01/15/abc123.jpg
func (p *ProxyMiddleware) handleFileProxy(ctx fiber.Ctx) error {
	// Decode URL-encoded characters (e.g., %E6%B5%8B -> 测)
	key, err := url.PathUnescape(ctx.Params("+"))
	if err != nil {
		return result.Err(
			i18n.T(result.ErrMessageInvalidFileKey),
			result.WithCode(result.ErrCodeInvalidFileKey),
		)
	}

	reader, err := p.service.GetObject(ctx.Context(), storage.GetObjectOptions{
		Key: key,
	})
	if err != nil {
		if errors.Is(err, storage.ErrObjectNotFound) {
			return result.Err(
				i18n.T(result.ErrMessageFileNotFound),
				result.WithCode(result.ErrCodeFileNotFound),
			)
		}

		logger.Errorf("Failed to get object %s: %v", key, err)
		return result.Err(i18n.T(result.ErrMessageFailedToGetFile))
	}

	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			logger.Errorf("Failed to close file reader: %v", closeErr)
		}
	}()

	stat, err := p.service.StatObject(ctx.Context(), storage.StatObjectOptions{
		Key: key,
	})
	if err != nil {
		logger.Warnf("Failed to stat object %s: %v", key, err)
	}

	if stat != nil && stat.ContentType != constants.Empty {
		ctx.Set(fiber.HeaderContentType, stat.ContentType)
	} else {
		// Fallback: detect Content-Type from file extension
		ext := filepath.Ext(key)
		if contentType := mime.TypeByExtension(ext); contentType != constants.Empty {
			ctx.Set(fiber.HeaderContentType, contentType)
		} else {
			ctx.Set(fiber.HeaderContentType, fiber.MIMEOctetStream)
		}
	}

	ctx.Set(fiber.HeaderCacheControl, "public, max-age=86400, must-revalidate")
	if stat != nil && stat.ETag != constants.Empty {
		ctx.Set(fiber.HeaderETag, stat.ETag)
	}

	return ctx.SendStream(reader)
}

// NewProxyMiddleware creates a new storage proxy middleware.
func NewProxyMiddleware(service storage.Service) app.Middleware {
	return &ProxyMiddleware{
		service: service,
	}
}
