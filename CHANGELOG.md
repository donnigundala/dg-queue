# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2025-12-27

### Added
- Initial stable release of the `dg-queue` plugin.
- **Core Queue System**: Job lifecycle management with retry logic and worker pools.
- **Multi-Driver Support**: Memory driver (development) and Redis driver (production).
- **Batch Processing**: Bulk job dispatching with chunking, progress callbacks, and rate limiting.
- **Container Integration**: Auto-registration with Injectable pattern and helper functions.
- **Observability**: OpenTelemetry metrics for job processing and queue performance.
- **Graceful Shutdown**: Proper cleanup and worker termination.

### Features
- Job status tracking with automatic UpdatedAt field
- Delayed jobs using Redis sorted sets
- Dead letter queue for failed jobs
- Shared Redis client support
- Configurable worker concurrency
- Error handling with continue-on-error option
- Production-ready with 42+ comprehensive tests

### Documentation
- Complete README with quick start and examples
- Container integration example
- Migration guide for scheduler extraction

### Performance
- Zero external dependencies (except go-redis for Redis driver)
- Efficient worker pool management
- Optimized batch processing

---

## Development History

The following versions represent the development journey leading to v1.0.0:

### 2025-12-03
- Scheduler functionality extracted to separate `dg-scheduler` package
- Fixed critical shutdown hang issue
- Improved driver cleanup mechanism

### 2025-11-24
- Initial implementation with memory and Redis drivers
- Batch processing capabilities
- Cron-based scheduling (later extracted)
