package storage

import (
	"reflect"

	"github.com/gofiber/fiber/v3"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/storage"
)

type StorageServiceHandlerParamResolver struct {
	service storage.Service
}

func NewStorageServiceHandlerParamResolver(service storage.Service) api.HandlerParamResolver {
	return &StorageServiceHandlerParamResolver{service: service}
}

func (r *StorageServiceHandlerParamResolver) Type() reflect.Type {
	return reflect.TypeFor[storage.Service]()
}

func (r *StorageServiceHandlerParamResolver) Resolve(fiber.Ctx) (reflect.Value, error) {
	return reflect.ValueOf(r.service), nil
}
