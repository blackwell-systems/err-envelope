package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/blackwell-systems/err-envelope"
)

func main() {
	mux := http.NewServeMux()

	// Example: validation error
	mux.HandleFunc("/validate", func(w http.ResponseWriter, r *http.Request) {
		email := r.URL.Query().Get("email")
		if email == "" {
			err := errenvelope.Validation(errenvelope.FieldErrors{
				"email": "is required",
			})
			errenvelope.Write(w, r, err)
			return
		}
		fmt.Fprintf(w, "Email validated: %s\n", email)
	})

	// Example: not found
	mux.HandleFunc("/user/", func(w http.ResponseWriter, r *http.Request) {
		userID := r.URL.Path[len("/user/"):]
		if userID == "" {
			errenvelope.Write(w, r, errenvelope.NotFound("User not found"))
			return
		}
		fmt.Fprintf(w, "User: %s\n", userID)
	})

	// Example: downstream error
	mux.HandleFunc("/downstream", func(w http.ResponseWriter, r *http.Request) {
		downErr := errors.New("payment service returned 502")
		errenvelope.Write(w, r, errenvelope.Downstream("payments", downErr))
	})

	// Example: unauthorized
	mux.HandleFunc("/protected", func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token == "" {
			errenvelope.Write(w, r, errenvelope.Unauthorized("Missing authorization token"))
			return
		}
		fmt.Fprintf(w, "Protected resource\n")
	})

	// Example: timeout
	mux.HandleFunc("/timeout", func(w http.ResponseWriter, r *http.Request) {
		errenvelope.Write(w, r, errenvelope.Timeout("Database query timed out"))
	})

	// Wrap with trace middleware
	handler := errenvelope.TraceMiddleware(mux)

	log.Println("Server listening on :8080")
	log.Println("Try:")
	log.Println("  curl http://localhost:8080/validate")
	log.Println("  curl http://localhost:8080/user/")
	log.Println("  curl http://localhost:8080/downstream")
	log.Println("  curl http://localhost:8080/protected")
	log.Println("  curl http://localhost:8080/timeout")

	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatal(err)
	}
}
