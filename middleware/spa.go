package middleware

import "io/fs"

// SpaConfig is the configuration for the Spa middleware.
type SpaConfig struct {
	Path string // Path is the path to the Spa.
	FS   fs.FS  // FS is the file system for the Spa.
}
