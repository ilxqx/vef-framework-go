package cron

import (
	"context"
	"time"
)

// JobDescriptorOption is a function type for configuring job descriptors using the options pattern.
// This allows for flexible and extensible job configuration.
type JobDescriptorOption func(*jobDescriptor)

// WithName sets the name of the job for identification and logging purposes.
func WithName(name string) JobDescriptorOption {
	return func(d *jobDescriptor) {
		d.name = name
	}
}

// WithTags assigns tags to the job for grouping and bulk operations.
func WithTags(tags ...string) JobDescriptorOption {
	return func(d *jobDescriptor) {
		d.tags = tags
	}
}

// WithConcurrent allows the job to run concurrently with other instances of itself.
// By default, jobs run in singleton mode (no concurrent execution).
func WithConcurrent() JobDescriptorOption {
	return func(d *jobDescriptor) {
		d.allowConcurrent = true
	}
}

// WithStartAt specifies when the job should start its schedule.
func WithStartAt(startAt time.Time) JobDescriptorOption {
	return func(d *jobDescriptor) {
		d.startAt = startAt
	}
}

// WithStartImmediately makes the job start immediately when the scheduler starts.
func WithStartImmediately() JobDescriptorOption {
	return func(d *jobDescriptor) {
		d.startImmediately = true
	}
}

// WithStopAt specifies when the job should stop running.
func WithStopAt(stopAt time.Time) JobDescriptorOption {
	return func(d *jobDescriptor) {
		d.stopAt = stopAt
	}
}

// WithLimitedRuns limits the job to run only the specified number of times.
func WithLimitedRuns(limitedRuns uint) JobDescriptorOption {
	return func(d *jobDescriptor) {
		d.limitedRuns = limitedRuns
	}
}

// WithContext sets a custom context for the job execution.
// If the context is canceled, the job will be canceled as well.
func WithContext(ctx context.Context) JobDescriptorOption {
	return func(d *jobDescriptor) {
		d.ctx = ctx
	}
}

// WithTask sets the handler function and its parameters for the job.
// The handler must be a function, and params are the arguments to pass to it.
func WithTask(handler any, params ...any) JobDescriptorOption {
	return func(d *jobDescriptor) {
		d.handler = handler
		d.params = params
	}
}
