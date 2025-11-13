package cron

import (
	"reflect"

	"github.com/gofiber/fiber/v3"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/cron"
)

// SchedulerHandlerParamResolver resolves cron.Scheduler for handler parameters.
type SchedulerHandlerParamResolver struct {
	scheduler cron.Scheduler
}

// NewSchedulerHandlerParamResolver creates a new cron scheduler parameter resolver.
func NewSchedulerHandlerParamResolver(scheduler cron.Scheduler) api.HandlerParamResolver {
	return &SchedulerHandlerParamResolver{scheduler: scheduler}
}

func (r *SchedulerHandlerParamResolver) Type() reflect.Type {
	return reflect.TypeFor[cron.Scheduler]()
}

func (r *SchedulerHandlerParamResolver) Resolve(fiber.Ctx) (reflect.Value, error) {
	return reflect.ValueOf(r.scheduler), nil
}
