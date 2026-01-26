package database

import (
	"github.com/uptrace/bun"

	"github.com/ilxqx/vef-framework-go/config"
	"github.com/ilxqx/vef-framework-go/internal/database/sqlguard"
	"github.com/ilxqx/vef-framework-go/log"
)

type databaseOptions struct {
	Config          *config.DatasourceConfig
	PoolConfig      *ConnectionPoolConfig
	EnableQueryHook bool
	Logger          log.Logger
	BunOptions      []bun.DBOption
	SqlGuardConfig  *sqlguard.Config
}

type Option func(*databaseOptions)

func newDefaultOptions(cfg *config.DatasourceConfig) *databaseOptions {
	var guardConfig *sqlguard.Config
	if cfg.EnableSqlGuard {
		guardConfig = sqlguard.DefaultConfig()
	}

	return &databaseOptions{
		Config:          cfg,
		PoolConfig:      NewDefaultConnectionPoolConfig(),
		EnableQueryHook: true,
		Logger:          logger,
		BunOptions:      []bun.DBOption{bun.WithDiscardUnknownColumns()},
		SqlGuardConfig:  guardConfig,
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

// WithSqlGuardConfig sets a custom sql guard configuration.
func WithSqlGuardConfig(cfg *sqlguard.Config) Option {
	return func(opts *databaseOptions) {
		opts.SqlGuardConfig = cfg
	}
}

// DisableSqlGuard disables the sql guard.
func DisableSqlGuard() Option {
	return func(opts *databaseOptions) {
		opts.SqlGuardConfig = nil
	}
}

func (opts *databaseOptions) apply(options ...Option) {
	for _, opt := range options {
		opt(opts)
	}
}
