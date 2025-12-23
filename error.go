// Package errenvelope provides a tiny, framework-agnostic
// server-side HTTP error envelope for Go services.
// It standardizes code/message/details/trace_id/retryable
// and includes helpers for validation, auth, timeouts,
// and downstream errors.
package errenvelope

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

// Error is a structured error envelope for HTTP APIs.
type Error struct {
	Code      Code   `json:"code"`
	Message   string `json:"message"`
	Details   any    `json:"details,omitempty"`
	TraceID   string `json:"trace_id,omitempty"`
	Retryable bool   `json:"retryable"`

	// Not serialized:
	Status     int           `json:"-"`
	Cause      error         `json:"-"`
	RetryAfter time.Duration `json:"-"` // Duration to wait before retrying
}

func (e *Error) Error() string {
	if e == nil {
		return "<nil>"
	}
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *Error) Unwrap() error { return e.Cause }

// MarshalJSON implements custom JSON serialization to include retry_after as a human-readable string.
// When RetryAfter is set, it appears in the JSON response as "retry_after": "30s" or "5m0s".
func (e *Error) MarshalJSON() ([]byte, error) {
	type Alias Error
	aux := &struct {
		*Alias
		RetryAfterStr string `json:"retry_after,omitempty"`
	}{
		Alias: (*Alias)(e),
	}
	if e.RetryAfter > 0 {
		aux.RetryAfterStr = e.RetryAfter.String()
	}
	return json.Marshal(aux)
}

// New creates a new Error with the given code, HTTP status, and message.
// If status is 0, defaults to 500. If message is empty, uses a default.
func New(code Code, status int, msg string) *Error {
	if status == 0 {
		status = http.StatusInternalServerError
	}
	if msg == "" {
		msg = defaultMessage(code)
	}
	return &Error{
		Code:      code,
		Message:   msg,
		Status:    status,
		Retryable: isRetryableDefault(code),
	}
}

// Wrap creates a new Error that wraps an underlying cause.
func Wrap(code Code, status int, msg string, cause error) *Error {
	e := New(code, status, msg)
	e.Cause = cause
	return e
}

// Newf creates a new Error with a formatted message.
// Mirrors fmt.Errorf ergonomics for creating errors with interpolated values.
func Newf(code Code, status int, format string, args ...any) *Error {
	return New(code, status, fmt.Sprintf(format, args...))
}

// Wrapf creates a new Error that wraps an underlying cause with a formatted message.
// Mirrors fmt.Errorf ergonomics while preserving error chain.
func Wrapf(code Code, status int, format string, cause error, args ...any) *Error {
	return Wrap(code, status, fmt.Sprintf(format, args...), cause)
}

// WithDetails adds structured details to the error.
// Returns a copy to avoid mutating shared error instances.
func (e *Error) WithDetails(details any) *Error {
	clone := *e
	clone.Details = details
	return &clone
}

// WithTraceID adds a trace ID for distributed tracing.
// Returns a copy to avoid mutating shared error instances.
func (e *Error) WithTraceID(id string) *Error {
	clone := *e
	clone.TraceID = id
	return &clone
}

// WithRetryable sets whether the error is retryable.
// Returns a copy to avoid mutating shared error instances.
func (e *Error) WithRetryable(v bool) *Error {
	clone := *e
	clone.Retryable = v
	return &clone
}

// WithStatus overrides the HTTP status code.
// Returns a copy to avoid mutating shared error instances.
func (e *Error) WithStatus(status int) *Error {
	clone := *e
	if status != 0 {
		clone.Status = status
	}
	return &clone
}

// WithRetryAfter sets the retry-after duration for rate-limited responses.
// The duration will be sent as a Retry-After header (in seconds).
// Returns a copy to avoid mutating shared error instances.
func (e *Error) WithRetryAfter(d time.Duration) *Error {
	clone := *e
	clone.RetryAfter = d
	return &clone
}

// LogValue implements slog.LogValuer for structured logging.
func (e *Error) LogValue() slog.Value {
	if e == nil {
		return slog.GroupValue()
	}
	attrs := []slog.Attr{
		slog.String("code", string(e.Code)),
		slog.String("message", e.Message),
		slog.Int("status", e.Status),
		slog.Bool("retryable", e.Retryable),
	}
	if e.TraceID != "" {
		attrs = append(attrs, slog.String("trace_id", e.TraceID))
	}
	if e.Details != nil {
		attrs = append(attrs, slog.Any("details", e.Details))
	}
	if e.RetryAfter > 0 {
		attrs = append(attrs, slog.Duration("retry_after", e.RetryAfter))
	}
	if e.Cause != nil {
		attrs = append(attrs, slog.String("cause", e.Cause.Error()))
	}
	return slog.GroupValue(attrs...)
}

// Is checks if an error has the given code.
func Is(err error, code Code) bool {
	var e *Error
	if errors.As(err, &e) {
		return e.Code == code
	}
	return false
}

func defaultMessage(code Code) string {
	switch code {
	case CodeBadRequest:
		return "Bad request"
	case CodeValidationFailed:
		return "Invalid input"
	case CodeUnauthorized:
		return "Unauthorized"
	case CodeForbidden:
		return "Forbidden"
	case CodeNotFound:
		return "Not found"
	case CodeGone:
		return "Resource no longer exists"
	case CodeConflict:
		return "Conflict"
	case CodePayloadTooLarge:
		return "Payload too large"
	case CodeUnprocessableEntity:
		return "Unprocessable entity"
	case CodeRateLimited:
		return "Rate limited"
	case CodeTimeout, CodeDownstreamTimeout:
		return "Request timed out"
	case CodeUnavailable:
		return "Service unavailable"
	case CodeCanceled:
		return "Request canceled"
	case CodeDownstream:
		return "Downstream service error"
	default:
		return "Internal error"
	}
}

func isRetryableDefault(code Code) bool {
	switch code {
	case CodeTimeout, CodeDownstreamTimeout, CodeUnavailable, CodeRateLimited:
		return true
	default:
		return false
	}
}
