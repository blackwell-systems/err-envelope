package errenvelope

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTraceIDFromRequestHeader(t *testing.T) {
	r := httptest.NewRequest("GET", "/test", nil)
	r.Header.Set(HeaderTraceID, "trace-from-header")

	traceID := TraceIDFromRequest(r)
	if traceID != "trace-from-header" {
		t.Errorf("expected trace-from-header, got %s", traceID)
	}
}

func TestTraceIDFromRequestContext(t *testing.T) {
	r := httptest.NewRequest("GET", "/test", nil)
	ctx := WithTraceID(r.Context(), "trace-from-context")
	r = r.WithContext(ctx)

	traceID := TraceIDFromRequest(r)
	if traceID != "trace-from-context" {
		t.Errorf("expected trace-from-context, got %s", traceID)
	}
}

func TestTraceIDFromRequestHeaderPriority(t *testing.T) {
	// Header should take priority over context
	r := httptest.NewRequest("GET", "/test", nil)
	r.Header.Set(HeaderTraceID, "header-trace")
	ctx := WithTraceID(r.Context(), "context-trace")
	r = r.WithContext(ctx)

	traceID := TraceIDFromRequest(r)
	if traceID != "header-trace" {
		t.Errorf("expected header-trace (header priority), got %s", traceID)
	}
}

func TestTraceIDFromRequestNil(t *testing.T) {
	traceID := TraceIDFromRequest(nil)
	if traceID != "" {
		t.Errorf("expected empty string for nil request, got %s", traceID)
	}
}

func TestTraceIDFromRequestEmpty(t *testing.T) {
	r := httptest.NewRequest("GET", "/test", nil)

	traceID := TraceIDFromRequest(r)
	if traceID != "" {
		t.Errorf("expected empty string, got %s", traceID)
	}
}

func TestMiddlewareWithTraceID(t *testing.T) {
	r := httptest.NewRequest("GET", "/test", nil)
	ctx := WithTraceID(r.Context(), "new-trace")

	// Extract from context
	value := ctx.Value(traceKey)
	if value == nil {
		t.Fatal("expected trace ID in context")
	}

	traceID, ok := value.(string)
	if !ok {
		t.Fatal("expected trace ID to be string")
	}

	if traceID != "new-trace" {
		t.Errorf("expected new-trace, got %s", traceID)
	}
}

func TestTraceMiddleware(t *testing.T) {
	var capturedTraceID string

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedTraceID = TraceIDFromRequest(r)
		w.WriteHeader(http.StatusOK)
	})

	middleware := TraceMiddleware(handler)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)

	middleware.ServeHTTP(w, r)

	// Should generate a trace ID
	if capturedTraceID == "" {
		t.Error("expected trace ID to be generated")
	}

	// Trace ID should be 32 characters (16 bytes hex encoded)
	if len(capturedTraceID) != 32 {
		t.Errorf("expected trace ID length 32, got %d", len(capturedTraceID))
	}
}

func TestTraceMiddlewareWithExistingHeader(t *testing.T) {
	var capturedTraceID string

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedTraceID = TraceIDFromRequest(r)
		w.WriteHeader(http.StatusOK)
	})

	middleware := TraceMiddleware(handler)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	r.Header.Set(HeaderTraceID, "existing-trace")

	middleware.ServeHTTP(w, r)

	// Should preserve existing trace ID
	if capturedTraceID != "existing-trace" {
		t.Errorf("expected existing-trace, got %s", capturedTraceID)
	}
}

func TestTraceMiddlewareContextPropagation(t *testing.T) {
	var contextHasTrace bool

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		value := r.Context().Value(traceKey)
		contextHasTrace = value != nil
		w.WriteHeader(http.StatusOK)
	})

	middleware := TraceMiddleware(handler)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)

	middleware.ServeHTTP(w, r)

	if !contextHasTrace {
		t.Error("trace ID should be in context")
	}
}

func TestNewTraceIDUniqueness(t *testing.T) {
	// Generate multiple trace IDs and check they're unique
	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id := newTraceID()
		if ids[id] {
			t.Errorf("duplicate trace ID: %s", id)
		}
		ids[id] = true

		// Check length
		if len(id) != 32 {
			t.Errorf("expected length 32, got %d for ID %s", len(id), id)
		}

		// Check it's hex (only 0-9a-f characters)
		for _, c := range id {
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
				t.Errorf("invalid hex character %c in trace ID %s", c, id)
			}
		}
	}
}

func TestTraceMiddlewareIntegration(t *testing.T) {
	// Test full integration with error writing
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := NotFound("resource not found")
		Write(w, r, err)
	})

	middleware := TraceMiddleware(handler)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)

	middleware.ServeHTTP(w, r)

	// Check trace ID is in response header
	traceID := w.Header().Get(HeaderTraceID)
	if traceID == "" {
		t.Error("expected trace ID in response header")
	}
	if len(traceID) != 32 {
		t.Errorf("expected trace ID length 32, got %d", len(traceID))
	}
}
