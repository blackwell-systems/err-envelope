# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed
- **BREAKING:** `WithDetails()`, `WithTraceID()`, `WithRetryable()`, `WithStatus()`, and `WithRetryAfter()` now return copies instead of mutating the receiver
  - Prevents shared-state bugs when reusing error instances
  - Aligns with Go's `context.WithValue()` pattern
  - Safe for global error variables
  - Migration: No code changes needed, but behavior is now safer

### Added
- Test for immutability guarantee (`TestImmutability`)

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
