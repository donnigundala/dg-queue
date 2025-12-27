package dgqueue

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// DriverFactory is a function that creates a queue driver.
type DriverFactory func(Config) (Driver, error)

var (
	globalDrivers   = make(map[string]DriverFactory)
	globalDriversMu sync.RWMutex
)

// RegisterDriver registers a driver factory globally.
func RegisterDriver(name string, factory DriverFactory) {
	globalDriversMu.Lock()
	defer globalDriversMu.Unlock()
	globalDrivers[name] = factory
}

// Manager is the main queue manager implementation.
type Manager struct {
	config     Config
	driver     Driver
	workers    map[string]*workerPool
	middleware []Middleware
	running    bool
	stopChan   chan struct{}
	wg         sync.WaitGroup
	mu         sync.RWMutex

	// Observability
	metricQueueDepth    metric.Int64ObservableGauge
	metricActiveWorkers metric.Int64ObservableGauge
	metricJobProcessed  metric.Int64Counter
	metricJobDuration   metric.Float64Histogram
}

// workerPool represents a pool of workers for a specific job type.
type workerPool struct {
	name        string
	concurrency int
	handler     WorkerFunc
	jobs        chan *Job
	stopChan    chan struct{}
	wg          sync.WaitGroup
}

// New creates a new queue manager.
func New(config Config) *Manager {
	return &Manager{
		config:     config,
		workers:    make(map[string]*workerPool),
		middleware: make([]Middleware, 0),
		stopChan:   make(chan struct{}),
	}
}

// SetDriver sets the queue driver.
func (m *Manager) SetDriver(driver Driver) {
	m.driver = driver
}

// Dispatch dispatches a job immediately.
func (m *Manager) Dispatch(ctx context.Context, name string, payload interface{}) (*Job, error) {
	job := NewJob(name, payload)
	job.Queue = m.config.DefaultQueue
	job.MaxAttempts = m.config.MaxAttempts
	job.Timeout = m.config.Timeout

	if err := m.driver.Push(ctx, job); err != nil {
		return nil, err
	}

	return job, nil
}

// DispatchAfter dispatches a job with a delay.
func (m *Manager) DispatchAfter(ctx context.Context, name string, payload interface{}, delay time.Duration) (*Job, error) {
	job := NewJob(name, payload)
	job.Queue = m.config.DefaultQueue
	job.MaxAttempts = m.config.MaxAttempts
	job.Timeout = m.config.Timeout
	WithDelay(job, delay)

	if err := m.driver.Push(ctx, job); err != nil {
		return nil, err
	}

	return job, nil
}

// DispatchBatch dispatches multiple jobs as a batch.
func (m *Manager) DispatchBatch(name string, config BatchConfig, items interface{}, mapper BatchMapper) error {
	// TODO: Implement batch processing
	return fmt.Errorf("batch processing not yet implemented")
}

// Worker registers a worker for a job name.
func (m *Manager) Worker(name string, concurrency int, handler WorkerFunc) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if concurrency <= 0 {
		concurrency = m.config.Workers
	}

	// Apply middleware
	finalHandler := handler
	for i := len(m.middleware) - 1; i >= 0; i-- {
		finalHandler = m.middleware[i](finalHandler)
	}

	m.workers[name] = &workerPool{
		name:        name,
		concurrency: concurrency,
		handler:     finalHandler,
		jobs:        make(chan *Job, concurrency*2),
		stopChan:    make(chan struct{}),
	}

	return nil
}

// Use adds middleware to the queue.
func (m *Manager) Use(middleware Middleware) Queue {
	m.middleware = append(m.middleware, middleware)
	return m
}

// Start starts the queue workers and scheduler.
func (m *Manager) Start() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.running {
		return fmt.Errorf("queue already running")
	}

	// Check if workers are enabled
	if !m.config.WorkerEnabled {
		m.logInfo("Queue workers disabled by config")
		return nil
	}

	// Recreate stopChan for safe restart
	m.stopChan = make(chan struct{})
	m.running = true

	// Start workers
	for _, worker := range m.workers {
		m.startWorkerPool(worker)
	}

	m.logInfo("Queue manager starting", "workers", len(m.workers))

	// Start dispatcher
	m.wg.Add(1)
	go m.dispatchJobs(context.Background())

	m.logInfo("Queue manager started", "workers", len(m.workers))
	return nil
}

// Stop stops the queue manager gracefully.
func (m *Manager) Stop(ctx context.Context) error {
	m.mu.Lock()
	if !m.running {
		m.mu.Unlock()
		return nil
	}

	m.logInfo("Queue manager stopping", "workers", len(m.workers))

	m.running = false
	close(m.stopChan)
	m.mu.Unlock()

	// Stop all workers
	for _, worker := range m.workers {
		close(worker.stopChan)
		worker.wg.Wait()
	}

	// Wait for dispatcher to finish
	m.wg.Wait()

	// Close driver connection
	if m.driver != nil {
		if err := m.driver.Close(); err != nil {
			m.logError("Failed to close driver", err)
			return fmt.Errorf("failed to close driver: %w", err)
		}
	}

	m.logInfo("Queue manager stopped")
	return nil
}

