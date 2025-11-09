package taskpool

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ilxqx/vef-framework-go/log"
)

type workerState int

const (
	workerStateIdle     workerState = 0
	workerStateRunning  workerState = 1
	workerStateStopping workerState = 2
	workerStateStopped  workerState = 3
)

// Worker executes tasks in a dedicated OS thread.
type Worker[TIn, TOut any] struct {
	id       int
	pool     *WorkerPool[TIn, TOut]
	delegate WorkerDelegate[TIn, TOut]
	logger   log.Logger

	state      workerState
	stateMu    sync.RWMutex
	lastActive atomic.Value // time.Time

	stopCh   chan struct{}
	initDone chan error

	tasksExecuted atomic.Uint64
}

func newWorker[TIn, TOut any](id int, pool *WorkerPool[TIn, TOut]) *Worker[TIn, TOut] {
	w := &Worker[TIn, TOut]{
		id:       id,
		pool:     pool,
		delegate: pool.config.DelegateFactory(),
		logger:   pool.logger.Named(fmt.Sprintf("worker-%d", id)),
		state:    workerStateIdle,
		stopCh:   make(chan struct{}),
		initDone: make(chan error, 1),
	}
	w.lastActive.Store(time.Now())

	return w
}

// run is the main worker loop, locks to OS thread for delegate compatibility.
func (w *Worker[TIn, TOut]) run() {
	runtime.LockOSThread()

	defer runtime.UnlockOSThread()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := w.delegate.Init(ctx, w.pool.config.DelegateConfig); err != nil {
		w.logger.Errorf("delegate init failed: %v", err)
		w.setState(workerStateStopped)

		w.initDone <- err

		return
	}

	w.initDone <- nil

	w.logger.Debug("worker started")

	defer func() {
		if err := w.delegate.Destroy(); err != nil {
			w.logger.Errorf("delegate destroy failed: %v", err)
		}

		w.setState(workerStateStopped)
		w.logger.Debug("worker stopped")
	}()

	idleCheckTicker := time.NewTicker(1 * time.Second)
	defer idleCheckTicker.Stop()

	for {
		select {
		case <-w.stopCh:
			return
		case <-idleCheckTicker.C:
			if w.shouldStopDueToIdle() {
				return
			}
		default:
			task := w.fetchTask()
			if task == nil {
				time.Sleep(10 * time.Millisecond)

				continue
			}

			w.executeTask(task)
		}
	}
}

// fetchTask retrieves next task by priority: High -> Medium -> Low.
func (w *Worker[TIn, TOut]) fetchTask() *Task[TIn, TOut] {
	select {
	case task := <-w.pool.highQueue:
		return task
	default:
	}

	select {
	case task := <-w.pool.mediumQueue:
		return task
	default:
	}

	select {
	case task := <-w.pool.lowQueue:
		return task
	case <-time.After(50 * time.Millisecond):
		return nil
	}
}

func (w *Worker[TIn, TOut]) executeTask(task *Task[TIn, TOut]) {
	w.setState(workerStateRunning)
	defer func() {
		w.setState(workerStateIdle)
		w.lastActive.Store(time.Now())

		if task.Done != nil {
			close(task.Done)
		}
	}()

	start := time.Now()

	w.tasksExecuted.Add(1)

	w.pool.stats.activeWorkers.Add(1)
	defer w.pool.stats.activeWorkers.Add(-1)

	result := w.execute(task)
	result.Duration = time.Since(start)

	if result.Error != nil {
		w.pool.stats.totalFailed.Add(1)
	} else {
		w.pool.stats.totalCompleted.Add(1)
	}

	if task.Result != nil {
		select {
		case task.Result <- result:
		case <-task.Context.Done():
		}
	}

	w.logger.Debugf("task completed: id=%s, duration=%v, error=%v",
		task.Id, result.Duration, result.Error)
}

func (w *Worker[TIn, TOut]) execute(task *Task[TIn, TOut]) Result[TOut] {
	result := Result[TOut]{TaskId: task.Id}

	ctx := task.Context
	if deadline, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc

		ctx, cancel = context.WithTimeout(ctx, w.pool.config.TaskTimeout)
		defer cancel()
	} else {
		remaining := time.Until(deadline)
		if remaining > w.pool.config.MaxTaskTimeout {
			var cancel context.CancelFunc

			ctx, cancel = context.WithTimeout(context.Background(), w.pool.config.TaskTimeout)
			defer cancel()
		}
	}

	data, err := w.delegate.Execute(ctx, task.Payload)
	result.Data = data
	result.Error = err

	return result
}

func (w *Worker[TIn, TOut]) shouldStopDueToIdle() bool {
	if w.pool.config.IdleTimeout == 0 {
		return false
	}

	lastActive := w.lastActive.Load().(time.Time)
	if time.Since(lastActive) < w.pool.config.IdleTimeout {
		return false
	}

	w.pool.workerMu.RLock()
	canStop := len(w.pool.workers) > w.pool.config.MinWorkers
	w.pool.workerMu.RUnlock()

	if canStop {
		w.logger.Infof("worker stopping due to idle timeout: idle_duration=%v",
			time.Since(lastActive))
		w.pool.removeWorker(w)
	}

	return canStop
}

func (w *Worker[TIn, TOut]) stop() {
	w.setState(workerStateStopping)
	close(w.stopCh)
}

func (w *Worker[TIn, TOut]) setState(state workerState) {
	w.stateMu.Lock()
	defer w.stateMu.Unlock()

	w.state = state
}
