// Package chi provides thin adapters for using err-envelope with chi router.
//
// Chi uses standard net/http handlers, so err-envelope works directly.
// This package exists for discoverability and convenience.
package chi

import (
	"net/http"

	errenvelope "github.com/blackwell-systems/err-envelope"
)

// Trace is a convenience wrapper around errenvelope.TraceMiddleware
// that returns a standard net/http middleware for chi.
//
// Chi can use errenvelope.TraceMiddleware directly; this exists for clarity.
//
// Example:
//
//	r := chi.NewRouter()
//	r.Use(chi.Trace)
func Trace(next http.Handler) http.Handler {
	return errenvelope.TraceMiddleware(next)
}
