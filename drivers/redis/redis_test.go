package redis

import (
	"context"
	"testing"
	"time"

	queue "github.com/donnigundala/dg-queue"
	"github.com/redis/go-redis/v9"
)

func setupRedisDriver(t *testing.T) *Driver {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		t.Skip("Redis not available, skipping test")
	}

	// Clear test data
	prefix := "test_queue"
	ctx = context.Background()
	keys, _ := client.Keys(ctx, prefix+":*").Result()
	if len(keys) > 0 {
		client.Del(ctx, keys...)
	}

	return NewDriverWithClient(client, prefix)
}

func TestRedisDriver_PushPop(t *testing.T) {
	driver := setupRedisDriver(t)
	defer driver.Close()

	job := queue.NewJob("test-job", map[string]string{"key": "value"})

	// Push
	err := driver.Push(job)
	if err != nil {
		t.Fatalf("Failed to push job: %v", err)
	}

	// Pop
	popped, err := driver.Pop("default")
	if err != nil {
		t.Fatalf("Failed to pop job: %v", err)
	}

	if popped.ID != job.ID {
		t.Errorf("Expected ID %s, got %s", job.ID, popped.ID)
	}
	if popped.Name != job.Name {
		t.Errorf("Expected name %s, got %s", job.Name, popped.Name)
	}
}

func TestRedisDriver_DelayedJob(t *testing.T) {
	driver := setupRedisDriver(t)
	defer driver.Close()

	// Create delayed job
	job := queue.NewJob("delayed-job", "payload").WithDelay(2 * time.Second)

	// Push
	err := driver.Push(job)
	if err != nil {
		t.Fatalf("Failed to push delayed job: %v", err)
	}

	// Try to pop immediately (should be empty)
	_, err = driver.Pop("default")
	if err != queue.ErrQueueEmpty {
		t.Error("Expected queue to be empty for delayed job")
	}

	// Wait for job to become available
	time.Sleep(2500 * time.Millisecond)

	// Pop should now work
	popped, err := driver.Pop("default")
	if err != nil {
		t.Fatalf("Failed to pop after delay: %v", err)
	}

	if popped.ID != job.ID {
		t.Errorf("Expected ID %s, got %s", job.ID, popped.ID)
	}
}

func TestRedisDriver_Failed(t *testing.T) {
	driver := setupRedisDriver(t)
	defer driver.Close()

	job := queue.NewJob("failed-job", "payload")

	// Move to failed queue
	err := driver.Failed(job)
	if err != nil {
		t.Fatalf("Failed to move job to failed queue: %v", err)
	}

	// Verify job is in failed queue
	ctx := context.Background()
	count, err := driver.client.LLen(ctx, driver.failedKey()).Result()
	if err != nil {
		t.Fatalf("Failed to get failed queue size: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected 1 job in failed queue, got %d", count)
	}
}

func TestRedisDriver_Size(t *testing.T) {
	driver := setupRedisDriver(t)
	defer driver.Close()

	// Initially empty
	size, err := driver.Size("default")
	if err != nil {
		t.Fatalf("Failed to get size: %v", err)
	}
	if size != 0 {
		t.Errorf("Expected size 0, got %d", size)
	}

	// Add regular job
	job1 := queue.NewJob("job1", "payload")
	driver.Push(job1)

	// Add delayed job
	job2 := queue.NewJob("job2", "payload").WithDelay(1 * time.Hour)
	driver.Push(job2)

	// Size should include both
	size, err = driver.Size("default")
	if err != nil {
		t.Fatalf("Failed to get size: %v", err)
	}
	if size != 2 {
		t.Errorf("Expected size 2, got %d", size)
	}
}

func TestRedisDriver_Retry(t *testing.T) {
	driver := setupRedisDriver(t)
	defer driver.Close()

	job := queue.NewJob("retry-job", "payload")
	job.Attempts = 1

	// Retry
	err := driver.Retry(job)
	if err != nil {
		t.Fatalf("Failed to retry job: %v", err)
	}

	// Pop and verify
	popped, err := driver.Pop("default")
	if err != nil {
		t.Fatalf("Failed to pop retried job: %v", err)
	}

	if popped.ID != job.ID {
		t.Errorf("Expected ID %s, got %s", job.ID, popped.ID)
	}
	if popped.Attempts != 1 {
		t.Errorf("Expected attempts 1, got %d", popped.Attempts)
	}
}

func TestRedisDriver_NewDriverWithClient(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		t.Skip("Redis not available, skipping test")
	}

	// Create driver with shared client
	driver := NewDriverWithClient(client, "shared_test")
	defer driver.Close()

	job := queue.NewJob("shared-job", "payload")
	err := driver.Push(job)
	if err != nil {
		t.Fatalf("Failed to push with shared client: %v", err)
	}

	popped, err := driver.Pop("default")
	if err != nil {
		t.Fatalf("Failed to pop with shared client: %v", err)
	}

	if popped.ID != job.ID {
		t.Errorf("Expected ID %s, got %s", job.ID, popped.ID)
	}
}
