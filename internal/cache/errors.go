package cache

import "errors"

// ErrDirectoryRequired indicates persistent storage requires directory.
var ErrDirectoryRequired = errors.New("directory path is required for persistent storage")
