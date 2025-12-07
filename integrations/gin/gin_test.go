package gin

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	errenvelope "github.com/blackwell-systems/err-envelope"
	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestTrace(t *testing.T) {
	r := gin.New()
	r.Use(Trace())

	r.GET("/test", func(c *gin.Context) {
		traceID := errenvelope.TraceIDFromRequest(c.Request)
		if traceID == "" {
			t.Error("expected trace ID to be set")
		}
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestTraceWithExistingHeader(t *testing.T) {
	r := gin.New()
	r.Use(Trace())

	existingTraceID := "existing-trace-id-123"

	r.GET("/test", func(c *gin.Context) {
		traceID := errenvelope.TraceIDFromRequest(c.Request)
		if traceID != existingTraceID {
			t.Errorf("expected trace ID %s, got %s", existingTraceID, traceID)
		}
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Request-Id", existingTraceID)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestWrite(t *testing.T) {
	r := gin.New()
	r.Use(Trace())

	r.GET("/error", func(c *gin.Context) {
		Write(c, errenvelope.NotFound("user not found"))
	})

	req := httptest.NewRequest("GET", "/error", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

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
	r := gin.New()
	r.Use(Trace())

	r.POST("/validate", func(c *gin.Context) {
		Write(c, errenvelope.Validation(errenvelope.FieldErrors{
			"email": "invalid format",
		}))
	})

	req := httptest.NewRequest("POST", "/validate", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

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
	r := gin.New()
	r.Use(Trace())

	r.POST("/bad", func(c *gin.Context) {
		Write(c, errenvelope.BadRequest("malformed JSON"))
	})

	req := httptest.NewRequest("POST", "/bad", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

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
	r := gin.New()
	r.Use(Trace())

	r.GET("/protected", func(c *gin.Context) {
		Write(c, errenvelope.Unauthorized("missing token"))
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

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
