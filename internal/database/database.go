package database

import (
	"fmt"
	"runtime"
	"time"

	"github.com/ilxqx/vef-framework-go/config"
	"github.com/ilxqx/vef-framework-go/internal/log"
	logPkg "github.com/ilxqx/vef-framework-go/log"
	"github.com/ilxqx/vef-framework-go/orm"
	"github.com/uptrace/bun"
	"go.uber.org/fx"
)

var logger = log.Named("database") // logger is the database module logger

// newDb creates a new Db
func newDb(lc fx.Lifecycle, config *config.DatasourceConfig) (*bun.DB, error) {
	sqlDb, dialect, err := selectDbDriver(config)
	if err != nil || sqlDb == nil {
		return nil, err
	}

	db := bun.NewDB(sqlDb, dialect, bun.WithDiscardUnknownColumns())
	addQueryHook(db)

	db = db.WithNamedArg(orm.PlaceholderKeyOperator, orm.OperatorSystem)

	// Setup database connection pool
	sqlDb.SetMaxIdleConns(max(runtime.GOMAXPROCS(0)*4, 10))
	sqlDb.SetMaxOpenConns(max(runtime.GOMAXPROCS(0)*8, 50))
	sqlDb.SetConnMaxIdleTime(5 * time.Minute)
	sqlDb.SetConnMaxLifetime(12 * time.Hour)

	if err := sqlDb.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}
	logger.Info("Successfully connected to database")

	if err := logDbVersion(config.Type, db, logger); err != nil {
		return nil, err
	}

	lc.Append(fx.StopHook(db.Close))

	return db, nil
}

// logDbVersion logs the version of the database.
func logDbVersion(dbType string, db *bun.DB, logger logPkg.Logger) error {
	version, err := queryVersion(dbType, db)
	if err != nil {
		return fmt.Errorf("failed to query database version: %v", err)
	}

	logger.Infof("Database type: %s | Database version: %s", dbType, version)
	return nil
}
