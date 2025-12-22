package memory

import (
	"context"
	"sync"

	"github.com/donnigundala/dg-core/contracts/queue"
	dgqueue "github.com/donnigundala/dg-queue"
)

// Driver is an in-memory queue driver for testing.
type Driver struct {
	queues map[string][]*queue.Job
	failed map[string]*queue.Job
	mu     sync.RWMutex
}

// NewDriver creates a new memory driver.
func NewDriver() *Driver {
	return &Driver{
		queues: make(map[string][]*queue.Job),
		failed: make(map[string]*queue.Job),
	}
}

// Push pushes a job to the queue.
func (d *Driver) Push(ctx context.Context, job *queue.Job) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.queues[job.Queue] == nil {
		d.queues[job.Queue] = make([]*queue.Job, 0)
	}

	d.queues[job.Queue] = append(d.queues[job.Queue], job)
	return nil
}

// Pop pops a job from the queue.
func (d *Driver) Pop(ctx context.Context, queueName string) (*queue.Job, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	jobs, exists := d.queues[queueName]
	if !exists || len(jobs) == 0 {
		return nil, dgqueue.ErrQueueEmpty
	}

	// Find first available job
	for i, job := range jobs {
		if dgqueue.IsAvailable(job) {
			// Remove from queue
			d.queues[queueName] = append(jobs[:i], jobs[i+1:]...)
			return job, nil
		}
	}

	// No available jobs (all delayed)
	return nil, dgqueue.ErrQueueEmpty
}

// Delete deletes a job.
func (d *Driver) Delete(ctx context.Context, jobID string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Search all queues
	for queueName, jobs := range d.queues {
		for i, job := range jobs {
			if job.ID == jobID {
				d.queues[queueName] = append(jobs[:i], jobs[i+1:]...)
				return nil
			}
		}
	}

	// Check failed jobs
	if _, exists := d.failed[jobID]; exists {
		delete(d.failed, jobID)
		return nil
	}

	return dgqueue.ErrJobNotFound
}

// Retry retries a failed job.
func (d *Driver) Retry(ctx context.Context, job *queue.Job) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Reset job state
	job.FailedAt = nil
	job.Error = ""

	// Push back to queue
	if d.queues[job.Queue] == nil {
		d.queues[job.Queue] = make([]*queue.Job, 0)
	}

	d.queues[job.Queue] = append(d.queues[job.Queue], job)
	return nil
}

// Failed moves a job to the dead letter queue.
func (d *Driver) Failed(ctx context.Context, job *queue.Job) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.failed[job.ID] = job
	return nil
}

// Get gets a job by ID.
func (d *Driver) Get(ctx context.Context, jobID string) (*queue.Job, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	// Search all queues
	for _, jobs := range d.queues {
		for _, job := range jobs {
			if job.ID == jobID {
				return job, nil
			}
		}
	}

	// Check failed jobs
	if job, exists := d.failed[jobID]; exists {
		return job, nil
	}

	return nil, dgqueue.ErrJobNotFound
}

// Size returns the number of jobs in a queue.
func (d *Driver) Size(ctx context.Context, queueName string) (int64, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if jobs, exists := d.queues[queueName]; exists {
		return int64(len(jobs)), nil
	}

	return 0, nil
}

// Close closes the driver.
func (d *Driver) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.queues = make(map[string][]*queue.Job)
	d.failed = make(map[string]*queue.Job)
	return nil
}
