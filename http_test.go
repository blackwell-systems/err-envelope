package errenvelope

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteWithError(t *testing.T) {
	err := NotFound("user not found")
	err = err.WithTraceID("trace123")

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)

	Write(w, r, err)

	// Check status code
	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}

	// Check Content-Type
	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", contentType)
	}

	// Check X-Request-Id header
	traceID := w.Header().Get(HeaderTraceID)
	if traceID != "trace123" {
		t.Errorf("expected X-Request-Id trace123, got %s", traceID)
	}

	// Check response body
	var response Error
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if response.Code != CodeNotFound {
		t.Errorf("expected code %s, got %s", CodeNotFound, response.Code)
	}
	if response.Message != "user not found" {
		t.Errorf("expected message 'user not found', got %s", response.Message)
	}
	if response.TraceID != "trace123" {
		t.Errorf("expected trace ID trace123, got %s", response.TraceID)
	}
}

func TestWriteWithNil(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)

	Write(w, r, nil)

	// Should return 204 No Content
	if w.Code != http.StatusNoContent {
		t.Errorf("expected status %d, got %d", http.StatusNoContent, w.Code)
	}

	// Body should be empty
	if w.Body.Len() != 0 {
		t.Error("expected empty body for nil error")
	}
}

func TestWriteWithGenericError(t *testing.T) {
	genericErr := errors.New("something broke")

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)

	Write(w, r, genericErr)

	// Should map to internal server error
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}

	var response Error
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if response.Code != CodeInternal {
		t.Errorf("expected code %s, got %s", CodeInternal, response.Code)
	}
}

func TestWriteWithTraceFromRequest(t *testing.T) {
	// Error without trace ID
	err := NotFound("not found")

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	r.Header.Set(HeaderTraceID, "request-trace-123")

	Write(w, r, err)

	// Should extract trace ID from request
	traceID := w.Header().Get(HeaderTraceID)
	if traceID != "request-trace-123" {
		t.Errorf("expected X-Request-Id request-trace-123, got %s", traceID)
	}

	var response Error
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if response.TraceID != "request-trace-123" {
		t.Errorf("expected trace ID request-trace-123, got %s", response.TraceID)
	}
}

func TestWriteWithTraceFromContext(t *testing.T) {
	err := NotFound("not found")

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	ctx := WithTraceID(r.Context(), "context-trace-456")
	r = r.WithContext(ctx)

	Write(w, r, err)

	// Should extract trace ID from context
	traceID := w.Header().Get(HeaderTraceID)
	if traceID != "context-trace-456" {
		t.Errorf("expected X-Request-Id context-trace-456, got %s", traceID)
	}
}

func TestWriteWithoutStatus(t *testing.T) {
	// Error with status 0 should default to 500
	err := &Error{
		Code:    CodeInternal,
		Message: "error",
		Status:  0, // No status
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)

	Write(w, r, err)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected default status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestWriteValidationError(t *testing.T) {
	fields := FieldErrors{
		"email": "invalid format",
		"age":   "must be positive",
	}
	err := Validation(fields)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/test", nil)

	Write(w, r, err)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var response Error
	if jsonErr := json.Unmarshal(w.Body.Bytes(), &response); jsonErr != nil {
		t.Fatalf("failed to unmarshal response: %v", jsonErr)
	}

	// Check details are preserved
	details, ok := response.Details.(map[string]any)
	if !ok {
		t.Fatal("expected details to be map")
	}
	fieldsMap, ok := details["fields"].(map[string]any)
	if !ok {
		t.Fatal("expected fields in details")
	}
	if fieldsMap["email"] != "invalid format" {
		t.Error("expected email error in fields")
	}
}

func TestWriteWithDeadlineExceeded(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)

	Write(w, r, context.DeadlineExceeded)

	if w.Code != http.StatusGatewayTimeout {
		t.Errorf("expected status %d, got %d", http.StatusGatewayTimeout, w.Code)
	}

	var response Error
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if response.Code != CodeTimeout {
		t.Errorf("expected code %s, got %s", CodeTimeout, response.Code)
	}
	if !response.Retryable {
		t.Error("timeout should be retryable")
	}
}

func TestWriteRetryableFlag(t *testing.T) {
	err := RateLimited("too many requests")

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)

	Write(w, r, err)

	var response Error
	if unmarshalErr := json.Unmarshal(w.Body.Bytes(), &response); unmarshalErr != nil {
		t.Fatalf("failed to unmarshal response: %v", unmarshalErr)
	}

	if !response.Retryable {
		t.Error("rate limited should be retryable")
	}
}
