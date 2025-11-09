package taskpool

import "context"

// WorkerDelegate defines pluggable task execution logic.
// Each worker owns its delegate instance, all methods run in worker's OS thread.
type WorkerDelegate[TIn, TOut any] interface {
	// Init is called once when worker starts.
	Init(ctx context.Context, config any) error

	// Execute runs for each task. Must respect context cancellation.
	Execute(ctx context.Context, payload TIn) (TOut, error)

	// Destroy is called once when worker stops.
	Destroy() error

	// HealthCheck is called periodically.
	HealthCheck() error
}

// Scheduler manages task submission and execution.
type Scheduler[TIn, TOut any] interface {
	// Submit blocks until task completes or context is canceled.
	Submit(ctx context.Context, payload TIn, opts ...SubmitOption) (Result[TOut], error)

	// SubmitAsync returns immediately with a result channel.
	SubmitAsync(ctx context.Context, payload TIn, opts ...SubmitOption) (<-chan Result[TOut], error)

	Stats() SchedulerStats

	// Shutdown waits for running tasks to complete.
	Shutdown(ctx context.Context) error
}
