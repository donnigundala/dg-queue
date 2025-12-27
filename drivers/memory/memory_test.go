package memory

import (
	"context"
	"testing"

	dgqueue "github.com/donnigundala/dg-queue"
)

func TestMemoryDriver_PushPop(t *testing.T) {
	driver, _ := NewDriver(dgqueue.DefaultConfig())
	ctx := context.Background()

	// Create and push a job
	job := dgqueue.NewJob("test-job", map[string]string{"key": "value"})
	err := driver.Push(ctx, job)
	if err != nil {
		t.Fatalf("Push failed: %v", err)
	}

	// Pop the job
	popped, err := driver.Pop(ctx, "default")
	if err != nil {
		t.Fatalf("Pop failed: %v", err)
	}

	if popped.ID != job.ID {
		t.Errorf("Expected job ID %s, got %s", job.ID, popped.ID)
	}
}

func TestMemoryDriver_Delete(t *testing.T) {
	driver, _ := NewDriver(dgqueue.DefaultConfig())
	ctx := context.Background()

	job := dgqueue.NewJob("test-job", "payload")
	driver.Push(ctx, job)

	err := driver.Delete(ctx, job.ID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Try to get deleted job
	_, err = driver.Get(ctx, job.ID)
	if err != dgqueue.ErrJobNotFound {
		t.Error("Expected ErrJobNotFound")
	}
}

func TestMemoryDriver_Retry(t *testing.T) {
	driver, _ := NewDriver(dgqueue.DefaultConfig())
	ctx := context.Background()

	job := dgqueue.NewJob("test-job", "payload")
	dgqueue.MarkFailed(job, nil)

	err := driver.Retry(ctx, job)
	if err != nil {
		t.Fatalf("Retry failed: %v", err)
	}

	// Job should be back in queue
	popped, err := driver.Pop(ctx, "default")
	if err != nil {
		t.Fatalf("Pop failed: %v", err)
	}

	if popped.ID != job.ID {
		t.Error("Retried job not found in queue")
	}
}

func TestMemoryDriver_Failed(t *testing.T) {
	driver, _ := NewDriver(dgqueue.DefaultConfig())
	ctx := context.Background()

	job := dgqueue.NewJob("test-job", "payload")
	dgqueue.MarkFailed(job, nil)

	err := driver.Failed(ctx, job)
	if err != nil {
		t.Fatalf("Failed failed: %v", err)
	}

	// Job should be in failed queue
	failed, err := driver.Get(ctx, job.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if failed.ID != job.ID {
		t.Error("Failed job not found")
	}
}

func TestMemoryDriver_Size(t *testing.T) {
	driver, _ := NewDriver(dgqueue.DefaultConfig())
	ctx := context.Background()

	// Push 3 jobs
	for i := 0; i < 3; i++ {
		job := dgqueue.NewJob("test-job", i)
		driver.Push(ctx, job)
	}

	size, err := driver.Size(ctx, "default")
	if err != nil {
		t.Fatalf("Size failed: %v", err)
	}

	if size != 3 {
		t.Errorf("Expected size 3, got %d", size)
	}
}
