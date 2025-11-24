# dg-queue

[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.21-blue)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)

A unified queue and scheduler system for the dg-framework. Simpler than Laravel Queue + Scheduler, but just as powerful.

## Features

### 🚀 Core Features
- **Queue System** - Asynchronous job processing
- **Built-in Scheduler** - Cron-like task scheduling (coming soon)
- 🚀 **Multiple Drivers** - Memory (testing) and Redis (production)
- ⏰ **Delayed Jobs** - Schedule jobs to run at a specific time
- 🔄 **Automatic Retries** - Configurable retry attempts with exponential backoff
- 💀 **Dead Letter Queue** - Failed jobs are moved to a separate queue
- 👷 **Worker Pools** - Concurrent job processing with configurable workers
- 🔗 **Shared Client** - Reuse Redis connections across components
- ✅ **Comprehensive Tests** - 27 tests covering all features

## Installation

```bash
go get github.com/donnigundala/dg-queue@v1.0.0
```

## Quick Start


import (
    "context"
    "fmt"
    "time"
    
    "github.com/donnigundala/dg-queue"
    "github.com/donnigundala/dg-queue/drivers/memory"
)

func main() {
    // Create queue
    q := queue.New(queue.DefaultConfig())
    q.SetDriver(memory.NewDriver())
    
    // Register worker
    q.Worker("send-email", 5, func(job *queue.Job) error {
        email := job.Payload.(map[string]interface{})
        fmt.Printf("Sending email to: %s\n", email["to"])
        return nil
    })
    
    // Start queue
    ctx := context.Background()
    q.Start(ctx)
    
    // Dispatch job
    q.Dispatch("send-email", map[string]interface{}{
        "to":      "user@example.com",
        "subject": "Welcome!",
    })
    
    // Let it process
    time.Sleep(1 * time.Second)
    
    // Stop gracefully
    q.Stop(ctx)
}
```

## Current Status

**✅ Implemented:**
- Core queue interfaces
- Job serialization and lifecycle
- Memory driver for testing
- Worker pools with concurrency
- Job retry with backoff
- Graceful shutdown

**🚧 In Progress:**
- Redis driver (Phase 3)
- Cron scheduler (Phase 4)
- Batch processing (Phase 5)

**📋 Planned:**
- Job chaining
- Middleware system
- Dead letter queue
- Database driver
- Metrics & monitoring

## Roadmap

- **v0.1.0** - Core queue + Memory driver ✅ (current)
- **v0.2.0** - Redis driver
- **v0.3.0** - Scheduler
- **v0.4.0** - Batch processing
- **v1.0.0** - Production ready

## Examples

See the [examples](./examples) directory for complete examples.

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Related Packages

- [dg-core](https://github.com/donnigundala/dg-core) - Core framework
- [dg-cache](https://github.com/donnigundala/dg-cache) - Cache abstraction
- [dg-database](https://github.com/donnigundala/dg-database) - Database plugin
