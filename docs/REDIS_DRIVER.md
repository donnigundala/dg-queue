# Redis Driver

Production-ready Redis driver with delayed jobs and dead letter queue support.

## Overview

- **Storage:** Redis (persistent)
- **Persistence:** Yes (survives restarts)
- **Thread-Safe:** Yes
- **Distributed:** Yes (multi-process)
- **Use Case:** Production workloads

## Installation

```bash
go get github.com/redis/go-redis/v9
```

## Quick Start

### Option 1: New Redis Client

```go
import (
    queue "github.com/donnigundala/dg-queue"
    "github.com/donnigundala/dg-queue/drivers/redis"
    goRedis "github.com/redis/go-redis/v9"
)

driver, err := redis.NewDriver("myapp", &goRedis.Options{
    Addr:     "localhost:6379",
    Password: "", // no password
    DB:       0,  // default DB
})

manager := queue.New(queue.DefaultConfig())
manager.SetDriver(driver)
```

### Option 2: Shared Redis Client (Recommended)

```go
// Create ONE Redis client
client := goRedis.NewClient(&goRedis.Options{
    Addr: "localhost:6379",
    PoolSize: 50,
    MinIdleConns: 10,
})

// Share with cache
cacheDriver := redisCache.NewDriverWithClient(client, "cache")

// Share with queue
queueDriver := redis.NewDriverWithClient(client, "queue")
```

**Benefits:**
- Single connection pool
- Lower memory footprint
- Better resource utilization

## Features

### Delayed Jobs

Schedule jobs to run in the future:

```go
// Run in 5 minutes
job := queue.NewJob("send-reminder", userData).
    WithDelay(5 * time.Minute)

manager.Dispatch("send-reminder", userData)
```

**Implementation:** Redis sorted sets with Unix timestamp as score.

### Dead Letter Queue

Failed jobs automatically moved to separate queue:

```go
// Job fails after 3 attempts
// Automatically moved to "queue:failed" list
```

**Access failed jobs:**

```go
// In Redis CLI
LRANGE queue:failed 0 -1
```

### Persistent Storage

Jobs survive application restarts:

```go
manager.Dispatch("important-job", data)
// App crashes
// App restarts
// Job still in queue ✅
```

## Configuration

### Redis Options

```go
driver, _ := redis.NewDriver("myapp", &goRedis.Options{
    // Connection
    Addr:     "localhost:6379",
    Password: "secret",
    DB:       0,
    
    // Connection Pool
    PoolSize:     50,
    MinIdleConns: 10,
    MaxRetries:   3,
    
    // Timeouts
    DialTimeout:  5 * time.Second,
    ReadTimeout:  3 * time.Second,
    WriteTimeout: 3 * time.Second,
    
    // Connection health
    PoolTimeout: 4 * time.Second,
})
```

### Queue Prefix

Separate applications on same Redis instance:

```go
// App 1
driver, _ := redis.NewDriver("app1", options) // Keys: app1:queues:*

// App 2
driver, _ := redis.NewDriver("app2", options) // Keys: app2:queues:*
```

## Redis Key Structure

### Regular Queue

```
{prefix}:queues:{queue_name}
```

Example: `myapp:queues:default`

**Type:** List (LPUSH/RPOP)

### Delayed Queue

```
{prefix}:queues:{queue_name}:delayed
```

Example: `myapp:queues:default:delayed`

**Type:** Sorted Set (ZADD/ZRANGEBYSCORE)  
**Score:** Unix timestamp when job becomes available

### Failed Queue

```
{prefix}:failed
```

Example: `myapp:failed`

**Type:** List (RPUSH)

## How It Works

### Job Dispatch

```go
manager.Dispatch("send-email", payload)
```

1. Job serialized to JSON
2. Pushed to Redis list: `RPUSH myapp:queues:default {json}`
3. Returns job ID

### Job Processing

```go
// Manager pops job
1. LPOP myapp:queues:default
2. Deserialize JSON → Job
3. Route to worker pool
4. Worker executes handler
```

### Delayed Jobs

```go
job.WithDelay(5 * time.Minute)
```

