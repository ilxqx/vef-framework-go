package param

import (
	"reflect"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/cron"
)

type SchedulerFactoryResolver struct {
	scheduler cron.Scheduler
}

func NewSchedulerFactoryResolver(scheduler cron.Scheduler) api.FactoryParamResolver {
	return &SchedulerFactoryResolver{scheduler: scheduler}
}

func (r *SchedulerFactoryResolver) Type() reflect.Type {
	return reflect.TypeFor[cron.Scheduler]()
}

func (r *SchedulerFactoryResolver) Resolve() (reflect.Value, error) {
	return reflect.ValueOf(r.scheduler), nil
}
