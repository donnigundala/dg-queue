# Batch Processing

Efficient bulk job dispatching with chunking, progress tracking, and error handling.

## Overview

The batch processor allows you to dispatch thousands of jobs efficiently with:
- **Chunking** - Process large datasets in manageable chunks
- **Progress Callbacks** - Track completion in real-time
- **Error Handling** - Cont inue-on-error or fail-fast modes
- **Map Function** - Transform items before dispatching
- **Rate Limiting** - Control dispatch speed

## Quick Start

```go
import queue "github.com/donnigundala/dg-queue"

manager := queue.New(queue.DefaultConfig())
batch := queue.NewBatch(manager)

items := []interface{}{
    map[string]string{"email": "user1@example.com"},
    map[string]string{"email": "user2@example.com"},
    // ... thousands more
}

config := queue.BatchConfig{
    ChunkSize:       100,
    ContinueOnError: true,
}

status, _ := batch.DispatchBatch("send-email", items, config)
fmt.Printf("Dispatched %d jobs\n", status.Total)
```

## Configuration

### BatchConfig

```go
type BatchConfig struct {
    ChunkSize       int                              // Items per chunk
    ContinueOnError bool                             // Continue if item fails
    OnProgress      func(processed, total int)       // Progress callback
    OnError         func(item interface{}, err error) // Error callback
    RateLimit       int                              // Max items/second (0 = unlimited)
}
```

### Default Configuration

```go
config := queue.DefaultBatchConfig()
// ChunkSize: 100
// ContinueOnError: true
// RateLimit: 0 (unlimited)
```

## Features

### Chunking

Process large datasets in smaller chunks:

```go
items := make([]interface{}, 10000) // 10K items

config := queue.BatchConfig{
    ChunkSize: 100, // Process 100 at a time
}

batch.DispatchBatch("process-item", items, config)
// Creates 100 chunks of 100 items each
```

**Benefits:**
- Lower memory pressure
- Better error isolation
- Progress tracking

### Progress Tracking

Monitor batch completion:

```go
config := queue.BatchConfig{
    ChunkSize: 100,
    OnProgress: func(processed, total int) {
        progress := float64(processed) / float64(total) * 100
        fmt.Printf("Progress: %.2f%% (%d/%d)\n", progress, processed, total)
    },
}

status, _ := batch.DispatchBatch("job", items, config)

// Wait for completion
for !status.IsComplete() {
    time.Sleep(100 * time.Millisecond)
}

fmt.Printf("Final progress: %.2f%%\n", status.Progress())
```

### Error Handling

#### Continue on Error

```go
config := queue.BatchConfig{
    ContinueOnError: true,
    OnError: func(item interface{}, err error) {
        log.Printf("Failed to process %v: %v", item, err)
    },
}

status, _ := batch.DispatchBatch("job", items, config)

fmt.Printf("Processed: %d, Failed: %d\n", status.Processed, status.Failed)
```

#### Fail Fast

```go
config := queue.BatchConfig{
    ContinueOnError: false,
    OnError: func(item interface{}, err error) {
        log.Printf("Batch failed at item %v: %v", item, err)
    },
}

status, err := batch.DispatchBatch("job", items, config)
if err != nil {
    // Batch stopped on first error
    fmt.Printf("Stopped after %d items\n", status.Processed)
}
```

### Map Function

Transform items before dispatching:

```go
items := []interface{}{1, 2, 3, 4, 5}

mapper := func(item interface{}) (interface{}, error) {
    num := item.(int)
    return map[string]int{
        "value":  num,
        "square": num * num,
    }, nil
}

status, _ := batch.Map("process-number", items, mapper, queue.DefaultBatchConfig())
```

**Use Cases:**
- Data enrichment
- Format conversion
- Validation before dispatch

### Rate Limiting

Control dispatch speed:

```go
config := queue.BatchConfig{
    ChunkSize: 100,
    RateLimit: 1000, // Max 1000 items/second
}

batch.DispatchBatch("api-call", items, config)
// Automatically throttles to stay under limit
```

## Complete Example

```go
package main

import (
    "fmt"
    "log"
    "time"
    
    queue "github.com/donnigundala/dg-queue"
)

func main() {
    // Setup
    manager := queue.New(queue.DefaultConfig())
    batch := queue.NewBatch(manager)
    
    // Register worker
    manager.Worker("send-email", 10, func(job *queue.Job) error {
        email := job.Payload.(map[string]interface{})
        fmt.Printf("Sending to: %s\n", email["to"])
        time.Sleep(100 * time.Millisecond) // Simulate work
        return nil
    })
    
    // Start queue
    manager.Start(context.Background())
    defer manager.Stop(context.Background())
    
    // Prepare items
    items := make([]interface{}, 1000)
    for i := 0; i < 1000; i++ {
        items[i] = map[string]interface{}{
            "to":      fmt.Sprintf("user%d@example.com", i),
            "subject": "Hello",
        }
    }
    
    // Configure batch
    config := queue.BatchConfig{
        ChunkSize: 100,
        OnProgress: func(processed, total int) {
            fmt.Printf("Progress: %d/%d (%.1f%%)\n", 
                processed, total, float64(processed)/float64(total)*100)
        },
        OnError: func(item interface{}, err error) {
            log.Printf("Error: %v", err)
        },
        ContinueOnError: true,
        RateLimit:       500, // 500 emails/sec
    }
    
    // Dispatch batch
    status, err := batch.DispatchBatch("send-email", items, config)
    if err != nil {
        log.Fatalf("Batch failed: %v", err)
    }
    
    // Monitor completion
    for !status.IsComplete() {
        time.Sleep(1 * time.Second)
        fmt.Printf("Status: Processed=%d, Failed=%d, Progress=%.1f%%\n",
            status.Processed, status.Failed, status.Progress())
    }
    
    fmt.Printf("Batch complete! Processed=%d, Failed=%d\n", 
        status.Processed, status.Failed)
}
```