1. Calculate available time: `now + 5 minutes`
2. Add to sorted set: `ZADD queue:delayed {timestamp} {json}`
3. Background process moves available jobs:
   ```
   ZRANGEBYSCORE queue:delayed -inf {now}
   → RPUSH queue:default {json}
   → ZREM queue:delayed {json}
   ```

## Best Practices

### 1. Use Shared Client

```go
// ✅ Good - One connection pool
client := redis.NewClient(options)
cacheDriver := redisCache.NewDriverWithClient(client, "cache")
queueDriver := redisQueue.NewDriverWithClient(client, "queue")

// ❌ Avoid - Multiple connection pools
cache := redis.NewClient(options)
queue := redis.NewClient(options)
```

### 2. Set Appropriate Pool Size

```go
driver, _ := redis.NewDriver("app", &redis.Options{
    PoolSize: runtime.NumCPU() * 10, // Good starting point
    MinIdleConns: 5,
})
```

### 3. Monitor Redis Memory

```bash
# Check memory usage
redis-cli INFO memory

# Check queue sizes
redis-cli LLEN myapp:queues:default
redis-cli ZCARD myapp:queues:default:delayed
```

### 4. Set TTL for Completed Jobs

```go
// In production, implement cleanup
// Redis doesn't auto-delete completed jobs
```

### 5. Use Connection Timeouts

```go
driver, _ := redis.NewDriver("app", &redis.Options{
    DialTimeout:  5 * time.Second,  // Prevent hanging on connect
    ReadTimeout:  3 * time.Second,  // Prevent slow reads
    WriteTimeout: 3 * time.Second,  // Prevent slow writes
})
```

## Troubleshooting

### Connection Refused

```
Error: dial tcp 127.0.0.1:6379: connect: connection refused
```

**Solution:**
```bash
# Start Redis
redis-server

# Or with Docker
docker run -d -p 6379:6379 redis:7-alpine
```

### Jobs Not Processing

**Check queue size:**
```bash
redis-cli LLEN myapp:queues:default
```

**Check delayed queue:**
```bash
redis-cli ZCARD myapp:queues:default:delayed
redis-cli ZRANGE myapp:queues:default:delayed 0 -1 WITHSCORES
```

**Verify workers registered:**
```go
manager.Worker("job-name", concurrency, handler) // Must match dispatch name
```

### Memory Growth

**Monitor failed queue:**
```bash
redis-cli LLEN myapp:failed
```

**Implement cleanup:**
```go
// Periodically clear old failed jobs
client.LTrim("myapp:failed", 0, 1000) // Keep last 1000
```

## High Availability

### Redis Sentinel

```go
import "github.com/redis/go-redis/v9"

client := goRedis.NewFailoverClient(&goRedis.FailoverOptions{
    MasterName:    "mymaster",
    SentinelAddrs: []string{":26379", ":26380", ":26381"},
})

driver := redis.NewDriverWithClient(client, "myapp")
```

### Redis Cluster

```go
client := goRedis.NewClusterClient(&goRedis.ClusterOptions{
    Addrs: []string{":7000", ":7001", ":7002"},
})

driver := redis.NewDriverWithClient(client, "myapp")
```

## Performance

### Throughput

- **Push:** ~10,000 ops/sec (single instance)
- **Pop:**  ~8,000 ops/sec (single instance)
- **Batch:** ~50,000 jobs/sec (100 items/batch)

### Latency

- **Local Redis:** <1ms per operation
- **Remote Redis:** 5-50ms depending on network

### Optimization

1. **Pipeline Commands** - Batch multiple operations
2. **Reduce Payload Size** - Smaller JSON = faster
3. **Connection Pooling** - Tune `PoolSize`
4. **Use Lua Scripts** - Atomic multi-command operations

## Migration from Memory

```go
// Development
manager.SetDriver(memory.NewDriver())

// Production
driver, _ := redis.NewDriver("myapp", &redis.Options{
    Addr: os.Getenv("REDIS_URL"),
})
manager.SetDriver(driver)
```

No code changes to job handlers - just swap drivers!
