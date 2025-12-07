package chi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	errenvelope "github.com/blackwell-systems/err-envelope"
	"github.com/go-chi/chi/v5"
)

func TestTrace(t *testing.T) {
	r := chi.NewRouter()
	r.Use(Trace)

	r.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		traceID := errenvelope.TraceIDFromRequest(r)
		if traceID == "" {
			t.Error("expected trace ID to be set")
		}
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestTraceWithExistingHeader(t *testing.T) {
	r := chi.NewRouter()
	r.Use(Trace)

	existingTraceID := "existing-trace-id-123"

	r.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		traceID := errenvelope.TraceIDFromRequest(r)
		if traceID != existingTraceID {
			t.Errorf("expected trace ID %s, got %s", existingTraceID, traceID)
		}
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Request-Id", existingTraceID)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestWriteIntegration(t *testing.T) {
	r := chi.NewRouter()
	r.Use(Trace)

	r.Get("/error", func(w http.ResponseWriter, r *http.Request) {
		errenvelope.Write(w, r, errenvelope.NotFound("user not found"))
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

func TestValidationErrorIntegration(t *testing.T) {
	r := chi.NewRouter()
	r.Use(Trace)

	r.Post("/validate", func(w http.ResponseWriter, r *http.Request) {
		errenvelope.Write(w, r, errenvelope.Validation(errenvelope.FieldErrors{
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
