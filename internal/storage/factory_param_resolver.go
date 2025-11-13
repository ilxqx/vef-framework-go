package storage

import (
	"reflect"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/storage"
)

type StorageServiceFactoryParamResolver struct {
	service storage.Service
}

func NewStorageServiceFactoryParamResolver(service storage.Service) api.FactoryParamResolver {
	return &StorageServiceFactoryParamResolver{service: service}
}

func (r *StorageServiceFactoryParamResolver) Type() reflect.Type {
	return reflect.TypeFor[storage.Service]()
}

func (r *StorageServiceFactoryParamResolver) Resolve() reflect.Value {
	return reflect.ValueOf(r.service)
}
