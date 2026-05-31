package middleware

import (
	"context"
	"net/http"
	"time"
)

func Timeout(duration time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Create context with timeout
			ctx, cancel := context.WithTimeout(r.Context(), duration)
			defer cancel()

			// Create channel to receive completion signal
			done := make(chan bool, 1)

			// Run handler in goroutine
			go func() {
				next.ServeHTTP(w, r.WithContext(ctx))
				done <- true
			}()

			// Wait for either completion or timeout
			select {
			case <-done:
				// Handler completed within timeout
				return
			case <-ctx.Done():
				// Timeout exceeded
				w.Header().Set("Content-Type", "text/plain")
				w.WriteHeader(http.StatusGatewayTimeout)
				w.Write([]byte("Gateway Timeout\n"))
				return
			}
		})
	}
}
