package param

import (
	"reflect"

	"github.com/gofiber/fiber/v3"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/storage"
)

type StorageResolver struct {
	service storage.Service
}

func NewStorageResolver(service storage.Service) api.HandlerParamResolver {
	return &StorageResolver{service: service}
}

func (*StorageResolver) Type() reflect.Type {
	return reflect.TypeFor[storage.Service]()
}

func (r *StorageResolver) Resolve(_ fiber.Ctx) (reflect.Value, error) {
	return reflect.ValueOf(r.service), nil
}
