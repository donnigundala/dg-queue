# Memory Driver

The memory driver provides an in-memory queue implementation perfect for development and testing.

## Overview

- **Storage:** In-memory maps and slices
- **Persistence:** None (data lost on restart)
- **Thread-Safe:** Yes (uses `sync.RWMutex`)
- **Use Case:** Development, testing, ephemeral jobs

## Usage

```go
import (
    queue "github.com/donnigundala/dg-queue"
    "github.com/donnigundala/dg-queue/drivers/memory"
)

manager := queue.New(queue.DefaultConfig())
manager.SetDriver(memory.NewDriver())
```

## Features

### Simple Setup

No configuration required - just create and use:

```go
driver := memory.NewDriver()
```

### FIFO Queue

Jobs processed in First-In-First-Out order:

```go
manager.Dispatch("job1", "payload1")
manager.Dispatch("job2", "payload2")
// job1 processes first, then job2
```

### Job Storage

All jobs stored in memory with full state tracking:

```go
status, _ := manager.Status(jobID)
fmt.Println(status.Status) // "pending", "processing", "completed", "failed"
```

## Testing Example

```go
func TestJobProcessing(t *testing.T) {
    manager := queue.New(queue.DefaultConfig())
    manager.SetDriver(memory.NewDriver())
    
    processed := false
    manager.Worker("test-job", 1, func(job *queue.Job) error {
        processed = true
        return nil
    })
    
    ctx := context.Background()
    manager.Start(ctx)
    defer manager.Stop(ctx)
    
    manager.Dispatch("test-job", "test-payload")
    
    time.Sleep(100 * time.Millisecond)
    assert.True(t, processed)
}
```

## Limitations

### No Persistence

Data is lost when process stops:

```go
manager.Dispatch("job", "data")
// Process restarts
// Job is gone ❌
```

**Solution:** Use Redis driver for production.

### Single Process Only

Cannot share queue across multiple processes:

```go
// Process A
manager.Dispatch("job", "data")

// Process B (different instance)
// Cannot see job ❌
```

**Solution:** Use Redis driver for distributed systems.

### Memory Growth

All jobs stored in memory until explicitly deleted:

```go
// Dispatching 1M jobs
for i := 0; i < 1000000; i++ {
    manager.Dispatch("job", data) // Memory grows continuously
}
```

**Best Practice:** Clean up completed jobs or use Redis with TTL.

## Best Practices

### 1. Use for Testing Only

```go
// ✅ Good
func TestQueue(t *testing.T) {
    manager.SetDriver(memory.NewDriver())
}

// ❌ Avoid in production
func main() {
    manager.SetDriver(memory.NewDriver()) // Data loss risk!
}
```

### 2. Clean Up Tests

```go
func TestWithCleanup(t *testing.T) {
    manager := queue.New(queue.DefaultConfig())
    manager.SetDriver(memory.NewDriver())
    
    defer manager.Stop(context.Background())
    
    // Your test code
}
```

### 3. Mock External Dependencies

```go
func TestEmailJob(t *testing.T) {
    // Use memory driver for queue
    manager.SetDriver(memory.NewDriver())
    
    // Mock email service
    emailMock := &MockEmailService{}
    
    manager.Worker("send-email", 1, func(job *queue.Job) error {
        return emailMock.Send(job.Payload)
    })
}
```

## Implementation Details

### Data Structures

```go
type Driver struct {
    mu      sync.RWMutex
    queues  map[string][]*Job  // Queue name → jobs
    jobs    map[string]*Job     // Job ID → job
    failed  []*Job              // Failed jobs
}
```

### Thread Safety

All operations protected by mutex:

```go
func (d *Driver) Push(job *Job) error {
    d.mu.Lock()
    defer d.mu.Unlock()
    
    // Add to queue
    d.queues[job.Queue] = append(d.queues[job.Queue], job)
    d.jobs[job.ID] = job
    
    return nil
}
```

### Memory Footprint

Approximate memory per job:
- Job struct: ~200 bytes
- Payload: varies (can be large!)
- Metadata: ~100 bytes

**Total:** ~300 bytes + payload size

## When to Use

✅ **Use When:**
- Unit testing queue logic
- Integration testing without external dependencies  
- Local development
- Temporary/ephemeral jobs
- Single-process applications (non-critical)

❌ **Don't Use When:**
- Production workloads
- Multi-process applications
- Jobs must survive restarts
- High job volume (>10K queued jobs)

## Migration to Redis

Easy migration when ready for production:

```go
// Development
driver := memory.NewDriver()

// Production  
driver, _ := redis.NewDriver("myapp", &redis.Options{
    Addr: "localhost:6379",
})
```

No code changes needed - just swap the driver!
