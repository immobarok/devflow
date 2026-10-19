package worker

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// Job represents a background task to be processed asynchronously.
type Job func(ctx context.Context) error

// Pool manages a pool of worker goroutines consuming jobs from a buffered channel.
type Pool struct {
	workerCount int
	jobQueue    chan Job
	wg          sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
	closed      bool
	mu          sync.Mutex
}

// NewPool initializes a new worker pool with a specified number of workers and job queue capacity.
func NewPool(workerCount, queueSize int) *Pool {
	if workerCount <= 0 {
		workerCount = 3
	}
	if queueSize <= 0 {
		queueSize = 100
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Pool{
		workerCount: workerCount,
		jobQueue:    make(chan Job, queueSize),
		ctx:         ctx,
		cancel:      cancel,
	}
}

// Start launches the background worker goroutines.
func (p *Pool) Start() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return
	}

	for i := 1; i <= p.workerCount; i++ {
		p.wg.Add(1)
		go p.worker(i)
	}

	slog.Info("background worker pool started", "workers", p.workerCount, "queue_capacity", cap(p.jobQueue))
}

func (p *Pool) worker(id int) {
	defer p.wg.Done()

	for {
		select {
		case <-p.ctx.Done():
			return
		case job, ok := <-p.jobQueue:
			if !ok {
				return
			}
			p.executeJob(id, job)
		}
	}
}

func (p *Pool) executeJob(workerID int, job Job) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("worker recovered from panic", "worker_id", workerID, "panic", r)
		}
	}()

	jobCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := job(jobCtx); err != nil {
		slog.Error("background job failed", "worker_id", workerID, "error", err)
	}
}

// Submit enqueues a job into the worker pool. It is non-blocking.
// Returns false if the queue is full or the pool is stopped.
func (p *Pool) Submit(job Job) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		slog.Warn("attempted to submit job to a closed worker pool")
		return false
	}

	select {
	case p.jobQueue <- job:
		return true
	default:
		slog.Warn("worker pool job queue is full; dropping job")
		return false
	}
}

// Stop gracefully shuts down the worker pool, draining all pending jobs.
func (p *Pool) Stop() {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return
	}
	p.closed = true
	close(p.jobQueue)
	p.mu.Unlock()

	slog.Info("draining and shutting down worker pool...")
	p.wg.Wait()
	p.cancel()
	slog.Info("worker pool stopped cleanly")
}
