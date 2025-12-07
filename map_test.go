package errenvelope

import (
	"context"
	"errors"
	"net"
	"net/http"
	"testing"
	"time"
)

func TestBadRequest(t *testing.T) {
	err := BadRequest("malformed JSON")

	if err.Code != CodeBadRequest {
		t.Errorf("expected code %s, got %s", CodeBadRequest, err.Code)
	}
	if err.Status != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, err.Status)
	}
	if err.Message != "malformed JSON" {
		t.Errorf("expected message 'malformed JSON', got %s", err.Message)
	}
	if err.Retryable {
		t.Error("bad request should not be retryable")
	}
}

func TestValidation(t *testing.T) {
	fields := FieldErrors{
		"email": "invalid format",
		"age":   "must be positive",
	}
	err := Validation(fields)

	if err.Code != CodeValidationFailed {
		t.Errorf("expected code %s, got %s", CodeValidationFailed, err.Code)
	}
	if err.Status != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, err.Status)
	}
	if err.Retryable {
		t.Error("validation errors should not be retryable")
	}

	details, ok := err.Details.(ValidationDetails)
	if !ok {
		t.Fatal("expected ValidationDetails")
	}
	if details.Fields["email"] != "invalid format" {
		t.Error("expected email field error")
	}
	if details.Fields["age"] != "must be positive" {
		t.Error("expected age field error")
	}
}

func TestUnauthorized(t *testing.T) {
	err := Unauthorized("missing token")

	if err.Code != CodeUnauthorized {
		t.Errorf("expected code %s, got %s", CodeUnauthorized, err.Code)
	}
	if err.Status != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, err.Status)
	}
	if err.Message != "missing token" {
		t.Errorf("expected message 'missing token', got %s", err.Message)
	}
	if err.Retryable {
		t.Error("unauthorized should not be retryable")
	}
}

func TestForbidden(t *testing.T) {
	err := Forbidden("insufficient permissions")

	if err.Code != CodeForbidden {
		t.Errorf("expected code %s, got %s", CodeForbidden, err.Code)
	}
	if err.Status != http.StatusForbidden {
		t.Errorf("expected status %d, got %d", http.StatusForbidden, err.Status)
	}
	if err.Retryable {
		t.Error("forbidden should not be retryable")
	}
}

func TestNotFound(t *testing.T) {
	err := NotFound("user not found")

	if err.Code != CodeNotFound {
		t.Errorf("expected code %s, got %s", CodeNotFound, err.Code)
	}
	if err.Status != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, err.Status)
	}
	if err.Retryable {
		t.Error("not found should not be retryable")
	}
}

func TestConflict(t *testing.T) {
	err := Conflict("email already exists")

	if err.Code != CodeConflict {
		t.Errorf("expected code %s, got %s", CodeConflict, err.Code)
	}
	if err.Status != http.StatusConflict {
		t.Errorf("expected status %d, got %d", http.StatusConflict, err.Status)
	}
	if err.Retryable {
		t.Error("conflict should not be retryable")
	}
}

func TestMethodNotAllowed(t *testing.T) {
	err := MethodNotAllowed("POST not allowed")

	if err.Code != CodeMethodNotAllowed {
		t.Errorf("expected code %s, got %s", CodeMethodNotAllowed, err.Code)
	}
	if err.Status != http.StatusMethodNotAllowed {
		t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, err.Status)
	}
	if err.Retryable {
		t.Error("method not allowed should not be retryable")
	}
}

func TestRequestTimeout(t *testing.T) {
	err := RequestTimeout("client timeout")

	if err.Code != CodeRequestTimeout {
		t.Errorf("expected code %s, got %s", CodeRequestTimeout, err.Code)
	}
	if err.Status != http.StatusRequestTimeout {
		t.Errorf("expected status %d, got %d", http.StatusRequestTimeout, err.Status)
	}
	if !err.Retryable {
		t.Error("request timeout should be retryable")
	}
}

func TestRateLimited(t *testing.T) {
	err := RateLimited("too many requests")

	if err.Code != CodeRateLimited {
		t.Errorf("expected code %s, got %s", CodeRateLimited, err.Code)
	}
	if err.Status != http.StatusTooManyRequests {
		t.Errorf("expected status %d, got %d", http.StatusTooManyRequests, err.Status)
	}
	if !err.Retryable {
		t.Error("rate limited should be retryable")
	}
}

func TestTimeout(t *testing.T) {
	err := Timeout("database query timed out")

	if err.Code != CodeTimeout {
		t.Errorf("expected code %s, got %s", CodeTimeout, err.Code)
	}
	if err.Status != http.StatusGatewayTimeout {
		t.Errorf("expected status %d, got %d", http.StatusGatewayTimeout, err.Status)
	}
	if !err.Retryable {
		t.Error("timeout should be retryable")
	}
}

