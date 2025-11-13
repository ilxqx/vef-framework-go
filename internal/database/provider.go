package database

import (
	"database/sql"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/schema"

	"github.com/ilxqx/vef-framework-go/config"
	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/internal/database/mysql"
	"github.com/ilxqx/vef-framework-go/internal/database/postgres"
	"github.com/ilxqx/vef-framework-go/internal/database/sqlite"
)

type DatabaseProvider interface {
	Connect(config *config.DatasourceConfig) (*sql.DB, schema.Dialect, error)
	Type() constants.DbType
	ValidateConfig(config *config.DatasourceConfig) error
	QueryVersion(db *bun.DB) (string, error)
}

type providerRegistry struct {
	providers map[constants.DbType]DatabaseProvider
}

func newProviderRegistry() *providerRegistry {
	registry := &providerRegistry{
		providers: make(map[constants.DbType]DatabaseProvider),
	}

	registry.register(sqlite.NewProvider())
	registry.register(postgres.NewProvider())
	registry.register(mysql.NewProvider())

	return registry
}

func (r *providerRegistry) register(provider DatabaseProvider) {
	r.providers[provider.Type()] = provider
}

func (r *providerRegistry) provider(dbType constants.DbType) (DatabaseProvider, bool) {
	provider, exists := r.providers[dbType]

	return provider, exists
}

var registry = newProviderRegistry()
