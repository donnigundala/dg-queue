package dgqueue

import (
	"time"

	"github.com/donnigundala/dg-core/contracts/queue"
)

// Alias types for convenience within the package
type Queue = queue.Queue
type Driver = queue.Driver
type Job = queue.Job
type JobStatus = queue.JobStatus
type WorkerFunc = queue.WorkerFunc
type Middleware = queue.Middleware

// BatchMapper is the function signature for batch item mapping.
type BatchMapper func(item interface{}) (interface{}, error)

// BatchConfig represents the configuration for a batch of jobs.
type BatchConfig struct {
	ChunkSize       int
	OnProgress      func(processed, total int)
	OnError         func(item interface{}, err error)
	ContinueOnError bool
	RateLimit       time.Duration
}
