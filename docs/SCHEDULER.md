# Scheduler

Cron-based job scheduler for recurring tasks.

## Overview

The scheduler uses [robfig/cron](https://github.com/robfig/cron) to execute jobs on a schedule.

## Installation

```bash
go get github.com/robfig/cron/v3
```

## Quick Start

```go
import queue "github.com/donnigundala/dg-queue"

manager := queue.New(queue.DefaultConfig())
scheduler := queue.NewScheduler(manager)

// Start scheduler
scheduler.Start()
defer scheduler.Stop()

// Schedule recurring job
scheduler.ScheduleJob("*/5 * * * *", "cleanup", map[string]string{
    "action": "clean_temp_files",
})
```

## Cron Syntax

Standard cron format with 5 fields:

```
┌───────────── minute (0 - 59)
│ ┌───────────── hour (0 - 23)
│ │ ┌───────────── day of month (1 - 31)
│ │ │ ┌───────────── month (1 - 12)
│ │ │ │ ┌───────────── day of week (0 - 6) (Sunday=0)
│ │ │ │ │
* * * * *
```

### Examples

```go
"* * * * *"      // Every minute
"*/5 * * * *"    // Every 5 minutes
"0 * * * *"      // Every hour
"0 0 * * *"      // Every day at midnight
"0 0 * * 0"      // Every Sunday at midnight
"0 0 1 * *"      // First day of month
"0 9-17 * * 1-5" // Mon-Fri, 9am-5pm
```

### Special Strings

```go
"@hourly"   // Every hour
"@daily"    // Every day at midnight
"@weekly"   // Every Sunday at midnight
"@monthly"  // First day of every month
"@yearly"   // January 1st at midnight
```

## Usage

### Schedule a Job

Dispatch job to queue on schedule:

```go
scheduler.ScheduleJob("0 2 * * *", "daily-report", map[string]interface{}{
    "report_type": "daily",
    "recipients": []string{"admin@example.com"},
})
```

This creates a job named `daily-report` every day at 2 AM.

### Custom Handler

Execute custom logic without queue:

```go
scheduler.Schedule("*/10 * * * *", "health-check", func() error {
    // Check system health
    if !isHealthy() {
        return errors.New("system unhealthy")
    }
    return nil
})
```

### Remove Schedule

```go
err := scheduler.Remove("health-check")
if err != nil {
    log.Printf("Failed to remove schedule: %v", err)
}
```

### Count Schedules

```go
count := scheduler.Count()
fmt.Printf("Active schedules: %d\n", count)
```

## Complete Example

```go
package main

import (
    "context"
    "fmt"
    "time"
    
    queue "github.com/donnigundala/dg-queue"
)

func main() {
    // Create queue manager
    manager := queue.New(queue.DefaultConfig())
    
    // Register worker
    manager.Worker("cleanup", 1, func(job *queue.Job) error {
        fmt.Println("Running cleanup...")
        // Cleanup logic here
        return nil
    })
    
    // Start queue
    ctx := context.Background()
    manager.Start(ctx)
    defer manager.Stop(ctx)
    
    // Create scheduler
    scheduler := queue.NewScheduler(manager)
    scheduler.Start()
    defer scheduler.Stop()
    
    // Schedule jobs
    scheduler.ScheduleJob("*/5 * * * *", "cleanup", map[string]string{
        "type": "temp_files",
    })
    
    scheduler.Schedule("0 * * * *", "heartbeat", func() error {
        fmt.Println("Heartbeat:", time.Now())
        return nil
    })
    
    // Keep running
    select {}
}
```

## Best Practices

### 1. Avoid Overlapping Executions

If job takes longer than schedule interval:

```go
// ❌ Avoid - Can overlap
scheduler.ScheduleJob("* * * * *", "slow-job", data)
// Job takes 2 minutes, but runs every minute

// ✅ Better - Use longer interval
scheduler.ScheduleJob("*/5 * * * *", "slow-job", data)
```

### 2. Handle Errors Gracefully

```go
scheduler.Schedule("*/5 * * * *", "api-sync", func() error {
    if err := syncWithAPI(); err != nil {
        log.Printf("API sync failed: %v", err)
        // Don't return error to keep schedule running
        return nil
    }
    return nil
})
```

### 3. Use Unique Schedule Names

```go
// ✅ Good - Descriptive unique names
scheduler.ScheduleJob("*/5 * * * *", "cleanup-temp-files", data)
scheduler.ScheduleJob("0 0 * * *", "cleanup-old-jobs", data)

// ❌ Avoid - Duplicate names cause errors
scheduler.ScheduleJob("*/5 * * * *", "cleanup", data)
scheduler.ScheduleJob("0 0 * * *", "cleanup", data) // Error!
```

### 4. Start Scheduler After Registering Jobs

```go
// ✅ Good
scheduler := queue.NewScheduler(manager)
scheduler.ScheduleJob("* * * * *", "job1", data)
scheduler.ScheduleJob("* * * * *", "job2", data)
scheduler.Start() // Start after all registered

// ❌ Might miss first execution
scheduler := queue.NewScheduler(manager)
scheduler.Start() // Started before jobs registered
scheduler.ScheduleJob("* * * * *", "job1", data)
```

### 5. Graceful Shutdown

```go
func main() {
    scheduler := queue.NewScheduler(manager)
    scheduler.Start()
    
    // Handle shutdown
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
    
    <-sigChan
    
    // Stop waits for running jobs
    ctx := scheduler.Stop()
    <-ctx.Done()  // Wait for completion
    
    fmt.Println("Scheduler stopped gracefully")
}
```

## Advanced Usage

### Dynamic Scheduling

Add/remove schedules at runtime:

```go
// Add schedule
scheduler.ScheduleJob("0 */6 * * *", "periodic-backup", data)

// Later, remove it
scheduler.Remove("periodic-backup")

// Add new one
scheduler.ScheduleJob("0 0 * * *", "daily-backup", data)
```

### Conditional Execution

```go
scheduler.Schedule("* * * * *", "conditional-job", func() error {
    if shouldRun() {
        return executeJob()
    }
    return nil // Skip this execution
})
```

### Monitoring

```go
// Track execution count
executions := 0
var mu sync.Mutex

scheduler.Schedule("* * * * *", "monitor", func() error {
    mu.Lock()
    executions++
    count := executions
    mu.Unlock()
    
    fmt.Printf("Executed %d times\n", count)
    return nil
})
```

## Troubleshooting

### Schedule Not Running

**Check cron syntax:**
```go
// ❌ Invalid
scheduler.ScheduleJob("*/5", "job", data) // Missing fields

// ✅ Valid
scheduler.ScheduleJob("*/5 * * * *", "job", data)
```

**Verify scheduler started:**
```go
scheduler.Start() // Must call this!
```

**Check for errors:**
```go
err := scheduler.ScheduleJob("*/5 * * * *", "job", data)
if err != nil {
    log.Printf("Schedule error: %v", err)
}
```

### Duplicate Name Error

```
Error: schedule 'job-name' already exists
```

**Solution:**
```go
// Remove existing schedule first
scheduler.Remove("job-name")

// Then add new one
scheduler.ScheduleJob("*/5 * * * *", "job-name", data)
```

### Jobs Running Late

Scheduler timing depends on system clock:

```go
// Jobs run at exact cron times
// "0 0 * * *" runs at 00:00:00, not when scheduler starts
```

**Solution:** If you need immediate execution:
```go
// Dispatch once immediately
manager.Dispatch("job-name", data)

// Then schedule for recurring
scheduler.ScheduleJob("0 0 * * *", "job-name", data)
```

## Performance

- **Overhead:** ~1ms per schedule check
- **Precision:** Second-level (not millisecond)
- **Capacity:** Tested with 1000+ schedules

## Comparison with Other Solutions

| Feature | dg-queue Scheduler | Cron Daemon | AWS EventBridge |
|---------|-------------------|-------------|-----------------|
| Cost | Free | Free | Paid |
| Setup | Code-based | Config file | Web console |
| Dynamic | Yes | No | Yes |
| Monitoring | Application logs | Syslog | CloudWatch |
| Best For | Application-specific | System-wide | Cloud services |
