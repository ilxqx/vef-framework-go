package middleware

import "io/fs"

// SPAConfig is the configuration for the SPA middleware.
type SPAConfig struct {
	Path string // Path is the path to the SPA.
	FS   fs.FS  // FS is the file system for the SPA.
}
