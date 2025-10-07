package database

import (
	"context"
	"database/sql"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/schema"
	"go.uber.org/fx"

	"github.com/ilxqx/vef-framework-go/config"
	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/internal/log"
	logPkg "github.com/ilxqx/vef-framework-go/log"
)

var logger = log.Named("database")

// newDb creates a new Db with lifecycle management.
func newDb(lc fx.Lifecycle, config *config.DatasourceConfig) (db *bun.DB, err error) {
	// Create database without validation and version logging
	// These will be done in the StartHook
	opts := []Option{
		WithQueryHook(true),
	}

	if db, err = CreateDb(config, opts...); err != nil {
		return db, err
	}

	// Get provider for StartHook validation
	provider, exists := registry.provider(config.Type)
	if !exists {
		return nil, newUnsupportedDbTypeError(config.Type)
	}

	// Register lifecycle hooks for proper startup and shutdown
	lc.Append(
		fx.StartStopHook(
			func(ctx context.Context) error {
				// Validate connection
				if err := db.PingContext(ctx); err != nil {
					return wrapPingError(provider.Type(), err)
				}

				// Log database version
				if err := logDbVersion(provider, db, logger); err != nil {
					return err
				}

				logger.Infof("Database client started successfully: %s", provider.Type())

				return nil
			},
			func() error {
				logger.Info("Closing database connection...")

				return db.Close()
			},
		),
	)

	return db, err
}

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

// CreateDb creates a new *bun.DB instance with custom options.
//
// Parameters:
//   - config: The datasource configuration containing database connection details
//   - options: Variadic list of Option functions to customize the database initialization
//
// Returns:
//   - *bun.DB: A configured bun database instance ready for use
//   - error: Any error that occurred during database initialization
func CreateDb(config *config.DatasourceConfig, options ...Option) (*bun.DB, error) {
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
