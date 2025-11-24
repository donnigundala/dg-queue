package queue

import "errors"

// Common queue errors.
var (
	ErrJobNotFound    = errors.New("job not found")
	ErrQueueNotFound  = errors.New("queue not found")
	ErrWorkerNotFound = errors.New("worker not found")
	ErrJobTimeout     = errors.New("job timeout")
	ErrMaxAttempts    = errors.New("max attempts exceeded")
	ErrInvalidCron    = errors.New("invalid cron expression")
	ErrQueueStopped   = errors.New("queue is stopped")
	ErrInvalidPayload = errors.New("invalid payload")
	ErrDriverNotFound = errors.New("driver not found")
	ErrInvalidConfig  = errors.New("invalid configuration")
	// ErrQueueEmpty is returned when the queue is empty.
	ErrQueueEmpty = errors.New("queue is empty")
)
