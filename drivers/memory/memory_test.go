package memory

import (
	"testing"

	queue "github.com/donnigundala/dg-queue"
)

func TestMemoryDriver_PushPop(t *testing.T) {
	driver := NewDriver()

	// Create and push a job
	job := queue.NewJob("test-job", map[string]string{"key": "value"})
	err := driver.Push(job)
	if err != nil {
		t.Fatalf("Push failed: %v", err)
	}

	// Pop the job
	popped, err := driver.Pop("default")
	if err != nil {
		t.Fatalf("Pop failed: %v", err)
	}

	if popped.ID != job.ID {
		t.Errorf("Expected job ID %s, got %s", job.ID, popped.ID)
	}
}

func TestMemoryDriver_Delete(t *testing.T) {
	driver := NewDriver()

	job := queue.NewJob("test-job", "payload")
	driver.Push(job)

	err := driver.Delete(job.ID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Try to get deleted job
	_, err = driver.Get(job.ID)
	if err != queue.ErrJobNotFound {
		t.Error("Expected ErrJobNotFound")
	}
}

func TestMemoryDriver_Retry(t *testing.T) {
	driver := NewDriver()

	job := queue.NewJob("test-job", "payload")
	job.MarkFailed(nil)

	err := driver.Retry(job)
	if err != nil {
		t.Fatalf("Retry failed: %v", err)
	}

	// Job should be back in queue
	popped, err := driver.Pop("default")
	if err != nil {
		t.Fatalf("Pop failed: %v", err)
	}

	if popped.ID != job.ID {
		t.Error("Retried job not found in queue")
	}
}

func TestMemoryDriver_Failed(t *testing.T) {
	driver := NewDriver()

	job := queue.NewJob("test-job", "payload")
	job.MarkFailed(nil)

	err := driver.Failed(job)
	if err != nil {
		t.Fatalf("Failed failed: %v", err)
	}

	// Job should be in failed queue
	failed, err := driver.Get(job.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if failed.ID != job.ID {
		t.Error("Failed job not found")
	}
}

func TestMemoryDriver_Size(t *testing.T) {
	driver := NewDriver()

	// Push 3 jobs
	for i := 0; i < 3; i++ {
		job := queue.NewJob("test-job", i)
		driver.Push(job)
	}

	size, err := driver.Size("default")
	if err != nil {
		t.Fatalf("Size failed: %v", err)
	}

	if size != 3 {
		t.Errorf("Expected size 3, got %d", size)
	}
}
