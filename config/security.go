package config

import "time"

// SecurityConfig contains security-related configuration.
type SecurityConfig struct {
	TokenExpires time.Duration `config:"token_expires"` // Token expiration time
}
