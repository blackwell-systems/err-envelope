package errenvelope

import (
	"context"
	"errors"
	"net"
	"net/http"
)

// FieldErrors is a simple, library-agnostic validation shape.
type FieldErrors map[string]string

// ValidationDetails holds field-level validation errors.
type ValidationDetails struct {
	Fields FieldErrors `json:"fields"`
}

// Validation creates a validation error with field-level details.
func Validation(fields FieldErrors) *Error {
	return New(CodeValidationFailed, http.StatusBadRequest, "").
		WithDetails(ValidationDetails{Fields: fields}).
		WithRetryable(false)
}

// Unauthorized creates an unauthorized error (401).
func Unauthorized(msg string) *Error {
	return New(CodeUnauthorized, http.StatusUnauthorized, msg).
		WithRetryable(false)
}

// Forbidden creates a forbidden error (403).
func Forbidden(msg string) *Error {
	return New(CodeForbidden, http.StatusForbidden, msg).
		WithRetryable(false)
}

// NotFound creates a not found error (404).
func NotFound(msg string) *Error {
	return New(CodeNotFound, http.StatusNotFound, msg).
		WithRetryable(false)
}

// Conflict creates a conflict error (409).
func Conflict(msg string) *Error {
	return New(CodeConflict, http.StatusConflict, msg).
		WithRetryable(false)
}

// RateLimited creates a rate limit error (429).
func RateLimited(msg string) *Error {
	return New(CodeRateLimited, http.StatusTooManyRequests, msg).
		WithRetryable(true)
}

// Timeout creates a timeout error (504).
func Timeout(msg string) *Error {
	return New(CodeTimeout, http.StatusGatewayTimeout, msg).
		WithRetryable(true)
}

// Unavailable creates an unavailable error (503).
func Unavailable(msg string) *Error {
	return New(CodeUnavailable, http.StatusServiceUnavailable, msg).
		WithRetryable(true)
}

// Downstream creates an error for downstream service failures (502).
func Downstream(service string, cause error) *Error {
	d := map[string]any{}
	if service != "" {
		d["service"] = service
	}
	return Wrap(CodeDownstream, http.StatusBadGateway, "", cause).
		WithDetails(d).
		WithRetryable(true)
}

// DownstreamTimeout creates a timeout error for downstream services (504).
func DownstreamTimeout(service string, cause error) *Error {
	d := map[string]any{}
	if service != "" {
		d["service"] = service
	}
	return Wrap(CodeDownstreamTimeout, http.StatusGatewayTimeout, "", cause).
		WithDetails(d).
		WithRetryable(true)
}

// From maps arbitrary errors into an *Error.
// Handles context errors, network timeouts, and wraps unknown errors.
func From(err error) *Error {
	if err == nil {
		return nil
	}

	var e *Error
	if errors.As(err, &e) {
		// Ensure status is sane
		if e.Status == 0 {
			e.Status = http.StatusInternalServerError
		}
		if e.Message == "" {
			e.Message = defaultMessage(e.Code)
		}
		return e
	}

	// Context-driven
	if errors.Is(err, context.DeadlineExceeded) {
		return Timeout("")
	}
	if errors.Is(err, context.Canceled) {
		return New(CodeCanceled, 499, "").WithRetryable(false) // 499 is common convention
	}

	// net.Error timeouts
	var ne net.Error
	if errors.As(err, &ne) && ne.Timeout() {
		return Timeout("")
	}

	// Default
	return Wrap(CodeInternal, http.StatusInternalServerError, "", err).
		WithRetryable(false)
}
