package cron

import (
	"reflect"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/cron"
)

// SchedulerFactoryParamResolver provides cron.Scheduler for handler factory functions.
type SchedulerFactoryParamResolver struct {
	scheduler cron.Scheduler
}

// NewSchedulerFactoryParamResolver creates a new SchedulerFactoryParamResolver.
func NewSchedulerFactoryParamResolver(scheduler cron.Scheduler) api.FactoryParamResolver {
	return &SchedulerFactoryParamResolver{scheduler: scheduler}
}

// Type returns the type this resolver handles.
func (r *SchedulerFactoryParamResolver) Type() reflect.Type {
	return reflect.TypeFor[cron.Scheduler]()
}

// Resolve returns the scheduler instance.
func (r *SchedulerFactoryParamResolver) Resolve() reflect.Value {
	return reflect.ValueOf(r.scheduler)
}