func TestUnavailable(t *testing.T) {
	err := Unavailable("service temporarily down")

	if err.Code != CodeUnavailable {
		t.Errorf("expected code %s, got %s", CodeUnavailable, err.Code)
	}
	if err.Status != http.StatusServiceUnavailable {
		t.Errorf("expected status %d, got %d", http.StatusServiceUnavailable, err.Status)
	}
	if !err.Retryable {
		t.Error("unavailable should be retryable")
	}
}

func TestDownstream(t *testing.T) {
	cause := errors.New("connection refused")
	err := Downstream("payments", cause)

	if err.Code != CodeDownstream {
		t.Errorf("expected code %s, got %s", CodeDownstream, err.Code)
	}
	if err.Status != http.StatusBadGateway {
		t.Errorf("expected status %d, got %d", http.StatusBadGateway, err.Status)
	}
	if !err.Retryable {
		t.Error("downstream should be retryable")
	}
	if err.Cause != cause {
		t.Error("cause should be preserved")
	}

	details, ok := err.Details.(map[string]any)
	if !ok {
		t.Fatal("expected map details")
	}
	if details["service"] != "payments" {
		t.Error("expected service 'payments' in details")
	}
}

func TestDownstreamTimeout(t *testing.T) {
	cause := errors.New("deadline exceeded")
	err := DownstreamTimeout("payments", cause)

	if err.Code != CodeDownstreamTimeout {
		t.Errorf("expected code %s, got %s", CodeDownstreamTimeout, err.Code)
	}
	if err.Status != http.StatusGatewayTimeout {
		t.Errorf("expected status %d, got %d", http.StatusGatewayTimeout, err.Status)
	}
	if !err.Retryable {
		t.Error("downstream timeout should be retryable")
	}

	details, ok := err.Details.(map[string]any)
	if !ok {
		t.Fatal("expected map details")
	}
	if details["service"] != "payments" {
		t.Error("expected service 'payments' in details")
	}
}

func TestFromNil(t *testing.T) {
	err := From(nil)
	if err != nil {
		t.Error("From(nil) should return nil")
	}
}

func TestFromError(t *testing.T) {
	original := New(CodeNotFound, http.StatusNotFound, "not found")
	err := From(original)

	if err != original {
		t.Error("From(*Error) should return the same error")
	}
}

func TestFromDeadlineExceeded(t *testing.T) {
	err := From(context.DeadlineExceeded)

	if err.Code != CodeTimeout {
		t.Errorf("expected code %s, got %s", CodeTimeout, err.Code)
	}
	if err.Status != http.StatusGatewayTimeout {
		t.Errorf("expected status %d, got %d", http.StatusGatewayTimeout, err.Status)
	}
	if !err.Retryable {
		t.Error("deadline exceeded should be retryable")
	}
}

func TestFromCanceled(t *testing.T) {
	err := From(context.Canceled)

	if err.Code != CodeCanceled {
		t.Errorf("expected code %s, got %s", CodeCanceled, err.Code)
	}
	if err.Status != 499 {
		t.Errorf("expected status 499, got %d", err.Status)
	}
	if err.Retryable {
		t.Error("canceled should not be retryable")
	}
}

type timeoutError struct{}

func (timeoutError) Error() string   { return "timeout" }
func (timeoutError) Timeout() bool   { return true }
func (timeoutError) Temporary() bool { return true }

func TestFromNetError(t *testing.T) {
	var netErr net.Error = timeoutError{}
	err := From(netErr)

	if err.Code != CodeTimeout {
		t.Errorf("expected code %s, got %s", CodeTimeout, err.Code)
	}
	if err.Status != http.StatusGatewayTimeout {
		t.Errorf("expected status %d, got %d", http.StatusGatewayTimeout, err.Status)
	}
	if !err.Retryable {
		t.Error("net timeout should be retryable")
	}
}

func TestFromGenericError(t *testing.T) {
	genericErr := errors.New("something went wrong")
	err := From(genericErr)

	if err.Code != CodeInternal {
		t.Errorf("expected code %s, got %s", CodeInternal, err.Code)
	}
	if err.Status != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, err.Status)
	}
	if err.Retryable {
		t.Error("generic error should not be retryable")
	}
	if err.Cause != genericErr {
		t.Error("cause should be preserved")
	}
}

func TestFromWrappedDeadline(t *testing.T) {
	// Test wrapped context.DeadlineExceeded
	wrapped := errors.Join(errors.New("outer"), context.DeadlineExceeded)
	err := From(wrapped)

	if err.Code != CodeTimeout {
		t.Errorf("expected code %s, got %s", CodeTimeout, err.Code)
	}
	if !err.Retryable {
		t.Error("wrapped deadline should be retryable")
	}
}

func TestFromActualNetTimeout(t *testing.T) {
	// Create a real network timeout scenario
	timeout := 1 * time.Millisecond
	conn, netErr := net.DialTimeout("tcp", "192.0.2.1:80", timeout) // TEST-NET-1, should timeout
	if conn != nil {
		_ = conn.Close()
	}

	if netErr != nil {
		err := From(netErr)
		// May be timeout or other net error
		if err.Code != CodeTimeout && err.Code != CodeInternal {
			t.Errorf("expected timeout or internal, got %s", err.Code)
		}
	}
}
