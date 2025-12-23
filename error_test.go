package errenvelope

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"testing"
	"time"
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
func TestWithRetryAfter(t *testing.T) {
	err := RateLimited("too many requests").WithRetryAfter(30 * time.Second)

	if err.RetryAfter != 30*time.Second {
		t.Errorf("expected retry after 30s, got %v", err.RetryAfter)
	}
	if err.Code != CodeRateLimited {
		t.Errorf("expected code %s, got %s", CodeRateLimited, err.Code)
	}
	if !err.Retryable {
		t.Error("rate limited should be retryable")
	}
}

func TestImmutability(t *testing.T) {
	original := New(CodeNotFound, http.StatusNotFound, "not found")
	
	modified := original.
		WithDetails(map[string]string{"id": "123"}).
		WithTraceID("trace-456").
		WithRetryable(true).
		WithStatus(http.StatusGone)
	
	if original.Details != nil {
		t.Error("WithDetails should not mutate original error")
	}
	if original.TraceID != "" {
		t.Error("WithTraceID should not mutate original error")
	}
	if original.Retryable {
		t.Error("WithRetryable should not mutate original error")
	}
	if original.Status != http.StatusNotFound {
		t.Error("WithStatus should not mutate original error")
	}
	
	if modified.Details == nil {
		t.Error("modified error should have details")
	}
	if modified.TraceID != "trace-456" {
		t.Errorf("modified error should have trace ID, got %s", modified.TraceID)
	}
	if !modified.Retryable {
		t.Error("modified error should be retryable")
	}
	if modified.Status != http.StatusGone {
		t.Errorf("modified error should have updated status, got %d", modified.Status)
	}
}

func TestLogValue(t *testing.T) {
	cause := errors.New("database timeout")
	err := Internal("processing failed").
		WithDetails(map[string]string{"request_id": "123"}).
		WithTraceID("trace-abc").
		WithRetryAfter(10 * time.Second)
	err.Cause = cause

	// Test that LogValue returns a valid slog.Value
	logVal := err.LogValue()
	if logVal.Kind() != slog.KindGroup {
		t.Error("LogValue should return a group")
	}

	// Test with slog to ensure it integrates correctly
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))
	logger.Info("error occurred", "error", err)

	// The buffer should contain JSON output with error fields
	output := buf.String()
	if !bytes.Contains([]byte(output), []byte("INTERNAL")) {
		t.Error("expected code in log output")
	}
	if !bytes.Contains([]byte(output), []byte("processing failed")) {
		t.Error("expected message in log output")
	}
	if !bytes.Contains([]byte(output), []byte("trace-abc")) {
		t.Error("expected trace_id in log output")
	}
}

func TestLogValueNil(t *testing.T) {
	var err *Error
	logVal := err.LogValue()
	if logVal.Kind() != slog.KindGroup {
		t.Error("LogValue on nil should return empty group")
	}
}

func TestNewf(t *testing.T) {
	userID := "12345"
	err := Newf(CodeNotFound, http.StatusNotFound, "user %s not found", userID)
	
	expected := "user 12345 not found"
	if err.Message != expected {
		t.Errorf("expected message %q, got %q", expected, err.Message)
	}
	if err.Code != CodeNotFound {
		t.Errorf("expected code %s, got %s", CodeNotFound, err.Code)
	}
	if err.Status != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, err.Status)
	}
}

func TestWrapf(t *testing.T) {
	cause := errors.New("connection refused")
	host := "db.example.com"
	err := Wrapf(CodeInternal, http.StatusInternalServerError, "failed to connect to %s", cause, host)
	
	expected := "failed to connect to db.example.com"
	if err.Message != expected {
		t.Errorf("expected message %q, got %q", expected, err.Message)
	}
	if err.Cause != cause {
		t.Error("expected cause to be set")
	}
	
	unwrapped := errors.Unwrap(err)
	if unwrapped != cause {
		t.Error("Unwrap should return the cause")
	}
}

