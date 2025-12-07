// Package gin provides adapters for using err-envelope with Gin framework.
package gin

import (
	"net/http"

	errenvelope "github.com/blackwell-systems/err-envelope"
	"github.com/gin-gonic/gin"
)

// Trace wires err-envelope trace ID middleware into Gin's middleware chain.
//
// This generates or propagates trace IDs and makes them available via
// errenvelope.TraceIDFromRequest(c.Request).
//
// Example:
//
//	r := gin.Default()
//	r.Use(Trace())
//	r.GET("/user", func(c *gin.Context) {
//	    traceID := errenvelope.TraceIDFromRequest(c.Request)
//	    // ...
//	})
func Trace() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Wrap remaining chain with err-envelope trace middleware
		handler := errenvelope.TraceMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Update context with traced request
			c.Request = r
			c.Next()
		}))

		handler.ServeHTTP(c.Writer, c.Request)
	}
}

// Write sends a structured error response using err-envelope format.
//
// This is a convenience wrapper that extracts c.Writer and c.Request
// to call errenvelope.Write.
//
// Example:
//
//	r.GET("/user", func(c *gin.Context) {
//	    if userID == "" {
//	        Write(c, errenvelope.BadRequest("Missing user ID"))
//	        return
//	    }
//	    // ...
//	})
func Write(c *gin.Context, err error) {
	errenvelope.Write(c.Writer, c.Request, err)
}
