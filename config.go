package queue

import "time"

// Logger is the interface for structured logging.
// Implement this interface to integrate with your logging system.
type Logger interface {
	// Info logs an informational message
	Info(msg string, keysAndValues ...interface{})

	// Error logs an error message
	Error(msg string, err error, keysAndValues ...interface{})

	// Warn logs a warning message
	Warn(msg string, keysAndValues ...interface{})
}

// Config represents the queue configuration.
type Config struct {
	// Driver specifies the queue driver (memory, redis, database)
	Driver string `mapstructure:"driver"`

	// Connection specifies the connection name
	Connection string `mapstructure:"connection"`

	// Prefix is the key prefix for the queue
	Prefix string `mapstructure:"prefix"`

	// DefaultQueue is the default queue name
	DefaultQueue string `mapstructure:"default_queue"`

	// MaxAttempts is the default maximum retry attempts
	MaxAttempts int `mapstructure:"max_attempts"`

	// Timeout is the default job timeout
	Timeout time.Duration `mapstructure:"timeout"`

	// RetryDelay is the delay between retries
	RetryDelay time.Duration `mapstructure:"retry_delay"`

	// Workers is the default number of workers
	Workers int `mapstructure:"workers"`

	// Options contains driver-specific options
	Options map[string]interface{} `mapstructure:"options"`

	// Logger is used for structured logging (optional)
	// If nil, no logging will be performed
	Logger Logger
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
		Logger:       nil, // No logging by default
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
