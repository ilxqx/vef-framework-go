package cron

import (
	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
	"github.com/samber/lo"
)

// schedulerAdapter implements the Scheduler interface by adapting a gocron.Scheduler.
// It provides a clean abstraction layer over the underlying gocron scheduler.
type schedulerAdapter struct {
	scheduler gocron.Scheduler
}

func (s *schedulerAdapter) Jobs() []Job {
	return lo.Map(
		s.scheduler.Jobs(),
		func(job gocron.Job, _ int) Job {
			return &jobAdapter{job: job}
		},
	)
}

func (s *schedulerAdapter) NewJob(definition JobDefinition) (Job, error) {
	def, task, options, err := definition.build()
	if err != nil {
		return nil, err
	}

	job, err := s.scheduler.NewJob(def, task, options...)
	if err != nil {
		return nil, err
	}

	return &jobAdapter{job: job}, nil
}

func (s *schedulerAdapter) RemoveByTags(tags ...string) {
	s.scheduler.RemoveByTags(tags...)
}

func (s *schedulerAdapter) RemoveJob(id string) error {
	uuid, err := uuid.Parse(id)
	if err != nil {
		return err
	}

	return s.scheduler.RemoveJob(uuid)
}

func (s *schedulerAdapter) Start() {
	s.scheduler.Start()
}

func (s *schedulerAdapter) StopJobs() error {
	return s.scheduler.StopJobs()
}

func (s *schedulerAdapter) Update(id string, definition JobDefinition) (Job, error) {
	uuid, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	def, task, options, err := definition.build()
	if err != nil {
		return nil, err
	}

	job, err := s.scheduler.Update(uuid, def, task, options...)
	if err != nil {
		return nil, err
	}

	return &jobAdapter{job: job}, nil
}

func (s *schedulerAdapter) JobsWaitingInQueue() int {
	return s.scheduler.JobsWaitingInQueue()
}

// NewScheduler creates a new Scheduler implementation wrapping the provided gocron.Scheduler.
// This is the main entry point for creating scheduler instances in the application.
func NewScheduler(scheduler gocron.Scheduler) Scheduler {
	return &schedulerAdapter{
		scheduler: scheduler,
	}
}
