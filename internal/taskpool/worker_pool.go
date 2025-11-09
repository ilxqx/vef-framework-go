package taskpool

import (
	"context"
	"fmt"
	"sync"

	"github.com/ilxqx/vef-framework-go/log"
)

// WorkerPool manages elastic worker scaling with priority queues.
type WorkerPool[TIn, TOut any] struct {
	config Config[TIn, TOut]
	logger log.Logger

	workers  []*Worker[TIn, TOut]
	workerMu sync.RWMutex
	wg       sync.WaitGroup

	highQueue   chan *Task[TIn, TOut]
	mediumQueue chan *Task[TIn, TOut]
	lowQueue    chan *Task[TIn, TOut]

	shutdown chan struct{}
	once     sync.Once

	stats poolStats
}

func newWorkerPool[TIn, TOut any](config Config[TIn, TOut]) (*WorkerPool[TIn, TOut], error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	pool := &WorkerPool[TIn, TOut]{
		config:      config,
		logger:      config.Logger.Named("pool"),
		workers:     make([]*Worker[TIn, TOut], 0, config.MaxWorkers),
		highQueue:   make(chan *Task[TIn, TOut], config.TaskQueueSize),
		mediumQueue: make(chan *Task[TIn, TOut], config.TaskQueueSize),
		lowQueue:    make(chan *Task[TIn, TOut], config.TaskQueueSize),
		shutdown:    make(chan struct{}),
	}

	for i := 0; i < config.MinWorkers; i++ {
		if err := pool.addWorker(); err != nil {
			_ = pool.Shutdown(context.Background())

			return nil, fmt.Errorf("failed to start worker %d: %w", i, err)
		}
	}

	pool.logger.Infof("worker pool started: min_workers=%d, max_workers=%d, queue_size=%d",
		config.MinWorkers, config.MaxWorkers, config.TaskQueueSize)

	return pool, nil
}

func (p *WorkerPool[TIn, TOut]) addWorker() error {
	p.workerMu.Lock()
	defer p.workerMu.Unlock()

	if len(p.workers) >= p.config.MaxWorkers {
		return ErrMaxWorkersReached
	}

	worker := newWorker(len(p.workers), p)

	p.wg.Go(func() {
		worker.run()
	})

	if err := <-worker.initDone; err != nil {
		return fmt.Errorf("worker initialization failed: %w", err)
	}

	p.workers = append(p.workers, worker)
	p.logger.Debugf("worker added: worker_id=%d, total_workers=%d", worker.id, len(p.workers))

	return nil
}

func (p *WorkerPool[TIn, TOut]) removeWorker(w *Worker[TIn, TOut]) {
	p.workerMu.Lock()
	defer p.workerMu.Unlock()

	for i, worker := range p.workers {
		if worker == w {
			p.workers[i] = p.workers[len(p.workers)-1]
			p.workers = p.workers[:len(p.workers)-1]
			p.logger.Debugf("worker removed: worker_id=%d, total_workers=%d", w.id, len(p.workers))

			return
		}
	}
}

// scaleIfNeeded adds workers when queue depth > workers * 2.
func (p *WorkerPool[TIn, TOut]) scaleIfNeeded() {
	queueDepth := len(p.highQueue) + len(p.mediumQueue) + len(p.lowQueue)

	p.workerMu.RLock()
	currentWorkers := len(p.workers)
	p.workerMu.RUnlock()

	if queueDepth > currentWorkers*2 && currentWorkers < p.config.MaxWorkers {
		if err := p.addWorker(); err != nil {
			p.logger.Warnf("failed to scale up: %v", err)
		} else {
			p.logger.Infof("scaled up workers: workers=%d, queue_depth=%d",
				currentWorkers+1, queueDepth)
		}
	}
}

func (p *WorkerPool[TIn, TOut]) submit(task *Task[TIn, TOut]) error {
	select {
	case <-p.shutdown:
		return ErrPoolShutdown
	default:
	}

	p.stats.totalSubmitted.Add(1)

	var queue chan *Task[TIn, TOut]
	switch task.Priority {
	case PriorityHigh:
		queue = p.highQueue
	case PriorityMedium:
		queue = p.mediumQueue
	case PriorityLow:
		queue = p.lowQueue
	default:
		return ErrInvalidPriority
	}

	select {
	case queue <- task:
		p.scaleIfNeeded()

		return nil
	case <-task.Context.Done():
		return task.Context.Err()
	case <-p.shutdown:
		return ErrPoolShutdown
	default:
		return ErrQueueFull
	}
}

func (p *WorkerPool[TIn, TOut]) Shutdown(ctx context.Context) error {
	var shutdownErr error

	p.once.Do(func() {
		p.logger.Info("shutting down worker pool")

		close(p.shutdown)
		close(p.highQueue)
		close(p.mediumQueue)
		close(p.lowQueue)

		p.workerMu.RLock()

		for _, worker := range p.workers {
			worker.stop()
		}

		p.workerMu.RUnlock()

		done := make(chan struct{})
		go func() {
			p.wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			p.logger.Info("worker pool shutdown complete")
		case <-ctx.Done():
			shutdownErr = fmt.Errorf("shutdown timeout: %w", ctx.Err())
			p.logger.Errorf("worker pool shutdown timeout: %v", shutdownErr)
		}
	})

	return shutdownErr
}

func (p *WorkerPool[TIn, TOut]) getStats() SchedulerStats {
	p.workerMu.RLock()
	totalWorkers := len(p.workers)
	p.workerMu.RUnlock()

	queuedTasks := len(p.highQueue) + len(p.mediumQueue) + len(p.lowQueue)

	return p.stats.snapshot(totalWorkers, queuedTasks)
}
