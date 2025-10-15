package database

import (
	"github.com/uptrace/bun"

	"github.com/ilxqx/vef-framework-go/config"
	logPkg "github.com/ilxqx/vef-framework-go/log"
)

// databaseOptions holds configuration options for database initialization.
type databaseOptions struct {
	Config          *config.DatasourceConfig
	PoolConfig      *ConnectionPoolConfig
	EnableQueryHook bool
	Logger          logPkg.Logger
	BunOptions      []bun.DBOption
}

// Option defines a function type for configuring databaseOptions.
type Option func(*databaseOptions)

// newDefaultOptions creates default database options.
func newDefaultOptions(config *config.DatasourceConfig) *databaseOptions {
	return &databaseOptions{
		Config:          config,
		PoolConfig:      NewDefaultConnectionPoolConfig(),
		EnableQueryHook: true,
		Logger:          logger,
		BunOptions:      []bun.DBOption{bun.WithDiscardUnknownColumns()},
	}
}

// WithConnectionPool sets a custom connection pool configuration.
func WithConnectionPool(poolConfig *ConnectionPoolConfig) Option {
	return func(opts *databaseOptions) {
		opts.PoolConfig = poolConfig
	}
}

// DisableQueryHook disables the query hook.
// By default, query hook is enabled for logging SQL queries.
func DisableQueryHook() Option {
	return func(opts *databaseOptions) {
		opts.EnableQueryHook = false
	}
}

// WithLogger sets a custom logger.
func WithLogger(logger logPkg.Logger) Option {
	return func(opts *databaseOptions) {
		opts.Logger = logger
	}
}

// WithBunOptions adds additional bun options.
func WithBunOptions(bunOpts ...bun.DBOption) Option {
	return func(opts *databaseOptions) {
		opts.BunOptions = append(opts.BunOptions, bunOpts...)
	}
}

// apply applies the given options to the databaseOptions.
func (opts *databaseOptions) apply(options ...Option) {
	for _, opt := range options {
		opt(opts)
	}
}
