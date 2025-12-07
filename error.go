// Package errenvelope provides a tiny, framework-agnostic
// server-side HTTP error envelope for Go services.
// It standardizes code/message/details/trace_id/retryable
// and includes helpers for validation, auth, timeouts,
// and downstream errors.
package errenvelope

import (
	"errors"
	"fmt"
	"net/http"
)

// Error is a structured error envelope for HTTP APIs.
type Error struct {
	Code      Code   `json:"code"`
	Message   string `json:"message"`
	Details   any    `json:"details,omitempty"`
	TraceID   string `json:"trace_id,omitempty"`
	Retryable bool   `json:"retryable"`

	// Not serialized:
	Status int   `json:"-"`
	Cause  error `json:"-"`
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

// WithDetails adds structured details to the error.
func (e *Error) WithDetails(details any) *Error {
	e.Details = details
	return e
}

// WithTraceID adds a trace ID for distributed tracing.
func (e *Error) WithTraceID(id string) *Error {
	e.TraceID = id
	return e
}

// WithRetryable sets whether the error is retryable.
func (e *Error) WithRetryable(v bool) *Error {
	e.Retryable = v
	return e
}

// WithStatus overrides the HTTP status code.
func (e *Error) WithStatus(status int) *Error {
	if status != 0 {
		e.Status = status
	}
	return e
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
