# Testing

Comprehensive testing guide for dg-queue.

## Test Coverage

**Total: 42 tests passing**

| Component | Tests | Coverage |
|-----------|-------|----------|
| Core (Job, Config) | 16 | ~95% |
| Memory Driver | 5 | 100% |
| Redis Driver | 6 | ~90% |
| Scheduler | 7 | ~85% |
| Batch Processing | 8 | ~90% |

## Running Tests

### All Tests

```bash
go test ./...
```

### Specific Package

```bash
go test ./drivers/memory -v
go test ./drivers/redis -v
```

### With Coverage

```bash
go test ./... -cover
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Short Mode (Skip Slow Tests)

```bash
go test ./... -short
```

## Test Structure

### Core Tests

Located in root package:

```
job_test.go       # Job struct and methods
config_test.go    # Configuration
scheduler_test.go # Scheduler functionality
batch_test.go     # Batch processing
```

### Driver Tests

Each driver has its own tests:

```
drivers/memory/memory_test.go
drivers/redis/redis_test.go
```

## Writing Tests

### Basic Job Test

```go
func TestJobCreation(t *testing.T) {
    job := queue.NewJob("test-job", map[string]string{
        "key": "value",
    })
    
    if job.Name != "test-job" {
        t.Errorf("Expected name 'test-job', got %s", job.Name)
    }
    
    if job.Queue != "default" {
        t.Errorf("Expected queue 'default', got %s", job.Queue)
    }
}
```

### Worker Test

```go
func TestWorker(t *testing.T) {
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
    
    manager.Dispatch("test-job", "payload")
    
    time.Sleep(100 * time.Millisecond)
    
    if !processed {
        t.Error("Job was not processed")
    }
}
```

### Error Handling Test

```go
func TestJobRetry(t *testing.T) {
    manager := queue.New(queue.Config{
        MaxAttempts: 3,
        RetryDelay:  10 * time.Millisecond,
    })
    manager.SetDriver(memory.NewDriver())
    
    attempts := 0
    manager.Worker("retry-job", 1, func(job *queue.Job) error {
        attempts++
        if attempts < 3 {
            return errors.New("temporary error")
        }
        return nil
    })
    
    ctx := context.Background()
    manager.Start(ctx)
    defer manager.Stop(ctx)
    
    manager.Dispatch("retry-job", "payload")
    
    time.Sleep(200 * time.Millisecond)
    
    if attempts != 3 {
        t.Errorf("Expected 3 attempts, got %d", attempts)
    }
}
```

## Testing with Memory Driver

Always use memory driver for tests:

```go
func TestExample(t *testing.T) {
    // ✅ Good - Fast, no external dependencies
    manager := queue.New(queue.DefaultConfig())
    manager.SetDriver(memory.NewDriver())
    
    // ❌ Avoid - Requires Redis
    driver, _ := redis.NewDriver("test", options)
    manager.SetDriver(driver)
}
```

## Testing Redis Driver

Tests automatically skip if Redis unavailable:

```go
func TestRedis(t *testing.T) {
    client := redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
    })
    
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()
    
    if err := client.Ping(ctx).Err(); err != nil {
        t.Skip("Redis not available, skipping test")
    }
    
    // Test code here
}
```

### Running Redis Tests

```bash
# Start Redis
 docker run -d -p 6379:6379 redis:7-alpine

# Run tests
go test ./drivers/redis -v
```

## Testing Scheduler

Use short intervals for testing:

```go
func TestScheduler(t *testing.T) {
    manager := queue.New(queue.DefaultConfig())
    scheduler := queue.NewScheduler(manager)
    scheduler.Start()
    defer scheduler.Stop()
    
    executed := false
    var mu sync.Mutex
    
    // Use every minute for real cron
    // For testing, check syntax only
    err := scheduler.Schedule("*/5 * * * *", "test", func() error {
        mu.Lock()
        executed = true
        mu.Unlock()
        return nil
    })
    
    if err != nil {
        t.Fatalf("Failed to schedule: %v", err)
    }
    
    // Verify schedule was added
    if scheduler.Count() != 1 {
        t.Error("Schedule not added")
    }
}
```

## Testing Batch Processing

```go
func TestBatch(t *testing.T) {
    manager := queue.New(queue.DefaultConfig())
    manager.SetDriver(memory.NewDriver())
    batch := queue.NewBatch(manager)
    
    items := make([]interface{}, 100)
    for i := 0; i < 100; i++ {
        items[i] = i
    }
    
    config := queue.BatchConfig{
        ChunkSize:       10,
        ContinueOnError: true,
    }
    
    status, err := batch.DispatchBatch("test", items, config)
    if err != nil {
        t.Fatalf("Batch failed: %v", err)
    }
    
    // Wait for completion
    time.Sleep(200 * time.Millisecond)
    
    if status.Total != 100 {
        t.Errorf("Expected 100 items, got %d", status.Total)
    }
    
    if !status.IsComplete() {
        t.Error("Batch not complete")
    }
}
```

## Test Helpers

### Setup Helper

```go
func setupTest(t *testing.T) (*queue.Manager, func()) {
    manager := queue.New(queue.DefaultConfig())
    manager.SetDriver(memory.NewDriver())
    
    ctx := context.Background()
    manager.Start(ctx)
    
    cleanup := func() {
        manager.Stop(ctx)
    }
    
    return manager, cleanup
}

