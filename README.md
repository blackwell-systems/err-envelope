# err-envelope

[![Blackwell Systems™](https://raw.githubusercontent.com/blackwell-systems/blackwell-docs-theme/main/badge-trademark.svg)](https://github.com/blackwell-systems)
[![Go Reference](https://pkg.go.dev/badge/github.com/blackwell-systems/err-envelope.svg)](https://pkg.go.dev/github.com/blackwell-systems/err-envelope)
[![CI](https://github.com/blackwell-systems/err-envelope/actions/workflows/ci.yml/badge.svg)](https://github.com/blackwell-systems/err-envelope/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/blackwell-systems/err-envelope)](https://goreportcard.com/report/github.com/blackwell-systems/err-envelope)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Sponsor](https://img.shields.io/badge/Sponsor-Buy%20Me%20a%20Coffee-yellow?logo=buy-me-a-coffee&logoColor=white)](https://buymeacoffee.com/blackwellsystems)

A tiny Go package for consistent HTTP error responses across services.

## Why

Without a standard, every endpoint returns errors differently:
- `{"error": "bad request"}`
- `{"message": "invalid email"}`
- `{"code": "E123", "details": {...}}`

This forces clients to handle each endpoint specially. `err-envelope` provides a single, predictable error shape.

## What You Get

```json
{
  "code": "VALIDATION_FAILED",
  "message": "Invalid input",
  "details": {
    "fields": {
      "email": "must be a valid email"
    }
  },
  "trace_id": "a1b2c3d4e5f6",
  "retryable": false
}
```

Every field has a purpose: stable codes for logic, messages for humans, details for context, trace IDs for debugging, and retry signals for resilience.

## Installation

```bash
go get github.com/blackwell-systems/err-envelope
```

## Quick Start

```go
package main

import (
    "net/http"
    errenvelope "github.com/blackwell-systems/err-envelope"
)

func main() {
    mux := http.NewServeMux()

    mux.HandleFunc("/user", func(w http.ResponseWriter, r *http.Request) {
        email := r.URL.Query().Get("email")
        if email == "" {
            // Validation error with field details
            err := errenvelope.Validation(errenvelope.FieldErrors{
                "email": "is required",
            })
            errenvelope.Write(w, r, err)
            return
        }
        // ... success path
    })

    // Wrap with trace middleware
    http.ListenAndServe(":8080", errenvelope.TraceMiddleware(mux))
}
```

### Response

```bash
$ curl http://localhost:8080/user
```

```json
{
  "code": "VALIDATION_FAILED",
  "message": "Invalid input",
  "details": {
    "fields": {
      "email": "is required"
    }
  },
  "trace_id": "f8e7d6c5b4a39281",
  "retryable": false
}
```

## API

### Common Constructors

```go
// Validation errors (400)
errenvelope.Validation(errenvelope.FieldErrors{
    "email": "invalid format",
    "age": "must be positive",
})

// Auth errors
errenvelope.Unauthorized("Missing token")    // 401
errenvelope.Forbidden("Insufficient permissions") // 403

// Resource errors
errenvelope.NotFound("User not found")          // 404
errenvelope.MethodNotAllowed("POST not allowed") // 405
errenvelope.RequestTimeout("Client timeout")     // 408
errenvelope.Conflict("Email already exists")     // 409

// Infrastructure errors
errenvelope.Timeout("Database query timed out")      // 504
errenvelope.Unavailable("Service temporarily down")  // 503
errenvelope.RateLimited("Too many requests")        // 429

// Downstream errors
errenvelope.Downstream("payments", err)              // 502
errenvelope.DownstreamTimeout("payments", err)       // 504
```

### Custom Errors

```go
// Low-level constructor
err := errenvelope.New(
    errenvelope.CodeInternal,
    http.StatusInternalServerError,
    "Database connection failed",
)

// Add details
err = err.WithDetails(map[string]any{
    "database": "postgres",
    "host": "db.example.com",
})

// Add trace ID
err = err.WithTraceID("abc123")

// Override retryable
err = err.WithRetryable(true)
```

### Writing Responses

```go
// Write error response
errenvelope.Write(w, r, err)

// Automatically handles:
// - Sets Content-Type: application/json
// - Sets X-Request-Id header (if trace ID present)
// - Sets correct HTTP status
// - Encodes error as JSON
```

### Mapping Arbitrary Errors

```go
// Convert any error to envelope
err := someLibrary.DoSomething()
errenvelope.Write(w, r, err)  // From() called automatically

// Handles:
// - context.DeadlineExceeded → Timeout
// - context.Canceled → Canceled (499)
// - net.Error with Timeout() → Timeout
// - *errenvelope.Error → passthrough
// - Unknown errors → Internal (500)
```

### Trace ID Middleware

```go
mux := http.NewServeMux()
// ... register handlers

// Wrap with trace middleware
handler := errenvelope.TraceMiddleware(mux)
http.ListenAndServe(":8080", handler)

// Generates trace ID if missing
// Propagates X-Request-Id header
// Adds to context for downstream access
```

## Error Codes

| Code | HTTP Status | Retryable | Use Case |
|------|-------------|-----------|----------|
| `INTERNAL` | 500 | No | Unexpected server errors |
| `BAD_REQUEST` | 400 | No | Malformed requests |
| `VALIDATION_FAILED` | 400 | No | Invalid input data |
| `UNAUTHORIZED` | 401 | No | Missing/invalid auth |
| `FORBIDDEN` | 403 | No | Insufficient permissions |
| `NOT_FOUND` | 404 | No | Resource doesn't exist |
| `METHOD_NOT_ALLOWED` | 405 | No | Invalid HTTP method |
| `REQUEST_TIMEOUT` | 408 | Yes | Client timeout |
| `CONFLICT` | 409 | No | State conflict (duplicate) |
| `RATE_LIMITED` | 429 | Yes | Too many requests |
| `CANCELED` | 499 | No | Client canceled request |
| `UNAVAILABLE` | 503 | Yes | Service temporarily down |
| `TIMEOUT` | 504 | Yes | Gateway timeout |
| `DOWNSTREAM_ERROR` | 502 | Yes | Upstream service failed |
| `DOWNSTREAM_TIMEOUT` | 504 | Yes | Upstream service timeout |

## Design Principles

**Minimal**: ~300 lines, stdlib only, single responsibility.

**Framework-Agnostic**: Works with `net/http` out of the box. Easy adapters for chi/gin/echo.

**Predictable**: Error codes are stable (never change). Messages may evolve for clarity. Sensible defaults for status codes and retryability.

**Observable**: Trace IDs for request correlation. Structured details for logging. Cause chains preserved via `errors.Unwrap`.

## Compatibility

If you already use RFC 7807 Problem Details, this can coexist—map between formats at the edge.

## JSON Schema

A [JSON Schema](schema.json) is included for client tooling and contract testing:

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "errenvelope.Error",
  "required": ["code", "message", "retryable"],
  "properties": {
    "code": { "type": "string" },
    "message": { "type": "string" },
    "retryable": { "type": "boolean" }
  }
}
```

Use this to validate responses, generate TypeScript types, or document your API.

## Examples

See [examples/nethttp](examples/nethttp) for a complete demo server.

Run it:
```bash
cd examples/nethttp
go run main.go
```

Test endpoints:
```bash
curl http://localhost:8080/validate
curl http://localhost:8080/user/
curl http://localhost:8080/downstream
curl http://localhost:8080/protected
curl http://localhost:8080/timeout
```

## Integration Patterns

**`net/http`** (default): Use `errenvelope.Write(w, r, err)` and `errenvelope.TraceMiddleware(mux)`.

**Chi/Gin/Echo**: Wrap errors in your middleware layer before writing responses.

**OpenAPI/TypeScript**: Use the included [JSON Schema](schema.json) to generate client types.

## Versioning

Follows semantic versioning. No breaking changes to envelope fields (`code`, `message`, `details`, `trace_id`, `retryable`) in minor releases.

## Used By

- [Pipeboard](https://github.com/blackwell-systems/pipeboard) - Clipboard sync service

## License

MIT

## Contributing

This is a reference implementation. Fork and adapt to your needs.

If you find a bug or have a suggestion, open an issue.
