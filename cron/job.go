package cron

import (
	"time"

	"github.com/go-co-op/gocron/v2"
)

// jobAdapter adapts gocron.Job to implement the framework's Job interface.
// It provides a clean abstraction layer over the underlying gocron job.
type jobAdapter struct {
	job gocron.Job
}

func (j *jobAdapter) Id() string {
	return j.job.ID().String()
}

func (j *jobAdapter) LastRun() (time.Time, error) {
	return j.job.LastRun()
}

func (j *jobAdapter) Name() string {
	return j.job.Name()
}

func (j *jobAdapter) NextRun() (time.Time, error) {
	return j.job.NextRun()
}

func (j *jobAdapter) NextRuns(count int) ([]time.Time, error) {
	return j.job.NextRuns(count)
}

func (j *jobAdapter) RunNow() error {
	return j.job.RunNow()
}

func (j *jobAdapter) Tags() []string {
	return j.job.Tags()
}
