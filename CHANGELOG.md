# dg-queue CHANGELOG

## [1.0.0] - 2025-11-24

### Added
- Core queue system with Job, Manager, and Worker interfaces
- Memory driver for development/testing
- Redis driver for production
  - Delayed jobs using Redis sorted sets
  - Dead letter queue for failed jobs
  - Shared client support
- Comprehensive testing (27 tests passing)
- Job lifecycle tracking (UpdatedAt field)
- Queue configuration with sensible defaults

### Fixed
- Critical job dispatcher bug (was misrouting jobs)
- UpdatedAt tracking for job state changes
- Lint errors in examples

### Breaking Changes
None (initial release)

## [0.1.0] - 2025-11-24

### Added
- Initial implementation of dg-queue core and memory driver
- Basic job processing with retry logic
- Worker pools with configurable concurrency
- 5 memory driver tests