func TestMarshalJSON(t *testing.T) {
	tests := []struct {
		name          string
		err           *Error
		wantRetryStr  string
		shouldContain bool
	}{
		{
			name:          "with retry_after",
			err:           RateLimited("too many requests").WithRetryAfter(30 * time.Second),
			wantRetryStr:  "30s",
			shouldContain: true,
		},
		{
			name:          "with longer retry_after",
			err:           Unavailable("maintenance").WithRetryAfter(5 * time.Minute),
			wantRetryStr:  "5m0s",
			shouldContain: true,
		},
		{
			name:          "without retry_after",
			err:           NotFound("user not found"),
			wantRetryStr:  "",
			shouldContain: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.err)
			if err != nil {
				t.Fatalf("failed to marshal: %v", err)
			}

			var result map[string]any
			if err := json.Unmarshal(data, &result); err != nil {
				t.Fatalf("failed to unmarshal: %v", err)
			}

			retryAfter, hasRetryAfter := result["retry_after"]
			if tt.shouldContain {
				if !hasRetryAfter {
					t.Error("expected retry_after field in JSON")
				}
				if retryAfter != tt.wantRetryStr {
					t.Errorf("expected retry_after %q, got %q", tt.wantRetryStr, retryAfter)
				}
			} else {
				if hasRetryAfter {
					t.Error("expected no retry_after field in JSON")
				}
			}

			if result["code"] != string(tt.err.Code) {
				t.Errorf("expected code %s, got %v", tt.err.Code, result["code"])
			}
			if result["message"] != tt.err.Message {
				t.Errorf("expected message %s, got %v", tt.err.Message, result["message"])
			}
			if result["retryable"] != tt.err.Retryable {
				t.Errorf("expected retryable %v, got %v", tt.err.Retryable, result["retryable"])
			}
		})
	}
}

func TestFormattedHelpers(t *testing.T) {
	tests := []struct {
		name     string
		err      *Error
		wantMsg  string
		wantCode Code
	}{
		{
			name:     "Internalf",
			err:      Internalf("database %s failed", "postgres"),
			wantMsg:  "database postgres failed",
			wantCode: CodeInternal,
		},
		{
			name:     "BadRequestf",
			err:      BadRequestf("invalid email: %s", "not-an-email"),
			wantMsg:  "invalid email: not-an-email",
			wantCode: CodeBadRequest,
		},
		{
			name:     "NotFoundf",
			err:      NotFoundf("user %d not found", 123),
			wantMsg:  "user 123 not found",
			wantCode: CodeNotFound,
		},
		{
			name:     "Unauthorizedf",
			err:      Unauthorizedf("missing header: %s", "Authorization"),
			wantMsg:  "missing header: Authorization",
			wantCode: CodeUnauthorized,
		},
		{
			name:     "Forbiddenf",
			err:      Forbiddenf("insufficient permissions for %s", "admin"),
			wantMsg:  "insufficient permissions for admin",
			wantCode: CodeForbidden,
		},
		{
			name:     "Conflictf",
			err:      Conflictf("email %s already exists", "test@example.com"),
			wantMsg:  "email test@example.com already exists",
			wantCode: CodeConflict,
		},
		{
			name:     "Timeoutf",
			err:      Timeoutf("query exceeded %dms", 5000),
			wantMsg:  "query exceeded 5000ms",
			wantCode: CodeTimeout,
		},
		{
			name:     "Unavailablef",
			err:      Unavailablef("service %s is down", "payments"),
			wantMsg:  "service payments is down",
			wantCode: CodeUnavailable,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Message != tt.wantMsg {
				t.Errorf("expected message %q, got %q", tt.wantMsg, tt.err.Message)
			}
			if tt.err.Code != tt.wantCode {
				t.Errorf("expected code %s, got %s", tt.wantCode, tt.err.Code)
			}
		})
	}
}
