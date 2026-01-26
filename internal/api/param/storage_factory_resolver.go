package param

import (
	"reflect"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/storage"
)

type StorageFactoryResolver struct {
	service storage.Service
}

func NewStorageFactoryResolver(service storage.Service) api.FactoryParamResolver {
	return &StorageFactoryResolver{service: service}
}

func (r *StorageFactoryResolver) Type() reflect.Type {
	return reflect.TypeFor[storage.Service]()
}

func (r *StorageFactoryResolver) Resolve() (reflect.Value, error) {
	return reflect.ValueOf(r.service), nil
}
