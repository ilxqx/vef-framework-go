package cron

import (
	"time"

	"github.com/go-co-op/gocron/v2"
)

// Job represents a scheduled task in the cron system.
// It provides methods to inspect and control individual job instances.
type Job interface {
	// Id returns the job's unique identifier as a string.
	Id() string
	// LastRun returns the time when the job was last executed.
	LastRun() (time.Time, error)
	// Name returns the human-readable name assigned to the job.
	Name() string
	// NextRun returns the time when the job is next scheduled to run.
	NextRun() (time.Time, error)
	// NextRuns returns the specified number of future scheduled run times.
	NextRuns(count int) ([]time.Time, error)
	// RunNow executes the job immediately without affecting its regular schedule.
	// This respects all job and scheduler limits and may affect future scheduling
	// if the job has run limits configured.
	RunNow() error
	// Tags returns the list of tags associated with the job for grouping and filtering.
	Tags() []string
}

// JobDefinition defines how a job should be scheduled and executed.
// Implementations specify different scheduling strategies (cron, duration, one-time, etc.).
type JobDefinition interface {
	// build converts the high-level job definition into gocron-specific components.
	// This is an internal method used by the scheduler implementation.
	build() (gocron.JobDefinition, gocron.Task, []gocron.JobOption, error)
}

// Scheduler manages the lifecycle and execution of cron jobs.
// It provides a high-level interface for job scheduling, management, and control.
type Scheduler interface {
	// Jobs returns all jobs currently registered with the scheduler.
	Jobs() []Job
	// NewJob creates and registers a new job with the scheduler.
	// The job will be scheduled according to its definition when the scheduler is running.
	// If the task function accepts a context.Context as its first parameter,
	// the scheduler will provide cancellation support for graceful shutdown.
	NewJob(definition JobDefinition) (Job, error)
	// RemoveByTags removes all jobs that have any of the specified tags.
	RemoveByTags(tags ...string)
	// RemoveJob removes the job with the specified unique identifier.
	RemoveJob(id string) error
	// Start begins scheduling and executing jobs according to their definitions.
	// Jobs added to a running scheduler are scheduled immediately. This method is non-blocking.
	Start()
	// StopJobs stops the execution of all jobs without removing them from the scheduler.
	// Jobs can be restarted by calling Start() again.
	StopJobs() error
	// Update replaces an existing job's definition while preserving its unique identifier.
	// This allows for dynamic job reconfiguration without losing job history.
	Update(id string, definition JobDefinition) (Job, error)
	// JobsWaitingInQueue returns the number of jobs waiting in the execution queue.
	// This is only relevant when using LimitModeWait; otherwise it returns zero.
	JobsWaitingInQueue() int
}
