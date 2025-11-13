package cron

import (
	"strings"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"

	"github.com/ilxqx/vef-framework-go/constants"
)

// jobMonitor implements gocron.Monitor interface to track job execution metrics. It provides detailed logging for job lifecycle events including timing and status.
type jobMonitor struct{}

func (m *jobMonitor) RecordJobTimingWithStatus(startTime, endTime time.Time, id uuid.UUID, name string, tags []string, status gocron.JobStatus, err error) {
	switch status {
	case gocron.Success:
		logger.Infof(
			"Job %q completed | id: %s | tags: %s | elapsed: %s | status: %s",
			name,
			id.String(),
			strings.Join(tags, constants.CommaSpace),
			endTime.Sub(startTime),
			status,
		)

	case gocron.Fail:
		logger.Errorf(
			"Job %q completed | id: %s | tags: %s | elapsed: %s | status: %s | error: %v",
			name,
			id.String(),
			strings.Join(tags, constants.CommaSpace),
			endTime.Sub(startTime),
			status,
			err,
		)

	default:
		logger.Warnf(
			"Job %q completed | id: %s | tags: %s | elapsed: %s | status: %s",
			name,
			id.String(),
			strings.Join(tags, constants.CommaSpace),
			endTime.Sub(startTime),
			status,
		)
	}
}

func (m *jobMonitor) IncrementJob(id uuid.UUID, name string, tags []string, status gocron.JobStatus) {
	logger.Infof(
		"Job %q scheduled | id: %s | tags: %s | status: %s",
		name,
		id.String(),
		strings.Join(tags, constants.CommaSpace),
		status,
	)
}

func (m *jobMonitor) RecordJobTiming(startTime, endTime time.Time, id uuid.UUID, name string, tags []string) {
	logger.Infof(
		"Job %q completed | id: %s | tags: %s | elapsed: %s",
		name,
		id.String(),
		strings.Join(tags, constants.CommaSpace),
		endTime.Sub(startTime),
	)
}

func newJobMonitor() *jobMonitor {
	return new(jobMonitor)
}
