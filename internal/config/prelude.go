package config

import (
	"fmt"

	configPkg "github.com/ilxqx/vef-framework-go/config"
)

func newAppConfig(config configPkg.Config) (*configPkg.AppConfig, error) {
	var appConfig configPkg.AppConfig
	// Unmarshal extracts app config from "vef.app" section
	if err := config.Unmarshal("vef.app", &appConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal app config: %w", err)
	}

	return &appConfig, nil
}

func newDatasourceConfig(config configPkg.Config) (*configPkg.DatasourceConfig, error) {
	var datasourceConfig configPkg.DatasourceConfig
	// Unmarshal extracts datasource config from "vef.datasource" section
	if err := config.Unmarshal("vef.datasource", &datasourceConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal datasource config: %w", err)
	}

	return &datasourceConfig, nil
}

func newCorsConfig(config configPkg.Config) (*configPkg.CorsConfig, error) {
	var corsConfig configPkg.CorsConfig
	// Unmarshal extracts CORS config from "vef.cors" section
	if err := config.Unmarshal("vef.cors", &corsConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cors config: %w", err)
	}

	return &corsConfig, nil
}

func newSecurityConfig(config configPkg.Config) (*configPkg.SecurityConfig, error) {
	var securityConfig configPkg.SecurityConfig
	// Unmarshal extracts security config from "vef.security" section
	if err := config.Unmarshal("vef.security", &securityConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal security config: %w", err)
	}

	return &securityConfig, nil
}

func newRedisConfig(config configPkg.Config) (*configPkg.RedisConfig, error) {
	var redisConfig configPkg.RedisConfig
	// Unmarshal extracts Redis config from "vef.redis" section
	if err := config.Unmarshal("vef.redis", &redisConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal redis config: %w", err)
	}

	return &redisConfig, nil
}

func newCacheConfig(config configPkg.Config) (*configPkg.CacheConfig, error) {
	var cacheConfig configPkg.CacheConfig
	// Unmarshal extracts cache config from "vef.cache" section
	if err := config.Unmarshal("vef.cache", &cacheConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cache config: %w", err)
	}

	return &cacheConfig, nil
}

func newI18nConfig(config configPkg.Config) (*configPkg.I18nConfig, error) {
	var i18nConfig configPkg.I18nConfig
	// Unmarshal extracts i18n config from "vef.i18n" section
	if err := config.Unmarshal("vef.i18n", &i18nConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal i18n config: %w", err)
	}

	return &i18nConfig, nil
}
