package async

import (
	"context"
	"log"
	"sync"
	"time"

	"ironnode/pkg/models"

	"gorm.io/gorm"
)

// AsyncLogger provides asynchronous logging functionality
type AsyncLogger struct {
	db          *gorm.DB
	logChannel  chan *models.RequestLog
	workerCount int
	wg          sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
}

// NewAsyncLogger creates a new async logger with specified buffer size and worker count
func NewAsyncLogger(db *gorm.DB, bufferSize, workerCount int) *AsyncLogger {
	ctx, cancel := context.WithCancel(context.Background())

	logger := &AsyncLogger{
		db:          db,
		logChannel:  make(chan *models.RequestLog, bufferSize),
		workerCount: workerCount,
		ctx:         ctx,
		cancel:      cancel,
	}

	// Start worker goroutines
	logger.start()

	return logger
}

// start spawns worker goroutines
func (l *AsyncLogger) start() {
	for i := 0; i < l.workerCount; i++ {
		l.wg.Add(1)
		go l.worker(i)
	}
}

// worker processes log entries from the channel
func (l *AsyncLogger) worker(id int) {
	defer l.wg.Done()

	log.Printf("[AsyncLogger] Worker %d started", id)

	for {
		select {
		case <-l.ctx.Done():
			log.Printf("[AsyncLogger] Worker %d shutting down", id)
			return
		case logEntry, ok := <-l.logChannel:
			if !ok {
				log.Printf("[AsyncLogger] Worker %d channel closed", id)
				return
			}
			l.processLog(logEntry)
		}
	}
}

// processLog saves log entry to database with retry logic
func (l *AsyncLogger) processLog(logEntry *models.RequestLog) {
	maxRetries := 3
	var err error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		err = l.db.Create(logEntry).Error
		if err == nil {
			return
		}

		log.Printf("[AsyncLogger] Failed to save log (attempt %d/%d): %v", attempt, maxRetries, err)

		if attempt < maxRetries {
			time.Sleep(time.Duration(attempt) * time.Second)
		}
	}

	log.Printf("[AsyncLogger] Dropped log entry after %d attempts: %v", maxRetries, err)
}

// Log sends a log entry to the async queue
func (l *AsyncLogger) Log(logEntry *models.RequestLog) {
	select {
	case l.logChannel <- logEntry:
		// Successfully queued
	case <-time.After(1 * time.Second):
		log.Printf("[AsyncLogger] Log queue full, dropping log entry")
	}
}

// LogWithTimeout sends a log entry with custom timeout
func (l *AsyncLogger) LogWithTimeout(logEntry *models.RequestLog, timeout time.Duration) bool {
	select {
	case l.logChannel <- logEntry:
		return true
	case <-time.After(timeout):
		log.Printf("[AsyncLogger] Timeout waiting for log queue")
		return false
	}
}

// Shutdown gracefully shuts down the async logger
func (l *AsyncLogger) Shutdown(timeout time.Duration) error {
	log.Printf("[AsyncLogger] Shutting down...")

	// Stop accepting new logs
	close(l.logChannel)

	// Wait for workers to finish with timeout
	done := make(chan struct{})
	go func() {
		l.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Printf("[AsyncLogger] All workers finished")
		return nil
	case <-time.After(timeout):
		l.cancel()
		log.Printf("[AsyncLogger] Shutdown timeout, forcing stop")
		return nil
	}
}

// Stats returns current logger statistics
func (l *AsyncLogger) Stats() map[string]interface{} {
	return map[string]interface{}{
		"queue_length": len(l.logChannel),
		"queue_capacity": cap(l.logChannel),
		"worker_count": l.workerCount,
	}
}