func TestWithHelper(t *testing.T) {
    manager, cleanup := setupTest(t)
    defer cleanup()
    
    // Test code
}
```

### Wait Helper

```go
func waitForJob(t *testing.T, done *bool, timeout time.Duration) {
    deadline := time.Now().Add(timeout)
    for time.Now().Before(deadline) {
        if *done {
            return
        }
        time.Sleep(10 * time.Millisecond)
    }
    t.Fatal("Timeout waiting for job")
}
```

## Best Practices

### 1. Use Table-Driven Tests

```go
func TestJobStates(t *testing.T) {
    tests := []struct {
        name     string
        setup    func(*queue.Job)
        expected string
    }{
        {
            name: "pending job",
            setup: func(j *queue.Job) {},
            expected: "pending",
        },
        {
            name: "started job",
            setup: func(j *queue.Job) { j.MarkStarted() },
            expected: "processing",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            job := queue.NewJob("test", "payload")
            tt.setup(job)
            
            if got := job.GetStatus(); got != tt.expected {
                t.Errorf("Expected %s, got %s", tt.expected, got)
            }
        })
    }
}
```

### 2. Clean Up Resources

```go
func TestExample(t *testing.T) {
    manager := queue.New(queue.DefaultConfig())
    ctx := context.Background()
    manager.Start(ctx)
    
    // ✅ Always clean up
    defer manager.Stop(ctx)
    
    // Test code
}
```

### 3. Test Error Cases

```go
func TestErrors(t *testing.T) {
    // Test nil payload
    job := queue.NewJob("test", nil)
    if job.Payload != nil {
        t.Error("Nil payload should be preserved")
    }
    
    // Test invalid config
    config := queue.Config{MaxAttempts: -1}
    // Should handle invalid config
}
```

### 4. Use Subtests

```go
func TestJob(t *testing.T) {
    t.Run("creation", func(t *testing.T) {
        job := queue.NewJob("test", "payload")
        // Assert creation
    })
    
    t.Run("with delay", func(t *testing.T) {
        job := queue.NewJob("test", "payload").
            WithDelay(5 * time.Minute)
        // Assert delay
    })
}
```

## Continuous Integration

### GitHub Actions

```yaml
name: Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    
    services:
      redis:
        image: redis:7-alpine
        ports:
          - 6379:6379
        options: >-
          --health-cmd "redis-cli ping"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
    
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Run tests
        run: go test ./... -v -race -coverprofile=coverage.out
      
      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          file: ./coverage.out
```

## Benchmarking

```go
func BenchmarkDispatch(b *testing.B) {
    manager := queue.New(queue.DefaultConfig())
    manager.SetDriver(memory.NewDriver())
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        manager.Dispatch("test", "payload")
    }
}

func BenchmarkBatchDispatch(b *testing.B) {
    manager := queue.New(queue.DefaultConfig())
    manager.SetDriver(memory.NewDriver())
    batch := queue.NewBatch(manager)
    
    items := make([]interface{}, 100)
    for i := 0; i < 100; i++ {
        items[i] = i
    }
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        batch.DispatchBatch("test", items, queue.DefaultBatchConfig())
    }
}
```

Run benchmarks:

```bash
go test -bench=. -benchmem
```

## Common Issues

### Race Conditions

Always test with race detector:

```bash
go test ./... -race
```

### Timing Issues

Use `time.Sleep` conservatively:

```go
// ❌ Flaky - May fail on slow CI
time.Sleep(10 * time.Millisecond)
assert.True(t, processed)

// ✅ Better - Wait with timeout
waitForJob(t, &processed, 1*time.Second)
```

### External Dependencies

Mock external services:

```go
type MockEmailService struct {
    SentEmails []string
}

func (m *MockEmailService) Send(to string) error {
    m.SentEmails = append(m.SentEmails, to)
    return nil
}
```
