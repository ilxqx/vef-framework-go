package config

import (
	"fmt"

	"github.com/ilxqx/vef-framework-go/config"
	"github.com/ilxqx/vef-framework-go/internal/monitor"
)

// newAppConfig creates and parses application configuration from "vef.app" section.
func newAppConfig(cfg config.Config) (*config.AppConfig, error) {
	var appConfig config.AppConfig
	// Unmarshal extracts app config from "vef.app" section
	if err := cfg.Unmarshal("vef.app", &appConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal app config: %w", err)
	}

	return &appConfig, nil
}

// newDatasourceConfig creates and parses datasource configuration from "vef.datasource" section.
func newDatasourceConfig(cfg config.Config) (*config.DatasourceConfig, error) {
	var datasourceConfig config.DatasourceConfig
	// Unmarshal extracts datasource config from "vef.datasource" section
	if err := cfg.Unmarshal("vef.datasource", &datasourceConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal datasource config: %w", err)
	}

	return &datasourceConfig, nil
}

// newCorsConfig creates and parses CORS configuration from "vef.cors" section.
func newCorsConfig(cfg config.Config) (*config.CorsConfig, error) {
	var corsConfig config.CorsConfig
	// Unmarshal extracts CORS config from "vef.cors" section
	if err := cfg.Unmarshal("vef.cors", &corsConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cors config: %w", err)
	}

	return &corsConfig, nil
}

// newSecurityConfig creates and parses security configuration from "vef.security" section.
func newSecurityConfig(cfg config.Config) (*config.SecurityConfig, error) {
	var securityConfig config.SecurityConfig
	// Unmarshal extracts security config from "vef.security" section
	if err := cfg.Unmarshal("vef.security", &securityConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal security config: %w", err)
	}

	return &securityConfig, nil
}

// newRedisConfig creates and parses Redis configuration from "vef.redis" section.
func newRedisConfig(cfg config.Config) (*config.RedisConfig, error) {
	var redisConfig config.RedisConfig
	// Unmarshal extracts Redis config from "vef.redis" section
	if err := cfg.Unmarshal("vef.redis", &redisConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal redis config: %w", err)
	}

	return &redisConfig, nil
}

// newStorageConfig creates and parses storage configuration from "vef.storage" section.
func newStorageConfig(cfg config.Config) (*config.StorageConfig, error) {
	var storageConfig config.StorageConfig
	// Unmarshal extracts storage config from "vef.storage" section
	if err := cfg.Unmarshal("vef.storage", &storageConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal storage config: %w", err)
	}

	return &storageConfig, nil
}

// newMonitorConfig creates and parses monitor configuration from "vef.monitor" section.
func newMonitorConfig(cfg config.Config) (*config.MonitorConfig, error) {
	monitorConfig := monitor.DefaultConfig()
	// Unmarshal extracts monitor config from "vef.monitor" section
	if err := cfg.Unmarshal("vef.monitor", &monitorConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal monitor config: %w", err)
	}

	return &monitorConfig, nil
}
