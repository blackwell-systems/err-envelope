package echo

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	errenvelope "github.com/blackwell-systems/err-envelope"
	"github.com/labstack/echo/v4"
)

func TestTrace(t *testing.T) {
	e := echo.New()
	e.Use(Trace)

	e.GET("/test", func(c echo.Context) error {
		traceID := errenvelope.TraceIDFromRequest(c.Request())
		if traceID == "" {
			t.Error("expected trace ID to be set")
		}
		return c.NoContent(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestTraceWithExistingHeader(t *testing.T) {
	e := echo.New()
	e.Use(Trace)

	existingTraceID := "existing-trace-id-123"

	e.GET("/test", func(c echo.Context) error {
		traceID := errenvelope.TraceIDFromRequest(c.Request())
		if traceID != existingTraceID {
			t.Errorf("expected trace ID %s, got %s", existingTraceID, traceID)
		}
		return c.NoContent(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Request-Id", existingTraceID)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestWrite(t *testing.T) {
	e := echo.New()
	e.Use(Trace)

	e.GET("/error", func(c echo.Context) error {
		return Write(c, errenvelope.NotFound("user not found"))
	})

	req := httptest.NewRequest("GET", "/error", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}

	var response map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response["code"] != "NOT_FOUND" {
		t.Errorf("expected code NOT_FOUND, got %v", response["code"])
	}
	if response["message"] != "user not found" {
		t.Errorf("expected message 'user not found', got %v", response["message"])
	}
	if response["trace_id"] == nil || response["trace_id"] == "" {
		t.Error("expected trace_id to be set")
	}
}

func TestWriteValidationError(t *testing.T) {
	e := echo.New()
	e.Use(Trace)

	e.POST("/validate", func(c echo.Context) error {
		return Write(c, errenvelope.Validation(errenvelope.FieldErrors{
			"email": "invalid format",
		}))
	})

	req := httptest.NewRequest("POST", "/validate", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	var response map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response["code"] != "VALIDATION_FAILED" {
		t.Errorf("expected code VALIDATION_FAILED, got %v", response["code"])
	}

	details, ok := response["details"].(map[string]any)
	if !ok {
		t.Fatal("expected details to be a map")
	}
	fields, ok := details["fields"].(map[string]any)
	if !ok {
		t.Fatal("expected fields to be a map")
	}
	if fields["email"] != "invalid format" {
		t.Errorf("expected email error 'invalid format', got %v", fields["email"])
	}
}

func TestWriteBadRequest(t *testing.T) {
	e := echo.New()
	e.Use(Trace)

	e.POST("/bad", func(c echo.Context) error {
		return Write(c, errenvelope.BadRequest("malformed JSON"))
	})

	req := httptest.NewRequest("POST", "/bad", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	var response map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response["code"] != "BAD_REQUEST" {
		t.Errorf("expected code BAD_REQUEST, got %v", response["code"])
	}
	if response["message"] != "malformed JSON" {
		t.Errorf("expected message 'malformed JSON', got %v", response["message"])
	}
}

func TestWriteUnauthorized(t *testing.T) {
	e := echo.New()
	e.Use(Trace)

	e.GET("/protected", func(c echo.Context) error {
		return Write(c, errenvelope.Unauthorized("missing token"))
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}

	var response map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response["code"] != "UNAUTHORIZED" {
		t.Errorf("expected code UNAUTHORIZED, got %v", response["code"])
	}
}

func TestWriteReturnsNil(t *testing.T) {
	e := echo.New()
	e.Use(Trace)

	e.GET("/error", func(c echo.Context) error {
		err := Write(c, errenvelope.NotFound("not found"))
		if err != nil {
			t.Errorf("Write should return nil, got %v", err)
		}
		return err
	})

	req := httptest.NewRequest("GET", "/error", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
}
