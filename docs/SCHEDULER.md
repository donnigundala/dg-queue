# Scheduler

> **⚠️ DEPRECATED**
>
> The scheduler functionality has been moved to a dedicated package: [dg-scheduler](https://github.com/donnigundala/dg-scheduler).
>
> This document is kept for historical purposes and will be removed in v2.0.0.

## Migration Guide

To use the scheduler in your application, please install the new package:

```bash
go get github.com/donnigundala/dg-scheduler
```

### Example Usage

```go
import (
    "github.com/donnigundala/dg-scheduler"
    "github.com/donnigundala/dg-queue"
)

// Create queue
q := queue.New(queue.DefaultConfig())

// Create scheduler with queue dispatcher
s := scheduler.New(q)

// Schedule job
s.Schedule("*/5 * * * *", "cleanup", nil)

// Start scheduler
s.Start()
```

For full documentation, please visit the [dg-scheduler repository](https://github.com/donnigundala/dg-scheduler).
