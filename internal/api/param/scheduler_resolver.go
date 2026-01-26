package param

import (
	"reflect"

	"github.com/gofiber/fiber/v3"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/cron"
)

type SchedulerResolver struct {
	scheduler cron.Scheduler
}

func NewSchedulerResolver(scheduler cron.Scheduler) api.HandlerParamResolver {
	return &SchedulerResolver{scheduler: scheduler}
}

func (r *SchedulerResolver) Type() reflect.Type {
	return reflect.TypeFor[cron.Scheduler]()
}

func (r *SchedulerResolver) Resolve(fiber.Ctx) (reflect.Value, error) {
	return reflect.ValueOf(r.scheduler), nil
}
