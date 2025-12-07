package errenvelope

import (
	"errors"
	"net/http"
	"testing"
)

func TestNew(t *testing.T) {
	err := New(CodeInternal, http.StatusInternalServerError, "something went wrong")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Code != CodeInternal {
		t.Errorf("expected code %s, got %s", CodeInternal, err.Code)
	}
	if err.Status != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, err.Status)
	}
	if err.Message != "something went wrong" {
		t.Errorf("expected message 'something went wrong', got %s", err.Message)
	}
	if err.Retryable {
		t.Error("expected retryable to be false by default")
	}
}

func TestWrap(t *testing.T) {
	cause := errors.New("root cause")
	err := Wrap(CodeInternal, http.StatusInternalServerError, "wrapper message", cause)

	if err.Cause != cause {
		t.Error("expected cause to be set")
	}
	if err.Message != "wrapper message" {
		t.Errorf("expected message 'wrapper message', got %s", err.Message)
	}

	// Test Unwrap
	unwrapped := errors.Unwrap(err)
	if unwrapped != cause {
		t.Error("Unwrap should return the cause")
	}
}

func TestError(t *testing.T) {
	err := New(CodeNotFound, http.StatusNotFound, "user not found")
	str := err.Error()

	expected := "NOT_FOUND: user not found"
	if str != expected {
		t.Errorf("expected error string %q, got %q", expected, str)
	}
}

func TestWithDetails(t *testing.T) {
	err := New(CodeValidationFailed, http.StatusBadRequest, "validation failed")
	details := map[string]any{"field": "email"}
	err = err.WithDetails(details)

	if err.Details == nil {
		t.Fatal("expected details to be set")
	}
	detailsMap, ok := err.Details.(map[string]any)
	if !ok {
		t.Fatal("expected details to be map[string]any")
	}
	if detailsMap["field"] != "email" {
		t.Error("expected field 'email' in details")
	}
}

func TestErrorWithTraceID(t *testing.T) {
	err := New(CodeInternal, http.StatusInternalServerError, "error")
	err = err.WithTraceID("trace123")

	if err.TraceID != "trace123" {
		t.Errorf("expected trace ID 'trace123', got %s", err.TraceID)
	}
}

func TestWithRetryable(t *testing.T) {
	err := New(CodeInternal, http.StatusInternalServerError, "error")

	// Default is false
	if err.Retryable {
		t.Error("expected retryable to be false by default")
	}

	// Set to true
	err = err.WithRetryable(true)
	if !err.Retryable {
		t.Error("expected retryable to be true")
	}

	// Set back to false
	err = err.WithRetryable(false)
	if err.Retryable {
		t.Error("expected retryable to be false")
	}
}

func TestWithStatus(t *testing.T) {
	err := New(CodeInternal, http.StatusInternalServerError, "error")
	err = err.WithStatus(http.StatusBadGateway)

	if err.Status != http.StatusBadGateway {
		t.Errorf("expected status %d, got %d", http.StatusBadGateway, err.Status)
	}
}

func TestIs(t *testing.T) {
	err := New(CodeNotFound, http.StatusNotFound, "not found")

	if !Is(err, CodeNotFound) {
		t.Error("Is() should return true for matching code")
	}
	if Is(err, CodeInternal) {
		t.Error("Is() should return false for non-matching code")
	}

	// Test with wrapped error
	wrapped := Wrap(CodeInternal, http.StatusInternalServerError, "wrapper", err)
	if !Is(wrapped, CodeInternal) {
		t.Error("Is() should match wrapped error code")
	}

	// Test with non-Error type
	regularErr := errors.New("regular error")
	if Is(regularErr, CodeInternal) {
		t.Error("Is() should return false for non-Error type")
	}

	// Test with nil
	if Is(nil, CodeInternal) {
		t.Error("Is() should return false for nil")
	}
}

func TestChaining(t *testing.T) {
	err := New(CodeValidationFailed, http.StatusBadRequest, "validation failed").
		WithDetails(map[string]any{"field": "email"}).
		WithTraceID("trace123").
		WithRetryable(false)

	if err.Code != CodeValidationFailed {
		t.Error("code should be preserved in chain")
	}
	if err.Details == nil {
		t.Error("details should be set")
	}
	if err.TraceID != "trace123" {
		t.Error("trace ID should be set")
	}
	if err.Retryable {
		t.Error("retryable should be false")
	}
}
