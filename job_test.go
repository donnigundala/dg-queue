package dgqueue

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestJob_NewJob(t *testing.T) {
	job := NewJob("test-job", map[string]string{"key": "value"})

	if job.ID == "" {
		t.Error("Expected job ID to be set")
	}
	if _, err := uuid.Parse(job.ID); err != nil {
		t.Errorf("Expected valid UUID, got error: %v", err)
	}
	if job.Name != "test-job" {
		t.Errorf("Expected name 'test-job', got %s", job.Name)
	}
	if job.Queue != "default" {
		t.Errorf("Expected queue 'default', got %s", job.Queue)
	}
	if job.MaxAttempts != 3 {
		t.Errorf("Expected MaxAttempts 3, got %d", job.MaxAttempts)
	}
	if job.Timeout != 30*time.Second {
		t.Errorf("Expected timeout 30s, got %v", job.Timeout)
	}
	if job.Attempts != 0 {
		t.Errorf("Expected 0 attempts, got %d", job.Attempts)
	}
}

func TestJob_WithQueue(t *testing.T) {
	job := NewJob("test", "payload")
	WithQueue(job, "custom")
	if job.Queue != "custom" {
		t.Errorf("Expected queue 'custom', got %s", job.Queue)
	}
}

func TestJob_WithMaxAttempts(t *testing.T) {
	job := NewJob("test", "payload")
	WithMaxAttempts(job, 5)
	if job.MaxAttempts != 5 {
		t.Errorf("Expected MaxAttempts 5, got %d", job.MaxAttempts)
	}
}

func TestJob_WithTimeout(t *testing.T) {
	job := NewJob("test", "payload")
	WithTimeout(job, 60*time.Second)
	if job.Timeout != 60*time.Second {
		t.Errorf("Expected timeout 60s, got %v", job.Timeout)
	}
}

func TestJob_WithDelay(t *testing.T) {
	job := NewJob("test", "payload")
	WithDelay(job, 10*time.Second)
	if job.Delay != 10*time.Second {
		t.Errorf("Expected delay 10s, got %v", job.Delay)
	}
	// AvailableAt should be CreatedAt + Delay
	expected := job.CreatedAt.Add(10 * time.Second)
	if !job.AvailableAt.Equal(expected) {
		t.Errorf("Expected AvailableAt %v, got %v", expected, job.AvailableAt)
	}
}

func TestJob_WithMetadata(t *testing.T) {
	job := NewJob("test", "payload")
	WithMetadata(job, "user_id", 123)
	if job.Metadata["user_id"] != 123 {
		t.Errorf("Expected metadata user_id=123, got %v", job.Metadata["user_id"])
	}
}

func TestJob_IsAvailable(t *testing.T) {
	// Job available now
	job1 := NewJob("test", "payload")
	if !IsAvailable(job1) {
		t.Error("Expected job to be available immediately")
	}

	// Job delayed
	job2 := NewJob("test", "payload")
	WithDelay(job2, 1*time.Hour)
	if IsAvailable(job2) {
		t.Error("Expected delayed job to not be available")
	}
}

func TestJob_CanRetry(t *testing.T) {
	job := NewJob("test", "payload")
	WithMaxAttempts(job, 3)

	// 0 attempts
	if !CanRetry(job) {
		t.Error("Expected job with 0 attempts to be retryable")
	}

	// 2 attempts
	job.Attempts = 2
	if !CanRetry(job) {
		t.Error("Expected job with 2/3 attempts to be retryable")
	}

	// 3 attempts (max reached)
	job.Attempts = 3
	if CanRetry(job) {
		t.Error("Expected job with 3/3 attempts to not be retryable")
	}
}

func TestJob_MarkStarted(t *testing.T) {
	job := NewJob("test", "payload")
	initialUpdatedAt := job.UpdatedAt

	time.Sleep(1 * time.Millisecond) // Ensure time difference
	MarkStarted(job)

	if job.StartedAt == nil {
		t.Error("Expected StartedAt to be set")
	}
	if job.Attempts != 1 {
		t.Errorf("Expected Attempts=1, got %d", job.Attempts)
	}
	if job.UpdatedAt.Equal(initialUpdatedAt) {
		t.Error("Expected UpdatedAt to be updated")
	}
}

func TestJob_MarkCompleted(t *testing.T) {
	job := NewJob("test", "payload")
	initialUpdatedAt := job.UpdatedAt

	time.Sleep(1 * time.Millisecond)
	MarkCompleted(job)

	if job.CompletedAt == nil {
		t.Error("Expected CompletedAt to be set")
	}
	if job.UpdatedAt.Equal(initialUpdatedAt) {
		t.Error("Expected UpdatedAt to be updated")
	}
}

func TestJob_MarkFailed(t *testing.T) {
	job := NewJob("test", "payload")
	initialUpdatedAt := job.UpdatedAt

	time.Sleep(1 * time.Millisecond)
	err := ErrJobTimeout
	MarkFailed(job, err)

	if job.FailedAt == nil {
		t.Error("Expected FailedAt to be set")
	}
	if job.Error != err.Error() {
		t.Errorf("Expected error '%s', got '%s'", err.Error(), job.Error)
	}
	if job.UpdatedAt.Equal(initialUpdatedAt) {
		t.Error("Expected UpdatedAt to be updated")
	}
}

func TestJob_GetStatus(t *testing.T) {
	// Pending
	job := NewJob("test", "payload")
	if GetJobStatus(job) != "pending" {
		t.Errorf("Expected status 'pending', got '%s'", GetJobStatus(job))
	}

	// Delayed
	WithDelay(job, 1*time.Hour)
	if GetJobStatus(job) != "delayed" {
		t.Errorf("Expected status 'delayed', got '%s'", GetJobStatus(job))
	}

	// Processing
	job = NewJob("test", "payload")
	MarkStarted(job)
	if GetJobStatus(job) != "processing" {
		t.Errorf("Expected status 'processing', got '%s'", GetJobStatus(job))
	}

	// Failed
	MarkFailed(job, ErrJobTimeout)
	if GetJobStatus(job) != "failed" {
		t.Errorf("Expected status 'failed', got '%s'", GetJobStatus(job))
	}

	// Completed
	job = NewJob("test", "payload")
	MarkCompleted(job)
	if GetJobStatus(job) != "completed" {
		t.Errorf("Expected status 'completed', got '%s'", GetJobStatus(job))
	}
}

func TestJob_Serialization(t *testing.T) {
	job := NewJob("test-job", map[string]interface{}{
		"user_id": 123,
		"email":   "test@example.com",
	})
	WithQueue(job, "emails")
	WithMaxAttempts(job, 5)

	// Marshal
	data, err := MarshalJob(job)
	if err != nil {
		t.Fatalf("Failed to marshal job: %v", err)
	}

	// Unmarshal
	unmarshaled, err := UnmarshalJob(data)
	if err != nil {
		t.Fatalf("Failed to unmarshal job: %v", err)
	}

	// Verify
	if unmarshaled.ID != job.ID {
		t.Errorf("ID mismatch: expected %s, got %s", job.ID, unmarshaled.ID)
	}
	if unmarshaled.Name != job.Name {
		t.Errorf("Name mismatch: expected %s, got %s", job.Name, unmarshaled.Name)
	}
	if unmarshaled.Queue != job.Queue {
		t.Errorf("Queue mismatch: expected %s, got %s", job.Queue, unmarshaled.Queue)
	}
	if unmarshaled.MaxAttempts != job.MaxAttempts {
		t.Errorf("MaxAttempts mismatch: expected %d, got %d", job.MaxAttempts, unmarshaled.MaxAttempts)
	}
}
