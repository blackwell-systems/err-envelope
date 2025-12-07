// Package echo provides adapters for using err-envelope with Echo framework.
package echo

import (
	"net/http"

	errenvelope "github.com/blackwell-systems/err-envelope"
	echofw "github.com/labstack/echo/v4"
)

// Trace adapts err-envelope trace middleware to Echo's middleware interface.
//
// This generates or propagates trace IDs and makes them available via
// errenvelope.TraceIDFromRequest(c.Request()).
//
// Example:
//
//	e := echo.New()
//	e.Use(Trace)
//	e.GET("/user", func(c echo.Context) error {
//	    traceID := errenvelope.TraceIDFromRequest(c.Request())
//	    // ...
//	    return nil
//	})
func Trace(next echofw.HandlerFunc) echofw.HandlerFunc {
	return func(c echofw.Context) error {
		// Wrap with err-envelope trace middleware
		handler := errenvelope.TraceMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Update context with traced request
			c.SetRequest(r)
			_ = next(c)
		}))

		handler.ServeHTTP(c.Response().Writer, c.Request())
		return nil
	}
}

// Write sends a structured error response using err-envelope format.
//
// This is a convenience wrapper that extracts c.Response().Writer and c.Request()
// to call errenvelope.Write.
//
// Example:
//
//	e.GET("/user", func(c echo.Context) error {
//	    if userID == "" {
//	        return Write(c, errenvelope.BadRequest("Missing user ID"))
//	    }
//	    // ...
//	    return nil
//	})
func Write(c echofw.Context, err error) error {
	errenvelope.Write(c.Response().Writer, c.Request(), err)
	return nil
}
