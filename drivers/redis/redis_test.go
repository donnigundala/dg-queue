package redis

import (
	"context"
	"testing"
	"time"

	dgqueue "github.com/donnigundala/dg-queue"
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
	ctx := context.Background()

	job := dgqueue.NewJob("test-job", map[string]string{"key": "value"})

	// Push
	err := driver.Push(ctx, job)
	if err != nil {
		t.Fatalf("Failed to push job: %v", err)
	}

	// Pop
	popped, err := driver.Pop(ctx, "default")
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
	ctx := context.Background()

	// Create delayed job
	job := dgqueue.NewJob("delayed-job", "payload")
	dgqueue.WithDelay(job, 2*time.Second)

	// Push
	err := driver.Push(ctx, job)
	if err != nil {
		t.Fatalf("Failed to push delayed job: %v", err)
	}

	// Try to pop immediately (should be empty)
	_, err = driver.Pop(ctx, "default")
	if err != dgqueue.ErrQueueEmpty {
		t.Error("Expected queue to be empty for delayed job")
	}

	// Wait for job to become available
	time.Sleep(2500 * time.Millisecond)

	// Pop should now work
	popped, err := driver.Pop(ctx, "default")
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
	ctx := context.Background()

	job := dgqueue.NewJob("failed-job", "payload")

	// Move to failed queue
	err := driver.Failed(ctx, job)
	if err != nil {
		t.Fatalf("Failed to move job to failed queue: %v", err)
	}

	// Verify job is in failed queue
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
	ctx := context.Background()

	// Initially empty
	size, err := driver.Size(ctx, "default")
	if err != nil {
		t.Fatalf("Failed to get size: %v", err)
	}
	if size != 0 {
		t.Errorf("Expected size 0, got %d", size)
	}

	// Add regular job
	job1 := dgqueue.NewJob("job1", "payload")
	driver.Push(ctx, job1)

	// Add delayed job
	job2 := dgqueue.NewJob("job2", "payload")
	dgqueue.WithDelay(job2, 1*time.Hour)
	driver.Push(ctx, job2)

	// Size should include both
	size, err = driver.Size(ctx, "default")
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
	ctx := context.Background()

	job := dgqueue.NewJob("retry-job", "payload")
	job.Attempts = 1

	// Retry
	err := driver.Retry(ctx, job)
	if err != nil {
		t.Fatalf("Failed to retry job: %v", err)
	}

	// Pop and verify
	popped, err := driver.Pop(ctx, "default")
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
	ctx = context.Background()

	job := dgqueue.NewJob("shared-job", "payload")
	err := driver.Push(ctx, job)
	if err != nil {
		t.Fatalf("Failed to push with shared client: %v", err)
	}

	popped, err := driver.Pop(ctx, "default")
	if err != nil {
		t.Fatalf("Failed to pop with shared client: %v", err)
	}

	if popped.ID != job.ID {
		t.Errorf("Expected ID %s, got %s", job.ID, popped.ID)
	}
}
