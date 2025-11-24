package queue

import "time"

// Config represents the queue configuration.
type Config struct {
	// Driver specifies the queue driver (memory, redis, database)
	Driver string

	// Connection specifies the connection name
	Connection string

	// Prefix is the key prefix for the queue
	Prefix string

	// DefaultQueue is the default queue name
	DefaultQueue string

	// MaxAttempts is the default maximum retry attempts
	MaxAttempts int

	// Timeout is the default job timeout
	Timeout time.Duration

	// RetryDelay is the delay between retries
	RetryDelay time.Duration

	// Workers is the default number of workers
	Workers int

	// Options contains driver-specific options
	Options map[string]interface{}
}

// DefaultConfig returns a configuration with sensible defaults.
func DefaultConfig() Config {
	return Config{
		Driver:       "memory",
		Connection:   "default",
		Prefix:       "queue",
		DefaultQueue: "default",
		MaxAttempts:  3,
		Timeout:      30 * time.Second,
		RetryDelay:   time.Second,
		Workers:      5,
		Options:      make(map[string]interface{}),
	}
}

// BatchConfig represents batch processing configuration.
type BatchConfig struct {
	// ChunkSize is the number of items to process per chunk
	ChunkSize int

	// RateLimit is the maximum number of jobs per second
	RateLimit int

	// OnProgress is called when progress is made
	OnProgress func(processed, total int)

	// OnError is called when an item fails
	OnError func(item interface{}, err error)

	// ContinueOnError determines if processing continues after an error
	ContinueOnError bool
}

// DefaultBatchConfig returns a batch configuration with sensible defaults.
func DefaultBatchConfig() BatchConfig {
	return BatchConfig{
		ChunkSize:       100,
		RateLimit:       0, // No limit
		ContinueOnError: true,
	}
}
