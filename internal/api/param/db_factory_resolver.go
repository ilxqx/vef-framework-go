package param

import (
	"reflect"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/orm"
)

type DBFactoryResolver struct {
	db orm.DB
}

func NewDBFactoryResolver(db orm.DB) api.FactoryParamResolver {
	return &DBFactoryResolver{db: db}
}

func (r *DBFactoryResolver) Type() reflect.Type {
	return reflect.TypeFor[orm.DB]()
}

func (r *DBFactoryResolver) Resolve() (reflect.Value, error) {
	return reflect.ValueOf(r.db), nil
}
