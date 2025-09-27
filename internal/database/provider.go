package database

import (
	"database/sql"

	"github.com/ilxqx/vef-framework-go/config"
	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/internal/database/mysql"
	"github.com/ilxqx/vef-framework-go/internal/database/postgres"
	"github.com/ilxqx/vef-framework-go/internal/database/sqlite"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/schema"
)

// DatabaseProvider defines the interface for database providers
type DatabaseProvider interface {
	// Connect establishes a connection to the database
	Connect(config *config.DatasourceConfig) (*sql.DB, schema.Dialect, error)

	// Type returns the database type identifier
	Type() constants.DbType

	// ValidateConfig validates the datasource configuration for this provider
	ValidateConfig(config *config.DatasourceConfig) error

	// QueryVersion queries the database version
	QueryVersion(db *bun.DB) (string, error)
}

// providerRegistry manages database providers
type providerRegistry struct {
	providers map[constants.DbType]DatabaseProvider
}

// newProviderRegistry creates a new provider registry with default providers
func newProviderRegistry() *providerRegistry {
	registry := &providerRegistry{
		providers: make(map[constants.DbType]DatabaseProvider),
	}

	// Register default providers
	registry.register(sqlite.NewProvider())
	registry.register(postgres.NewProvider())
	registry.register(mysql.NewProvider())

	return registry
}

// register registers a database provider
func (r *providerRegistry) register(provider DatabaseProvider) {
	r.providers[provider.Type()] = provider
}

// provider returns a provider by type
func (r *providerRegistry) provider(dbType constants.DbType) (DatabaseProvider, bool) {
	provider, exists := r.providers[dbType]
	return provider, exists
}

// Global provider registry instance
var registry = newProviderRegistry()
