package database

import (
	"database/sql"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/schema"

	"github.com/ilxqx/vef-framework-go/config"
	"github.com/ilxqx/vef-framework-go/constants"
	logPkg "github.com/ilxqx/vef-framework-go/log"
)

// logDbVersion logs the version of the database using the provider's QueryVersion method.
func logDbVersion(provider DatabaseProvider, db *bun.DB, logger logPkg.Logger) error {
	version, err := provider.QueryVersion(db)
	if err != nil {
		return wrapVersionQueryError(provider.Type(), err)
	}

	logger.Infof("Database type: %s | Database version: %s", provider.Type(), version)

	return nil
}

// setupBunDB creates and configures a bun.DB instance with the provided SQL database and dialect.
func setupBunDB(sqlDb *sql.DB, dialect schema.Dialect, opts *databaseOptions) *bun.DB {
	db := bun.NewDB(sqlDb, dialect, opts.BunOptions...)

	if opts.EnableQueryHook {
		addQueryHook(db, opts.Logger)
	}

	db = db.WithNamedArg(constants.PlaceholderKeyOperator, constants.OperatorSystem)

	return db
}

// configureConnectionPool applies connection pool configuration to the SQL database.
func configureConnectionPool(sqlDb *sql.DB, opts *databaseOptions) {
	if opts.PoolConfig != nil {
		opts.PoolConfig.ApplyToDB(sqlDb)
	}
}

// initializeDatabase performs the complete database initialization process.
func initializeDatabase(sqlDb *sql.DB, dialect schema.Dialect, opts *databaseOptions) (*bun.DB, error) {
	// Setup bun.DB instance
	db := setupBunDB(sqlDb, dialect, opts)

	// Configure connection pool
	configureConnectionPool(sqlDb, opts)

	return db, nil
}

// New creates a new *bun.DB instance with custom options.
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
