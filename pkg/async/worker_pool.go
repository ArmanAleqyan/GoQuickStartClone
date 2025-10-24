package async

import (
	"context"
	"log"
	"sync"
	"time"
)

// Task represents a unit of work
type Task func(ctx context.Context) error

// WorkerPool manages a pool of worker goroutines
type WorkerPool struct {
	workerCount int
	taskQueue   chan Task
	wg          sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
	stats       *PoolStats
	mu          sync.RWMutex
}

// PoolStats holds worker pool statistics
type PoolStats struct {
	TasksProcessed   int64
	TasksFailed      int64
	ActiveWorkers    int
	TotalWorkers     int
	QueueLength      int
	QueueCapacity    int
}

// NewWorkerPool creates a new worker pool
func NewWorkerPool(workerCount, queueSize int) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())

	pool := &WorkerPool{
		workerCount: workerCount,
		taskQueue:   make(chan Task, queueSize),
		ctx:         ctx,
		cancel:      cancel,
		stats: &PoolStats{
			TotalWorkers:  workerCount,
			QueueCapacity: queueSize,
		},
	}

	pool.start()

	return pool
}

// start spawns worker goroutines
func (p *WorkerPool) start() {
	for i := 0; i < p.workerCount; i++ {
		p.wg.Add(1)
		go p.worker(i)
	}
	log.Printf("[WorkerPool] Started %d workers", p.workerCount)
}

// worker processes tasks from the queue
func (p *WorkerPool) worker(id int) {
	defer p.wg.Done()

	log.Printf("[WorkerPool] Worker %d started", id)

	p.mu.Lock()
	p.stats.ActiveWorkers++
	p.mu.Unlock()

	defer func() {
		p.mu.Lock()
		p.stats.ActiveWorkers--
		p.mu.Unlock()
	}()

	for {
		select {
		case <-p.ctx.Done():
			log.Printf("[WorkerPool] Worker %d shutting down", id)
			return
		case task, ok := <-p.taskQueue:
			if !ok {
				log.Printf("[WorkerPool] Worker %d channel closed", id)
				return
			}
			p.executeTask(task)
		}
	}
}

// executeTask runs a task with timeout and error handling
func (p *WorkerPool) executeTask(task Task) {
	// Create task context with timeout
	taskCtx, cancel := context.WithTimeout(p.ctx, 30*time.Second)
	defer cancel()

	// Execute task
	err := task(taskCtx)

	// Update stats
	p.mu.Lock()
	if err != nil {
		p.stats.TasksFailed++
		log.Printf("[WorkerPool] Task failed: %v", err)
	} else {
		p.stats.TasksProcessed++
	}
	p.mu.Unlock()
}

// Submit submits a task to the worker pool
func (p *WorkerPool) Submit(task Task) bool {
	select {
	case p.taskQueue <- task:
		return true
	case <-time.After(100 * time.Millisecond):
		log.Printf("[WorkerPool] Queue full, task rejected")
		return false
	}
}

// SubmitWithTimeout submits a task with custom timeout
func (p *WorkerPool) SubmitWithTimeout(task Task, timeout time.Duration) bool {
	select {
	case p.taskQueue <- task:
		return true
	case <-time.After(timeout):
		log.Printf("[WorkerPool] Timeout waiting for queue")
		return false
	}
}

// SubmitBlocking submits a task and blocks until accepted
func (p *WorkerPool) SubmitBlocking(task Task) {
	p.taskQueue <- task
}

// Shutdown gracefully shuts down the worker pool
func (p *WorkerPool) Shutdown(timeout time.Duration) error {
	log.Printf("[WorkerPool] Shutting down...")

	// Stop accepting new tasks
	close(p.taskQueue)

	// Wait for workers with timeout
	done := make(chan struct{})
	go func() {
		p.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Printf("[WorkerPool] All workers finished")
		return nil
	case <-time.After(timeout):
		p.cancel()
		log.Printf("[WorkerPool] Shutdown timeout, forcing stop")
		return nil
	}
}

// Stats returns current pool statistics
func (p *WorkerPool) Stats() PoolStats {
	p.mu.RLock()
	defer p.mu.RUnlock()

	stats := *p.stats
	stats.QueueLength = len(p.taskQueue)

	return stats
}

// IsHealthy checks if the pool is healthy
func (p *WorkerPool) IsHealthy() bool {
	stats := p.Stats()

	// Pool is unhealthy if queue is 90% full
	queueUsage := float64(stats.QueueLength) / float64(stats.QueueCapacity)
	if queueUsage > 0.9 {
		return false
	}

	// Pool is unhealthy if no active workers
	if stats.ActiveWorkers == 0 {
		return false
	}

	return true
}

// Resize dynamically adjusts the number of workers
func (p *WorkerPool) Resize(newWorkerCount int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	currentCount := p.workerCount
	if newWorkerCount > currentCount {
		// Add workers
		for i := 0; i < (newWorkerCount - currentCount); i++ {
			p.wg.Add(1)
			go p.worker(currentCount + i)
		}
		log.Printf("[WorkerPool] Added %d workers (total: %d)", newWorkerCount-currentCount, newWorkerCount)
	}

	p.workerCount = newWorkerCount
	p.stats.TotalWorkers = newWorkerCount
}
