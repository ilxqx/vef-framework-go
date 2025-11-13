package database

import (
	"github.com/uptrace/bun"

	"github.com/ilxqx/vef-framework-go/config"
	"github.com/ilxqx/vef-framework-go/log"
)

type databaseOptions struct {
	Config          *config.DatasourceConfig
	PoolConfig      *ConnectionPoolConfig
	EnableQueryHook bool
	Logger          log.Logger
	BunOptions      []bun.DBOption
}

type Option func(*databaseOptions)

func newDefaultOptions(config *config.DatasourceConfig) *databaseOptions {
	return &databaseOptions{
		Config:          config,
		PoolConfig:      NewDefaultConnectionPoolConfig(),
		EnableQueryHook: true,
		Logger:          logger,
		BunOptions:      []bun.DBOption{bun.WithDiscardUnknownColumns()},
	}
}

func WithConnectionPool(poolConfig *ConnectionPoolConfig) Option {
	return func(opts *databaseOptions) {
		opts.PoolConfig = poolConfig
	}
}

// DisableQueryHook disables query logging which is enabled by default.
func DisableQueryHook() Option {
	return func(opts *databaseOptions) {
		opts.EnableQueryHook = false
	}
}

func WithLogger(logger log.Logger) Option {
	return func(opts *databaseOptions) {
		opts.Logger = logger
	}
}

func WithBunOptions(bunOpts ...bun.DBOption) Option {
	return func(opts *databaseOptions) {
		opts.BunOptions = append(opts.BunOptions, bunOpts...)
	}
}

func (opts *databaseOptions) apply(options ...Option) {
	for _, opt := range options {
		opt(opts)
	}
}
