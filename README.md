# dg-queue

[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.21-blue)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)

A robust, production-ready queue system for Go with Redis backend and batch processing.

## Features

- 🚀 **Multiple Drivers** - Memory (testing) and Redis (production)
- ⏰ **Delayed Jobs** - Schedule jobs to run at a specific time
- 📦 **Batch Processing** - Efficient bulk operations with chunking
- 🔄 **Automatic Retries** - Configurable retry attempts with backoff
- 💀 **Dead Letter Queue** - Failed jobs automatically moved to separate queue
- 👷 **Worker Pools** - Concurrent job processing with configurable workers
- 🔗 **Shared Client** - Reuse Redis connections across components
- 🧹 **Graceful Shutdown** - Proper cleanup of connections and goroutines
- ✅ **Comprehensive Tests** - Full test coverage

## Installation

```bash
go get github.com/donnigundala/dg-queue@v1.0.0
```

## Quick Start

### Basic Usage (Memory Driver)

```go
package main

import (
    "context"
    "fmt"
    "github.com/donnigundala/dg-queue"
)

func main() {
    // Create queue with default config (uses memory driver)
    q := queue.New(queue.DefaultConfig())

    // Register worker
    q.Worker("send-email", 5, func(job *queue.Job) error {
        email := job.Payload.(map[string]interface{})
        fmt.Printf("Sending email to: %s\n", email["to"])
        return nil
    })

    // Start queue
    ctx := context.Background()
    q.Start(ctx)
    defer q.Stop(ctx)

    // Dispatch job
    job, _ := q.Dispatch("send-email", map[string]interface{}{
        "to":      "user@example.com",
        "subject": "Welcome!",
    })
    
    fmt.Printf("Job %s dispatched\n", job.ID)
}
```

### Production Usage (Redis Driver)

```go
import (
    "github.com/donnigundala/dg-queue"
    "github.com/donnigundala/dg-queue/drivers/redis"
    goRedis "github.com/redis/go-redis/v9"
)

// Create Redis driver
driver, _ := redis.NewDriver("myapp", &goRedis.Options{
    Addr: "localhost:6379",
})

q := queue.New(queue.DefaultConfig())
q.SetDriver(driver)
```

### Delayed Jobs

```go
// Dispatch job to run in 5 minutes
q.DispatchAfter("process-payment", payload, 5*time.Minute)
```

### Cron Scheduler

> **Note**: Scheduler has been moved to a separate package!  
> Use [dg-scheduler](https://github.com/donnigundala/dg-scheduler) for cron-based scheduling.

```go
import "github.com/donnigundala/dg-scheduler"

scheduler := scheduler.New(q)
scheduler.Start()
defer scheduler.Stop()

// Run every 5 minutes
scheduler.ScheduleJob("*/5 * * * *", "cleanup", map[string]string{
    "action": "clean_temp_files",
})
```

### Batch Processing

```go
batch := queue.NewBatch(q)

items := []interface{}{
    map[string]string{"email": "user1@test.com"},
    map[string]string{"email": "user2@test.com"},
    // ... 1000 more items
}

config := queue.BatchConfig{
    ChunkSize: 100,  // Process 100 at a time
    OnProgress: func(processed, total int) {
        fmt.Printf("Progress: %d/%d\n", processed, total)
    },
    ContinueOnError: true,
}

status, _ := batch.DispatchBatch("send-email", items, config)
fmt.Printf("Dispatched %d jobs\n", status.Total)
```

## Configuration

```go
config := queue.Config{
    Driver:       "redis",          // or "memory"
    DefaultQueue: "default",
    MaxAttempts:  3,                // Retry up to 3 times
    Timeout:      30 * time.Second, // Job timeout
    RetryDelay:   time.Second,      // Delay between retries
    Workers:      5,                // Worker pool size
}

q := queue.New(config)
```
## Roadmap

- **v1.0.0** - Core queue + Memory driver ✅
- **v1.1.0** - Redis driver ✅
- **v1.2.0** - Batch processing ✅
- **v1.3.0** - Graceful shutdown ✅
- **v2.0.0** - Remove deprecated scheduler (use dg-scheduler)

**📋 Planned:**
- Job chaining
- Middleware system
- Database driver
- Metrics & monitoring

## Examples

See the [examples](./examples) directory for complete examples.

## Migration from v1.x

If you were using the built-in scheduler (`Manager.Schedule()`), please migrate to [dg-scheduler](https://github.com/donnigundala/dg-scheduler):

```go
// Before (deprecated)
q.Schedule("*/5 * * * *", "cleanup", handler)

// After
import "github.com/donnigundala/dg-scheduler"
s := scheduler.New(q)
s.Schedule("*/5 * * * *", "cleanup", handler)
```

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Related Packages

- [dg-scheduler](https://github.com/donnigundala/dg-scheduler) - Cron-based job scheduler
- [dg-core](https://github.com/donnigundala/dg-core) - Core framework
- [dg-cache](https://github.com/donnigundala/dg-cache) - Cache abstraction
- [dg-database](https://github.com/donnigundala/dg-database) - Database plugin
