package errenvelope

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

type ctxKey string

const traceKey ctxKey = "errenvelope.trace_id"

// TraceIDFromRequest extracts the trace ID from the request header or context.
func TraceIDFromRequest(r *http.Request) string {
	if r == nil {
		return ""
	}
	// Prefer header
	if id := r.Header.Get(HeaderTraceID); id != "" {
		return id
	}
	// Then context
	if v := r.Context().Value(traceKey); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// WithTraceID adds a trace ID to the context.
func WithTraceID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, traceKey, id)
}

// TraceMiddleware generates or propagates a trace ID for each request.
func TraceMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(HeaderTraceID)
		if id == "" {
			id = newTraceID()
		}
		ctx := WithTraceID(r.Context(), id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func newTraceID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}