## Best Practices

### 1. Choose Appropriate Chunk Size

```go
// ✅ Good - 100 for API calls
config.ChunkSize = 100

// ✅ Good - 1000 for database inserts
config.ChunkSize = 1000

// ❌ Too small - overhead
config.ChunkSize = 1

// ❌ Too large - memory issues
config.ChunkSize = 100000
```

**Rule of thumb:** Start with 100, adjust based on job complexity.

### 2. Always Handle Errors

```go
config := queue.BatchConfig{
    OnError: func(item interface{}, err error) {
        // ✅ Log error
        log.Printf("Failed: %v - %v", item, err)
        
        // ✅ Store in database
        db.SaveFailedItem(item, err)
        
        // ✅ Send alert
        if criticalError(err) {
            sendAlert(err)
        }
    },
    ContinueOnError: true,
}
```

### 3. Use Progress Callbacks Wisely

```go
// ✅ Good - Update every N items
lastUpdate := 0
config.OnProgress = func(processed, total int) {
    if processed-lastUpdate >= 100 {
        fmt.Printf("Progress: %d/%d\n", processed, total)
        lastUpdate = processed
    }
}

// ❌ Avoid - Too frequent
config.OnProgress = func(processed, total int) {
    fmt.Printf("Progress: %d/%d\n", processed, total) // Every item!
}
```

### 4. Set Rate Limits for External APIs

```go
// Respecting API rate limits
config := queue.BatchConfig{
    RateLimit: 100, // Max 100 calls/sec per API docs
    OnError: func(item interface{}, err error) {
        if isRateLimitError(err) {
            time.Sleep(1 * time.Second) // Backoff
        }
    },
}
```

### 5. Validate Items Before Batch

```go
// ✅ Pre-validate
validItems := []interface{}{}
for _, item := range items {
    if isValid(item) {
        validItems = append(validItems, item)
    }
}
batch.DispatchBatch("job", validItems, config)

// ❌ Validate in dispatcher (wastes resources)
batch.DispatchBatch("job", items, config)
// Worker validates each item
```

## Performance

### Throughput

With default settings:
- **Small Jobs** (<1ms): ~10,000 items/sec
- **Medium Jobs** (~10ms): ~1,000 items/sec  
- **Large Jobs** (~100ms): ~100 items/sec

### Optimization Tips

1. **Increase Chunk Size** - Reduce overhead
   ```go
   config.ChunkSize = 1000 // vs 100
   ```

2. **Reduce Callbacks** - Less logging
   ```go
   config.OnProgress = nil // Disable if not needed
   ```

3. **Use Worker Pools** - Parallel processing
   ```go
   manager.Worker("job", 20, handler) // 20 concurrent workers
   ```

4. **Batch Database Operations** - In handler
   ```go
   manager.Worker("save-user", 5, func(job *queue.Job) error {
       // Batch insert 100 users at once
       return db.BatchInsert(users)
   })
   ```

## Troubleshooting

### Slow Processing

**Check worker concurrency:**
```go
// Increase workers
manager.Worker("slow-job", 50, handler) // Was 5, now 50
```

**Check chunk size:**
```go
// Larger chunks = less overhead
config.ChunkSize = 500 // Was 100
```

### High Memory Usage

**Reduce chunk size:**
```go
config.ChunkSize = 50 // Was 1000
```

**Process in smaller batches:**
```go
// Instead of all 100K at once
for i := 0; i < len(allItems); i += 10000 {
    chunk := allItems[i:min(i+10000, len(allItems))]
    batch.DispatchBatch("job", chunk, config)
}
```

### Items Not Processing

**Verify worker registered:**
```go
// Worker name must match dispatch name
manager.Worker("send-email", 5, handler)
batch.DispatchBatch("send-email", items, config) // ✅ Matches

batch.DispatchBatch("send_email", items, config)  // ❌ No worker
```

**Check for errors:**
```go
status, err := batch.DispatchBatch("job", items, config)
if err != nil {
    log.Printf("Batch error: %v", err)
}
if status.Failed > 0 {
    log.Printf("%d items failed", status.Failed)
}
```

## Comparison

| Approach | Throughput | Memory | Error Handling | Progress |
|----------|-----------|--------|----------------|----------|
| Manual Loop | Low | Low | Manual |  Manual |
| Batch (ChunkSize=100) | High | Medium | Built-in | Built-in |
| Batch (ChunkSize=1000) | Highest | High | Built-in | Built-in |

## When to Use

✅ **Use Batch When:**
- Dispatching >100 jobs
- Need progress tracking
- Need error handling
- Want rate limiting

❌ **Use Regular Dispatch When:**
- Single job or few jobs
- No need for progress tracking
- Individual job control needed
