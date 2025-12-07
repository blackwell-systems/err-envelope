package errenvelope

import (
	"encoding/json"
	"net/http"
)

const (
	// HeaderTraceID is the standard header name for trace/request IDs.
	HeaderTraceID = "X-Request-Id"
)

// Write writes a consistent JSON error envelope to the response.
// If TraceID is missing on the error, it tries to derive it from the request.
func Write(w http.ResponseWriter, r *http.Request, err error) {
	e := From(err)
	if e == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if e.TraceID == "" {
		e.TraceID = TraceIDFromRequest(r)
	}

	if e.TraceID != "" {
		w.Header().Set(HeaderTraceID, e.TraceID)
	}

	status := e.Status
	if status == 0 {
		status = http.StatusInternalServerError
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(e)
}
