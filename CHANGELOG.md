# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.1.0] - TBD

### Added
- Formatted error constructors: `Newf()` and `Wrapf()` for `fmt.Printf`-style messages
- Formatted helper functions: `Internalf()`, `BadRequestf()`, `NotFoundf()`, `Unauthorizedf()`, `Forbiddenf()`, `Conflictf()`, `Timeoutf()`, `Unavailablef()`
- Custom JSON serialization: `retry_after` field now included in JSON responses when `RetryAfter` is set (e.g., `"retry_after": "30s"`)
- Comprehensive tests for formatted constructors and JSON serialization

### Changed
- **BREAKING (from v1.0.0):** `WithDetails()`, `WithTraceID()`, `WithRetryable()`, `WithStatus()`, and `WithRetryAfter()` now return copies instead of mutating the receiver
  - Prevents shared-state bugs when reusing error instances
  - Aligns with Go's `context.WithValue()` pattern
  - Safe for global error variables
  - Migration: No code changes needed, but behavior is now safer

## [1.0.0] - 2025-12-22

### Added
- Initial release
- Core error envelope structure with code/message/details/trace_id/retryable
- HTTP middleware for automatic trace ID injection
- Framework integrations: Chi, Echo, Gin
- Validation error helpers with field-level errors
- Constructor functions: `Internal()`, `BadRequest()`, `Validation()`, etc.
- Error code constants: `CodeInternal`, `CodeValidationFailed`, etc.
- `Retry-After` header support for rate limiting
- `slog.LogValuer` implementation for structured logging
- Comprehensive test suite with 99% coverage
- Examples and documentation

[Unreleased]: https://github.com/blackwell-systems/err-envelope/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/blackwell-systems/err-envelope/releases/tag/v1.0.0
