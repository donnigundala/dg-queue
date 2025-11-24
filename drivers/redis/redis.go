package redis

import (
	"context"
	"fmt"
	"time"

	queue "github.com/donnigundala/dg-queue"
	"github.com/redis/go-redis/v9"
)

// Driver is a Redis queue driver.
type Driver struct {
	client *redis.Client
	prefix string
}

// NewDriver creates a new Redis queue driver.
func NewDriver(prefix string, options *redis.Options) (*Driver, error) {
	client := redis.NewClient(options)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &Driver{
		client: client,
		prefix: prefix,
	}, nil
}

// NewDriverWithClient creates a new Redis queue driver with an existing client.
func NewDriverWithClient(client *redis.Client, prefix string) *Driver {
	return &Driver{
		client: client,
		prefix: prefix,
	}
}

// Push pushes a job to the queue.
func (d *Driver) Push(job *queue.Job) error {
	ctx := context.Background()

	data, err := job.Marshal()
	if err != nil {
		return err
	}

	// If job has delay, add to delayed queue (sorted set)
	if job.Delay > 0 || !job.IsAvailable() {
		score := float64(job.AvailableAt.Unix())
		return d.client.ZAdd(ctx, d.delayedKey(job.Queue), redis.Z{
			Score:  score,
			Member: data,
		}).Err()
	}

	// Otherwise, push to regular queue (list)
	return d.client.RPush(ctx, d.queueKey(job.Queue), data).Err()
}

// Pop pops a job from the queue.
func (d *Driver) Pop(queueName string) (*queue.Job, error) {
	ctx := context.Background()

	// First, check delayed queue and move available jobs
	d.moveDelayedJobs(queueName)

	// Pop from regular queue
	data, err := d.client.LPop(ctx, d.queueKey(queueName)).Bytes()
	if err == redis.Nil {
		return nil, queue.ErrQueueEmpty
	}
	if err != nil {
		return nil, err
	}

	return queue.UnmarshalJob(data)
}

// moveDelayedJobs moves delayed jobs that are now available to the regular queue.
func (d *Driver) moveDelayedJobs(queueName string) {
	ctx := context.Background()
	now := float64(time.Now().Unix())

	// Get all jobs with score <= now
	results, err := d.client.ZRangeByScoreWithScores(ctx, d.delayedKey(queueName), &redis.ZRangeBy{
		Min: "-inf",
		Max: fmt.Sprintf("%f", now),
	}).Result()

	if err != nil || len(results) == 0 {
		return
	}

	// Move jobs to regular queue
	pipe := d.client.Pipeline()
	for _, result := range results {
		pipe.RPush(ctx, d.queueKey(queueName), result.Member)
		pipe.ZRem(ctx, d.delayedKey(queueName), result.Member)
	}
	pipe.Exec(ctx)
}

// Delete deletes a job from the queue.
func (d *Driver) Delete(jobID string) error {
	// For simplicity, we don't track individual jobs in Redis
	// Jobs are deleted when popped
	return nil
}

// Retry pushes a job back to the queue for retry.
func (d *Driver) Retry(job *queue.Job) error {
	return d.Push(job)
}

// Failed moves a job to the failed queue.
func (d *Driver) Failed(job *queue.Job) error {
	ctx := context.Background()

	data, err := job.Marshal()
	if err != nil {
		return err
	}

	return d.client.RPush(ctx, d.failedKey(), data).Err()
}

// Get retrieves a job by ID (not supported in Redis driver).
func (d *Driver) Get(jobID string) (*queue.Job, error) {
	return nil, fmt.Errorf("Get not supported in Redis driver")
}

// Size returns the number of jobs in the queue.
func (d *Driver) Size(queueName string) (int, error) {
	ctx := context.Background()

	regularSize, err := d.client.LLen(ctx, d.queueKey(queueName)).Result()
	if err != nil {
		return 0, err
	}

	delayedSize, err := d.client.ZCard(ctx, d.delayedKey(queueName)).Result()
	if err != nil {
		return 0, err
	}

	return int(regularSize + delayedSize), nil
}

// Close closes the Redis connection.
func (d *Driver) Close() error {
	return d.client.Close()
}

// Helper methods for key generation
func (d *Driver) queueKey(name string) string {
	return fmt.Sprintf("%s:queues:%s", d.prefix, name)
}

func (d *Driver) delayedKey(name string) string {
	return fmt.Sprintf("%s:queues:%s:delayed", d.prefix, name)
}

func (d *Driver) failedKey() string {
	return fmt.Sprintf("%s:failed", d.prefix)
}