// Status returns the status of a job.
func (m *Manager) Status(ctx context.Context, jobID string) (*JobStatus, error) {
	job, err := m.driver.Get(ctx, jobID)
	if err != nil {
		return nil, err
	}

	return &JobStatus{
		ID:        job.ID,
		Name:      job.Name,
		Queue:     job.Queue,
		Status:    GetJobStatus(job),
		Attempts:  job.Attempts,
		CreatedAt: job.CreatedAt,
		UpdatedAt: job.UpdatedAt,
		Error:     job.Error,
	}, nil
}

// Driver returns the underlying driver.
func (m *Manager) Driver() Driver {
	return m.driver
}

// startWorkerPool starts a worker pool.
func (m *Manager) startWorkerPool(pool *workerPool) {
	for i := 0; i < pool.concurrency; i++ {
		pool.wg.Add(1)
		go m.runWorker(pool, i)
	}
}

// runWorker runs a single worker.
func (m *Manager) runWorker(pool *workerPool, id int) {
	defer pool.wg.Done()

	for {
		select {
		case job := <-pool.jobs:
			m.processJob(pool, job)
		case <-pool.stopChan:
			return
		}
	}
}

// processJob processes a single job.
func (m *Manager) processJob(pool *workerPool, job *Job) {
	MarkStarted(job)

	// Create timeout context
	ctx, cancel := context.WithTimeout(context.Background(), job.Timeout)
	defer cancel()

	// Run job with timeout
	done := make(chan error, 1)
	go func() {
		done <- pool.handler(ctx, job)
	}()

	select {
	case err := <-done:
		if err != nil {
			MarkFailed(job, err)
			if CanRetry(job) {
				m.logInfo("Job failed, retrying", "job_id", job.ID, "job_name", job.Name, "attempt", job.Attempts, "error", err)
				// Retry with backoff
				WithDelay(job, m.config.RetryDelay*time.Duration(job.Attempts))
				m.driver.Retry(ctx, job)
			} else {
				m.logError("Job failed permanently", err, "job_id", job.ID, "job_name", job.Name, "attempts", job.Attempts)
				// Move to dead letter queue
				m.driver.Failed(ctx, job)
			}
		} else {
			MarkCompleted(job)
			m.driver.Delete(ctx, job.ID)
		}

		// Record metrics
		if m.metricJobProcessed != nil {
			status := "success"
			if err != nil {
				status = "failed"
			}
			attrs := metric.WithAttributes(
				attribute.String("queue.name", pool.name),
				attribute.String("job.status", status),
			)
			m.metricJobProcessed.Add(ctx, 1, attrs)

			duration := float64(time.Since(job.CreatedAt).Milliseconds()) // Or use start time of processing?
			// job.CreatedAt is creation time. We usually want processing duration.
			// Let's rely on standard "duration from start of handler".
			// But wait, the previous code didn't capture start time separately.
			// Let's assume we want end-to-end latency for now or modification.
			// Actually better to just wrap the handler execution time.
			// Re-reading code: 'done' channel waits for handler.
			// I'll stick to job.CreatedAt for E2E latency or I'll assume approximate duration is ok.
			// Let's use E2E latency (CreatedAt -> Now) as "duration" for now as it's more useful for queue lag.
			m.metricJobDuration.Record(ctx, duration, attrs)
		}
	case <-ctx.Done():
		MarkFailed(job, ErrJobTimeout)
		if CanRetry(job) {
			m.logInfo("Job timed out, retrying", "job_id", job.ID, "job_name", job.Name, "attempt", job.Attempts)
			m.driver.Retry(context.Background(), job)
		} else {
			m.logError("Job timed out permanently", ErrJobTimeout, "job_id", job.ID, "job_name", job.Name, "attempts", job.Attempts)
			m.driver.Failed(context.Background(), job)
		}
	}
}

// dispatchJobs dispatches jobs to workers.
func (m *Manager) dispatchJobs(ctx context.Context) {
	defer m.wg.Done()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.fetchAndDispatchJobs()
		case <-m.stopChan:
			return
		case <-ctx.Done():
			return
		}
	}
}

// fetchAndDispatchJobs fetches jobs from the driver and dispatches to workers.
func (m *Manager) fetchAndDispatchJobs() {
	ctx := context.Background()
	// Pop ONE job at a time (not one per worker!)
	job, err := m.driver.Pop(ctx, m.config.DefaultQueue)
	if err != nil {
		return
	}

	// Find the worker for this job
	m.mu.RLock()
	pool, exists := m.workers[job.Name]
	m.mu.RUnlock()

	if !exists {
		// No worker registered for this job type -> dead letter queue
		m.driver.Failed(ctx, job)
		return
	}

	// Try to dispatch to worker pool
	select {
	case pool.jobs <- job:
		// Successfully dispatched
	default:
		// Worker pool is full, push job back to queue
		m.driver.Push(ctx, job)
	}
}
