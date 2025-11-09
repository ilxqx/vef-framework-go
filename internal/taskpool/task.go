package taskpool

import (
	"context"
	"time"

	"github.com/rs/xid"
)

type Priority int

const (
	PriorityLow    Priority = 0
	PriorityMedium Priority = 1
	PriorityHigh   Priority = 2
)

// Task is a unit of work executed by a worker.
type Task[TIn, TOut any] struct {
	Id          string
	Context     context.Context
	Priority    Priority
	Payload     TIn
	Result      chan<- Result[TOut]
	Done        chan struct{} // Closed by worker after execution
	SubmittedAt time.Time
}

// Result contains the outcome of task execution.
type Result[TOut any] struct {
	TaskId   string
	Data     TOut
	Error    error
	Duration time.Duration
}

func generateTaskId() string {
	return xid.New().String()
}
