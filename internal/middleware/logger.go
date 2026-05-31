package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

// logger middleware logs each request with method, path, status, duration, request id, and client ip
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// wrap response writer to capture status code
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(rw, r)

		duration := time.Since(start).Milliseconds()

		// get request id from context
		reqID := GetRequestID(r.Context())
		// get real ip from context
		clientIP := GetRealIP(r)

		slog.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rw.statusCode,
			"duration_ms", duration,
			"request_id", reqID,
			"client_ip", clientIP,
		)
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
