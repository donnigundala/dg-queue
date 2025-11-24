package queue

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Job represents a queued job.
type Job struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Queue       string                 `json:"queue"`
	Payload     interface{}            `json:"payload"`
	Attempts    int                    `json:"attempts"`
	MaxAttempts int                    `json:"max_attempts"`
	Timeout     time.Duration          `json:"timeout"`
	Delay       time.Duration          `json:"delay"`
	AvailableAt time.Time              `json:"available_at"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	StartedAt   *time.Time             `json:"started_at,omitempty"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	FailedAt    *time.Time             `json:"failed_at,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// NewJob creates a new job.
func NewJob(name string, payload interface{}) *Job {
	now := time.Now()
	return &Job{
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
func (j *Job) WithQueue(queue string) *Job {
	j.Queue = queue
	return j
}

// WithMaxAttempts sets the maximum retry attempts.
func (j *Job) WithMaxAttempts(attempts int) *Job {
	j.MaxAttempts = attempts
	return j
}

// WithTimeout sets the job timeout.
func (j *Job) WithTimeout(timeout time.Duration) *Job {
	j.Timeout = timeout
	return j
}

// WithDelay sets the job delay.
func (j *Job) WithDelay(delay time.Duration) *Job {
	j.Delay = delay
	j.AvailableAt = j.CreatedAt.Add(delay)
	return j
}

// WithMetadata adds metadata to the job.
func (j *Job) WithMetadata(key string, value interface{}) *Job {
	j.Metadata[key] = value
	return j
}

// IsAvailable returns true if the job is available for processing.
func (j *Job) IsAvailable() bool {
	return time.Now().After(j.AvailableAt) || time.Now().Equal(j.AvailableAt)
}

// CanRetry returns true if the job can be retried.
func (j *Job) CanRetry() bool {
	return j.Attempts < j.MaxAttempts
}

// MarkStarted marks the job as started.
func (j *Job) MarkStarted() {
	now := time.Now()
	j.StartedAt = &now
	j.UpdatedAt = now
	j.Attempts++
}

// MarkCompleted marks the job as completed.
func (j *Job) MarkCompleted() {
	now := time.Now()
	j.CompletedAt = &now
	j.UpdatedAt = now
}

// MarkFailed marks the job as failed.
func (j *Job) MarkFailed(err error) {
	now := time.Now()
	j.FailedAt = &now
	j.UpdatedAt = now
	if err != nil {
		j.Error = err.Error()
	}
}

// Marshal marshals the job to JSON.
func (j *Job) Marshal() ([]byte, error) {
	return json.Marshal(j)
}

// Unmarshal unmarshals a job from JSON.
func UnmarshalJob(data []byte) (*Job, error) {
	var job Job
	if err := json.Unmarshal(data, &job); err != nil {
		return nil, err
	}
	return &job, nil
}

// JobStatus represents the status of a job.
type JobStatus struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Queue     string    `json:"queue"`
	Status    string    `json:"status"` // pending, processing, completed, failed
	Attempts  int       `json:"attempts"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Error     string    `json:"error,omitempty"`
}

// GetStatus returns the current status of the job.
func (j *Job) GetStatus() string {
	if j.CompletedAt != nil {
		return "completed"
	}
	if j.FailedAt != nil {
		return "failed"
	}
	if j.StartedAt != nil {
		return "processing"
	}
	if !j.IsAvailable() {
		return "delayed"
	}
	return "pending"
}
