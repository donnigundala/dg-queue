package queue

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Manager is the main queue manager implementation.
type Manager struct {
	config     Config
	driver     Driver
	workers    map[string]*workerPool
	middleware []Middleware
	schedules  []*schedule
	running    bool
	stopChan   chan struct{}
	wg         sync.WaitGroup
	mu         sync.RWMutex
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

// schedule represents a scheduled job.
type schedule struct {
	cron    string
	name    string
	handler ScheduleHandler
	next    time.Time
}

// New creates a new queue manager.
func New(config Config) *Manager {
	return &Manager{
		config:     config,
		workers:    make(map[string]*workerPool),
		middleware: make([]Middleware, 0),
		schedules:  make([]*schedule, 0),
		stopChan:   make(chan struct{}),
	}
}

// SetDriver sets the queue driver.
func (m *Manager) SetDriver(driver Driver) {
	m.driver = driver
}

// Dispatch dispatches a job immediately.
func (m *Manager) Dispatch(name string, payload interface{}) (*Job, error) {
	job := NewJob(name, payload)
	job.Queue = m.config.DefaultQueue
	job.MaxAttempts = m.config.MaxAttempts
	job.Timeout = m.config.Timeout

	if err := m.driver.Push(job); err != nil {
		return nil, err
	}

	return job, nil
}

// DispatchAfter dispatches a job with a delay.
func (m *Manager) DispatchAfter(name string, payload interface{}, delay time.Duration) (*Job, error) {
	job := NewJob(name, payload)
	job.Queue = m.config.DefaultQueue
	job.MaxAttempts = m.config.MaxAttempts
	job.Timeout = m.config.Timeout
	job.WithDelay(delay)

	if err := m.driver.Push(job); err != nil {
		return nil, err
	}

	return job, nil
}

// DispatchBatch dispatches multiple jobs as a batch.
func (m *Manager) DispatchBatch(name string, config BatchConfig, items interface{}, mapper BatchMapper) error {
	// TODO: Implement batch processing
	return fmt.Errorf("batch processing not yet implemented")
}

// Schedule schedules a job using cron syntax.
func (m *Manager) Schedule(cron string, name string, handler ScheduleHandler) error {
	// TODO: Implement cron parsing
	m.mu.Lock()
	defer m.mu.Unlock()

	m.schedules = append(m.schedules, &schedule{
		cron:    cron,
		name:    name,
		handler: handler,
		next:    time.Now(), // TODO: Calculate next run time
	})

	return nil
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
func (m *Manager) Start(ctx context.Context) error {
	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		return fmt.Errorf("queue already running")
	}
	m.running = true
	m.mu.Unlock()

	// Start workers
	for _, worker := range m.workers {
		m.startWorkerPool(worker)
	}

	// Start job dispatcher
	m.wg.Add(1)
	go m.dispatchJobs(ctx)

	// Start scheduler (if schedules exist)
	if len(m.schedules) > 0 {
		m.wg.Add(1)
		go m.runScheduler(ctx)
	}

	return nil
}

// Stop stops the queue gracefully.
func (m *Manager) Stop(ctx context.Context) error {
	m.mu.Lock()
	if !m.running {
		m.mu.Unlock()
		return nil
	}
	m.running = false
	m.mu.Unlock()

	// Signal stop
	close(m.stopChan)

	// Stop all workers
	for _, worker := range m.workers {
		close(worker.stopChan)
		worker.wg.Wait()
	}

	// Wait for dispatcher and scheduler
	m.wg.Wait()

	return nil
}

// Status returns the status of a job.
func (m *Manager) Status(jobID string) (*JobStatus, error) {
	job, err := m.driver.Get(jobID)
	if err != nil {
		return nil, err
	}

	return &JobStatus{
		ID:        job.ID,
		Name:      job.Name,
		Queue:     job.Queue,
		Status:    job.GetStatus(),
		Attempts:  job.Attempts,
		CreatedAt: job.CreatedAt,
		UpdatedAt: job.CreatedAt, // TODO: Track updates
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
	job.MarkStarted()

	// Create timeout context
	ctx, cancel := context.WithTimeout(context.Background(), job.Timeout)
	defer cancel()

	// Run job with timeout
	done := make(chan error, 1)
	go func() {
		done <- pool.handler(job)
	}()

	select {
	case err := <-done:
		if err != nil {
			job.MarkFailed(err)
			if job.CanRetry() {
				// Retry with backoff
				job.WithDelay(m.config.RetryDelay * time.Duration(job.Attempts))
				m.driver.Retry(job)
			} else {
				// Move to dead letter queue
				m.driver.Failed(job)
			}
		} else {
			job.MarkCompleted()
			m.driver.Delete(job.ID)
		}
	case <-ctx.Done():
		job.MarkFailed(ErrJobTimeout)
		if job.CanRetry() {
			m.driver.Retry(job)
		} else {
			m.driver.Failed(job)
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
	for name, pool := range m.workers {
		// Try to pop a job
		job, err := m.driver.Pop(m.config.DefaultQueue)
		if err != nil {
			continue
		}

		// Check if job matches this worker
		if job.Name == name {
			select {
			case pool.jobs <- job:
				// Job dispatched
			default:
				// Worker busy, push back
				m.driver.Push(job)
			}
		} else {
			// Wrong worker, push back
			m.driver.Push(job)
		}
	}
}

// runScheduler runs the job scheduler.
func (m *Manager) runScheduler(ctx context.Context) {
	defer m.wg.Done()

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.checkSchedules()
		case <-m.stopChan:
			return
		case <-ctx.Done():
			return
		}
	}
}

// checkSchedules checks and executes due schedules.
func (m *Manager) checkSchedules() {
	now := time.Now()

	for _, sched := range m.schedules {
		if now.After(sched.next) || now.Equal(sched.next) {
			// Execute scheduled job
			go func(s *schedule) {
				if err := s.handler(); err != nil {
					// Log error (TODO: Add proper logging)
					fmt.Printf("Scheduled job %s failed: %v\n", s.name, err)
				}
			}(sched)

			// TODO: Calculate next run time based on cron
			sched.next = now.Add(1 * time.Hour) // Placeholder
		}
	}
}
