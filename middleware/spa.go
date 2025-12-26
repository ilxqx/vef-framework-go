package middleware

import (
	"io/fs"
)

// SpaConfig is the configuration for the Spa middleware.
type SpaConfig struct {
	// Path is the path to the Spa.
	Path string
	// Fs is the file system for the Spa.
	Fs fs.FS
	// ExcludePaths are paths that should not be handled by the Spa (e.g., "/api", "/ws").
	ExcludePaths []string
}
