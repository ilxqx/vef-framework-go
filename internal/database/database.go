package database

import (
	"database/sql"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/schema"

	"github.com/ilxqx/vef-framework-go/config"
	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/log"
)

func logDBVersion(provider DatabaseProvider, db *bun.DB, logger log.Logger) error {
	version, err := provider.QueryVersion(db)
	if err != nil {
		return wrapVersionQueryError(provider.Type(), err)
	}

	logger.Infof("Database type: %s | Database version: %s", provider.Type(), version)

	return nil
}

func setupBunDB(sqlDB *sql.DB, dialect schema.Dialect, opts *databaseOptions) *bun.DB {
	db := bun.NewDB(sqlDB, dialect, opts.BunOptions...)

	if opts.EnableQueryHook {
		addQueryHook(db, opts.Logger, opts.SQLGuardConfig)
	}

	db = db.WithNamedArg(constants.PlaceholderKeyOperator, constants.OperatorSystem)

	return db
}

func New(cfg *config.DatasourceConfig, options ...Option) (*bun.DB, error) {
	provider, exists := registry.provider(cfg.Type)
	if !exists {
		return nil, newUnsupportedDBTypeError(cfg.Type)
	}

	sqlDB, dialect, err := provider.Connect(cfg)
	if err != nil || sqlDB == nil {
		return nil, err
	}

	opts := newDefaultOptions(cfg)
	opts.apply(options...)

	if opts.PoolConfig != nil {
		opts.PoolConfig.ApplyToDB(sqlDB)
	}

	return setupBunDB(sqlDB, dialect, opts), nil
}
