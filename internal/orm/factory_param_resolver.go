package orm

import (
	"reflect"

	"github.com/ilxqx/vef-framework-go/api"
)

// DbFactoryParamResolver provides Db for handler factory functions.
type DbFactoryParamResolver struct {
	db Db
}

// NewDbFactoryParamResolver creates a new DbFactoryParamResolver.
func NewDbFactoryParamResolver(db Db) api.FactoryParamResolver {
	return &DbFactoryParamResolver{db: db}
}

// Type returns the type this resolver handles.
func (r *DbFactoryParamResolver) Type() reflect.Type {
	return reflect.TypeFor[Db]()
}

// Resolve returns the database instance.
func (r *DbFactoryParamResolver) Resolve() reflect.Value {
	return reflect.ValueOf(r.db)
}
