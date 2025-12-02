# dg-queue CHANGELOG

## [1.0.0] - 2025-11-24

### Added
- **Core Queue System**
  - Job lifecycle management with retry logic
  - Worker pools with configurable concurrency
  - Graceful shutdown support
  - Job status tracking with UpdatedAt field
  
- **Drivers**
  - Memory driver for development/testing (5 tests)
  - Redis driver for production (6 tests)
    - Delayed jobs using Redis sorted sets
    - Dead letter queue for failed jobs
    - Shared client support via `NewDriverWithClient`
  
- **Scheduler** (7 tests)
  - Cron-based job scheduling using robfig/cron
  - Schedule management (add, remove, count)
  - Convenience `ScheduleJob` method
  
- **Batch Processing** (8 tests)
  - Bulk job dispatching with chunking
  - Progress callbacks
  - Error handling with continue-on-error option
  - Map function for item transformation
  - Rate limiting support

### Fixed
- Critical job dispatcher bug (was misrouting jobs to wrong workers)
- Job UpdatedAt tracking for all state changes
- Lint errors in example files

### Technical Details
- 42 comprehensive tests passing
- Zero external dependencies (except go-redis for Redis driver)
- Production-ready with full test coverage

## [1.3.0] - 2025-12-03

### Changed
- **BREAKING**: Scheduler functionality extracted to separate `dg-scheduler` package
- `Manager.Schedule()` method now deprecated (returns error with migration message)
- Improved shutdown mechanism with proper driver cleanup
- Memory driver now returns `ErrQueueEmpty` instead of `ErrJobNotFound` for consistency
- Updated README with scheduler migration guide

### Added
- Driver connection cleanup in `Manager.Stop()` - fixes shutdown hang
- `stopChan` recreation in `Manager.Start()` - allows safe restart
- Nil check in Redis driver `Close()` method
- Deprecation notice for `Schedule()` method
- Migration guide in README

### Fixed
- **Critical**: Application no longer hangs on shutdown (Ctrl+C)
- Redis connections now properly closed on shutdown
- Queue can be safely stopped and restarted
- Error consistency between memory and Redis drivers

### Removed
- Scheduler logic from Manager (moved to dg-scheduler)
- `Schedule()` method from Queue interface
- `ScheduleHandler` type from queue.go
- Internal `schedule` struct and scheduler goroutines

### Migration Guide
Users of `Manager.Schedule()` should migrate to the new `dg-scheduler` package:

```go
// Before (v1.2.x - now deprecated)
q.Schedule("*/5 * * * *", "cleanup", handler)

// After (v1.3.0+)
import "github.com/donnigundala/dg-scheduler"
s := scheduler.New(q)
s.Schedule("*/5 * * * *", "cleanup", handler)
```

See [dg-scheduler](https://github.com/donnigundala/dg-scheduler) for full documentation.

### Dependencies
- Removed: `github.com/robfig/cron/v3` (moved to dg-scheduler)

## [0.1.0] - 2025-11-24

### Added
- Initial implementation of dg-queue core and memory driver
- Basic job processing with retry logic
- Worker pools with configurable concurrency
- 5 memory driver tests
