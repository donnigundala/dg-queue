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

```go
package main

import (
    "github.com/donnigundala/dg-core/foundation"
    "github.com/donnigundala/dg-queue"
)

func main() {
    app := foundation.New(".")
    
    // Register provider (uses 'queue' key in config)
    app.Register(dgqueue.NewQueueServiceProvider(nil))
    
    app.Start()
    
    // Usage
    q := dgqueue.MustResolve(app)
    q.Dispatch("send-email", map[string]interface{}{"to": "user@test.com"})
}
```

### Integration via InfrastructureSuite
In your `bootstrap/app.go`, you typically use the declarative suite pattern:

```go
func InfrastructureSuite(workerMode bool) []foundation.ServiceProvider {
	// 1. Add Queue (Always register for dispatching)
	queueProvider := dgqueue.NewQueueServiceProvider(nil)
    
	// 2. Inject mode-based worker state
	queueProvider.Config.WorkerEnabled = workerMode
    
	return []foundation.ServiceProvider{
		queueProvider,
		// ... other providers
	}
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

The plugin uses the `queue` key in your configuration file.

### Configuration Mapping (YAML vs ENV)

| YAML Key | Environment Variable | Default | Description |
| :--- | :--- | :--- | :--- |
| `queue.driver` | `QUEUE_DRIVER` | `memory` | `redis`, `memory` |
| `queue.connection` | `QUEUE_CONNECTION` | `default` | Redis connection name |
| `queue.prefix` | `QUEUE_PREFIX` | `queue_` | Key prefix |
| `queue.default_queue` | `QUEUE_DEFAULT_QUEUE` | `default` | Default queue name |
| `queue.max_attempts` | `QUEUE_MAX_ATTEMPTS` | `3` | Max retry attempts |
| `queue.timeout` | `QUEUE_TIMEOUT` | `30s` | Job timeout duration |
| `queue.worker_enabled` | `QUEUE_WORKER_ENABLED` | `true` | Start worker loop |
| `queue.workers` | `QUEUE_WORKERS` | `5` | Number of concurrent workers |

### Example YAML

```yaml
queue:
  driver: "redis"
  connection: "default"
  worker_enabled: false # Set to true on Worker nodes
  workers: 10
  timeout: 60s
```

## Container Integration (v1.6.0+)

dg-queue provides first-class support for the `dg-core` container system.

```go
import (
    "github.com/donnigundala/dg-queue"
    "github.com/donnigundala/dg-core/contracts/foundation"
)

// 1. Resolve using helper functions
q := queue.MustResolve(app)
q.Dispatch("email", payload)

// 2. Inject into your services
type UserService struct {
    *queue.Injectable
}

func NewUserService(app foundation.Application) *UserService {
    return &UserService{
        Injectable: queue.NewInjectable(app),
    }
}

func (s *UserService) Register() {
    // Access queue directly
    s.Queue().Dispatch("welcome-email", payload)
}
```
## Roadmap

- **v1.0.0** - Core queue + Memory driver ✅
- **v1.1.0** - Redis driver ✅
- **v1.2.0** - Batch processing ✅
- **v1.3.0** - Graceful shutdown ✅
- **v2.0.0** - Remove deprecated scheduler
- **v2.1.0** - Metrics & monitoring ✅

**📋 Planned:**
- Job chaining
- Middleware system
- Database driver

## 📊 Observability

`dg-queue` is instrumented with OpenTelemetry metrics. If `dg-observability` is registered and enabled, the following metrics are automatically collected:

*   `queue.job.count`: Counter (labels: `queue`, `job_name`, `status`) - tracks processed jobs.
*   `queue.job.duration`: Histogram (labels: `queue`, `job_name`, `status`) - execution time in milliseconds.
*   `queue.depth`: Gauge (labels: `queue`) - number of pending jobs (Redis only).
*   `queue.workers.active`: Gauge (labels: `queue`) - number of workers currently processing jobs.

To enable observability, ensure the `dg-observability` plugin is registered and configured:

```yaml
observability:
  enabled: true
  service_name: "my-app"
```

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
