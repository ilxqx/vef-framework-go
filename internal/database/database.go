package database

import (
	"database/sql"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/schema"

	"github.com/ilxqx/vef-framework-go/config"
	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/log"
)

func logDbVersion(provider DatabaseProvider, db *bun.DB, logger log.Logger) error {
	version, err := provider.QueryVersion(db)
	if err != nil {
		return wrapVersionQueryError(provider.Type(), err)
	}

	logger.Infof("Database type: %s | Database version: %s", provider.Type(), version)

	return nil
}

func setupBunDB(sqlDb *sql.DB, dialect schema.Dialect, opts *databaseOptions) *bun.DB {
	db := bun.NewDB(sqlDb, dialect, opts.BunOptions...)

	if opts.EnableQueryHook {
		addQueryHook(db, opts.Logger)
	}

	db = db.WithNamedArg(constants.PlaceholderKeyOperator, constants.OperatorSystem)

	return db
}

func configureConnectionPool(sqlDb *sql.DB, opts *databaseOptions) {
	if opts.PoolConfig != nil {
		opts.PoolConfig.ApplyToDB(sqlDb)
	}
}

func initializeDatabase(sqlDb *sql.DB, dialect schema.Dialect, opts *databaseOptions) (*bun.DB, error) {
	db := setupBunDB(sqlDb, dialect, opts)

	configureConnectionPool(sqlDb, opts)

	return db, nil
}

func New(config *config.DatasourceConfig, options ...Option) (*bun.DB, error) {
	provider, exists := registry.provider(config.Type)
	if !exists {
		return nil, newUnsupportedDbTypeError(config.Type)
	}

	sqlDb, dialect, err := provider.Connect(config)
	if err != nil || sqlDb == nil {
		return nil, err
	}

	opts := newDefaultOptions(config)
	opts.apply(options...)

	return initializeDatabase(sqlDb, dialect, opts)
}
