package queue

import (
	"context"
	"time"
)

// Queue is the main interface for the queue system.
type Queue interface {
	// Dispatch dispatches a job immediately
	Dispatch(name string, payload interface{}) (*Job, error)

	// DispatchAfter dispatches a job with a delay
	DispatchAfter(name string, payload interface{}, delay time.Duration) (*Job, error)

	// DispatchBatch dispatches multiple jobs as a batch
	DispatchBatch(name string, config BatchConfig, items interface{}, mapper BatchMapper) error

	// Schedule schedules a job using cron syntax
	Schedule(cron string, name string, handler ScheduleHandler) error

	// Worker registers a worker for a job name
	Worker(name string, concurrency int, handler WorkerFunc) error

	// Use adds middleware to the queue
	Use(middleware Middleware) Queue

	// Start starts the queue workers and scheduler
	Start(ctx context.Context) error

	// Stop stops the queue gracefully
	Stop(ctx context.Context) error

	// Status returns the status of a job
	Status(jobID string) (*JobStatus, error)

	// Driver returns the underlying driver
	Driver() Driver
}

// WorkerFunc is the function signature for job handlers.
type WorkerFunc func(job *Job) error

// ScheduleHandler is the function signature for scheduled job handlers.
type ScheduleHandler func() error

// BatchMapper is the function signature for batch item mapping.
type BatchMapper func(item interface{}) (interface{}, error)

// Middleware is the function signature for queue middleware.
type Middleware func(next WorkerFunc) WorkerFunc

// Driver is the interface for queue storage drivers.
type Driver interface {
	// Push pushes a job to the queue
	Push(job *Job) error

	// Pop pops a job from the queue
	Pop(queueName string) (*Job, error)

	// Delete deletes a job
	Delete(jobID string) error

	// Retry retries a failed job
	Retry(job *Job) error

	// Failed moves a job to the dead letter queue
	Failed(job *Job) error

	// Get gets a job by ID
	Get(jobID string) (*Job, error)

	// Size returns the number of jobs in a queue
	Size(queueName string) (int64, error)

	// Close closes the driver
	Close() error
}
