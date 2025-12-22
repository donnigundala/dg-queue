package dgqueue

import (
	"encoding/json"
	"time"

	"github.com/donnigundala/dg-core/contracts/queue"
	"github.com/google/uuid"
)

// NewJob creates a new job.
func NewJob(name string, payload interface{}) *Job {
	now := time.Now()
	return &queue.Job{
		ID:          uuid.New().String(),
		Name:        name,
		Queue:       "default",
		Payload:     payload,
		Attempts:    0,
		MaxAttempts: 3,
		Timeout:     30 * time.Second,
		Delay:       0,
		CreatedAt:   now,
		UpdatedAt:   now,
		AvailableAt: now,
		Metadata:    make(map[string]interface{}),
	}
}

// WithQueue sets the queue name.
func WithQueue(j *Job, queue string) *Job {
	j.Queue = queue
	return j
}

// WithMaxAttempts sets the maximum retry attempts.
func WithMaxAttempts(j *Job, attempts int) *Job {
	j.MaxAttempts = attempts
	return j
}

// WithTimeout sets the job timeout.
func WithTimeout(j *Job, timeout time.Duration) *Job {
	j.Timeout = timeout
	return j
}

// WithDelay sets the job delay.
func WithDelay(j *Job, delay time.Duration) *Job {
	j.Delay = delay
	j.AvailableAt = j.CreatedAt.Add(delay)
	return j
}

// WithMetadata adds metadata to the job.
func WithMetadata(j *Job, key string, value interface{}) *Job {
	j.Metadata[key] = value
	return j
}

// IsAvailable returns true if the job is available for processing.
func IsAvailable(j *Job) bool {
	return time.Now().After(j.AvailableAt) || time.Now().Equal(j.AvailableAt)
}

// CanRetry returns true if the job can be retried.
func CanRetry(j *Job) bool {
	return j.Attempts < j.MaxAttempts
}

// MarkStarted marks the job as started.
func MarkStarted(j *Job) {
	now := time.Now()
	j.StartedAt = &now
	j.UpdatedAt = now
	j.Attempts++
}

// MarkCompleted marks the job as completed.
func MarkCompleted(j *Job) {
	now := time.Now()
	j.CompletedAt = &now
	j.UpdatedAt = now
}

// MarkFailed marks the job as failed.
func MarkFailed(j *Job, err error) {
	now := time.Now()
	j.FailedAt = &now
	j.UpdatedAt = now
	if err != nil {
		j.Error = err.Error()
	}
}

// MarshalJob marshals the job to JSON.
func MarshalJob(j *Job) ([]byte, error) {
	return json.Marshal(j)
}

// UnmarshalJob unmarshals a job from JSON.
func UnmarshalJob(data []byte) (*Job, error) {
	var job queue.Job
	if err := json.Unmarshal(data, &job); err != nil {
		return nil, err
	}
	return &job, nil
}

// GetJobStatus returns the current status of the job.
func GetJobStatus(j *Job) string {
	if j.CompletedAt != nil {
		return "completed"
	}
	if j.FailedAt != nil {
		return "failed"
	}
	if j.StartedAt != nil {
		return "processing"
	}
	if !IsAvailable(j) {
		return "delayed"
	}
	return "pending"
}
