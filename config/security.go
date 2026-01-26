package config

import "time"

// SecurityConfig defines security settings.
type SecurityConfig struct {
	TokenExpires time.Duration `config:"token_expires"`
}
