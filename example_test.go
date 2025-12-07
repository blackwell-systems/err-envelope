package errenvelope_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	errenvelope "github.com/blackwell-systems/err-envelope"
)

// ExampleWrite_validation demonstrates handling validation errors with field details.
func ExampleWrite_validation() {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		email := r.URL.Query().Get("email")
		password := r.URL.Query().Get("password")

		if email == "" || password == "" {
			err := errenvelope.Validation(errenvelope.FieldErrors{
				"email":    "is required",
				"password": "is required",
			})
			errenvelope.Write(w, r, err)
			return
		}

		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/signup", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	fmt.Printf("Status: %d\n", w.Code)
	fmt.Printf("Content-Type: %s\n", w.Header().Get("Content-Type"))
	// Output:
	// Status: 400
	// Content-Type: application/json
}

// ExampleWrite_unauthorized demonstrates handling authentication errors.
func ExampleWrite_unauthorized() {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token == "" {
			err := errenvelope.Unauthorized("Missing authorization token")
			errenvelope.Write(w, r, err)
			return
		}

		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	fmt.Printf("Status: %d\n", w.Code)
	// Output:
	// Status: 401
}

// ExampleWrite_notFound demonstrates handling 404 errors.
func ExampleWrite_notFound() {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.PathValue("id")
		// Simulate user not found
		if userID == "999" {
			err := errenvelope.NotFound("User not found")
			errenvelope.Write(w, r, err)
			return
		}

		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/users/999", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	fmt.Printf("Status: %d\n", w.Code)
	// Output:
	// Status: 404
}

// ExampleWrite_conflict demonstrates handling conflict errors (e.g., duplicate resources).
func ExampleWrite_conflict() {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate duplicate email
		err := errenvelope.Conflict("User with this email already exists")
		errenvelope.Write(w, r, err)
	})

	req := httptest.NewRequest("POST", "/signup", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	fmt.Printf("Status: %d\n", w.Code)
	// Output:
	// Status: 409
}

// ExampleWrite_downstream demonstrates handling downstream service errors.
func ExampleWrite_downstream() {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate payment service failure
		paymentErr := fmt.Errorf("connection refused")
		err := errenvelope.Downstream("payment-service", paymentErr)
		errenvelope.Write(w, r, err)
	})

	req := httptest.NewRequest("POST", "/checkout", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	fmt.Printf("Status: %d\n", w.Code)
	// Output:
	// Status: 502
}

// ExampleTraceMiddleware demonstrates adding trace ID middleware.
func ExampleTraceMiddleware() {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/user", func(w http.ResponseWriter, r *http.Request) {
		// Trace ID is available in context
		traceID := errenvelope.GetTraceID(r.Context())
		fmt.Printf("Trace ID present: %v\n", traceID != "")

		w.WriteHeader(http.StatusOK)
	})

	// Wrap with trace middleware
	handler := errenvelope.TraceMiddleware(mux)

	req := httptest.NewRequest("GET", "/api/user", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// Output:
	// Trace ID present: true
}

// ExampleNew demonstrates creating a custom error with details.
func ExampleNew() {
	err := errenvelope.New(
		errenvelope.CodeInternal,
		http.StatusInternalServerError,
		"Database connection failed",
	)

	err = err.WithDetails(map[string]any{
		"database": "postgres",
		"host":     "db.example.com",
	})

	err = err.WithRetryable(true)

	fmt.Printf("Code: %s\n", err.Code)
	fmt.Printf("Message: %s\n", err.Message)
	fmt.Printf("Retryable: %v\n", err.Retryable)
	// Output:
	// Code: INTERNAL
	// Message: Database connection failed
	// Retryable: true
}

// ExampleFrom demonstrates mapping arbitrary errors to envelopes.
func ExampleFrom() {
	// Standard library error
	err := fmt.Errorf("something went wrong")
	envelope := errenvelope.From(err)

	fmt.Printf("Code: %s\n", envelope.Code)
	fmt.Printf("Status: %d\n", envelope.Status)
	// Output:
	// Code: INTERNAL
	// Status: 500
}

// ExampleValidation demonstrates field-level validation errors.
func ExampleValidation() {
	err := errenvelope.Validation(errenvelope.FieldErrors{
		"email":    "must be a valid email address",
		"age":      "must be at least 18",
		"password": "must be at least 8 characters",
	})

	fmt.Printf("Code: %s\n", err.Code)
	fmt.Printf("Status: %d\n", err.Status)
	fmt.Printf("Retryable: %v\n", err.Retryable)
	// Output:
	// Code: VALIDATION_FAILED
	// Status: 400
	// Retryable: false
}
