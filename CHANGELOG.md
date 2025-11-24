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

## [0.1.0] - 2025-11-24

### Added
- Initial implementation of dg-queue core and memory driver
- Basic job processing with retry logic
- Worker pools with configurable concurrency
- 5 memory driver tests
