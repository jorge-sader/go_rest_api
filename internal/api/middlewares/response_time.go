package middlewares

import (
	"fmt"
	"net/http"
	"time"
)

// ===========> WARNING: THIS IS NOT ACCURATE! <===========
//

// FIXME: Find way to measure accurately or remove
func ResponseTime(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		fmt.Println("Received request in response time middleware")

		// Create a custom ResponseWriter to capture the status code
		wrappedWriter := &responseWriter{ResponseWriter: w, status: http.StatusOK}

		// Calculate duration
		duration := time.Since(start)
		w.Header().Set("X-Response-Time", duration.String()) // We need the value of duration to pass it but we want to account for next in our computation

		next.ServeHTTP(wrappedWriter, r) // once we pass the writer we cant update the heading any more.

		// log the request details
		fmt.Printf("Method: %s, URL: %s, Status: %d, Duration: %v\n", r.Method, r.URL, wrappedWriter.status, duration.String())
	})
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}
